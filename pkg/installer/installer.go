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

package installer

import (
	"github.com/sirupsen/logrus"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/configupload"
	"github.com/kubermatic/kubeone/pkg/installer/installation"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

// Options groups the various possible options for running
// the Kubernetes installation.
type Options struct {
	Verbose        bool
	BackupFile     string
	DestroyWorkers bool
	RemoveBinaries bool
	AWSProfilePath string
	AWSProfileName string
}

// Installer is entrypoint for installation process
type Installer struct {
	cluster *kubeoneapi.KubeOneCluster
	logger  *logrus.Logger
}

// NewInstaller returns a new installer, responsible for dispatching
// between the different supported Kubernetes versions and running the
func NewInstaller(cluster *kubeoneapi.KubeOneCluster, logger *logrus.Logger) *Installer {
	return &Installer{
		cluster: cluster,
		logger:  logger,
	}
}

// Install run the installation process
func (i *Installer) Install(options *Options) error {
	return installation.Install(i.createState(options))
}

// Reset resets cluster:
// * destroys all the worker machines
// * kubeadm reset masters
func (i *Installer) Reset(options *Options) error {
	return installation.Reset(i.createState(options))
}

// createState creates a basic, non-host bound state with
// all relevant information, but *no* Runner yet. The various
// task helper functions will take care of setting up Runner
// structs for each task individually.
func (i *Installer) createState(options *Options) *state.State {
	return &state.State{
		Cluster:        i.cluster,
		Connector:      ssh.NewConnector(),
		Configuration:  configupload.NewConfiguration(),
		WorkDir:        "kubeone",
		Logger:         i.logger,
		Verbose:        options.Verbose,
		BackupFile:     options.BackupFile,
		DestroyWorkers: options.DestroyWorkers,
		RemoveBinaries: options.RemoveBinaries,
		AWSProfilePath: options.AWSProfilePath,
		AWSProfileName: options.AWSProfileName,
	}
}
