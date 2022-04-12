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

package features

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"
)

func activateKubeadmOIDC(feature *kubeoneapi.OpenIDConnect, args *kubeadmargs.Args) {
	if feature == nil || !feature.Enable {
		return
	}

	args.APIServer.ExtraArgs["oidc-issuer-url"] = feature.Config.IssuerURL
	args.APIServer.ExtraArgs["oidc-client-id"] = feature.Config.ClientID
	optionalMapSet(args.APIServer.ExtraArgs, "oidc-username-claim", feature.Config.UsernameClaim)
	optionalMapSet(args.APIServer.ExtraArgs, "oidc-username-prefix", feature.Config.UsernamePrefix)
	optionalMapSet(args.APIServer.ExtraArgs, "oidc-groups-claim", feature.Config.GroupsClaim)

	// While we have to handle the "-" value for GroupsPrefix, the same is done by the kube-apiserver for the
	// UsernamePrefix. This is why there is no `if feature.Config.UsernamePrefix != "-" { ... }`
	//
	// Docs at the https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/ says:
	// `--oidc-username-prefix string` If provided, all usernames will be prefixed with this value. If not provided,
	// username claims other than 'email' are prefixed by the issuer URL to avoid clashes. To skip any prefixing,
	// provide the value '-'.
	if feature.Config.GroupsPrefix != "-" {
		optionalMapSet(args.APIServer.ExtraArgs, "oidc-groups-prefix", feature.Config.GroupsPrefix)
	}
	optionalMapSet(args.APIServer.ExtraArgs, "oidc-required-claim", feature.Config.RequiredClaim)
	optionalMapSet(args.APIServer.ExtraArgs, "oidc-signing-algs", feature.Config.SigningAlgs)
	optionalMapSet(args.APIServer.ExtraArgs, "oidc-ca-file", feature.Config.CAFile)
}

func optionalMapSet(m map[string]string, key string, val string) {
	if val == "" {
		return
	}

	m[key] = val
}
