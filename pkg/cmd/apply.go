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

package cmd

import (
	"fmt"
	"reflect"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tasks"

	apiserverconfigv1 "k8s.io/apiserver/pkg/apis/config/v1"
	kyaml "sigs.k8s.io/yaml"
)

type applyOpts struct {
	globalOptions
	AutoApprove bool `longflag:"auto-approve" shortflag:"y"`
	// Install flags
	BackupFile   string `longflag:"backup" shortflag:"b"`
	NoInit       bool   `longflag:"no-init"`
	ForceInstall bool   `longflag:"force-install"`
	// Upgrade flags
	ForceUpgrade              bool `longflag:"force-upgrade"`
	UpgradeMachineDeployments bool `longflag:"upgrade-machine-deployments"`
	CreateMachineDeployments  bool `longflag:"create-machine-deployments"`
	RotateEncryptionKey       bool `longflag:"rotate-encryption-key"`
}

func (opts *applyOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, err
	}

	s.BackupFile = defaultBackupPath(opts.BackupFile, opts.ManifestFile, s.Cluster.Name)
	s.ForceInstall = opts.ForceInstall
	s.ForceUpgrade = opts.ForceUpgrade
	s.UpgradeMachineDeployments = opts.UpgradeMachineDeployments
	s.CreateMachineDeployments = opts.CreateMachineDeployments

	return s, initBackup(s.BackupFile)
}

func applyCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &applyOpts{}

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Reconcile the cluster",
		Long: heredoc.Doc(`
			Reconcile (Install/Upgrade/Repair/Restore) Kubernetes cluster on pre-existing machines. MachineDeployments get
			initialized but won't get modified by default, see '--upgrade-machine-deployments'.

			This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
			It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.
		`),
		SilenceErrors: true,
		Example:       `kubeone apply -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			opts.globalOptions = *gopts
			st, err := opts.BuildState()
			if err != nil {
				return err
			}

			return runApply(st, opts)
		},
	}

	cmd.Flags().BoolVarP(
		&opts.AutoApprove,
		longFlagName(opts, "AutoApprove"),
		shortFlagName(opts, "AutoApprove"),
		false,
		"auto approve plan")

	cmd.Flags().StringVarP(
		&opts.BackupFile,
		longFlagName(opts, "BackupFile"),
		shortFlagName(opts, "BackupFile"),
		"",
		"path to where the PKI backup .tar.gz file should be placed (default: location of cluster config file)")

	cmd.Flags().BoolVar(
		&opts.NoInit,
		longFlagName(opts, "NoInit"),
		false,
		"don't initialize the cluster (only install binaries)")

	cmd.Flags().BoolVar(
		&opts.ForceInstall,
		longFlagName(opts, "ForceInstall"),
		false,
		"use force to install new binary versions (!dangerous!)")

	cmd.Flags().BoolVar(
		&opts.ForceUpgrade,
		longFlagName(opts, "ForceUpgrade"),
		false,
		"force start upgrade process")

	cmd.Flags().BoolVar(
		&opts.UpgradeMachineDeployments,
		longFlagName(opts, "UpgradeMachineDeployments"),
		false,
		"upgrade MachineDeployments objects")

	cmd.Flags().BoolVar(
		&opts.CreateMachineDeployments,
		longFlagName(opts, "CreateMachineDeployments"),
		true,
		"create MachineDeployments objects")

	cmd.Flags().BoolVar(
		&opts.RotateEncryptionKey,
		longFlagName(opts, "RotateEncryptionKey"),
		false,
		"rotate Encryption Provider encryption key")

	return cmd
}

func runApply(st *state.State, opts *applyOpts) error {
	// Validate credentials
	if err := validateCredentials(st, opts.CredentialsFile); err != nil {
		return err
	}

	// Probe the cluster for the actual state and the needed tasks.
	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbesAndSafeguard(probbing)

	if err := probbing.Run(st); err != nil {
		return err
	}

	if st.Verbose {
		// Print information about hosts collected by probes
		for _, host := range st.LiveCluster.ControlPlane {
			printHostInformation(host)
		}

		for _, host := range st.LiveCluster.StaticWorkers {
			printHostInformation(host)
		}
	}

	// Reconcile the cluster based on the probe status
	if !st.LiveCluster.IsProvisioned() {
		return runApplyInstall(st, opts)
	}

	if !st.LiveCluster.Healthy() {
		if opts.RotateEncryptionKey {
			return fail.RuntimeError{
				Op:  "checking encryption key rotation",
				Err: errors.New("cluster is not healthy, encryption key rotation is not supported"),
			}
		}

		brokenHosts := st.LiveCluster.BrokenHosts()
		if len(brokenHosts) > 0 {
			for _, node := range brokenHosts {
				st.Logger.Errorf("Host %q is broken and needs to be manually removed\n", node)
			}

			st.Logger.Warnf("Hosts must be removed in a correct order to preserve the Etcd quorum.")
			st.Logger.Warnf("Loss of the Etcd quorum can cause loss of all data!!!")
			st.Logger.Warnf("After removing the recommended hosts, run 'kubeone apply' before removing any other host.")

			safeToDelete := st.LiveCluster.SafeToDeleteHosts()
			if len(safeToDelete) > 0 {
				st.Logger.Warnf("The recommended removal order:")
				for _, safe := range safeToDelete {
					st.Logger.Warnf("- %q", safe)
				}
			} else {
				st.Logger.Warnf("No other broken node can be removed without losing quorum.")
			}
		}

		runRepair := false
		for _, node := range st.LiveCluster.ControlPlane {
			if !node.IsInCluster {
				runRepair = true

				break
			}
		}

		if !runRepair {
			for _, node := range st.LiveCluster.StaticWorkers {
				if !node.IsInCluster {
					runRepair = true

					break
				}
			}
		}

		if safeRepair, higherVer := st.LiveCluster.SafeToRepair(st.Cluster.Versions.Kubernetes); !safeRepair {
			st.Logger.Errorln("Repair and upgrade are not supported at the same time!")
			st.Logger.Warnf("Requested version: %s\n", st.Cluster.Versions.Kubernetes)
			st.Logger.Warnf("Highest version: %s\n", higherVer)
			st.Logger.Warnf("Use version %s to repair the cluster, then run apply with the new version\n", higherVer)

			return fail.ConfigValidation(fmt.Errorf("repair and upgrade are not supported at the same time"))
		}

		if runRepair {
			return runApplyInstall(st, opts)
		}

		if len(brokenHosts) > 0 {
			return fail.NewConfigError("broken hosts check", "broken host(s) found, remove it manually")
		}

		return nil
	}

	if opts.RotateEncryptionKey {
		if !st.EncryptionEnabled() {
			return fail.ConfigValidation(fmt.Errorf("encryption Providers support is not enabled for this cluster"))
		}

		if st.Cluster.Features.EncryptionProviders != nil &&
			st.Cluster.Features.EncryptionProviders.CustomEncryptionConfiguration != "" {
			return fail.ConfigValidation(fmt.Errorf("key rotation of custom providers file is not supported"))
		}

		return runApplyRotateKey(st, opts)
	}

	return runApplyUpgradeIfNeeded(st, opts)
}

func runApplyInstall(s *state.State, opts *applyOpts) error { // Print the expected changes
	fmt.Println("The following actions will be taken: ")
	fmt.Println("Run with --verbose flag for more information.")

	for _, node := range s.LiveCluster.ControlPlane {
		if !node.IsInCluster {
			if node.Config.IsLeader {
				fmt.Printf("\t+ initialize control plane node %q (%s) using %s\n", node.Config.Hostname, node.Config.PrivateAddress, s.Cluster.Versions.Kubernetes)
			} else {
				fmt.Printf("\t+ join control plane node %q (%s) using %s\n", node.Config.Hostname, node.Config.PrivateAddress, s.Cluster.Versions.Kubernetes)
			}
		}
	}

	for _, node := range s.LiveCluster.StaticWorkers {
		if !node.IsInCluster {
			fmt.Printf("\t+ join static worker node %q (%s)\n", node.Config.Hostname, node.Config.PrivateAddress)
		}
	}

	if opts.NoInit {
		fmt.Println("\t! NoInit option provided: only binaries will be installed")
	}

	if opts.ForceInstall {
		fmt.Println("\t! force-install option provided: force install new binary versions (!dangerous!)")
	}

	if !s.LiveCluster.IsProvisioned() {
		for _, node := range s.Cluster.DynamicWorkers {
			fmt.Printf("\t+ ensure machinedeployment %q with %d replica(s) exists\n", node.Name, resolveInt(node.Replicas))
		}
	}

	if s.Cluster.Addons.Enabled() && s.Cluster.Addons.Path != "" {
		fmt.Printf("\t+ apply embedded and custom addons defined in %q\n", s.Cluster.Addons.Path)
	} else if s.Cluster.Addons.Enabled() {
		fmt.Print("\t+ apply embedded addons")
	}

	fmt.Println()
	confirm, err := confirmCommand(opts.AutoApprove)
	if err != nil {
		return err
	}

	if !confirm {
		s.Logger.Println("Operation canceled.")

		return nil
	}

	if opts.NoInit {
		return tasks.WithBinariesOnly(nil).Run(s)
	}

	return tasks.WithFullInstall(nil).Run(s)
}

func runApplyUpgradeIfNeeded(s *state.State, opts *applyOpts) error {
	fmt.Println("The following actions will be taken: ")
	if !opts.Verbose {
		fmt.Println("Run with --verbose flag for more information.")
	}

	upgradeNeeded, err := s.LiveCluster.UpgradeNeeded()
	if err != nil {
		s.Logger.Errorf("Upgrade not allowed: %v\n", err)

		return err
	}

	operations := []string{}

	var tasksToRun tasks.Tasks

	if upgradeNeeded || opts.ForceUpgrade {
		// disable case, we do this as early as possible.
		if s.ShouldDisableEncryption() {
			tasksToRun = tasks.WithDisableEncryptionProviders(tasksToRun, s.LiveCluster.EncryptionConfiguration.Custom)
		}

		tasksToRun = tasks.WithUpgrade(tasksToRun)

		if s.ShouldEnableEncryption() {
			operations = append(operations, "enable Encryption Provider support")
			tasksToRun = tasks.WithRewriteSecrets(tasksToRun)
		}

		// custom encryption configuration was modified
		if s.LiveCluster.CustomEncryptionEnabled() &&
			s.Cluster.Features.EncryptionProviders != nil &&
			s.Cluster.Features.EncryptionProviders.CustomEncryptionConfiguration != "" {
			config := &apiserverconfigv1.EncryptionConfiguration{}
			err = kyaml.UnmarshalStrict([]byte(s.Cluster.Features.EncryptionProviders.CustomEncryptionConfiguration), config)
			if err != nil {
				return err
			}

			if !reflect.DeepEqual(config, s.LiveCluster.EncryptionConfiguration.Config) {
				operations = append(operations, []string{"update Encryption Provider configuration", "restart KubeAPI"}...)
				tasksToRun = tasks.WithCustomEncryptionConfigUpdated(tasksToRun)
			}
		}

		for _, node := range s.LiveCluster.ControlPlane {
			forceFlag := ""
			if opts.ForceUpgrade {
				forceFlag = "force "
			}

			operations = append(operations,
				fmt.Sprintf("%supgrade control plane node %q (%s): %s -> %s",
					forceFlag,
					node.Config.Hostname,
					node.Config.PrivateAddress,
					node.Kubelet.Version,
					s.Cluster.Versions.Kubernetes))
		}

		for _, node := range s.LiveCluster.StaticWorkers {
			forceFlag := ""
			if opts.ForceUpgrade {
				forceFlag = "force "
			}
			operations = append(operations,
				fmt.Sprintf("%supgrade worker node %q (%s): %s -> %s",
					forceFlag,
					node.Config.Hostname,
					node.Config.PrivateAddress,
					node.Kubelet.Version,
					s.Cluster.Versions.Kubernetes))
		}
	} else {
		tasksToRun = tasks.WithResources(nil)
	}

	fmt.Println()
	for _, op := range operations {
		fmt.Printf("\t~ %s\n", op)
	}

	for _, op := range tasksToRun.Descriptions(s) {
		fmt.Printf("\t~ %s\n", op)
	}

	fmt.Println()
	confirm, err := confirmCommand(opts.AutoApprove)
	if err != nil {
		return err
	}

	if !confirm {
		s.Logger.Println("Operation canceled.")

		return nil
	}

	return tasksToRun.Run(s)
}

func runApplyRotateKey(s *state.State, opts *applyOpts) error {
	if !opts.ForceUpgrade {
		s.Logger.Error("rotating encryption keys requires the --force-upgrade flag")

		return fail.ConfigValidation(fmt.Errorf("rotating encryption keys requires the --force-upgrade flag"))
	}
	if !s.EncryptionEnabled() {
		s.Logger.Error("rotating encryption keys failed: Encryption Providers support is not enabled")

		return fail.ConfigValidation(fmt.Errorf("rotating encryption keys failed: Encryption Providers support is not enabled"))
	}

	fmt.Println("The following actions will be taken: ")
	fmt.Println("Run with --verbose flag for more information.")
	tasksToRun := tasks.WithRotateKey(nil)

	for _, op := range tasksToRun.Descriptions(s) {
		fmt.Printf("\t~ %s\n", op)
	}

	fmt.Println()
	confirm, err := confirmCommand(opts.AutoApprove)
	if err != nil {
		return err
	}

	if !confirm {
		s.Logger.Println("Operation canceled.")

		return nil
	}

	return tasksToRun.Run(s)
}

func printHostInformation(host state.Host) {
	containerdCR := host.ContainerRuntimeContainerd
	dockerCR := host.ContainerRuntimeDocker

	fmt.Printf("Host: %q\n", host.Config.Hostname)
	fmt.Printf("\tHost initialized: %s\n", boolStr(host.Initialized()))

	fmt.Printf("\t%s healthy: %s (%s)\n", containerdCR.Name, boolStr(containerdCR.Healthy()), printVersion(containerdCR.Version))
	if dockerCR.IsProvisioned() {
		fmt.Printf("\t%s healthy: %s (%s)\n", dockerCR.Name, boolStr(dockerCR.Healthy()), printVersion(dockerCR.Version))
	}

	fmt.Printf("\tKubelet healthy: %s (%s)\n", boolStr(host.Kubelet.Healthy()), printVersion(host.Kubelet.Version))
	fmt.Println()

	componentStatusReport(containerdCR)

	if dockerCR.IsProvisioned() {
		fmt.Println()
		componentStatusReport(dockerCR)
	}

	fmt.Println()
	componentStatusReport(host.Kubelet)
	fmt.Println()
}

func componentStatusReport(component state.ComponentStatus) {
	fmt.Printf("\t%s is installed: %s\n", component.Name, boolStr(component.Status&state.ComponentInstalled != 0))
	fmt.Printf("\t%s is running: %s\n", component.Name, boolStr(component.Status&state.SystemDStatusRunning != 0))
	fmt.Printf("\t%s is active: %s\n", component.Name, boolStr(component.Status&state.SystemDStatusActive != 0))
	fmt.Printf("\t%s is restarting: %s\n", component.Name, boolStr(component.Status&state.SystemDStatusRestarting != 0))
}

func boolStr(b bool) string {
	if b {
		return yes
	}

	return "no"
}

func resolveInt(i *int) int {
	if i == nil {
		return 0
	}

	return *i
}

func printVersion(version *semver.Version) string {
	if version == nil {
		return "unknown"
	}

	return version.String()
}
