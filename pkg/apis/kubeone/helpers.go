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

	"github.com/Masterminds/semver"
)

// Leader returns the first configured host. Only call this after
// validating the cluster config to ensure a leader exists.
func (c *KubeOneCluster) Leader() (*HostConfig, error) {
	for i := range c.Spec.Hosts {
		if c.Spec.Hosts[i].IsLeader {
			return c.Spec.Hosts[i], nil
		}
	}
	return nil, errors.New("leader not found")
}

// Followers returns all but the first configured host. Only call
// this after validating the cluster config to ensure hosts exist.
func (c *KubeOneCluster) Followers() []*HostConfig {
	return c.Spec.Hosts[1:]
}

// CloudProviderInTree detects is there in-tree cloud provider implementation for specified provider.
// List of in-tree provider can be found here: https://github.com/kubernetes/kubernetes/tree/master/pkg/cloudprovider
func (p *ProviderConfig) CloudProviderInTree() bool {
	switch p.Name {
	case ProviderNameAWS, ProviderNameGCE, ProviderNameOpenStack, ProviderNameVSphere:
		return true
	default:
		return false
	}
}

// KubernetesCNIVersion returns kubernetes-cni package version
func (m *VersionConfig) KubernetesCNIVersion() string {
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
