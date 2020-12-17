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

package state

import (
	"context"

	"github.com/Masterminds/semver"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/configupload"
	"k8c.io/kubeone/pkg/runner"
	"k8c.io/kubeone/pkg/ssh"

	"k8s.io/client-go/rest"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func New(ctx context.Context) (*State, error) {
	joinToken, err := bootstraputil.GenerateBootstrapToken()
	return &State{
		JoinToken:     joinToken,
		Connector:     ssh.NewConnector(ctx),
		Configuration: configupload.NewConfiguration(),
		Context:       ctx,
		WorkDir:       "./kubeone",
	}, err
}

// State holds together currently test flags and parsed info, along with
// utilities like logger
type State struct {
	Cluster                   *kubeoneapi.KubeOneCluster
	LiveCluster               *Cluster
	Logger                    logrus.FieldLogger
	Connector                 *ssh.Connector
	Configuration             *configupload.Configuration
	Runner                    *runner.Runner
	Context                   context.Context
	WorkDir                   string
	JoinCommand               string
	JoinToken                 string
	RESTConfig                *rest.Config
	DynamicClient             dynclient.Client
	Verbose                   bool
	BackupFile                string
	DestroyWorkers            bool
	RemoveBinaries            bool
	ForceUpgrade              bool
	ForceInstall              bool
	UpgradeMachineDeployments bool
	PatchCNI                  bool
	CredentialsFilePath       string
	ManifestFilePath          string
	PauseImage                string
}

// ContainerRuntimeConfig return API object that is product of KubeOne manifest and live cluster probing
func (s *State) ContainerRuntimeConfig() kubeoneapi.ContainerRuntimeConfig {
	crCfg := *s.Cluster.ContainerRuntime.DeepCopy()
	condition, _ := semver.NewConstraint(">= 1.23")

	if condition.Check(s.LiveCluster.ExpectedVersion) {
		// forced containerd for clusters version >= 1.23
		crCfg.Docker = nil
		crCfg.Containerd = &kubeoneapi.ContainerRuntimeContainerd{}
		return crCfg
	}

	switch {
	case crCfg.Docker != nil:
	case crCfg.Containerd != nil:
	default:
		crCfg.Docker = &kubeoneapi.ContainerRuntimeDocker{}
	}

	return crCfg
}

func (s *State) KubeadmVerboseFlag() string {
	if s.Verbose {
		return "--v=6"
	}
	return ""
}

// Clone returns a shallow copy of the State.
func (s *State) Clone() *State {
	newState := *s
	return &newState
}
