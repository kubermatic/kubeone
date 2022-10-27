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

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/semverutil"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"
	"k8c.io/kubeone/pkg/templates/weave"
)

const (
	// addonLabel is applied to all objects deployed using addons
	addonLabel = "kubeone.io/addon"

	// greaterThan23Constraint defines a semver constraint that validates Kubernetes versions is greater than 1.23
	greaterThan23Constraint = ">= 1.23"

	// defaultStorageClass addon defines name of the default-storage-class addon
	defaultStorageClassAddonName = "default-storage-class"
)

var (
	// embeddedAddons is a list of addons that are embedded in the KubeOne
	// binary. Those addons are skipped when applying a user-provided addon with the same name.
	embeddedAddons = map[string]string{
		resources.AddonCCMAws:                 "",
		resources.AddonCCMAzure:               "",
		resources.AddonCCMDigitalOcean:        "",
		resources.AddonCCMHetzner:             "",
		resources.AddonCCMOpenStack:           "",
		resources.AddonCCMEquinixMetal:        "",
		resources.AddonCCMPacket:              "",
		resources.AddonCCMVsphere:             "",
		resources.AddonCNICanal:               "",
		resources.AddonCNICilium:              "",
		resources.AddonCNIWeavenet:            "",
		resources.AddonCSIAwsEBS:              "",
		resources.AddonCSIAzureDisk:           "",
		resources.AddonCSIAzureFile:           "",
		resources.AddonCSIDigitalOcean:        "",
		resources.AddonCSIHetzner:             "",
		resources.AddonCSIGCPComputePD:        "",
		resources.AddonCSINutanix:             "",
		resources.AddonCSIOpenStackCinder:     "",
		resources.AddonCSIVMwareCloudDirector: "",
		resources.AddonCSIVsphere:             "",
		resources.AddonMachineController:      "",
		resources.AddonMetricsServer:          "",
		resources.AddonNodeLocalDNS:           "",
		resources.AddonOperatingSystemManager: "",
	}

	greaterThan23 = semverutil.MustParseConstraint(greaterThan23Constraint)
)

type addonAction struct {
	name      string
	supportFn func() error
}

//nolint:nakedret
func collectAddons(s *state.State) (addonsToDeploy []addonAction) {
	if *s.Cluster.Features.CoreDNS.DeployPodDisruptionBudget {
		addonsToDeploy = append(addonsToDeploy, addonAction{
			name: resources.AddonCoreDNSPDB,
		})
	}

	if s.Cluster.Features.MetricsServer.Enable {
		addonsToDeploy = append(addonsToDeploy, addonAction{
			name: resources.AddonMetricsServer,
		})
	}

	switch {
	case s.Cluster.ClusterNetwork.CNI.Canal != nil:
		addonsToDeploy = append(addonsToDeploy, addonAction{
			name: resources.AddonCNICanal,
		})
	case s.Cluster.ClusterNetwork.CNI.Cilium != nil:
		addonsToDeploy = append(addonsToDeploy, addonAction{
			name: resources.AddonCNICilium,
		})
	case s.Cluster.ClusterNetwork.CNI.WeaveNet != nil:
		addonsToDeploy = append(addonsToDeploy, addonAction{
			name: resources.AddonCNIWeavenet,
			supportFn: func() error {
				if s.Cluster.ClusterNetwork.CNI.WeaveNet.Encrypted {
					if err := weave.EnsureSecret(s); err != nil {
						return err
					}
				}

				return nil
			},
		})
	}

	if s.Cluster.Features.NodeLocalDNS.Deploy {
		addonsToDeploy = append(addonsToDeploy, addonAction{
			name: resources.AddonNodeLocalDNS,
		})
	}

	if s.Cluster.MachineController.Deploy {
		addonsToDeploy = append(addonsToDeploy, addonAction{
			name: resources.AddonMachineController,
		})
	}

	if s.Cluster.OperatingSystemManager.Deploy {
		addonsToDeploy = append(addonsToDeploy, addonAction{
			name: resources.AddonOperatingSystemManager,
		})
	}

	addonsToDeploy = ensureCSIAddons(s, addonsToDeploy)

	if s.Cluster.CloudProvider.External {
		addonsToDeploy = ensureCCMAddons(s, addonsToDeploy)
	}

	return
}

func cleanupAddons(s *state.State) error {
	if !*s.Cluster.Features.CoreDNS.DeployPodDisruptionBudget {
		if err := DeleteAddonByName(s, resources.AddonCoreDNSPDB); err != nil {
			return err
		}
	}

	return nil
}

func Ensure(s *state.State) error {
	addonsToDeploy := collectAddons(s)

	for _, add := range addonsToDeploy {
		if add.supportFn != nil {
			if err := add.supportFn(); err != nil {
				return err
			}
		}
		if err := EnsureAddonByName(s, add.name); err != nil {
			return err
		}
	}

	if err := cleanupAddons(s); err != nil {
		return err
	}

	return nil
}

// EnsureUserAddons deploys addons that are provided by the user and that are
// not embedded.
func EnsureUserAddons(s *state.State) error {
	applier, err := newAddonsApplier(s)
	if err != nil {
		return err
	}

	s.Logger.Infof("Applying user provided addons...")
	combinedAddons := map[string]string{}

	if applier.LocalFS != nil {
		customAddons, err := fs.ReadDir(applier.LocalFS, ".")
		if err != nil {
			return fail.Runtime(err, "reading local addons directory")
		}

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
	}

	for _, embeddedAddon := range s.Cluster.Addons.Addons {
		if _, ok := embeddedAddons[embeddedAddon.Name]; ok {
			continue
		}

		if embeddedAddon.Delete {
			if err := applier.loadAndDeleteAddon(s, applier.EmbeddedFS, embeddedAddon.Name); err != nil {
				return err
			}

			continue
		}

		if _, ok := combinedAddons[embeddedAddon.Name]; !ok {
			combinedAddons[embeddedAddon.Name] = ""
		}
	}

	for addonName := range combinedAddons {
		// NB: We can't migrate StorageClass when applying the CSI driver because
		// CSI driver is deployed only for Kubernetes 1.23+ clusters, but this
		// issue affects older clusters as well.
		if addonName == defaultStorageClassAddonName && s.Cluster.CloudProvider.GCE != nil {
			if err := migrateGCEStandardStorageClass(s); err != nil {
				return err
			}
		}
		if err := EnsureAddonByName(s, addonName); err != nil {
			return err
		}
	}

	if applier.LocalFS != nil {
		s.Logger.Info("Applying addons from the root directory...")
		if err := applier.loadAndApplyAddon(s, applier.LocalFS, ""); err != nil {
			return err
		}
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
			return fail.Runtime(lErr, "reading local addons directory")
		}

		for _, a := range addons {
			if !a.IsDir() {
				continue
			}
			if a.Name() == addonName {
				if err := applier.loadAndApplyAddon(s, applier.LocalFS, a.Name()); err != nil {
					return err
				}

				return nil
			}
		}
	}

	addons, eErr := fs.ReadDir(applier.EmbeddedFS, ".")
	if eErr != nil {
		return fail.Runtime(eErr, "reading embedded addons directory")
	}

	for _, a := range addons {
		if !a.IsDir() {
			continue
		}
		if a.Name() == addonName {
			if err := applier.loadAndApplyAddon(s, applier.EmbeddedFS, a.Name()); err != nil {
				return err
			}

			return nil
		}
	}

	return fail.RuntimeError{
		Op:  fmt.Sprintf("installing %q addon", addonName),
		Err: errors.New("addon does not exist"),
	}
}

// DeleteAddonByName deletes an addon by its name. It's required to keep the
// old addon manifest for this to work, however, it's enough to keep only
// metadata (i.e. spec is not needed). If the addon is not found in the addons
// directory, or if the addons are not enabled, it will search
// for the embedded addons.
func DeleteAddonByName(s *state.State, addonName string) error {
	applier, err := newAddonsApplier(s)
	if err != nil {
		return err
	}

	if applier.LocalFS != nil {
		addons, lErr := fs.ReadDir(applier.LocalFS, ".")
		if lErr != nil {
			return fail.Runtime(lErr, "reading local addons directory")
		}

		for _, a := range addons {
			if !a.IsDir() {
				continue
			}
			if a.Name() == addonName {
				if err := applier.loadAndDeleteAddon(s, applier.LocalFS, a.Name()); err != nil {
					return err
				}

				return nil
			}
		}
	}

	addons, eErr := fs.ReadDir(applier.EmbeddedFS, ".")
	if eErr != nil {
		return fail.Runtime(eErr, "reading embedded addons directory")
	}

	for _, a := range addons {
		if !a.IsDir() {
			continue
		}
		if a.Name() == addonName {
			if err := applier.loadAndDeleteAddon(s, applier.EmbeddedFS, a.Name()); err != nil {
				return err
			}

			return nil
		}
	}

	return fail.RuntimeError{
		Op:  fmt.Sprintf("installing %q addon", addonName),
		Err: errors.New("addon does not exist"),
	}
}

func ensureCSIAddons(s *state.State, addonsToDeploy []addonAction) []addonAction {
	k8sVersion := semver.MustParse(s.Cluster.Versions.Kubernetes)
	gte23 := greaterThan23.Check(k8sVersion)

	// We deploy available CSI drivers un-conditionally for k8s v1.23+
	//
	// CSIMigration, if applicable, for the cloud providers is turned on by default and requires installation of CSI drviers even if we
	// don't use external CCM. Although mount operations would fall-back to in-tree solution if CSI driver is not available. Fallback
	// for provision operations is NOT supported by in-tree solution.

	switch {
	// CSI driver is required for k8s v1.23+
	case s.Cluster.CloudProvider.AWS != nil && (gte23 || s.Cluster.CloudProvider.External):
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSIAwsEBS,
				supportFn: func() error {
					return migrateAWSCSIDriver(s)
				},
			},
		)
	// CSI driver is required for k8s v1.23+
	case s.Cluster.CloudProvider.Azure != nil && (gte23 || s.Cluster.CloudProvider.External):
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSIAzureDisk,
				supportFn: func() error {
					return migrateAzureDiskCSIDriver(s)
				},
			},
			addonAction{
				name: resources.AddonCSIAzureFile,
			},
		)
		// CSI driver is required for k8s v1.23+
	case s.Cluster.CloudProvider.GCE != nil && (gte23 || s.Cluster.CloudProvider.External):
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSIGCPComputePD,
			},
		)
	// Install CSI driver unconditionally
	case s.Cluster.CloudProvider.DigitalOcean != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSIDigitalOcean,
			},
		)
	// Install CSI driver unconditionally
	case s.Cluster.CloudProvider.Hetzner != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSIHetzner,
			},
		)
	// Install CSI driver unconditionally
	case s.Cluster.CloudProvider.Nutanix != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSINutanix,
			},
		)
	// Install CSI driver unconditionally
	case s.Cluster.CloudProvider.Openstack != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSIOpenStackCinder,
			},
		)
	// Install CSI driver unconditionally
	case s.Cluster.CloudProvider.VMwareCloudDirector != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSIVMwareCloudDirector,
			},
		)
	// Install CSI driver only if external cloud provider is used
	case s.Cluster.CloudProvider.Vsphere != nil && s.Cluster.CloudProvider.External:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCSIVsphere,
				supportFn: func() error {
					return removeCSIVsphereFromKubeSystem(s)
				},
			},
		)
	default:
		s.Logger.Infof("CSI driver for %q not yet supported, skipping", s.Cluster.CloudProvider.CloudProviderName())
	}

	return addonsToDeploy
}

func ensureCCMAddons(s *state.State, addonsToDeploy []addonAction) []addonAction {
	switch {
	case s.Cluster.CloudProvider.AWS != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCCMAws,
			},
		)
	case s.Cluster.CloudProvider.Azure != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCCMAzure,
			},
		)
	case s.Cluster.CloudProvider.DigitalOcean != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCCMDigitalOcean,
			},
		)
	case s.Cluster.CloudProvider.Hetzner != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCCMHetzner,
			},
		)
	case s.Cluster.CloudProvider.Openstack != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCCMOpenStack,
			},
		)

	case s.Cluster.CloudProvider.Vsphere != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCCMVsphere,
				supportFn: func() error {
					return migrateVsphereAddon(s)
				},
			},
		)
	case s.Cluster.CloudProvider.EquinixMetal != nil:
		addonsToDeploy = append(addonsToDeploy,
			addonAction{
				name: resources.AddonCCMEquinixMetal,
				supportFn: func() error {
					return migratePacketToEquinixCCM(s)
				},
			},
		)
	default:
		s.Logger.Infof("CCM driver for %q not yet supported, skipping", s.Cluster.CloudProvider.CloudProviderName())
	}

	return addonsToDeploy
}
