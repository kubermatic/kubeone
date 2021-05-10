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

package images

import (
	"sort"

	"github.com/docker/distribution/reference"
)

type Resource int

func (res Resource) namedReference() reference.Named {
	named, _ := reference.ParseNormalizedNamed(allResources()[res])
	return named
}

const (
	CalicoCNI Resource = iota
	CalicoController
	CalicoNode
	DigitaloceanCCM
	DNSNodeCache
	Flannel
	HetznerCCM
	MachineController
	MetricsServer
	OpenstackCCM
	PacketCCM
	VsphereCCM
	WeaveNetCNIKube
	WeaveNetCNINPC
)

func baseResources() map[Resource]string {
	return map[Resource]string{
		CalicoCNI:         "docker.io/calico/cni:v3.16.5",
		CalicoController:  "docker.io/calico/kube-controllers:v3.16.5",
		CalicoNode:        "docker.io/calico/node:v3.16.5",
		DNSNodeCache:      "k8s.gcr.io/k8s-dns-node-cache:1.15.13",
		Flannel:           "quay.io/coreos/flannel:v0.13.0",
		MachineController: "docker.io/kubermatic/machine-controller:v1.29.0",
		MetricsServer:     "k8s.gcr.io/metrics-server:v0.3.6",
	}
}

func optionalResources() map[Resource]string {
	return map[Resource]string{
		DigitaloceanCCM: "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.23",
		HetznerCCM:      "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.8.1",
		OpenstackCCM:    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.17.0",
		PacketCCM:       "docker.io/packethost/packet-ccm:v1.0.0",
		VsphereCCM:      "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.2.1",
		WeaveNetCNIKube: "docker.io/weaveworks/weave-kube:2.7.0",
		WeaveNetCNINPC:  "docker.io/weaveworks/weave-npc:2.7.0",
	}
}

func allResources() map[Resource]string {
	ret := map[Resource]string{}
	for k, v := range baseResources() {
		ret[k] = v
	}
	for k, v := range optionalResources() {
		ret[k] = v
	}
	return ret
}

type Opt func(*Resolver)

func WithOverwriteRegistryGetter(getter func() string) Opt {
	return func(r *Resolver) {
		r.overwriteRegistryGetter = getter
	}
}

func NewResolver(opts ...Opt) *Resolver {
	r := &Resolver{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type Resolver struct {
	overwriteRegistryGetter func() string
}

type ListFilter int

const (
	ListFilterNone ListFilter = iota
	ListFilterBase
	ListFilterOpional
)

func (r *Resolver) List(lf ListFilter) []string {
	var list []string

	fn := allResources
	switch lf {
	case ListFilterBase:
		fn = baseResources
	case ListFilterOpional:
		fn = optionalResources
	}

	for res := range fn() {
		list = append(list, r.Get(res))
	}

	sort.Strings(list)
	return list
}

func (r *Resolver) Tag(res Resource) string {
	named := res.namedReference()
	if tagged, ok := named.(reference.Tagged); ok {
		return tagged.Tag()
	}

	return "latest"
}

type GetOpt func(ref string) string

func WithDomain(domain string) GetOpt {
	return func(ref string) string {
		named, _ := reference.ParseNormalizedNamed(ref)
		nt := named.(reference.NamedTagged)
		path := reference.Path(named)

		return domain + "/" + path + ":" + nt.Tag()
	}
}

func WithTag(tag string) GetOpt {
	return func(ref string) string {
		named, _ := reference.ParseNormalizedNamed(ref)
		return named.Name() + ":" + tag
	}
}

func (r *Resolver) Get(res Resource, opts ...GetOpt) string {
	named := res.namedReference()
	domain := reference.Domain(named)
	reminder := reference.Path(named)

	if tagged, ok := named.(reference.Tagged); ok {
		reminder += ":" + tagged.Tag()
	} else {
		reminder += ":latest"
	}

	if r.overwriteRegistryGetter != nil {
		if reg := r.overwriteRegistryGetter(); reg != "" {
			domain = reg
		}
	}

	ret := domain + "/" + reminder
	for _, opt := range opts {
		ret = opt(ret)
	}

	return ret
}
