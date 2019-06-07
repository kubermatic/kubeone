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
	"fmt"

	"github.com/kubermatic/kubeone/test/e2e/testutil"
)

const (
	// AWS cloud provider
	AWS = "aws"
	// DigitalOcean cloud provider
	DigitalOcean = "digitalocean"
	// Hetzner cloud provider
	Hetzner = "hetzner"

	// tfStateFileName is name of the Terraform state file
	tfStateFileName = "terraform.tfstate"
)

// Provisioner contains cluster management operations such as provision and cleanup
type Provisioner interface {
	Provision() (string, error)
	Cleanup() error
}

// CreateProvisioner returns interface for specific provisioner
func CreateProvisioner(testPath string, identifier string, provider string) (Provisioner, error) {
	switch provider {
	case AWS:
		return NewAWSProvisioner(testPath, identifier)
	case DigitalOcean:
		return NewDOProvisioner(testPath, identifier)
	case Hetzner:
		return NewHetznerProvisioner(testPath, identifier)
	default:
		return nil, fmt.Errorf("unsuported provider %v", provider)
	}
}

// terraform contains information needed to provision infrastructure using Terraform
type terraform struct {
	// terraformDir is the path to the Terraform scripts
	terraformDir string
	// identifier is an unique identifier for the test run
	identifier string
}

// initAndApply method initializes a Terraform working directory and applies scripts
func (p *terraform) initAndApply() (string, error) {
	initCmd := []string{"init"}
	if len(p.identifier) > 0 {
		initCmd = append(initCmd, fmt.Sprintf("--backend-config=key=%s", p.identifier))
	}

	_, err := testutil.ExecuteCommand(p.terraformDir, "terraform", initCmd, nil)
	if err != nil {
		return "", fmt.Errorf("terraform init command failed: %v", err)
	}

	_, err = testutil.ExecuteCommand(p.terraformDir, "terraform", []string{"apply", "-auto-approve"}, nil)
	if err != nil {
		return "", fmt.Errorf("terraform apply command failed: %v", err)
	}

	return p.getTFJson()
}

// destroy destories the infrastructure
func (p *terraform) destroy() error {
	_, err := testutil.ExecuteCommand(p.terraformDir, "terraform", []string{"destroy", "-auto-approve"}, nil)
	if err != nil {
		return fmt.Errorf("terraform destroy command failed: %v", err)
	}
	return nil
}

// getTFJson reads an output from a state file
func (p *terraform) getTFJson() (string, error) {
	tf, err := testutil.ExecuteCommand(p.terraformDir, "terraform", []string{"output", fmt.Sprintf("-state=%v", tfStateFileName), "-json"}, nil)
	if err != nil {
		return "", fmt.Errorf("generating tf json failed: %v", err)
	}

	return tf, nil
}
