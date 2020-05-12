/*
Copyright 2020 The KubeOne Authors.

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

package addons

import (
	"fmt"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	addonLabel = "kubeone.io/addon"

	kubectlApplyScript = `kubectl apply -f {{.FILE_NAME}} --prune -l "%s"`
)

// TemplateData is data available in the addons render template
type TemplateData struct {
	Config      *kubeoneapi.KubeOneCluster
	Credentials map[string]string
}

func Ensure(s *state.State) error {
	if s.Cluster.Addons == nil || !s.Cluster.Addons.Enable {
		s.Logger.Infoln("Skipping applying addons because addons are not enabled…")
		return nil
	}
	s.Logger.Infoln("Applying addons…")

	creds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "unable to fetch credentials")
	}
	templateData := TemplateData{
		Config:      s.Cluster,
		Credentials: creds,
	}
	if err := getManifests(s, templateData); err != nil {
		return errors.WithStack(err)
	}

	if err := applyAddons(s); err != nil {
		return errors.Wrap(err, "failed to apply addons")
	}

	return nil
}

func applyAddons(s *state.State) error {
	return errors.Wrap(s.RunTaskOnLeader(runKubectl), "failed to apply addons")
}

func runKubectl(s *state.State, _ *kubeoneapi.HostConfig, _ ssh.Connection) error {
	if err := s.Configuration.UploadTo(s.Runner.Conn, s.WorkDir); err != nil {
		return errors.Wrap(err, "failed to upload manifests")
	}
	_, _, err := s.Runner.Run(fmt.Sprintf(kubectlApplyScript, addonLabel), runner.TemplateVariables{
		"FILE_NAME": fmt.Sprintf("%s/addons/", s.WorkDir),
	})
	return err
}
