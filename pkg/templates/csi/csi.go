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

package csi

import (
	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/addons"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"
)

// Ensure external CCM deployen if Provider.External
func Ensure(s *state.State) error {
	if !s.Cluster.CloudProvider.External {
		return nil
	}

	s.Logger.Info("Ensure CSI driver is up to date...")
	var err error

	switch {
	case s.Cluster.CloudProvider.AWS != nil:
		if err = addons.EnsureAddonByName(s, resources.AddonCSIAwsEBS); err != nil {
			return errors.Wrap(err, "failed to deploy AWS EBS CSI driver")
		}
	case s.Cluster.CloudProvider.Azure != nil:
		if s.Cluster.CloudProvider.CloudConfig == "" {
			return errors.New("cloudConfig not defined")
		}

		// Deploy AzureDisk CSI driver
		if err = addons.EnsureAddonByName(s, resources.AddonCSIAzureDisk); err != nil {
			return errors.Wrap(err, "failed to deploy azuredisk CSI driver")
		}

		// Deploy AzureFile CSI driver
		err = addons.EnsureAddonByName(s, resources.AddonCSIAzureFile)
		return errors.Wrap(err, "failed to deploy azurefile CSI driver")
	case s.Cluster.CloudProvider.Hetzner != nil:
		err = addons.EnsureAddonByName(s, resources.AddonCSIHetnzer)
	case s.Cluster.CloudProvider.Openstack != nil:
		if s.Cluster.CloudProvider.CloudConfig == "" {
			return errors.New("cloudConfig not defined")
		}
		err = addons.EnsureAddonByName(s, resources.AddonCSIOpenStackCinder)
	case s.Cluster.CloudProvider.Vsphere != nil:
		if s.Cluster.CloudProvider.CSIConfig == "" {
			s.Logger.Warnln("vSphere CSI driver requires CSI config to be provided via .cloudProvider.csiConfig. Skipping...")
			return nil
		}
		err = addons.EnsureAddonByName(s, resources.AddonCSIVsphere)
	default:
		s.Logger.Infof("CSI driver for %q not yet supported, skipping", s.Cluster.CloudProvider.CloudProviderName())
		return nil
	}

	return errors.Wrap(err, "failed to ensure CSI driver is installed")
}
