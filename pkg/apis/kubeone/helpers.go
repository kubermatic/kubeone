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
	"errors"
	"math/rand"

	"github.com/Masterminds/semver"
)

// Leader returns the first configured host. Only call this after
// validating the cluster config to ensure a leader exists.
func (c KubeOneCluster) Leader() (HostConfig, error) {
	for _, host := range c.Hosts {
		if host.IsLeader {
			return host, nil
		}
	}
	return HostConfig{}, errors.New("leader not found")
}

func (c KubeOneCluster) RandomHost() HostConfig {
	n := rand.Int31n(int32(len(c.Hosts)))
	return c.Hosts[n]
}

// Followers returns all but the first configured host. Only call
// this after validating the cluster config to ensure hosts exist.
func (c KubeOneCluster) Followers() []HostConfig {
	followers := []HostConfig{}
	for _, h := range c.Hosts {
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
func (h *HostConfig) SetOperatingSystem(os string) {
	h.OperatingSystem = os
}

// SetLeader sets is the given host leader
func (h *HostConfig) SetLeader(leader bool) {
	h.IsLeader = leader
}

// CloudProviderInTree detects is there in-tree cloud provider implementation for specified provider.
// List of in-tree provider can be found here: https://github.com/kubernetes/kubernetes/tree/master/pkg/cloudprovider
func (p CloudProviderSpec) CloudProviderInTree() bool { //nolint:stylecheck
	switch p.Name {
	case CloudProviderNameAWS, CloudProviderNameGCE, CloudProviderNameOpenStack:
		return true
	case CloudProviderNameVSphere, CloudProviderNameAzure:
		return true
	default:
		return false
	}
}

// KubernetesCNIVersion returns kubernetes-cni package version
func (m VersionConfig) KubernetesCNIVersion() string {
	s := semver.MustParse(m.Kubernetes)
	c, _ := semver.NewConstraint(">= 1.13.0, <= 1.13.4")

	switch {
	// Validation ensures that the oldest cluster version is 1.13.0.
	// Versions 1.13.0-1.13.4 uses 0.6.0, so it's safe to return 0.6.0
	// if >= 1.13.0, <= 1.13.4 constraint check successes.
	case c.Check(s):
		return "0.6.0"
	default:
		return "0.7.5"
	}
}
