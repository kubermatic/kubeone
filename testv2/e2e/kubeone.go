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
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/apis/kubeone/config"
	"k8c.io/kubeone/test/e2e/testutil"
)

var (
	kubeoneVerboseFlag = flag.Bool("kubeone-verbose", false, "run kubeone actions with --verbose flag")
	credentialsFlag    = flag.String("credentials", "", "run kubeone with --credentials flag")
)

type kubeoneBin struct {
	bin             string
	dir             string
	tfjsonPath      string
	manifestPath    string
	credentialsPath string
	verbose         bool
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

func (k1 *kubeoneBin) AsyncProxy(ctx context.Context) (string, func() error, error) {
	list, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}

	hostPort := list.Addr().String()
	if err = list.Close(); err != nil {
		return "", nil, err
	}

	proxyURL := url.URL{
		Scheme: "http",
		Host:   hostPort,
	}

	cmd := k1.build("proxy", "--listen", hostPort).BuildCmd(ctx)
	if err = cmd.Start(); err != nil {
		return "", nil, err
	}

	return proxyURL.String(), cmd.Wait, nil
}

func (k1 *kubeoneBin) ClusterManifest() (*kubeoneapi.KubeOneCluster, error) {
	var buf bytes.Buffer

	args := k1.globalFlags()
	exe := k1.build(append(args, "config", "dump")...)
	testutil.StdoutTo(&buf)(exe)

	if err := exe.Run(); err != nil {
		return nil, fmt.Errorf("rendering manifest failed: %w", err)
	}

	logger := logrus.New()
	k1Manifest, err := config.BytesToKubeOneCluster(buf.Bytes(), nil, nil, logger)

	return k1Manifest, err
}

func (k1 *kubeoneBin) globalFlags() []string {
	args := []string{"--tfjson", k1.tfjsonPath}

	if k1.verbose {
		args = append(args, "--verbose")
	}

	if k1.manifestPath != "" {
		args = append(args, "--manifest", k1.manifestPath)
	}

	if k1.credentialsPath != "" {
		args = append(args, "--credentials", k1.credentialsPath)
	}

	return args
}

func (k1 *kubeoneBin) kubeconfigPath(tmpDir string) (string, error) {
	kubeconfig, err := os.CreateTemp(tmpDir, "kubeconfig-*")
	if err != nil {
		return "", err
	}
	defer kubeconfig.Close()

	buf, err := k1.Kubeconfig()
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(kubeconfig.Name(), buf, 0600); err != nil {
		return "", err
	}

	return kubeconfig.Name(), nil
}

func (k1 *kubeoneBin) run(args ...string) error {
	return k1.build(args...).Run()
}

func (k1 *kubeoneBin) build(args ...string) *testutil.Exec {
	bin := "kubeone"
	if k1.bin != "" {
		bin = k1.bin
	}

	return testutil.NewExec(bin,
		testutil.WithArgs(append(k1.globalFlags(), args...)...),
		testutil.WithEnv(os.Environ()),
		testutil.InDir(k1.dir),
		testutil.StdoutDebug,
	)
}
