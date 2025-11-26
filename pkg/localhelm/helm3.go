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
	"path"
	"sort"
	"strings"

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
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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

	helmSettings := newHelmSettings(st.Verbose)
	helmCfg, err := newActionConfiguration(helmSettings.Debug)
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
		List(releasesFilterFn(st.Cluster.Addons.OnlyHelmReleases(), st.Logger))
	if err != nil {
		return err
	}

	for _, release := range st.Cluster.Addons.OnlyHelmReleases() {
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

		restClientGetter := newRestClientGetter(tmpKubeConf.Name(), release.Namespace, st)
		if err = helmCfg.Init(restClientGetter, release.Namespace, helmStorageDriver, st.Logger.Debugf); err != nil {
			return fail.Runtime(err, "initializing helm action configuration")
		}

		histClient := helmaction.NewHistory(helmCfg)
		histClient.Max = 1
		existingReleases, err := histClient.Run(release.ReleaseName)

		switch {
		case errors.Is(err, driver.ErrReleaseNotFound):
			if err = installRelease(st.Context, helmCfg, release, helmSettings, providers, st.DynamicClient, vals, st.Logger); err != nil {
				return err
			}
		case err == nil:
			if err = upgradeRelease(st.Context, helmCfg, release, helmSettings, providers, st.DynamicClient, vals, existingReleases, st.Logger); err != nil {
				return err
			}
		default:
			return fail.Runtime(err, "helm releases history")
		}
	}

	return uninstallReleases(releasesToUninstall, helmCfg, tmpKubeConf.Name(), st)
}

func releasesFilterFn(helmReleases []kubeoneapi.HelmRelease, logger logrus.FieldLogger) func(rel *helmrelease.Release) bool {
	return func(rel *helmrelease.Release) bool {
		for _, hr := range helmReleases {
			if rel.Name == hr.ReleaseName && rel.Namespace == hr.Namespace {
				chartName := hr.Chart
				if hr.RepoURL == "" && !strings.HasPrefix(hr.ChartURL, "oci://") {
					chartName, _ = GetChartNameFromChartYAML(chartName)
				}
				if chartName == rel.Chart.Name() {
					return false
				}
			}
		}
		_, found := rel.Labels[releasedByKubeone]
		if found {
			logger.Infof("queue %s/%s v%d helm release to uninstall", rel.Namespace, rel.Name, rel.Version)
		}

		return found
	}
}

// GetChartNameFromChartYAML is used to extract chart name from Chart.yaml in case of local charts
func GetChartNameFromChartYAML(chartPath string) (string, error) {
	chartYAMLPath := path.Join(chartPath, "Chart.yaml")
	yamlFile, err := os.ReadFile(chartYAMLPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chartMetadata struct {
		Name string `yaml:"name"`
	}
	err = yaml.Unmarshal(yamlFile, &chartMetadata)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal Chart.yaml: %w", err)
	}

	return chartMetadata.Name, nil
}

func newHelmSettings(verbose bool) *helmcli.EnvSettings {
	helmSettings := helmcli.New()
	helmSettings.Debug = verbose

	return helmSettings
}

func newRestClientGetter(kubeConfigFileName, namespace string, st *state.State) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		Namespace:  ptr.To(namespace),
		KubeConfig: ptr.To(kubeConfigFileName),
		WrapConfigFn: func(rc *rest.Config) *rest.Config {
			tunnelErr := kubeconfig.TunnelRestConfig(st, rc)
			if tunnelErr != nil {
				panic(tunnelErr)
			}

			return rc
		},
	}
}

// cleanManifest removes Helm-generated comment lines (any line where '#' is the first
// non-whitespace character) from the rendered manifest before comparison.
//
// Helm injects comments like "# Source: ..." into manifests for debugging, which
// causes byte-for-byte manifest equality checks to fail even when the actual
// Kubernetes resources are identical. Since these comments carry no semantic
// meaning for resource reconciliation, we safely ignore them during equality
// comparisons to avoid unnecessary Helm upgrades or false drift detection.
//
// Note: Only lines starting with '#' (after leading whitespace) are removed.
// Inline comments (e.g., "replicas: 2  # desired count") are preserved because
// the '#' is not at the start of the line.
func cleanManifest(manifest string) string {
	if manifest == "" {
		return ""
	}

	var buf strings.Builder

	for line := range strings.Lines(manifest) {
		if strings.HasPrefix(line, "#") {
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	return strings.TrimSpace(buf.String())
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

	return cleanManifest(rel.Manifest) == cleanManifest(latestHelmRelease.Manifest)
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

	if release.Auth != nil {
		helmUpgrade.Username = release.Auth.Username
		helmUpgrade.Password = release.Auth.Password
	}

	chartRequested, err := getChart(release, helmUpgrade.ChartPathOptions, helmSettings, providers)
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
	helmInstall.Wait = release.Wait
	helmInstall.Timeout = release.WaitTimeout.Duration

	if release.Auth != nil {
		helmInstall.Username = release.Auth.Username
		helmInstall.Password = release.Auth.Password
	}

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
	chartRequested, err := getChart(release, client.ChartPathOptions, helmSettings, providers)
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
	helmCfg *helmaction.Configuration,
	kubeconfPath string,
	st *state.State,
) error {
	logger := st.Logger
	for _, release := range toUninstall {
		restClientGetter := newRestClientGetter(kubeconfPath, release.Namespace, st)
		if err := helmCfg.Init(restClientGetter, release.Namespace, helmStorageDriver, logger.Debugf); err != nil {
			return fail.Runtime(err, "initializing helm action configuration")
		}

		helmUninstall := helmaction.NewUninstall(helmCfg)
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
	release kubeoneapi.HelmRelease,
	chartPathOpts helmaction.ChartPathOptions,
	helmSettings *helmcli.EnvSettings,
	providers getter.Providers,
) (*chart.Chart, error) {
	chartName := release.Chart
	if release.ChartURL != "" {
		chartName = release.ChartURL
	}

	chartPath, err := chartPathOpts.LocateChart(chartName, helmSettings)
	if err != nil {
		return nil, fail.Runtime(err, "locating helm chart")
	}

	return newChart(chartPath, chartName, providers, helmSettings)
}

func newChart(chartPath string, chartName string, providers getter.Providers, helmSettings *helmcli.EnvSettings) (*chart.Chart, error) {
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
			chartRequested, err = dependencyUpdate(chartName, helmSettings, providers)
			if err != nil {
				return nil, err
			}
		}
	}

	return chartRequested, nil
}

func dependencyUpdate(chartPath string, helmSettings *helmcli.EnvSettings, providers []getter.Provider) (*chart.Chart, error) {
	mgr := &downloader.Manager{
		Out:              os.Stdout,
		ChartPath:        chartPath,
		SkipUpdate:       false,
		Getters:          providers,
		RepositoryConfig: helmSettings.RepositoryConfig,
		RepositoryCache:  helmSettings.RepositoryCache,
		Debug:            helmSettings.Debug,
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
	registryWriter := io.Discard
	if debug {
		registryWriter = os.Stdout
	}
	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(registryWriter),
	)

	return &helmaction.Configuration{
		RegistryClient: registryClient,
	}, fail.Runtime(err, "initializing new helm registry client")
}

func makeKey(rlsname string, version int) string {
	return fmt.Sprintf("%s.%s.v%d", helmStorageType, rlsname, version)
}
