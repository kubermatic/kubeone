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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"k8c.io/kubeone/test/testexec"
)

const sonobuoyResultsFile = "results.tar.gz"

type sonobuoyReport struct {
	Name    string                 `json:"name"`
	Status  string                 `json:"status"`
	Details *sonobuoyReportDetails `json:"details,omitempty"`
	Items   []sonobuoyReport       `json:"items,omitempty"`
}

type sonobuoyReportDetails struct {
	Stdout   string   `json:"stdout,omitempty"`
	Messages []string `json:"messages,omitempty"`
}

type sonobuoyMode string

const (
	sonobuoyConformance     sonobuoyMode = "non-disruptive-conformance"
	sonobuoyConformanceLite sonobuoyMode = "conformance-lite"
	sonobuoyQuick           sonobuoyMode = "quick"
)

type sonobuoyBin struct {
	dir        string
	kubeconfig string
	proxyURL   string
}

func (sbb *sonobuoyBin) Run(mode sonobuoyMode) error {
	return sbb.run(context.Background(), "run", fmt.Sprintf("--mode=%s", mode))
}

func (sbb *sonobuoyBin) Wait(ctx context.Context) error {
	return sbb.run(ctx, "wait")
}

func (sbb *sonobuoyBin) Retrieve() error {
	return sbb.run(context.Background(), "retrieve", "--filename", sonobuoyResultsFile)
}

func (sbb *sonobuoyBin) Results() ([]sonobuoyReport, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rpipe, wpipe, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	exe := sbb.build("results", sonobuoyResultsFile, "--mode", "detailed", "--plugin", "e2e")
	cmd := exe.BuildCmd(ctx)
	cmd.Stdout = wpipe
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var waitErr error
	go func() {
		waitErr = cmd.Wait()
		_ = wpipe.Close() // send EOF to break the reading loop (with EOF), ignore the error
	}()

	dec := json.NewDecoder(rpipe)
	failedCases := []sonobuoyReport{}
	for {
		var rep sonobuoyReport
		if err := dec.Decode(&rep); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		if rep.Status == "failed" {
			// we are interested only in failed test cases
			failedCases = append(failedCases, rep)
		}
	}

	return failedCases, waitErr
}

func (sbb *sonobuoyBin) run(ctx context.Context, args ...string) error {
	return sbb.build(args...).BuildCmd(ctx).Run()
}

func (sbb *sonobuoyBin) build(args ...string) *testexec.Exec {
	exe := testexec.NewExec("sonobuoy",
		testexec.WithArgs(args...),
		testexec.WithEnv(os.Environ()),
		testexec.InDir(sbb.dir),
		testexec.StdoutDebug,
	)

	if sbb.kubeconfig != "" {
		testexec.WithEnvs(fmt.Sprintf("KUBECONFIG=%s", sbb.kubeconfig))(exe)
	}

	if sbb.proxyURL != "" {
		proxies := []string{
			fmt.Sprintf("HTTPS_PROXY=%s", sbb.proxyURL),
			fmt.Sprintf("HTTP_PROXY=%s", sbb.proxyURL),
		}
		testexec.WithEnv(proxies)(exe)
	}

	return exe
}
