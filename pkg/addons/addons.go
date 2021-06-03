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
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
)

const (
	addonLabel = "kubeone.io/addon"
)

var (
	kubectlApplyScript = heredoc.Doc(`
		sudo KUBECONFIG=/etc/kubernetes/admin.conf \
		kubectl apply -f - --prune -l "%s=%s"
	`)
)

// TemplateData is data available in the addons render template
type TemplateData struct {
	Config      *kubeoneapi.KubeOneCluster
	Credentials map[string]string
}

func Ensure(s *state.State) error {
	if s.Cluster.Addons == nil || !s.Cluster.Addons.Enable {
		s.Logger.Infoln("Skipping applying addons because addons are not enabled...")
		return nil
	}
	s.Logger.Infoln("Applying addons...")

	creds, err := credentials.Any(s.CredentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "unable to fetch credentials")
	}

	templateData := TemplateData{
		Config:      s.Cluster,
		Credentials: creds,
	}

	addonsPath, dirs, err := traverseAddonsDirectory(s)
	if err != nil {
		return errors.Wrap(err, "failed to parse the addons directory")
	}

	for _, addonDir := range dirs {
		if len(addonDir) == 0 {
			s.Logger.Info("Applying addons from the root directory...")
		} else {
			s.Logger.Infof("Applying addon %q...", addonDir)
		}

		manifest, err := getManifestsFromDirectory(s, templateData, addonsPath, addonDir)
		if err != nil {
			return errors.WithStack(err)
		}
		if len(strings.TrimSpace(manifest)) == 0 {
			if len(addonDir) != 0 {
				s.Logger.Warnf("Addon directory %q is empty, skipping...", addonDir)
			}
			continue
		}

		if err := applyAddons(s, manifest, addonDir); err != nil {
			return errors.Wrap(err, "failed to apply addons")
		}
	}

	return nil
}

func applyAddons(s *state.State, manifest string, addonName string) error {
	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
		var (
			cmd            = fmt.Sprintf(kubectlApplyScript, addonLabel, addonName)
			r              = strings.NewReader(manifest)
			stdout, stderr strings.Builder
		)

		_, err := conn.POpen(cmd, r, &stdout, &stderr)
		if s.Verbose {
			fmt.Printf("+ %s\n", cmd)
			fmt.Printf("%s", stderr.String())
			fmt.Printf("%s", stdout.String())
		}

		return err
	})
}
