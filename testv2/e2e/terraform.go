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

package e2e

import (
	"os"

	"k8c.io/kubeone/test/e2e/testutil"
)

var (
	defaultTFEnvironment = []string{
		"TF_IN_AUTOMATION=true",
		"TF_CLI_ARGS=-no-color",
	}
)

type terraformBin struct {
	path    string
	vars    []string
	varFile string
}

func (tf *terraformBin) Init() error {
	return tf.run("init")
}

func (tf *terraformBin) Apply() error {
	args := []string{"apply", "-auto-approve"}
	args = append(args, tf.varFlags()...)

	return tf.run(args...)
}

func (tf *terraformBin) Destroy() error {
	args := []string{"destroy", "-auto-approve"}
	args = append(args, tf.varFlags()...)

	return tf.run(args...)
}

func (tf *terraformBin) varFlags() []string {
	var args []string

	for _, arg := range tf.vars {
		args = append(args, "-var", arg)
	}

	if tf.varFile != "" {
		args = append(args, "-var-file", mustAbsolutePath(tf.varFile))
	}

	return args
}

func (tf *terraformBin) run(args ...string) error {
	return tf.build(args...).Run()
}

func (tf *terraformBin) build(args ...string) *testutil.Exec {
	return testutil.NewExec("terraform",
		testutil.WithArgs(args...),
		testutil.WithEnv(append(os.Environ(), defaultTFEnvironment...)),
		testutil.InDir(tf.path),
		testutil.StdoutDebug,
	)
}
