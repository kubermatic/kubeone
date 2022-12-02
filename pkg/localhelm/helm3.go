/*
Copyright 2022 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package localhelm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	helmaction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/pointer"
	"k8c.io/kubeone/pkg/state"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	helmStorageDriver = "secret"
	releasedByKubeone = "released-by-kubeone"
)

func Deploy(st *state.State) error {
	if len(st.Cluster.HelmReleases) == 0 {
		return nil
	}

	konfigBuf, err := kubeconfig.Download(st)
	if err != nil {
		return err
	}

	tmpKubeConf, err := os.CreateTemp("", "kubeone-kubeconfig-*")
	if err != nil {
		return fail.Runtime(err, "creating temp file for helm kubeconfig")
	}
	defer func() {
		name := tmpKubeConf.Name()
		tmpKubeConf.Close()
		os.Remove(name)
	}()

	n, err := tmpKubeConf.Write(konfigBuf)
	if err != nil {
		return fail.Runtime(err, "wring temp file for helm kubeconfig")
	}
	if n != len(konfigBuf) {
		return fail.NewRuntimeError("incorrect number of bytes written to temp kubeconfig", "")
	}

	restClientGetter := &genericclioptions.ConfigFlags{
		Namespace:  pointer.New("default"),
		KubeConfig: pointer.New(tmpKubeConf.Name()),
		WrapConfigFn: func(rc *rest.Config) *rest.Config {
			tunnelErr := kubeconfig.TunnelRestConfig(st, rc)
			if tunnelErr != nil {
				panic(tunnelErr)
			}

			return rc
		},
	}

	var helmSettings = helmcli.New()
	helmSettings.Debug = st.Verbose

	cfg, err := newActionConfiguration(helmSettings.Debug)
	if err != nil {
		return err
	}

	kubeClinet, err := kubernetes.NewForConfig(st.RESTConfig)
	if err != nil {
		return fail.Config(err, "init new kubernetes client")
	}

	// all namespaces
	releasesToUninstall, err := driver.NewSecrets(kubeClinet.CoreV1().Secrets("")).List(func(rel *release.Release) bool {
		for _, hr := range st.Cluster.HelmReleases {
			if rel.Name == hr.ReleaseName && rel.Namespace == hr.Namespace && rel.Chart.Name() == hr.Chart {
				return false
			}
		}

		_, ok := rel.Labels[releasedByKubeone]

		return ok
	})
	if err != nil {
		return err
	}

	for _, rh := range st.Cluster.HelmReleases {
		st.Logger.Infof("Deploying helm chart %s as release %s", rh.Chart, rh.ReleaseName)

		var valueFiles []string
		for _, value := range rh.Values {
			if value.ValuesFile != "" {
				valueFiles = append(valueFiles, value.ValuesFile)
			}

			if value.Inline != nil {
				inlineValues, errTmp := os.CreateTemp("", "inline-helm-values-*")
				if errTmp != nil {
					return fail.Runtime(errTmp, "creating temp file for helm inline values")
				}

				inlineValuesName := inlineValues.Name()
				defer os.Remove(inlineValuesName)

				valuesBuf := bytes.NewBuffer(value.Inline)
				_, err = io.Copy(inlineValues, valuesBuf)
				if err != nil {
					inlineValues.Close()

					return fail.Runtime(err, "copying helm inline values to the temp file")
				}

				inlineValues.Close()
				valueFiles = append(valueFiles, inlineValuesName)
			}
		}

		valueOpts := &values.Options{
			ValueFiles: valueFiles,
		}
		providers := getter.All(helmSettings)
		vals, errMerge := valueOpts.MergeValues(providers)
		if errMerge != nil {
			return fail.Runtime(errMerge, "merging helm values")
		}

		if err = cfg.Init(restClientGetter, rh.Namespace, helmStorageDriver, st.Logger.Debugf); err != nil {
			return fail.Runtime(err, "initializing helm action configuration")
		}

		histClient := helmaction.NewHistory(cfg)
		histClient.Max = 1
		_, err = histClient.Run(rh.ReleaseName)
		switch {
		case errors.Is(err, driver.ErrReleaseNotFound):
			helmInstall := helmaction.NewInstall(cfg)
			helmInstall.DependencyUpdate = true
			helmInstall.CreateNamespace = true
			helmInstall.Namespace = rh.Namespace
			helmInstall.ReleaseName = rh.ReleaseName
			helmInstall.RepoURL = rh.RepoURL
			helmInstall.Version = rh.Version

			chartRequested, chartErr := getChart(rh.Chart, helmInstall.ChartPathOptions, helmSettings, providers)
			if chartErr != nil {
				return chartErr
			}

			rel, errInstall := helmInstall.RunWithContext(st.Context, chartRequested, vals)
			if errInstall != nil {
				return fail.Runtime(errInstall, "installing helm release %q from chart %q", rh.Chart, rh.ReleaseName)
			}

			rel.Labels[releasedByKubeone] = "yes"
			if err = driver.NewSecrets(kubeClinet.CoreV1().Secrets(rh.Namespace)).Update(rel.Name, rel); err != nil {
				return fail.Runtime(err, "adding kubeone labels to helm release %s/%s", rh.Namespace, rh.ReleaseName)
			}
		case err == nil:
			helmUpgrade := helmaction.NewUpgrade(cfg)
			helmUpgrade.Install = true
			helmUpgrade.DependencyUpdate = true
			helmUpgrade.MaxHistory = 5
			helmUpgrade.Namespace = rh.Namespace
			helmUpgrade.RepoURL = rh.RepoURL
			helmUpgrade.Version = rh.Version

			chartRequested, chartErr := getChart(rh.Chart, helmUpgrade.ChartPathOptions, helmSettings, providers)
			if chartErr != nil {
				return chartErr
			}

			rel, errUpgrade := helmUpgrade.RunWithContext(st.Context, rh.ReleaseName, chartRequested, vals)
			if errUpgrade != nil {
				return fail.Runtime(errUpgrade, "upgrading helm release %q from chart %q", rh.Chart, rh.ReleaseName)
			}

			rel.Labels[releasedByKubeone] = "yes"
			if err = driver.NewSecrets(kubeClinet.CoreV1().Secrets(rh.Namespace)).Update(rel.Name, rel); err != nil {
				return fail.Runtime(err, "adding kubeone labels to helm release %s/%s", rh.Namespace, rh.ReleaseName)
			}
		default:
			return fail.Runtime(err, "helm releases history")
		}
	}

	for _, rel := range releasesToUninstall {
		if err = cfg.Init(restClientGetter, rel.Namespace, helmStorageDriver, st.Logger.Debugf); err != nil {
			return fail.Runtime(err, "initializing helm action configuration")
		}

		helmUninstall := helmaction.NewUninstall(cfg)
		if _, err = helmUninstall.Run(rel.Name); err != nil {
			return fail.Runtime(err, "uninstalling helm release %s/%s", rel.Namespace, rel.Name)
		}
	}

	return nil
}

func getChart(
	chartName string,
	chartPathOpts helmaction.ChartPathOptions,
	helmSettings *helmcli.EnvSettings,
	providers getter.Providers,
) (*chart.Chart, error) {
	chartPath, err := chartPathOpts.LocateChart(chartName, helmSettings)
	if err != nil {
		return nil, fail.Runtime(err, "locating helm chart")
	}

	return newChart(chartPath, chartName, providers, helmSettings)
}

func newChart(chartPath string, chartName string, providers getter.Providers, settings *helmcli.EnvSettings) (*chart.Chart, error) {
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, fail.Runtime(err, "loading helm chart")
	}

	switch chartRequested.Metadata.Type {
	case "", "application":
	default:
		return nil, fail.ConfigValidation(fmt.Errorf("%s charts are not installable", chartRequested.Metadata.Type))
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if errMiss := helmaction.CheckDependencies(chartRequested, req); errMiss != nil {
			chartRequested, err = dependencyUpdate(chartName, settings, providers)
			if err != nil {
				return nil, err
			}
		}
	}

	return chartRequested, nil
}

func dependencyUpdate(chartPath string, settings *helmcli.EnvSettings, providers []getter.Provider) (*chart.Chart, error) {
	mgr := &downloader.Manager{
		Out:              os.Stdout,
		ChartPath:        chartPath,
		SkipUpdate:       false,
		Getters:          providers,
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
		Debug:            settings.Debug,
	}
	if err := mgr.Update(); err != nil {
		return nil, fail.Runtime(err, "getting helm chart dependencies")
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, fail.Runtime(err, "loading helm chart after update")
	}

	return chartRequested, nil
}

func newActionConfiguration(debug bool) (*helmaction.Configuration, error) {
	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stdout),
	)

	return &helmaction.Configuration{
		RegistryClient: registryClient,
	}, fail.Runtime(err, "initializing new helm registry client")
}
