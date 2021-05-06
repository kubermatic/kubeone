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
	"github.com/kubermatic/machine-controller/pkg/machines/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tasks"
)

type resetOpts struct {
	globalOptions
	AutoApprove    bool `longflag:"auto-approve" shortflag:"y"`
	DestroyWorkers bool `longflag:"destroy-workers"`
	RemoveBinaries bool `longflag:"remove-binaries"`
}

func (opts *resetOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build State")
	}

	s.DestroyWorkers = opts.DestroyWorkers
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
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
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
		"auto approve reset (NO-OP/NOT YET ENABLED)")

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

	return cmd
}

// runReset resets all machines provisioned by KubeOne
func runReset(opts *resetOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	s.Logger.Warnln("this command will require an explicit confirmation starting with the next minor release (v1.3)")

	fmt.Println("The following nodes will be reset. The Kubernetes cluster running on those nodes will be permanently destroyed ")
	for _, node := range s.Cluster.ControlPlane.Hosts {
		fmt.Printf("\t- reset control plane node %q (%s)\n", node.Hostname, node.PrivateAddress)
	}
	for _, node := range s.Cluster.StaticWorkers.Hosts {
		fmt.Printf("\t- reset static worker nodes %q (%s)\n", node.Hostname, node.PrivateAddress)
	}

	err = kubeconfig.BuildKubernetesClientset(s)
	if err == nil {
		// Gather information about machine-controller managed nodes
		machines := v1alpha1.MachineList{}
		if err := s.DynamicClient.List(s.Context, &machines); err != nil {
			s.Logger.Warnln("Failed to list Machines. Worker nodes will not be deleted. If there are worker nodes in the cluster, you might have to delete them manually.")
		}
		for _, machine := range machines.Items {
			fmt.Printf("\t+ reset machine-controller managed machines %q\n", machine.Name)
		}
	} else {
		s.Logger.Warnln("Failed to list Machines. Worker nodes will not be deleted. If there are worker nodes in the cluster, you might have to delete them manually.")
	}

	confirm, err := confirmCommand(opts.AutoApprove)
	if err != nil {
		return err
	}

	if !confirm {
		s.Logger.Println("Operation canceled.")
		return nil
	}

	return errors.Wrap(tasks.WithReset(nil).Run(s), "failed to reset the cluster")
}
