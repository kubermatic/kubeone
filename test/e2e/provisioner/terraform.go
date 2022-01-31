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

package provisioner

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"k8c.io/kubeone/test/e2e/testutil"
)

const (
	applyRetryNumber  = 5
	applyRetryTimeout = 30 * time.Second
)

// terraform contains information needed to provision infrastructure using Terraform
type terraform struct {
	// terraformDir is the path to the Terraform scripts
	terraformDir string
	// identifier is an unique identifier for the test run
	identifier string
}

// initAndApply method initializes a Terraform working directory and applies scripts
func (p *terraform) initAndApply(applyArgs ...string) (string, error) {
	initCmd := []string{"init"}
	if len(p.identifier) > 0 {
		initCmd = append(initCmd, fmt.Sprintf("--backend-config=key=%s", p.identifier))
	}

	err := p.run(initCmd...)
	if err != nil {
		return "", fmt.Errorf("terraform init command failed: %w", err)
	}

	args := []string{"apply", "-auto-approve"}
	if applyArgs != nil {
		args = append(args, applyArgs...)
	}

	var applyErr error
	for i := 0; i < applyRetryNumber; i++ {
		// In case when apply fails due to the CIDR conflict error, we need to destroy resources and start over
		// or otherwise apply will always fail.
		// This is because the random_integer resource used for the CIDR creation is not recreated on the subsequent
		// runs of terraform apply, so terraform always tries to create the same CIDR.
		destroyErr := p.destroy()
		if destroyErr != nil {
			return "", fmt.Errorf("terraform destroy command failed: %w", destroyErr)
		}
		applyErr = p.run(args...)
		if applyErr == nil {
			break
		}
		time.Sleep(applyRetryTimeout)
	}

	if applyErr != nil {
		return "", fmt.Errorf("terraform apply command failed: %w", applyErr)
	}

	return p.getTFJson()
}

// destroy destories the infrastructure
func (p *terraform) destroy() error {
	err := p.run("destroy", "-auto-approve")
	if err != nil {
		return fmt.Errorf("terraform destroy command failed: %w", err)
	}

	return nil
}

// getTFJson reads an output from a state file
func (p *terraform) getTFJson() (string, error) {
	var jsonBuf bytes.Buffer

	tfcmd := p.build("output", fmt.Sprintf("-state=%v", tfStateFileName), "-json")
	testutil.StdoutTo(&jsonBuf)(tfcmd)

	if err := tfcmd.Run(); err != nil {
		return "", fmt.Errorf("generating tf json failed: %w", err)
	}

	return jsonBuf.String(), nil
}

func (p *terraform) build(args ...string) *testutil.Exec {
	return testutil.NewExec("terraform",
		testutil.WithArgs(args...),
		testutil.WithEnv(os.Environ()),
		testutil.InDir(p.terraformDir),
	)
}

func (p *terraform) run(args ...string) error {
	return p.build(args...).Run()
}
