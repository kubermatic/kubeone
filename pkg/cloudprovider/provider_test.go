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

package cloudprovider

import (
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/provisioner"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
)

func testNodeSets() []kubeoneapi.NodeSet {
	return []kubeoneapi.NodeSet{
		{Name: "a", Replicas: 2},
		{Name: "b", Replicas: 1, SSH: kubeoneapi.SSHSpec{Username: "rocky", PrivateKeyFile: "/b.key"}},
	}
}

func TestMachineNodeSets(t *testing.T) {
	machines := make([]clusterv1alpha1.Machine, 3)

	got, err := MachineNodeSets(machines, testNodeSets())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"a", "a", "b"}
	for i, nodeSet := range got {
		if nodeSet.Name != want[i] {
			t.Errorf("machine %d: expected nodeSet %q, got %q", i, want[i], nodeSet.Name)
		}
	}

	if _, err = MachineNodeSets(machines[:2], testNodeSets()); err == nil {
		t.Error("expected an error when machines don't match the replicas")
	}
}

// EnsureVM maps a single machine, which must get the settings of its own
// nodeSet rather than the first one.
func TestHostConfigsFromMachinesSingleNodeSet(t *testing.T) {
	nodeSet := testNodeSets()[1]

	hosts := HostConfigsFromMachines([]provisioner.Machine{{PublicAddress: "1.2.3.4"}}, []kubeoneapi.NodeSet{nodeSet})
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].SSHUsername != "rocky" || hosts[0].SSHPrivateKeyFile != "/b.key" || hosts[0].PublicAddress != "1.2.3.4" {
		t.Errorf("unexpected host %+v", hosts[0])
	}
}

func TestHostConfigsFromMachines(t *testing.T) {
	machines := []provisioner.Machine{{Hostname: "a0"}, {Hostname: "a1"}, {Hostname: "b0"}}

	hosts := HostConfigsFromMachines(machines, testNodeSets())
	if len(hosts) != 3 {
		t.Fatalf("expected 3 hosts, got %d", len(hosts))
	}

	wantUsers := []string{"root", "root", "rocky"}
	for i, host := range hosts {
		if host.Hostname != machines[i].Hostname || host.SSHUsername != wantUsers[i] {
			t.Errorf("host %d: unexpected hostname/user %q/%q", i, host.Hostname, host.SSHUsername)
		}
	}
}
