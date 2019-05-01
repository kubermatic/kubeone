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
	DigitalOceanTokenKey    = "DO_TOKEN"
	GoogleServiceAccountKey = "GOOGLE_SERVICE_ACCOUNT"
	HetznerTokenKey         = "HZ_TOKEN"
	OpenStackAuthURL        = "OS_AUTH_URL"
	OpenStackDomainName     = "OS_DOMAIN_NAME"
	OpenStackPassword       = "OS_PASSWORD"
	OpenStackTenantName     = "OS_TENANT_NAME"
	OpenStackUserName       = "OS_USER_NAME"
	VSphereAddress          = "VSPHERE_ADDRESS"
	VSpherePasswords        = "VSPHERE_PASSWORD"
	VSphereUsername         = "VSPHERE_USERNAME"
)

// ProviderEnvironmentVariable is used to match environment variable used by KubeOne to environment variable used by
// machine-controller.
type ProviderEnvironmentVariable struct {
	Name                  string
	MachineControllerName string
}

// ProviderCredentials match the cloudprovider and parses its credentials from environment
func ProviderCredentials(p kubeone.CloudProviderName) (map[string]string, error) {
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
	case kubeone.CloudProviderNameOpenStack:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "OS_AUTH_URL"},
			{Name: "OS_USERNAME", MachineControllerName: "OS_USER_NAME"},
			{Name: "OS_PASSWORD"},
			{Name: "OS_DOMAIN_NAME"},
			{Name: "OS_TENANT_NAME"},
		})
	case kubeone.CloudProviderNameHetzner:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "HCLOUD_TOKEN", MachineControllerName: "HZ_TOKEN"},
		})
	case kubeone.CloudProviderNameDigitalOcean:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "DIGITALOCEAN_TOKEN", MachineControllerName: "DO_TOKEN"},
		})
	case kubeone.CloudProviderNameGCE:
		gsa, err := parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "GOOGLE_CREDENTIALS", MachineControllerName: "GOOGLE_SERVICE_ACCOUNT"},
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		// encode it before sending to secret to be consumed by
		// machine-controller, as machine-controller assumes it will be double encoded
		gsa["GOOGLE_SERVICE_ACCOUNT"] = base64.StdEncoding.EncodeToString([]byte(gsa["GOOGLE_SERVICE_ACCOUNT"]))
		return gsa, nil
	case kubeone.CloudProviderNameVSphere:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "VSPHERE_ADDRESS"},
			{Name: "VSPHERE_USERNAME"},
			{Name: "VSPHERE_PASSWORD"},
		})
	}

	return nil, errors.New("no provider matched")
}

func parseCredentialVariables(envVars []ProviderEnvironmentVariable) (map[string]string, error) {
	creds := make(map[string]string)
	for _, env := range envVars {
		if len(env.MachineControllerName) == 0 {
			env.MachineControllerName = env.Name
		}
		creds[env.MachineControllerName] = strings.TrimSpace(os.Getenv(env.Name))
		if creds[env.MachineControllerName] == "" {
			return nil, errors.Errorf("environment variable %s is not set, but is required", env.Name)
		}
	}
	return creds, nil
}
