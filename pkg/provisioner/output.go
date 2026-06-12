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

	"k8c.io/kubeone/pkg/maputils"
	cloud "k8c.io/machine-controller/pkg/cloudprovider/instance"

	corev1 "k8s.io/api/core/v1"
)

var privateIPNets = []*net.IPNet{
	mustParseCIDR("10.0.0.0/8"),
	mustParseCIDR("172.16.0.0/12"),
	mustParseCIDR("192.168.0.0/16"),
	mustParseCIDR("100.64.0.0/10"),
	mustParseCIDR("fc00::/7"),
	mustParseCIDR("fe80::/10"),
}

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

	for address, addressType := range maputils.IterateInOrder(instance.Addresses()) {
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
		case "":
			// we will try to guess the type
			ip := net.ParseIP(address)
			if ip == nil {
				// not an IP, guess this is a hostname
				if hostname != "" {
					hostname = address
				}

				continue
			}

			ipv4 := ip.To4()
			if isPrivateIP(ip) {
				if ipv4 != nil {
					if privateAddress == "" {
						privateAddress = address
					}
				} else if privateAddressIPv6 == "" {
					privateAddressIPv6 = address
				}
				continue
			} else {
				if ipv4 != nil {
					if publicAddress == "" {
						publicAddress = address
					}
				} else if publicAddressIPv6 == "" {
					publicAddressIPv6 = address
				}
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

	for _, addressType := range addresses {
		// we only care about ExternalIP and InternalIP specifically, thus nolint
		//nolint:exhaustive
		switch addressType {
		case corev1.NodeExternalIP:
			publicIPExists = true
		case corev1.NodeInternalIP:
			privateIPExists = true
		}
	}

	return publicIPExists && privateIPExists
}

func mustParseCIDR(cidr string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}

	return ipNet
}

func isPrivateIP(ip net.IP) bool {
	for _, network := range privateIPNets {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
