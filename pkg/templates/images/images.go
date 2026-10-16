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
	CalicoThirdPartyPlugins
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
	NutanixCSIAuthAgent

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
		CalicoCNI:               {"*": "quay.io/calico/calico:v3.33.0"},
		CalicoController:        {"*": "quay.io/calico/calico:v3.33.0"},
		CalicoNode:              {"*": "quay.io/calico/node:v3.33.0"},
		CalicoThirdPartyPlugins: {"*": "quay.io/calico/third-party-cni-plugins:v3.33.0"},
		DNSNodeCache:            {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.27.1"},
		Flannel:                 {"*": "docker.io/flannel/flannel:v0.24.4"},
		MachineController:       {"*": "quay.io/kubermatic/machine-controller:86813a09db8bb425ad6b2124147c2e78280e68cb"},
		MetricsServer:           {"*": "registry.k8s.io/metrics-server/metrics-server:v0.9.0"},
		OperatingSystemManager:  {"*": "quay.io/kubermatic/operating-system-manager:7f2bf34da9bae3f5c6d90170df355680ba8c9321"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		AwsCCM: {
			"1.35.x":   "registry.k8s.io/provider-aws/cloud-controller-manager:v1.35.2",
			"1.36.x":   "registry.k8s.io/provider-aws/cloud-controller-manager:v1.36.1",
			">=1.37.x": "registry.k8s.io/provider-aws/cloud-controller-manager:v1.37.0",
		},

		CSISnapshotController: {"*": "registry.k8s.io/sig-storage/snapshot-controller:v8.7.0"},
		CSISnapshotWebhook:    {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v8.1.1"},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.67.0"},
		AwsEbsCSIAttacher:            {"*": "public.ecr.aws/csi-components/csi-attacher:v4.13.0-eksbuild.1"},
		AwsEbsCSILivenessProbe:       {"*": "public.ecr.aws/csi-components/livenessprobe:v2.20.0-eksbuild.1"},
		AwsEbsCSINodeDriverRegistrar: {"*": "public.ecr.aws/csi-components/csi-node-driver-registrar:v2.18.0-eksbuild.1"},
		AwsEbsCSIProvisioner:         {"*": "public.ecr.aws/csi-components/csi-provisioner:v6.4.0-eksbuild.1"},
		AwsEbsCSIResizer:             {"*": "public.ecr.aws/csi-components/csi-resizer:v2.3.0-eksbuild.1"},
		AwsEbsCSISnapshotter:         {"*": "public.ecr.aws/csi-components/csi-snapshotter:v8.6.0-eksbuild.10"},

		// Azure CCM
		AzureCCM: {
			"1.35.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.35.9",
			"1.36.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.36.7",
			">= 1.37.x": "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.37.0",
		},
		AzureCNM: {
			"1.35.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.35.9",
			"1.36.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.36.7",
			">= 1.37.x": "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.37.0",
		},

		// AzureFile CSI driver
		AzureFileCSI:                   {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/azurefile-csi:v1.35.10"},
		AzureFileCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/livenessprobe:v2.19.0"},
		AzureFileCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-node-driver-registrar:v2.17.0"},
		AzureFileCSIProvisioner:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-provisioner:v6.3.0"},
		AzureFileCSIResizer:            {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-resizer:v2.2.0"},
		AzureFileCSISnapshotter:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-snapshotter:v8.5.0"},

		// AzureDisk CSI driver5
		AzureDiskCSI:                   {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/azuredisk-csi:v1.36.0"},
		AzureDiskCSIAttacher:           {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-attacher:v4.12.0"},
		AzureDiskCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/livenessprobe:v2.19.0"},
		AzureDiskCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-node-driver-registrar:v2.17.0"},
		AzureDiskCSIProvisioner:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-provisioner:v6.3.0"},
		AzureDiskCSIResizer:            {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-resizer:v2.2.1"},
		AzureDiskCSISnapshotter:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-snapshotter:v8.6.0"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.71"},

		// DigitalOcean CSI
		DigitalOceanCSI:                   {"*": "docker.io/digitalocean/do-csi-plugin:v4.19.0"},
		DigitalOceanCSIAlpine:             {"*": "docker.io/alpine:3"},
		DigitalOceanCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.13.0"},
		DigitalOceanCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.18.0"},
		DigitalOceanCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.3.0"},
		DigitalOceanCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.2.1"},
		DigitalOceanCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.6.0"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.39.0"},

		// Hetzner CSI
		HetznerCSI:                   {"*": "docker.io/hetznercloud/hcloud-csi-driver:v2.24.0"},
		HetznerCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.13.0"},
		HetznerCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.2.1"},
		HetznerCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.3.0"},
		HetznerCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.20.0"},
		HetznerCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.18.0"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.35.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.35.0",
			"1.36.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.36.0",
			">= 1.37.x": "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.36.0",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.35.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.35.0",
			"1.36.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.36.0",
			">= 1.37.x": "registry.k8s.io/provider-os/cinder-csi-plugin:v1.36.0",
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
			"1.35.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.35.1",
			"1.36.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.36.0",
			">= 1.37.x": "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.37.0",
		},

		// vSphere CSI
		VsphereCSIDriver:             {"*": "registry.k8s.io/csi-vsphere/driver:v3.8.0"},
		VsphereCSISyncer:             {"*": "registry.k8s.io/csi-vsphere/syncer:v3.8.0"},
		VsphereCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.10.0"},
		VsphereCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.17.0"},
		VsphereCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.14.0"},
		VsphereCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v4.0.1"},
		VsphereCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.0.0"},
		VsphereCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.4.0"},

		// Nutanix CCM
		NutanixCCM: {"*": "ghcr.io/nutanix-cloud-native/cloud-provider-nutanix/controller:v0.6.3"},

		// Nutanix CSI
		NutanixCSI:                      {"*": "docker.io/nutanix/ntnx-csi:3.9.0"},
		NutanixCSILivenessProbe:         {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.19.0"},
		NutanixCSIExternalHealthMonitor: {"*": "registry.k8s.io/sig-storage/csi-external-health-monitor-controller:v0.18.0"},
		NutanixCSIAttacher:              {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.12.0"},
		NutanixCSIProvisioner:           {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.3.0"},
		NutanixCSIRegistrar:             {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.17.0"},
		NutanixCSIPrecheck:              {"*": "docker.io/nutanix/ntnx-csi-precheck:3.9.0"},
		NutanixCSIResizer:               {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.2.1"},
		NutanixCSISnapshotter:           {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.6.0"},
		NutanixCSIAuthAgent:             {"*": "docker.io/nutanix/ntnx-csi-authagent:3.9.0"},

		// GCP CCM
		GCPCCM: {"*": "registry.k8s.io/cloud-provider-gcp/cloud-controller-manager:v37.0.1"},

		// GCP Compute Persistent Disk CSI
		// see: https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/blob/master/deploy/kubernetes/images/stable-master/image.yaml
		GCPComputeCSIDriver:              {"*": "registry.k8s.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.27.1"},
		GCPComputeCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.0.0"},
		GCPComputeCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.8.1"},
		GCPComputeCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.0.0"},
		GCPComputeCSISnapshotter:         {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.2.1"},
		GCPComputeCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.13.0"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.20.2@sha256:2939231d0d3e3ebddcd80fffa168b7ddcc78fdf0dc864d1c8c126ff523c54f01"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.20.2@sha256:64d8798350e8569b8e7622563fed6e44dce2625f311e4651b774816516c744fc"},
		CiliumEnvoy:    {"*": "quay.io/cilium/cilium-envoy:v1.37.6-1789133542-cbec91f666af0bf742da986d43832932dbb26b82@sha256:af7382699576b9e65e9184efa52eeca0b58aea70ad6e511bf260c91d9f740463"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.20.2@sha256:d309c977870e9dbede7122a10eee09a4c9d66685e8d52af62d9c3c113f815b0f"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.13.6@sha256:049a80a03585c043d0c3121bdc518c72b0fd42bae07e4bcce0fdb8f1e88041d0"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.13.6@sha256:83b3fc763f3d49948c306bf49b08277383e195904b9ca5c0c0022b2112628787"},
		CiliumCertGen:   {"*": "quay.io/cilium/certgen:v0.4.11@sha256:14d29a176fc96d15b154559ce1a5baf01367c10baabfa7bb8028b72f087a56f0"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.35.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.35.2",
			"1.36.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.36.1",
			">= 1.37.x": "registry.k8s.io/autoscaling/cluster-autoscaler:v1.36.1",
		},

		// CSI Vault Secret Provider
		CSIVaultSecretProvider: {"*": "docker.io/hashicorp/vault-csi-provider:1.1.0"},

		// CSI Secrets Driver
		SecretStoreCSIDriverNodeRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		SecretStoreCSIDriver:              {"*": "registry.k8s.io/csi-secrets-store/driver:v1.2.1"},
		SecretStoreCSIDriverLivenessProbe: {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.7.0"},
		SecretStoreCSIDriverCRDs:          {"*": "registry.k8s.io/csi-secrets-store/driver-crds:v1.2.1"},

		// KubeVirt's CCM
		KubeVirtCCM: {"*": "quay.io/kubevirt/kubevirt-cloud-controller-manager:v0.6.0"},

		// KubeVirt CSI
		KubeVirtCSI:                    {"*": "quay.io/kubermatic/kubevirt-csi-driver:v0.4.5"},
		KubeVirtCSINodeDriverRegistrar: {"*": "quay.io/openshift/origin-csi-node-driver-registrar:4.20.0"},
		KubeVirtCSILivenessProbe:       {"*": "quay.io/openshift/origin-csi-livenessprobe:4.20.0"},
		KubeVirtCSIProvisioner:         {"*": "quay.io/openshift/origin-csi-external-provisioner:4.20.0"},
		KubeVirtCSIAttacher:            {"*": "quay.io/openshift/origin-csi-external-attacher:4.20.0"},

		// Backup Restic
		BackupResticSnapshotter: {"*": "registry.k8s.io/etcd:3.5.32-0"},
		BackupResticUploader:    {"*": "ghcr.io/restic/restic:0.19.1"},

		// Unattended upgrades
		UUApline: {"*": "docker.io/library/alpine:3.23"},
		UUFluo:   {"*": "ghcr.io/flatcar/flatcar-linux-update-operator:v0.10.0"},
	}
}

func allResources() map[Resource]map[string]string {
	ret := map[Resource]map[string]string{}
	maps.Copy(ret, baseResources())
	maps.Copy(ret, optionalResources())

	return ret
}

// ProviderNone is a special provider value that disables filtering images by
// the cloud provider.
const ProviderNone = "none"

// providerResources maps cloud provider names (as returned by
// kubeoneapi.CloudProviderSpec.CloudProviderName) to the provider-specific
// resources (CCM/CSI). Resources not listed here are provider-neutral.
func providerResources() map[string][]Resource {
	return map[string][]Resource{
		"aws": {
			AwsCCM,
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
			AzureFileCSI,
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
			DigitalOceanCSI,
			DigitalOceanCSIAlpine,
			DigitalOceanCSIAttacher,
			DigitalOceanCSINodeDriverRegistar,
			DigitalOceanCSIProvisioner,
			DigitalOceanCSIResizer,
			DigitalOceanCSISnapshotter,
		},
		"equinixmetal": {
			EquinixMetalCCM,
		},
		"gce": {
			GCPCCM,
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
			NutanixCSILivenessProbe,
			NutanixCSIExternalHealthMonitor,
			NutanixCSIAttacher,
			NutanixCSIPrecheck,
			NutanixCSI,
			NutanixCSIProvisioner,
			NutanixCSIRegistrar,
			NutanixCSIResizer,
			NutanixCSISnapshotter,
			NutanixCSIAuthAgent,
		},
		"openstack": {
			OpenstackCCM,
			OpenstackCSI,
			OpenstackCSINodeDriverRegistar,
			OpenstackCSILivenessProbe,
			OpenstackCSIAttacher,
			OpenstackCSIProvisioner,
			OpenstackCSIResizer,
			OpenstackCSISnapshotter,
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
			VsphereCSIDriver,
			VsphereCSISyncer,
			VsphereCSIAttacher,
			VsphereCSILivenessProbe,
			VsphereCSINodeDriverRegistar,
			VsphereCSIProvisioner,
			VsphereCSIResizer,
			VsphereCSISnapshotter,
		},
	}
}

// Providers returns sorted list of cloud provider names that can be used to
// filter images.
func Providers() []string {
	return slices.Sorted(maps.Keys(providerResources()))
}

// ValidateProviders checks that given providers are known. Empty list or
// a single ProviderNone value means no filtering.
func ValidateProviders(providers []string) error {
	valid := Providers()

	for _, provider := range providers {
		if provider == ProviderNone {
			if len(providers) > 1 {
				return fail.RuntimeError{
					Op:  "checking provider flag",
					Err: fmt.Errorf("--provider %q can't be combined with other providers", ProviderNone),
				}
			}

			continue
		}

		if !slices.Contains(valid, provider) {
			return fail.RuntimeError{
				Op:  "checking provider flag",
				Err: fmt.Errorf("unknown provider %q, --provider can be only %q or any of %v", provider, ProviderNone, valid),
			}
		}
	}

	return nil
}

type listOptions struct {
	providers []string
}

// skip reports whether the resource belongs to a cloud provider that was not
// selected.
func (lo listOptions) skip(res Resource) bool {
	if len(lo.providers) == 0 || slices.Contains(lo.providers, ProviderNone) {
		return false
	}

	for provider, resources := range providerResources() {
		if slices.Contains(resources, res) {
			return !slices.Contains(lo.providers, provider)
		}
	}

	return false
}

type ListOpt func(*listOptions)

// WithProviders limits the provider-specific images (CCM/CSI) to the given
// cloud providers. Provider-neutral images are always listed.
func WithProviders(providers ...string) ListOpt {
	return func(lo *listOptions) {
		lo.providers = providers
	}
}

func newListOptions(opts []ListOpt) listOptions {
	var lo listOptions
	for _, opt := range opts {
		opt(&lo)
	}

	return lo
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

func (r *Resolver) List(lf ListFilter, opts ...ListOpt) []string {
	var list []string

	lo := newListOptions(opts)

	fn := allResources
	switch lf {
	case ListFilterBase:
		fn = baseResources
	case ListFilterOptional:
		fn = optionalResources
	case ListFilterNone:
	}

	for res := range fn() {
		if lo.skip(res) {
			continue
		}

		img := r.Get(res)
		if img != "" {
			list = append(list, img)
		}
	}

	sort.Strings(list)

	return list
}

func (r *Resolver) ListAll(opts ...ListOpt) []string {
	resources := allResources()
	lo := newListOptions(opts)

	// create a bool map, to deduplicate the images
	listMap := make(map[string]bool)
	for res := range resources {
		if lo.skip(res) {
			continue
		}

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
