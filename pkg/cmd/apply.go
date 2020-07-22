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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/tasks"
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
}

func (opts *applyOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build state")
	}

	s.BackupFile = opts.BackupFile
	s.ForceInstall = opts.ForceInstall
	s.ForceUpgrade = opts.ForceUpgrade
	s.UpgradeMachineDeployments = opts.UpgradeMachineDeployments

	if s.BackupFile == "" {
		fullPath, _ := filepath.Abs(opts.ManifestFile)
		clusterName := s.Cluster.Name
		s.BackupFile = filepath.Join(filepath.Dir(fullPath), fmt.Sprintf("%s.tar.gz", clusterName))
	}

	// refuse to overwrite existing backups (NB: since we attempt to
	// write to the file later on to check for write permissions, we
	// inadvertently create a zero byte file even if the first step
	// of the installer fails; for this reason it's okay to find an
	// existing, zero byte backup)
	stat, err := os.Stat(s.BackupFile)
	if err != nil && stat != nil && stat.Size() > 0 {
		return nil, errors.Errorf("backup %s already exists, refusing to overwrite", opts.BackupFile)
	}

	// try to write to the file before doing anything else
	f, err := os.OpenFile(s.BackupFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %q for writing", opts.BackupFile)
	}

	return s, f.Close()
}

func applyCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &applyOpts{}

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Reconcile the cluster",
		Long: `
Reconcile (Install/Upgrade/Repair/Restore) Kubernetes cluster on pre-existing machines

This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.
`,
		Example: `kubeone apply -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			opts.globalOptions = *gopts

			return runApply(opts)
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

	return cmd
}

func runApply(opts *applyOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	// Validate credentials
	_, err = credentials.ProviderCredentials(s.Cluster.CloudProvider, opts.CredentialsFile)
	if err != nil {
		return errors.Wrap(err, "failed to validate credentials")
	}

	// Probe the cluster for the actual state and the needed tasks.
	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbes(probbing)

	if err = probbing.Run(s); err != nil {
		return err
	}

	if s.Verbose {
		// Print information about hosts collected by probes
		for _, host := range s.LiveCluster.ControlPlane {
			printHostInformation(host)
		}

		for _, host := range s.LiveCluster.StaticWorkers {
			printHostInformation(host)
		}
	}

	// Reconcile the cluster based on the probe status
	if !s.LiveCluster.IsProvisioned() {
		return runApplyInstall(s, opts)
	}

	if !s.LiveCluster.Healthy() {
		brokenHosts := s.LiveCluster.BrokenHosts()
		if len(brokenHosts) > 0 {
			for _, node := range brokenHosts {
				s.Logger.Errorf("Host %q is broken and needs to be manually removed\n", node)
			}

			s.Logger.Warnf("Hosts must be removed in a correct order to preserve the Etcd quorum.")
			s.Logger.Warnf("Loss of the Etcd quorum can cause loss of all data!!!")
			s.Logger.Warnf("After removing the recommended hosts, run 'kubeone apply' before removing any other host.")

			safeToDelete := s.LiveCluster.SafeToDeleteHosts()
			if len(safeToDelete) > 0 {
				s.Logger.Warnf("The recommended removal order:")
				for _, safe := range safeToDelete {
					s.Logger.Warnf("- %q", safe)
				}
			} else {
				s.Logger.Warnf("No other broken node can be removed without losing quorum.")
			}
		}

		runRepair := false
		for _, node := range s.LiveCluster.ControlPlane {
			if !node.IsInCluster {
				runRepair = true
				break
			}
		}

		if !runRepair {
			for _, node := range s.LiveCluster.StaticWorkers {
				if !node.IsInCluster {
					runRepair = true
					break
				}
			}
		}

		if safeRepair, higherVer := s.LiveCluster.SafeToRepair(s.Cluster.Versions.Kubernetes); !safeRepair {
			s.Logger.Errorln("Repair and upgrade are not supported at the same time!")
			s.Logger.Warnf("Requested version: %s\n", s.Cluster.Versions.Kubernetes)
			s.Logger.Warnf("Highest version: %s\n", higherVer)
			s.Logger.Warnf("Use version %s to repair the cluster, then run apply with the new version\n", higherVer)
			return errors.New("repair and upgrade are not supported at the same time")
		}

		if runRepair {
			return runApplyInstall(s, opts)
		}

		if len(brokenHosts) > 0 {
			return errors.New("broken host(s) found, remove it manually")
		}

		return nil
	}

	return runApplyUpgradeIfNeeded(s, opts)
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
			fmt.Printf("\t+ join worker node %q (%s)\n", node.Config.Hostname, node.Config.PrivateAddress)
		}
	}

	if opts.NoInit {
		fmt.Println("\t! NoInit option provided: only binaries will be installed")
	}

	if opts.ForceInstall {
		fmt.Println("\t! force-install option provided: force install new binary versions (!dangerous!)")
	}

	for _, node := range s.Cluster.DynamicWorkers {
		fmt.Printf("\t+ ensure machinedeployment %q with %d replica(s) exists\n", node.Name, resolveInt(node.Replicas))
	}

	if s.Cluster.Addons != nil && s.Cluster.Addons.Enable {
		fmt.Printf("\t+ apply addons defined in %q\n", s.Cluster.Addons.Path)
	}

	fmt.Println()
	confirm, err := confirmApply(opts.AutoApprove)
	if err != nil {
		return err
	}

	if !confirm {
		s.Logger.Println("Operation canceled.")
		return nil
	}

	if opts.NoInit {
		return errors.Wrap(tasks.WithBinariesOnly(nil).Run(s), "failed to install kubernetes binaries")
	}

	return errors.Wrap(tasks.WithFullInstall(nil).Run(s), "failed to install the cluster")
}

func runApplyUpgradeIfNeeded(s *state.State, opts *applyOpts) error {
	fmt.Println("The following actions will be taken: ")
	fmt.Println("Run with --verbose flag for more information.")

	var tasksToRun tasks.Tasks

	upgradeNeeded, err := s.LiveCluster.UpgradeNeeded()
	if err != nil {
		s.Logger.Errorf("Upgrade not allowed: %v\n", err)
		return err
	}

	operations := []string{}

	if upgradeNeeded || opts.ForceUpgrade {
		tasksToRun = tasks.WithUpgrade(nil)

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
		tasksToRun = tasks.WithRefreshResources(nil)
	}

	fmt.Println()
	for _, op := range operations {
		fmt.Printf("\t~ %s\n", op)
	}

	for _, op := range tasksToRun.Descriptions(s) {
		fmt.Printf("\t~ %s\n", op)
	}

	fmt.Println()
	confirm, err := confirmApply(opts.AutoApprove)
	if err != nil {
		return err
	}

	if !confirm {
		s.Logger.Println("Operation canceled.")
		return nil
	}

	return errors.Wrap(tasksToRun.Run(s), "failed to reconcile the cluster")
}

func confirmApply(autoApprove bool) (bool, error) {
	if autoApprove {
		return true, nil
	}

	if !terminal.IsTerminal(int(os.Stdin.Fd())) || !terminal.IsTerminal(int(os.Stdout.Fd())) {
		return false, errors.New("not running in the terminal")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to proceed (yes/no): ")

	confirmation, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	fmt.Println()

	return strings.Trim(confirmation, "\n") == "yes", nil
}

func printHostInformation(host state.Host) {
	fmt.Printf("Host: %q\n", host.Config.Hostname)
	fmt.Printf("\tHost initialized: %s\n", boolStr(host.Initialized()))
	fmt.Printf("\tDocker healthy: %s (%s)\n", boolStr(host.ContainerRuntime.Healthy()), printVersion(host.ContainerRuntime.Version))
	fmt.Printf("\tKubelet healthy: %s (%s)\n", boolStr(host.Kubelet.Healthy()), printVersion(host.Kubelet.Version))

	fmt.Println()
	fmt.Printf("\tDocker is installed: %s\n", boolStr(host.ContainerRuntime.Status&state.ComponentInstalled != 0))
	fmt.Printf("\tDocker is running: %s\n", boolStr(host.ContainerRuntime.Status&state.SystemDStatusRunning != 0))
	fmt.Printf("\tDocker is active: %s\n", boolStr(host.ContainerRuntime.Status&state.SystemDStatusActive != 0))
	fmt.Printf("\tDocker is restarting: %s\n", boolStr(host.ContainerRuntime.Status&state.SystemDStatusRestarting != 0))

	fmt.Println()
	fmt.Printf("\tKubelet is installed: %s\n", boolStr(host.Kubelet.Status&state.ComponentInstalled != 0))
	fmt.Printf("\tKubelet is running: %s\n", boolStr(host.Kubelet.Status&state.SystemDStatusRunning != 0))
	fmt.Printf("\tKubelet is active: %s\n", boolStr(host.Kubelet.Status&state.SystemDStatusActive != 0))
	fmt.Printf("\tKubelet is restarting: %s\n", boolStr(host.Kubelet.Status&state.SystemDStatusRestarting != 0))
	fmt.Println()
}

func boolStr(b bool) string {
	if b {
		return "yes"
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
		return "not installed"
	}
	return version.String()
}
