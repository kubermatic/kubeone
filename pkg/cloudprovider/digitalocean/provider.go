/*
Copyright 2026 The KubeOne Authors.

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

package digitalocean

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/cloudprovider"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	digitaloceantypes "k8c.io/machine-controller/sdk/cloudprovider/digitalocean"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// maxTagLength is the max length DigitalOcean accepts for tag names
	maxTagLength = 255
)

var (
	_ cloudprovider.ControlPlaneCloudProvider = &Provider{}
	_ cloudprovider.LoadBalancerProvider      = &Provider{}
	_ cloudprovider.CleanupPreparer           = &Provider{}

	// invalidTagChars matches characters not allowed in DigitalOcean tag names
	invalidTagChars = regexp.MustCompile(`[^a-zA-Z0-9:_-]`)
)

func init() {
	cloudprovider.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() string { return "DigitalOcean" }

func (p *Provider) Enabled(s *state.State) bool {
	return s.Cluster.CloudProvider.DigitalOcean != nil && len(s.Cluster.ControlPlane.NodeSets) > 0
}

func (p *Provider) MatchesConfig(cluster *kubeoneapi.KubeOneCluster) bool {
	return cluster.CloudProvider.DigitalOcean != nil
}

func (p *Provider) GenerateMachines(clusterName string, nodeSets []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	return generateControlPlaneMachines(clusterName, nodeSets, kubeletVersion)
}

func (p *Provider) EnsureVM(s *state.State, capimachine clusterv1alpha1.Machine) error {
	if err := exportCredentials(s); err != nil {
		return err
	}

	provMachines, err := provisioner.FindOrCreateMachines(s.Context, []clusterv1alpha1.Machine{capimachine}, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

func (p *Provider) LookupVMs(s *state.State) error {
	if err := exportCredentials(s); err != nil {
		return err
	}

	capimachines, err := p.GenerateMachines(
		s.Cluster.Name,
		s.Cluster.ControlPlane.NodeSets,
		s.Cluster.Versions.Kubernetes,
	)
	if err != nil {
		return err
	}

	provMachines, err := provisioner.FindMachines(s.Context, capimachines, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

func (p *Provider) PrepareCleanup(s *state.State) error {
	return exportCredentials(s)
}

// exportCredentials exports the DigitalOcean token for the in-process
// machine-controller driver, which reads it from the DO_TOKEN environment
// variable.
func exportCredentials(s *state.State) error {
	token, err := apiToken(s)
	if err != nil {
		return err
	}

	if err := os.Setenv(credentials.DigitalOceanTokenKeyMC, token); err != nil {
		return fail.Runtime(err, "setting %s", credentials.DigitalOceanTokenKeyMC)
	}

	return nil
}

func apiToken(s *state.State) (string, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return "", err
	}

	return providerCreds[credentials.DigitalOceanTokenKeyMC], nil
}

// sanitizeTag replaces characters that are not allowed in DigitalOcean tags
// and truncates the tag to the maximum allowed length.
func sanitizeTag(tag string) string {
	tag = invalidTagChars.ReplaceAllString(tag, "-")
	if len(tag) > maxTagLength {
		tag = tag[:maxTagLength]
	}

	return tag
}

// ClusterTag is the tag applied to all control plane droplets of the cluster,
// it's used by the load balancer to select its backends.
func ClusterTag(clusterName string) string {
	return sanitizeTag("kubeone-cluster:" + clusterName)
}

// OwnedTag marks droplets created and owned by KubeOne for the cluster.
// DigitalOcean doesn't persist tags on load balancers, see ownedByKubeOne.
func OwnedTag(clusterName string) string {
	return sanitizeTag("kubeone-owned:" + clusterName)
}

// labelTags converts labels into DigitalOcean tags in the "key:value" form,
// DigitalOcean tags have no separate value. The order is deterministic.
func labelTags(labels map[string]string) []string {
	tags := make([]string, 0, len(labels))
	for _, key := range slices.Sorted(maps.Keys(labels)) {
		tags = append(tags, sanitizeTag(key+":"+labels[key]))
	}

	return tags
}

func generateControlPlaneMachines(clusterName string, nodeSets []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSets {
		labels := map[string]string{
			"kubeone_cluster_name": clusterName,
			"kubeone_role":         "control-plane",
		}

		if node.NodeSettings.Labels == nil {
			node.NodeSettings.Labels = map[string]string{}
		}
		maps.Copy(node.NodeSettings.Labels, labels)

		osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
		if err != nil {
			return nil, fail.Config(err, "marshaling operating system spec")
		}

		for idx := range node.Replicas {
			var doConfig digitaloceantypes.RawConfig
			if err = jsonutil.StrictUnmarshal(node.CloudProviderSpec, &doConfig); err != nil {
				return nil, fail.Config(err, "decode digitalocean config")
			}

			// Private networking is required for CP-to-CP traffic and the load balancer
			if doConfig.PrivateNetworking.Value == nil {
				doConfig.PrivateNetworking.Value = new(true)
			}

			// Same kubeone_* keys as the other providers put on their VMs
			doConfig.Tags = appendMissingTags(doConfig.Tags, append([]string{ClusterTag(clusterName), OwnedTag(clusterName)}, labelTags(labels)...)...)

			doSpec, err := json.Marshal(doConfig)
			if err != nil {
				return nil, fail.Config(err, "marshaling cloud provider spec")
			}

			providerConfig := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderDigitalocean,
				CloudProviderSpec: runtime.RawExtension{
					Raw: doSpec,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerConfig)
			if err != nil {
				return nil, fail.Cloud(err, "digitalocean", "json marshaling provider config")
			}

			name := fmt.Sprintf("%s-%s-%d", clusterName, node.Name, idx)
			machines = append(machines, clusterv1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					// machine-controller tags the droplet with the UID, so it
					// must be a valid DigitalOcean tag
					UID: types.UID(sanitizeTag(name)),
				},
				Spec: clusterv1alpha1.MachineSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Labels:      node.NodeSettings.Labels,
						Annotations: node.NodeSettings.Annotations,
					},
					Taints: node.NodeSettings.Taints,
					Versions: clusterv1alpha1.MachineVersionInfo{
						Kubelet: kubeletVersion,
					},
					ProviderSpec: clusterv1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: providerSpecRaw,
						},
					},
				},
			})
		}
	}

	return machines, nil
}

func appendMissingTags(tags []providerconfig.ConfigVarString, newTags ...string) []providerconfig.ConfigVarString {
	existing := map[string]bool{}
	for _, t := range tags {
		existing[t.Value] = true
	}

	for _, t := range newTags {
		if !existing[t] {
			tags = append(tags, providerconfig.ConfigVarString{Value: t})
			existing[t] = true
		}
	}

	return tags
}
