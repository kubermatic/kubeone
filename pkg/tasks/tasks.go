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
	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/addons"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/clusterstatus"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/features"
	"k8c.io/kubeone/pkg/kubeconfig"
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
			return errors.Wrap(err, step.ErrMsg)
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
			Task{Fn: installPrerequisites, ErrMsg: "failed to install prerequisites"},
		)
}

// WithHostnameOS will prepend passed tasks with 2 basic tasks:
//  * detect OS on all cluster hosts
//  * detect hostnames  on all cluster hosts
func WithHostnameOS(t Tasks) Tasks {
	return t.prepend(
		Task{Fn: determineHostname, ErrMsg: "failed to detect hostname"},
		Task{Fn: determineOS, ErrMsg: "failed to detect OS"},
	)
}

// WithProbes will run different probes over the defined cluster
func WithProbes(t Tasks) Tasks {
	return t.append(
		Task{Fn: runProbes, ErrMsg: "probes failed"},
	)
}

func WithProbesAndSafeguard(t Tasks) Tasks {
	return t.append(
		Task{Fn: runProbes, ErrMsg: "probes failed"},
		Task{Fn: safeguard, ErrMsg: "probes analysis failed"},
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
			ErrMsg: "failed to disable nm-cloud-setup",
		},
		{
			Fn:     installPrerequisites,
			ErrMsg: "failed to install prerequisites",
		},
	}...).
		append(kubernetesConfigFiles()...).
		append(Tasks{
			{
				Fn: func(s *state.State) error {
					s.Logger.Infoln("Configuring certs and etcd on control plane node...")

					return s.RunTaskOnLeader(kubeadmCertsExecutor)
				},
				ErrMsg: "failed to provision certs and etcd on leader",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Downloading PKI...")

					return s.RunTaskOnLeader(certificate.DownloadKubePKI)
				},
				ErrMsg: "failed to download Kubernetes PKI from the leader",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Uploading PKI...")

					return s.RunTaskOnFollowers(certificate.UploadKubePKI, state.RunParallel)
				},
				ErrMsg: "failed to upload Kubernetes PKI",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Infoln("Configuring certs and etcd on consecutive control plane node...")

					return s.RunTaskOnFollowers(kubeadmCertsExecutor, state.RunParallel)
				},
				ErrMsg: "failed to provision certs and etcd on followers",
			},
			{Fn: initKubernetesLeader, ErrMsg: "failed to init kubernetes on leader"},
			{Fn: kubeconfig.BuildKubernetesClientset, ErrMsg: "failed to build kubernetes clientset"},
			{
				Fn: func(s *state.State) error {
					return s.RunTaskOnLeader(approvePendingCSR)
				},
				ErrMsg: "failed to approve leader's kubelet CSR",
			},
			{Fn: repairClusterIfNeeded, ErrMsg: "failed to repair cluster"},
			{Fn: joinControlplaneNode, ErrMsg: "failed to join other masters a cluster"},
			{Fn: restartKubeAPIServer, ErrMsg: "failed to restart unhealthy kube-apiserver"},
		}...).
		append(WithResources(nil)...).
		append(
			Task{
				Fn:        createMachineDeployments,
				ErrMsg:    "failed to create worker machines",
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
				Fn:     patchStaticPods,
				ErrMsg: "failed to patch static pods",
			},
			{
				Fn:          renewControlPlaneCerts,
				ErrMsg:      "failed to renew certificates",
				Description: "renew all certificates",
				Predicate: func(s *state.State) bool {
					return s.LiveCluster.CertsToExpireInLessThen90Days()
				},
			},
			{
				Fn:     saveKubeconfig,
				ErrMsg: "failed to save kubeconfig to the local machine",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Downloading PKI...")

					return s.RunTaskOnLeader(certificate.DownloadKubePKI)
				},
				ErrMsg: "failed to download Kubernetes PKI from the leader",
			},
			{
				Fn:     features.Activate,
				ErrMsg: "failed to activate features",
			},
			{
				Fn:     patchCoreDNS,
				ErrMsg: "failed to patch CoreDNS",
			},
			{
				Fn:          addons.Ensure,
				ErrMsg:      "failed to apply addons",
				Description: "ensure embedded addons",
			},
			{
				Fn:          ensureCNI,
				ErrMsg:      "failed to install cni plugin",
				Description: "ensure CNI",
				Predicate:   func(s *state.State) bool { return s.Cluster.ClusterNetwork.CNI.External == nil },
			},
			{
				Fn:          ensureCABundleConfigMap,
				ErrMsg:      "failed to ensure caBundle configMap",
				Description: "ensure caBundle configMap",
				Predicate:   func(s *state.State) bool { return s.Cluster.CABundle != "" },
			},
			{
				Fn:          addons.EnsureUserAddons,
				ErrMsg:      "failed to apply addons",
				Description: "ensure custom addons",
				Predicate:   func(s *state.State) bool { return s.Cluster.Addons != nil && s.Cluster.Addons.Enable },
			},
			{
				Fn:          credentials.Ensure,
				ErrMsg:      "failed to ensure credentials secret",
				Description: "ensure credential",
			},
			{
				Fn:          externalccm.Ensure,
				ErrMsg:      "failed to ensure external CCM",
				Description: "ensure external CCM",
				Predicate:   func(s *state.State) bool { return s.Cluster.CloudProvider.External },
			},
			{
				Fn:     joinStaticWorkerNodes,
				ErrMsg: "failed to join worker nodes to the cluster",
			},
			{
				Fn:     labelNodeOSes,
				ErrMsg: "failed to label nodes with their OS",
			},
			{
				Fn:     machinecontroller.WaitReady,
				ErrMsg: "failed to wait for machine-controller",
			},
			{
				Fn:        operatingsystemmanager.WaitReady,
				ErrMsg:    "failed to wait for operating-system-manager",
				Predicate: func(s *state.State) bool { return s.Cluster.OperatingSystemManagerEnabled() },
			},
		}...,
	)
}

func WithUpgrade(t Tasks) Tasks {
	return WithHostnameOSAndProbes(t).
		append(kubernetesConfigFiles()...). // this, in the upgrade process where config rails are handled
		append(Tasks{
			{Fn: kubeconfig.BuildKubernetesClientset, ErrMsg: "failed to build kubernetes clientset"},
			{Fn: runPreflightChecks, ErrMsg: "preflight checks failed", Retries: 1},
			{Fn: upgradeLeader, ErrMsg: "failed to upgrade leader control plane"},
			{Fn: upgradeFollower, ErrMsg: "failed to upgrade follower control plane"},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Downloading PKI...")

					return s.RunTaskOnLeader(certificate.DownloadKubePKI)
				},
				ErrMsg: "failed to download Kubernetes PKI from the leader",
			},
		}...).
		append(WithResources(nil)...).
		append(
			Task{Fn: restartKubeAPIServer, ErrMsg: "failed to restart unhealthy kube-apiserver"},
			Task{Fn: upgradeStaticWorkers, ErrMsg: "unable to upgrade static worker nodes"},
			Task{
				Fn:          upgradeMachineDeployments,
				ErrMsg:      "failed to upgrade MachineDeployments",
				Description: "upgrade MachineDeployments",
				Predicate:   func(s *state.State) bool { return s.UpgradeMachineDeployments },
			},
		)
}

func WithReset(t Tasks) Tasks {
	return t.append(Tasks{
		{Fn: destroyWorkers, ErrMsg: "failed to destroy workers"},
		{Fn: resetAllNodes, ErrMsg: "failed to reset nodes"},
		{Fn: removeBinariesAllNodes, ErrMsg: "failed to remove binaries from nodes"},
	}...)
}

func WithContainerDMigration(t Tasks) Tasks {
	return WithHostnameOS(t).
		append(Tasks{
			{Fn: validateContainerdInConfig, ErrMsg: "failed to validate config", Retries: 1},
			{Fn: kubeconfig.BuildKubernetesClientset, ErrMsg: "failed to build kubernetes clientset"},
			{Fn: migrateToContainerd, ErrMsg: "failed to migrate to containerd"},
			{Fn: patchCRISocketAnnotation, ErrMsg: "failed to patch Node objects"},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Downloading PKI...")

					return s.RunTaskOnLeader(certificate.DownloadKubePKI)
				},
				ErrMsg: "failed to download Kubernetes PKI from the leader",
			},
			{
				Fn:          addons.Ensure,
				ErrMsg:      "failed to apply addons",
				Description: "ensure embedded addons",
			},
			{
				Fn: func(s *state.State) error {
					s.Logger.Warn("Now please rolling restart your machineDeployments to get containerd")
					s.Logger.Warn("see more at: https://docs.kubermatic.com/kubeone/v1.3/cheat_sheets/rollout_machinedeployment/")

					return nil
				},
				Predicate: func(s *state.State) bool { return s.Cluster.MachineController.Deploy },
			},
		}...)
}

func WithClusterStatus(t Tasks) Tasks {
	return WithHostnameOS(t).
		append(Tasks{
			{Fn: kubeconfig.BuildKubernetesClientset, ErrMsg: "failed to build kubernetes clientset"},
			{Fn: clusterstatus.Print, ErrMsg: "failed to get cluster status"},
		}...)
}

func kubernetesConfigFiles() Tasks {
	return Tasks{
		{Fn: generateKubeadm, ErrMsg: "failed to generate kubeadm config files"},
		{Fn: generateConfigurationFiles, ErrMsg: "failed to generate config files"},
		{Fn: uploadConfigurationFiles, ErrMsg: "failed to upload config files"},
	}
}

func WithDisableEncryptionProviders(t Tasks, customConfig bool) Tasks {
	t = WithHostnameOSAndProbes(t)
	if customConfig {
		return t.append(Tasks{
			{
				Fn:          removeEncryptionProviderFile,
				ErrMsg:      "failed to remove encryption providers configuration",
				Description: "remove old Encryption Providers configuration file",
			},
			{
				Fn:          ensureRestartKubeAPIServer,
				ErrMsg:      "failed to restart KubeAPI",
				Description: "restart KubeAPI containers",
			},

			{
				Fn:          rewriteClusterSecrets,
				ErrMsg:      "failed to rewrite cluster secrets",
				Description: "rewrite all cluster secrets",
			},
		}...)
	}

	return t.append(Tasks{
		{
			Fn:          fetchEncryptionProvidersFile,
			ErrMsg:      "failed to fetch EncryptionProviders config",
			Description: "fetch current Encryption Providers configuration file "},
		{
			Fn:          uploadIdentityFirstEncryptionConfiguration,
			ErrMsg:      "failed to upload encryption providers configuration",
			Description: "upload updated Encryption Providers configuration file"},
		{
			Fn:          ensureRestartKubeAPIServer,
			ErrMsg:      "failed to restart KubeAPI",
			Description: "restart KubeAPI containers",
		},
		{
			Fn:          rewriteClusterSecrets,
			ErrMsg:      "failed to rewrite cluster secrets",
			Description: "rewrite all cluster secrets",
		},
		{
			Fn:          removeEncryptionProviderFile,
			ErrMsg:      "failed to remove encryption providers configuration",
			Description: "remove old Encryption Providers configuration file",
		},
	}...)
}

func WithRewriteSecrets(t Tasks) Tasks {
	return t.append(
		Task{
			Fn:          rewriteClusterSecrets,
			ErrMsg:      "failed to rewrite cluster secrets",
			Description: "rewrite all cluster secrets",
		})
}

func WithCustomEncryptionConfigUpdated(t Tasks) Tasks {
	return t.append(Tasks{
		{
			Fn:          ensureRestartKubeAPIServer,
			ErrMsg:      "failed to restart KubeAPI",
			Description: "restart KubeAPI containers",
		},
		{
			Fn:          rewriteClusterSecrets,
			ErrMsg:      "failed to rewrite cluster secrets",
			Description: "rewrite all cluster secrets",
		},
	}...)
}

func WithRotateKey(t Tasks) Tasks {
	return WithHostnameOSAndProbes(t).
		append(Tasks{
			{
				Fn:          fetchEncryptionProvidersFile,
				ErrMsg:      "failed to fetch EncryptionProviders config",
				Description: "fetch current Encryption Providers configuration file ",
			},
			{
				Fn:          uploadEncryptionConfigurationWithNewKey,
				ErrMsg:      "failed to upload encryption providers configuration",
				Description: "upload updated Encryption Providers configuration file",
			},
			{
				Fn:          ensureRestartKubeAPIServer,
				ErrMsg:      "failed to restart KubeAPI",
				Description: "restart KubeAPI containers",
			},
			{
				Fn:          rewriteClusterSecrets,
				ErrMsg:      "failed to rewrite cluster secrets",
				Description: "rewrite all cluster secrets",
			},
			{
				Fn:          uploadEncryptionConfigurationWithoutOldKey,
				ErrMsg:      "failed to upload encryption providers configuration",
				Description: "upload updated Encryption Providers configuration file",
			},
			{
				Fn:          ensureRestartKubeAPIServer,
				ErrMsg:      "failed to restart KubeAPI",
				Description: "restart KubeAPI containers",
			},
		}...)
}

func WithCCMCSIMigration(t Tasks) Tasks {
	return t.append(Tasks{
		{Fn: ccmMigrationValidateConfig, ErrMsg: "failed to validate config", Retries: 1},
		{
			Fn:     readyToCompleteCCMMigration,
			ErrMsg: "failed to validate readiness to complete migration",
			Predicate: func(s *state.State) bool {
				return s.CCMMigrationComplete
			},
		},
	}...).
		append(kubernetesConfigFiles()...).
		append(
			Task{Fn: ccmMigrationRegenerateControlPlaneManifests, ErrMsg: "failed to regenerate static pod manifests"},
			Task{Fn: ccmMigrationUpdateControlPlaneKubeletConfig, ErrMsg: "failed to update kubelet config on control plane nodes"},
			Task{
				Fn:     ccmMigrationUpdateStaticWorkersKubeletConfig,
				ErrMsg: "failed to update kubelet config on static worker nodes",
				Predicate: func(s *state.State) bool {
					return len(s.Cluster.StaticWorkers.Hosts) > 0
				},
			},
		).
		append(WithResources(nil)...).
		append(
			Task{
				Fn:        migrateOpenStackPVs,
				ErrMsg:    "failed to migrate openstack persistentvolumes",
				Predicate: func(s *state.State) bool { return s.Cluster.CloudProvider.Openstack != nil },
			},
			Task{
				Fn: func(s *state.State) error {
					s.Logger.Warn("Now please rolling restart your machineDeployments to migrate to ccm/csi")
					s.Logger.Warn("see more at: https://docs.kubermatic.com/kubeone/v1.3/cheat_sheets/rollout_machinedeployment/")
					s.Logger.Warn("Once you're done, please run this command again with the '--complete' flag to finish migration")

					return nil
				},
				ErrMsg:    "failed to show next steps",
				Predicate: func(s *state.State) bool { return s.Cluster.MachineController.Deploy && !s.CCMMigrationComplete },
			},
		)
}
