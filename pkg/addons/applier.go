/*
Copyright 2021 The KubeOne Authors.

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

package addons

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"
	"k8c.io/kubeone/pkg/templates/resources"

	"sigs.k8s.io/yaml"
)

const (
	webhookCertsCSI = "CSI"
)

var (
	kubectlApplyScript = heredoc.Doc(`
		sudo KUBECONFIG=/etc/kubernetes/admin.conf \
		kubectl apply -f - --prune -l "%s=%s"
	`)

	kubectlDeleteScript = heredoc.Doc(`
		sudo KUBECONFIG=/etc/kubernetes/admin.conf \
		kubectl delete -f - -l "%s=%s" --ignore-not-found=true
	`)
)

// Applier holds structure used to fetch, parse, and apply addons
type applier struct {
	TemplateData templateData
	LocalFS      fs.FS
	EmbeddedFS   fs.FS
}

// TemplateData is data available in the addons render template
type templateData struct {
	Config                                   *kubeoneapi.KubeOneCluster
	Certificates                             map[string]string
	Credentials                              map[string]string
	CredentialsCCM                           map[string]string
	CredentialsCCMHash                       string
	CCMClusterName                           string
	CSIMigration                             bool
	CSIMigrationFeatureGates                 string
	CalicoIptablesBackend                    string
	DeployCSIAddon                           bool
	MachineControllerCredentialsEnvVars      string
	MachineControllerCredentialsHash         string
	OperatingSystemManagerEnabled            bool
	OperatingSystemManagerCredentialsEnvVars string
	OperatingSystemManagerCredentialsHash    string
	RegistryCredentials                      []registryCredentialsContainer
	InternalImages                           *internalImages
	Resources                                map[string]string
	Params                                   map[string]string
}

type registryCredentialsContainer struct {
	RegistryName string
	Auth         kubeoneapi.ContainerdRegistryAuthConfig
}

func newAddonsApplier(s *state.State) (*applier, error) {
	localFS, err := addonsLocalFS(s.Cluster.Addons, s.ManifestFilePath)
	if err != nil {
		return nil, err
	}

	creds, err := credentials.Any(s.CredentialsFilePath)
	if err != nil {
		return nil, err
	}

	credsEnvVarsMC, err := mcCredentialsEnvVars(s)
	if err != nil {
		return nil, err
	}

	mcCredsHash, err := credentialsHash(s, credentials.TypeMC)
	if err != nil {
		return nil, err
	}

	kubeCAPrivateKey, kubeCACert, err := certificate.CAKeyPair(s.Configuration)
	if err != nil {
		return nil, err
	}

	// We want this to be true in two cases:
	// 	* if the CSI migration is already enabled
	//	* if we are starting the CCM/CSI migration process
	csiMigration := s.CCMMigration
	if !csiMigration && s.LiveCluster != nil && s.LiveCluster.CCMStatus != nil {
		csiMigration = s.LiveCluster.CCMStatus.CSIMigrationEnabled
	}

	// We're intentionally ignoring the error here. If the provider is not supported
	// the function will return an empty string (""), which we can easily detect in
	// the templates
	csiMigrationFeatureGates := ""
	if s.ShouldEnableCSIMigration() {
		_, csiMigrationFeatureGates, _ = s.Cluster.CSIMigrationFeatureGates(s.ShouldUnregisterInTreeCloudProvider())
	}

	// Certs for machine-controller-webhook
	mcCertsMap, err := certificate.NewSignedTLSCert(
		resources.MachineControllerWebhookName,
		resources.MachineControllerNameSpace,
		s.Cluster.ClusterNetwork.ServiceDomainName,
		kubeCAPrivateKey,
		kubeCACert,
	)
	if err != nil {
		return nil, err
	}

	// Certs for metrics-server
	msCertsMap, err := certificate.NewSignedTLSCert(
		resources.MetricsServerName,
		resources.MetricsServerNamespace,
		s.Cluster.ClusterNetwork.ServiceDomainName,
		kubeCAPrivateKey,
		kubeCACert,
	)
	if err != nil {
		return nil, err
	}

	params := map[string]string{}
	if s.Cluster.Addons.Enabled() {
		for k, v := range s.Cluster.Addons.GlobalParams {
			params[k] = v
		}
	}

	credsCCM, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeCCM)
	if err != nil {
		return nil, err
	}

	credsCCMHash, err := credentialsHash(s, credentials.TypeCCM)
	if err != nil {
		return nil, err
	}

	// Check are we deploying the CSI driver
	deployCSI := len(ensureCSIAddons(s, []addonAction{})) > 0

	calicoIptablesBackend := "Auto"
	for _, cp := range s.LiveCluster.ControlPlane {
		if cp.Config.OperatingSystem == kubeoneapi.OperatingSystemNameFlatcar || cp.Config.OperatingSystem == kubeoneapi.OperatingSystemNameRHEL {
			calicoIptablesBackend = "NFT"

			break
		}
	}

	data := templateData{
		Config: s.Cluster,
		Certificates: map[string]string{
			"MachineControllerWebhookCert": mcCertsMap[resources.TLSCertName],
			"MachineControllerWebhookKey":  mcCertsMap[resources.TLSKeyName],
			"MetricsServerCert":            msCertsMap[resources.TLSCertName],
			"MetricsServerKey":             msCertsMap[resources.TLSKeyName],
			"KubernetesCA":                 mcCertsMap[resources.KubernetesCACertName],
		},
		Credentials:                         creds,
		CredentialsCCM:                      credsCCM,
		CredentialsCCMHash:                  credsCCMHash,
		CCMClusterName:                      s.LiveCluster.CCMClusterName,
		CSIMigration:                        csiMigration,
		CSIMigrationFeatureGates:            csiMigrationFeatureGates,
		CalicoIptablesBackend:               calicoIptablesBackend,
		DeployCSIAddon:                      deployCSI,
		MachineControllerCredentialsEnvVars: string(credsEnvVarsMC),
		MachineControllerCredentialsHash:    mcCredsHash,
		OperatingSystemManagerEnabled:       s.Cluster.OperatingSystemManager.Deploy,
		RegistryCredentials:                 containerdRegistryCredentials(s.Cluster.ContainerRuntime.Containerd),
		InternalImages: &internalImages{
			pauseImage: s.PauseImage,
			resolver:   s.Images.Get,
		},
		Resources: resources.All(),
		Params:    params,
	}

	if err := csiWebhookCerts(s, &data, csiMigration, kubeCAPrivateKey, kubeCACert); err != nil {
		return nil, err
	}

	// Certs for operating-system-manager-webhook
	if s.Cluster.OperatingSystemManager.Deploy {
		if err := webhookCerts(data.Certificates,
			"OSM",
			resources.OperatingSystemManagerWebhookName,
			resources.OperatingSystemManagerNamespace,
			s.Cluster.ClusterNetwork.ServiceDomainName,
			kubeCAPrivateKey,
			kubeCACert,
		); err != nil {
			return nil, err
		}

		credsOSM, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeOSM)
		if err != nil {
			return nil, err
		}

		envVarsOSM := credentials.EnvVarBindings(credentials.SecretNameOSM, credsOSM)
		credsEnvVarsOSM, err := yaml.Marshal(envVarsOSM)
		if err != nil {
			return nil, fail.Runtime(err, "marshalling OSM credentials env variables")
		}

		osmCredsHash, err := credentialsHash(s, credentials.TypeOSM)
		if err != nil {
			return nil, err
		}

		data.OperatingSystemManagerCredentialsEnvVars = string(credsEnvVarsOSM)
		data.OperatingSystemManagerCredentialsHash = osmCredsHash
	}

	return &applier{
		TemplateData: data,
		LocalFS:      localFS,
		EmbeddedFS:   embeddedaddons.FS,
	}, nil
}

func csiWebhookCerts(s *state.State, data *templateData, csiMigration bool, kubeCAPrivateKey *rsa.PrivateKey, kubeCACert *x509.Certificate) error {
	// Certs for CSI plugins
	switch {
	case s.Cluster.CloudProvider.DigitalOcean != nil,
		s.Cluster.CloudProvider.Openstack != nil,
		s.Cluster.CloudProvider.GCE != nil:
		if err := webhookCerts(data.Certificates,
			webhookCertsCSI,
			resources.GenericCSIWebhookName,
			resources.GenericCSIWebhookNamespace,
			s.Cluster.ClusterNetwork.ServiceDomainName,
			kubeCAPrivateKey,
			kubeCACert,
		); err != nil {
			return err
		}
	// Certs for vsphere-csi-webhook (deployed only if CSIMigration is enabled)
	case s.Cluster.CloudProvider.Vsphere != nil:
		if err := webhookCerts(data.Certificates,
			webhookCertsCSI,
			resources.GenericCSIWebhookName,
			resources.VsphereCSIWebhookNamespace,
			s.Cluster.ClusterNetwork.ServiceDomainName,
			kubeCAPrivateKey,
			kubeCACert,
		); err != nil {
			return err
		}
		if csiMigration {
			if err := webhookCerts(data.Certificates,
				"CSIMigration",
				resources.VsphereCSIWebhookName,
				resources.VsphereCSIWebhookNamespace,
				s.Cluster.ClusterNetwork.ServiceDomainName,
				kubeCAPrivateKey,
				kubeCACert,
			); err != nil {
				return err
			}
		}
	case s.Cluster.CloudProvider.Nutanix != nil:
		if err := webhookCerts(data.Certificates,
			webhookCertsCSI,
			resources.NutanixCSIWebhookName,
			resources.GenericCSIWebhookNamespace,
			s.Cluster.ClusterNetwork.ServiceDomainName,
			kubeCAPrivateKey,
			kubeCACert,
		); err != nil {
			return err
		}
	}

	return nil
}

func webhookCerts(certs map[string]string, prefix, webhookName, webhookNamespace, serviceDomainName string, kubeCAPrivateKey *rsa.PrivateKey, kubeCACert *x509.Certificate) error {
	certsMap, err := certificate.NewSignedTLSCert(
		webhookName,
		webhookNamespace,
		serviceDomainName,
		kubeCAPrivateKey,
		kubeCACert,
	)
	if err != nil {
		return err
	}

	certs[fmt.Sprintf("%sWebhookCert", prefix)] = certsMap[resources.TLSCertName]
	certs[fmt.Sprintf("%sWebhookKey", prefix)] = certsMap[resources.TLSKeyName]

	return nil
}

func containerdRegistryCredentials(containerdConfig *kubeoneapi.ContainerRuntimeContainerd) []registryCredentialsContainer {
	if containerdConfig == nil {
		return nil
	}

	var (
		regCredentials []registryCredentialsContainer
		regNames       []string
	)

	for reg := range containerdConfig.Registries {
		regNames = append(regNames, reg)
	}

	sort.Strings(regNames)

	for _, reg := range regNames {
		regConfig := containerdConfig.Registries[reg]
		if regConfig.Auth != nil {
			regCredentials = append(regCredentials, registryCredentialsContainer{
				RegistryName: reg,
				Auth:         *regConfig.Auth,
			})
		}
	}

	return regCredentials
}

func mcCredentialsEnvVars(s *state.State) ([]byte, error) {
	var credsEnvVarsMC []byte

	if s.Cluster.MachineController.Deploy {
		credsMC, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeMC)
		if err != nil {
			return nil, err
		}

		envVarsMC := credentials.EnvVarBindings(credentials.SecretNameMC, credsMC)
		credsEnvVarsMC, err = yaml.Marshal(envVarsMC)
		if err != nil {
			return nil, fail.Runtime(err, "marshalling machine-controller credentials env variables")
		}
	}

	return credsEnvVarsMC, nil
}

func credentialsHash(s *state.State, credsType credentials.Type) (string, error) {
	creds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credsType)
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("kubeone-%s", s.Cluster.CloudProvider.CloudProviderName())

	keys := make([]string, 0, len(creds))
	for k := range creds {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		hash += fmt.Sprintf("%s%s", k, creds[k])
	}

	h := sha256.New()
	h.Write([]byte(hash))

	return hex.EncodeToString(h.Sum(nil)), nil
}

func addonsLocalFS(clusterAddons *kubeoneapi.Addons, manifestFilePath string) (fs.FS, error) {
	var localFS fs.FS

	if clusterAddons.Enabled() && clusterAddons.Path != "" {
		addonsPath, err := clusterAddons.RelativePath(manifestFilePath)
		if err != nil {
			return nil, err
		}

		localFS = os.DirFS(addonsPath)
	}

	return localFS, nil
}

// loadAndApplyAddon parses the addons manifests and runs kubectl apply.
func (a *applier) loadAndApplyAddon(s *state.State, fsys fs.FS, addonName string) error {
	s.Logger.Infof("Applying addon %s...", addonName)

	manifest, err := a.getManifestsFromDirectory(s, fsys, addonName)
	if err != nil {
		return err
	}

	if len(strings.TrimSpace(manifest)) == 0 {
		if len(addonName) != 0 {
			s.Logger.Warnf("Addon directory %q is empty, skipping...", addonName)
		}

		return nil
	}

	return runKubectlApply(s, manifest, addonName)
}

// loadAndApplyAddon parses the addons manifests and runs kubectl apply.
func (a *applier) loadAndDeleteAddon(s *state.State, fsys fs.FS, addonName string) error {
	s.Logger.Infof("Deleting addon %q...", addonName)

	manifest, err := a.getManifestsFromDirectory(s, fsys, addonName)
	if err != nil {
		return err
	}

	if len(strings.TrimSpace(manifest)) == 0 {
		if len(addonName) != 0 {
			s.Logger.Warnf("Addon directory %q is empty, skipping...", addonName)
		}

		return nil
	}

	return runKubectlDelete(s, manifest, addonName)
}

// runKubectlApply runs kubectl apply command
func runKubectlApply(s *state.State, manifest string, addonName string) error {
	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, conn executor.Interface) error {
		var (
			cmd            = fmt.Sprintf(kubectlApplyScript, addonLabel, addonName)
			stdin          = strings.NewReader(manifest)
			stdout, stderr strings.Builder
		)

		_, err := conn.POpen(cmd, stdin, &stdout, &stderr)
		if s.Verbose {
			fmt.Printf("+ %s\n", cmd)
			fmt.Printf("%s", stderr.String())
			fmt.Printf("%s", stdout.String())
		}

		return err
	})
}

// runKubectlDelete runs kubectl delete command
func runKubectlDelete(s *state.State, manifest string, addonName string) error {
	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, conn executor.Interface) error {
		var (
			cmd            = fmt.Sprintf(kubectlDeleteScript, addonLabel, addonName)
			stdin          = strings.NewReader(manifest)
			stdout, stderr strings.Builder
		)

		_, err := conn.POpen(cmd, stdin, &stdout, &stderr)
		if s.Verbose {
			fmt.Printf("+ %s\n", cmd)
			fmt.Printf("%s", stderr.String())
			fmt.Printf("%s", stdout.String())
		}

		return err
	})
}

type internalImages struct {
	pauseImage string
	resolver   func(images.Resource, ...images.GetOpt) string
}

func (im *internalImages) Get(imgName string) (string, error) {
	// TODO: somehow handle this the other way around
	if imgName == "PauseImage" {
		return im.pauseImage, nil
	}

	res, err := images.FindResource(imgName)
	if err != nil {
		return "", err
	}

	return im.resolver(res), nil
}
