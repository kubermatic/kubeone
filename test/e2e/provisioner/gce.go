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
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/test/e2e/testutil"
)

// GCEProvisioner contains implementation of GCE provisioner interface
type GCEProvisioner struct {
	testPath  string
	terraform *terraform

	init bool
}

// NewGCEProvisioner creates and initialize GCE provisioner
func NewGCEProvisioner(creds func() error, testPath, identifier string) (*GCEProvisioner, error) {
	if err := creds(); err != nil {
		return nil, errors.Wrap(err, "unable to validate credentials")
	}

	terraform := &terraform{
		terraformDir: "../../examples/terraform/gce/",
		identifier:   identifier,
	}

	return &GCEProvisioner{
		terraform: terraform,
		testPath:  testPath,

		init: false,
	}, nil
}

// Provision provisions a cluster using Terraform
func (p *GCEProvisioner) Provision() (string, error) {
	args := []string{}
	if !p.init {
		args = []string{"-var", "control_plane_target_pool_members_count=1"}
		p.init = true
	}
	tf, err := p.terraform.initAndApply(args)
	if err != nil {
		return "", err
	}

	return tf, nil
}

// Cleanup destroys infrastructure created by Terraform
func (p *GCEProvisioner) Cleanup() error {
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
