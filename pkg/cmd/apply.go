package cmd

import (
	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type applyOpts struct {
	globalOptions
	NoInit      bool `longflag:"no-init"`
	Force       bool `longflag:"force"`
	AutoApprove bool `longflag:"auto-approve"`
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
		&opts.NoInit,
		longFlagName(opts, "NoInit"),
		false,
		"don't initialize the cluster (only install binaries)")

	cmd.Flags().BoolVar(
		&opts.AutoApprove,
		longFlagName(opts, "AutoApprove"),
		false,
		"auto approve plan")

	cmd.Flags().BoolVar(
		&opts.Force,
		longFlagName(opts, "Force"),
		false,
		"use force to install new binary versions (!dangerous!)")

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

	return nil
}
