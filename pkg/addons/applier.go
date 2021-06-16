/*
Copyright 2021 The KubeOne Authors.

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
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"

	"k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"
)

var (
	kubectlApplyScript = heredoc.Doc(`
		sudo KUBECONFIG=/etc/kubernetes/admin.conf \
		kubectl apply -f - --prune -l "%s=%s"
	`)
)

// Applier holds structure used to fetch, parse, and apply addons
type applier struct {
	TemplateData templateData
	LocalFS      fs.FS
	EmbededFS    embed.FS
}

// TemplateData is data available in the addons render template
type templateData struct {
	Config      *kubeoneapi.KubeOneCluster
	Credentials map[string]string
	Resources   map[string]string
}

func newAddonsApplier(s *state.State) (*applier, error) {
	var localFS fs.FS
	if s.Cluster.Addons != nil && s.Cluster.Addons.Enable {
		addonsPath := s.Cluster.Addons.Path
		if !filepath.IsAbs(addonsPath) && s.ManifestFilePath != "" {
			manifestAbsPath, err := filepath.Abs(filepath.Dir(s.ManifestFilePath))
			if err != nil {
				return nil, errors.Wrap(err, "unable to get absolute path to the cluster manifest")
			}
			addonsPath = filepath.Join(manifestAbsPath, addonsPath)
		}

		localFS = os.DirFS(addonsPath)
	}

	creds, err := credentials.Any(s.CredentialsFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch credentials")
	}

	td := templateData{
		Config:      s.Cluster,
		Credentials: creds,
		Resources:   resources.All(),
	}

	return &applier{
		TemplateData: td,
		LocalFS:      localFS,
		EmbededFS:    addons.F,
	}, nil
}
