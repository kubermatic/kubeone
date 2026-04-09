/*
Copyright 2025 The KubeOne Authors.

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
	"net"

	cloud "k8c.io/machine-controller/pkg/cloudprovider/instance"

	corev1 "k8s.io/api/core/v1"
)

type Machine struct {
	PublicAddress  string `json:"public_address,omitempty"`
	PrivateAddress string `json:"private_address,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
}

func getMachineProvisionerOutput(instances []cloud.Instance) []Machine {
	var out []Machine

	for _, instance := range instances {
		machine := GetMachineInfo(instance)
		out = append(out, machine)
	}

	return out
}

func GetMachineInfo(instance cloud.Instance) Machine {
	var publicAddress, privateAddress, hostname string
	var publicAddressIPv6, privateAddressIPv6 string

	for address, addressType := range instance.Addresses() {
		switch addressType {
		case corev1.NodeExternalIP:
			if ip := net.ParseIP(address); ip != nil && ip.To4() != nil {
				if publicAddress == "" {
					publicAddress = address
				}
			} else if ip != nil {
				if publicAddressIPv6 == "" {
					publicAddressIPv6 = address
				}
			}
		case corev1.NodeInternalIP:
			if ip := net.ParseIP(address); ip != nil && ip.To4() != nil {
				if privateAddress == "" {
					privateAddress = address
				}
			} else if ip != nil {
				if privateAddressIPv6 == "" {
					privateAddressIPv6 = address
				}
			}
		case corev1.NodeHostName:
			if hostname == "" {
				hostname = address
			}
		case corev1.NodeInternalDNS:
			if hostname == "" {
				hostname = address
			}
		case corev1.NodeExternalDNS:
			if hostname == "" {
				hostname = address
			}
		}
	}

	if publicAddress == "" {
		publicAddress = publicAddressIPv6
	}
	if privateAddress == "" {
		privateAddress = privateAddressIPv6
	}

	return Machine{
		PublicAddress:  publicAddress,
		PrivateAddress: privateAddress,
		Hostname:       hostname,
	}
}

func publicAndPrivateIPExist(addresses map[string]corev1.NodeAddressType) bool {
	var publicIPExists, privateIPExists bool
	// we only care about ExternalIP and InternalIP specifically, thus nolint
	//nolint:exhaustive
	for _, addressType := range addresses {
		switch addressType {
		case corev1.NodeExternalIP:
			publicIPExists = true
		case corev1.NodeInternalIP:
			privateIPExists = true
		}
	}

	return publicIPExists && privateIPExists
}
