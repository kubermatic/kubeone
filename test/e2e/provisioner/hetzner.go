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
	"errors"
	"fmt"
	"os"

	"github.com/kubermatic/kubeone/test/e2e/testutil"
)

// HetznerProvisioner is provisioner used to provision cluster on Hetzner
type HetznerProvisioner struct {
	testPath  string
	terraform *terraform
}

// NewHetznerProvisioner creates and initialize the HetznerProvisioner structure
func NewHetznerProvisioner(testPath, identifier string) (*HetznerProvisioner, error) {
	terraform := &terraform{
		terraformDir: "../../examples/terraform/hetzner/",
		identifier:   identifier,
	}

	return &HetznerProvisioner{
		terraform: terraform,
		testPath:  testPath,
	}, nil
}

// Provision provisions a Hetzner cluster
func (p *HetznerProvisioner) Provision() (string, error) {
	hcloudToken := os.Getenv("HCLOUD_TOKEN")

	if len(hcloudToken) == 0 {
		return "", errors.New("unable to run the test suite, HCLOUD_TOKEN environment variable cannot be empty")
	}

	tf, err := p.terraform.initAndApply()
	if err != nil {
		return "", err
	}

	return tf, nil
}

// Cleanup destroys infrastructure created by Terraform
func (p *HetznerProvisioner) Cleanup() error {
	err := p.terraform.destroy()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	_, err = testutil.ExecuteCommand("", "rm", []string{"-rf", p.testPath}, nil)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}
