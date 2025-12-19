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
	corev1 "k8s.io/api/core/v1"
)

type Machine struct {
	PublicAddress  string `json:"public_address,omitempty"`
	PrivateAddress string `json:"private_address,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	SSHUser        string `json:"ssh_user,omitempty"`
	Bastion        bool   `json:"bastion,omitempty"`
}

func getMachineProvisionerOutput(instances []MachineInstance) []Machine {
	var out []Machine

	for _, instance := range instances {
		machine := getMachineInfo(instance)
		out = append(out, machine)
	}

	return out
}

func getMachineInfo(instance MachineInstance) Machine {
	var publicAddress, privateAddress, hostname string
	for address, addressType := range instance.inst.Addresses() {
		switch addressType {
		case corev1.NodeExternalIP:
			publicAddress = address
		case corev1.NodeInternalIP:
			privateAddress = address
		case corev1.NodeHostName:
			hostname = address
		case corev1.NodeInternalDNS:
			hostname = address
		case corev1.NodeExternalDNS:
			if hostname == "" {
				hostname = address
			}
		}
	}

	return Machine{
		PublicAddress:  publicAddress,
		PrivateAddress: privateAddress,
		Hostname:       hostname,
		SSHUser:        instance.sshUser,
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
