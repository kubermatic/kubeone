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
	"github.com/distribution/distribution/v3/reference"

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

	// Nutanix CSI
	NutanixCSILivenessProbe
	NutanixCSI
	NutanixCSIProvisioner
	NutanixCSIRegistrar
	NutanixCSIResizer
	NutanixCSISnapshotter
	NutanixCSISnapshotController
	NutanixCSISnapshotValidationWebhook

	// DigitalOcean CSI
	DigitalOceanCSI
	DigitalOceanCSIAlpine
	DigitalOceanCSIAttacher
	DigitalOceanCSINodeDriverRegistar
	DigitalOceanCSIProvisioner
	DigitalOceanCSIResizer
	DigitalOceanCSISnapshotController
	DigitalOceanCSISnapshotValidationWebhook
	DigitalOceanCSISnapshotter

	// OpenStack CSI
	OpenstackCSI
	OpenstackCSINodeDriverRegistar
	OpenstackCSILivenessProbe
	OpenstackCSIAttacher
	OpenstackCSIProvisioner
	OpenstackCSIResizer
	OpenstackCSISnapshotter
	OpenstackCSISnapshotController
	OpenstackCSISnapshotWebhook

	// Hetzner CSI
	HetznerCSI
	HetznerCSIAttacher
	HetznerCSIResizer
	HetznerCSIProvisioner
	HetznerCSILivenessProbe
	HetznerCSINodeDriverRegistar

	// CCMs and CSI plugins
	DigitaloceanCCM
	HetznerCCM
	OpenstackCCM
	EquinixMetalCCM
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
	VsphereCSISnapshotController
	VsphereCSISnapshotValidationWebhook

	// GCP Compute Persistent Disk CSI
	GCPComputeCSIDriver
	GCPComputeCSIProvisioner
	GCPComputeCSIAttacher
	GCPComputeCSIResizer
	GCPComputeCSISnapshotter
	GCPComputeCSISnapshotController
	GCPComputeCSISnapshotValidationWebhook
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
		CalicoCNI:        {"*": "quay.io/calico/cni:v3.26.0"},
		CalicoController: {"*": "quay.io/calico/kube-controllers:v3.26.0"},
		CalicoNode:       {"*": "quay.io/calico/node:v3.26.0"},
		DNSNodeCache:     {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.22.23"},
		Flannel:          {"*": "docker.io/flannel/flannel:v0.21.3"},
		// TODO: Switch to semver image before release.
		MachineController:      {"*": "quay.io/kubermatic/machine-controller:v1.57.0"},
		MetricsServer:          {"*": "registry.k8s.io/metrics-server/metrics-server:v0.6.3"},
		OperatingSystemManager: {"*": "quay.io/kubermatic/operating-system-manager:v1.3.0"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		AwsCCM: {
			"1.24.x":    "registry.k8s.io/provider-aws/cloud-controller-manager:v1.24.4",
			"1.25.x":    "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.3",
			"1.26.x":    "registry.k8s.io/provider-aws/cloud-controller-manager:v1.26.1",
			">= 1.27.0": "registry.k8s.io/provider-aws/cloud-controller-manager:v1.27.1",
		},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.20.0"},
		AwsEbsCSIAttacher:            {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-attacher:v4.3.0-eks-1-27-3"},
		AwsEbsCSILivenessProbe:       {"*": "public.ecr.aws/eks-distro/kubernetes-csi/livenessprobe:v2.10.0-eks-1-27-3"},
		AwsEbsCSINodeDriverRegistrar: {"*": "public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.8.0-eks-1-27-3"},
		AwsEbsCSIProvisioner:         {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-provisioner:v3.5.0-eks-1-27-3"},
		AwsEbsCSIResizer:             {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-resizer:v1.8.0-eks-1-27-3"},
		AwsEbsCSISnapshotter:         {"*": "public.ecr.aws/eks-distro/kubernetes-csi/external-snapshotter/csi-snapshotter:v6.2.1-eks-1-27-3"},
		AwsEbsCSISnapshotController:  {"*": "registry.k8s.io/sig-storage/snapshot-controller:v6.2.1"},

		// Azure CCM
		AzureCCM: {
			"1.24.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.11",
			"1.25.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.5",
			">= 1.26.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.26.0",
		},
		AzureCNM: {
			"1.24.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.11",
			"1.25.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.5",
			">= 1.26.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.26.0",
		},

		// AzureFile CSI driver
		AzureFileCSI:                      {"*": "mcr.microsoft.com/oss/kubernetes-csi/azurefile-csi:v1.24.0"},
		AzureFileCSIAttacher:              {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v3.5.0"},
		AzureFileCSILivenessProbe:         {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.7.0"},
		AzureFileCSINodeDriverRegistar:    {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.5.1"},
		AzureFileCSIProvisioner:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v3.2.0"},
		AzureFileCSIResizer:               {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.5.0"},
		AzureFileCSISnapshotter:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v5.0.1"},
		AzureFileCSISnapshotterController: {"*": "mcr.microsoft.com/oss/kubernetes-csi/snapshot-controller:v5.0.1"},

		// AzureDisk CSI driver
		AzureDiskCSI:                      {"*": "mcr.microsoft.com/oss/kubernetes-csi/azuredisk-csi:v1.25.0"},
		AzureDiskCSIAttacher:              {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v3.5.0"},
		AzureDiskCSILivenessProbe:         {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.7.0"},
		AzureDiskCSINodeDriverRegistar:    {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.5.1"},
		AzureDiskCSIProvisioner:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v3.2.0"},
		AzureDiskCSIResizer:               {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.5.0"},
		AzureDiskCSISnapshotter:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v5.0.1"},
		AzureDiskCSISnapshotterController: {"*": "mcr.microsoft.com/oss/kubernetes-csi/snapshot-controller:v5.0.1"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.41"},

		DigitalOceanCSI:                          {"*": "docker.io/digitalocean/do-csi-plugin:v4.5.0"},
		DigitalOceanCSIAlpine:                    {"*": "docker.io/alpine:3"},
		DigitalOceanCSIAttacher:                  {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.1.0"},
		DigitalOceanCSINodeDriverRegistar:        {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.7.0"},
		DigitalOceanCSIProvisioner:               {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.4.0"},
		DigitalOceanCSIResizer:                   {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.7.0"},
		DigitalOceanCSISnapshotController:        {"*": "registry.k8s.io/sig-storage/snapshot-controller:v6.2.1"},
		DigitalOceanCSISnapshotValidationWebhook: {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v6.2.1"},
		DigitalOceanCSISnapshotter:               {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v6.2.1"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.16.0"},

		// Hetzner CSI
		HetznerCSI:                   {"*": "docker.io/hetznercloud/hcloud-csi-driver:v2.3.2"},
		HetznerCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.1.0"},
		HetznerCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.7.0"},
		HetznerCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.4.0"},
		HetznerCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.9.0"},
		HetznerCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.7.0"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.24.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.24.5",
			"1.25.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.25.3",
			">= 1.26.0": "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.0",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.24.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.24.5",
			"1.25.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.25.3",
			">= 1.26.0": "docker.io/k8scloudprovider/cinder-csi-plugin:v1.26.0",
		},
		OpenstackCSINodeDriverRegistar: {
			"< 1.26.0":  "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.5.1",
			">= 1.26.0": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.6.2",
		},
		OpenstackCSILivenessProbe: {
			"< 1.26.0":  "registry.k8s.io/sig-storage/livenessprobe:v2.7.0",
			">= 1.26.0": "registry.k8s.io/sig-storage/livenessprobe:v2.8.0",
		},
		OpenstackCSIAttacher: {
			"< 1.26.0":  "registry.k8s.io/sig-storage/csi-attacher:v3.4.0",
			">= 1.26.0": "registry.k8s.io/sig-storage/csi-attacher:v4.0.0",
		},
		OpenstackCSIProvisioner: {
			"< 1.26.0":  "registry.k8s.io/sig-storage/csi-provisioner:v3.1.0",
			">= 1.26.0": "registry.k8s.io/sig-storage/csi-provisioner:v3.4.0",
		},
		OpenstackCSIResizer: {
			"< 1.26.0":  "registry.k8s.io/sig-storage/csi-resizer:v1.4.0",
			">= 1.26.0": "registry.k8s.io/sig-storage/csi-resizer:v1.6.0",
		},
		OpenstackCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v6.1.0"},
		OpenstackCSISnapshotController: {"*": "registry.k8s.io/sig-storage/snapshot-controller:v6.1.0"},
		OpenstackCSISnapshotWebhook:    {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v6.1.0"},

		// Equinix Metal CCM
		EquinixMetalCCM: {"*": "ghcr.io/equinix/cloud-provider-equinix-metal:v3.5.0"},

		// vSphere CCM
		VsphereCCM: {
			"1.24.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.24.2",
			">= 1.25.0": "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.25.0",
		},

		// VMware Cloud Director CSI
		VMwareCloudDirectorCSI:                    {"*": "projects.registry.vmware.com/vmware-cloud-director/cloud-director-named-disk-csi-driver:1.3.2"},
		VMwareCloudDirectorCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v3.2.1"},
		VMwareCloudDirectorCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v2.2.2"},
		VMwareCloudDirectorCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.2.0"},

		// vSphere CSI
		VsphereCSIDriver:                    {"*": "gcr.io/cloud-provider-vsphere/csi/release/driver:v2.7.0"},
		VsphereCSISyncer:                    {"*": "gcr.io/cloud-provider-vsphere/csi/release/syncer:v2.7.0"},
		VsphereCSIAttacher:                  {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.0.0"},
		VsphereCSILivenessProbe:             {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.7.0"},
		VsphereCSINodeDriverRegistar:        {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		VsphereCSIProvisioner:               {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.2.1"},
		VsphereCSIResizer:                   {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.6.0"},
		VsphereCSISnapshotter:               {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v6.0.1"},
		VsphereCSISnapshotController:        {"*": "registry.k8s.io/sig-storage/snapshot-controller:v6.0.1"},
		VsphereCSISnapshotValidationWebhook: {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v6.0.1"},

		// Nutanix CSI
		NutanixCSI:                          {"*": "quay.io/karbon/ntnx-csi:v2.6.3"},
		NutanixCSILivenessProbe:             {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.10.0"},
		NutanixCSIProvisioner:               {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.5.0"},
		NutanixCSIRegistrar:                 {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.8.0"},
		NutanixCSIResizer:                   {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.8.0"},
		NutanixCSISnapshotter:               {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v3.0.3"},
		NutanixCSISnapshotController:        {"*": "registry.k8s.io/sig-storage/snapshot-controller:v6.2.1"},
		NutanixCSISnapshotValidationWebhook: {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v6.2.1"},

		// GCP Compute Persistent Disk CSI
		GCPComputeCSIDriver:                    {"*": "registry.k8s.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.8.1"},
		GCPComputeCSIProvisioner:               {"*": "registry.k8s.io/sig-storage/csi-provisioner:v3.1.0"},
		GCPComputeCSIAttacher:                  {"*": "registry.k8s.io/sig-storage/csi-attacher:v3.4.0"},
		GCPComputeCSIResizer:                   {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.4.0"},
		GCPComputeCSISnapshotter:               {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v6.1.0"},
		GCPComputeCSISnapshotController:        {"*": "registry.k8s.io/sig-storage/snapshot-controller:v6.1.0"},
		GCPComputeCSISnapshotValidationWebhook: {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v6.1.0"},
		GCPComputeCSINodeDriverRegistrar:       {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.5.0"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.13.3@sha256:77176464a1e11ea7e89e984ac7db365e7af39851507e94f137dcf56c87746314"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.13.3@sha256:fa7003cbfdf8358cb71786afebc711b26e5e44a2ed99bd4944930bba915b8910"},

		// Calico VXLAN
		CalicoVXLANCNI:        {"*": "quay.io/calico/cni:v3.23.3"},
		CalicoVXLANController: {"*": "quay.io/calico/kube-controllers:v3.23.3"},
		CalicoVXLANNode:       {"*": "quay.io/calico/node:v3.23.3"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.13.3@sha256:19e4aae5ff72cd9fbcb7d2d16a1570533320a478acc015fc91a4d41a177cadf6"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.11.0@sha256:bcb369c47cada2d4257d63d3749f7f87c91dde32e010b223597306de95d1ecc8"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.11.0@sha256:14c04d11f78da5c363f88592abae8d2ecee3cbe009f443ef11df6ac5f692d839"},
		CiliumCertGen:   {"*": "quay.io/cilium/certgen:v0.1.8@sha256:4a456552a5f192992a6edcec2febb1c54870d665173a33dc7d876129b199ddbd"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.24.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.24.0",
			"1.25.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.0",
			">= 1.26.0": "registry.k8s.io/autoscaling/cluster-autoscaler:v1.26.1",
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
