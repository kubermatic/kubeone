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
	"k8c.io/kubeone/pkg/addons"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/clusterstatus"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
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
			return fail.Runtime(err, step.Operation)
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
//  * detect OS on all cluster hosts
//  * detect hostnames  on all cluster hosts
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
			},
			{
				Fn:          addons.Ensure,
				Operation:   "applying addons",
				Description: "ensure embedded addons",
			},
			{
				Fn:          ensureCNI,
				Operation:   "installing CNI plugin",
				Description: "ensure CNI",
				Predicate:   func(s *state.State) bool { return s.Cluster.ClusterNetwork.CNI.External == nil },
			},
			{
				Fn:          ensureCABundleConfigMap,
				Operation:   "ensuring caBundle configMap",
				Description: "ensure caBundle configMap",
				Predicate:   func(s *state.State) bool { return s.Cluster.CABundle != "" },
			},
			{
				Fn:          addons.EnsureUserAddons,
				Operation:   "applying addons",
				Description: "ensure custom addons",
				Predicate:   func(s *state.State) bool { return s.Cluster.Addons != nil && s.Cluster.Addons.Enable },
			},
			{
				Fn:          externalccm.Ensure,
				Operation:   "ensuring external CCM",
				Description: "ensure external CCM",
				Predicate:   func(s *state.State) bool { return s.Cluster.CloudProvider.External },
			},
			{
				Fn:        joinStaticWorkerNodes,
				Operation: "joining static worker nodes to the cluster",
			},
			{
				Fn:        labelNodes,
				Operation: "labeling nodes",
			},
			{
				Fn:        machinecontroller.WaitReady,
				Operation: "waiting for machine-controller",
			},
			{
				Fn:        operatingsystemmanager.WaitReady,
				Operation: "waiting for operating-system-manager",
				Predicate: func(s *state.State) bool { return s.Cluster.OperatingSystemManagerEnabled() },
			},
		}...,
	)
}

func WithUpgrade(t Tasks) Tasks {
	return WithHostnameOSAndProbes(t).
		append(kubernetesConfigFiles()...). // this, in the upgrade process where config rails are handled
		append(Tasks{
			{Fn: kubeconfig.BuildKubernetesClientset, Operation: "building kubernetes clientset"},
			{Fn: runPreflightChecks, Operation: "checking preflight safetynet", Retries: 1},
			{Fn: upgradeLeader, Operation: "upgrading leader control plane"},
			{Fn: upgradeFollower, Operation: "upgrading follower control plane"},
			{
				Fn: func(s *state.State) error {
					s.Logger.Info("Downloading PKI...")

					return s.RunTaskOnLeader(certificate.DownloadKubePKI)
				},
				Operation: "downloading Kubernetes PKI from the leader",
			},
		}...).
		append(WithResources(nil)...).
		append(
			Task{Fn: restartKubeAPIServer, Operation: "restarting unhealthy kube-apiserver"},
			Task{Fn: upgradeStaticWorkers, Operation: "upgrading static worker nodes"},
			Task{
				Fn:          upgradeMachineDeployments,
				Operation:   "upgrading MachineDeployments",
				Description: "upgrade MachineDeployments",
				Predicate:   func(s *state.State) bool { return s.UpgradeMachineDeployments },
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
					s.Logger.Warn("see more at: https://docs.kubermatic.com/kubeone/v1.4/cheat_sheets/rollout_machinedeployment/")

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
			Description: "fetch current Encryption Providers configuration file "},
		{
			Fn:          uploadIdentityFirstEncryptionConfiguration,
			Operation:   "uploading encryption providers configuration",
			Description: "upload updated Encryption Providers configuration file"},
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
	}...).
		append(kubernetesConfigFiles()...).
		append(
			Task{Fn: ccmMigrationRegenerateControlPlaneManifests, Operation: "regenerating static pod manifests"},
			Task{Fn: ccmMigrationUpdateControlPlaneKubeletConfig, Operation: "updating kubelet config on control plane nodes"},
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
			Task{
				Fn:        migrateOpenStackPVs,
				Operation: "migrating openstack persistentvolumes",
				Predicate: func(s *state.State) bool { return s.Cluster.CloudProvider.Openstack != nil },
			},
			Task{
				Fn: func(s *state.State) error {
					s.Logger.Warn("Now please rolling restart your machineDeployments to migrate to ccm/csi")
					s.Logger.Warn("see more at: https://docs.kubermatic.com/kubeone/v1.4/cheat_sheets/rollout_machinedeployment/")
					s.Logger.Warn("Once you're done, please run this command again with the '--complete' flag to finish migration")

					return nil
				},
				Operation: "show next steps",
				Predicate: func(s *state.State) bool { return s.Cluster.MachineController.Deploy && !s.CCMMigrationComplete },
			},
		)
}
