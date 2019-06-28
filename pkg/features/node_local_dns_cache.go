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
	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/templates/dnscache"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/kubeadmargs"
	"github.com/kubermatic/kubeone/pkg/util"
)

func installNodeLocalDNSCache(nodelocalcache *kubeone.NodeLocalDNSCache, ctx *util.Context) error {
	if nodelocalcache == nil {
		return nil
	}

	if !nodelocalcache.Enable {
		return nil
	}

	return dnscache.Deploy(ctx)
}

func updateNodeLocalDNSCacheKubeadmConfig(feature *kubeoneapi.NodeLocalDNSCache, args *kubeadmargs.Args) {
	if feature == nil {
		return
	}

	if !feature.Enable {
		return
	}

	args.Kubelet.ExtraArgs["cluster-dns"] = dnscache.VirtualIP
}
