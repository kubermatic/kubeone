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

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tasks"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
)

type resetOpts struct {
	globalOptions
	AutoApprove      bool `longflag:"auto-approve" shortflag:"y"`
	DestroyWorkers   bool `longflag:"destroy-workers"`
	RemoveVolumes    bool `longflag:"remove-volumes"`
	RemoveLBServices bool `longflag:"remove-lb-services"`
	RemoveBinaries   bool `longflag:"remove-binaries"`
}

func (opts *resetOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, err
	}

	s.DestroyWorkers = opts.DestroyWorkers
	s.RemoveVolumes = opts.RemoveVolumes
	s.RemoveLBServices = opts.RemoveLBServices
	s.RemoveBinaries = opts.RemoveBinaries

	return s, nil
}

// resetCmd setups reset command
func resetCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &resetOpts{}

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Revert changes",
		Long: heredoc.Doc(`
			Undo all changes done by KubeOne to the configured machines.

			This command takes KubeOne manifest which contains information about hosts. It's possible to source information about
			hosts from Terraform output, using the '--tfjson' flag.
		`),
		Example: `kubeone reset -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			opts.globalOptions = *gopts

			return runReset(opts)
		},
	}

	cmd.Flags().BoolVarP(
		&opts.AutoApprove,
		longFlagName(opts, "AutoApprove"),
		shortFlagName(opts, "AutoApprove"),
		false,
		"auto approve reset")

	cmd.Flags().BoolVar(
		&opts.DestroyWorkers,
		longFlagName(opts, "DestroyWorkers"),
		true,
		"destroy all worker machines before resetting the cluster")

	cmd.Flags().BoolVar(
		&opts.RemoveBinaries,
		longFlagName(opts, "RemoveBinaries"),
		false,
		"remove kubernetes binaries after resetting the cluster")

	cmd.Flags().BoolVar(
		&opts.RemoveVolumes,
		longFlagName(opts, "RemoveVolumes"),
		true,
		"remove all dynamically provisioned and unretained volumes before resetting the cluster")

	cmd.Flags().BoolVar(
		&opts.RemoveLBServices,
		longFlagName(opts, "RemoveLBServices"),
		true,
		"remove all load balancers services before resetting the cluster")

	return cmd
}

// runReset resets all machines provisioned by KubeOne
func runReset(opts *resetOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return err
	}

	if opts.DestroyWorkers || opts.RemoveVolumes || opts.RemoveLBServices {
		if cErr := kubeconfig.BuildKubernetesClientset(s); cErr != nil {
			s.Logger.Errorln("Failed to build the Kubernetes clientset.")
			if opts.RemoveLBServices {
				s.Logger.Warnln("Unable to list and delete load balancers in the cluster.")
				s.Logger.Warnln("You can skip this phase by using '--cleanup-load-balancer=false'.")
				s.Logger.Warnln("If there are load balancers in the cluster, you might have to delete them manually.")
			}
			if opts.RemoveVolumes {
				s.Logger.Warnln("Unable to list and delete dynamically provisioned and unretained volumes in the cluster.")
				s.Logger.Warnln("You can skip this phase by using '--cleanup-volumes=false'.")
				s.Logger.Warnln("If there are unretained volumes in the cluster, you might have to delete them manually.")
			}
			if opts.DestroyWorkers {
				s.Logger.Warnln("Unable to list and delete machine-controller managed nodes.")
				s.Logger.Warnln("You can skip this phase by using '--destroy-workers=false' flag.")
				s.Logger.Warnln("If there are worker nodes in the cluster, you might have to delete them manually.")
			}

			return cErr
		}
	}

	if opts.RemoveLBServices {
		s.Logger.Warnln("remove-lb-services command will PERMANENTLY delete the load balancers from the cluster.")
	}

	if opts.RemoveVolumes {
		s.Logger.Warnln("remove-volumes command will PERMANENTLY delete the unretained volumes from the cluster.")
	}

	if opts.DestroyWorkers {
		s.Logger.Warnln("destroy-workers command will PERMANENTLY destroy the Kubernetes cluster running on the following nodes:")

		// Gather information about machine-controller managed nodes
		machines := clusterv1alpha1.MachineList{}
		if err = s.DynamicClient.List(s.Context, &machines); err != nil {
			s.Logger.Errorln("Failed to list machine-controller managed Machines.")
			s.Logger.Warnln("Worker nodes might not be deleted. If there are worker nodes in the cluster, you might have to delete them manually.")
		}

		if len(machines.Items) > 0 {
			fmt.Printf("\nThe following machine-controller managed worker nodes will be destroyed:\n")
			for _, machine := range machines.Items {
				fmt.Printf("\t- %s/%s\n", machine.Namespace, machine.Name)
			}
		}
	} else {
		s.Logger.Warnln("KubeOne will NOT remove machine-controller managed Machines.")
		s.Logger.Warnln("If there are worker nodes in the cluster, you might have to delete them manually.")
	}

	for _, node := range s.Cluster.ControlPlane.Hosts {
		fmt.Printf("\t- reset control plane node %q (%s)\n", node.Hostname, node.PrivateAddress)
	}
	for _, node := range s.Cluster.StaticWorkers.Hosts {
		fmt.Printf("\t- reset static worker nodes %q (%s)\n", node.Hostname, node.PrivateAddress)
	}

	fmt.Printf("\nAfter the command is complete, there's NO way to recover the cluster or its data!\n")

	confirm, err := confirmCommand(opts.AutoApprove)
	if err != nil {
		return err
	}

	if !confirm {
		s.Logger.Println("Operation canceled.")

		return nil
	}

	return tasks.WithReset(nil).Run(s)
}
