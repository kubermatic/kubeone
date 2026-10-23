/*
Copyright 2026 The KubeOne Authors.

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

package features

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"
)

func TestActivateKubeadmAuthenticationConfiguration(t *testing.T) {
	args := kubeadmargs.New()

	activateKubeadmAuthenticationConfiguration(&kubeoneapi.AuthenticationConfiguration{Enable: true, ConfigFilePath: "auth.yaml"}, args)

	if got := args.APIServer.ExtraArgs[apiServerAuthenticationConfigFlag]; got != AuthenticationConfigPath {
		t.Fatalf("unexpected authentication config flag: got %q, want %q", got, AuthenticationConfigPath)
	}
}

func TestActivateKubeadmAuthenticationConfigurationDisabled(t *testing.T) {
	for _, feature := range []*kubeoneapi.AuthenticationConfiguration{nil, {Enable: false}} {
		args := kubeadmargs.New()

		activateKubeadmAuthenticationConfiguration(feature, args)

		if got, ok := args.APIServer.ExtraArgs[apiServerAuthenticationConfigFlag]; ok {
			t.Fatalf("expected no authentication config flag, got %q", got)
		}
	}
}

func TestValidateAuthenticationConfiguration(t *testing.T) {
	validJWT := heredoc.Doc(`
		apiVersion: apiserver.config.k8s.io/v1
		kind: AuthenticationConfiguration
		jwt:
		- issuer:
		    url: https://issuer.example.com
		    audiences:
		    - kubernetes
		    - ci
		    audienceMatchPolicy: MatchAny
		  claimValidationRules:
		  - expression: 'claims.hd == "example.com"'
		    message: hosted domain must be example.com
		  claimMappings:
		    username:
		      expression: '"oidc:" + claims.sub'
		- issuer:
		    url: https://token.actions.githubusercontent.com
		    audiences:
		    - kubernetes
		  claimMappings:
		    username:
		      claim: sub
		      prefix: "github:"
	`)

	anonymous := heredoc.Doc(`
		apiVersion: apiserver.config.k8s.io/v1
		kind: AuthenticationConfiguration
		anonymous:
		  enabled: true
		  conditions:
		  - path: /livez
		  - path: /readyz
		  - path: /healthz
	`)

	tests := []struct {
		name           string
		config         string
		apiServerFlags map[string]string
		wantErr        bool
	}{
		{
			name:   "multiple JWT issuers",
			config: validJWT,
		},
		{
			name:   "anonymous with conditions",
			config: anonymous,
		},
		{
			name: "v1beta1 is accepted",
			config: heredoc.Doc(`
				apiVersion: apiserver.config.k8s.io/v1beta1
				kind: AuthenticationConfiguration
				anonymous:
				  enabled: false
			`),
		},
		{
			name:           "anonymous with anonymous-auth flag",
			config:         anonymous,
			apiServerFlags: map[string]string{"anonymous-auth": "false"},
			wantErr:        true,
		},
		{
			name:           "anonymous-auth flag without anonymous section",
			config:         validJWT,
			apiServerFlags: map[string]string{"anonymous-auth": "false"},
		},
		{
			name: "anonymous conditions without enabled",
			config: heredoc.Doc(`
				apiVersion: apiserver.config.k8s.io/v1
				kind: AuthenticationConfiguration
				anonymous:
				  enabled: false
				  conditions:
				  - path: /livez
			`),
			wantErr: true,
		},
		{
			name: "invalid CEL expression",
			config: heredoc.Doc(`
				apiVersion: apiserver.config.k8s.io/v1
				kind: AuthenticationConfiguration
				jwt:
				- issuer:
				    url: https://issuer.example.com
				    audiences:
				    - kubernetes
				  claimMappings:
				    username:
				      expression: 'claims.sub +'
			`),
			wantErr: true,
		},
		{
			name: "issuer conflicting with default service account issuer",
			config: heredoc.Doc(`
				apiVersion: apiserver.config.k8s.io/v1
				kind: AuthenticationConfiguration
				jwt:
				- issuer:
				    url: https://kubernetes.default.svc.cluster.local
				    audiences:
				    - kubernetes
				  claimMappings:
				    username:
				      claim: sub
				      prefix: ""
			`),
			wantErr: true,
		},
		{
			name:           "issuer conflicting with custom service account issuer",
			config:         validJWT,
			apiServerFlags: map[string]string{"service-account-issuer": "https://issuer.example.com"},
			wantErr:        true,
		},
		{
			name: "unknown field",
			config: heredoc.Doc(`
				apiVersion: apiserver.config.k8s.io/v1
				kind: AuthenticationConfiguration
				anonymus:
				  enabled: true
			`),
			wantErr: true,
		},
		{
			name: "wrong kind",
			config: heredoc.Doc(`
				apiVersion: apiserver.config.k8s.io/v1
				kind: EncryptionConfiguration
				resources: []
			`),
			wantErr: true,
		},
		{
			name:    "not yaml",
			config:  "{",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &kubeoneapi.KubeOneCluster{
				ClusterNetwork: kubeoneapi.ClusterNetworkConfig{
					ServiceDomainName: "cluster.local",
				},
				ControlPlaneComponents: &kubeoneapi.ControlPlaneComponents{
					APIServer: &kubeoneapi.ControlPlaneComponentConfig{
						Flags: tt.apiServerFlags,
					},
				},
			}

			err := ValidateAuthenticationConfiguration([]byte(tt.config), cluster)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateAuthenticationConfiguration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
