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
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"

	"github.com/digitalocean/godo"

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
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// maxTagLength is the max length DigitalOcean accepts for tag names
	maxTagLength = 255

	dropletPageSize = 200
)

var (
	_ cloudprovider.ControlPlaneCloudProvider = &provider{}
	_ cloudprovider.LoadBalancerProvider      = &provider{}
	_ cloudprovider.CleanupPreparer           = &provider{}
	_ cloudprovider.Preflighter               = &provider{}

	// invalidTagChars matches characters not allowed in DigitalOcean tag names
	invalidTagChars = regexp.MustCompile(`[^a-zA-Z0-9:_-]`)

	// newDropletsClient builds the droplets client. Variable to allow
	// injecting fakes in tests.
	newDropletsClient = func(token string) dropletLister {
		return godo.NewFromToken(token).Droplets
	}
)

// dropletLister is the part of godo.DropletsService used to find the droplets
// of the cluster.
type dropletLister interface {
	ListByTag(ctx context.Context, tag string, opts *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

func init() {
	cloudprovider.Register(&provider{})
}

type provider struct{}

func (p *provider) Name() string { return "DigitalOcean" }

func (p *provider) Enabled(s *state.State) bool {
	return s.Cluster.CloudProvider.DigitalOcean != nil && len(s.Cluster.ControlPlane.NodeSets) > 0
}

func (p *provider) MatchesConfig(cluster *kubeoneapi.KubeOneCluster) bool {
	return cluster.CloudProvider.DigitalOcean != nil
}

func (p *provider) GenerateMachines(clusterName string, nodeSets []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	return generateControlPlaneMachines(clusterName, nodeSets, kubeletVersion)
}

func (p *provider) EnsureVM(s *state.State, nodeSet kubeoneapi.NodeSet, capimachine clusterv1alpha1.Machine) error {
	if err := exportCredentials(s); err != nil {
		return err
	}

	provMachines, err := provisioner.FindOrCreateMachines(s.Context, []clusterv1alpha1.Machine{capimachine}, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, []kubeoneapi.NodeSet{nodeSet})...)

	return nil
}

func (p *provider) LookupVMs(s *state.State) error {
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

func (p *provider) PrepareCleanup(s *state.State) error {
	return exportCredentials(s)
}

func (p *provider) Preflight(s *state.State) error {
	token, err := apiToken(s)
	if err != nil {
		return err
	}

	return checkDropletRegion(s.Context, newDropletsClient(token), s.Cluster)
}

// checkDropletRegion fails if droplets of the cluster exist outside the
// nodeSets region. DigitalOcean can't move droplets between regions, so after
// a region change apply would create a second control plane next to the old
// droplets, which reset would then no longer find. The cluster name is the
// only identity of the droplets, so this also keeps a cluster from adopting
// the droplets of a cluster with the same name in another region.
func checkDropletRegion(ctx context.Context, client dropletLister, cluster *kubeoneapi.KubeOneCluster) error {
	region, err := nodeSetsRegion(cluster.ControlPlane.NodeSets)
	if err != nil || region == "" {
		return err
	}

	tag := OwnedTag(cluster.Name)
	otherRegions := sets.New[string]()
	opts := &godo.ListOptions{Page: 1, PerPage: dropletPageSize}

	for {
		droplets, resp, err := client.ListByTag(ctx, tag, opts)
		if err != nil {
			return fail.Cloud(err, "digitalocean", "listing droplets tagged %q", tag)
		}

		for _, droplet := range droplets {
			if droplet.Region != nil && droplet.Region.Slug != region {
				otherRegions.Insert(droplet.Region.Slug)
			}
		}

		if resp == nil || resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return fail.Cloud(err, "digitalocean", "paginating droplets")
		}
		opts.Page = page + 1
	}

	if otherRegions.Len() > 0 {
		return fail.Config(
			fmt.Errorf(
				"cluster %q has droplets tagged %q in region(s) %v, but the control plane nodeSets use region %q; "+
					"DigitalOcean can't move droplets between regions, revert the nodeSets region or reset the cluster first",
				cluster.Name, tag, sets.List(otherRegions), region,
			),
			"checking control plane droplet regions",
		)
	}

	return nil
}

// nodeSetsRegion returns the region of the control plane nodeSets, validation
// ensures all nodeSets use the same one.
func nodeSetsRegion(nodeSets []kubeoneapi.NodeSet) (string, error) {
	if len(nodeSets) == 0 {
		return "", nil
	}

	var doConfig digitaloceantypes.RawConfig
	if err := jsonutil.StrictUnmarshal(nodeSets[0].CloudProviderSpec, &doConfig); err != nil {
		return "", fail.Config(err, "decode digitalocean config")
	}

	return doConfig.Region.Value, nil
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
// with "_" and truncates the tag to the maximum allowed length. Cluster names
// are DNS-1123 subdomains and never contain "_", so e.g. "a.b" and "a-b" map
// to different tags.
func sanitizeTag(tag string) string {
	tag = invalidTagChars.ReplaceAllString(tag, "_")
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

		// Copy, so the user's nodeSet labels (which become node labels) are
		// left unchanged
		machineLabels := maps.Clone(node.NodeSettings.Labels)
		if machineLabels == nil {
			machineLabels = map[string]string{}
		}
		maps.Copy(machineLabels, labels)

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

			name := node.MachineName(clusterName, idx)
			machines = append(machines, clusterv1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					// machine-controller tags the droplet with the UID and
					// finds it by name and UID in all regions. The region is
					// left out, so a changed region still finds the existing
					// droplets and Preflight can refuse the change.
					UID: types.UID(sanitizeTag(name)),
				},
				Spec: clusterv1alpha1.MachineSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Labels:      machineLabels,
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
