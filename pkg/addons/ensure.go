/*
Copyright 2020 The KubeOne Authors.

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

package addons

import (
	"io/fs"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"
)

const (
	// addonLabel is applied to all objects deployed using addons
	addonLabel = "kubeone.io/addon"
)

var (
	// embeddedAddons is a list of addons that are embedded in the KubeOne
	// binary. Those addons are skipped when applying a user-provided addon with the same name.
	embeddedAddons = map[string]string{
		resources.AddonCCMAws:             "",
		resources.AddonCCMAzure:           "",
		resources.AddonCCMDigitalOcean:    "",
		resources.AddonCCMHetzner:         "",
		resources.AddonCCMOpenStack:       "",
		resources.AddonCCMPacket:          "",
		resources.AddonCCMVsphere:         "",
		resources.AddonCNICanal:           "",
		resources.AddonCNICilium:          "",
		resources.AddonCNIWeavenet:        "",
		resources.AddonCSIAwsEBS:          "",
		resources.AddonCSIAzureDisk:       "",
		resources.AddonCSIAzureFile:       "",
		resources.AddonCSIHetnzer:         "",
		resources.AddonCSIOpenStackCinder: "",
		resources.AddonCSIVsphere:         "",
		resources.AddonMachineController:  "",
		resources.AddonMetricsServer:      "",
		resources.AddonNodeLocalDNS:       "",
	}
)

// EnsureUserAddons deploys addons that are provided by the user and that are
// not embedded.
func EnsureUserAddons(s *state.State) error {
	applier, err := newAddonsApplier(s)
	if err != nil {
		return err
	}

	if applier.LocalFS == nil {
		s.Logger.Infoln("Skipping applying addons because addons are not enabled...")
		return nil
	}

	s.Logger.Infof("Applying user provided addons...")

	customAddons, err := fs.ReadDir(applier.LocalFS, ".")
	if err != nil {
		return errors.Wrap(err, "failed to read addons directory")
	}

	combinedAddons := map[string]string{}
	for _, useraddon := range customAddons {
		if !useraddon.IsDir() {
			continue
		}

		if _, ok := embeddedAddons[useraddon.Name()]; ok {
			continue
		}

		if _, ok := combinedAddons[useraddon.Name()]; !ok {
			combinedAddons[useraddon.Name()] = ""
		}
	}

	for _, embeddedAddon := range s.Cluster.Addons.Addons {
		if _, ok := embeddedAddons[embeddedAddon.Name]; ok {
			continue
		}

		if embeddedAddon.Delete {
			s.Logger.Infof("Deleting addon %q...", embeddedAddon.Name)
			if err := applier.loadAndDeleteAddon(s, applier.EmbededFS, embeddedAddon.Name); err != nil {
				return errors.Wrapf(err, "failed to load and delete the addon %q", embeddedAddon.Name)
			}
			continue
		}

		if _, ok := combinedAddons[embeddedAddon.Name]; !ok {
			combinedAddons[embeddedAddon.Name] = ""
		}
	}

	for addonName := range combinedAddons {
		s.Logger.Infof("Applying addon %q...", addonName)

		if err := EnsureAddonByName(s, addonName); err != nil {
			return errors.Wrapf(err, "failed to load and apply the addon %q", addonName)
		}
	}

	s.Logger.Info("Applying addons from the root directory...")
	if err := applier.loadAndApplyAddon(s, applier.LocalFS, ""); err != nil {
		return errors.Wrap(err, "failed to load and apply addons from the root directory")
	}

	return nil
}

// EnsureAddonByName deploys an addon by its name. If the addon is not found
// in the addons directory, or if the addons are not enabled, it will search
// for the embedded addons.
func EnsureAddonByName(s *state.State, addonName string) error {
	applier, err := newAddonsApplier(s)
	if err != nil {
		return err
	}

	if applier.LocalFS != nil {
		addons, lErr := fs.ReadDir(applier.LocalFS, ".")
		if lErr != nil {
			return errors.Wrap(lErr, "failed to read addons directory")
		}

		for _, a := range addons {
			if !a.IsDir() {
				continue
			}
			if a.Name() == addonName {
				if err := applier.loadAndApplyAddon(s, applier.LocalFS, a.Name()); err != nil {
					return errors.Wrap(err, "failed to load and apply addon")
				}
				return nil
			}
		}
	}

	addons, eErr := fs.ReadDir(applier.EmbededFS, ".")
	if eErr != nil {
		return errors.Wrap(eErr, "failed to read embedded addons")
	}

	for _, a := range addons {
		if !a.IsDir() {
			continue
		}
		if a.Name() == addonName {
			if err := applier.loadAndApplyAddon(s, applier.EmbededFS, a.Name()); err != nil {
				return errors.Wrap(err, "failed to load and apply embedded addon")
			}
			return nil
		}
	}

	return errors.Errorf("addon %q does not exist", addonName)
}
