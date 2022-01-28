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
	OperatingSystemUbuntu  OperatingSystem = "ubuntu"
	OperatingSystemCentOS7 OperatingSystem = "centos7"
	OperatingSystemCentOS8 OperatingSystem = "centos"
	OperatingSystemFlatcar OperatingSystem = "flatcar"
	OperatingSystemAmazon  OperatingSystem = "amzn2"
	OperatingSystemRHEL    OperatingSystem = "rhel"
	OperatingSystemDefault OperatingSystem = ""
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
		OperatingSystemRHEL,
		OperatingSystemDefault:
		return nil
	}

	return errors.New("failed to validate operating system")
}

// ControlPlaneImageFlags returns Terraform flags for control plane image and SSH username
func ControlPlaneImageFlags(provider string, osName OperatingSystem) ([]string, error) {
	if provider == provisioner.AWS {
		user, err := sshUsername(osName)
		if err != nil {
			return nil, err
		}

		switch {
		case osName == OperatingSystemCentOS7:
			return []string{
				"-var", fmt.Sprintf("ami=%s", AWSCentOS7AMI),
				"-var", fmt.Sprintf("ssh_username=%s", user),
				"-var", fmt.Sprintf("bastion_user=%s", user),
			}, nil
		default:
			return []string{
				"-var", fmt.Sprintf("os=%s", osName),
				"-var", fmt.Sprintf("ssh_username=%s", user),
				"-var", fmt.Sprintf("bastion_user=%s", user),
			}, nil
		}
	}

	return nil, errors.New("custom operating system is not supported for selected provider")
}

func sshUsername(osName OperatingSystem) (string, error) {
	switch osName {
	case OperatingSystemUbuntu:
		return "ubuntu", nil
	case OperatingSystemCentOS7, OperatingSystemCentOS8:
		return "centos", nil
	case OperatingSystemFlatcar:
		return "core", nil
	case OperatingSystemRHEL, OperatingSystemAmazon:
		return "ec2-user", nil
	case OperatingSystemDefault:
	}

	return "", errors.New("operating system not matched")
}
