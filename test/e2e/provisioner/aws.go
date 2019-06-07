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

// AWSProvisioner is provisioner used to provision cluster on AWS
type AWSProvisioner struct {
	testPath  string
	terraform *terraform
}

// NewAWSProvisioner creates and initialize AWSProvisioner structure
func NewAWSProvisioner(testPath, identifier string) (*AWSProvisioner, error) {
	terraform := &terraform{
		terraformDir: "../../examples/terraform/aws/",
		identifier:   identifier,
	}

	return &AWSProvisioner{
		terraform: terraform,
		testPath:  testPath,
	}, nil
}

// Provision provisions an AWS cluster
func (p *AWSProvisioner) Provision() (string, error) {
	awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if len(awsKeyID) == 0 || len(awsSecret) == 0 {
		return "", errors.New("unable to run the test suite, AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY environment variables cannot be empty")
	}

	tf, err := p.terraform.initAndApply()
	if err != nil {
		return "", err
	}

	return tf, nil
}

// Cleanup destroys infrastructure created by Terraform
func (p *AWSProvisioner) Cleanup() error {
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
