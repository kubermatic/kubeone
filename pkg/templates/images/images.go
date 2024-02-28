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
	"github.com/distribution/reference"

	"k8c.io/kubeone/pkg/fail"
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
	CiliumCertGen

	// WeaveNet CNI
	WeaveNetCNIKube
	WeaveNetCNINPC

	// Core components (MC, metrics-server...)
	DNSNodeCache
	MachineController
	MetricsServer
	OperatingSystemManager

	// Addons
	ClusterAutoscaler

	// AWS CCM
	AwsCCM

	// Azure CCM
	AzureCCM
	AzureCNM

	// CSI external snapshotter
	CSISnapshotController
	CSISnapshotWebhook

	// AWS EBS CSI
	AwsEbsCSI
	AwsEbsCSIAttacher
	AwsEbsCSILivenessProbe
	AwsEbsCSINodeDriverRegistrar
	AwsEbsCSIProvisioner
	AwsEbsCSIResizer
	AwsEbsCSISnapshotter

	// AzureFile CSI
	AzureFileCSI
	AzureFileCSIAttacher
	AzureFileCSILivenessProbe
	AzureFileCSINodeDriverRegistar
	AzureFileCSIProvisioner
	AzureFileCSIResizer
	AzureFileCSISnapshotter

	// AzureDisk CSI
	AzureDiskCSI
	AzureDiskCSIAttacher
	AzureDiskCSILivenessProbe
	AzureDiskCSINodeDriverRegistar
	AzureDiskCSIProvisioner
	AzureDiskCSIResizer
	AzureDiskCSISnapshotter

	// Nutanix CSI
	NutanixCSILivenessProbe
	NutanixCSI
	NutanixCSIProvisioner
	NutanixCSIRegistrar
	NutanixCSIResizer
	NutanixCSISnapshotter

	// DigitalOcean CSI
	DigitalOceanCSI
	DigitalOceanCSIAlpine
	DigitalOceanCSIAttacher
	DigitalOceanCSINodeDriverRegistar
	DigitalOceanCSIProvisioner
	DigitalOceanCSIResizer
	DigitalOceanCSISnapshotter

	// OpenStack CSI
	OpenstackCSI
	OpenstackCSINodeDriverRegistar
	OpenstackCSILivenessProbe
	OpenstackCSIAttacher
	OpenstackCSIProvisioner
	OpenstackCSIResizer
	OpenstackCSISnapshotter

	// Hetzner CSI
	HetznerCSI
	HetznerCSIAttacher
	HetznerCSIResizer
	HetznerCSIProvisioner
	HetznerCSILivenessProbe
	HetznerCSINodeDriverRegistar

	// CCMs and CSI plugins
	DigitaloceanCCM
	EquinixMetalCCM
	HetznerCCM
	GCPCCM
	NutanixCCM
	OpenstackCCM
	VsphereCCM

	// CSI Vault Secret Provider
	CSIVaultSecretProvider // hashicorp/vault-csi-provider:1.1.0

	// CSI Secrets Driver
	SecretStoreCSIDriverNodeRegistrar
	SecretStoreCSIDriver
	SecretStoreCSIDriverLivenessProbe
	SecretStoreCSIDriverCRDs

	// VMwareCloud Director CSI
	VMwareCloudDirectorCSI
	VMwareCloudDirectorCSIAttacher
	VMwareCloudDirectorCSIProvisioner
	VMwareCloudDirectorCSINodeDriverRegistrar

	// vSphere CSI
	VsphereCSIDriver
	VsphereCSISyncer
	VsphereCSIAttacher
	VsphereCSILivenessProbe
	VsphereCSINodeDriverRegistar
	VsphereCSIProvisioner
	VsphereCSIResizer
	VsphereCSISnapshotter

	// GCP Compute Persistent Disk CSI
	GCPComputeCSIDriver
	GCPComputeCSIProvisioner
	GCPComputeCSIAttacher
	GCPComputeCSIResizer
	GCPComputeCSISnapshotter
	GCPComputeCSINodeDriverRegistrar

	// Calico VXLAN
	CalicoVXLANCNI
	CalicoVXLANController
	CalicoVXLANNode
)

func FindResource(name string) (Resource, error) {
	for res := range allResources() {
		if res.String() == name {
			return res, nil
		}
	}

	return 0, fail.Runtime(fmt.Errorf("no such resource"), "image lookup %q", name)
}

func baseResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		CalicoCNI:              {"*": "quay.io/calico/cni:v3.26.3"},
		CalicoController:       {"*": "quay.io/calico/kube-controllers:v3.26.3"},
		CalicoNode:             {"*": "quay.io/calico/node:v3.26.3"},
		DNSNodeCache:           {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.22.28"},
		Flannel:                {"*": "docker.io/flannel/flannel:v0.21.3"},
		MachineController:      {"*": "quay.io/kubermatic/machine-controller:v1.58.1"},
		MetricsServer:          {"*": "registry.k8s.io/metrics-server/metrics-server:v0.7.0"},
		OperatingSystemManager: {"*": "quay.io/kubermatic/operating-system-manager:d16de939d7164849ab22f5903d528fcca215b9e9"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		AwsCCM: {
			"1.27.x":    "registry.k8s.io/provider-aws/cloud-controller-manager:v1.27.5",
			"1.28.x":    "registry.k8s.io/provider-aws/cloud-controller-manager:v1.28.4",
			">= 1.29.0": "registry.k8s.io/provider-aws/cloud-controller-manager:v1.29.1",
		},

		CSISnapshotController: {"*": "registry.k8s.io/sig-storage/snapshot-controller:v6.3.3"},
		CSISnapshotWebhook:    {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v6.3.3"},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.26.1"},
		AwsEbsCSIAttacher:            {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-attacher:v4.4.3-eks-1-29-2"},
		AwsEbsCSILivenessProbe:       {"*": "public.ecr.aws/eks-distro/kubernetes-csi/livenessprobe:v2.11.0-eks-1-29-2"},
		AwsEbsCSINodeDriverRegistrar: {"*": "public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.9.3-eks-1-29-2"},
		AwsEbsCSIProvisioner:         {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-provisioner:v3.6.3-eks-1-29-2"},
		AwsEbsCSIResizer:             {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-resizer:v1.9.3-eks-1-29-2"},
		AwsEbsCSISnapshotter:         {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-snapshotter/csi-snapshotter:v6.3.3-eks-1-29-2"},

		// Azure CCM
		AzureCCM: {
			"1.26.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.26.12",
			"1.27.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.27.6",
			"1.28.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.28.5",
			">= 1.29.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.29.0",
		},
		AzureCNM: {
			"1.26.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.26.12",
			"1.27.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.27.6",
			"1.28.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.28.5",
			">= 1.29.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.29.0",
		},

		// AzureFile CSI driver
		AzureFileCSI:                   {"*": "mcr.microsoft.com/oss/kubernetes-csi/azurefile-csi:v1.29.3"},
		AzureFileCSIAttacher:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v4.3.0"},
		AzureFileCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.10.0"},
		AzureFileCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.8.0"},
		AzureFileCSIProvisioner:        {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v3.5.0"},
		AzureFileCSIResizer:            {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.8.0"},
		AzureFileCSISnapshotter:        {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v6.3.3"},

		// AzureDisk CSI driver
		AzureDiskCSI:                   {"*": "mcr.microsoft.com/oss/kubernetes-csi/azuredisk-csi:v1.29.3"},
		AzureDiskCSIAttacher:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v4.3.0"},
		AzureDiskCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.10.0"},
		AzureDiskCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.8.0"},
		AzureDiskCSIProvisioner:        {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v3.5.0"},
		AzureDiskCSIResizer:            {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.8.0"},
		AzureDiskCSISnapshotter:        {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v6.3.3"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.47"},

		// DigitalOcean CSI
		DigitalOceanCSI:                   {"*": "docker.io/digitalocean/do-csi-plugin:v4.8.0"},
		DigitalOceanCSIAlpine:             {"*": "docker.io/alpine:3"},
		DigitalOceanCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.4.3"},
		DigitalOceanCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.3"},
		DigitalOceanCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.6.3"},
		DigitalOceanCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.9.3"},
		DigitalOceanCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v6.3.3"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.19.0"},

		// Hetzner CSI
		HetznerCSI:                   {"*": "docker.io/hetznercloud/hcloud-csi-driver:v2.6.0"},
		HetznerCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.1.0"},
		HetznerCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.7.0"},
		HetznerCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.4.0"},
		HetznerCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.9.0"},
		HetznerCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.7.0"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.26.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.26.4",
			"1.27.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.27.3",
			"1.28.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.28.1",
			">= 1.29.0": "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.29.0",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.26.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.26.4",
			"1.27.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.27.3",
			"1.28.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.28.1",
			">= 1.29.0": "registry.k8s.io/provider-os/cinder-csi-plugin:v1.29.0",
		},
		OpenstackCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.2"},
		OpenstackCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.11.0"},
		OpenstackCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.4.2"},
		OpenstackCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.6.2"},
		OpenstackCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.9.2"},
		OpenstackCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v6.3.3"},

		// Equinix Metal CCM
		EquinixMetalCCM: {"*": "quay.io/equinix-oss/cloud-provider-equinix-metal:v3.8.0"},

		// VMware Cloud Director CSI
		VMwareCloudDirectorCSI:                    {"*": "projects.registry.vmware.com/vmware-cloud-director/cloud-director-named-disk-csi-driver:1.5.0"},
		VMwareCloudDirectorCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v3.2.1"},
		VMwareCloudDirectorCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v2.2.2"},
		VMwareCloudDirectorCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.2.0"},

		// vSphere CPI (A.K.A. CCM)
		VsphereCCM: {
			"1.26.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.26.2",
			"1.27.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.27.0",
			"1.28.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.28.0",
			">= 1.29.0": "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.29.0",
		},

		// vSphere CSI
		VsphereCSIDriver:             {"*": "gcr.io/cloud-provider-vsphere/csi/release/driver:v3.1.2"},
		VsphereCSISyncer:             {"*": "gcr.io/cloud-provider-vsphere/csi/release/syncer:v3.1.2"},
		VsphereCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.4.2"},
		VsphereCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.11.0"},
		VsphereCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.2"},
		VsphereCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.6.2"},
		VsphereCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.9.2"},
		VsphereCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v6.3.3"},

		// Nutanix CCM
		NutanixCCM: {"*": "ghcr.io/nutanix-cloud-native/cloud-provider-nutanix/controller:v0.3.2"},

		// Nutanix CSI
		NutanixCSI:              {"*": "quay.io/karbon/ntnx-csi:v2.6.6"},
		NutanixCSILivenessProbe: {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.11.0"},
		NutanixCSIProvisioner:   {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.6.2"},
		NutanixCSIRegistrar:     {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.1"},
		NutanixCSIResizer:       {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.9.2"},
		NutanixCSISnapshotter:   {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v3.0.3"},

		// GCP CCM
		GCPCCM: {"*": "registry.k8s.io/cloud-provider-gcp/cloud-controller-manager:v28.2.1"},

		// GCP Compute Persistent Disk CSI
		// see: https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/blob/master/deploy/kubernetes/images/stable-master/image.yaml
		GCPComputeCSIDriver:              {"*": "registry.k8s.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.13.0"},
		GCPComputeCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.6.3"},
		GCPComputeCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.4.3"},
		GCPComputeCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.9.3"},
		GCPComputeCSISnapshotter:         {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v6.3.3"},
		GCPComputeCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.3"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.14.3@sha256:e5ca22526e01469f8d10c14e2339a82a13ad70d9a359b879024715540eef4ace"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.14.3@sha256:c9613277b72103ed36e9c0d16b9a17cafd507461d59340e432e3e9c23468b5e2"},

		// Calico VXLAN
		CalicoVXLANCNI:        {"*": "quay.io/calico/cni:v3.26.3"},
		CalicoVXLANController: {"*": "quay.io/calico/kube-controllers:v3.26.3"},
		CalicoVXLANNode:       {"*": "quay.io/calico/node:v3.26.3"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.14.3@sha256:3f5c425faca4f8a38e29d4f86db2aadd8f33c9f05debd78a2bb2a24dd9e565d8"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.12.1@sha256:9e5f81ee747866480ea1ac4630eb6975ff9227f9782b7c93919c081c33f38267"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.12.1@sha256:1f86f3400827a0451e6332262467f894eeb7caf0eb8779bd951e2caa9d027cbe"},
		CiliumCertGen:   {"*": "quay.io/cilium/certgen:v0.1.9@sha256:89a0847753686444daabde9474b48340993bd19c7bea66a46e45b2974b82041f"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.25.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.3",
			"1.26.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.26.4",
			"1.27.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.27.3",
			">= 1.28.0": "registry.k8s.io/autoscaling/cluster-autoscaler:v1.28.0",
		},

		// CSI Vault Secret Provider
		CSIVaultSecretProvider: {"*": "docker.io/hashicorp/vault-csi-provider:1.1.0"},

		// CSI Secrets Driver
		SecretStoreCSIDriverNodeRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		SecretStoreCSIDriver:              {"*": "registry.k8s.io/csi-secrets-store/driver:v1.2.1"},
		SecretStoreCSIDriverLivenessProbe: {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.7.0"},
		SecretStoreCSIDriverCRDs:          {"*": "registry.k8s.io/csi-secrets-store/driver-crds:v1.2.1"},
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
	case ListFilterNone:
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
		nt, _ := named.(reference.NamedTagged)
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
