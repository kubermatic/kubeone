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
		CalicoCNI:              {"*": "quay.io/calico/cni:v3.28.2"},
		CalicoController:       {"*": "quay.io/calico/kube-controllers:v3.28.2"},
		CalicoNode:             {"*": "quay.io/calico/node:v3.28.2"},
		DNSNodeCache:           {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.23.1"},
		Flannel:                {"*": "docker.io/flannel/flannel:v0.24.3"},
		MachineController:      {"*": "quay.io/kubermatic/machine-controller:v1.60.0"},
		MetricsServer:          {"*": "registry.k8s.io/metrics-server/metrics-server:v0.7.2"},
		OperatingSystemManager: {"*": "quay.io/kubermatic/operating-system-manager:v1.6.0"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		AwsCCM: {
			"1.29.x":    "registry.k8s.io/provider-aws/cloud-controller-manager:v1.29.6",
			"1.30.x":    "registry.k8s.io/provider-aws/cloud-controller-manager:v1.30.3",
			">= 1.31.0": "registry.k8s.io/provider-aws/cloud-controller-manager:v1.31.1",
		},

		CSISnapshotController: {"*": "registry.k8s.io/sig-storage/snapshot-controller:v8.1.0"},
		CSISnapshotWebhook:    {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v8.1.0"},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.35.0"},
		AwsEbsCSIAttacher:            {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-attacher:v4.7.0-eks-1-31-3"},
		AwsEbsCSILivenessProbe:       {"*": "public.ecr.aws/eks-distro/kubernetes-csi/livenessprobe:v2.14.0-eks-1-31-3"},
		AwsEbsCSINodeDriverRegistrar: {"*": "public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.12.0-eks-1-31-3"},
		AwsEbsCSIProvisioner:         {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-provisioner:v5.1.0-eks-1-31-3"},
		AwsEbsCSIResizer:             {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-resizer:v1.12.0-eks-1-31-3"},
		AwsEbsCSISnapshotter:         {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.1.0"},

		// Azure CCM
		AzureCCM: {
			"1.29.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.29.11",
			"1.30.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.30.7",
			">= 1.31.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.31.1",
		},
		AzureCNM: {
			"1.29.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.29.11",
			"1.30.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.30.7",
			">= 1.31.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.31.1",
		},

		// AzureFile CSI driver
		AzureFileCSI:                   {"*": "mcr.microsoft.com/oss/kubernetes-csi/azurefile-csi:v1.30.5"},
		AzureFileCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.13.1"},
		AzureFileCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.12.0"},
		AzureFileCSIProvisioner:        {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v5.1.0"},
		AzureFileCSIResizer:            {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.11.1"},
		AzureFileCSISnapshotter:        {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v8.1.0"},

		// AzureDisk CSI driver
		AzureDiskCSI:                   {"*": "mcr.microsoft.com/oss/kubernetes-csi/azuredisk-csi:v1.30.3"},
		AzureDiskCSIAttacher:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v4.6.1"},
		AzureDiskCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.13.1"},
		AzureDiskCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.12.0"},
		AzureDiskCSIProvisioner:        {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v5.1.0"},
		AzureDiskCSIResizer:            {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.11.1"},
		AzureDiskCSISnapshotter:        {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v8.1.0"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.56"},

		// DigitalOcean CSI
		DigitalOceanCSI:                   {"*": "digitalocean/do-csi-plugin:v4.12.0"},
		DigitalOceanCSIAlpine:             {"*": "docker.io/alpine:3"},
		DigitalOceanCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.7.0"},
		DigitalOceanCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.12.0"},
		DigitalOceanCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.1.0"},
		DigitalOceanCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.12.0"},
		DigitalOceanCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.1.0"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.20.0"},

		// Hetzner CSI
		HetznerCSI:                   {"*": "docker.io/hetznercloud/hcloud-csi-driver:v2.9.0"},
		HetznerCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.6.1"},
		HetznerCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.11.2"},
		HetznerCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.0.2"},
		HetznerCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.13.1"},
		HetznerCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.11.1"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.29.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.29.1",
			"1.30.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.30.2",
			">= 1.31.0": "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.31.2",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.29.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.29.1",
			"1.30.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.30.2",
			">= 1.31.0": "registry.k8s.io/provider-os/cinder-csi-plugin:v1.31.2",
		},
		OpenstackCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.2"},
		OpenstackCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.11.0"},
		OpenstackCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.4.2"},
		OpenstackCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.6.2"},
		OpenstackCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.9.2"},
		OpenstackCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.1.0"},

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
			"1.29.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.29.0",
			"1.30.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.30.1",
			">= 1.31.0": "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.31.0",
		},

		// vSphere CSI
		VsphereCSIDriver:             {"*": "registry.k8s.io/csi-vsphere/driver:v3.3.1"},
		VsphereCSISyncer:             {"*": "registry.k8s.io/csi-vsphere/syncer:v3.3.1"},
		VsphereCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.5.1"},
		VsphereCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.12.0"},
		VsphereCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.10.1"},
		VsphereCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v4.0.1"},
		VsphereCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.10.1"},
		VsphereCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.1.0"},

		// Nutanix CCM
		NutanixCCM: {"*": "ghcr.io/nutanix-cloud-native/cloud-provider-nutanix/controller:v0.4.1"},

		// Nutanix CSI
		NutanixCSI:              {"*": "quay.io/karbon/ntnx-csi:v2.6.10"},
		NutanixCSILivenessProbe: {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.11.0"},
		NutanixCSIProvisioner:   {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.6.2"},
		NutanixCSIRegistrar:     {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.1"},
		NutanixCSIResizer:       {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.9.2"},
		NutanixCSISnapshotter:   {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.1.0"},

		// GCP CCM
		GCPCCM: {
			"1.29.x":    "registry.k8s.io/cloud-provider-gcp/cloud-controller-manager:v29.0.0",
			"1.30.x":    "registry.k8s.io/cloud-provider-gcp/cloud-controller-manager:v30.0.0",
			">= 1.31.0": "registry.k8s.io/cloud-provider-gcp/cloud-controller-manager:v30.0.0",
		},

		// GCP Compute Persistent Disk CSI
		// see: https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/blob/master/deploy/kubernetes/images/stable-master/image.yaml
		GCPComputeCSIDriver:              {"*": "registry.k8s.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.15.0"},
		GCPComputeCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.1.0"},
		GCPComputeCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.4.3"},
		GCPComputeCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.11.1"},
		GCPComputeCSISnapshotter:         {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.1.0"},
		GCPComputeCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.3"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Calico VXLAN
		CalicoVXLANCNI:        {"*": "quay.io/calico/cni:v3.28.2"},
		CalicoVXLANController: {"*": "quay.io/calico/kube-controllers:v3.28.2"},
		CalicoVXLANNode:       {"*": "quay.io/calico/node:v3.28.2"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.16.3@sha256:62d2a09bbef840a46099ac4c69421c90f84f28d018d479749049011329aa7f28"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.16.3@sha256:6e2925ef47a1c76e183c48f95d4ce0d34a1e5e848252f910476c3e11ce1ec94b"},
		CiliumEnvoy:    {"*": "quay.io/cilium/cilium-envoy:v1.29.9-1728346947-0d05e48bfbb8c4737ec40d5781d970a550ed2bbd@sha256:42614a44e508f70d03a04470df5f61e3cffd22462471a0be0544cf116f2c50ba"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.16.3@sha256:feb60efd767e0e7863a94689f4a8db56a0acc7c1d2b307dee66422e3dc25a089"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.13.1@sha256:e2e9313eb7caf64b0061d9da0efbdad59c6c461f6ca1752768942bfeda0796c6"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.13.1@sha256:0e0eed917653441fded4e7cdb096b7be6a3bddded5a2dd10812a27b1fc6ed95b"},
		CiliumCertGen:   {"*": "quay.io/cilium/certgen:v0.2.0@sha256:169d93fd8f2f9009db3b9d5ccd37c2b753d0989e1e7cd8fe79f9160c459eef4f"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.29.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.29.4",
			"1.30.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.30.2",
			">= 1.31.0": "registry.k8s.io/autoscaling/cluster-autoscaler:v1.31.0",
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
