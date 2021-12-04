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

	awscredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"k8c.io/kubeone/pkg/apis/kubeone"
)

// The environment variable names with credential in them
const (
	// Variables that KubeOne (and Terraform) expect to see
	AWSAccessKeyID          = "AWS_ACCESS_KEY_ID"
	AWSSecretAccessKey      = "AWS_SECRET_ACCESS_KEY" //nolint:gosec
	AzureClientID           = "ARM_CLIENT_ID"
	AzureClientSecret       = "ARM_CLIENT_SECRET" //nolint:gosec
	AzureTenantID           = "ARM_TENANT_ID"
	AzureSubscribtionID     = "ARM_SUBSCRIPTION_ID"
	DigitalOceanTokenKey    = "DIGITALOCEAN_TOKEN"
	GoogleServiceAccountKey = "GOOGLE_CREDENTIALS"
	HetznerTokenKey         = "HCLOUD_TOKEN"
	OpenStackAuthURL        = "OS_AUTH_URL"
	OpenStackDomainName     = "OS_DOMAIN_NAME"
	OpenStackPassword       = "OS_PASSWORD"
	OpenStackRegionName     = "OS_REGION_NAME"
	OpenStackTenantID       = "OS_TENANT_ID"
	OpenStackTenantName     = "OS_TENANT_NAME"
	OpenStackUserName       = "OS_USERNAME"
	PacketAPIKey            = "PACKET_AUTH_TOKEN" //nolint:gosec
	PacketProjectID         = "PACKET_PROJECT_ID"
	VSphereAddress          = "VSPHERE_SERVER"
	VSpherePassword         = "VSPHERE_PASSWORD"
	VSphereUsername         = "VSPHERE_USER"

	// Variables that machine-controller expects
	AzureClientIDMC           = "AZURE_CLIENT_ID"
	AzureClientSecretMC       = "AZURE_CLIENT_SECRET" //nolint:gosec
	AzureTenantIDMC           = "AZURE_TENANT_ID"
	AzureSubscribtionIDMC     = "AZURE_SUBSCRIPTION_ID"
	DigitalOceanTokenKeyMC    = "DO_TOKEN"
	GoogleServiceAccountKeyMC = "GOOGLE_SERVICE_ACCOUNT"
	HetznerTokenKeyMC         = "HZ_TOKEN"
	OpenStackUserNameMC       = "OS_USER_NAME"
	PacketAPIKeyMC            = "PACKET_API_KEY" //nolint:gosec
	VSphereAddressMC          = "VSPHERE_ADDRESS"
	VSphereUsernameMC         = "VSPHERE_USERNAME"
)

var (
	allKeys = []string{
		AWSAccessKeyID,
		AWSSecretAccessKey,
		AzureClientID,
		AzureClientSecret,
		AzureTenantID,
		AzureSubscribtionID,
		DigitalOceanTokenKey,
		GoogleServiceAccountKey,
		HetznerTokenKey,
		OpenStackAuthURL,
		OpenStackDomainName,
		OpenStackPassword,
		OpenStackRegionName,
		OpenStackTenantID,
		OpenStackTenantName,
		OpenStackUserName,
		PacketAPIKey,
		PacketProjectID,
		VSphereAddress,
		VSpherePassword,
		VSphereUsername,
	}
)

// ProviderEnvironmentVariable is used to match environment variable used by KubeOne to environment variable used by
// machine-controller.
type ProviderEnvironmentVariable struct {
	Name                  string
	MachineControllerName string
}

func Any(credentialsFilePath string) (map[string]string, error) {
	credentialsFinder, err := newCredsFinder(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	creds := map[string]string{}

	for _, key := range allKeys {
		if val := credentialsFinder(key); val != "" {
			creds[key] = val
		}
	}

	return creds, nil
}

// ProviderCredentials implements fetching credentials for each supported provider
func ProviderCredentials(cloudProvider kubeone.CloudProviderSpec, credentialsFilePath string) (map[string]string, error) {
	credentialsFinder, err := newCredsFinder(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	switch {
	case cloudProvider.AWS != nil:
		return credentialsFinder.aws()
	case cloudProvider.Azure != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: AzureClientID, MachineControllerName: AzureClientIDMC},
			{Name: AzureClientSecret, MachineControllerName: AzureClientSecretMC},
			{Name: AzureTenantID, MachineControllerName: AzureTenantIDMC},
			{Name: AzureSubscribtionID, MachineControllerName: AzureSubscribtionIDMC},
		}, defaultValidationFunc)
	case cloudProvider.DigitalOcean != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: DigitalOceanTokenKey, MachineControllerName: DigitalOceanTokenKeyMC},
		}, defaultValidationFunc)
	case cloudProvider.GCE != nil:
		gsa, err := credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: GoogleServiceAccountKey, MachineControllerName: GoogleServiceAccountKeyMC},
		}, defaultValidationFunc)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		// encode it before sending to secret to be consumed by
		// machine-controller, as machine-controller assumes it will be double encoded
		gsa[GoogleServiceAccountKeyMC] = base64.StdEncoding.EncodeToString([]byte(gsa[GoogleServiceAccountKeyMC]))
		return gsa, nil
	case cloudProvider.Hetzner != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: HetznerTokenKey, MachineControllerName: HetznerTokenKeyMC},
		}, defaultValidationFunc)
	case cloudProvider.Openstack != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: OpenStackAuthURL},
			{Name: OpenStackUserName, MachineControllerName: OpenStackUserNameMC},
			{Name: OpenStackPassword},
			{Name: OpenStackDomainName},
			{Name: OpenStackRegionName},
			{Name: OpenStackTenantID},
			{Name: OpenStackTenantName},
		}, openstackValidationFunc)
	case cloudProvider.Packet != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: PacketAPIKey, MachineControllerName: PacketAPIKeyMC},
			{Name: PacketProjectID},
		}, defaultValidationFunc)
	case cloudProvider.Vsphere != nil:
		vscreds, err := credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: VSphereAddress, MachineControllerName: VSphereAddressMC},
			{Name: VSphereUsername, MachineControllerName: VSphereUsernameMC},
			{Name: VSpherePassword},
		}, defaultValidationFunc)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		// force scheme, as machine-controller requires it while terraform does not
		vscreds[VSphereAddressMC] = "https://" + vscreds[VSphereAddressMC]
		return vscreds, nil
	case cloudProvider.None != nil:
		return map[string]string{}, nil
	}

	return nil, errors.New("no provider matched")
}

func newCredsFinder(credentialsFilePath string) (lookupFunc, error) {
	staticMap := map[string]string{}
	finder := func(name string) string {
		if val := os.Getenv(name); val != "" {
			return val
		}
		return staticMap[name]
	}

	if credentialsFilePath == "" {
		return finder, nil
	}

	buf, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load credentials file")
	}

	if err = yaml.Unmarshal(buf, &staticMap); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal credentials file")
	}

	return finder, nil
}

// lookupFunc is function that retrieves credentials from the sources
type lookupFunc func(string) string

func (lookup lookupFunc) aws() (map[string]string, error) {
	creds := make(map[string]string)
	accessKeyID := lookup(AWSAccessKeyID)
	secretAccessKey := lookup(AWSSecretAccessKey)

	if accessKeyID != "" && secretAccessKey != "" {
		creds[AWSAccessKeyID] = accessKeyID
		creds[AWSSecretAccessKey] = secretAccessKey
		return creds, nil
	}

	if os.Getenv("AWS_PROFILE") == "" {
		// no profile is specified, we refuse to totally implicitly use shared
		// credentials. This is needed as a precaution, to avoid accidental
		// exposure of credentials not meant for sharing with cluster.
		return nil, errors.New("no ENV credentials found, AWS_PROFILE is empty")
	}

	// If env fails resort to config file
	sharedCredsProvider := awscredentials.NewSharedCredentials("", "")

	// will error out in case when ether ID or KEY are missing from shared file
	configCreds, err := sharedCredsProvider.Get()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// safe to assume credentials were found
	creds[AWSAccessKeyID] = configCreds.AccessKeyID
	creds[AWSSecretAccessKey] = configCreds.SecretAccessKey

	return creds, nil
}

func (lookup lookupFunc) parseCredentialVariables(envVars []ProviderEnvironmentVariable, validationFunc func(map[string]string) error) (map[string]string, error) {
	creds := make(map[string]string)
	for _, env := range envVars {
		creds[env.Name] = strings.TrimSpace(lookup(env.Name))
	}

	// Validate credentials using given validation function
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
