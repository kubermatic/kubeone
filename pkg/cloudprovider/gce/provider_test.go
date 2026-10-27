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

package gce

import (
	"context"
	"encoding/base64"
	"os"
	"slices"
	"strings"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	gcetypes "k8c.io/machine-controller/sdk/cloudprovider/gce"
	"k8c.io/machine-controller/sdk/providerconfig"
)

const testServiceAccount = `{"type":"service_account","project_id":"proj","client_email":"sa@proj.iam.gserviceaccount.com"}`

func testNodeSets() []kubeoneapi.NodeSet {
	return []kubeoneapi.NodeSet{
		{
			Name:              "cp-a",
			Replicas:          2,
			CloudProviderSpec: []byte(`{"zone":"europe-west3-a","machineType":"e2-standard-2","tags":["user-tag"],"labels":{"team":"x"}}`),
			NodeSettings:      kubeoneapi.NodeSettingsSpec{Labels: map[string]string{"user": "label"}},
		},
		{
			Name:              "cp-b",
			Replicas:          1,
			CloudProviderSpec: []byte(`{"zone":"europe-west3-b","machineType":"e2-standard-2"}`),
		},
	}
}

func TestGenerateMachines(t *testing.T) {
	machines, err := generateGCEControlPlaneMachines("demo", testNodeSets(), "1.36.0")
	if err != nil {
		t.Fatalf("generating machines: %v", err)
	}

	var names []string
	for _, m := range machines {
		names = append(names, m.Name)

		if string(m.UID) != m.Name {
			t.Errorf("machine %q has UID %q, CleanupVMs needs it to be the deterministic name", m.Name, m.UID)
		}

		pconfig, cfgErr := providerconfig.GetConfig(m.Spec.ProviderSpec)
		if cfgErr != nil {
			t.Fatalf("reading provider config: %v", cfgErr)
		}
		if pconfig.CloudProvider != providerconfig.CloudProviderGoogle {
			t.Errorf("machine %q has cloud provider %q", m.Name, pconfig.CloudProvider)
		}

		gceConfig, cfgErr := gcetypes.GetConfig(*pconfig)
		if cfgErr != nil {
			t.Fatalf("decoding gce config: %v", cfgErr)
		}

		tags := 0
		for _, tag := range gceConfig.Tags {
			if tag == "demo-control-plane" {
				tags++
			}
		}
		if tags != 1 {
			t.Errorf("machine %q has tags %v, want the control plane tag exactly once", m.Name, gceConfig.Tags)
		}

		if gceConfig.Labels["kubeone_cluster_name"] != "demo" || gceConfig.Labels["kubeone_role"] != "control-plane" {
			t.Errorf("machine %q has labels %v", m.Name, gceConfig.Labels)
		}
	}

	want := []string{"demo-cp-a-0", "demo-cp-a-1", "demo-cp-b-0"}
	if !slices.Equal(names, want) {
		t.Errorf("got machines %v, want %v", names, want)
	}

	pconfig, err := providerconfig.GetConfig(machines[0].Spec.ProviderSpec)
	if err != nil {
		t.Fatalf("reading provider config: %v", err)
	}
	first, err := gcetypes.GetConfig(*pconfig)
	if err != nil {
		t.Fatalf("decoding gce config: %v", err)
	}
	if !slices.Contains(first.Tags, "user-tag") || first.Labels["team"] != "x" {
		t.Errorf("user tags/labels were not kept: %v %v", first.Tags, first.Labels)
	}
}

func TestGenerateMachinesRejectsUnknownFields(t *testing.T) {
	nodeSets := []kubeoneapi.NodeSet{{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(`{"zone":"europe-west3-a","foo":"bar"}`)}}

	if _, err := generateGCEControlPlaneMachines("demo", nodeSets, "1.36.0"); err == nil {
		t.Fatal("expected an error for an unknown cloudProviderSpec field")
	}
}

func TestHostConfigsFromMachinesUseInstanceNames(t *testing.T) {
	nodeSets := testNodeSets()[:1]
	capimachines, err := generateGCEControlPlaneMachines("demo", nodeSets, "1.36.0")
	if err != nil {
		t.Fatalf("generating machines: %v", err)
	}
	machines := []provisioner.Machine{
		{PublicAddress: "1.2.3.4", PrivateAddress: "10.0.0.2", Hostname: "demo-cp-a-0.c.a-very-long-project-name.internal"},
		{PublicAddress: "1.2.3.5", PrivateAddress: "10.0.0.3", Hostname: "demo-cp-a-1.europe-west3-a.c.proj.internal"},
	}

	hosts := hostConfigsFromMachines(machines, capimachines, nodeSets)
	if len(hosts) != 2 {
		t.Fatalf("got %d hosts, want 2", len(hosts))
	}

	for i, want := range []string{"demo-cp-a-0", "demo-cp-a-1"} {
		if hosts[i].Hostname != want {
			t.Errorf("host %s has hostname %q, want the instance name %q", hosts[i].PublicAddress, hosts[i].Hostname, want)
		}
	}
	if hosts[0].Labels["user"] != "label" {
		t.Errorf("nodeSet labels not kept: %v", hosts[0].Labels)
	}
}

func testState() *state.State {
	return &state.State{
		Cluster: &kubeoneapi.KubeOneCluster{
			Name: "demo",
			CloudProvider: kubeoneapi.CloudProviderSpec{
				GCE: &kubeoneapi.GCESpec{ControlPlane: &kubeoneapi.GCEControlPlane{
					LoadBalancer: kubeoneapi.GCELoadBalancer{Name: "demo-kubeapi", Region: "europe-west3"},
				}},
			},
		},
	}
}

func TestExportCredentials(t *testing.T) {
	t.Setenv(credentials.GoogleServiceAccountKey, testServiceAccount)
	t.Setenv(credentials.GoogleServiceAccountKeyMC, "")

	if err := exportCredentials(testState()); err != nil {
		t.Fatalf("exportCredentials: %v", err)
	}

	got, err := base64.StdEncoding.DecodeString(os.Getenv(credentials.GoogleServiceAccountKeyMC))
	if err != nil {
		t.Fatalf("%s is not base64: %v", credentials.GoogleServiceAccountKeyMC, err)
	}
	if string(got) != testServiceAccount {
		t.Errorf("got %s, want %s", got, testServiceAccount)
	}
}

func TestResolveLB(t *testing.T) {
	t.Setenv(credentials.GoogleServiceAccountKey, testServiceAccount)

	t.Run("project from the service account", func(t *testing.T) {
		cfg, err := resolveLB(testState())
		if err != nil {
			t.Fatalf("resolveLB: %v", err)
		}
		if cfg.ProjectID != "proj" || cfg.owner() != "kubeone:demo" || cfg.negName("europe-west3-a") != "demo-kubeapi-europe-west3-a" {
			t.Errorf("unexpected config: %+v", cfg)
		}
	})

	t.Run("configured project wins", func(t *testing.T) {
		s := testState()
		s.Cluster.CloudProvider.GCE.ControlPlane.LoadBalancer.ProjectID = "other"

		cfg, err := resolveLB(s)
		if err != nil {
			t.Fatalf("resolveLB: %v", err)
		}
		if cfg.ProjectID != "other" {
			t.Errorf("got project %q, want other", cfg.ProjectID)
		}
	})

	t.Run("no region", func(t *testing.T) {
		s := testState()
		s.Cluster.CloudProvider.GCE.ControlPlane.LoadBalancer.Region = ""

		if _, err := resolveLB(s); err == nil {
			t.Fatal("expected an error without a region")
		}
	})
}

func TestResourcePath(t *testing.T) {
	for _, in := range []string{
		"https://www.googleapis.com/compute/v1/projects/p/zones/z/instances/i",
		"https://compute.googleapis.com/compute/v1/projects/p/zones/z/instances/i",
		"projects/p/zones/z/instances/i",
	} {
		if got := resourcePath(in); got != "projects/p/zones/z/instances/i" {
			t.Errorf("resourcePath(%q) = %q", in, got)
		}
	}
}

func TestCheckInstancesOwned(t *testing.T) {
	machines, err := generateGCEControlPlaneMachines("demo", testNodeSets(), "1.36.0")
	if err != nil {
		t.Fatalf("generating machines: %v", err)
	}

	owned := func(name string) map[string]any {
		return map[string]any{"labels": map[string]any{
			"kubeone_cluster_name": "demo",
			"kubeone_role":         "control-plane",
			"machine_uid":          name,
		}}
	}

	tests := []struct {
		name      string
		instances map[string]map[string]any
		wantErr   string
	}{
		{
			name: "no instances",
		},
		{
			name: "all instances owned",
			instances: map[string]map[string]any{
				"projects/proj/zones/europe-west3-a/instances/demo-cp-a-0": owned("demo-cp-a-0"),
				"projects/proj/zones/europe-west3-a/instances/demo-cp-a-1": owned("demo-cp-a-1"),
				"projects/proj/zones/europe-west3-b/instances/demo-cp-b-0": owned("demo-cp-b-0"),
			},
		},
		{
			name: "instance without labels",
			instances: map[string]map[string]any{
				"projects/proj/zones/europe-west3-a/instances/demo-cp-a-0": owned("demo-cp-a-0"),
				"projects/proj/zones/europe-west3-b/instances/demo-cp-b-0": {},
			},
			wantErr: `instance "demo-cp-b-0"`,
		},
		{
			name: "instance of another cluster",
			instances: map[string]map[string]any{
				"projects/proj/zones/europe-west3-a/instances/demo-cp-a-0": {"labels": map[string]any{
					"kubeone_cluster_name": "other",
					"kubeone_role":         "control-plane",
					"machine_uid":          "demo-cp-a-0",
				}},
			},
			wantErr: "kubeone_cluster_name",
		},
		{
			name: "instance not created by machine-controller",
			instances: map[string]map[string]any{
				"projects/proj/zones/europe-west3-a/instances/demo-cp-a-0": {"labels": map[string]any{
					"kubeone_cluster_name": "demo",
					"kubeone_role":         "control-plane",
				}},
			},
			wantErr: "machine_uid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, svc := newFakeCompute(t)
			for p, obj := range tc.instances {
				f.put(p, obj)
			}

			err := checkInstancesOwned(context.Background(), svc, testLBConfig(), "demo", machines)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}

			if got := f.mutations(); len(got) != 0 {
				t.Errorf("check must not change anything, got %v", got)
			}
		})
	}
}

func TestLeaderMachineName(t *testing.T) {
	nodeSets := testNodeSets()

	machines, err := generateGCEControlPlaneMachines("demo", nodeSets, "1.36.0")
	if err != nil {
		t.Fatalf("generating machines: %v", err)
	}

	// KubeOne makes the first control plane host the leader, the hosts are
	// added in the order of the generated machines
	if got := leaderMachineName("demo", nodeSets); got != machines[0].Name {
		t.Errorf("got leader %q, want the first machine %q", got, machines[0].Name)
	}

	nodeSets[0].Replicas = 0
	if got := leaderMachineName("demo", nodeSets); got != "demo-cp-b-0" {
		t.Errorf("got leader %q with an empty first nodeSet, want demo-cp-b-0", got)
	}

	if got := leaderMachineName("demo", nil); got != "" {
		t.Errorf("got leader %q without nodeSets, want none", got)
	}
}

func TestMachineZoneProject(t *testing.T) {
	machine := func(spec string) clusterv1alpha1.Machine {
		machines, err := generateGCEControlPlaneMachines("demo", []kubeoneapi.NodeSet{{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(spec)}}, "1.36.0")
		if err != nil {
			t.Fatalf("generating machines: %v", err)
		}

		return machines[0]
	}

	tests := []struct {
		name      string
		lbProject string
		spec      string
		wantErr   bool
	}{
		{
			name:      "VM and load balancer in the service account project",
			lbProject: "proj",
			spec:      `{"zone":"europe-west3-a"}`,
		},
		{
			// The VM is created in the service account project "proj"
			name:      "load balancer in another project than the service account",
			lbProject: "other",
			spec:      `{"zone":"europe-west3-a"}`,
			wantErr:   true,
		},
		{
			name:      "nodeSet in the load balancer project",
			lbProject: "other",
			spec:      `{"zone":"europe-west3-a","projectID":"other"}`,
		},
		{
			name:      "nodeSet in another project than the load balancer",
			lbProject: "proj",
			spec:      `{"zone":"europe-west3-a","projectID":"other"}`,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testLBConfig()
			cfg.ProjectID = tc.lbProject

			zone, err := machineZone(machine(tc.spec), cfg)
			switch {
			case tc.wantErr && err == nil:
				t.Fatal("expected a project error")
			case !tc.wantErr && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case !tc.wantErr && zone != "europe-west3-a":
				t.Errorf("got zone %q", zone)
			}
		})
	}
}

func TestCheckInstancesOwnedInVMProject(t *testing.T) {
	machines, err := generateGCEControlPlaneMachines("demo", testNodeSets()[1:], "1.36.0")
	if err != nil {
		t.Fatalf("generating machines: %v", err)
	}

	// The load balancer is in another project than the service account, the
	// VM was created in the service account project. Reset must still find
	// and check it there instead of failing on the project mismatch.
	f, svc := newFakeCompute(t)
	f.put("projects/proj/zones/europe-west3-b/instances/demo-cp-b-0", map[string]any{"labels": map[string]any{
		"kubeone_cluster_name": "other",
		"kubeone_role":         "control-plane",
		"machine_uid":          "demo-cp-b-0",
	}})

	cfg := testLBConfig()
	cfg.ProjectID = "lb-project"

	err = checkInstancesOwned(context.Background(), svc, cfg, "demo", machines)
	if err == nil || !strings.Contains(err.Error(), "kubeone_cluster_name") {
		t.Fatalf("expected the foreign VM in the service account project to be found, got %v", err)
	}
}

func TestGenerateMachinesKeepsNodeSetLabels(t *testing.T) {
	nodeSets := testNodeSets()

	for range 2 {
		if _, err := generateGCEControlPlaneMachines("demo", nodeSets, "1.36.0"); err != nil {
			t.Fatalf("generating machines: %v", err)
		}
	}

	// The labels of the cluster config become node labels, the kubeone_*
	// machine labels (with a new timestamp every time) must not leak into them
	if got := nodeSets[0].NodeSettings.Labels; len(got) != 1 || got["user"] != "label" {
		t.Errorf("nodeSet labels were modified: %v", got)
	}
	if nodeSets[1].NodeSettings.Labels != nil {
		t.Errorf("nodeSet without labels got labels: %v", nodeSets[1].NodeSettings.Labels)
	}
}
