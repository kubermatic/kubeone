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
	"os"

	"github.com/pkg/errors"
)

const (
	// AWS cloud provider
	AWS = "aws"
	// DigitalOcean cloud provider
	DigitalOcean = "digitalocean"
	// Hetzner cloud provider
	Hetzner = "hetzner"
	// GCE cloud provider
	GCE = "gce"
	// Equinix Metal cloud provider
	EquinixMetal = "equinixmetal"
	// OpenStack provider
	OpenStack = "openstack"

	// tfStateFileName is name of the Terraform state file
	tfStateFileName = "terraform.tfstate"
)

// Provisioner contains cluster management operations such as provision and cleanup
type Provisioner interface {
	Provision(args ...string) (string, error)
	Cleanup() error
}

// CreateProvisioner returns interface for specific provisioner
func CreateProvisioner(testPath string, identifier string, provider string) (Provisioner, error) {
	switch provider {
	case AWS:
		creds := verifyCredentials("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY")

		return NewDefaultProvisioner(creds, testPath, identifier, provider)
	case DigitalOcean:
		creds := verifyCredentials("DIGITALOCEAN_TOKEN")

		return NewDefaultProvisioner(creds, testPath, identifier, provider)
	case Hetzner:
		creds := verifyCredentials("HCLOUD_TOKEN")

		return NewDefaultProvisioner(creds, testPath, identifier, provider)
	case GCE:
		creds := verifyCredentials("GOOGLE_CREDENTIALS")

		return NewDefaultProvisioner(creds, testPath, identifier, provider)
	case EquinixMetal:
		creds := verifyCredentials("METAL_AUTH_TOKEN", "METAL_PROJECT_ID")

		return NewDefaultProvisioner(creds, testPath, identifier, provider)
	case OpenStack:
		creds := verifyCredentials("OS_AUTH_URL", "OS_DOMAIN_NAME", "OS_PASSWORD",
			"OS_REGION_NAME", "OS_TENANT_NAME", "OS_USERNAME")

		return NewDefaultProvisioner(creds, testPath, identifier, provider)
	default:
		return nil, fmt.Errorf("unsupported provider %v", provider)
	}
}

func verifyCredentials(envs ...string) func() error {
	return func() error {
		for _, env := range envs {
			_, ok := os.LookupEnv(env)
			if !ok {
				return errors.New("key not found")
			}
		}

		return nil
	}
}
