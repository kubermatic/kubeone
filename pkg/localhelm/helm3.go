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
	"helm.sh/helm/v3/pkg/storage/driver"

	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
)

const helmStorageDriver = "secret"

func Deploy(st *state.State) error {
	if len(st.Cluster.HelmReleases) == 0 {
		return nil
	}

	konfig, err := kubeconfig.Download(st)
	if err != nil {
		return err
	}

	restClientGetter := &genericclioptions.ConfigFlags{
		Namespace:  pointer.String("default"),
		KubeConfig: pointer.String(string(konfig)),
		WrapConfigFn: func(rc *rest.Config) *rest.Config {
			err := kubeconfig.TunnelRestConfig(st, rc)
			if err != nil {
				panic(err)
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

	for _, rh := range st.Cluster.HelmReleases {
		st.Logger.Infof("Deploying helm chart %s as release %s", rh.Chart, rh.ReleaseName)

		var valueFiles []string
		for _, value := range rh.Values {
			if value.File != "" {
				valueFiles = append(valueFiles, value.File)
			}

			if value.Inline != nil {
				inlineValues, err := os.CreateTemp("", "inline-helm-values-*")
				if err != nil {
					return err
				}

				inlineValuesName := inlineValues.Name()
				defer os.Remove(inlineValuesName)

				valuesBuf := bytes.NewBuffer(value.Inline)
				_, err = io.Copy(inlineValues, valuesBuf)
				if err != nil {
					inlineValues.Close()

					return err
				}

				inlineValues.Close()
				valueFiles = append(valueFiles, inlineValuesName)
			}
		}

		valueOpts := &values.Options{
			ValueFiles: valueFiles,
		}
		providers := getter.All(helmSettings)
		vals, err := valueOpts.MergeValues(providers)

		if err := cfg.Init(restClientGetter, rh.Namespace, helmStorageDriver, st.Logger.Debugf); err != nil {
			return err
		}

		histClient := helmaction.NewHistory(cfg)
		histClient.Max = 1
		_, err = histClient.Run(rh.ReleaseName)
		switch {
		case errors.Is(err, driver.ErrReleaseNotFound):
			helmInstall := helmaction.NewInstall(cfg)
			helmInstall.DependencyUpdate = true
			helmInstall.CreateNamespace = true
			helmInstall.ReleaseName = rh.ReleaseName
			helmInstall.RepoURL = rh.RepoURL
			helmInstall.Version = rh.Version

			chartRequested, err := getChart(rh.Chart, helmInstall.ChartPathOptions, helmSettings, providers)
			if err != nil {
				return err
			}

			_, err = helmInstall.RunWithContext(st.Context, chartRequested, vals)

			return err
		case err == nil:
			helmUpgrade := helmaction.NewUpgrade(cfg)
			helmUpgrade.Namespace = rh.Namespace
			helmUpgrade.RepoURL = rh.RepoURL
			helmUpgrade.Version = rh.Version
			helmUpgrade.Install = true
			helmUpgrade.DependencyUpdate = true
			helmUpgrade.MaxHistory = 5

			chartRequested, err := getChart(rh.Chart, helmUpgrade.ChartPathOptions, helmSettings, providers)
			if err != nil {
				return err
			}

			_, err = helmUpgrade.RunWithContext(st.Context, rh.ReleaseName, chartRequested, vals)

			return err
		default:
			return err
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
		return nil, err
	}

	return newChart(chartPath, chartName, providers, helmSettings)
}

func newChart(chartPath string, chartName string, providers getter.Providers, settings *helmcli.EnvSettings) (*chart.Chart, error) {
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	switch chartRequested.Metadata.Type {
	case "", "application":
	default:
		panic(fmt.Errorf("%s charts are not installable", chartRequested.Metadata.Type))
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
		return nil, err
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	return chartRequested, nil
}

func newActionConfiguration(debug bool) (*helmaction.Configuration, error) {
	actionConfig := &helmaction.Configuration{}
	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stdout),
	)
	if err != nil {
		return nil, err
	}
	actionConfig.RegistryClient = registryClient

	return actionConfig, nil
}
