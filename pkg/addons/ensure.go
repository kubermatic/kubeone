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
	"fmt"
	"io/fs"
	"strings"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"
)

const (
	// addonLabel is applied to all objects deployed using addons
	addonLabel = "kubeone.io/addon"
)

var (
	// embeddedAddons is a list of addons that are embedded in the KubeOne
	// binary. Those addons are skipped when applying the user-provided addons
	embeddedAddons = map[string]string{
		resources.AddonCCMDigitalOcean: "",
		resources.AddonCCMHetzner:      "",
		resources.AddonCCMOpenStack:    "",
		resources.AddonCCMPacket:       "",
		resources.AddonCNICanal:        "",
		resources.AddonCNIWeavenet:     "",
		resources.AddonNodeLocalDNS:    "",
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

	addons, err := fs.ReadDir(applier.LocalFS, ".")
	if err != nil {
		return errors.Wrap(err, "failed to read addons directory")
	}

	for _, a := range addons {
		if !a.IsDir() {
			continue
		}
		if _, ok := embeddedAddons[a.Name()]; ok {
			continue
		}

		s.Logger.Infof("Applying addon %q...", a.Name())

		if err := applier.loadAndApplyAddon(s, applier.LocalFS, a.Name()); err != nil {
			return errors.Wrapf(err, "failed to load and apply the addon %q", a.Name())
		}
	}

	s.Logger.Info("Applying addons from the root directory...")
	if err := applier.loadAndApplyAddon(s, applier.LocalFS, ""); err != nil {
		return errors.Wrap(err, "failed to load and apply addons from the root directory")
	}

	return nil
}

// EnsureAddonByName deploys an addon by its name. If the addon is not found
// in the addons directory, or if the addons or not enabled, it will search
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

// loadAndApplyAddon parses the addons manifests and runs kubectl apply.
func (a *applier) loadAndApplyAddon(s *state.State, f fs.FS, addonName string) error {
	manifest, err := a.getManifestsFromDirectory(s, f, addonName)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(strings.TrimSpace(manifest)) == 0 {
		if len(addonName) != 0 {
			s.Logger.Warnf("Addon directory %q is empty, skipping...", addonName)
		}
		return nil
	}

	if err := runKubectlApply(s, manifest, addonName); err != nil {
		return errors.Wrap(err, "failed to apply addons")
	}

	return nil
}

// runKubectlApply runs kubectl apply command
func runKubectlApply(s *state.State, manifest string, addonName string) error {
	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
		var (
			cmd            = fmt.Sprintf(kubectlApplyScript, addonLabel, addonName)
			stdin          = strings.NewReader(manifest)
			stdout, stderr strings.Builder
		)

		_, err := conn.POpen(cmd, stdin, &stdout, &stderr)
		if s.Verbose {
			fmt.Printf("+ %s\n", cmd)
			fmt.Printf("%s", stderr.String())
			fmt.Printf("%s", stdout.String())
		}

		return err
	})
}
