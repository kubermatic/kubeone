package config

import (
	"fmt"
	"os"
	"strings"
)

type providerConfig struct {
	CloudProviderSpec   map[string]interface{} `json:"cloudProviderSpec"`
	Labels              map[string]string      `json:"labels"`
	SSHPublicKeys       []string               `json:"sshPublicKeys"`
	OperatingSystem     string                 `json:"operatingSystem"`
	OperatingSystemSpec map[string]interface{} `json:"operatingSystemSpec"`
}

// ProviderConfig describes the cloud provider that is running the machines.
type ProviderConfig struct {
	Name        ProviderName      `json:"name"`
	CloudConfig string            `json:"cloud_config"`
	Credentials map[string]string `json:"credentials"`
}

// Validate checks the ProviderConfig for errors
func (p *ProviderConfig) Validate() error {
	switch p.Name {
	case ProviderNameAWS, ProviderNameOpenStack, ProviderNameHetzner, ProviderNameDigitalOcean, ProviderNameVSphere:
	default:
		return fmt.Errorf("unknown provider name %q", p.Name)
	}

	for _, varName := range p.Name.CredentialsEnvironmentVariables() {
		if p.Credentials[varName] == "" {
			return fmt.Errorf("environment variable %s is not set", varName)
		}
	}

	return nil
}

// ApplyEnvironment reads cloud provider credentials from
// environment variables.
func (p *ProviderConfig) ApplyEnvironment() error {
	if p.Credentials == nil {
		p.Credentials = make(map[string]string)
	}

	for _, varName := range p.Name.CredentialsEnvironmentVariables() {
		p.Credentials[varName] = strings.TrimSpace(os.Getenv(varName))
	}

	return nil
}

// ProviderName represents the name of an provider
type ProviderName string

// ProviderName values
const (
	ProviderNameAWS          ProviderName = "aws"
	ProviderNameOpenStack    ProviderName = "openstack"
	ProviderNameHetzner      ProviderName = "hetzner"
	ProviderNameDigitalOcean ProviderName = "digitalocean"
	ProviderNameVSphere      ProviderName = "vshere"
)

func (p ProviderName) CredentialsEnvironmentVariables() []string {
	switch p {
	case ProviderNameAWS:
		return []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}
	case ProviderNameOpenStack:
		return []string{"OS_AUTH_URL", "OS_USER_NAME", "OS_PASSWORD", "OS_DOMAIN_NAME", "OS_TENANT_NAME"}
	case ProviderNameHetzner:
		return []string{"HZ_TOKEN"}
	case ProviderNameDigitalOcean:
		return []string{"DO_TOKEN"}
	case ProviderNameVSphere:
		return []string{"VSPHERE_ADDRESS", "VSPHERE_USERNAME", "VSPHERE_PASSWORD"}
	}

	return nil
}
