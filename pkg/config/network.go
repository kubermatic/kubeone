package config

import (
	"fmt"
	"net"
)

// NetworkConfig describes the node network.
type NetworkConfig struct {
	PodSubnetVal     string `json:"pod_subnet"`
	ServiceSubnetVal string `json:"service_subnet"`
	NodePortRangeVal string `json:"node_port_range"`
}

// PodSubnet returns the pod subnet or the default value.
func (m *NetworkConfig) PodSubnet() string {
	if m.PodSubnetVal != "" {
		return m.PodSubnetVal
	}

	return "10.244.0.0/16"
}

// ServiceSubnet returns the service subnet or the default value.
func (m *NetworkConfig) ServiceSubnet() string {
	if m.ServiceSubnetVal != "" {
		return m.ServiceSubnetVal
	}

	return "10.96.0.0/12"
}

// NodePortRange returns the node port range or the default value.
func (m *NetworkConfig) NodePortRange() string {
	if m.NodePortRangeVal != "" {
		return m.NodePortRangeVal
	}

	return "30000-32767"
}

// Validate checks the NetworkConfig for errors
func (m *NetworkConfig) Validate() error {
	if m.PodSubnetVal != "" {
		if _, _, err := net.ParseCIDR(m.PodSubnetVal); err != nil {
			return fmt.Errorf("invalid pod subnet specified: %v", err)
		}
	}

	if m.ServiceSubnetVal != "" {
		if _, _, err := net.ParseCIDR(m.ServiceSubnetVal); err != nil {
			return fmt.Errorf("invalid service subnet specified: %v", err)
		}
	}

	return nil
}
