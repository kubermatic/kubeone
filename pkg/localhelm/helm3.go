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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	helmaction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/pointer"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	helmStorageDriver = "secret"
	helmStorageType   = "sh.helm.release.v1"
	releasedByKubeone = "kubeone.k8c.io/released-by-kubeone"
)

func Deploy(st *state.State) error {
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

	restClientGetter := newRestClientGetter(tmpKubeConf.Name(), st)
	helmSettings := newHelmSettings(st.Verbose)
	cfg, err := newActionConfiguration(helmSettings.Debug)
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(st.RESTConfig)
	if err != nil {
		return fail.Config(err, "init new kubernetes client")
	}

	// all namespaces
	noNamespaceSecretsClient := kubeClient.CoreV1().Secrets("")
	releasesToUninstall, err := driver.
		NewSecrets(noNamespaceSecretsClient).
		List(releasesFilterFn(st.Cluster.HelmReleases, st.Logger))
	if err != nil {
		return err
	}

	for _, release := range st.Cluster.HelmReleases {
		var valueFiles []string
		for _, value := range release.Values {
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

		if err = cfg.Init(restClientGetter, release.Namespace, helmStorageDriver, st.Logger.Debugf); err != nil {
			return fail.Runtime(err, "initializing helm action configuration")
		}

		histClient := helmaction.NewHistory(cfg)
		histClient.Max = 1
		existingReleases, err := histClient.Run(release.ReleaseName)

		switch {
		case errors.Is(err, driver.ErrReleaseNotFound):
			if err = installRelease(st.Context, cfg, release, helmSettings, providers, st.DynamicClient, vals, st.Logger); err != nil {
				return err
			}
		case err == nil:
			if err = upgradeRelease(st.Context, cfg, release, helmSettings, providers, st.DynamicClient, vals, existingReleases, st.Logger); err != nil {
				return err
			}
		default:
			return fail.Runtime(err, "helm releases history")
		}
	}

	return uninstallReleases(releasesToUninstall, cfg, restClientGetter, st.Logger)
}

func releasesFilterFn(helmReleases []kubeoneapi.HelmRelease, logger logrus.FieldLogger) func(rel *helmrelease.Release) bool {
	return func(rel *helmrelease.Release) bool {
		for _, hr := range helmReleases {
			if rel.Name == hr.ReleaseName && rel.Namespace == hr.Namespace && rel.Chart.Name() == hr.Chart {
				return false
			}
		}

		_, found := rel.Labels[releasedByKubeone]
		if found {
			logger.Infof("queue %s/%s v%d helm release to uninstall", rel.Namespace, rel.Name, rel.Version)
		}

		return found
	}
}

func newHelmSettings(verbose bool) *helmcli.EnvSettings {
	helmSettings := helmcli.New()
	helmSettings.Debug = verbose

	return helmSettings
}

func newRestClientGetter(kubeConfigFileName string, st *state.State) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		Namespace:  pointer.New("default"),
		KubeConfig: pointer.New(kubeConfigFileName),
		WrapConfigFn: func(rc *rest.Config) *rest.Config {
			tunnelErr := kubeconfig.TunnelRestConfig(st, rc)
			if tunnelErr != nil {
				panic(tunnelErr)
			}

			return rc
		},
	}
}

func helmReleasesEqual(rel *helmrelease.Release, oldRels []*helmrelease.Release) bool {
	if len(oldRels) == 0 {
		return false
	}

	sort.Slice(oldRels, func(i, j int) bool {
		return oldRels[i].Version > oldRels[j].Version
	})
	latestHelmRelease := oldRels[0]

	if rel.Chart.Metadata.Version != latestHelmRelease.Chart.Metadata.Version {
		return false
	}

	if !cmp.Equal(neverNilMap(rel.Config), neverNilMap(latestHelmRelease.Config)) {
		return false
	}

	return rel.Manifest == latestHelmRelease.Manifest
}

func neverNilMap(m1 map[string]any) map[string]any {
	if m1 == nil {
		return map[string]any{}
	}

	return m1
}

func upgradeRelease(
	ctx context.Context,
	cfg *helmaction.Configuration,
	release kubeoneapi.HelmRelease,
	helmSettings *helmcli.EnvSettings,
	providers getter.Providers,
	dynclient ctrlruntimeclient.Client,
	vals map[string]interface{},
	existingHelmReleases []*helmrelease.Release,
	logger logrus.FieldLogger,
) error {
	helmInstall := newHelmInstallClient(cfg, release)
	helmInstall.DryRun = true
	dryRunHelmRelease, err := runInstallRelease(ctx, release, helmInstall, helmSettings, providers, vals)
	if err != nil {
		return err
	}

	if helmReleasesEqual(dryRunHelmRelease, existingHelmReleases) {
		logger.Infof("Skip upgrading helm chart %s as release %s", release.Chart, release.ReleaseName)

		return nil
	}

	helmUpgrade := helmaction.NewUpgrade(cfg)
	helmUpgrade.Install = true
	helmUpgrade.DependencyUpdate = true
	helmUpgrade.ResetValues = true
	helmUpgrade.MaxHistory = 5
	helmUpgrade.Namespace = release.Namespace
	helmUpgrade.RepoURL = release.RepoURL
	helmUpgrade.Version = release.Version

	chartRequested, err := getChart(release.Chart, helmUpgrade.ChartPathOptions, helmSettings, providers)
	if err != nil {
		return err
	}

	logger.Infof("Upgrading helm chart %s as release %s", release.Chart, release.ReleaseName)
	rel, err := helmUpgrade.RunWithContext(ctx, release.ReleaseName, chartRequested, vals)
	if err != nil {
		return fail.Runtime(err, "upgrading helm release %q from chart %q", release.Chart, release.ReleaseName)
	}

	secretObjectKey := ctrlruntimeclient.ObjectKey{
		Name:      makeKey(rel.Name, rel.Version),
		Namespace: release.Namespace,
	}

	return addReleaseSecretLabels(ctx, secretObjectKey, dynclient)
}

func newHelmInstallClient(cfg *helmaction.Configuration, release kubeoneapi.HelmRelease) *helmaction.Install {
	helmInstall := helmaction.NewInstall(cfg)
	helmInstall.DependencyUpdate = true
	helmInstall.CreateNamespace = true
	helmInstall.Namespace = release.Namespace
	helmInstall.ReleaseName = release.ReleaseName
	helmInstall.RepoURL = release.RepoURL
	helmInstall.Version = release.Version

	return helmInstall
}

func installRelease(
	ctx context.Context,
	cfg *helmaction.Configuration,
	release kubeoneapi.HelmRelease,
	helmSettings *helmcli.EnvSettings,
	providers getter.Providers,
	dynclient ctrlruntimeclient.Client,
	vals map[string]interface{},
	logger logrus.FieldLogger,
) error {
	logger.Infof("Deploying helm chart %s as release %s", release.Chart, release.ReleaseName)
	helmInstall := newHelmInstallClient(cfg, release)
	rel, err := runInstallRelease(ctx, release, helmInstall, helmSettings, providers, vals)
	if err != nil {
		return err
	}

	secretObjectKey := ctrlruntimeclient.ObjectKey{
		Name:      makeKey(rel.Name, rel.Version),
		Namespace: release.Namespace,
	}

	return addReleaseSecretLabels(ctx, secretObjectKey, dynclient)
}

func runInstallRelease(
	ctx context.Context,
	release kubeoneapi.HelmRelease,
	client *helmaction.Install,
	helmSettings *helmcli.EnvSettings,
	providers getter.Providers,
	vals map[string]interface{},
) (*helmrelease.Release, error) {
	chartRequested, err := getChart(release.Chart, client.ChartPathOptions, helmSettings, providers)
	if err != nil {
		return nil, err
	}

	rel, err := client.RunWithContext(ctx, chartRequested, vals)
	if err != nil {
		return nil, fail.Runtime(err, "installing helm release %q from chart %q", release.Chart, release.ReleaseName)
	}

	return rel, nil
}

func uninstallReleases(
	toUninstall []*helmrelease.Release,
	cfg *helmaction.Configuration,
	restClientGetter *genericclioptions.ConfigFlags,
	logger logrus.FieldLogger,
) error {
	for _, release := range toUninstall {
		if err := cfg.Init(restClientGetter, release.Namespace, helmStorageDriver, logger.Debugf); err != nil {
			return fail.Runtime(err, "initializing helm action configuration")
		}

		helmUninstall := helmaction.NewUninstall(cfg)
		resp, err := helmUninstall.Run(release.Name)
		if err != nil {
			return fail.Runtime(err, "uninstalling helm release %s/%s", release.Namespace, release.Name)
		}

		logger.Infof("uninstalled helm release %s/%s: %s", release.Namespace, release.Name, resp.Info)
	}

	return nil
}

func addReleaseSecretLabels(ctx context.Context, releaseNamespacedName ctrlruntimeclient.ObjectKey, dynclient ctrlruntimeclient.Client) error {
	releaseSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      releaseNamespacedName.Name,
			Namespace: releaseNamespacedName.Namespace,
		},
	}

	if err := dynclient.Get(ctx, releaseNamespacedName, &releaseSecret); err != nil {
		return fail.Runtime(err, "getting secret object for the %s secret", releaseNamespacedName)
	}

	releaseSecretOld := releaseSecret.DeepCopy()
	if releaseSecret.Labels == nil {
		releaseSecret.Labels = map[string]string{}
	}
	releaseSecret.Labels[releasedByKubeone] = ""
	err := dynclient.Patch(ctx, &releaseSecret, ctrlruntimeclient.MergeFrom(releaseSecretOld))

	return fail.Runtime(err, "patching labels of helm release secret %s", releaseNamespacedName)
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

func makeKey(rlsname string, version int) string {
	return fmt.Sprintf("%s.%s.v%d", helmStorageType, rlsname, version)
}
