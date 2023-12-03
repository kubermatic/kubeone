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
	"io"
	"os"

	"k8c.io/kubeone/test/testexec"
)

type protokolBin struct {
	namespaces []string
	outputDir  string
}

func (p *protokolBin) Start(ctx context.Context, kubeconfigPath string, proxyURL string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("starting protokol: %w", err)
	}

	if len(p.namespaces) == 0 {
		return nil, errors.New("refusing to dump *everything*, please specify namespaces")
	}

	args := []string{"--output", p.outputDir, "--kubeconfig", kubeconfigPath}
	for _, ns := range p.namespaces {
		args = append(args, "--namespace", ns)
	}

	protocolCtx, cancel := context.WithCancel(ctx)
	exe := p.build(proxyURL, args...).BuildCmd(protocolCtx)

	if err := exe.Start(); err != nil {
		cancel()

		return nil, err
	}

	return cancel, nil
}

func (p *protokolBin) build(proxyURL string, args ...string) *testexec.Exec {
	env := os.Environ()

	if proxyURL != "" {
		env = append(env,
			fmt.Sprintf("HTTPS_PROXY=%s", proxyURL),
			fmt.Sprintf("HTTP_PROXY=%s", proxyURL),
		)
	}

	return testexec.NewExec("protokol",
		testexec.WithArgs(args...),
		testexec.WithEnv(env),
		testexec.StderrTo(io.Discard),
	)
}
