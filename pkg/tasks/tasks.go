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

package tasks

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/clusterstatus"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/features"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/localhelm"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/externalccm"
	"k8c.io/kubeone/pkg/templates/machinecontroller"
	"k8c.io/kubeone/pkg/templates/operatingsystemmanager"
)

type Tasks []Task

func (t Tasks) Run(s *state.State) error {
	for _, step := range t {
		if step.Predicate != nil && !step.Predicate(s) {
			continue
		}
		if err := step.Run(s); err != nil {
			return fail.RuntimeError{
				Op:  step.Operation,
				Err: errors.WithStack(err),
			}
		}
	}

	return nil
}

func (t Tasks) Descriptions(s *state.State) []string {
	var descriptions []string

	for _, step := range t {
		if step.Predicate != nil && !step.Predicate(s) {
			continue
		}
		if step.Description != "" {
			descriptions = append(descriptions, step.Description)
		}
	}

	return descriptions
}

func (t Tasks) append(newtasks ...Task) Tasks {
	return append(t, newtasks...)
}

func (t Tasks) prepend(newtasks ...Task) Tasks {
	return append(newtasks, t...)
}

// WithBinariesOnly will prepend passed tasks with tasks WithHostnameOS() and
// append install prerequisite binaries (docker, kubeadm, kubelet, etc...) on
// all hosts
func WithBinariesOnly(t Tasks) Tasks {
	return WithHostnameOSAndProbes(t).
		append(
			Task{Fn: installPrerequisites, Operation: "installing prerequisites"},
		)
}

// WithHostnameOS will prepend passed tasks with 2 basic tasks:
//   - detect OS on all cluster hosts
//   - detect hostnames  on all cluster hosts
func WithHostnameOS(t Tasks) Tasks {
	return t.prepend(
		Task{Fn: determineHostname, Operation: "detecting hostname"},
		Task{Fn: determineOS, Operation: "detecting OS"},
	)
}

// WithProbes will run different probes over the defined cluster
func WithProbes(t Tasks) Tasks {
	return t.append(
		Task{Fn: runProbes, Operation: "running probes"},
	)
}

func WithProbesAndSafeguard(t Tasks) Tasks {
	return t.append(
		Task{Fn: runProbes, Operation: "running probes"},
		Task{Fn: safeguard, Operation: "checking safeguards"},
	)
}

func WithHostnameOSAndProbes(t Tasks) Tasks {
	return WithProbesAndSafeguard(WithHostnameOS(t))
}

// WithFullInstall with install binaries (using WithBinariesOnly) and
// orchestrate complete cluster init
func WithFullInstall(t Tasks) Tasks {
	return WithHostnameOSAndProbes(t).append(Tasks{
		{
			Fn: func(s *state.State) error {
				return s.RunTaskOnAllNodes(disableNMCloudSetup, state.RunParallel)
			},
			Operation: "disabling nm-cloud-setup",
		},
		{
			Fn:        installPrerequisites,
			Operation: "installing prerequisites",
		},
	}...).
		append(kubernetesConfigFiles()...).
		append(Tasks{
			{
				Fn:        kubeadmPreflightChecks,
				Operation: "kubeadm preflight checks",
			},
			{Fn: prePullImages, Operation: "pre-pull images"},
			{
				Fn: func(s *state.State) error {
					s.Logger.Infoln("Configuring certs and etcd on control plane node...")

					return s.RunTaskOnLeader(kubeadmCertsExecutor)
				},
				Operation: "provisioning certificates on the leader",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Downloading PKI...")

					return s.RunTaskOnLeader(certificate.DownloadKubePKI)
				},
				Operation: "downloading Kubernetes PKI from the leader",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Uploading PKI...")

					return s.RunTaskOnFollowers(certificate.UploadKubePKI, state.RunParallel)
				},
				Operation: "uploading Kubernetes PKI",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Infoln("Configuring certs and etcd on consecutive control plane node...")

					return s.RunTaskOnFollowers(kubeadmCertsExecutor, state.RunParallel)
				},
				Operation: "provisioning certificates on the followers",
			},
			{Fn: initKubernetesLeader, Operation: "initializing kubernetes on leader"},
			{Fn: kubeconfig.BuildKubernetesClientset, Operation: "building kubernetes clientset"},
			{
				Fn: func(s *state.State) error {
					return s.RunTaskOnLeader(approvePendingCSR)
				},
				Operation: "approving leader's kubelet CSR",
			},
			{Fn: repairClusterIfNeeded, Operation: "repairing cluster"},
			{Fn: joinControlplaneNode, Operation: "joining followers control plane nodes"},
			{Fn: restartKubeAPIServer, Operation: "restarting unhealthy kube-apiserver"},
		}...).
		append(WithResources(nil)...).
		append(
			Task{
				// Node might emit one more CSR for kubelet serving certificates
				// after external CCM initializes the node. That's because
				// CCM modifies IP addresses in the Node object to properly set
				// private and public addresses, DNS names, etc...
				// To ensure that we approve those CSRs, we need to force kubelet
				// to generate new CSRs as soon as possible, and then approve
				// those new CSRs.
				// NB: We intentionally do this only on FullInstall because in
				// other cases we already have CCM deployed, so this is not
				// an issue. Additionally, we do this only for control plane
				// nodes because static workers are joined after the CCM is
				// deployed.
				Fn: func(s *state.State) error {
					if err := restartKubeletOnControlPlane(s); err != nil {
						return err
					}

					return s.RunTaskOnAllNodes(approvePendingCSR, true)
				},
				Operation: "removing old and approving new kubelet CSRs",
				Predicate: func(s *state.State) bool { return s.Cluster.CloudProvider.External },
			},
		).
		append(
			Task{
				Fn:        createMachineDeployments,
				Operation: "creating worker machines",
				Predicate: func(s *state.State) bool { return !s.LiveCluster.IsProvisioned() },
			},
		)
}

func WithResources(t Tasks) Tasks {
	return t.append(
		Tasks{
			{
				Fn: saveCABundle,
				Predicate: func(s *state.State) bool {
					return s.Cluster.CABundle != ""
				},
			},
			{
				Fn:        determinePauseImage,
				Operation: "determining the pause image",
			},
			{
				Fn:        patchStaticPods,
				Operation: "patching static pods",
			},
			{
				Fn:          renewControlPlaneCerts,
				Operation:   "renewing certificates",
				Description: "renew all certificates",
				Predicate: func(s *state.State) bool {
					return s.LiveCluster.CertsToExpireInLessThen90Days()
				},
			},
			{
				Fn:        saveKubeconfig,
				Operation: "saving kubeconfig",
			},
			{
				Fn:        removeSuperKubeconfig,
				Operation: "removing " + superAdminConfPath,
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Downloading PKI...")

					return s.RunTaskOnLeader(certificate.DownloadKubePKI)
				},
				Operation: "downloading Kubernetes PKI from the leader",
			},
			{
				Fn:        features.Activate,
				Operation: "activating features",
			},
			{
				Fn:        patchCoreDNS,
				Operation: "patching CoreDNS",
			},
			{
				Fn:          credentials.Ensure,
				Operation:   "ensuring credentials secret",
				Description: "ensure credential",
				Predicate:   func(s *state.State) bool { return s.Cluster.CloudProvider.SecretProviderClassName == "" },
			},
			{
				Fn:          ensureCABundleConfigMap,
				Operation:   "ensuring caBundle configMap",
				Description: "ensure caBundle configMap",
				Predicate:   func(s *state.State) bool { return s.Cluster.CABundle != "" },
			},
			{
				Fn:          labelNodes,
				Operation:   "labeling control-plane nodes",
				Description: "labeling control-plane nodes",
			},
			{
				Fn:          annotateNodes,
				Operation:   "annotating control-plane nodes",
				Description: "annotating control-plane nodes",
			},
			{
				Fn:          cleanupStaleObjects,
				Operation:   "cleaning up any leftovers from addons",
				Description: "clean up any leftovers from addons",
			},
			{
				Fn:          addons.Ensure,
				Operation:   "applying addons",
				Description: "ensure embedded addons",
			},
			{
				Fn:          addons.EnsureUserAddons,
				Operation:   "applying addons",
				Description: "ensure custom addons",
				Predicate:   func(s *state.State) bool { return s.Cluster.Addons != nil },
			},
			{
				Fn:        localhelm.Deploy,
				Operation: "releasing core helm charts",
			},
			{
				Fn:          externalccm.Ensure,
				Operation:   "ensuring external CCM",
				Description: "ensure external CCM",
				Predicate:   func(s *state.State) bool { return s.Cluster.CloudProvider.External },
			},
			{
				Fn:          ensureVsphereCSICABundleConfigMap,
				Operation:   "ensure vSphere CSI caBundle configMap",
				Description: "ensure vSphere CSI caBundle configMap",
				Predicate: func(s *state.State) bool {
					return s.Cluster.CABundle != "" && s.Cluster.CloudProvider.Vsphere != nil && s.Cluster.CloudProvider.External && !s.Cluster.CloudProvider.DisableBundledCSIDrivers
				},
			},
			{
				Fn:        joinStaticWorkerNodes,
				Operation: "joining static worker nodes to the cluster",
			},
			{
				Fn:          labelNodes,
				Operation:   "labeling nodes",
				Description: "labeling nodes",
			},
			{
				Fn:          annotateNodes,
				Operation:   "annotating nodes",
				Description: "annotating nodes",
			},
			{
				Fn:        fixFilePermissions,
				Operation: "Fix permissions of system files",
			},
			{
				Fn:        machinecontroller.WaitReady,
				Operation: "waiting for machine-controller",
			},
			{
				Fn:        operatingsystemmanager.WaitReady,
				Operation: "waiting for operating-system-manager",
				Predicate: func(s *state.State) bool { return s.Cluster.OperatingSystemManager.Deploy },
			},
			{
				Fn:          upgradeMachineDeployments,
				Operation:   "upgrading MachineDeployments",
				Description: "upgrade MachineDeployments",
				Predicate:   func(s *state.State) bool { return s.UpgradeMachineDeployments },
			},
		}...,
	)
}

func WithUpgrade(t Tasks, followers ...kubeoneapi.HostConfig) Tasks {
	return WithHostnameOSAndProbes(t).
		append(kubernetesConfigFiles()...). // this, in the upgrade process where config rails are handled
		append(Tasks{
			{Fn: kubeconfig.BuildKubernetesClientset, Operation: "building kubernetes clientset"},
			{Fn: uploadKubeadmToConfigMaps, Operation: "updating kubeadm configmaps"},
			{Fn: runPreflightChecks, Operation: "checking preflight safetynet", Retries: 1},
			{Fn: upgradeLeader, Operation: "upgrading leader control plane"},
		}...).
		append(generateUpgradeFollowersTasks(followers)...).
		append(Task{
			Fn: func(s *state.State) error {
				s.Logger.Info("Downloading PKI...")

				return s.RunTaskOnLeader(certificate.DownloadKubePKI)
			},
			Operation: "downloading Kubernetes PKI from the leader",
		}).
		append(WithResources(nil)...).
		append(
			Task{Fn: restartKubeAPIServer, Operation: "restarting unhealthy kube-apiserver"},
			Task{Fn: upgradeStaticWorkers, Operation: "upgrading static worker nodes"},
			Task{Fn: updateAllKubelets, Operation: "upgrading kubelets"},
			Task{
				Fn:          migratePVCAllocatedResourceStatus,
				Operation:   "migrating PVCs",
				Description: "migrate PVCs with AllocatedResourceStatuses",
				Predicate: func(s *state.State) bool {
					targetVersion, err := semver.NewVersion(s.Cluster.Versions.Kubernetes)
					if err != nil {
						// This should never ever happen because we validate the version.
						panic(err)
					}

					liveCP := s.LiveCluster.ControlPlane
					if len(liveCP) == 0 || liveCP[0].Kubelet.Version == nil {
						// This might only happen if the control plane doesn't exist,
						// but that's not the case when upgrading the cluster.
						return false
					}

					// Run operation only when upgrading to Kubernetes 1.31.
					return targetVersion.Minor() == 31 && liveCP[0].Kubelet.Version.Minor() == 30
				},
			},
			Task{
				Fn:          pruneImagesOnAllNodes,
				Operation:   "deleting unused container images",
				Description: "delete unused container images",
				Predicate:   func(s *state.State) bool { return s.PruneImages },
			},
		)
}

func WithReset(t Tasks) Tasks {
	return t.append(Tasks{
		{Fn: destroyWorkers, Operation: "destroying workers"},
		{Fn: resetAllNodes, Operation: "resetting all nodes"},
		{Fn: removeBinariesAllNodes, Operation: "removing kubernetes binaries from nodes"},
	}...)
}

func WithRemoveExtraEtcdMembers(t Tasks) Tasks {
	return t.append(Tasks{
		{Fn: repairClusterIfNeeded, Operation: "repairing cluster"},
	}...)
}

func WithContainerDMigration(t Tasks) Tasks {
	return WithHostnameOS(t).
		append(Tasks{
			{Fn: validateContainerdInConfig, Operation: "validating config", Retries: 1},
			{Fn: kubeconfig.BuildKubernetesClientset, Operation: "building kubernetes clientset"},
			{Fn: migrateToContainerd, Operation: "migrating to containerd"},
			{Fn: patchCRISocketAnnotation, Operation: "patching Node objects"},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Downloading PKI...")

					return s.RunTaskOnLeader(certificate.DownloadKubePKI)
				},
				Operation: "downloading Kubernetes PKI from the leader",
			},
			{
				Fn:          addons.Ensure,
				Operation:   "applying addons",
				Description: "ensure embedded addons",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Warn("Now please rolling restart your machineDeployments to get containerd")
					s.Logger.Warn("see more at: https://docs.kubermatic.com/kubeone/v1.10/cheat-sheets/rollout-machinedeployment/")

					return nil
				},
				Operation: "deploying machine-controller",
				Predicate: func(s *state.State) bool { return s.Cluster.MachineController.Deploy },
			},
		}...)
}

func WithClusterStatus(t Tasks) Tasks {
	return WithHostnameOS(t).
		append(Tasks{
			{Fn: kubeconfig.BuildKubernetesClientset, Operation: "building kubernetes clientset"},
			{Fn: clusterstatus.Print, Operation: "getting cluster status"},
		}...)
}

func kubernetesConfigFiles() Tasks {
	return Tasks{
		{Fn: generateKubeadm, Operation: "generating kubeadm config files"},
		{Fn: generateConfigurationFiles, Operation: "generating config files"},
		{Fn: uploadConfigurationFiles, Operation: "uploading config files"},
	}
}

func WithDisableEncryptionProviders(t Tasks, customConfig bool) Tasks {
	t = WithHostnameOSAndProbes(t)
	if customConfig {
		return t.append(Tasks{
			{
				Fn:          removeEncryptionProviderFile,
				Operation:   "removing encryption providers configuration",
				Description: "remove old Encryption Providers configuration file",
			},
			{
				Fn:          ensureRestartKubeAPIServer,
				Operation:   "restarting KubeAPI",
				Description: "restart KubeAPI containers",
			},

			{
				Fn:          rewriteClusterSecrets,
				Operation:   "rewriting cluster secrets",
				Description: "rewrite all cluster secrets",
			},
		}...)
	}

	return t.append(Tasks{
		{
			Fn:          fetchEncryptionProvidersFile,
			Operation:   "fetching EncryptionProviders config",
			Description: "fetch current Encryption Providers configuration file ",
		},
		{
			Fn:          uploadIdentityFirstEncryptionConfiguration,
			Operation:   "uploading encryption providers configuration",
			Description: "upload updated Encryption Providers configuration file",
		},
		{
			Fn:          ensureRestartKubeAPIServer,
			Operation:   "restarting kube-apiserver pods",
			Description: "restart KubeAPI containers",
		},
		{
			Fn:          rewriteClusterSecrets,
			Operation:   "rewriting cluster secrets",
			Description: "rewrite all cluster secrets",
		},
		{
			Fn:          removeEncryptionProviderFile,
			Operation:   "removing encryption providers configuration",
			Description: "remove old Encryption Providers configuration file",
		},
	}...)
}

func WithRewriteSecrets(t Tasks) Tasks {
	return t.append(
		Task{
			Fn:          rewriteClusterSecrets,
			Operation:   "rewriting cluster secrets",
			Description: "rewrite all cluster secrets",
		})
}

func WithCustomEncryptionConfigUpdated(t Tasks) Tasks {
	return t.append(Tasks{
		{
			Fn:          ensureRestartKubeAPIServer,
			Operation:   "restarting KubeAPI",
			Description: "restart KubeAPI containers",
		},
		{
			Fn:          rewriteClusterSecrets,
			Operation:   "rewriting cluster secrets",
			Description: "rewrite all cluster secrets",
		},
	}...)
}

func WithRotateKey(t Tasks) Tasks {
	return WithHostnameOSAndProbes(t).
		append(Tasks{
			{
				Fn:          fetchEncryptionProvidersFile,
				Operation:   "fetching EncryptionProviders config",
				Description: "fetch current Encryption Providers configuration file ",
			},
			{
				Fn:          uploadEncryptionConfigurationWithNewKey,
				Operation:   "uploading encryption providers configuration",
				Description: "upload updated Encryption Providers configuration file",
			},
			{
				Fn:          ensureRestartKubeAPIServer,
				Operation:   "restarting KubeAPI",
				Description: "restart KubeAPI containers",
			},
			{
				Fn:          rewriteClusterSecrets,
				Operation:   "rewriting cluster secrets",
				Description: "rewrite all cluster secrets",
			},
			{
				Fn:          uploadEncryptionConfigurationWithoutOldKey,
				Operation:   "uploading encryption providers configuration",
				Description: "upload updated Encryption Providers configuration file",
			},
			{
				Fn:          ensureRestartKubeAPIServer,
				Operation:   "restarting kube-apiserver pods",
				Description: "restart KubeAPI containers",
			},
		}...)
}

func WithCCMCSIMigration(t Tasks) Tasks {
	return t.append(Tasks{
		{Fn: ccmMigrationValidateConfig, Operation: "validating config", Retries: 1},
		{
			Fn:        readyToCompleteCCMMigration,
			Operation: "validating readiness to complete migration",
			Predicate: func(s *state.State) bool {
				return s.CCMMigrationComplete
			},
		},
		{Fn: generateKubeadm, Operation: "generating kubeadm config files"},
	}...).
		append(
			Task{Fn: ccmMigrationRegenerateControlPlaneManifestsAndKubeletConfig, Operation: "regenerating static pod manifests and kubelet config"},
			Task{
				Fn:        ccmMigrationUpdateStaticWorkersKubeletConfig,
				Operation: "updating kubelet config on static worker nodes",
				Predicate: func(s *state.State) bool {
					return len(s.Cluster.StaticWorkers.Hosts) > 0
				},
			},
		).
		append(WithResources(nil)...).
		append(
			// Regenerate files only when finishing the CCM/CSI migration.
			Task{
				Fn:        generateConfigurationFiles,
				Operation: "generating config files",
				Predicate: func(s *state.State) bool {
					return s.CCMMigrationComplete
				},
			},
			Task{
				Fn:        uploadConfigurationFiles,
				Operation: "uploading config files",
				Predicate: func(s *state.State) bool {
					return s.CCMMigrationComplete
				},
			},
			Task{
				Fn:        migrateOpenStackPVs,
				Operation: "migrating openstack persistentvolumes",
				Predicate: func(s *state.State) bool { return s.Cluster.CloudProvider.Openstack != nil },
			},
			Task{
				Fn: func(s *state.State) error {
					s.Logger.Warn("Now please rolling restart your machineDeployments to migrate to ccm/csi")
					s.Logger.Warn("see more at: https://docs.kubermatic.com/kubeone/v1.10/cheat-sheets/rollout-machinedeployment/")
					s.Logger.Warn("Once you're done, please run this command again with the '--complete' flag to finish migration")

					return nil
				},
				Operation: "show next steps",
				Predicate: func(s *state.State) bool { return s.Cluster.MachineController.Deploy && !s.CCMMigrationComplete },
			},
		)
}

func updateAllKubelets(s *state.State) error {
	return s.RunTaskOnAllNodes(func(s *state.State, node *kubeoneapi.HostConfig, _ executor.Interface) error {
		logger := s.Logger.WithField("node", node.PublicAddress)

		logger.Infoln("Upgrading kubelet")
		if err := upgradeKubernetesBinaries(s, *node, scripts.Params{Kubelet: true}); err != nil {
			return err
		}

		sleep := 30 * time.Second
		s.Logger.Infof("Sleeping %s seconds, giving time for kubeapi server to restart", sleep)
		time.Sleep(sleep)

		return nil
	}, state.RunSequentially)
}
