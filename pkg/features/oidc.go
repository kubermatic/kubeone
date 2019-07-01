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
	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/kubeadmargs"
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
	optionalMapSet(args.APIServer.ExtraArgs, "oidc-groups-prefix", feature.Config.GroupsPrefix)
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
