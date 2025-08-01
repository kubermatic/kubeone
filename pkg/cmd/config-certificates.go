/*
Copyright 2025 The KubeOne Authors.

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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/certificate"
)

func configCertificatesCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificates",
		Short: "certificates manipulations",
	}
	cmd.AddCommand(configCertificatesRenewCmd(rootFlags))

	return cmd
}

func configCertificatesRenewCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	return &cobra.Command{
		Use:     "renew",
		Short:   "renew all the ceritificates of the control-plane",
		Long:    "",
		Example: "",
		RunE: func(*cobra.Command, []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			st, err := gopts.BuildState()
			if err != nil {
				return err
			}

			return certificate.RenewAll(st)
		},
	}
}
