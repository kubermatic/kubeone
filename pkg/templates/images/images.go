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
	OpenstackCSISnapshotController
	OpenstackCSISnapshotWebhook

	// CCMs and CSI plugins
	DigitaloceanCCM
	HetznerCCM
	HetznerCSI
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
		CalicoCNI:              {"*": "quay.io/calico/cni:v3.23.3"},
		CalicoController:       {"*": "quay.io/calico/kube-controllers:v3.23.3"},
		CalicoNode:             {"*": "quay.io/calico/node:v3.23.3"},
		DNSNodeCache:           {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.22.13"},
		Flannel:                {"*": "quay.io/coreos/flannel:v0.15.1"},
		MachineController:      {"*": "quay.io/kubermatic/machine-controller:v1.55.0"},
		MetricsServer:          {"*": "k8s.gcr.io/metrics-server/metrics-server:v0.6.1"},
		OperatingSystemManager: {"*": "quay.io/kubermatic/operating-system-manager:v1.1.1"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		// General CSI images (could be used for all providers)
		CSINodeDriverRegistar: {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.3.0"},
		CSILivenessProbe:      {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.4.0"},
		CSIAttacher:           {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.3.0"},
		CSIProvisioner:        {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v2.2.2"},
		CSIResizer:            {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.3.0"},
		CSISnapshotter:        {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v4.2.0"},

		AwsCCM: {
			"1.23.x":    "k8s.gcr.io/provider-aws/cloud-controller-manager:v1.23.2",
			"1.24.x":    "k8s.gcr.io/provider-aws/cloud-controller-manager:v1.24.1",
			">= 1.25.0": "k8s.gcr.io/provider-aws/cloud-controller-manager:v1.25.1",
		},

		// Azure CCM
		AzureCCM: {
			"1.23.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.23.21",
			"1.24.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.8",
			">= 1.25.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.3",
		},
		AzureCNM: {
			"1.23.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.23.21",
			"1.24.x":    "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.8",
			">= 1.25.0": "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.3",
		},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.12.1"},
		AwsEbsCSIAttacher:            {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.4.0"},
		AwsEbsCSILivenessProbe:       {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.6.0"},
		AwsEbsCSINodeDriverRegistrar: {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		AwsEbsCSIProvisioner:         {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.1.0"},
		AwsEbsCSIResizer:             {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.4.0"},
		AwsEbsCSISnapshotter:         {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v6.0.1"},
		AwsEbsCSISnapshotController:  {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v6.0.1"},

		// AzureFile CSI driver
		AzureFileCSI:                      {"*": "mcr.microsoft.com/oss/kubernetes-csi/azurefile-csi:v1.22.0"},
		AzureFileCSIAttacher:              {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v3.5.0"},
		AzureFileCSILivenessProbe:         {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.7.0"},
		AzureFileCSINodeDriverRegistar:    {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.5.1"},
		AzureFileCSIProvisioner:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v3.2.0"},
		AzureFileCSIResizer:               {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.5.0"},
		AzureFileCSISnapshotter:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v5.0.1"},
		AzureFileCSISnapshotterController: {"*": "mcr.microsoft.com/oss/kubernetes-csi/snapshot-controller:v5.0.1"},

		// AzureDisk CSI driver
		AzureDiskCSI:                      {"*": "mcr.microsoft.com/oss/kubernetes-csi/azuredisk-csi:v1.23.0"},
		AzureDiskCSIAttacher:              {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-attacher:v3.5.0"},
		AzureDiskCSILivenessProbe:         {"*": "mcr.microsoft.com/oss/kubernetes-csi/livenessprobe:v2.7.0"},
		AzureDiskCSINodeDriverRegistar:    {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-node-driver-registrar:v2.5.1"},
		AzureDiskCSIProvisioner:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-provisioner:v3.2.0"},
		AzureDiskCSIResizer:               {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-resizer:v1.5.0"},
		AzureDiskCSISnapshotter:           {"*": "mcr.microsoft.com/oss/kubernetes-csi/csi-snapshotter:v5.0.1"},
		AzureDiskCSISnapshotterController: {"*": "mcr.microsoft.com/oss/kubernetes-csi/snapshot-controller:v5.0.1"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.40"},

		DigitalOceanCSI:                          {"*": "docker.io/digitalocean/do-csi-plugin:v4.4.1"},
		DigitalOceanCSIAlpine:                    {"*": "docker.io/alpine:3"},
		DigitalOceanCSIAttacher:                  {"*": "k8s.gcr.io/sig-storage/csi-attacher:v4.0.0"},
		DigitalOceanCSINodeDriverRegistar:        {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.6.0"},
		DigitalOceanCSIProvisioner:               {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.3.0"},
		DigitalOceanCSIResizer:                   {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.6.0"},
		DigitalOceanCSISnapshotController:        {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v6.1.0"},
		DigitalOceanCSISnapshotValidationWebhook: {"*": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v6.1.0"},
		DigitalOceanCSISnapshotter:               {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v6.1.0"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.13.2"},

		// Hetzner CSI
		HetznerCSI: {"*": "docker.io/hetznercloud/hcloud-csi-driver:1.6.0"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.23.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.23.4",
			"1.24.x":    "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.24.5",
			">= 1.25.0": "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.25.3",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.23.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.23.4",
			"1.24.x":    "docker.io/k8scloudprovider/cinder-csi-plugin:v1.24.5",
			">= 1.25.0": "docker.io/k8scloudprovider/cinder-csi-plugin:v1.25.3",
		},
		OpenstackCSINodeDriverRegistar: {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		OpenstackCSILivenessProbe:      {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.7.0"},
		OpenstackCSIAttacher:           {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.4.0"},
		OpenstackCSIProvisioner:        {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.1.0"},
		OpenstackCSIResizer:            {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.4.0"},
		OpenstackCSISnapshotter:        {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v6.0.1"},
		OpenstackCSISnapshotController: {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v6.0.1"},
		OpenstackCSISnapshotWebhook:    {"*": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v6.0.1"},

		// Equinix Metal CCM
		EquinixMetalCCM: {"*": "ghcr.io/equinix/cloud-provider-equinix-metal:v3.5.0"},

		// vSphere CCM
		VsphereCCM: {
			"1.23.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.23.2",
			"1.24.x":    "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.24.2",
			">= 1.25.0": "gcr.io/cloud-provider-vsphere/cpi/release/manager:v1.25.0",
		},

		// VMware Cloud Director CSI
		VMwareCloudDirectorCSI: {"*": "projects.registry.vmware.com/vmware-cloud-director/cloud-director-named-disk-csi-driver:1.2.0.latest"},

		// vSphere CSI
		VsphereCSIDriver:                    {"*": "gcr.io/cloud-provider-vsphere/csi/release/driver:v2.7.0"},
		VsphereCSISyncer:                    {"*": "gcr.io/cloud-provider-vsphere/csi/release/syncer:v2.7.0"},
		VsphereCSIAttacher:                  {"*": "k8s.gcr.io/sig-storage/csi-attacher:v4.0.0"},
		VsphereCSILivenessProbe:             {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.7.0"},
		VsphereCSINodeDriverRegistar:        {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		VsphereCSIProvisioner:               {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.2.1"},
		VsphereCSIResizer:                   {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.6.0"},
		VsphereCSISnapshotter:               {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v6.0.1"},
		VsphereCSISnapshotController:        {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v6.0.1"},
		VsphereCSISnapshotValidationWebhook: {"*": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v6.0.1"},

		// Nutanix CSI
		NutanixCSI:                          {"*": "quay.io/karbon/ntnx-csi:v2.6.1"},
		NutanixCSILivenessProbe:             {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.7.0"},
		NutanixCSIProvisioner:               {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.2.1"},
		NutanixCSIRegistrar:                 {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		NutanixCSIResizer:                   {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.6.0"},
		NutanixCSISnapshotter:               {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v6.0.1"},
		NutanixCSISnapshotController:        {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v6.0.1"},
		NutanixCSISnapshotValidationWebhook: {"*": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v6.0.1"},

		// GCP Compute Persistent Disk CSI
		GCPComputeCSIDriver:                    {"*": "k8s.gcr.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.8.0"},
		GCPComputeCSIProvisioner:               {"*": "k8s.gcr.io/sig-storage/csi-provisioner:v3.1.0"},
		GCPComputeCSIAttacher:                  {"*": "k8s.gcr.io/sig-storage/csi-attacher:v3.4.0"},
		GCPComputeCSIResizer:                   {"*": "k8s.gcr.io/sig-storage/csi-resizer:v1.4.0"},
		GCPComputeCSISnapshotter:               {"*": "k8s.gcr.io/sig-storage/csi-snapshotter:v6.1.0"},
		GCPComputeCSISnapshotController:        {"*": "k8s.gcr.io/sig-storage/snapshot-controller:v6.1.0"},
		GCPComputeCSISnapshotValidationWebhook: {"*": "k8s.gcr.io/sig-storage/snapshot-validation-webhook:v6.1.0"},
		GCPComputeCSINodeDriverRegistrar:       {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.0"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1"},

		// Cilium
		Cilium:         {"*": "quay.io/cilium/cilium:v1.12.3@sha256:30de50c4dc0a1e1077e9e7917a54d5cab253058b3f779822aec00f5c817ca826"},
		CiliumOperator: {"*": "quay.io/cilium/operator-generic:v1.12.3@sha256:816ec1da586139b595eeb31932c61a7c13b07fb4a0255341c0e0f18608e84eff"},

		// Calico VXLAN
		CalicoVXLANCNI:        {"*": "quay.io/calico/cni:v3.23.3"},
		CalicoVXLANController: {"*": "quay.io/calico/kube-controllers:v3.23.3"},
		CalicoVXLANNode:       {"*": "quay.io/calico/node:v3.23.3"},

		// Hubble
		HubbleRelay:     {"*": "quay.io/cilium/hubble-relay:v1.12.3@sha256:320dff9389e3fc6e2d33863510d497e8bcf245a5755236ae466a0729cc656a79"},
		HubbleUI:        {"*": "quay.io/cilium/hubble-ui:v0.9.2@sha256:d3596efc94a41c6b772b9afe6fe47c17417658956e04c3e2a28d293f2670663e"},
		HubbleUIBackend: {"*": "quay.io/cilium/hubble-ui-backend:v0.9.2@sha256:a3ac4d5b87889c9f7cc6323e86d3126b0d382933bd64f44382a92778b0cde5d7"},
		CiliumCertGen:   {"*": "quay.io/cilium/certgen:v0.1.8@sha256:4a456552a5f192992a6edcec2febb1c54870d665173a33dc7d876129b199ddbd"},

		// Cluster-autoscaler addon
		ClusterAutoscaler: {
			"1.23.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.1",
			"1.24.x":    "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.24.0",
			">= 1.25.0": "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.25.0",
		},

		// CSI Vault Secret Provider
		CSIVaultSecretProvider: {"*": "docker.io/hashicorp/vault-csi-provider:1.1.0"},

		// CSI Secrets Driver
		SecretStoreCSIDriverNodeRegistrar: {"*": "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.1"},
		SecretStoreCSIDriver:              {"*": "k8s.gcr.io/csi-secrets-store/driver:v1.2.1"},
		SecretStoreCSIDriverLivenessProbe: {"*": "k8s.gcr.io/sig-storage/livenessprobe:v2.7.0"},
		SecretStoreCSIDriverCRDs:          {"*": "k8s.gcr.io/csi-secrets-store/driver-crds:v1.2.1"},
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
