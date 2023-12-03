/*
Copyright 2021 The KubeOne Authors.

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
	"errors"
	"testing"

	"k8c.io/kubeone/pkg/fail"
)

func TestOpenstackValidationFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		creds map[string]string
		err   error
	}{
		{
			name:  "no-credentials",
			creds: map[string]string{},
			err:   errors.New("key OS_AUTH_URL is required but is not present"),
		},
		{
			name: "application-credentials",
			creds: map[string]string{
				"OS_TENANT_NAME":                   "test",
				"OS_AUTH_URL":                      "https://localhost:5000",
				"OS_DOMAIN_NAME":                   "test",
				"OS_REGION_NAME":                   "de",
				"OS_APPLICATION_CREDENTIAL_ID":     "1234",
				"OS_APPLICATION_CREDENTIAL_SECRET": "5678",
			},
			err: nil,
		},
		{
			name: "no-credential-secret",
			creds: map[string]string{
				"OS_TENANT_NAME":               "test",
				"OS_AUTH_URL":                  "https://localhost:5000",
				"OS_DOMAIN_NAME":               "test",
				"OS_REGION_NAME":               "de",
				"OS_APPLICATION_CREDENTIAL_ID": "1234",
			},
			err: errors.New("only one of OS_APPLICATION_CREDENTIAL_ID, OS_APPLICATION_CREDENTIAL_SECRET is set for application credentials"),
		},
		{
			name: "no-credential-id",
			creds: map[string]string{
				"OS_TENANT_NAME":                   "test",
				"OS_AUTH_URL":                      "https://localhost:5000",
				"OS_DOMAIN_NAME":                   "test",
				"OS_REGION_NAME":                   "de",
				"OS_APPLICATION_CREDENTIAL_SECRET": "5678",
			},
			err: errors.New("only one of OS_APPLICATION_CREDENTIAL_ID, OS_APPLICATION_CREDENTIAL_SECRET is set for application credentials"),
		},
		{
			name: "user-credentials",
			creds: map[string]string{
				"OS_TENANT_NAME": "test",
				"OS_AUTH_URL":    "https://localhost:5000",
				"OS_DOMAIN_NAME": "test",
				"OS_REGION_NAME": "de",
				"OS_USERNAME":    "1234",
				"OS_PASSWORD":    "5678",
			},
			err: nil,
		},
		{
			name: "no-password",
			creds: map[string]string{
				"OS_TENANT_NAME": "test",
				"OS_AUTH_URL":    "https://localhost:5000",
				"OS_DOMAIN_NAME": "test",
				"OS_REGION_NAME": "de",
				"OS_USERNAME":    "1234",
			},
			err: errors.New("only one of OS_USERNAME, OS_PASSWORD is set for user credentials"),
		},
		{
			name: "no-username",
			creds: map[string]string{
				"OS_TENANT_NAME": "test",
				"OS_AUTH_URL":    "https://localhost:5000",
				"OS_DOMAIN_NAME": "test",
				"OS_REGION_NAME": "de",
				"OS_PASSWORD":    "5678",
			},
			err: errors.New("only one of OS_USERNAME, OS_PASSWORD is set for user credentials"),
		},
		{
			name: "mixed-credentials-1",
			creds: map[string]string{
				"OS_TENANT_NAME":               "test",
				"OS_AUTH_URL":                  "https://localhost:5000",
				"OS_DOMAIN_NAME":               "test",
				"OS_REGION_NAME":               "de",
				"OS_APPLICATION_CREDENTIAL_ID": "1234",
				"OS_PASSWORD":                  "5678",
			},
			err: errors.New("both app credentials (OS_APPLICATION_CREDENTIAL_ID OS_APPLICATION_CREDENTIAL_SECRET) and user credentials (OS_USERNAME OS_PASSWORD) found"),
		},
		{
			name: "mixed-credentials-2",
			creds: map[string]string{
				"OS_TENANT_NAME":                   "test",
				"OS_AUTH_URL":                      "https://localhost:5000",
				"OS_DOMAIN_NAME":                   "test",
				"OS_REGION_NAME":                   "de",
				"OS_APPLICATION_CREDENTIAL_SECRET": "5678",
				"OS_USERNAME":                      "1234",
			},
			err: errors.New("both app credentials (OS_APPLICATION_CREDENTIAL_ID OS_APPLICATION_CREDENTIAL_SECRET) and user credentials (OS_USERNAME OS_PASSWORD) found"),
		},
		{
			name: "mixed-credentials-3",
			creds: map[string]string{
				"OS_TENANT_NAME":                   "test",
				"OS_AUTH_URL":                      "https://localhost:5000",
				"OS_DOMAIN_NAME":                   "test",
				"OS_REGION_NAME":                   "de",
				"OS_APPLICATION_CREDENTIAL_ID":     "1234",
				"OS_APPLICATION_CREDENTIAL_SECRET": "5678",
				"OS_USERNAME":                      "1234",
				"OS_PASSWORD":                      "5678",
			},
			err: errors.New("both app credentials (OS_APPLICATION_CREDENTIAL_ID OS_APPLICATION_CREDENTIAL_SECRET) and user credentials (OS_USERNAME OS_PASSWORD) found"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := openstackValidationFunc(tt.creds)
			if tt.err != nil && err != nil {
				var credsErr fail.CredentialsError
				if !errors.As(err, &credsErr) {
					t.Errorf("extected %T error type", credsErr)
				}
				if credsErr.Err.Error() != tt.err.Error() {
					t.Errorf("expected error = '%v', got error = '%v'", tt.err.Error(), err.Error())
				}
			} else if !errors.Is(err, tt.err) {
				t.Errorf("%s: expected error = %v, got error = %v", tt.name, tt.err, err)
			}
		})
	}
}

func TestVmwareCloudDirectorValidationFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		creds map[string]string
		err   error
	}{
		{
			name:  "empty",
			creds: map[string]string{},
			err:   errors.New("key VCD_ORG is required but is not present"),
		},
		{
			name: "no-credentials",
			creds: map[string]string{
				"VCD_ORG": "test",
				"VCD_URL": "http://localhost:8080",
				"VCD_VDC": "vdc",
			},
			err: errors.New("no valid credentials (either api token or user) found"),
		},
		{
			name: "username-password",
			creds: map[string]string{
				"VCD_ORG":      "test",
				"VCD_URL":      "http://localhost:8080",
				"VCD_VDC":      "vdc",
				"VCD_USER":     "user",
				"VCD_PASSWORD": "password",
			},
			err: nil,
		},
		{
			name: "username-no-password",
			creds: map[string]string{
				"VCD_ORG":  "test",
				"VCD_URL":  "http://localhost:8080",
				"VCD_VDC":  "vdc",
				"VCD_USER": "user",
			},
			err: errors.New("key VCD_USER and VCD_PASSWORD are required but not present"),
		},
		{
			name: "password-no-username",
			creds: map[string]string{
				"VCD_ORG":      "test",
				"VCD_URL":      "http://localhost:8080",
				"VCD_VDC":      "vdc",
				"VCD_PASSWORD": "password",
			},
			err: errors.New("key VCD_USER and VCD_PASSWORD are required but not present"),
		},
		{
			name: "api-token",
			creds: map[string]string{
				"VCD_ORG":       "test",
				"VCD_URL":       "http://localhost:8080",
				"VCD_VDC":       "vdc",
				"VCD_API_TOKEN": "token",
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := vmwareCloudDirectorValidationFunc(tt.creds)
			if tt.err != nil && err != nil {
				var credsErr fail.CredentialsError
				if !errors.As(err, &credsErr) {
					t.Errorf("extected %T error type", credsErr)
				}
				if credsErr.Err.Error() != tt.err.Error() {
					t.Errorf("expected error = '%v', got error = '%v'", tt.err.Error(), err.Error())
				}
			} else if !errors.Is(err, tt.err) {
				t.Errorf("%s: expected error = %v, got error = %v", tt.name, tt.err, err)
			}
		})
	}
}
