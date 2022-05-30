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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"k8c.io/kubeone/test/e2e/testutil"
)

type sonobuoyReport struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type sonobuoyMode string

const (
	sonobuoyQuick           sonobuoyMode = "quick"
	sonobuoyConformance     sonobuoyMode = "conformance"
	sonobuoyConformanceLite sonobuoyMode = "conformance-lite"
)

const (
	sonobuoyResultsFile = "results.tar.gz"
)

type sonobuoyBin struct {
	dir string
}

func (sbb *sonobuoyBin) Run(mode sonobuoyMode) error {
	return sbb.run("run", fmt.Sprintf("-mode=%s", mode))
}

func (sbb *sonobuoyBin) Wait() error {
	return sbb.run("wait")
}

func (sbb *sonobuoyBin) Retrive() error {
	return sbb.run("retrive", "--filename", sonobuoyResultsFile)
}

func (sbb *sonobuoyBin) Results() ([]sonobuoyReport, error) {
	rpipe, wpipe, _ := os.Pipe()
	failedCases := []sonobuoyReport{}

	exe := sbb.build("results", sonobuoyResultsFile, "--mode", "detailed", "--plugin", "e2e")
	testutil.StdoutTo(wpipe)(exe)
	var (
		runErr error
		wg     sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		runErr = exe.Run()
		wg.Done()
	}()

	dec := json.NewDecoder(rpipe)

	for {
		var rep sonobuoyReport
		if err := dec.Decode(&rep); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if rep.Status == "failed" {
			failedCases = append(failedCases, rep)
		}
	}

	wg.Wait()
	return failedCases, runErr
}

func (sbb *sonobuoyBin) run(args ...string) error {
	return sbb.build(args...).Run()
}

func (sbb *sonobuoyBin) build(args ...string) *testutil.Exec {
	return testutil.NewExec("sonobuoy",
		testutil.WithArgs(args...),
		testutil.WithEnv(os.Environ()),
		testutil.InDir(sbb.dir),
		testutil.WithDryRun(),
		testutil.StdoutDebug,
	)
}
