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
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"
	"k8c.io/kubeone/pkg/templates/resources"

	"sigs.k8s.io/yaml"
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
	EmbededFS    fs.FS
}

// TemplateData is data available in the addons render template
type templateData struct {
	Config                                   *kubeoneapi.KubeOneCluster
	Certificates                             map[string]string
	Credentials                              map[string]string
	CredentialsCCM                           map[string]string
	CCMClusterName                           string
	CSIMigration                             bool
	CSIMigrationFeatureGates                 string
	MachineControllerCredentialsEnvVars      string
	OperatingSystemManagerEnabled            bool
	OperatingSystemManagerCredentialsEnvVars string
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
	var localFS fs.FS

	if s.Cluster.Addons.Enabled() && s.Cluster.Addons.Path != "" {
		addonsPath, err := s.Cluster.Addons.RelativePath(s.ManifestFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get addons path")
		}

		localFS = os.DirFS(addonsPath)
	}

	creds, err := credentials.Any(s.CredentialsFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch credentials")
	}

	credsCCM, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeCCM)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch cloud provider credentials")
	}

	envVarsMC, err := credentials.EnvVarBindings(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.SecretNameMC, credentials.TypeMC)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch env var bindings for credentials")
	}

	credsEnvVarsMC, err := yaml.Marshal(envVarsMC)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert env var bindings for credentials to yaml")
	}

	var credsEnvVarsOSM []byte
	if s.Cluster.OperatingSystemManagerEnabled() {
		var envVarsOSM []corev1.EnvVar
		envVarsOSM, err = credentials.EnvVarBindings(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.SecretNameOSM, credentials.TypeOSM)
		if err != nil {
			return nil, errors.Wrap(err, "unable to fetch env var bindings for credentials")
		}

		credsEnvVarsOSM, err = yaml.Marshal(envVarsOSM)
		if err != nil {
			return nil, errors.Wrap(err, "unable to convert env var bindings for credentials to yaml")
		}
	}

	kubeCAPrivateKey, kubeCACert, err := certificate.CAKeyPair(s.Configuration)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load CA keypair")
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

	regCredentials := []registryCredentialsContainer{}

	if s.Cluster.ContainerRuntime.Containerd != nil {
		regNames := []string{}

		for reg := range s.Cluster.ContainerRuntime.Containerd.Registries {
			regNames = append(regNames, reg)
		}

		sort.Strings(regNames)

		for _, reg := range regNames {
			regConfig := s.Cluster.ContainerRuntime.Containerd.Registries[reg]
			if regConfig.Auth != nil {
				regCredentials = append(regCredentials, registryCredentialsContainer{
					RegistryName: reg,
					Auth:         *regConfig.Auth,
				})
			}
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
		CCMClusterName:                      s.LiveCluster.CCMClusterName,
		CSIMigration:                        csiMigration,
		CSIMigrationFeatureGates:            csiMigrationFeatureGates,
		MachineControllerCredentialsEnvVars: string(credsEnvVarsMC),
		OperatingSystemManagerEnabled:       s.Cluster.OperatingSystemManagerEnabled(),
		RegistryCredentials:                 regCredentials,
		InternalImages: &internalImages{
			pauseImage: s.PauseImage,
			resolver:   s.Images.Get,
		},
		Resources: resources.All(),
		Params:    params,
	}

	// Certs for CSI plugins
	switch {
	// Certs for vsphere-csi-webhook (deployed only if CSIMigration is enabled)
	case csiMigration && s.Cluster.CloudProvider.Vsphere != nil:
		vsphereCSICertsMap, err := certificate.NewSignedTLSCert(
			resources.VsphereCSIWebhookName,
			resources.VsphereCSIWebhookNamespace,
			s.Cluster.ClusterNetwork.ServiceDomainName,
			kubeCAPrivateKey,
			kubeCACert,
		)
		if err != nil {
			return nil, err
		}
		data.Certificates["vSphereCSIWebhookCert"] = vsphereCSICertsMap[resources.TLSCertName]
		data.Certificates["vSphereCSIWebhookKey"] = vsphereCSICertsMap[resources.TLSKeyName]
	case s.Cluster.CloudProvider.Nutanix != nil:
		nutanixCSICertsMap, err := certificate.NewSignedTLSCert(
			resources.NutanixCSIWebhookName,
			resources.NutanixCSIWebhookNamespace,
			s.Cluster.ClusterNetwork.ServiceDomainName,
			kubeCAPrivateKey,
			kubeCACert,
		)
		if err != nil {
			return nil, err
		}
		data.Certificates["NutanixCSIWebhookCert"] = nutanixCSICertsMap[resources.TLSCertName]
		data.Certificates["NutanixCSIWebhookKey"] = nutanixCSICertsMap[resources.TLSKeyName]
	case s.Cluster.CloudProvider.DigitalOcean != nil && s.Cluster.CloudProvider.External:
		digitaloceanCSICertsMap, err := certificate.NewSignedTLSCert(
			resources.DigitalOceanCSIWebhookName,
			resources.DigitalOceanCSIWebhookNamespace,
			s.Cluster.ClusterNetwork.ServiceDomainName,
			kubeCAPrivateKey,
			kubeCACert,
		)
		if err != nil {
			return nil, err
		}
		data.Certificates["DigitalOceanCSIWebhookCert"] = digitaloceanCSICertsMap[resources.TLSCertName]
		data.Certificates["DigitalOceanCSIWebhookKey"] = digitaloceanCSICertsMap[resources.TLSKeyName]
	}

	// Certs for operating-system-manager-webhook
	if s.Cluster.OperatingSystemManagerEnabled() {
		osmCertsMap, err := certificate.NewSignedTLSCert(
			resources.OperatingSystemManagerWebhookName,
			resources.OperatingSystemManagerNamespace,
			s.Cluster.ClusterNetwork.ServiceDomainName,
			kubeCAPrivateKey,
			kubeCACert,
		)
		if err != nil {
			return nil, err
		}
		data.Certificates["OperatingSystemManagerWebhookCert"] = osmCertsMap[resources.TLSCertName]
		data.Certificates["OperatingSystemManagerWebhookKey"] = osmCertsMap[resources.TLSKeyName]
		data.OperatingSystemManagerCredentialsEnvVars = string(credsEnvVarsOSM)
	}

	return &applier{
		TemplateData: data,
		LocalFS:      localFS,
		EmbededFS:    embeddedaddons.FS,
	}, nil
}

// loadAndApplyAddon parses the addons manifests and runs kubectl apply.
func (a *applier) loadAndApplyAddon(s *state.State, fsys fs.FS, addonName string) error {
	s.Logger.Infof("Applying addon %s...", addonName)

	manifest, err := a.getManifestsFromDirectory(s, fsys, addonName)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(strings.TrimSpace(manifest)) == 0 {
		if len(addonName) != 0 {
			s.Logger.Warnf("Addon directory %q is empty, skipping...", addonName)
		}

		return nil
	}

	return errors.Wrap(
		runKubectlApply(s, manifest, addonName),
		"failed to apply addons",
	)
}

// loadAndApplyAddon parses the addons manifests and runs kubectl apply.
func (a *applier) loadAndDeleteAddon(s *state.State, fsys fs.FS, addonName string) error {
	s.Logger.Infof("Deleting addon %q...", addonName)

	manifest, err := a.getManifestsFromDirectory(s, fsys, addonName)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(strings.TrimSpace(manifest)) == 0 {
		if len(addonName) != 0 {
			s.Logger.Warnf("Addon directory %q is empty, skipping...", addonName)
		}

		return nil
	}

	return errors.Wrap(
		runKubectlDelete(s, manifest, addonName),
		"failed to apply addons",
	)
}

// runKubectlApply runs kubectl apply command
func runKubectlApply(s *state.State, manifest string, addonName string) error {
	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
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
	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
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
