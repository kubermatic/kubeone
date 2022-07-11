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
	HubbleProxy
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

	// CCMs and CSI plugins
	DigitaloceanCCM
	HetznerCCM
	HetznerCSI
	OpenstackCCM
	EquinixMetalCCM
	VsphereCCM

	// VMwareCloud Director CSI
	VMwareCloudDirectorCSI

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
		CalicoCNI:         {"*": "quay.io/calico/cni:v3.22.2"},
		CalicoController:  {"*": "quay.io/calico/kube-controllers:v3.22.2"},
		CalicoNode:        {"*": "quay.io/calico/node:v3.22.2"},
		DNSNodeCache:      {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.21.1"},
		Flannel:           {"*": "quay.io/coreos/flannel:v0.15.1"},
		MachineController: {"*": "quay.io/kubermatic/machine-controller:v1.52.0"},
		MetricsServer:     {"*": "k8s.gcr.io/metrics-server/metrics-server:v0.6.1"},
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
			"1.20.x":    "k8s.gcr.io/provider-aws/cloud-controller-manager:v1.20.1",
			"1.21.x":    "k8s.gcr.io/provider-aws/cloud-controller-manager:v1.21.1",
			"1.22.x":    "k8s.gcr.io/provider-aws/cloud-controller-manager:v1.22.2",
			"1.23.x":    "k8s.gcr.io/provider-aws/cloud-controller-manager:v1.23.1",
			">= 1.24.0": "k8s.gcr.io/provider-aws/cloud-controller-manager:v1.24.0",
		},

		// Azure CCM
		AzureCCM: {
			"1.20.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v0.7.21",
			"1.21.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.0.18",
			"1.22.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.1.14",
			"1.23.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.23.11",
			">= 1.24.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.0",
		},
		AzureCNM: {
			"1.20.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v0.7.21",
			"1.21.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.0.18",
			"1.22.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.1.14",
			"1.23.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.23.11",
			">= 1.24.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.0",
		},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.6.2"},
		AwsEbsCSIAttacher:            {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.1.0"},
		AwsEbsCSILivenessProbe:       {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.4.0"},
		AwsEbsCSINodeDriverRegistrar: {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.3.0"},
		AwsEbsCSIProvisioner:         {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v2.2.2"},
		AwsEbsCSIResizer:             {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.1.0"},
		AwsEbsCSISnapshotter:         {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v4.2.1"},
		AwsEbsCSISnapshotController:  {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v4.2.1"},

		// AzureFile CSI driver
		AzureFileCSI:                      {"*": "mcr.microsoft.com/k8s/csi/azurefile-csi:v1.18.0"},
		AzureFileCSIAttacher:              {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v3.4.0"},
		AzureFileCSILivenessProbe:         {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.6.0"},
		AzureFileCSINodeDriverRegistar:    {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.5.0"},
		AzureFileCSIProvisioner:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v3.1.0"},
		AzureFileCSIResizer:               {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.4.0"},
		AzureFileCSISnapshotter:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v5.0.1"},
		AzureFileCSISnapshotterController: {"*": "mcr.microsoft.com/oss/kubernetes-csi/snapshot-controller:v5.0.1"},

		// AzureDisk CSI driver
		AzureDiskCSI:                      {"*": "mcr.microsoft.com/k8s/csi/azuredisk-csi:v1.18.0"},
		AzureDiskCSIAttacher:              {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v3.4.0"},
		AzureDiskCSILivenessProbe:         {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.6.0"},
		AzureDiskCSINodeDriverRegistar:    {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.5.0"},
		AzureDiskCSIProvisioner:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v3.1.0"},
		AzureDiskCSIResizer:               {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.4.0"},
		AzureDiskCSISnapshotter:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v5.0.1"},
		AzureDiskCSISnapshotterController: {"*": "mcr.microsoft.com/oss/kubernetes-csi/snapshot-controller:v5.0.1"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.37"},

		DigitalOceanCSI:                          {"*": "docker.io/digitalocean/do-csi-plugin:v4.0.0"},
		DigitalOceanCSIAlpine:                    {"*": "docker.io/alpine:3"},
		DigitalOceanCSIAttacher:                  {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.3.0"},
		DigitalOceanCSINodeDriverRegistar:        {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.4.0"},
		DigitalOceanCSIProvisioner:               {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.0.0"},
		DigitalOceanCSIResizer:                   {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.3.0"},
		DigitalOceanCSISnapshotController:        {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v5.0.0"},
		DigitalOceanCSISnapshotValidationWebhook: {"*": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v5.0.0"},
		DigitalOceanCSISnapshotter:               {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v5.0.0"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.12.1"},

		// Hetzner CSI
		HetznerCSI: {"*": "docker.io/hetznercloud/hcloud-csi-driver:1.6.0"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.20.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.20.2",
			"1.21.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.21.0",
			"1.22.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.22.0",
			"1.23.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.23.1",
			">= 1.24.0": "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.24.0",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.20.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.20.3",
			"1.21.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.21.0",
			"1.22.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.22.0",
			"1.23.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.23.0",
			">= 1.24.0": "docker.io/k8scloudprovider/cinder-csi-plugin:v1.24.0",
		},
		OpenstackCSINodeDriverRegistar: {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.0"},
		OpenstackCSILivenessProbe:      {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.6.0"},
		OpenstackCSIAttacher:           {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.4.0"},
		OpenstackCSIProvisioner:        {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.1.0"},
		OpenstackCSIResizer:            {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.4.0"},
		OpenstackCSISnapshotter:        {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v5.0.1"},

		// Equinix Metal CCM
		EquinixMetalCCM: {"*": "docker.io/equinix/cloud-provider-equinix-metal:v3.4.2"},

		// vSphere CCM
		VsphereCCM: {
			"1.20.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.20.1",
			"1.21.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.21.3",
			"1.22.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.22.6",
			">= 1.23.0": "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.23.0",
		},

		// VMware Cloud Director CSI
		VMwareCloudDirectorCSI: {"*": "projects.registry.vmware.com/vmware-cloud-director/cloud-director-named-disk-csi-driver:1.2.0.latest"},

		// vSphere CSI
		VsphereCSIDriver:                    {"*": "gcr.io/cloud-provider-vsphere/csi/release/driver:v2.5.1"},
		VsphereCSISyncer:                    {"*": "gcr.io/cloud-provider-vsphere/csi/release/syncer:v2.5.1"},
		VsphereCSIAttacher:                  {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.4.0"},
		VsphereCSILivenessProbe:             {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.6.0"},
		VsphereCSINodeDriverRegistar:        {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.0"},
		VsphereCSIProvisioner:               {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.1.0"},
		VsphereCSIResizer:                   {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.4.0"},
		VsphereCSISnapshotter:               {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v5.0.1"},
		VsphereCSISnapshotController:        {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v5.0.1"},
		VsphereCSISnapshotValidationWebhook: {"*": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v5.0.1"},

		// Nutanix CSI
		NutanixCSI:                          {"*": "quay.io/karbon/ntnx-csi:v2.5.0"},
		NutanixCSILivenessProbe:             {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.3.0"},
		NutanixCSIProvisioner:               {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v2.2.2"},
		NutanixCSIRegistrar:                 {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.2.0"},
		NutanixCSIResizer:                   {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.2.0"},
		NutanixCSISnapshotter:               {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v4.2.1"},
		NutanixCSISnapshotController:        {">= 1.20.0": "k8s.gcr.io/sig-storage/snapshot-controller:v4.2.1"},
		NutanixCSISnapshotValidationWebhook: {">= 1.20.0": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v4.2.1"},

		// GCP Compute Persistent Disk CSI
		GCPComputeCSIDriver: {
			"1.21.x":    "gke.gcr.io/gcp-compute-persistent-disk-csi-driver:v1.4.0",
			">= 1.22.0": "k8s.gcr.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.7.1",
		},
		GCPComputeCSIProvisioner:               {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.1.0"},
		GCPComputeCSIAttacher:                  {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.4.0"},
		GCPComputeCSIResizer:                   {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.4.0"},
		GCPComputeCSISnapshotter:               {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v4.0.1"},
		GCPComputeCSISnapshotController:        {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v4.0.1"},
		GCPComputeCSISnapshotValidationWebhook: {"*": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v4.0.1"},
		GCPComputeCSINodeDriverRegistrar:       {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.0"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.11.5@sha256:79e66c3c2677e9ecc3fd5b2ed8e4ea7e49cf99ed6ee181f2ef43400c4db5eef0"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.11.5@sha256:8ace281328b27d4216218c604d720b9a63a8aec2bd1996057c79ab0168f9d6d8"},

		// Calico VXLAN
		CalicoVXLANCNI:        {"*": "quay.io/calico/cni:v3.22.2"},
		CalicoVXLANController: {"*": "quay.io/calico/kube-controllers:v3.22.2"},
		CalicoVXLANNode:       {"*": "quay.io/calico/node:v3.22.2"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.11.5@sha256:8498f27a9c85ff74e56e18cfce4f0ccfae6f55d4134d708d364d273f3043f817"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.8.5@sha256:4eaca1ec1741043cfba6066a165b3bf251590cf4ac66371c4f63fbed2224ebb4"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.8.5@sha256:2bce50cf6c32719d072706f7ceccad654bfa907b2745a496da99610776fe31ed"},
		HubbleProxy:     {"*": "docker.io/envoyproxy/envoy:v1.18.4@sha256:e5c2bb2870d0e59ce917a5100311813b4ede96ce4eb0c6bfa879e3fbe3e83935"},
		CiliumCertGen:   {"*": "quay.io/cilium/certgen:v0.1.5"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.20.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.20.2",
			"1.21.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.21.2",
			"1.22.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.22.2",
			"1.23.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.0",
			">= 1.24.0": "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.24.0",
		},
		// operating-system-manager addon
		OperatingSystemManager: {"*": "quay.io/kubermatic/operating-system-manager:v0.5.0"},
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
