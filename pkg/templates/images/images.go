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
	"maps"
	"slices"
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
	CiliumEnvoy
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
	NutanixCSIExternalHealthMonitor
	NutanixCSIAttacher
	NutanixCSIPrecheck
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
	VMwareCloudDirectorCSIResizer
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

	// KubeVirt's CCM
	KubeVirtCCM

	// KubeVirt CSI
	KubeVirtCSI
	KubeVirtCSINodeDriverRegistrar
	KubeVirtCSILivenessProbe
	KubeVirtCSIProvisioner
	KubeVirtCSIAttacher
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
		CalicoCNI:              {"*": "quay.io/calico/cni:v3.30.3"},
		CalicoController:       {"*": "quay.io/calico/kube-controllers:v3.30.3"},
		CalicoNode:             {"*": "quay.io/calico/node:v3.30.3"},
		DNSNodeCache:           {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.26.4"},
		Flannel:                {"*": "docker.io/flannel/flannel:v0.24.4"},
		MachineController:      {"*": "quay.io/kubermatic/machine-controller:1b7f31166cf4827955e52acb8581a7aa48eebba4"},
		MetricsServer:          {"*": "registry.k8s.io/metrics-server/metrics-server:v0.8.0"},
		OperatingSystemManager: {"*": "quay.io/kubermatic/operating-system-manager:6ee458b6ff4783518bd77b80b0bef859b2e7c44f"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		AwsCCM: {
			"1.32.x":    "registry.k8s.io/provider-aws/cloud-controller-manager:v1.32.5",
			">= 1.33.0": "registry.k8s.io/provider-aws/cloud-controller-manager:v1.33.0",
		},

		CSISnapshotController: {"*": "registry.k8s.io/sig-storage/snapshot-controller:v8.1.1"},
		CSISnapshotWebhook:    {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v8.1.1"},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.49.0"},
		AwsEbsCSIAttacher:            {"*": "public.ecr.aws/csi-components/csi-attacher:v4.9.0-eksbuild.4"},
		AwsEbsCSILivenessProbe:       {"*": "public.ecr.aws/csi-components/livenessprobe:v2.16.0-eksbuild.5"},
		AwsEbsCSINodeDriverRegistrar: {"*": "public.ecr.aws/csi-components/csi-node-driver-registrar:v2.14.0-eksbuild.5"},
		AwsEbsCSIProvisioner:         {"*": "public.ecr.aws/csi-components/csi-provisioner:v5.3.0-eksbuild.4"},
		AwsEbsCSIResizer:             {"*": "public.ecr.aws/csi-components/csi-resizer:v1.14.0-eksbuild.4"},
		AwsEbsCSISnapshotter:         {"*": "public.ecr.aws/csi-components/csi-snapshotter:v8.3.0-eksbuild.2"},

		// Azure CCM
		AzureCCM: {
			"1.32.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.32.8",
			"1.33.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.33.3",
			">= 1.34.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.33.1",
		},
		AzureCNM: {
			"1.32.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.32.8",
			"1.33.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.33.3",
			">= 1.33.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.34.1",
		},

		// AzureFile CSI driver
		AzureFileCSI:                   {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/azurefile-csi:v1.34.0"},
		AzureFileCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/livenessprobe:v2.17.0"},
		AzureFileCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-node-driver-registrar:v2.15.0"},
		AzureFileCSIProvisioner:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-provisioner:v5.3.0"},
		AzureFileCSIResizer:            {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-resizer:v1.14.0"},
		AzureFileCSISnapshotter:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-snapshotter:v8.3.0"},

		// AzureDisk CSI driver5
		AzureDiskCSI:                   {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/azuredisk-csi:v1.33.5"},
		AzureDiskCSIAttacher:           {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-attacher:v4.10.1"},
		AzureDiskCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/livenessprobe:v2.17.0"},
		AzureDiskCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-node-driver-registrar:v2.15.0"},
		AzureDiskCSIProvisioner:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-provisioner:v5.3.0"},
		AzureDiskCSIResizer:            {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-resizer:v1.14.0"},
		AzureDiskCSISnapshotter:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-snapshotter:v8.3.0"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.63"},

		// DigitalOcean CSI
		DigitalOceanCSI:                   {"*": "digitalocean/do-csi-plugin:v4.14.0"},
		DigitalOceanCSIAlpine:             {"*": "docker.io/alpine:3"},
		DigitalOceanCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.8.1"},
		DigitalOceanCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.13.0"},
		DigitalOceanCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.2.0"},
		DigitalOceanCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.13.2"},
		DigitalOceanCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.2.0"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.26.0"},

		// Hetzner CSI
		HetznerCSI:                   {"*": "docker.io/hetznercloud/hcloud-csi-driver:v2.17.0"},
		HetznerCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.9.0"},
		HetznerCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.14.0"},
		HetznerCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.3.0"},
		HetznerCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.16.0"},
		HetznerCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.14.0"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.32.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.32.1",
			"1.33.x": "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.33.1",
			">= 1.34.0": "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.34.0",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.32.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.32.1",
			"1.33.0": "registry.k8s.io/provider-os/cinder-csi-plugin:v1.33.1",
			">= 1.34.0": "registry.k8s.io/provider-os/cinder-csi-plugin:v1.34.0",
		},
		OpenstackCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.15.0"},
		OpenstackCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.17.0"},
		OpenstackCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.10.0"},
		OpenstackCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.3.0"},
		OpenstackCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.14.0"},
		OpenstackCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.3.0"},

		// Equinix Metal CCM
		EquinixMetalCCM: {"*": "quay.io/equinix-oss/cloud-provider-equinix-metal:v3.8.1"},

		// VMware Cloud Director CSI
		VMwareCloudDirectorCSI:                    {"*": "projects.registry.vmware.com/vmware-cloud-director/cloud-director-named-disk-csi-driver:1.6.0"},
		VMwareCloudDirectorCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v3.2.1"},
		VMwareCloudDirectorCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v2.2.2"},
		VMwareCloudDirectorCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.4.0"},
		VMwareCloudDirectorCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.2.0"},

		// vSphere CPI (A.K.A. CCM)
		VsphereCCM: {
			"1.32.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.32.2",
			"1.33.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.33.0",
			">= 1.34.0": "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.34.0",
		},

		// vSphere CSI
		VsphereCSIDriver:             {"*": "registry.k8s.io/csi-vsphere/driver:v3.5.0"},
		VsphereCSISyncer:             {"*": "registry.k8s.io/csi-vsphere/syncer:v3.5.0"},
		VsphereCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.8.1"},
		VsphereCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.15.0"},
		VsphereCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.13.0"},
		VsphereCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v4.0.1"},
		VsphereCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.12.0"},
		VsphereCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.2.0"},

		// Nutanix CCM
		NutanixCCM: {"*": "ghcr.io/nutanix-cloud-native/cloud-provider-nutanix/controller:v0.5.2"},

		// Nutanix CSI
		NutanixCSI:                      {"*": "docker.io/nutanix/ntnx-csi:3.3.4"},
		NutanixCSILivenessProbe:         {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.15.0"},
		NutanixCSIExternalHealthMonitor: {"*": "registry.k8s.io/sig-storage/csi-external-health-monitor-controller:v0.14.0"},
		NutanixCSIAttacher:              {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.8.1"},
		NutanixCSIProvisioner:           {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.2.0"},
		NutanixCSIRegistrar:             {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.13.0"},
		NutanixCSIPrecheck:              {"*": "docker.io/nutanix/ntnx-csi-precheck:3.3.4"},
		NutanixCSIResizer:               {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.13.2"},
		NutanixCSISnapshotter:           {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v3.0.3"},

		// GCP CCM
		GCPCCM: {"*": "registry.k8s.io/cloud-provider-gcp/cloud-controller-manager:v33.1.1"},

		// GCP Compute Persistent Disk CSI
		// see: https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/blob/master/deploy/kubernetes/images/stable-master/image.yaml
		GCPComputeCSIDriver:              {"*": "registry.k8s.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.17.3"},
		GCPComputeCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.2.0"},
		GCPComputeCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.8.1"},
		GCPComputeCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.13.2"},
		GCPComputeCSISnapshotter:         {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.2.1"},
		GCPComputeCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.13.0"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.18.2@"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.18.2@sha256:cb4e4ffc5789fd5ff6a534e3b1460623df61cba00f5ea1c7b40153b5efb81805"},
		CiliumEnvoy:    {"*": "quay.io/cilium/v1.34.7-1757592137-1a52bb680a956879722f48c591a2ca90f7791324@sha256:7932d656b63f6f866b6732099d33355184322123cfe1182e6f05175a3bc2e0e0"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.18.2@sha256:6079308ee15e44dff476fb522612732f7c5c4407a1017bc3470916242b0405ac"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.13.3@sha256:661d5de7050182d495c6497ff0b007a7a1e379648e60830dd68c4d78ae21761d"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.13.3@sha256:db1454e45dc39ca41fbf7cad31eec95d99e5b9949c39daaad0fa81ef29d56953"},
		CiliumCertGen:   {"*": "quay.io/cilium/certgen:v0.2.4@sha256:de7b97b1d19a34b674d0c4bc1da4db999f04ae355923a9a994ac3a81e1a1b5ff"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.32.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.32.3",
			">= 1.33.0": "registry.k8s.io/autoscaling/cluster-autoscaler:v1.33.1",
		},

		// CSI Vault Secret Provider
		CSIVaultSecretProvider: {"*": "docker.io/hashicorp/vault-csi-provider:1.1.0"},

		// CSI Secrets Driver
		SecretStoreCSIDriverNodeRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		SecretStoreCSIDriver:              {"*": "registry.k8s.io/csi-secrets-store/driver:v1.2.1"},
		SecretStoreCSIDriverLivenessProbe: {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.7.0"},
		SecretStoreCSIDriverCRDs:          {"*": "registry.k8s.io/csi-secrets-store/driver-crds:v1.2.1"},

		// KubeVirt's CCM
		KubeVirtCCM: {"*": "quay.io/kubevirt/kubevirt-cloud-controller-manager:v0.5.1"},

		// KubeVirt CSI
		KubeVirtCSI:                    {"*": "quay.io/kubermatic/kubevirt-csi-driver:v0.4.4"},
		KubeVirtCSINodeDriverRegistrar: {"*": "quay.io/openshift/origin-csi-node-driver-registrar:4.20.0"},
		KubeVirtCSILivenessProbe:       {"*": "quay.io/openshift/origin-csi-livenessprobe:4.20.0"},
		KubeVirtCSIProvisioner:         {"*": "quay.io/openshift/origin-csi-external-provisioner:4.20.0"},
		KubeVirtCSIAttacher:            {"*": "quay.io/openshift/origin-csi-external-attacher:4.20.0"},
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

func (r *Resolver) ListAll() []string {
	resources := allResources()

	// create a bool map, to deduplicate the images
	listMap := make(map[string]bool)
	for res := range resources {
		for _, img := range resources[res] {
			listMap[img] = true
		}
	}

	list := slices.Collect(maps.Keys(listMap))

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
