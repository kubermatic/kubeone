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
	"strings"

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

	// Backup Restic
	BackupResticSnapshotter
	BackupResticUploader

	// Unatttended Upgrades
	UUApline
	UUFluo
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
		CalicoCNI:              {"*": "quay.io/calico/cni:v3.32.0"},
		CalicoController:       {"*": "quay.io/calico/kube-controllers:v3.32.0"},
		CalicoNode:             {"*": "quay.io/calico/node:v3.32.0"},
		DNSNodeCache:           {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.26.7"},
		Flannel:                {"*": "docker.io/flannel/flannel:v0.24.4"},
		MachineController:      {"*": "quay.io/kubermatic/machine-controller:090279e10f6557926c29fce95405df4eaea44241"},
		MetricsServer:          {"*": "registry.k8s.io/metrics-server/metrics-server:v0.8.1"},
		OperatingSystemManager: {"*": "quay.io/kubermatic/operating-system-manager:0d531b49eca2a91b0e6c61609a32468668a615bb"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		AwsCCM: {
			"1.34.x":   "registry.k8s.io/provider-aws/cloud-controller-manager:v1.34.0",
			"1.35.x":   "registry.k8s.io/provider-aws/cloud-controller-manager:v1.35.0",
			">=1.36.x": "registry.k8s.io/provider-aws/cloud-controller-manager:v1.35.0",
		},

		CSISnapshotController: {"*": "registry.k8s.io/sig-storage/snapshot-controller:v8.1.1"},
		CSISnapshotWebhook:    {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v8.1.1"},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.61.1"},
		AwsEbsCSIAttacher:            {"*": "public.ecr.aws/csi-components/csi-attacher:v4.12.0-eksbuild.2"},
		AwsEbsCSILivenessProbe:       {"*": "public.ecr.aws/csi-components/livenessprobe:v2.19.0-eksbuild.2"},
		AwsEbsCSINodeDriverRegistrar: {"*": "public.ecr.aws/csi-components/csi-node-driver-registrar:v2.17.0-eksbuild.2"},
		AwsEbsCSIProvisioner:         {"*": "public.ecr.aws/csi-components/csi-provisioner:v6.2.0-eksbuild.7"},
		AwsEbsCSIResizer:             {"*": "public.ecr.aws/csi-components/csi-resizer:v2.2.0-eksbuild.2"},
		AwsEbsCSISnapshotter:         {"*": "public.ecr.aws/csi-components/csi-snapshotter:v8.6.0-eksbuild.2"},

		// Azure CCM
		AzureCCM: {
			"1.34.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.34.10",
			"1.35.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.35.5",
			">= 1.36.x": "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.36.1",
		},
		AzureCNM: {
			"1.34.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.34.10",
			"1.35.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.35.5",
			">= 1.36.x": "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.36.1",
		},

		// AzureFile CSI driver
		AzureFileCSI:                   {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/azurefile-csi:v1.35.3"},
		AzureFileCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/livenessprobe:v2.18.0"},
		AzureFileCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-node-driver-registrar:v2.16.0"},
		AzureFileCSIProvisioner:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-provisioner:v6.1.1"},
		AzureFileCSIResizer:            {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-resizer:v2.1.0"},
		AzureFileCSISnapshotter:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-snapshotter:v8.5.0"},

		// AzureDisk CSI driver5
		AzureDiskCSI:                   {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/azuredisk-csi:v1.34.4"},
		AzureDiskCSIAttacher:           {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-attacher:v4.11.0"},
		AzureDiskCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/livenessprobe:v2.18.0"},
		AzureDiskCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-node-driver-registrar:v2.16.0"},
		AzureDiskCSIProvisioner:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-provisioner:v6.2.0"},
		AzureDiskCSIResizer:            {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-resizer:v2.1.0"},
		AzureDiskCSISnapshotter:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-snapshotter:v8.5.0"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.67"},

		// DigitalOcean CSI
		DigitalOceanCSI:                   {"*": "docker.io/digitalocean/do-csi-plugin:v4.17.0"},
		DigitalOceanCSIAlpine:             {"*": "docker.io/alpine:3"},
		DigitalOceanCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.11.0"},
		DigitalOceanCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.16.0"},
		DigitalOceanCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.2.0"},
		DigitalOceanCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.1.0"},
		DigitalOceanCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.5.0"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.31.1"},

		// Hetzner CSI
		HetznerCSI:                   {"*": "docker.io/hetznercloud/hcloud-csi-driver:v2.21.2"},
		HetznerCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.11.0"},
		HetznerCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.1.0"},
		HetznerCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.2.0"},
		HetznerCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.18.0"},
		HetznerCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.16.0"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.34.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.34.1",
			"1.35.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.35.0",
			">= 1.36.x": "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.36.0",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.34.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.34.1",
			"1.35.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.35.0",
			">= 1.36.x": "registry.k8s.io/provider-os/cinder-csi-plugin:v1.36.0",
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
			"1.34.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.34.0",
			"1.35.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.35.1",
			">= 1.36.x": "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.36.0",
		},

		// vSphere CSI
		VsphereCSIDriver:             {"*": "registry.k8s.io/csi-vsphere/driver:v3.7.1"},
		VsphereCSISyncer:             {"*": "registry.k8s.io/csi-vsphere/syncer:v3.7.1"},
		VsphereCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.9.0"},
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
		GCPCCM: {"*": "registry.k8s.io/cloud-provider-gcp/cloud-controller-manager:v36.0.7"},

		// GCP Compute Persistent Disk CSI
		// see: https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/blob/master/deploy/kubernetes/images/stable-master/image.yaml
		GCPComputeCSIDriver:              {"*": "registry.k8s.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.26.0"},
		GCPComputeCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.0.0"},
		GCPComputeCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.8.1"},
		GCPComputeCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.0.0"},
		GCPComputeCSISnapshotter:         {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.2.1"},
		GCPComputeCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.13.0"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.19.4@sha256:2eb67991eaa9368ba199c2fac2c573cb0ffdeb79184533344f42fc9a7ff6af3c"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.19.4@sha256:1aa2b62735e7d8ab49ee840ae59c346932024c88901579121395c1271b435f71"},
		CiliumEnvoy:    {"*": "quay.io/cilium/cilium-envoy:v1.36.6-1778235340-b87d1e32f522b33bd51701c6476d199326f01496@sha256:71d4fa0ec45e8d546dbd5604e169dc77fe92be63b799313bff031d00d89762e3"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.19.4@sha256:59af8c0d561e560c2a042e7600a3496bc0367df8fbf868aa68d5834c8ec1a431"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.13.5@sha256:f7d514fc54d784ed6df9d58cf0e97648b143f92b766dd1780ed3fc845bd4c516"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.13.5@sha256:fac0c300ae119274edca11fd89b1ad23c788792d8bc4ea2ba631c709e8d3c688"},
		CiliumCertGen:   {"*": "quay.io/cilium/certgen:v0.4.3@sha256:d63d1cb3ee6db2167cb1ca9e7e31f30b6197846fcf42505d8f59e2f090a42c86"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.34.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.34.3",
			"1.35.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.35.0",
			">= 1.36.x": "registry.k8s.io/autoscaling/cluster-autoscaler:v1.35.0",
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
		KubeVirtCSI:                    {"*": "quay.io/kubermatic/kubevirt-csi-driver:v0.4.5"},
		KubeVirtCSINodeDriverRegistrar: {"*": "quay.io/openshift/origin-csi-node-driver-registrar:4.20.0"},
		KubeVirtCSILivenessProbe:       {"*": "quay.io/openshift/origin-csi-livenessprobe:4.20.0"},
		KubeVirtCSIProvisioner:         {"*": "quay.io/openshift/origin-csi-external-provisioner:4.20.0"},
		KubeVirtCSIAttacher:            {"*": "quay.io/openshift/origin-csi-external-attacher:4.20.0"},

		// Backup Restic
		BackupResticSnapshotter: {"*": "registry.k8s.io/etcd:3.5.16-0"},
		BackupResticUploader:    {"*": "ghcr.io/restic/restic:0.18.1"},

		// Unattended upgrades
		UUApline: {"*": "docker.io/library/alpine:3.23"},
		UUFluo:   {"*": "ghcr.io/flatcar/flatcar-linux-update-operator:v0.10.0-rc1"},
	}
}

func allResources() map[Resource]map[string]string {
	ret := map[Resource]map[string]string{}
	maps.Copy(ret, baseResources())
	maps.Copy(ret, optionalResources())

	return ret
}

// providerToResources maps a cloud provider name (as returned by
// CloudProviderSpec.Name()) to the set of optional Resource constants that are
// deployed for that provider.  Shared infrastructure images such as the CSI
// snapshot controller are included in every provider that relies on them.
func providerToResources() map[string][]Resource {
	return map[string][]Resource{
		"aws": {
			AwsCCM,
			CSISnapshotController,
			CSISnapshotWebhook,
			AwsEbsCSI,
			AwsEbsCSIAttacher,
			AwsEbsCSILivenessProbe,
			AwsEbsCSINodeDriverRegistrar,
			AwsEbsCSIProvisioner,
			AwsEbsCSIResizer,
			AwsEbsCSISnapshotter,
		},
		"azure": {
			AzureCCM,
			AzureCNM,
			CSISnapshotController,
			CSISnapshotWebhook,
			AzureFileCSI,
			AzureFileCSIAttacher,
			AzureFileCSILivenessProbe,
			AzureFileCSINodeDriverRegistar,
			AzureFileCSIProvisioner,
			AzureFileCSIResizer,
			AzureFileCSISnapshotter,
			AzureDiskCSI,
			AzureDiskCSIAttacher,
			AzureDiskCSILivenessProbe,
			AzureDiskCSINodeDriverRegistar,
			AzureDiskCSIProvisioner,
			AzureDiskCSIResizer,
			AzureDiskCSISnapshotter,
		},
		"digitalocean": {
			DigitaloceanCCM,
			CSISnapshotController,
			CSISnapshotWebhook,
			DigitalOceanCSI,
			DigitalOceanCSIAlpine,
			DigitalOceanCSIAttacher,
			DigitalOceanCSINodeDriverRegistar,
			DigitalOceanCSIProvisioner,
			DigitalOceanCSIResizer,
			DigitalOceanCSISnapshotter,
		},
		"gce": {
			GCPCCM,
			CSISnapshotController,
			CSISnapshotWebhook,
			GCPComputeCSIDriver,
			GCPComputeCSIProvisioner,
			GCPComputeCSIAttacher,
			GCPComputeCSIResizer,
			GCPComputeCSISnapshotter,
			GCPComputeCSINodeDriverRegistrar,
		},
		"hetzner": {
			HetznerCCM,
			HetznerCSI,
			HetznerCSIAttacher,
			HetznerCSIResizer,
			HetznerCSIProvisioner,
			HetznerCSILivenessProbe,
			HetznerCSINodeDriverRegistar,
		},
		"kubevirt": {
			KubeVirtCCM,
			KubeVirtCSI,
			KubeVirtCSINodeDriverRegistrar,
			KubeVirtCSILivenessProbe,
			KubeVirtCSIProvisioner,
			KubeVirtCSIAttacher,
		},
		"nutanix": {
			NutanixCCM,
			NutanixCSI,
			NutanixCSILivenessProbe,
			NutanixCSIExternalHealthMonitor,
			NutanixCSIAttacher,
			NutanixCSIPrecheck,
			NutanixCSIProvisioner,
			NutanixCSIRegistrar,
			NutanixCSIResizer,
			NutanixCSISnapshotter,
		},
		"openstack": {
			OpenstackCCM,
			CSISnapshotController,
			CSISnapshotWebhook,
			OpenstackCSI,
			OpenstackCSINodeDriverRegistar,
			OpenstackCSILivenessProbe,
			OpenstackCSIAttacher,
			OpenstackCSIProvisioner,
			OpenstackCSIResizer,
			OpenstackCSISnapshotter,
		},
		"equinixmetal": {
			EquinixMetalCCM,
		},
		"vmwareCloudDirector": {
			VMwareCloudDirectorCSI,
			VMwareCloudDirectorCSIAttacher,
			VMwareCloudDirectorCSIProvisioner,
			VMwareCloudDirectorCSIResizer,
			VMwareCloudDirectorCSINodeDriverRegistrar,
		},
		"vsphere": {
			VsphereCCM,
			CSISnapshotController,
			CSISnapshotWebhook,
			VsphereCSIDriver,
			VsphereCSISyncer,
			VsphereCSIAttacher,
			VsphereCSILivenessProbe,
			VsphereCSINodeDriverRegistar,
			VsphereCSIProvisioner,
			VsphereCSIResizer,
			VsphereCSISnapshotter,
		},
		"none": {},
	}
}

// SupportedProviders returns the sorted list of cloud provider names that can
// be used with --provider.
func SupportedProviders() []string {
	providers := slices.Collect(maps.Keys(providerToResources()))
	sort.Strings(providers)

	return providers
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
	ListFilterOptional
)

func (r *Resolver) List(lf ListFilter) []string {
	var list []string

	fn := allResources
	switch lf {
	case ListFilterBase:
		fn = baseResources
	case ListFilterOptional:
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

// ListForProvider returns the sorted list of images required by the given cloud
// provider.  The provider name must match one of the values returned by
// SupportedProviders() (i.e. the string returned by CloudProviderSpec.Name()).
// Shared infra images (e.g. CSISnapshotController) are included for every
// provider that uses them.
func (r *Resolver) ListForProvider(provider string) ([]string, error) {
	resources, ok := providerToResources()[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q, must be one of: %s",
			provider, strings.Join(SupportedProviders(), ", "))
	}

	// deduplicate with a map (multiple resources can resolve to the same image)
	listMap := make(map[string]bool)
	for _, res := range resources {
		img := r.Get(res)
		if img != "" {
			listMap[img] = true
		}
	}

	list := slices.Collect(maps.Keys(listMap))
	if list == nil {
		list = []string{}
	}
	sort.Strings(list)

	return list, nil
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
