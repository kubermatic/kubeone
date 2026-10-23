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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"k8c.io/kubeone/test/testexec"
)

var defaultTFEnvironment = []string{
	"TF_IN_AUTOMATION=true",
	"TF_CLI_ARGS=-no-color",
}

type terraformBin struct {
	path    string
	vars    []string
	varFile string
}

func (tf *terraformBin) Init() error {
	return tf.run("init")
}

func (tf *terraformBin) Apply(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	args := []string{"apply", "-auto-approve"}
	args = append(args, tf.varFlags()...)
	exe := tf.build(args...).BuildCmd(context.Background()) // ctx is explicitly ignored here, handled in goroutine

	go func() {
		// instead of passing down the parent context, which will cause SIGKILL of the terraform in case when context is
		// Done(), resort to gently letting the terraform to quit voluntarily
		<-ctx.Done()
		_ = exe.Process.Signal(syscall.SIGTERM)
	}()

	if err := exe.Run(); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	return nil
}

func (tf *terraformBin) Destroy() error {
	args := []string{"destroy", "-auto-approve"}
	args = append(args, tf.varFlags()...)

	return tf.run(args...)
}

// Output returns the values of the terraform outputs, keyed by the output name
func (tf *terraformBin) Output() (map[string]any, error) {
	var buf bytes.Buffer

	exe := testexec.StdoutTo(&buf)(tf.build("output", "-json"))
	if err := exe.Run(); err != nil {
		return nil, fmt.Errorf("terraform output: %w", err)
	}

	var outputs map[string]struct {
		Value any `json:"value"`
	}

	if err := json.Unmarshal(buf.Bytes(), &outputs); err != nil {
		return nil, fmt.Errorf("decoding terraform output: %w", err)
	}

	values := make(map[string]any, len(outputs))
	for name, output := range outputs {
		values[name] = output.Value
	}

	return values, nil
}

func (tf *terraformBin) varFlags() []string {
	var args []string

	if tf.varFile != "" {
		args = append(args, "-var-file", mustAbsolutePath(tf.varFile))
	}

	for _, arg := range tf.vars {
		args = append(args, "-var", arg)
	}

	return args
}

func (tf *terraformBin) run(args ...string) error {
	return tf.build(args...).Run()
}

func (tf *terraformBin) build(args ...string) *testexec.Exec {
	return testexec.NewExec("terraform",
		testexec.WithArgs(args...),
		testexec.WithEnv(append(os.Environ(), defaultTFEnvironment...)),
		testexec.InDir(tf.path),
		testexec.StdoutDebug,
	)
}
