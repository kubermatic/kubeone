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
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"k8c.io/kubeone/test/testexec"
)

type protokolBin struct {
	namespaces []string
	outputDir  string
}

var (
	protokolBuildOnce sync.Once
	protokolBinPath   string
	errProtokolBuild  error
)

// buildProtokol compiles the protokol tool once per test binary. It is built
// up front, without the kubeone proxy in the environment, so that fetching
// modules does not go through the SSH tunnel. Running the binary directly
// (rather than via `go tool`) also means cancelling the context kills
// protokol itself instead of orphaning it behind the `go` wrapper process.
func buildProtokol(ctx context.Context) (string, error) {
	protokolBuildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "protokol-*")
		if err != nil {
			errProtokolBuild = err

			return
		}

		bin := filepath.Join(dir, "protokol")
		out, err := exec.CommandContext(ctx, "go", "build", "-o", bin, "go.xrstf.de/protokol").CombinedOutput()
		if err != nil {
			errProtokolBuild = fmt.Errorf("building protokol: %w: %s", err, out)

			return
		}

		protokolBinPath = bin
	})

	return protokolBinPath, errProtokolBuild
}

// Start runs protokol in the background, following the logs of all
// containers in the configured namespaces until the returned stop function is
// called.
func (p *protokolBin) Start(ctx context.Context, kubeconfigPath, proxyURL string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("starting protokol: %w", err)
	}

	protocolCtx, cancel := context.WithCancel(ctx)

	cmd, logFile, err := p.cmd(protocolCtx, p.outputDir, kubeconfigPath, proxyURL)
	if err != nil {
		cancel()

		return nil, err
	}

	if err = cmd.Start(); err != nil {
		cancel()
		logFile.Close()

		return nil, fmt.Errorf("starting protokol: %w", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = cmd.Wait()
		logFile.Close()
	}()

	return func() {
		cancel()
		<-done
	}, nil
}

// Dump does a one-shot dump of the current container logs and events in the
// configured namespaces into a subdirectory of the output directory. It is
// meant to be used when a test fails before Start has been called, e.g. when
// `kubeone apply` itself fails.
func (p *protokolBin) Dump(ctx context.Context, subdir, kubeconfigPath, proxyURL string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd, logFile, err := p.cmd(cmdCtx, filepath.Join(p.outputDir, subdir), kubeconfigPath, proxyURL, "--oneshot")
	if err != nil {
		return err
	}
	defer logFile.Close()

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("running protokol (see %s): %w", logFile.Name(), err)
	}

	return nil
}

func (p *protokolBin) cmd(ctx context.Context, outputDir, kubeconfigPath, proxyURL string, extraArgs ...string) (*exec.Cmd, *os.File, error) {
	if len(p.namespaces) == 0 {
		return nil, nil, errors.New("refusing to dump *everything*, please specify namespaces")
	}

	bin, err := buildProtokol(ctx)
	if err != nil {
		return nil, nil, err
	}

	if err = os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating protokol output dir: %w", err)
	}

	// protokol's own output goes next to the logs so failures to connect to
	// the cluster are visible instead of silently producing an empty dir
	logFile, err := os.CreateTemp(outputDir, "protokol-*.log")
	if err != nil {
		return nil, nil, fmt.Errorf("creating protokol log file: %w", err)
	}

	args := []string{"--output", outputDir, "--kubeconfig", kubeconfigPath, "--events"}
	for _, ns := range p.namespaces {
		args = append(args, "--namespace", ns)
	}
	args = append(args, extraArgs...)

	env := os.Environ()
	if proxyURL != "" {
		env = append(
			env,
			fmt.Sprintf("HTTPS_PROXY=%s", proxyURL),
			fmt.Sprintf("HTTP_PROXY=%s", proxyURL),
		)
	}

	cmd := testexec.NewExec(
		bin,
		testexec.WithArgs(args...),
		testexec.WithEnv(env),
		testexec.StdoutTo(logFile),
		testexec.StderrTo(logFile),
	).BuildCmd(ctx)

	return cmd, logFile, nil
}
