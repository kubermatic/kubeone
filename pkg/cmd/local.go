/*
Copyright 2022 The KubeOne Authors.

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
	"context"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/apis/kubeone/config"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	kubeonevalidation "k8c.io/kubeone/pkg/apis/kubeone/validation"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8net "k8s.io/apimachinery/pkg/util/net"
)

type localOpts struct {
	globalOptions
	AutoApprove       bool   `longflag:"auto-approve" shortflag:"y"`
	BackupFile        string `longflag:"backup" shortflag:"b"`
	NoInit            bool   `longflag:"no-init"`
	ForceInstall      bool   `longflag:"force-install"`
	ForceUpgrade      bool   `longflag:"force-upgrade"`
	KubernetesVersion string `longflag:"kubernetes-version"`
	APIEndpoint       string `longflag:"api-endpoint"`
}

func (opts *localOpts) BuildState() (*state.State, error) {
	logger := newLogger(opts.Verbose, opts.LogFormat)

	var (
		cluster      *kubeoneapi.KubeOneCluster
		haveManifest bool
		err          error
	)

	_, err = os.Stat(opts.ManifestFile)
	if err == nil {
		haveManifest = true
	}

	if haveManifest {
		cluster, err = loadClusterConfig(opts.ManifestFile, "", "", logger)
		if err != nil {
			return nil, err
		}
		convertToLocalCluster(cluster, logger)
	} else {
		cluster = generateLocalCluster(logger, opts.KubernetesVersion, opts.APIEndpoint)
	}
	rootContext := context.Background()

	localExec := executor.NewLocal(rootContext)

	s, err := state.New(rootContext, state.WithExecutorAdapter(localExec))
	if err != nil {
		return nil, err
	}

	s.Logger = logger
	s.Cluster = cluster

	// Validate Addons path if provided
	if s.Cluster.Addons.Enabled() {
		addonsPath, err := s.Cluster.Addons.RelativePath(s.ManifestFilePath)
		if err != nil {
			return nil, err
		}

		// Check if only embedded addons are being used; path is not required for embedded addons and no validation is required
		embeddedAddonsOnly, err := addons.EmbeddedAddonsOnly(s.Cluster.Addons.Addons)
		if err != nil {
			return nil, err
		}

		// If custom addons are being used then addons path is required and should be a valid directory
		if !embeddedAddonsOnly {
			if _, err := os.Stat(addonsPath); os.IsNotExist(err) {
				return nil, fail.Runtime(err, "checking addons directory")
			}
		}
	}

	s.ManifestFilePath = opts.ManifestFile
	s.Verbose = opts.Verbose
	s.BackupFile = defaultBackupPath(opts.BackupFile, opts.ManifestFile, s.Cluster.Name)
	s.ForceInstall = opts.ForceInstall
	s.ForceUpgrade = opts.ForceUpgrade

	return s, initBackup(s.BackupFile)
}

func localCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &localOpts{}

	cmd := &cobra.Command{
		Use:   "local",
		Short: "Reconcile the local one-node-all-in-one cluster",
		Long: heredoc.Doc(`
			Initialize the all-in-one node cluster on a local machine.
		`),
		SilenceErrors: true,
		Example:       `kubeone local`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			opts.globalOptions = *gopts

			return runLocal(opts)
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

	cmd.Flags().StringVar(
		&opts.KubernetesVersion,
		longFlagName(opts, "KubernetesVersion"),
		"1.24.2",
		"kubernetes version to install when there is no manifest")

	cmd.Flags().StringVar(
		&opts.APIEndpoint,
		longFlagName(opts, "ApiEndpoint"),
		"",
		"kube-apiserver endpoint to init, defaut to autodetect")

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

	return cmd
}

func runLocal(opts *localOpts) error {
	st, err := opts.BuildState()
	if err != nil {
		return err
	}

	aopts := &applyOpts{
		globalOptions: opts.globalOptions,
		AutoApprove:   opts.AutoApprove,
		BackupFile:    opts.BackupFile,
		NoInit:        opts.NoInit,
		ForceInstall:  opts.ForceInstall,
		ForceUpgrade:  opts.ForceUpgrade,
	}

	return runApply(st, aopts)
}

func generateLocalCluster(logger logrus.FieldLogger, kubeVersion, apiEndpoint string) *kubeoneapi.KubeOneCluster {
	ownIP, err := k8net.ChooseHostInterface()
	if err != nil {
		panic(err)
	}

	cls := &kubeonev1beta2.KubeOneCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubeonev1beta2.SchemeGroupVersion.String(),
			Kind:       "KubeOneCluster",
		},
		APIEndpoint: kubeonev1beta2.APIEndpoint{
			Host: apiEndpoint,
		},
		Name: "local",
		ControlPlane: kubeonev1beta2.ControlPlaneConfig{
			Hosts: []kubeonev1beta2.HostConfig{
				{
					PublicAddress: ownIP.String(),
					IsLeader:      true,
					Taints:        []corev1.Taint{},
				},
			},
		},
		CloudProvider: kubeonev1beta2.CloudProviderSpec{
			None: &kubeonev1beta2.NoneSpec{},
		},
		MachineController: &kubeonev1beta2.MachineControllerConfig{
			Deploy: false,
		},
		Versions: kubeonev1beta2.VersionConfig{
			Kubernetes: kubeVersion,
		},
	}

	kubeonev1beta2.SetObjectDefaults_KubeOneCluster(cls)

	internalCluster, err := config.DefaultedV1Beta2KubeOneCluster(cls, nil, nil, logger)
	if err != nil {
		// this should never happen
		panic(err)
	}

	if err = kubeonevalidation.ValidateKubeOneCluster(*internalCluster).ToAggregate(); err != nil {
		// this should never happen
		panic(err)
	}

	return internalCluster
}

func convertToLocalCluster(in *kubeoneapi.KubeOneCluster, logger logrus.FieldLogger) {
	genCluster := generateLocalCluster(logger, in.Versions.Kubernetes, in.APIVersion)

	in.Name = genCluster.Name
	in.ControlPlane = genCluster.ControlPlane
	in.CloudProvider = genCluster.CloudProvider
	in.MachineController = genCluster.MachineController
	in.Versions = genCluster.Versions
}
