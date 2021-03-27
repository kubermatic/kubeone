/*
Copyright 2020 The KubeOne Authors.

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

func activateEncryptionProviders(feature *kubeoneapi.EncryptionProviders, args *kubeadmargs.Args) {
	if feature == nil || !feature.Enable {
		return
	}
	if feature.CustomEncryptionConfiguration == "" {
		args.APIServer.ExtraArgs[apiServerEncryptionProviderFlag] = apiServerEncryptionProviderConfigPath
	} else {
		args.APIServer.ExtraArgs[apiServerEncryptionProviderFlag] = apiServerEncryptionProviderCustomConfigPath
	}
}
