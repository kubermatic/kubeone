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

package runner

import (
	"io/fs"
	"os"

	"github.com/koron-go/prefixw"
	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/ssh/sshiofs"
)

// Runner bundles a connection to a host with the verbosity and
// other options for running commands via SSH.
type Runner struct {
	Conn    ssh.Connection
	Prefix  string
	OS      kubeoneapi.OperatingSystemName
	Verbose bool
}

// TemplateVariables is a render context for templates
type TemplateVariables map[string]interface{}

func (r *Runner) NewFS() fs.FS {
	return sshiofs.New(r.Conn)
}

func (r *Runner) RunRaw(cmd string) (string, string, error) {
	if r.Conn == nil {
		return "", "", errors.New("runner is not tied to an opened SSH connection")
	}

	if !r.Verbose {
		stdout, stderr, _, err := r.Conn.Exec(cmd)
		if err != nil {
			err = errors.Wrap(err, stderr)
		}

		return stdout, stderr, err
	}

	stdout := NewTee(prefixw.New(os.Stdout, r.Prefix))
	defer stdout.Close()

	stderr := NewTee(prefixw.New(os.Stderr, r.Prefix))
	defer stderr.Close()

	// run the command
	_, err := r.Conn.POpen(cmd, nil, stdout, stderr)

	return stdout.String(), stderr.String(), err
}

// Run executes a given command/script, optionally printing its output to
// stdout/stderr.
func (r *Runner) Run(cmd string, variables TemplateVariables) (string, string, error) {
	cmd, err := scripts.Render(cmd, variables)
	if err != nil {
		return "", "", err
	}

	return r.RunRaw(cmd)
}
