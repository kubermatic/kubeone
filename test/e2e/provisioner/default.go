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

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/test/e2e/testutil"
)

// DefaultProvisioner contains default implementation of provisioner interface
type DefaultProvisioner struct {
	testPath  string
	terraform *terraform
}

// NewDefaultProvisioner creates and initialize universal provisioner
func NewDefaultProvisioner(creds func() error, testPath, identifier, provider string) (*DefaultProvisioner, error) {
	if err := creds(); err != nil {
		return nil, errors.Wrap(err, "unable to validate credentials")
	}

	terraform := &terraform{
		terraformDir: fmt.Sprintf("../../examples/terraform/%s/", provider),
		identifier:   identifier,
	}

	return &DefaultProvisioner{
		terraform: terraform,
		testPath:  testPath,
	}, nil
}

// Provision provisions a cluster using Terraform
func (p *DefaultProvisioner) Provision() (string, error) {
	tf, err := p.terraform.initAndApply()
	if err != nil {
		return "", err
	}

	return tf, nil
}

// Cleanup destroys infrastructure created by Terraform
func (p *DefaultProvisioner) Cleanup() error {
	err := p.terraform.destroy()
	if err != nil {
		return err
	}

	_, err = testutil.ExecuteCommand("", "rm", []string{"-rf", p.testPath}, nil)
	if err != nil {
		return err
	}

	return nil
}
