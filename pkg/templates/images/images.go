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

//go:generate go run golang.org/x/tools/cmd/stringer -type=Resource

package images

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/docker/distribution/reference"
)

type Resource int

func (res Resource) namedReference(kubernetesVersionGetter func() string) reference.Named {
	kubeVer, _ := semver.NewVersion(kubernetesVersionGetter())

	for ver, img := range allResources()[res] {
		sv, _ := semver.NewConstraint(ver)
		if sv.Check(kubeVer) {
			named, _ := reference.ParseNormalizedNamed(img)
			return named
		}
	}

	return nil
}

const (
	// default 0 index has no meaning
	// Canal CNI
	CalicoCNI Resource = iota + 1
	CalicoController
	CalicoNode
	Flannel

	// Cilium CNI
	Cilium
	CiliumOperator
	HubbleRelay
	HubbleUI
	HubbleUIBackend
	HubbleProxy

	// WeaveNet CNI
	WeaveNetCNIKube
	WeaveNetCNINPC

	// Core components (MC, metrics-server...)
	DNSNodeCache
	MachineController
	MetricsServer

	// Addons
	ClusterAutoscaler

	// General CSI images (to be removed)
	CSIAttacher
	CSINodeDriverRegistar
	CSIProvisioner
	CSISnapshotter
	CSIResizer
	CSILivenessProbe

	AwsCCM

	// Azure CCM
	AzureCCM
	AzureCNM

	// AWS EBS CSI
	AwsEbsCSI
	AwsEbsCSIAttacher
	AwsEbsCSILivenessProbe
	AwsEbsCSINodeDriverRegistrar
	AwsEbsCSIProvisioner
	AwsEbsCSIResizer
	AwsEbsCSISnapshotter
	AwsEbsCSISnapshotController

	// AzureFile CSI
	AzureFileCSI
	AzureFileCSIAttacher
	AzureFileCSILivenessProbe
	AzureFileCSINodeDriverRegistar
	AzureFileCSIProvisioner
	AzureFileCSIResizer
	AzureFileCSISnapshotter
	AzureFileCSISnapshotterController

	// AzureDisk CSI
	AzureDiskCSI
	AzureDiskCSIAttacher
	AzureDiskCSILivenessProbe
	AzureDiskCSINodeDriverRegistar
	AzureDiskCSIProvisioner
	AzureDiskCSIResizer
	AzureDiskCSISnapshotter
	AzureDiskCSISnapshotterController

	// CCMs and CSI plugins
	DigitaloceanCCM
	HetznerCCM
	HetznerCSI
	OpenstackCCM
	OpenstackCSI
	EquinixMetalCCM
	VsphereCCM
	VsphereCSIDriver
	VsphereCSISyncer
	VsphereCSIProvisioner
)

func FindResource(name string) (Resource, error) {
	for res := range allResources() {
		if res.String() == name {
			return res, nil
		}
	}

	return 0, fmt.Errorf("no such resource: %q", name)
}

func baseResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		CalicoCNI:         {"*": "docker.io/calico/cni:v3.19.1"},
		CalicoController:  {"*": "docker.io/calico/kube-controllers:v3.19.1"},
		CalicoNode:        {"*": "docker.io/calico/node:v3.19.1"},
		DNSNodeCache:      {"*": "k8s.gcr.io/k8s-dns-node-cache:1.15.13"},
		Flannel:           {"*": "quay.io/coreos/flannel:v0.13.0"},
		MachineController: {"*": "docker.io/kubermatic/machine-controller:v1.40.1"},
		MetricsServer:     {"*": "k8s.gcr.io/metrics-server/metrics-server:v0.5.0"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		// General CSI images (could be used for all providers)
		CSINodeDriverRegistar: {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.3.0"},
		CSILivenessProbe:      {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.4.0"},
		CSIAttacher:           {">= 1.19.0": "k8s.gcr.io/sig-storage/csi-attacher:v3.3.0"},
		CSIProvisioner:        {">= 1.19.0": "k8s.gcr.io/sig-storage/csi-provisioner:v2.2.2"},
		CSIResizer:            {">= 1.19.0": "k8s.gcr.io/sig-storage/csi-resizer:v1.3.0"},
		CSISnapshotter: {
			">= 1.19.0, < 1.20.0": "k8s.gcr.io/sig-storage/csi-snapshotter:v3.0.3",
			">= 1.20.0":           "k8s.gcr.io/sig-storage/csi-snapshotter:v4.2.0",
		},

		AwsCCM: {
			"1.19.x":    "us.gcr.io/k8s-artifacts-prod/provider-aws/cloud-controller-manager:v1.19.0-alpha.1",
			"1.20.x":    "us.gcr.io/k8s-artifacts-prod/provider-aws/cloud-controller-manager:v1.20.0-alpha.0",
			"1.21.x":    "us.gcr.io/k8s-artifacts-prod/provider-aws/cloud-controller-manager:v1.21.0-alpha.0",
			">= 1.22.0": "us.gcr.io/k8s-artifacts-prod/provider-aws/cloud-controller-manager:v1.22.0-alpha.0",
		},

		// Azure CCM
		AzureCCM: {
			"1.19.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v0.6.0",
			"1.20.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v0.7.8",
			"1.21.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.0.5",
			">= 1.22.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.1.1",
		},
		AzureCNM: {
			"1.19.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v0.6.0",
			"1.20.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v0.7.8",
			"1.21.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.0.5",
			">= 1.22.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.1.1",
		},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "k8s.gcr.io/provider-aws/aws-ebs-csi-driver:v1.4.0"},
		AwsEbsCSIAttacher:            {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.1.0"},
		AwsEbsCSILivenessProbe:       {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.2.0"},
		AwsEbsCSINodeDriverRegistrar: {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.1.0"},
		AwsEbsCSIProvisioner:         {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v2.1.1"},
		AwsEbsCSIResizer:             {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.1.0"},
		AwsEbsCSISnapshotter: {
			">= 1.19.0, < 1.20.0": "k8s.gcr.io/sig-storage/csi-snapshotter:v3.0.3",
			">= 1.20.0":           "k8s.gcr.io/sig-storage/csi-snapshotter:v4.2.1",
		},
		AwsEbsCSISnapshotController: {
			">= 1.19.0, < 1.20.0": "k8s.gcr.io/sig-storage/snapshot-controller:v3.0.3",
			">= 1.20.0":           "k8s.gcr.io/sig-storage/snapshot-controller:v4.2.1",
		},

		// AzureFile CSI driver
		AzureFileCSI:                      {"*": "mcr.microsoft.com/k8s/csi/azurefile-csi:v1.7.0"},
		AzureFileCSIAttacher:              {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.3.0"},
		AzureFileCSILivenessProbe:         {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.4.0"},
		AzureFileCSINodeDriverRegistar:    {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.3.0"},
		AzureFileCSIProvisioner:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v2.2.2"},
		AzureFileCSIResizer:               {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.3.0"},
		AzureFileCSISnapshotter:           {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v3.0.3"},
		AzureFileCSISnapshotterController: {"*": "mcr.microsoft.com/oss/kubernetes-csi/snapshot-controller:v3.0.3"},

		// AzureDisk CSI driver
		AzureDiskCSI:                      {"*": "mcr.microsoft.com/k8s/csi/azuredisk-csi:v1.8.0"},
		AzureDiskCSIAttacher:              {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v3.3.0"},
		AzureDiskCSILivenessProbe:         {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.4.0"},
		AzureDiskCSINodeDriverRegistar:    {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.3.0"},
		AzureDiskCSIProvisioner:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v2.2.2"},
		AzureDiskCSIResizer:               {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.3.0"},
		AzureDiskCSISnapshotter:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v3.0.3"},
		AzureDiskCSISnapshotterController: {"*": "mcr.microsoft.com/oss/kubernetes-csi/snapshot-controller:v3.0.3"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.33"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.12.0"},

		// Hetzner CSI
		HetznerCSI: {"*": "docker.io/hetznercloud/hcloud-csi-driver:1.6.0"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.19.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.19.2",
			"1.20.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.20.2",
			"1.21.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.21.0",
			">= 1.22.0": "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.22.0",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.19.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.19.0",
			"1.20.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.20.3",
			"1.21.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.21.0",
			">= 1.22.0": "docker.io/k8scloudprovider/cinder-csi-plugin:v1.22.0",
		},

		// Equinix Metal CCM
		EquinixMetalCCM: {"*": "docker.io/equinix/cloud-provider-equinix-metal:v3.3.0"},

		// vSphere CCM
		VsphereCCM: {
			"1.19.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.19.1",
			"1.20.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.20.0",
			"1.21.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.21.1",
			">= 1.22.0": "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.22.4",
		},

		// vSphere CSI
		VsphereCSIDriver:      {"*": "gcr.io/cloud-provider-vsphere/csi/release/driver:v2.4.0"},
		VsphereCSISyncer:      {"*": "gcr.io/cloud-provider-vsphere/csi/release/syncer:v2.4.0"},
		VsphereCSIProvisioner: {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.0.0"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.11.0"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.11.0"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.11.0"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.8.3"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.8.3"},
		HubbleProxy:     {"*": "docker.io/envoyproxy/envoy:v1.18.4"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.19.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.19.0",
			"1.20.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.20.0",
			"1.21.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.21.0",
			">= 1.22.0": "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.22.0",
		},
	}
}

func allResources() map[Resource]map[string]string {
	ret := map[Resource]map[string]string{}
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

func WithKubernetesVersionGetter(getter func() string) Opt {
	return func(r *Resolver) {
		r.kubernetesVersionGetter = getter
	}
}

func NewResolver(opts ...Opt) *Resolver {
	r := &Resolver{}
	for _, opt := range opts {
		opt(r)
	}

	// If KubernetesVersionGetter is not provided, we'll default to 0.0.0,
	// so that we can at least get images that are version-independent.
	if r.kubernetesVersionGetter == nil {
		r.kubernetesVersionGetter = func() string {
			return "9.9.9"
		}
	}

	return r
}

type Resolver struct {
	overwriteRegistryGetter func() string
	kubernetesVersionGetter func() string
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
		img := r.Get(res)
		if img != "" {
			list = append(list, img)
		}
	}

	sort.Strings(list)
	return list
}

func (r *Resolver) Tag(res Resource) string {
	named := res.namedReference(r.kubernetesVersionGetter)
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
	named := res.namedReference(r.kubernetesVersionGetter)
	if named == nil {
		return ""
	}

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
