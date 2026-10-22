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
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	digitaloceantypes "k8c.io/machine-controller/sdk/cloudprovider/digitalocean"
	"k8c.io/machine-controller/sdk/providerconfig"
)

const validSpec = `{"region":"fra1","size":"s-2vcpu-4gb","tags":["team-a"]}`

func testNodeSets(spec string, replicas int) []kubeoneapi.NodeSet {
	return []kubeoneapi.NodeSet{
		{
			Name:              "cp",
			Replicas:          replicas,
			OperatingSystem:   kubeoneapi.OperatingSystemNameUbuntu,
			SSH:               kubeoneapi.SSHSpec{PublicKeys: []string{"ssh-ed25519 AAAA test"}},
			CloudProviderSpec: json.RawMessage(spec),
		},
	}
}

func decodeMachine(t *testing.T, m clusterv1alpha1.Machine) (providerconfig.Config, digitaloceantypes.RawConfig) {
	t.Helper()

	var pc providerconfig.Config
	if err := json.Unmarshal(m.Spec.ProviderSpec.Value.Raw, &pc); err != nil {
		t.Fatalf("decoding providerSpec: %v", err)
	}

	var raw digitaloceantypes.RawConfig
	if err := json.Unmarshal(pc.CloudProviderSpec.Raw, &raw); err != nil {
		t.Fatalf("decoding cloudProviderSpec: %v", err)
	}

	return pc, raw
}

func tagValues(tags []providerconfig.ConfigVarString) []string {
	var out []string
	for _, t := range tags {
		out = append(out, t.Value)
	}

	return out
}

func TestGenerateMachines(t *testing.T) {
	machines, err := generateControlPlaneMachines("demo", testNodeSets(validSpec, 3), "v1.36.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(machines) != 3 {
		t.Fatalf("expected 3 machines, got %d", len(machines))
	}

	for i, m := range machines {
		wantName := "demo-cp-" + string(rune('0'+i))
		if m.Name != wantName || m.Spec.Name != wantName {
			t.Errorf("machine %d: expected name %q, got %q/%q", i, wantName, m.Name, m.Spec.Name)
		}
		if want := wantName; string(m.UID) != want {
			t.Errorf("machine %d: expected UID %q, got %q", i, want, m.UID)
		}
		if m.Spec.Versions.Kubelet != "v1.36.0" {
			t.Errorf("machine %d: unexpected kubelet version %q", i, m.Spec.Versions.Kubelet)
		}

		pc, raw := decodeMachine(t, m)
		if pc.CloudProvider != providerconfig.CloudProviderDigitalocean {
			t.Errorf("machine %d: unexpected cloud provider %q", i, pc.CloudProvider)
		}
		if pc.OperatingSystem != providerconfig.OperatingSystemUbuntu {
			t.Errorf("machine %d: unexpected operating system %q", i, pc.OperatingSystem)
		}
		if !reflect.DeepEqual(pc.SSHPublicKeys, []string{"ssh-ed25519 AAAA test"}) {
			t.Errorf("machine %d: unexpected ssh keys %v", i, pc.SSHPublicKeys)
		}
		if raw.Region.Value != "fra1" || raw.Size.Value != "s-2vcpu-4gb" {
			t.Errorf("machine %d: unexpected region/size %q/%q", i, raw.Region.Value, raw.Size.Value)
		}
		if raw.PrivateNetworking.Value == nil || !*raw.PrivateNetworking.Value {
			t.Errorf("machine %d: private networking must default to true", i)
		}
		tags := tagValues(raw.Tags)
		for _, want := range []string{"team-a", "kubeone-cluster:demo", "kubeone-owned:demo", "kubeone_cluster_name:demo", "kubeone_role:control-plane"} {
			if !slices.Contains(tags, want) {
				t.Errorf("machine %d: expected tag %q in %v", i, want, tags)
			}
		}
		if raw.Token.Value != "" {
			t.Errorf("machine %d: token must not be set in the providerSpec", i)
		}
		if m.Spec.Labels["kubeone_cluster_name"] != "demo" || m.Spec.Labels["kubeone_role"] != "control-plane" {
			t.Errorf("machine %d: unexpected labels %v", i, m.Spec.Labels)
		}
	}
}

// CleanupVMs regenerates the machines to find what to delete, so the output
// must be deterministic.
func TestGenerateMachinesDeterministic(t *testing.T) {
	first, err := generateControlPlaneMachines("demo", testNodeSets(validSpec, 3), "v1.36.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := generateControlPlaneMachines("demo", testNodeSets(validSpec, 3), "v1.36.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatal("expected GenerateMachines to be deterministic")
	}
}

func TestGenerateMachinesKeepsExplicitValues(t *testing.T) {
	spec := `{"region":"fra1","size":"s-2vcpu-4gb","private_networking":true,"tags":["kubeone-cluster:demo"]}`

	machines, err := generateControlPlaneMachines("demo", testNodeSets(spec, 1), "v1.36.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, raw := decodeMachine(t, machines[0])
	tags := tagValues(raw.Tags)
	count := 0
	for _, tag := range tags {
		if tag == "kubeone-cluster:demo" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected cluster tag exactly once, got tags %v", tags)
	}
}

func TestGenerateMachinesRejectsBadSpec(t *testing.T) {
	tests := []struct {
		name string
		spec string
	}{
		{name: "unknown field", spec: `{"region":"fra1","size":"s-1vcpu-1gb","image":"ubuntu"}`},
		{name: "invalid json", spec: `{"region":`},
		{name: "wrong type", spec: `{"region":"fra1","size":"s-1vcpu-1gb","tags":"a"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := generateControlPlaneMachines("demo", testNodeSets(tc.spec, 1), "v1.36.0"); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestSanitizeTag(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "kubeone-cluster:demo", want: "kubeone-cluster:demo"},
		{in: "kubeone-cluster:my.cluster", want: "kubeone-cluster:my_cluster"},
		{in: "kubeone-cluster:my-cluster", want: "kubeone-cluster:my-cluster"},
		{in: "Demo_Cluster 1", want: "Demo_Cluster_1"},
		{in: strings.Repeat("a", 300), want: strings.Repeat("a", maxTagLength)},
	}

	for _, tc := range tests {
		if got := sanitizeTag(tc.in); got != tc.want {
			t.Errorf("sanitizeTag(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestGenerateMachinesSanitizesUID(t *testing.T) {
	machines, err := generateControlPlaneMachines("my.cluster", testNodeSets(validSpec, 1), "v1.36.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if machines[0].Name != "my.cluster-cp-0" {
		t.Errorf("unexpected name %q", machines[0].Name)
	}
	if string(machines[0].UID) != "my_cluster-cp-0" {
		t.Errorf("expected sanitized UID, got %q", machines[0].UID)
	}
}

func testState(lb *kubeoneapi.DigitalOceanLoadBalancer, nodeSets []kubeoneapi.NodeSet) *state.State {
	cluster := &kubeoneapi.KubeOneCluster{
		Name: "demo",
		CloudProvider: kubeoneapi.CloudProviderSpec{
			DigitalOcean: &kubeoneapi.DigitalOceanSpec{},
		},
		ControlPlane: kubeoneapi.ControlPlaneConfig{NodeSets: nodeSets},
	}
	if lb != nil {
		cluster.CloudProvider.DigitalOcean.ControlPlane = &kubeoneapi.DigitalOceanControlPlane{LoadBalancer: *lb}
	}

	return &state.State{Cluster: cluster, Logger: logrus.New()}
}

func TestEnabledAndHasLoadBalancer(t *testing.T) {
	p := &provider{}

	tests := []struct {
		name        string
		lb          *kubeoneapi.DigitalOceanLoadBalancer
		nodeSets    []kubeoneapi.NodeSet
		wantEnabled bool
		wantLB      bool
	}{
		{name: "no nodeSets", lb: &kubeoneapi.DigitalOceanLoadBalancer{}, wantEnabled: false, wantLB: false},
		{name: "create mode", lb: &kubeoneapi.DigitalOceanLoadBalancer{Name: "demo-kubeapi"}, nodeSets: testNodeSets(validSpec, 1), wantEnabled: true, wantLB: true},
		{name: "existing mode", lb: &kubeoneapi.DigitalOceanLoadBalancer{ID: "abc"}, nodeSets: testNodeSets(validSpec, 1), wantEnabled: true, wantLB: true},
		{name: "nil controlPlane", lb: nil, nodeSets: testNodeSets(validSpec, 1), wantEnabled: true, wantLB: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := testState(tc.lb, tc.nodeSets)
			if got := p.Enabled(s); got != tc.wantEnabled {
				t.Errorf("Enabled() = %v, want %v", got, tc.wantEnabled)
			}
			if got := p.HasLoadBalancer(s); got != tc.wantLB {
				t.Errorf("HasLoadBalancer() = %v, want %v", got, tc.wantLB)
			}
		})
	}
}

func TestGenerateMachinesKeepsNodeSetLabels(t *testing.T) {
	nodeSets := testNodeSets(validSpec, 1)
	nodeSets[0].NodeSettings.Labels = map[string]string{"team": "a"}

	machines, err := generateControlPlaneMachines("demo", nodeSets, "v1.36.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(nodeSets[0].NodeSettings.Labels, map[string]string{"team": "a"}) {
		t.Errorf("nodeSet labels must not change, got %v", nodeSets[0].NodeSettings.Labels)
	}
	if got := machines[0].Spec.Labels; got["team"] != "a" || got["kubeone_role"] != "control-plane" {
		t.Errorf("unexpected machine labels %v", got)
	}
}

// machine-controller finds droplets by name and UID tag across all regions,
// the UID must not depend on the region, or droplets are lost when the region
// changes.
func TestGenerateMachinesUIDIgnoresRegion(t *testing.T) {
	fra1, err := generateControlPlaneMachines("demo", testNodeSets(`{"region":"fra1","size":"s-2vcpu-4gb"}`, 1), "v1.36.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ams3, err := generateControlPlaneMachines("demo", testNodeSets(`{"region":"ams3","size":"s-2vcpu-4gb"}`, 1), "v1.36.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fra1[0].Name != ams3[0].Name || fra1[0].UID != ams3[0].UID {
		t.Errorf("expected the same name and UID, got %q/%q and %q/%q", fra1[0].Name, fra1[0].UID, ams3[0].Name, ams3[0].UID)
	}
}

// fakeDropletLister returns the given droplets for the tag, one slice per
// page.
type fakeDropletLister struct {
	pages [][]godo.Droplet
	err   error

	tags []string
}

func (f *fakeDropletLister) ListByTag(_ context.Context, tag string, opts *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	f.tags = append(f.tags, tag)
	if f.err != nil {
		return nil, nil, f.err
	}
	if len(f.pages) == 0 {
		return nil, &godo.Response{Links: &godo.Links{}}, nil
	}

	page := 1
	if opts != nil && opts.Page > 0 {
		page = opts.Page
	}

	pages := &godo.Pages{}
	if page > 1 {
		pages.Prev = "https://api.digitalocean.com/v2/droplets?page=" + strconv.Itoa(page-1)
	}
	if page < len(f.pages) {
		pages.Next = "https://api.digitalocean.com/v2/droplets?page=" + strconv.Itoa(page+1)
	}

	return f.pages[page-1], &godo.Response{Links: &godo.Links{Pages: pages}}, nil
}

func droplet(name, region string) godo.Droplet {
	return godo.Droplet{Name: name, Region: &godo.Region{Slug: region}}
}

func TestCheckDropletRegion(t *testing.T) {
	tests := []struct {
		name    string
		pages   [][]godo.Droplet
		listErr error
		wantErr string
	}{
		{name: "no droplets"},
		{name: "same region", pages: [][]godo.Droplet{{droplet("demo-cp-0", "fra1"), droplet("demo-cp-1", "fra1")}}},
		{name: "droplet without region", pages: [][]godo.Droplet{{{Name: "demo-cp-0"}}}},
		{name: "other region", pages: [][]godo.Droplet{{droplet("demo-cp-0", "ams3")}}, wantErr: `in region(s) [ams3], but the control plane nodeSets use region "fra1"`},
		{name: "other region on second page", pages: [][]godo.Droplet{{droplet("demo-cp-0", "fra1")}, {droplet("demo-cp-1", "nyc1"), droplet("demo-cp-2", "ams3")}}, wantErr: "in region(s) [ams3 nyc1]"},
		{name: "list error", listErr: errors.New("boom"), wantErr: "boom"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeDropletLister{pages: tc.pages, err: tc.listErr}
			s := testState(nil, testNodeSets(validSpec, 3))

			err := checkDropletRegion(context.Background(), client, s.Cluster)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}

			if len(client.tags) == 0 || client.tags[0] != OwnedTag("demo") {
				t.Errorf("expected droplets listed by tag %q, got %v", OwnedTag("demo"), client.tags)
			}
			if want := max(len(tc.pages), 1); tc.listErr == nil && len(client.tags) != want {
				t.Errorf("expected %d list calls, got %d", want, len(client.tags))
			}
		})
	}
}
