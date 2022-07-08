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
	"os"

	"github.com/koron-go/prefixw"
	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/scripts"
)

// Runner bundles a connection to a host with the verbosity and
// other options for running commands via SSH.
type Runner struct {
	Executor executor.Interface
	Prefix   string
	OS       kubeoneapi.OperatingSystemName
	Verbose  bool
}

// TemplateVariables is a render context for templates
type TemplateVariables map[string]interface{}

func (r *Runner) NewFS() executor.MkdirFS {
	return executorfs.New(r.Executor)
}

func (r *Runner) RunRaw(cmd string) (string, string, error) {
	if r.Executor == nil {
		return "", "", fail.RuntimeError{
			Op:  "checking available executor adapter",
			Err: errors.New("runner has no open adapter"),
		}
	}

	if !r.Verbose {
		stdout, stderr, _, err := r.Executor.Exec(cmd)
		if err != nil {
			r.Executor.Close()
			r.Executor = nil
		}

		return stdout, stderr, err
	}

	stdout := NewTee(prefixw.New(os.Stdout, r.Prefix))
	defer stdout.Close()

	stderr := NewTee(prefixw.New(os.Stderr, r.Prefix))
	defer stderr.Close()

	// run the command
	_, err := r.Executor.POpen(cmd, nil, stdout, stderr)

	if err != nil {
		r.Executor.Close()
		r.Executor = nil
	}

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
