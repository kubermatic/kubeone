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

package credentials

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
)

// The environment variable names with credential in them that machine-controller expects to see
const (
	AWSAccessKeyID          = "AWS_ACCESS_KEY_ID"
	AWSSecretAccessKey      = "AWS_SECRET_ACCESS_KEY"
	AzureClientID           = "ARM_CLIENT_ID"
	AzureClientSecret       = "ARM_CLIENT_SECRET"
	AzureTenantID           = "ARM_TENANT_ID"
	AzureSubscribtionID     = "ARM_SUBSCRIPTION_ID"
	DigitalOceanTokenKey    = "DO_TOKEN"
	GoogleServiceAccountKey = "GOOGLE_SERVICE_ACCOUNT"
	HetznerTokenKey         = "HZ_TOKEN"
	OpenStackAuthURL        = "OS_AUTH_URL"
	OpenStackDomainName     = "OS_DOMAIN_NAME"
	OpenStackPassword       = "OS_PASSWORD"
	OpenStackRegionName     = "OS_REGION_NAME"
	OpenStackTenantID       = "OS_TENANT_ID"
	OpenStackTenantName     = "OS_TENANT_NAME"
	OpenStackUserName       = "OS_USER_NAME"
	PacketAPIKey            = "PACKET_API_KEY"
	PacketProjectID         = "PACKET_PROJECT_ID"
	VSphereAddress          = "VSPHERE_ADDRESS"
	VSpherePassword         = "VSPHERE_PASSWORD"
	VSphereUsername         = "VSPHERE_USERNAME"
)

// ProviderEnvironmentVariable is used to match environment variable used by KubeOne to environment variable used by
// machine-controller.
type ProviderEnvironmentVariable struct {
	Name                  string
	MachineControllerName string
}

// ProviderCredentials match the cloudprovider and parses its credentials from environment
func ProviderCredentials(p kubeone.CloudProviderName, s *kubeone.KubeOneSecrets) (map[string]string, error) {
	switch p {
	case kubeone.CloudProviderNameAWS:
		creds := make(map[string]string)
		envCredsProvider := credentials.NewEnvCredentials()
		envCreds, err := envCredsProvider.Get()
		if err != nil {
			return nil, err
		}
		if envCreds.AccessKeyID != "" && envCreds.SecretAccessKey != "" {
			creds["AWS_ACCESS_KEY_ID"] = envCreds.AccessKeyID
			creds["AWS_SECRET_ACCESS_KEY"] = envCreds.SecretAccessKey
			return creds, nil
		}

		// If env fails resort to config file
		configCredsProvider := credentials.NewSharedCredentials("", "")
		configCreds, err := configCredsProvider.Get()
		if err != nil {
			return nil, err
		}
		if configCreds.AccessKeyID != "" && configCreds.SecretAccessKey != "" {
			creds["AWS_ACCESS_KEY_ID"] = configCreds.AccessKeyID
			creds["AWS_SECRET_ACCESS_KEY"] = configCreds.SecretAccessKey
			return creds, nil
		}

		return nil, errors.New("error parsing aws credentials")
	case kubeone.CloudProviderNameAzure:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: AzureClientID, MachineControllerName: "AZURE_CLIENT_ID"},
			{Name: AzureClientSecret, MachineControllerName: "AZURE_CLIENT_SECRET"},
			{Name: AzureTenantID, MachineControllerName: "AZURE_TENANT_ID"},
			{Name: AzureSubscribtionID, MachineControllerName: "AZURE_SUBSCRIPTION_ID"},
		}, s, defaultValidationFunc)
	case kubeone.CloudProviderNameOpenStack:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "OS_AUTH_URL"},
			{Name: "OS_USERNAME", MachineControllerName: "OS_USER_NAME"},
			{Name: "OS_PASSWORD"},
			{Name: "OS_DOMAIN_NAME"},
			{Name: "OS_REGION_NAME"},
			{Name: "OS_TENANT_ID"},
			{Name: "OS_TENANT_NAME"},
		}, s, openstackValidationFunc)
	case kubeone.CloudProviderNameHetzner:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "HCLOUD_TOKEN", MachineControllerName: "HZ_TOKEN"},
		}, s, defaultValidationFunc)
	case kubeone.CloudProviderNameDigitalOcean:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "DIGITALOCEAN_TOKEN", MachineControllerName: "DO_TOKEN"},
		}, s, defaultValidationFunc)
	case kubeone.CloudProviderNameGCE:
		gsa, err := parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "GOOGLE_CREDENTIALS", MachineControllerName: "GOOGLE_SERVICE_ACCOUNT"},
		}, s, defaultValidationFunc)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		// encode it before sending to secret to be consumed by
		// machine-controller, as machine-controller assumes it will be double encoded
		gsa["GOOGLE_SERVICE_ACCOUNT"] = base64.StdEncoding.EncodeToString([]byte(gsa["GOOGLE_SERVICE_ACCOUNT"]))
		return gsa, nil
	case kubeone.CloudProviderNamePacket:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "PACKET_AUTH_TOKEN", MachineControllerName: PacketAPIKey},
			{Name: PacketProjectID},
		}, s, defaultValidationFunc)
	case kubeone.CloudProviderNameVSphere:
		vscreds, err := parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "VSPHERE_SERVER", MachineControllerName: VSphereAddress},
			{Name: "VSPHERE_USER", MachineControllerName: VSphereUsername},
			{Name: VSpherePassword},
		}, s, defaultValidationFunc)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// force scheme, as machine-controller requires it while terraform does not
		vscreds[VSphereAddress] = "https://" + vscreds[VSphereAddress]
		return vscreds, nil
	}

	return nil, errors.New("no provider matched")
}

func parseCredentialVariables(envVars []ProviderEnvironmentVariable, secrets *kubeone.KubeOneSecrets, validationFunc func(map[string]string) error) (map[string]string, error) {
	// Validate credentials using given validation function
	creds := make(map[string]string)
	for _, env := range envVars {
		if secrets != nil {
			if v, ok := secrets.Secrets[env.Name]; ok {
				creds[env.Name] = v
			}
		}
		if len(creds[env.Name]) == 0 {
			creds[env.Name] = strings.TrimSpace(os.Getenv(env.Name))
		}
	}
	if err := validationFunc(creds); err != nil {
		return nil, errors.Wrap(err, "unable to validate credentials")
	}

	// Prepare credentials to be used by machine-controller
	mcCreds := make(map[string]string)
	for _, env := range envVars {
		name := env.MachineControllerName
		if len(name) == 0 {
			name = env.Name
		}
		mcCreds[name] = creds[env.Name]
	}

	return mcCreds, nil
}

func defaultValidationFunc(creds map[string]string) error {
	for k, v := range creds {
		if len(v) == 0 {
			return errors.Errorf("key %v is required but isn't present", k)
		}
	}
	return nil
}

func openstackValidationFunc(creds map[string]string) error {
	for k, v := range creds {
		if k == OpenStackTenantID || k == OpenStackTenantName {
			continue
		}
		if len(v) == 0 {
			return errors.Errorf("key %v is required but isn't present", k)
		}
	}

	if v, ok := creds[OpenStackTenantID]; !ok || len(v) == 0 {
		if v, ok := creds[OpenStackTenantName]; !ok || len(v) == 0 {
			return errors.Errorf("key %v or %v is required but isn't present", OpenStackTenantID, OpenStackTenantName)
		}
	}

	return nil
}
