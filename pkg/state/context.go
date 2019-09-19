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
	"github.com/sirupsen/logrus"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/configupload"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"

	"k8s.io/client-go/rest"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func New() (*State, error) {
	joinToken, err := bootstraputil.GenerateBootstrapToken()
	return &State{
		JoinToken: joinToken,
	}, err
}

// State holds together currently test flags and parsed info, along with
// utilities like logger
type State struct {
	Cluster                   *kubeoneapi.KubeOneCluster
	Logger                    logrus.FieldLogger
	Connector                 *ssh.Connector
	Configuration             *configupload.Configuration
	Runner                    *runner.Runner
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
	UpgradeMachineDeployments bool
	PatchCNI                  bool
	CredentialsFilePath       string
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
