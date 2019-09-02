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

	"github.com/kubermatic/kubeone/test/e2e/provisioner"
)

type OperatingSystem string

const (
	OperatingSystemUbuntu  OperatingSystem = "ubuntu"
	OperatingSystemCentOS  OperatingSystem = "centos"
	OperatingSystemCoreOS  OperatingSystem = "coreos"
	OperatingSystemDefault OperatingSystem = ""
)

func ValidateOperatingSystem(osName string) error {
	switch OperatingSystem(osName) {
	case OperatingSystemUbuntu, OperatingSystemCentOS, OperatingSystemCoreOS, OperatingSystemDefault:
		return nil
	}
	return errors.New("failed to validate operating system")
}

// DiscoverControlPlaneOSImage returns Terraform flag with the image to be used for provisioning
func DiscoverControlPlaneOSImage(provider string, osName OperatingSystem) (string, error) {
	if provider == provisioner.AWS {
		img, err := discoverAWSImage(osName)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("ami=%s", img), nil
	}
	return "", errors.New("custom operating system is not supported for selected provider")
}

func discoverAWSImage(osName OperatingSystem) (string, error) {
	switch osName {
	case OperatingSystemUbuntu:
		return "ami-0119667e27598718e", nil
	case OperatingSystemCentOS:
		return "ami-0e1ab783dc9489f34", nil
	case OperatingSystemCoreOS:
		return "ami-04de4c2943ebaa320", nil
	}
	return "", errors.New("operating system not matched")
}
