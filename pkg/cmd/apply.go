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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/tasks"
)

type applyOpts struct {
	globalOptions
	AutoApprove bool `longflag:"auto-approve"`
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
		Hidden: true, // for now
		Use:    "apply",
		Short:  "apply reconcile the cluster",
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

	cmd.Flags().BoolVar(
		&opts.AutoApprove,
		longFlagName(opts, "AutoApprove"),
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

	// Reconcile the cluster based on the probe status
	if !s.LiveCluster.IsProvisioned() {
		return runApplyInstall(s, opts)
	}
	if !s.LiveCluster.Healthy() {
		if broken, nodes := s.LiveCluster.IsBroken(); broken {
			for _, node := range nodes {
				s.Logger.Errorf("host %q is broken and needs to be manually removed\n", node)
			}
			s.Logger.Warnf("You can remove %d hosts at the same or otherwise quorum be lost!!!\n", s.LiveCluster.EtcdToleranceRemain())
			s.Logger.Warnf("After removing host(s), run kubeone apply again\n")
		}
		// TODO: Should we return at the beginning after install?
		for _, node := range s.LiveCluster.ControlPlane {
			if !node.IsInCluster {
				return runApplyInstall(s, opts)
			}
		}
		return nil
	}

	return runApplyUpgradeIfNeeded(s, opts)
}

func runApplyInstall(s *state.State, opts *applyOpts) error { // Print the expected changes
	fmt.Println("The following actions will be taken: ")
	fmt.Println()

	for _, node := range s.LiveCluster.ControlPlane {
		if !node.IsInCluster {
			fmt.Printf("+ provision host %q (%s)\n", node.Config.Hostname, node.Config.PrivateAddress)
		}
	}

	if opts.NoInit {
		fmt.Println("+ NoInit option provided: only binaries will be installed")
	}
	if opts.ForceInstall {
		fmt.Println("! force-install option provided: force install new binary versions (!dangerous!)")
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
	if s.LiveCluster.UpgradeNeeded() || opts.ForceUpgrade {
		fmt.Println("The following actions will be taken: ")
		fmt.Println()

		// TODO: Maybe it's not needed to upgrade each node, check version
		for _, node := range s.Cluster.ControlPlane.Hosts {
			if opts.ForceUpgrade {
				fmt.Printf("~ force upgrade node %q (%s) to Kubernetes %s\n", node.Hostname, node.PrivateAddress, s.Cluster.Versions.Kubernetes)
			} else {
				fmt.Printf("~ upgrade node %q (%s) to Kubernetes %s\n", node.Hostname, node.PrivateAddress, s.Cluster.Versions.Kubernetes)
			}
		}

		if s.UpgradeMachineDeployments {
			fmt.Printf("~ replace all worker machines with %s\n", s.Cluster.Versions.Kubernetes)
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

		return errors.Wrap(tasks.WithUpgrade(nil).Run(s), "failed to upgrade cluster")
	}
	s.Logger.Println("The expected state matches actual, no action needed.")
	return nil
}

func confirmApply(autoApprove bool) (bool, error) {
	if autoApprove {
		return true, nil
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
