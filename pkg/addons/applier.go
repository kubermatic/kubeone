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
	"io/fs"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/credentials"
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
	Config                              *kubeoneapi.KubeOneCluster
	Certificates                        map[string]string
	Credentials                         map[string]string
	CSIMigration                        bool
	CSIMigrationFeatureGates            string
	MachineControllerCredentialsEnvVars string
	InternalImages                      *internalImages
	Resources                           map[string]string
	Params                              map[string]string
}

func newAddonsApplier(s *state.State) (*applier, error) {
	var localFS fs.FS

	if s.Cluster.Addons.Enabled() {
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

	envVars, err := credentials.EnvVarBindings(s.Cluster.CloudProvider, s.CredentialsFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch env var bindings for credentials")
	}

	credsEnvVars, err := yaml.Marshal(envVars)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert env var bindings for credentials to yaml")
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
		CSIMigration:                        csiMigration,
		CSIMigrationFeatureGates:            csiMigrationFeatureGates,
		MachineControllerCredentialsEnvVars: string(credsEnvVars),
		InternalImages: &internalImages{
			pauseImage: s.PauseImage,
			resolver:   s.Images.Get,
		},
		Resources: resources.All(),
		Params:    params,
	}

	// Certs for vsphere-csi-webhook (deployed only if CSIMigration is enabled)
	if csiMigration && s.Cluster.CloudProvider.Vsphere != nil {
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
	}

	return &applier{
		TemplateData: data,
		LocalFS:      localFS,
		EmbededFS:    embeddedaddons.F,
	}, nil
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
