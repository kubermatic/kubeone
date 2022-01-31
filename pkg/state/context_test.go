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

package state

import (
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func TestShouldEnableInTreeCloudProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		cluster              *kubeoneapi.KubeOneCluster
		liveCluster          *Cluster
		ccmMigrationComplete bool
		want                 bool
	}{
		{
			name: "new OpenStack cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 true,
		},
		{
			name: "new OpenStack cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "new Hetzner cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "new Hetzner cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "new vSphere cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Vsphere:  &kubeoneapi.VsphereSpec{},
					External: false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 true,
		},
		{
			name: "new vSphere cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Vsphere:  &kubeoneapi.VsphereSpec{},
					External: true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "existing OpenStack cluster with in-tree cloud provider enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					InTreeCloudProviderEnabled: true,
				},
			},
			ccmMigrationComplete: false,
			want:                 true,
		},
		{
			name: "existing OpenStack cluster with in-tree cloud provider disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					InTreeCloudProviderEnabled: false,
				},
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "existing OpenStack cluster with in-tree cloud provider and ccmMigrationComplete enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					InTreeCloudProviderEnabled: false,
				},
			},
			ccmMigrationComplete: true,
			want:                 false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := &State{
				Cluster:     tc.cluster,
				LiveCluster: tc.liveCluster,
			}

			if got := s.ShouldEnableInTreeCloudProvider(); got != tc.want {
				t.Errorf("State.ShouldEnableInTreeCloudProvider() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestShouldEnableCSIMigration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cluster      *kubeoneapi.KubeOneCluster
		liveCluster  *Cluster
		ccmMigration bool
		want         bool
	}{
		{
			name: "new OpenStack cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigration: false,
			want:         false,
		},
		{
			name: "new OpenStack cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigration: false,
			want:         true,
		},
		{
			name: "new Hetzner cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigration: false,
			want:         false,
		},
		{
			name: "new Hetzner cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigration: false,
			want:         false,
		},
		{
			name: "new vSphere cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigration: false,
			want:         false,
		},
		{
			name: "new vSphere cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigration: false,
			want:         false,
		},
		{
			name: "existing OpenStack cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					CSIMigrationEnabled: false,
				},
			},
			ccmMigration: false,
			want:         false,
		},
		{
			name: "existing OpenStack cluster with external enabled and disabled CSI migration",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					CSIMigrationEnabled: false,
				},
			},
			ccmMigration: false,
			want:         false,
		},
		{
			name: "existing OpenStack cluster with already enabled CSI feature gates",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					CSIMigrationEnabled: true,
				},
			},
			ccmMigration: false,
			want:         true,
		},
		{
			name: "existing OpenStack cluster with disabled CSI feature gates with CSI migration started",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					CSIMigrationEnabled: false,
				},
			},
			ccmMigration: true,
			want:         true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := &State{
				Cluster:      tc.cluster,
				LiveCluster:  tc.liveCluster,
				CCMMigration: tc.ccmMigration,
			}

			if got := s.ShouldEnableCSIMigration(); got != tc.want {
				t.Errorf("State.ShouldEnableCSIMigration() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestShouldUnregisterInTreeProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		cluster              *kubeoneapi.KubeOneCluster
		liveCluster          *Cluster
		ccmMigrationComplete bool
		want                 bool
	}{
		{
			name: "new OpenStack cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "new OpenStack cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 true,
		},
		{
			name: "new Hetzner cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "new Hetzner cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "new vSphere cluster with external disabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: false,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "new vSphere cluster with external enabled",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner:  &kubeoneapi.HetznerSpec{},
					External: true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: nil,
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "existing OpenStack cluster with external disabled and in-tree provider registered",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					InTreeCloudProviderUnregistered: false,
				},
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "existing OpenStack cluster with external enabled and in-tree provider registered",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					InTreeCloudProviderUnregistered: false,
				},
			},
			ccmMigrationComplete: false,
			want:                 false,
		},
		{
			name: "existing OpenStack cluster with external enabled and in-tree provider unregistered",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					InTreeCloudProviderUnregistered: true,
				},
			},
			ccmMigrationComplete: false,
			want:                 true,
		},
		{
			name: "existing OpenStack cluster with external enabled, in-tree provider registered, and CSI migration completed",
			cluster: &kubeoneapi.KubeOneCluster{
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Openstack: &kubeoneapi.OpenstackSpec{},
					External:  true,
				},
			},
			liveCluster: &Cluster{
				CCMStatus: &CCMStatus{
					InTreeCloudProviderUnregistered: false,
				},
			},
			ccmMigrationComplete: true,
			want:                 true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &State{
				Cluster:              tc.cluster,
				LiveCluster:          tc.liveCluster,
				CCMMigrationComplete: tc.ccmMigrationComplete,
			}

			if got := s.ShouldUnregisterInTreeCloudProvider(); got != tc.want {
				t.Errorf("State.ShouldUnregisterInTreeProvider() = %v, want %v", got, tc.want)
			}
		})
	}
}
