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
	"fmt"
	"strings"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	apiserverapi "k8s.io/apiserver/pkg/apis/apiserver"
	apiserverinstall "k8s.io/apiserver/pkg/apis/apiserver/install"
	apiservervalidation "k8s.io/apiserver/pkg/apis/apiserver/validation"
	authenticationcel "k8s.io/apiserver/pkg/authentication/cel"
)

const (
	apiServerAuthenticationConfigFlag = "authentication-config"
	apiServerAnonymousAuthFlag        = "anonymous-auth"
	apiServerServiceAccountIssuerFlag = "service-account-issuer"

	// AuthenticationConfigDir is the directory on control plane nodes holding the AuthenticationConfiguration
	AuthenticationConfigDir = "/etc/kubernetes/authentication"

	// AuthenticationConfigPath is the path on control plane nodes to the AuthenticationConfiguration
	AuthenticationConfigPath = AuthenticationConfigDir + "/config.yaml"
)

var (
	authenticationConfigScheme = runtime.NewScheme()
	authenticationConfigCodecs = serializer.NewCodecFactory(authenticationConfigScheme, serializer.EnableStrict)
)

func init() {
	apiserverinstall.Install(authenticationConfigScheme)
}

// AuthenticationConfigurationEnabled returns true if the AuthenticationConfiguration feature is enabled
func AuthenticationConfigurationEnabled(f kubeoneapi.Features) bool {
	return f.AuthenticationConfiguration != nil && f.AuthenticationConfiguration.Enable
}

func activateKubeadmAuthenticationConfiguration(feature *kubeoneapi.AuthenticationConfiguration, args *kubeadmargs.Args) {
	if feature == nil || !feature.Enable {
		return
	}

	args.APIServer.ExtraArgs[apiServerAuthenticationConfigFlag] = AuthenticationConfigPath
}

// ValidateAuthenticationConfiguration decodes the AuthenticationConfiguration manifest and validates it the same way
// the kube-apiserver does, so that an invalid configuration is caught before it's rolled out to the control plane.
func ValidateAuthenticationConfiguration(data []byte, cluster *kubeoneapi.KubeOneCluster) error {
	obj, err := runtime.Decode(authenticationConfigCodecs.UniversalDecoder(), data)
	if err != nil {
		return fail.ConfigValidation(fmt.Errorf("decoding authentication configuration: %w", err))
	}

	config, ok := obj.(*apiserverapi.AuthenticationConfiguration)
	if !ok {
		return fail.ConfigValidation(fmt.Errorf("expected AuthenticationConfiguration, got %T", obj))
	}

	var apiServerFlags map[string]string
	if cluster.ControlPlaneComponents != nil && cluster.ControlPlaneComponents.APIServer != nil {
		apiServerFlags = cluster.ControlPlaneComponents.APIServer.Flags
	}

	// kube-apiserver refuses to start when anonymous authentication is configured both ways
	if _, ok := apiServerFlags[apiServerAnonymousAuthFlag]; ok && config.Anonymous != nil {
		return fail.ConfigValidation(fmt.Errorf("authentication configuration with anonymous section can't be used together with the %q kube-apiserver flag", apiServerAnonymousAuthFlag))
	}

	// JWT issuers must not overlap with the service account token issuers
	disallowedIssuers := []string{fmt.Sprintf("https://kubernetes.default.svc.%s", cluster.ClusterNetwork.ServiceDomainName)}
	if issuers, ok := apiServerFlags[apiServerServiceAccountIssuerFlag]; ok {
		disallowedIssuers = strings.Split(issuers, ",")
	}

	if errs := apiservervalidation.ValidateAuthenticationConfiguration(authenticationcel.NewDefaultCompiler(), config, disallowedIssuers); len(errs) > 0 {
		return fail.ConfigValidation(fmt.Errorf("invalid authentication configuration: %w", errs.ToAggregate()))
	}

	return nil
}
