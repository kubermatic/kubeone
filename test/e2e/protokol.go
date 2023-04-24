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

	"k8c.io/kubeone/test/testexec"
)

type protokolBin struct {
	namespaces []string
	outputDir  string
}

func (p *protokolBin) Start(ctx context.Context, kubeconfigPath string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("protokol start: %w", err)
	}

	if len(p.namespaces) == 0 {
		return errors.New("refusing to dump *everything*, please specify namespaces")
	}

	args := []string{"--output", p.outputDir, "--kubeconfig", kubeconfigPath}
	for _, ns := range p.namespaces {
		args = append(args, "--namespace", ns)
	}

	exe := p.build(args...).BuildCmd(ctx)

	if err := exe.Start(); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("protokol apply: %w", err)
	}

	return nil
}

func (p *protokolBin) build(args ...string) *testexec.Exec {
	return testexec.NewExec("protokol",
		testexec.WithArgs(args...),
		testexec.WithEnv(os.Environ()),
		testexec.StderrDebug,
	)
}
