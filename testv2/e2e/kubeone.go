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
	"fmt"
	"os"

	"k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/test/e2e/testutil"

	"sigs.k8s.io/yaml"
)

type kubeoneBin struct {
	bin             string
	dir             string
	tfjsonPath      string
	manifestPath    string
	credentialsPath string
}

func (k1 *kubeoneBin) globalFlags() []string {
	args := []string{"--tfjson", k1.tfjsonPath}

	if k1.manifestPath != "" {
		args = append(args, "--manifest", k1.manifestPath)
	}

	if k1.credentialsPath != "" {
		args = append(args, "--credentials", k1.credentialsPath)
	}

	return args
}

func (k1 *kubeoneBin) Apply() error {
	return k1.run("apply", "--auto-approve")
}

func (k1 *kubeoneBin) Kubeconfig() ([]byte, error) {
	var buf bytes.Buffer

	args := k1.globalFlags()
	exe := k1.build(append(args, "kubeconfig")...)
	testutil.StdoutTo(&buf)(exe)

	if err := exe.Run(); err != nil {
		return nil, fmt.Errorf("fetching kubeconfig failed: %w", err)
	}

	return buf.Bytes(), nil
}

func (k1 *kubeoneBin) Reset() error {
	return k1.run("reset", "--auto-approve", "--destroy-workers", "--remove-binaries")
}

func (k1 *kubeoneBin) Manifest() (*kubeone.KubeOneCluster, error) {
	var buf bytes.Buffer

	args := k1.globalFlags()
	exe := k1.build(append(args, "config", "dump")...)
	testutil.StdoutTo(&buf)(exe)

	if err := exe.Run(); err != nil {
		return nil, fmt.Errorf("rendering manifest failed: %w", err)
	}

	var k1Manifest kubeone.KubeOneCluster
	err := yaml.UnmarshalStrict(buf.Bytes(), &k1Manifest)

	return &k1Manifest, err
}

func (k1 *kubeoneBin) run(args ...string) error {
	return k1.build(append(k1.globalFlags(), args...)...).Run()
}

func (k1 *kubeoneBin) build(args ...string) *testutil.Exec {
	bin := "kubeone"
	if k1.bin != "" {
		bin = k1.bin
	}

	return testutil.NewExec(bin,
		testutil.WithArgs(args...),
		testutil.WithEnv(os.Environ()),
		testutil.InDir(k1.dir),
		testutil.WithDryRun(),
		testutil.StdoutDebug,
	)
}
