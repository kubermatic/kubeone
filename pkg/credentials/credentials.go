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

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
)

// Type is a type of credentials that should be fetched
type Type string

const (
	TypeUniversal Type = ""
	TypeCCM       Type = "CCM"
	TypeMC        Type = "MC"
	TypeOSM       Type = "OSM"
)

// The environment variable names with credential in them
const (
	// Variables that KubeOne (and Terraform) expect to see
	AWSAccessKeyID                       = "AWS_ACCESS_KEY_ID"
	AWSSecretAccessKey                   = "AWS_SECRET_ACCESS_KEY" //nolint:gosec
	AzureClientID                        = "ARM_CLIENT_ID"
	AzureClientSecret                    = "ARM_CLIENT_SECRET" //nolint:gosec
	AzureTenantID                        = "ARM_TENANT_ID"
	AzureSubscriptionID                  = "ARM_SUBSCRIPTION_ID"
	DigitalOceanTokenKey                 = "DIGITALOCEAN_TOKEN"
	GoogleServiceAccountKey              = "GOOGLE_CREDENTIALS"
	HetznerTokenKey                      = "HCLOUD_TOKEN"
	KubevirtKubeconfigKey                = "KUBEVIRT_KUBECONFIG"
	NutanixEndpoint                      = "NUTANIX_ENDPOINT"
	NutanixPort                          = "NUTANIX_PORT"
	NutanixUsername                      = "NUTANIX_USERNAME"
	NutanixPassword                      = "NUTANIX_PASSWORD"
	NutanixInsecure                      = "NUTANIX_INSECURE"
	NutanixProxyURL                      = "NUTANIX_PROXY_URL"
	NutanixClusterName                   = "NUTANIX_CLUSTER_NAME"
	NutanixPEEndpoint                    = "NUTANIX_PE_ENDPOINT"
	NutanixPEUsername                    = "NUTANIX_PE_USERNAME"
	NutanixPEPassword                    = "NUTANIX_PE_PASSWORD" //nolint:gosec
	OpenStackAuthURL                     = "OS_AUTH_URL"
	OpenStackDomainName                  = "OS_DOMAIN_NAME"
	OpenStackPassword                    = "OS_PASSWORD"
	OpenStackRegionName                  = "OS_REGION_NAME"
	OpenStackTenantID                    = "OS_TENANT_ID"
	OpenStackTenantName                  = "OS_TENANT_NAME"
	OpenStackUserName                    = "OS_USERNAME"
	OpenStackApplicationCredentialID     = "OS_APPLICATION_CREDENTIAL_ID"
	OpenStackApplicationCredentialSecret = "OS_APPLICATION_CREDENTIAL_SECRET"
	EquinixMetalAuthToken                = "METAL_AUTH_TOKEN" //nolint:gosec
	EquinixMetalProjectID                = "METAL_PROJECT_ID"
	// TODO: Remove Packet env vars after deprecation period.
	PacketAPIKey    = "PACKET_API_KEY"    //nolint:gosec
	PacketProjectID = "PACKET_PROJECT_ID" //nolint:gosec
	VSphereAddress  = "VSPHERE_SERVER"
	VSpherePassword = "VSPHERE_PASSWORD"
	VSphereUsername = "VSPHERE_USER"
	// VMware Cloud Director Credentials
	VMwareCloudDirectorUsername     = "VCD_USER"
	VMwareCloudDirectorPassword     = "VCD_PASSWORD"
	VMwareCloudDirectorAPIToken     = "VCD_API_TOKEN" //nolint:gosec
	VMwareCloudDirectorOrganization = "VCD_ORG"
	VMwareCloudDirectorURL          = "VCD_URL"
	VMwareCloudDirectorVDC          = "VCD_VDC"
	VMwareCloudDirectorSkipTLS      = "VCD_ALLOW_UNVERIFIED_SSL"

	// Variables that machine-controller expects
	AzureClientIDMC           = "AZURE_CLIENT_ID"
	AzureClientSecretMC       = "AZURE_CLIENT_SECRET" //nolint:gosec
	AzureTenantIDMC           = "AZURE_TENANT_ID"
	AzureSubscriptionIDMC     = "AZURE_SUBSCRIPTION_ID"
	DigitalOceanTokenKeyMC    = "DO_TOKEN"
	GoogleServiceAccountKeyMC = "GOOGLE_SERVICE_ACCOUNT"
	HetznerTokenKeyMC         = "HZ_TOKEN"
	OpenStackUserNameMC       = "OS_USER_NAME"
	VSphereAddressMC          = "VSPHERE_ADDRESS"
	VSphereUsernameMC         = "VSPHERE_USERNAME"
)

var allKeys = []string{
	AWSAccessKeyID,
	AWSSecretAccessKey,
	AzureClientID,
	AzureClientSecret,
	AzureTenantID,
	AzureSubscriptionID,
	DigitalOceanTokenKey,
	GoogleServiceAccountKey,
	HetznerTokenKey,
	NutanixEndpoint,
	NutanixPort,
	NutanixUsername,
	NutanixPassword,
	NutanixInsecure,
	NutanixProxyURL,
	NutanixClusterName,
	NutanixPEEndpoint,
	NutanixPEUsername,
	NutanixPEPassword,
	OpenStackAuthURL,
	OpenStackDomainName,
	OpenStackPassword,
	OpenStackRegionName,
	OpenStackTenantID,
	OpenStackTenantName,
	OpenStackUserName,
	EquinixMetalAuthToken,
	EquinixMetalProjectID,
	PacketAPIKey,
	PacketProjectID,
	VSphereAddress,
	VSpherePassword,
	VSphereUsername,
	VMwareCloudDirectorUsername,
	VMwareCloudDirectorPassword,
	VMwareCloudDirectorAPIToken,
	VMwareCloudDirectorOrganization,
	VMwareCloudDirectorURL,
	VMwareCloudDirectorVDC,
	VMwareCloudDirectorSkipTLS,
}

// ProviderEnvironmentVariable is used to match environment variable used by KubeOne to environment variable used by
// machine-controller.
type ProviderEnvironmentVariable struct {
	Name                  string
	MachineControllerName string
}

func Any(credentialsFilePath string) (map[string]string, error) {
	credentialsFinder, err := newCredentialsFinder(withYAMLFile(credentialsFilePath))
	if err != nil {
		return nil, err
	}

	creds := map[string]string{}

	for _, key := range allKeys {
		if val := credentialsFinder.get(key); val != "" {
			creds[key] = val
			// NB: We want to use Equinix Metal env vars everywhere, even if
			// users has PACKET_ env vars on their systems.
			// TODO: Remove after deprecation period.
			switch key {
			case PacketAPIKey:
				creds[EquinixMetalAuthToken] = val
			case PacketProjectID:
				creds[EquinixMetalProjectID] = val
			}
		}
	}

	return creds, nil
}

// ProviderCredentials implements fetching credentials for each supported provider
func ProviderCredentials(cloudProvider kubeoneapi.CloudProviderSpec, credentialsFilePath string, credentialsType Type) (map[string]string, error) {
	credentialsFinderStore, err := newCredentialsFinder(withYAMLFile(credentialsFilePath), withType(credentialsType))
	if err != nil {
		return nil, err
	}

	credentialsFinder := credentialsFinderStore.lookupFunc()
	switch {
	case cloudProvider.AWS != nil:
		return credentialsFinder.aws()
	case cloudProvider.Azure != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: AzureClientID, MachineControllerName: AzureClientIDMC},
			{Name: AzureClientSecret, MachineControllerName: AzureClientSecretMC},
			{Name: AzureTenantID, MachineControllerName: AzureTenantIDMC},
			{Name: AzureSubscriptionID, MachineControllerName: AzureSubscriptionIDMC},
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
			return nil, err
		}

		if credentialsType == TypeMC || credentialsType == TypeOSM {
			// encode it before sending to secret to be consumed by
			// machine-controller, as machine-controller assumes it will be double encoded
			gsa[GoogleServiceAccountKeyMC] = base64.StdEncoding.EncodeToString([]byte(gsa[GoogleServiceAccountKeyMC]))
		}

		return gsa, nil
	case cloudProvider.Hetzner != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: HetznerTokenKey, MachineControllerName: HetznerTokenKeyMC},
		}, defaultValidationFunc)
	case cloudProvider.Kubevirt != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: KubevirtKubeconfigKey},
		}, defaultValidationFunc)
	case cloudProvider.Nutanix != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: NutanixEndpoint},
			{Name: NutanixPort},
			{Name: NutanixUsername},
			{Name: NutanixPassword},
			{Name: NutanixInsecure},
			{Name: NutanixProxyURL},
			{Name: NutanixClusterName},
			{Name: NutanixPEEndpoint},
			{Name: NutanixPEUsername},
			{Name: NutanixPEPassword},
		}, nutanixValidationFunc)
	case cloudProvider.Openstack != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: OpenStackAuthURL},
			{Name: OpenStackUserName, MachineControllerName: OpenStackUserNameMC},
			{Name: OpenStackPassword},
			{Name: OpenStackApplicationCredentialID},
			{Name: OpenStackApplicationCredentialSecret},
			{Name: OpenStackDomainName},
			{Name: OpenStackRegionName},
			{Name: OpenStackTenantID},
			{Name: OpenStackTenantName},
		}, openstackValidationFunc)
	case cloudProvider.EquinixMetal != nil:
		return credentialsFinder.equinixmetal()
	case cloudProvider.VMwareCloudDirector != nil:
		return credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: VMwareCloudDirectorUsername},
			{Name: VMwareCloudDirectorPassword},
			{Name: VMwareCloudDirectorAPIToken},
			{Name: VMwareCloudDirectorOrganization},
			{Name: VMwareCloudDirectorURL},
			{Name: VMwareCloudDirectorVDC},
			{Name: VMwareCloudDirectorSkipTLS},
		}, vmwareCloudDirectorValidationFunc)
	case cloudProvider.Vsphere != nil:
		vscreds, err := credentialsFinder.parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: VSphereAddress, MachineControllerName: VSphereAddressMC},
			{Name: VSphereUsername, MachineControllerName: VSphereUsernameMC},
			{Name: VSpherePassword},
		}, defaultValidationFunc)
		if err != nil {
			return nil, err
		}
		// force scheme, as machine-controller requires it while terraform does not
		vscreds[VSphereAddressMC] = "https://" + vscreds[VSphereAddressMC]

		return vscreds, nil
	case cloudProvider.None != nil:
		return map[string]string{}, nil
	}

	return nil, fail.CredentialsError{
		Op:       "lookup",
		Provider: "unknown",
		Err:      errors.New("unknown provider"),
	}
}

func withYAMLFile(filePath string) func(*credentialsFinder) error {
	return func(cf *credentialsFinder) error {
		if filePath == "" {
			return nil
		}

		buf, err := os.ReadFile(filePath)
		if err != nil {
			return fail.Runtime(err, "reading credentials file")
		}

		if err = yaml.Unmarshal(buf, &cf.static); err != nil {
			return fail.Runtime(err, "unmarshalling credentials file")
		}

		return nil
	}
}

func withType(typ Type) func(*credentialsFinder) error {
	return func(cf *credentialsFinder) error {
		cf.typ = typ

		return nil
	}
}

func newCredentialsFinder(opts ...func(*credentialsFinder) error) (*credentialsFinder, error) {
	cf := credentialsFinder{
		static:  map[string]string{},
		dynamic: os.Getenv,
	}

	for _, optFn := range opts {
		if err := optFn(&cf); err != nil {
			return nil, err
		}
	}

	return &cf, nil
}

type credentialsFinder struct {
	static  map[string]string
	dynamic func(string) string
	typ     Type
}

func (cf *credentialsFinder) lookupFunc() lookupFunc { return cf.get }

func (cf *credentialsFinder) typedKey(name string) string {
	return string(cf.typ) + "_" + name
}

func (cf *credentialsFinder) fetch(name string) string {
	if val := cf.static[name]; val != "" {
		return val
	}

	return cf.dynamic(name)
}

func (cf *credentialsFinder) get(name string) string {
	if cf.typ != TypeUniversal {
		if val := cf.fetch(cf.typedKey(name)); val != "" {
			return val
		}
	}

	return cf.fetch(name)
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
		return nil, fail.CredentialsError{
			Op:       "lookup",
			Provider: "AWS",
			Err:      errors.New("no ENV credentials found, AWS_PROFILE is empty"),
		}
	}

	// If env fails resort to config file
	sharedCredsProvider := awscredentials.NewSharedCredentials("", "")

	// will error out in case when ether ID or KEY are missing from shared file
	configCreds, err := sharedCredsProvider.Get()
	if err != nil {
		return nil, fail.CredentialsError{
			Op:       "lookup",
			Provider: "AWS",
			Err:      errors.WithStack(err),
		}
	}

	// safe to assume credentials were found
	creds[AWSAccessKeyID] = configCreds.AccessKeyID
	creds[AWSSecretAccessKey] = configCreds.SecretAccessKey

	return creds, nil
}

func (lookup lookupFunc) equinixmetal() (map[string]string, error) {
	creds := make(map[string]string)
	packetAPIKey := lookup(PacketAPIKey)
	packetProjectID := lookup(PacketProjectID)
	metalAuthToken := lookup(EquinixMetalAuthToken)
	metalProjectID := lookup(EquinixMetalProjectID)

	if packetAPIKey != "" && packetProjectID != "" && metalAuthToken != "" && metalProjectID != "" {
		return nil, fail.CredentialsError{
			Op:       "lookup",
			Provider: "Equinixmetal",
			Err:      errors.New("found both PACKET_ and METAL_ environment variables, but only one can be used"),
		}
	}

	if (packetAPIKey != "" && packetProjectID == "") || (packetAPIKey == "" && packetProjectID != "") {
		return nil, fail.CredentialsError{
			Op:       "lookup",
			Provider: "Equinixmetal",
			Err:      errors.New("both PACKET_API_KEY and PACKET_PROJECT_ID environment variables are required, but found only one"),
		}
	}

	if (metalAuthToken != "" && metalProjectID == "") || (metalAuthToken == "" && metalProjectID != "") {
		return nil, fail.CredentialsError{
			Op:       "lookup",
			Provider: "Equinixmetal",
			Err:      errors.New("both METAL_AUTH_TOKEN and METAL_PROJECT_ID environment variables are required, but found only one"),
		}
	}

	if packetAPIKey == "" && packetProjectID == "" && metalAuthToken == "" && metalProjectID == "" {
		return nil, fail.CredentialsError{
			Op:       "lookup",
			Provider: "Equinixmetal",
			Err:      errors.New("METAL_AUTH_TOKEN and METAL_PROJECT_ID environment variables are required"),
		}
	}

	if packetAPIKey != "" && packetProjectID != "" {
		creds[EquinixMetalAuthToken] = packetAPIKey
		creds[EquinixMetalProjectID] = packetProjectID

		return creds, nil
	}

	creds[EquinixMetalAuthToken] = metalAuthToken
	creds[EquinixMetalProjectID] = metalProjectID

	return creds, nil
}

func (lookup lookupFunc) parseCredentialVariables(envVars []ProviderEnvironmentVariable, validationFunc func(map[string]string) error) (map[string]string, error) {
	creds := make(map[string]string)
	for _, env := range envVars {
		creds[env.Name] = strings.TrimSpace(lookup(env.Name))
	}

	// Validate credentials using given validation function
	if err := validationFunc(creds); err != nil {
		return nil, err
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
			return fail.CredentialsError{
				Op:  "validating",
				Err: errors.Errorf("key %v is required but isn't present", k),
			}
		}
	}

	return nil
}

func nutanixValidationFunc(creds map[string]string) error {
	alwaysRequired := []string{
		NutanixEndpoint,
		NutanixPort,
		NutanixUsername,
		NutanixPassword,
		NutanixPEEndpoint,
		NutanixPEUsername,
		NutanixPEPassword,
	}

	for _, key := range alwaysRequired {
		if v, ok := creds[key]; !ok || len(v) == 0 {
			return fail.CredentialsError{
				Op:       "validating",
				Provider: "Nutanix",
				Err:      errors.Errorf("key %v is required but is not present", key),
			}
		}
	}

	return nil
}

func openstackValidationFunc(creds map[string]string) error {
	alwaysRequired := []string{OpenStackAuthURL, OpenStackRegionName}

	var (
		appCredsIDOkay        bool
		appCredsSecretOkay    bool
		userCredsUsernameOkay bool
		userCredsPasswordOkay bool
	)

	if v, ok := creds[OpenStackApplicationCredentialID]; ok && len(v) != 0 {
		appCredsIDOkay = true
	}
	if v, ok := creds[OpenStackApplicationCredentialSecret]; ok && len(v) != 0 {
		appCredsSecretOkay = true
	}

	// Domain name is only required when using default credentials i.e. username and password
	if !appCredsIDOkay && !appCredsSecretOkay {
		alwaysRequired = append(alwaysRequired, OpenStackDomainName)
	}

	for _, key := range alwaysRequired {
		if v, ok := creds[key]; !ok || len(v) == 0 {
			return fail.CredentialsError{
				Op:       "validating",
				Provider: "Openstack",
				Err:      errors.Errorf("key %v is required but is not present", key),
			}
		}
	}

	if v, ok := creds[OpenStackUserName]; ok && len(v) != 0 {
		userCredsUsernameOkay = true
	}

	if v, ok := creds[OpenStackPassword]; ok && len(v) != 0 {
		userCredsPasswordOkay = true
	}

	if (appCredsIDOkay || appCredsSecretOkay) && (userCredsUsernameOkay || userCredsPasswordOkay) {
		return fail.CredentialsError{
			Op:       "validating",
			Provider: "Openstack",
			Err: errors.Errorf(
				"both app credentials (%s %s) and user credentials (%s %s) found",
				OpenStackApplicationCredentialID,
				OpenStackApplicationCredentialSecret,
				OpenStackUserName,
				OpenStackPassword,
			),
		}
	}

	if (appCredsIDOkay && !appCredsSecretOkay) || (!appCredsIDOkay && appCredsSecretOkay) {
		return fail.CredentialsError{
			Op:       "validating",
			Provider: "Openstack",
			Err: errors.Errorf(
				"only one of %s, %s is set for application credentials",
				OpenStackApplicationCredentialID,
				OpenStackApplicationCredentialSecret,
			),
		}
	}

	if (userCredsUsernameOkay && !userCredsPasswordOkay) || (!userCredsUsernameOkay && userCredsPasswordOkay) {
		return fail.CredentialsError{
			Op:       "validating",
			Provider: "Openstack",
			Err: errors.Errorf(
				"only one of %s, %s is set for user credentials",
				OpenStackUserName,
				OpenStackPassword,
			),
		}
	}

	if (!appCredsIDOkay && !appCredsSecretOkay) && (!userCredsUsernameOkay && !userCredsPasswordOkay) {
		return fail.CredentialsError{
			Op:       "validating",
			Provider: "Openstack",
			Err:      errors.New("no valid credentials (either application or user) found"),
		}
	}

	// Tenant ID/Name are not required when using application credentials
	if userCredsUsernameOkay && userCredsPasswordOkay {
		if v, ok := creds[OpenStackTenantID]; !ok || len(v) == 0 {
			if v, ok := creds[OpenStackTenantName]; !ok || len(v) == 0 {
				return fail.CredentialsError{
					Op:       "validating",
					Provider: "Openstack",
					Err: errors.Errorf(
						"key %v or %v is required but isn't present",
						OpenStackTenantID,
						OpenStackTenantName,
					),
				}
			}
		}
	}

	return nil
}

func vmwareCloudDirectorValidationFunc(creds map[string]string) error {
	alwaysRequired := []string{
		VMwareCloudDirectorOrganization,
		VMwareCloudDirectorURL,
		VMwareCloudDirectorVDC,
	}

	for _, key := range alwaysRequired {
		if v, ok := creds[key]; !ok || len(v) == 0 {
			return fail.CredentialsError{
				Op:       "validating",
				Provider: "VMware Cloud Director",
				Err:      errors.Errorf("key %v is required but is not present", key),
			}
		}
	}

	var (
		apiToken bool
		password bool
		username bool
	)

	if v, ok := creds[VMwareCloudDirectorAPIToken]; ok && len(v) != 0 {
		apiToken = true
	}

	if v, ok := creds[VMwareCloudDirectorUsername]; ok && len(v) != 0 {
		username = true
	}

	if v, ok := creds[VMwareCloudDirectorPassword]; ok && len(v) != 0 {
		password = true
	}

	userCreds := username || password

	// API token and user credentials are mutually exclusive.
	if userCreds && apiToken {
		return fail.CredentialsError{
			Op:       "validating",
			Provider: "VMware Cloud Director",
			Err: errors.Errorf(
				"both api token (%s) and user credentials (%s %s) found",
				VMwareCloudDirectorAPIToken,
				VMwareCloudDirectorUsername,
				VMwareCloudDirectorPassword,
			),
		}
	}

	// No valid credentials found.
	if !userCreds && !apiToken {
		return fail.CredentialsError{
			Op:       "validating",
			Provider: "VMware Cloud Director",
			Err:      errors.New("no valid credentials (either api token or user) found"),
		}
	}

	// Username and password are required when using default credentials i.e. username and password.
	if !apiToken && (!username || !password) {
		return fail.CredentialsError{
			Op:       "validating",
			Provider: "VMware Cloud Director",
			Err: errors.Errorf(
				"key %v and %v are required but not present",
				VMwareCloudDirectorUsername,
				VMwareCloudDirectorPassword,
			),
		}
	}

	return nil
}
