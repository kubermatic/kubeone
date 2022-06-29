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

package e2e

import (
	"errors"
	"fmt"

	"k8c.io/kubeone/test/e2e/provisioner"
)

type OperatingSystem string

const (
	OperatingSystemUbuntu     OperatingSystem = "ubuntu"
	OperatingSystemCentOS7    OperatingSystem = "centos7"
	OperatingSystemCentOS8    OperatingSystem = "centos"
	OperatingSystemFlatcar    OperatingSystem = "flatcar"
	OperatingSystemAmazon     OperatingSystem = "amzn"
	OperatingSystemAmazonMC   OperatingSystem = "amzn2"
	OperatingSystemRHEL       OperatingSystem = "rhel"
	OperatingSystemRockyLinux OperatingSystem = "rockylinux"
	OperatingSystemDefault    OperatingSystem = ""
)

const (
	AWSCentOS7AMI = "ami-04552009264cbe9f4"
)

func ValidateOperatingSystem(osName string) error {
	switch OperatingSystem(osName) {
	case OperatingSystemUbuntu,
		OperatingSystemFlatcar,
		OperatingSystemCentOS7,
		OperatingSystemCentOS8,
		OperatingSystemAmazon,
		OperatingSystemAmazonMC,
		OperatingSystemRHEL,
		OperatingSystemRockyLinux,
		OperatingSystemDefault:
		return nil
	}

	return errors.New("failed to validate operating system")
}

// ControlPlaneImageFlags returns Terraform flags for control plane image and SSH username
func ControlPlaneImageFlags(provider string, osName OperatingSystem) ([]string, error) {
	if provider == provisioner.AWS {
		switch {
		case osName == OperatingSystemCentOS7:
			return []string{
				"-var", fmt.Sprintf("ami=%s", AWSCentOS7AMI),
				"-var", "os=centos",
				"-var", "ssh_username=centos",
				"-var", "bastion_user=centos",
			}, nil
		default:
			return []string{
				"-var", fmt.Sprintf("os=%s", osName),
			}, nil
		}
	}

	return nil, errors.New("custom operating system is not supported for selected provider")
}
