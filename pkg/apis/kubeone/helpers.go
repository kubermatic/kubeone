/*
Copyright 2019 The KubeOne Authors.

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

package kubeone

import (
	"math/rand"

	"github.com/pkg/errors"
)

// Leader returns the first configured host. Only call this after
// validating the cluster config to ensure a leader exists.
func (c KubeOneCluster) Leader() (HostConfig, error) {
	for _, host := range c.ControlPlane.Hosts {
		if host.IsLeader {
			return host, nil
		}
	}
	return HostConfig{}, errors.New("leader not found")
}

func (c KubeOneCluster) RandomHost() HostConfig {
	//nolint:gosec
	// G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec)
	n := rand.Int31n(int32(len(c.ControlPlane.Hosts)))
	return c.ControlPlane.Hosts[n]
}

// Followers returns all but the first configured host. Only call
// this after validating the cluster config to ensure hosts exist.
func (c KubeOneCluster) Followers() []HostConfig {
	followers := []HostConfig{}
	for _, h := range c.ControlPlane.Hosts {
		if !h.IsLeader {
			followers = append(followers, h)
		}
	}
	return followers
}

// SetHostname sets the hostname for the given host
func (h *HostConfig) SetHostname(hostname string) {
	h.Hostname = hostname
}

// SetOperatingSystem sets the operating system for the given host
func (h *HostConfig) SetOperatingSystem(os OperatingSystemName) {
	h.OperatingSystem = os
}

// SetLeader sets is the given host leader
func (h *HostConfig) SetLeader(leader bool) {
	h.IsLeader = leader
}

// CloudProviderName returns name of the cloud provider
func (p CloudProviderSpec) CloudProviderName() string {
	switch {
	case p.AWS != nil:
		return "aws"
	case p.Azure != nil:
		return "azure"
	case p.DigitalOcean != nil:
		return "digitalocean"
	case p.GCE != nil:
		return "gce"
	case p.Hetzner != nil:
		return "hetzner"
	case p.Openstack != nil:
		return "openstack"
	case p.Packet != nil:
		return "packet"
	case p.Vsphere != nil:
		return "vsphere"
	case p.None != nil:
		return "none"
	}

	return ""
}

// CloudProviderInTree detects is there in-tree cloud provider implementation for specified provider.
// List of in-tree provider can be found here: https://github.com/kubernetes/kubernetes/tree/master/pkg/cloudprovider
func (p CloudProviderSpec) CloudProviderInTree() bool {
	if p.Openstack != nil {
		return !p.External
	} else if p.AWS != nil || p.GCE != nil || p.Vsphere != nil || p.Azure != nil {
		return true
	}

	return false
}

func (p CloudProviderSpec) CSIMigrationFeatureGates() map[string]bool {
	featureGates := map[string]bool{}

	switch {
	case p.AWS != nil:
		if p.AWS.CSIMigration {
			featureGates["CSIMigration"] = true
			featureGates["CSIMigrationAWS"] = true
		}
		if p.AWS.CSIMigrationComplete {
			featureGates["CSIMigrationAWSComplete"] = true
		}
	case p.Azure != nil:
		if p.Azure.CSIMigrationDisk {
			featureGates["CSIMigration"] = true
			featureGates["CSIMigrationAzureDisk"] = true
		}
		if p.Azure.CSIMigrationDiskComplete {
			featureGates["CSIMigrationAzureDiskComplete"] = true
		}
		if p.Azure.CSIMigrationFile {
			featureGates["CSIMigration"] = true
			featureGates["CSIMigrationAzureFile"] = true
		}
		if p.Azure.CSIMigrationFileComplete {
			featureGates["CSIMigrationAzureFileComplete"] = true
		}
	case p.GCE != nil:
		if p.GCE.CSIMigration {
			featureGates["CSIMigration"] = true
			featureGates["CSIMigrationGCE"] = true
		}
		if p.GCE.CSIMigrationComplete {
			featureGates["CSIMigrationGCEComplete"] = true
		}
	case p.Openstack != nil:
		if p.Openstack.CSIMigration {
			featureGates["CSIMigration"] = true
			featureGates["CSIMigrationOpenStack"] = true
		}
		if p.Openstack.CSIMigrationComplete {
			featureGates["CSIMigrationOpenStackComplete"] = true
		}
	case p.Vsphere != nil:
		if p.Vsphere.CSIMigration {
			featureGates["CSIMigration"] = true
			featureGates["CSIMigrationvSphere"] = true
		}
		if p.Vsphere.CSIMigrationComplete {
			featureGates["CSIMigrationvSphereComplete"] = true
		}
	}

	return featureGates
}
