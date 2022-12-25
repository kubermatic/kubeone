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

package initcmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/Masterminds/sprig/v3"
	"github.com/iancoleman/orderedmap"

	"k8c.io/kubeone/examples"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/pointer"
	"k8c.io/kubeone/pkg/tabwriter"

	kyaml "sigs.k8s.io/yaml"
)

type GenerateOpts struct {
	validProviders map[string]initProvider

	path string

	providerName      string
	clusterName       string
	kubernetesVersion string

	generateTerraform bool
	terraformVars     map[string]string

	cni string

	enableFeatureEncryption bool
	enableFeatureCoreDNSPDB bool

	enableAddonAutoscaler bool
	enableAddonBackups    bool

	addonBackupsPassword         string
	addonBackupsS3Bucket         string
	addonBackupsDefaultAWSRegion string
}

func NewGenerateOpts(path, providerName, clusterName, kubernetesVersion string, generateTerraform bool) *GenerateOpts {
	// Ensure non-interactive init will populate terraform.tfvars with stubs for required variables
	requiredTFVars := map[string]string{}
	prov := ValidProviders[providerName]

	for _, param := range prov.requiredTFVars {
		requiredTFVars[param.Name] = ""
	}

	return &GenerateOpts{
		validProviders:          ValidProviders,
		path:                    path,
		providerName:            providerName,
		clusterName:             clusterName,
		kubernetesVersion:       kubernetesVersion,
		cni:                     cniCanalValue,
		enableFeatureEncryption: false,
		enableFeatureCoreDNSPDB: false,
		enableAddonAutoscaler:   false,
		enableAddonBackups:      false,
		generateTerraform:       generateTerraform,
		terraformVars:           requiredTFVars,
	}
}

func GenerateConfigs(opts *GenerateOpts) error {
	if opts.clusterName == "" {
		return fmt.Errorf("cluster name is required")
	}

	ybuf, err := genKubeOneClusterYAML(opts)
	if err != nil {
		return fail.Runtime(err, "generating KubeOneCluster")
	}

	// special case to generate JUST yaml and no terraform
	if opts.path == "-" && !opts.generateTerraform {
		_, err = fmt.Printf("%s", ybuf)

		return err
	}

	err = os.MkdirAll(opts.path, 0750)
	if err != nil {
		return err
	}

	k1config, err := os.Create(filepath.Join(opts.path, "kubeone.yaml"))
	if err != nil {
		return fail.Runtime(err, "creating manifest file")
	}
	defer k1config.Close()

	_, err = io.Copy(k1config, bytes.NewBuffer(ybuf))
	if err != nil {
		return fail.Runtime(err, "writing KubeOneCluster")
	}

	prov := ValidProviders[opts.providerName]
	if err := createTerraformVars(opts, prov); err != nil {
		return err
	}

	return nil
}

var (
	manifestTemplateSource = heredoc.Doc(`
		apiVersion: {{ .APIVersion }}
		kind: {{ .Kind }}
		{{- with .Name}}
		name: {{ . }}
		{{- end }}

		versions:
		  kubernetes: {{ .Versions.Kubernetes }}

		cloudProvider:
		  {{ .CloudProvider.Name }}: {}
		{{- with .CloudProvider.External }}
		  external: true
		{{ end -}}
		{{- with .CloudProvider.CloudConfig }}
		  cloudConfig: |
		{{ . | indent 4 -}}
		{{ end -}}
		{{- with .CloudProvider.CSIConfig }}
		  csiConfig: |
		{{ . | indent 4 -}}
		{{ end }}
		containerRuntime:
		  containerd: {}
		
		{{  if .ClusterNetwork.CNI -}}
		clusterNetwork:
		  cni:
		  {{- with .ClusterNetwork.CNI.Canal }}
		    canal: {}
		  {{ end -}}
		  {{- with .ClusterNetwork.CNI.Cilium }}
		    cilium:
		      enableHubble: {{ .EnableHubble }}
		      kubeProxyReplacement: {{ .KubeProxyReplacement }}
		  {{- end -}}
		  {{- with .ClusterNetwork.CNI.External }}
		    external: {}
		  {{- end -}}
		  {{- with .ClusterNetwork.KubeProxy }}
		  kubeProxy:
		    skipInstallation: {{ .SkipInstallation }}
          {{- end -}}
		{{- end }}

		{{- if or .Features.EncryptionProviders .Features.CoreDNS }}
		features:
		{{- with .Features.EncryptionProviders }}
		  encryptionProviders:
		    enable: {{ .Enable }}
		{{- end -}}
		{{- with .Features.CoreDNS }}
		  coreDNS:
		    deployPodDisruptionBudget: {{ .DeployPodDisruptionBudget }}
		{{ end -}}
		{{ end -}}

		{{ with .MachineController }}
		machineController:
		  deploy: false
		{{ end -}}

		{{- with .OperatingSystemManager }}
		operatingSystemManager:
		  deploy: false
		{{ end }}

		{{- with .Addons }}
		addons:
		  enable: true
		  addons:
		{{- range .Addons }}
		    - name: {{ .Name -}}
		      {{- with .Params }}
		      params:
		      {{- range $key, $value := . }}
		        {{ $key }}: {{ $value -}}
		      {{- end -}}
		      {{- end -}}
		{{- end }}
		{{- end }}
	`)

	manifestTemplate = template.Must(
		template.New("manifest").Funcs(sprig.TxtFuncMap()).
			Parse(manifestTemplateSource),
	)
)

func genKubeOneClusterYAML(params *GenerateOpts) ([]byte, error) {
	prov := ValidProviders[params.providerName]
	clusterName := params.clusterName

	if params.generateTerraform && params.providerName != "none" {
		clusterName = ""
	}

	cluster := kubeonev1beta2.KubeOneCluster{
		TypeMeta: kubeonev1beta2.NewKubeOneCluster().TypeMeta,
		Name:     clusterName,
		CloudProvider: kubeonev1beta2.CloudProviderSpec{
			External:    prov.external,
			CloudConfig: prov.cloudConfig,
			CSIConfig:   prov.csiConfig,
		},
		ContainerRuntime: kubeonev1beta2.ContainerRuntimeConfig{
			Containerd: &kubeonev1beta2.ContainerRuntimeContainerd{},
		},
		Versions: kubeonev1beta2.VersionConfig{
			Kubernetes: params.kubernetesVersion,
		},
		Addons: &kubeonev1beta2.Addons{
			Enable: true,
			Addons: []kubeonev1beta2.Addon{
				{
					Name: "default-storage-class",
				},
			},
		},
	}

	providerName := prov.alternativeName
	if providerName == "" {
		providerName = params.providerName
	}

	err := kubeonev1beta2.SetCloudProvider(&cluster.CloudProvider, providerName)
	if err != nil {
		return nil, err
	}

	if cluster.CloudProvider.None != nil {
		cluster.Addons = nil
		cluster.MachineController = &kubeonev1beta2.MachineControllerConfig{
			Deploy: false,
		}
		cluster.OperatingSystemManager = &kubeonev1beta2.OperatingSystemManagerConfig{
			Deploy: false,
		}
	}

	clusterAdditionalParams(&cluster, params)

	var buf bytes.Buffer
	err = manifestTemplate.Execute(&buf, &cluster)
	if err != nil {
		return nil, fail.Runtime(err, "generating kubeone manifest")
	}

	dummy := kubeonev1beta2.NewKubeOneCluster()
	if err = kyaml.UnmarshalStrict(buf.Bytes(), &dummy); err != nil {
		return nil, fail.Runtime(err, "kubeone manifest testing marshal/unmarshal")
	}

	return buf.Bytes(), err
}

func clusterAdditionalParams(cluster *kubeonev1beta2.KubeOneCluster, generateOpts *GenerateOpts) {
	// CNI
	switch generateOpts.cni {
	case cniCanalValue:
		cluster.ClusterNetwork.CNI = &kubeonev1beta2.CNI{
			Canal: &kubeonev1beta2.CanalSpec{},
		}
	case cniCiliumValue:
		cluster.ClusterNetwork.CNI = &kubeonev1beta2.CNI{
			Cilium: &kubeonev1beta2.CiliumSpec{
				EnableHubble:         true,
				KubeProxyReplacement: kubeonev1beta2.KubeProxyReplacementDisabled,
			},
		}
	case cniCiliumReplacementValue:
		cluster.ClusterNetwork.CNI = &kubeonev1beta2.CNI{
			Cilium: &kubeonev1beta2.CiliumSpec{
				EnableHubble:         true,
				KubeProxyReplacement: kubeonev1beta2.KubeProxyReplacementStrict,
			},
		}
		cluster.ClusterNetwork.KubeProxy = &kubeonev1beta2.KubeProxyConfig{
			SkipInstallation: true,
		}
	case cniExternalValue:
		cluster.ClusterNetwork.CNI = &kubeonev1beta2.CNI{
			External: &kubeonev1beta2.ExternalCNISpec{},
		}
	}

	// Features
	if generateOpts.enableFeatureEncryption {
		cluster.Features.EncryptionProviders = &kubeonev1beta2.EncryptionProviders{
			Enable: true,
		}
	}
	if generateOpts.enableFeatureCoreDNSPDB {
		cluster.Features.CoreDNS = &kubeonev1beta2.CoreDNS{
			DeployPodDisruptionBudget: pointer.New(true),
		}
	}

	// Addons
	if (generateOpts.enableAddonAutoscaler || generateOpts.enableAddonBackups) && cluster.Addons == nil {
		if cluster.Addons == nil {
			cluster.Addons = &kubeonev1beta2.Addons{
				Enable: true,
				Addons: []kubeonev1beta2.Addon{},
			}
		}
	}
	if generateOpts.enableAddonAutoscaler {
		cluster.Addons.Addons = append(cluster.Addons.Addons, kubeonev1beta2.Addon{
			Name: "cluster-autoscaler",
		})
	}
	if generateOpts.enableAddonBackups {
		cluster.Addons.Addons = append(cluster.Addons.Addons, kubeonev1beta2.Addon{
			Name: "backups-restic",
			Params: map[string]string{
				"resticPassword":   generateOpts.addonBackupsPassword,
				"s3Bucket":         generateOpts.addonBackupsS3Bucket,
				"awsDefaultRegion": generateOpts.addonBackupsDefaultAWSRegion,
			},
		})
	}
}

func createTerraformVars(opts *GenerateOpts, prov initProvider) error {
	if !opts.generateTerraform || prov.terraformPath == "" {
		return nil
	}

	if err := examples.CopyTo(opts.path, prov.terraformPath); err != nil {
		return fail.Runtime(err, "copying terraform configuration")
	}

	tfvars, err := os.Create(filepath.Join(opts.path, "terraform.tfvars"))
	if err != nil {
		return err
	}
	defer tfvars.Close()

	buf := genTerraformVars(opts)
	if buf == nil {
		return fail.Runtime(fmt.Errorf("terraform vars buffer is empty"), "generating terraform configuration")
	}

	fmt.Fprint(tfvars, buf.String())

	return nil
}

func genTerraformVars(opts *GenerateOpts) *bytes.Buffer {
	tfVars := bytes.NewBuffer([]byte{})

	fmt.Fprintf(tfVars, "cluster_name = %q\n\n", opts.clusterName)

	omap := orderedmap.New()
	for k, v := range opts.terraformVars {
		omap.Set(k, v)
	}
	omap.SortKeys(sort.Strings)

	tab := tabwriter.NewWithPadding(tfVars, 1)
	defer tab.Flush()

	for _, k := range omap.Keys() {
		v, _ := omap.Get(k)
		fmt.Fprintf(tab, "%s\t= %q\n", k, v)
	}

	return tfVars
}
