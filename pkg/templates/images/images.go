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
	KubeVirtCSISnapshotter
	KubeVirtCSIResizer
	KubeVirtCSIDriverOperator

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
		CalicoCNI:               {"*": "quay.io/calico/calico:v3.33.0@sha256:4ebc20730127bc9e8bf1db5f5a2d8b58cf37fe0e891b3c7b5471992221d213e5"},
		CalicoController:        {"*": "quay.io/calico/calico:v3.33.0@sha256:4ebc20730127bc9e8bf1db5f5a2d8b58cf37fe0e891b3c7b5471992221d213e5"},
		CalicoNode:              {"*": "quay.io/calico/node:v3.33.0@sha256:565d8db3200e443955557f8f28ff44ab8a4a456754bc74a19c4f2f3444af9b1c"},
		CalicoThirdPartyPlugins: {"*": "quay.io/calico/third-party-cni-plugins:v3.33.0@sha256:0f32f48279484c0d6f0e0d0e767b2d0e1a6c6ca34cb3421a61b989d992f3df41"},
		DNSNodeCache:            {"*": "registry.k8s.io/dns/k8s-dns-node-cache:1.27.1@sha256:6748acde83ac08626a27a554da5364e8cac64bc950c6a1b6b939a3d27b364ae7"},
		Flannel:                 {"*": "docker.io/flannel/flannel:v0.24.4@sha256:b6b426436c5b221226033a90d77544243a2010692dcf6fa7dfe76f732256fd2e"},
		MachineController:       {"*": "quay.io/kubermatic/machine-controller:d54db023cb7270ada5e1fc8a38de632a30530ebd@sha256:3448a570f36d43e7d501ecdeef51327f063a7b3f755c52dec45245289671728b"},
		MetricsServer:           {"*": "registry.k8s.io/metrics-server/metrics-server:v0.9.0@sha256:d9862115e7c7881280d3d75ca26bda8ffc0fc213315979575bf23ce9826205c0"},
		OperatingSystemManager:  {"*": "quay.io/kubermatic/operating-system-manager:6e5b8a7c774084e77b9bfd61767cc385b46607f6@sha256:667d4784ad1aa03888dbb7d9fbe384d7c14aae1aa4f761ccca5a62706a426927"},
	}
}

func optionalResources() map[Resource]map[string]string {
	return map[Resource]map[string]string{
		AwsCCM: {
			"1.35.x":   "registry.k8s.io/provider-aws/cloud-controller-manager:v1.35.2@sha256:5a5d4e5942ad078a184c12bf5435f8226ccb157ee15e9e1f87c183a4bf4b8f1a",
			"1.36.x":   "registry.k8s.io/provider-aws/cloud-controller-manager:v1.36.1@sha256:bea2cdab2bf136f7fe69ca1767c9cda2b28a6470bf908bda299ec34764fd7ed4",
			">=1.37.x": "registry.k8s.io/provider-aws/cloud-controller-manager:v1.37.0@sha256:af3cbf70e40471e334bacddb7f35bf7ade422bf2f23bc4b47513bea3f3ef1875",
		},

		CSISnapshotController: {"*": "registry.k8s.io/sig-storage/snapshot-controller:v8.7.0@sha256:8511f417d0072efcd937aedd7cc79a937d5066e64ee17beb6097301c92742f19"},
		CSISnapshotWebhook:    {"*": "registry.k8s.io/sig-storage/snapshot-validation-webhook:v8.1.1@sha256:979842f9a6c23ae1b2ddd26603c27412dfc4d3c027d9cda1cb87a67b91ae9ac8"},

		// AWS EBS CSI driver
		AwsEbsCSI:                    {"*": "public.ecr.aws/ebs-csi-driver/aws-ebs-csi-driver:v1.67.0@sha256:aa142ceb797c55396b12afd7f5890b59c791d739227f8c066aef1e1962b1b5f0"},
		AwsEbsCSIAttacher:            {"*": "public.ecr.aws/csi-components/csi-attacher:v4.13.0-eksbuild.1@sha256:0265ec08c22f1ad498c926d7b1b317bedcf5006a3f15162404bded3c00f00336"},
		AwsEbsCSILivenessProbe:       {"*": "public.ecr.aws/csi-components/livenessprobe:v2.20.0-eksbuild.1@sha256:abbf7a40cec1e13d03e888a82051a363c5e345609adbdb9f85202184f370bb88"},
		AwsEbsCSINodeDriverRegistrar: {"*": "public.ecr.aws/csi-components/csi-node-driver-registrar:v2.18.0-eksbuild.1@sha256:245812c3d18e1267926c11c891f1b53001a92d9f2492982945ee2e39cfa57453"},
		AwsEbsCSIProvisioner:         {"*": "public.ecr.aws/csi-components/csi-provisioner:v6.4.0-eksbuild.1@sha256:0f2268740b699b4bdb917465cdfc24e362c3c3b82a3369a225ab7ce1c02ccc8f"},
		AwsEbsCSIResizer:             {"*": "public.ecr.aws/csi-components/csi-resizer:v2.3.0-eksbuild.1@sha256:438d1339cea3251dbac530aa851a447eff1334aa01c2c9c73cbf495871f73481"},
		AwsEbsCSISnapshotter:         {"*": "public.ecr.aws/csi-components/csi-snapshotter:v8.6.0-eksbuild.10@sha256:5bbd4f08715bfbb4548610f12c0f4a76040d6ee6edb6d6540e1ad8751f5196cf"},

		// Azure CCM
		AzureCCM: {
			"1.35.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.35.9@sha256:8add1bd1dfffda732bb8bc0b5bcaaae61e14c8983cce74dba7105b6988fd3625",
			"1.36.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.36.7@sha256:1a1fb615a2190f27ae92ae0ef3024bf631073a0d30fe015e6abcc35500f70344",
			">= 1.37.x": "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.37.0@sha256:ae5ea0d4d5266c117e28983e65281abf59cdf46854fcade32c08a15412606d80",
		},
		AzureCNM: {
			"1.35.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.35.9@sha256:745b4ccf7613e7cc89b97bd71611c08ee883f0e78c588d42c06d178ef58fc3cd",
			"1.36.x":    "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.36.7@sha256:dfe3c003c318264374a7767bf2f68f94189ee1deb4f444843fd2425d9f13ea6a",
			">= 1.37.x": "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.37.0@sha256:5eefaaf64e151096d863fe816d0ac3ef18d34242df54a40e6bc91f8014297cda",
		},

		// AzureFile CSI driver
		AzureFileCSI:                   {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/azurefile-csi:v1.35.10@sha256:ed3a798088b271e72e3a043460bcf496e3089d54b3a09bbf04c8b0d0f8f58e09"},
		AzureFileCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/livenessprobe:v2.19.0@sha256:7a8812cdab00de32ca4fb9109abafbe6b844bf72803875f23cd36c6037cb6b6e"},
		AzureFileCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-node-driver-registrar:v2.17.0@sha256:0b8469fdabc44daf993f26f0e6ad9bd493b972ef9a8be3fecef33e76c200450e"},
		AzureFileCSIProvisioner:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-provisioner:v6.3.0@sha256:6bfdffdbe1101db8dc803ace173c320bdf2e922ef90ab372f2d51c15a561cad2"},
		AzureFileCSIResizer:            {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-resizer:v2.2.0@sha256:2aa3ce6438a19cf48b19ee1a41692c86685694271864903c04fbeb9273ba8b5b"},
		AzureFileCSISnapshotter:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-snapshotter:v8.5.0@sha256:77ccdc782f9432e93af7d87e3dc40767a7e98ddadf6ed7a18aed2f12ece67ed5"},

		// AzureDisk CSI driver
		AzureDiskCSI:                   {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/azuredisk-csi:v1.36.1@sha256:7ec150cb9c23f98c489100855f57f770b4650a1af2be3d8fefe4fa60339eba70"},
		AzureDiskCSIAttacher:           {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-attacher:v4.12.0@sha256:4318420d641012a1d9cc2529f6e047527bc05a87524a7ed8deb7156ac5353b00"},
		AzureDiskCSILivenessProbe:      {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/livenessprobe:v2.19.0@sha256:7a8812cdab00de32ca4fb9109abafbe6b844bf72803875f23cd36c6037cb6b6e"},
		AzureDiskCSINodeDriverRegistar: {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-node-driver-registrar:v2.17.0@sha256:0b8469fdabc44daf993f26f0e6ad9bd493b972ef9a8be3fecef33e76c200450e"},
		AzureDiskCSIProvisioner:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-provisioner:v6.3.0@sha256:6bfdffdbe1101db8dc803ace173c320bdf2e922ef90ab372f2d51c15a561cad2"},
		AzureDiskCSIResizer:            {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-resizer:v2.2.1@sha256:75130dd170f64c201b77555c0bcbc28ffb4df39a211d63d0155360b7c1c3cdce"},
		AzureDiskCSISnapshotter:        {"*": "mcr.microsoft.com/oss/v2/kubernetes-csi/csi-snapshotter:v8.6.0@sha256:a7e8891fa9d51346c68e11d154e62e6dfad1ce7aaadcd46d164646d141796799"},

		// DigitalOcean CCM
		DigitaloceanCCM: {"*": "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.71@sha256:a13b4c188e1d22cc2661a26f40fc7b10a4296b2aee34ba1b9287a488929bec5d"},

		// DigitalOcean CSI
		DigitalOceanCSI:                   {"*": "docker.io/digitalocean/do-csi-plugin:v4.19.0@sha256:32176ee3f34b87dd35ea66517c6bd121068e5c9b637b395f907bc385dc033bcd"},
		DigitalOceanCSIAlpine:             {"*": "docker.io/alpine:3@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf"},
		DigitalOceanCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.13.0@sha256:d1a26170efed45491074ca4545c81844eee2b9043fb885ac01fb4aed6f8b72d2"},
		DigitalOceanCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.18.0@sha256:b7fefd08651f00ac4df1a196ed4c621c2451fbb899297fe9fad9e36c667af0c0"},
		DigitalOceanCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.3.0@sha256:a4b0b1a37605b7b04a293e136edf7006ec1786a8eb3f4e5a945f81d667dcc371"},
		DigitalOceanCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.2.1@sha256:ea1d25e23479000c7e8eeb92d827df66258df4e482ca054c5e7ce3fc0f5c41a5"},
		DigitalOceanCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.6.0@sha256:42af0929bcd60a43499825c078a60ff0534e08af8fbeb283aa391be40feb9f3e"},

		// Hetzner CCM
		HetznerCCM: {"*": "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.39.2@sha256:74a757d1fa70acf06be1b84b6d5ebd73290b35d8ae2ac44f1c667565a54eca8b"},

		// Hetzner CSI
		HetznerCSI:                   {"*": "docker.io/hetznercloud/hcloud-csi-driver:v2.24.3@sha256:3cb24d853d637893512c5a289956c19dd9f5ecefb3ceabb79a42f95fd7614b73"},
		HetznerCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.13.1@sha256:b398d0f5a30202b5261412023b24057bc5a493d8b2b5dad25ebf5c89966e1559"},
		HetznerCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.3.1@sha256:d5b040c2f6ffb25127169feeb845aaf1fa3279e510d71fbebe386b5becc625cc"},
		HetznerCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.4.1@sha256:9e76a2c599ffd7a66f08feeb62ddec61e3545eaa442f8f8f6a495569c0b06829"},
		HetznerCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.20.1@sha256:d06ac4f9242654f278d9c8a4e343b0516542bfc614218d469959c6428765e97f"},
		HetznerCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.18.1@sha256:f847fb218cdf1d0f0fa79772ca6f46e11d25635eb49c06efb5dccd00b5245ef5"},

		// OpenStack CCM
		OpenstackCCM: {
			"1.35.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.35.0@sha256:f12e022d54753ac4023e212f3ac803ccebd84b3a9dc28e0cc961df9d9308645f",
			"1.36.x":    "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.36.0@sha256:e354e40db2d0726ae7c3e40543f1408b4baa6272d983df28775ec351b1e51613",
			">= 1.37.x": "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.37.1@sha256:5b1bf75c418ec59b61d6b817329b780a1004b891fca725a2b3898dc4e67b732f",
		},

		// OpenStack CSI
		OpenstackCSI: {
			"1.35.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.35.0@sha256:eb3b6d4408598503561873c22ebf5fd1dfc9fabcb93c8cfba4909aad8defa556",
			"1.36.x":    "registry.k8s.io/provider-os/cinder-csi-plugin:v1.36.0@sha256:78adaeb154c73c25ba8ebccded98e96b0c993f9fd1cec7daaa5e5ae37de7fff9",
			">= 1.37.x": "registry.k8s.io/provider-os/cinder-csi-plugin:v1.37.1@sha256:ad3832945f1eef2e0cfd1f7e12966d38d894c888dfa153fd34d9650710497a8e",
		},
		OpenstackCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.15.0@sha256:11f199f6bec47403b03cb49c79a41f445884b213b382582a60710b8c6fdc316a"},
		OpenstackCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.17.0@sha256:9b75b9ade162136291d5e8f13a1dfc3dec71ee61419b1bfc112e0796ff8a6aa9"},
		OpenstackCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.10.0@sha256:be59d0556508d3dd419cc34c53062f170a28902ef36e832b947c7796458a083d"},
		OpenstackCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.3.0@sha256:bb057f866177d5f4139a1527e594499cbe0feeb67b63aaca8679dfdf0a6016f9"},
		OpenstackCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.14.0@sha256:5e7cbb63fd497fa913caa21fee1a69f727c220c6fa83c5f8bb0995e2ad73a474"},
		OpenstackCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.4.0@sha256:c7e0a3718832b6197ce8b29fefb3fed3d84f4fbcdf08f4606140dbec2566501d"},

		// Equinix Metal CCM
		EquinixMetalCCM: {"*": "quay.io/equinix-oss/cloud-provider-equinix-metal:v3.8.1@sha256:489474abe43eefceca9fd53a305c071d4f5dfb25f08a771f2450553dd01b785a"},

		// VMware Cloud Director CSI
		VMwareCloudDirectorCSI:                    {"*": "projects.registry.vmware.com/vmware-cloud-director/cloud-director-named-disk-csi-driver:1.6.0@sha256:5ee26850b7ff6fffb1de1442a15b7b82924d4a3e0022db39231679a46ae14c54"},
		VMwareCloudDirectorCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v3.2.1@sha256:60ab9b3e6a030d3038c87c0d6bca2930f58d1d72823e6a4af09767dc83b696a2"},
		VMwareCloudDirectorCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v2.2.2@sha256:04c55b93a032ad8075cae540d60fc740a72c31006590dbdb61ad4ab9813959d0"},
		VMwareCloudDirectorCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v1.4.0@sha256:9ebbf9f023e7b41ccee3d52afe39a89e3ddacdbb69269d583abfc25847cfd9e4"},
		VMwareCloudDirectorCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.2.0@sha256:2dee3fe5fe861bb66c3a4ac51114f3447a4cd35870e0f2e2b558c7a400d89589"},

		// vSphere CPI (A.K.A. CCM)
		VsphereCCM: {
			"1.35.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.35.1@sha256:fab194e856a6d0fa24b22c80c7e6c3e72e0ccc19dad87eae5ae9c23cefe85b79",
			"1.36.x":    "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.36.0@sha256:c0f3933f69006683bed53a67976bbb9106d52b910727492a0bb04c9d6a0bb8ac",
			">= 1.37.x": "registry.k8s.io/cloud-pv-vsphere/cloud-provider-vsphere:v1.37.0@sha256:df7f7380d12d1e26c08556769f6c778b8e7c282463eb972ea9c98bdb96dfa715",
		},

		// vSphere CSI
		VsphereCSIDriver:             {"*": "registry.k8s.io/csi-vsphere/driver:v3.8.0@sha256:35e9468a51b1393f537f9a4e7ead205425f30e11d645eb3e2d9e7c51b1f5c34d"},
		VsphereCSISyncer:             {"*": "registry.k8s.io/csi-vsphere/syncer:v3.8.0@sha256:35760650bed1a4519282861bb1ae668e3cc0c7ef8f22713913079a9b6348ee8c"},
		VsphereCSIAttacher:           {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.10.0@sha256:be59d0556508d3dd419cc34c53062f170a28902ef36e832b947c7796458a083d"},
		VsphereCSILivenessProbe:      {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.17.0@sha256:9b75b9ade162136291d5e8f13a1dfc3dec71ee61419b1bfc112e0796ff8a6aa9"},
		VsphereCSINodeDriverRegistar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.14.0@sha256:5244abbe87e01b35adeb8bb13882a74785df0c0619f8325c9e950395c3f72a97"},
		VsphereCSIProvisioner:        {"*": "registry.k8s.io/sig-storage/csi-provisioner:v4.0.1@sha256:bf5a235b67d8aea00f5b8ec24d384a2480e1017d5458d8a63b361e9eeb1608a9"},
		VsphereCSIResizer:            {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.0.0@sha256:4a95d94e57ad82f6977cd8d4fdcfcfc0b83f02d990e4e7715b688c20970a906d"},
		VsphereCSISnapshotter:        {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.4.0@sha256:c7e0a3718832b6197ce8b29fefb3fed3d84f4fbcdf08f4606140dbec2566501d"},

		// Nutanix CCM
		NutanixCCM: {"*": "ghcr.io/nutanix-cloud-native/cloud-provider-nutanix/controller:v0.6.3@sha256:63ab9b35927d9599d9c227fcb6c59041313d278b5d180998013960f6371aa586"},

		// Nutanix CSI
		NutanixCSI:                      {"*": "docker.io/nutanix/ntnx-csi:3.9.0@sha256:fd6a552868cd2f659731d7f6eaec69cb8a63bc6947ec4d27ef7344366cffd170"},
		NutanixCSILivenessProbe:         {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.19.0@sha256:06da0d5b8908072f2e4522692aee8dc119fba7247a9658497e1153992cd777e9"},
		NutanixCSIExternalHealthMonitor: {"*": "registry.k8s.io/sig-storage/csi-external-health-monitor-controller:v0.18.0@sha256:430c1f2267152ce6e4547ebbad225a2b399d03e5894f2f39ba233a38e4750a47"},
		NutanixCSIAttacher:              {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.12.0@sha256:b9dc9a714a484ccdeeb6f86d88d4db9b7a5ecfc5a55da6db3a60bb3fa33c278a"},
		NutanixCSIProvisioner:           {"*": "registry.k8s.io/sig-storage/csi-provisioner:v5.3.0@sha256:bb057f866177d5f4139a1527e594499cbe0feeb67b63aaca8679dfdf0a6016f9"},
		NutanixCSIRegistrar:             {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.17.0@sha256:f9de845b170155199f2a2a3f9531cf13d78e31235e9db6b6582a8b0db0a50dad"},
		NutanixCSIPrecheck:              {"*": "docker.io/nutanix/ntnx-csi-precheck:3.9.0@sha256:2bf3674172ab46adfb68032af4ac49af22b103beae06473e3f9e2da2047fd45a"},
		NutanixCSIResizer:               {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.2.1@sha256:ea1d25e23479000c7e8eeb92d827df66258df4e482ca054c5e7ce3fc0f5c41a5"},
		NutanixCSISnapshotter:           {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.6.0@sha256:42af0929bcd60a43499825c078a60ff0534e08af8fbeb283aa391be40feb9f3e"},
		NutanixCSIAuthAgent:             {"*": "docker.io/nutanix/ntnx-csi-authagent:3.9.0@sha256:8ee934bc30f646baa1cbb10962558d01e36edf09027cb662e6acd878f7b907bf"},

		// GCP CCM
		GCPCCM: {"*": "registry.k8s.io/cloud-provider-gcp/cloud-controller-manager:v37.0.1@sha256:46f1594881b8829d7dc7bbb2c4a69c71d413cc5ad9bbd13b8a9e11c67a6a8bcf"},

		// GCP Compute Persistent Disk CSI
		// see: https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/blob/master/deploy/kubernetes/images/stable-master/image.yaml
		GCPComputeCSIDriver:              {"*": "registry.k8s.io/cloud-provider-gcp/gcp-compute-persistent-disk-csi-driver:v1.27.1@sha256:3c40baa435c0bf7c66dac277bf13a3cad9c3a3a992d93d39f63ef02bbada939d"},
		GCPComputeCSIProvisioner:         {"*": "registry.k8s.io/sig-storage/csi-provisioner:v6.0.0@sha256:0c537015fe9cf9d53d79d0181e17a78ee784303cd46bf957016605488f212327"},
		GCPComputeCSIAttacher:            {"*": "registry.k8s.io/sig-storage/csi-attacher:v4.8.1@sha256:69888dba58159c8bc0d7c092b9fb97900c9ca8710d088b0b7ea7bd9052df86f6"},
		GCPComputeCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.0.0@sha256:4a95d94e57ad82f6977cd8d4fdcfcfc0b83f02d990e4e7715b688c20970a906d"},
		GCPComputeCSISnapshotter:         {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.2.1@sha256:5f4bb469fec51147ce157329dab598c758da1b018bad6dad26f0ff469326d769"},
		GCPComputeCSINodeDriverRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.13.0@sha256:d7138bcc3aa5f267403d45ad4292c95397e421ea17a0035888850f424c7de25d"},

		// WeaveNet CNI plugin
		WeaveNetCNIKube: {"*": "docker.io/weaveworks/weave-kube:2.8.1@sha256:d797338e7beb17222e10757b71400d8471bdbd9be13b5da38ce2ebf597fb4e63"},
		WeaveNetCNINPC:  {"*": "docker.io/weaveworks/weave-npc:2.8.1@sha256:38d3e30a97a2260558f8deb0fc4c079442f7347f27c86660dbfc8ca91674f14c"},

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
			"1.35.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.35.2@sha256:aac369dc283927a623deb1af54696efcc722ae79255aa07788422e495bab887d",
			"1.36.x":    "registry.k8s.io/autoscaling/cluster-autoscaler:v1.36.1@sha256:461388d03eb5b55db58941ee7a81d9a619dd15302b01cbab9f5a9946715994b4",
			">= 1.37.x": "registry.k8s.io/autoscaling/cluster-autoscaler:v1.36.1@sha256:461388d03eb5b55db58941ee7a81d9a619dd15302b01cbab9f5a9946715994b4",
		},

		// CSI Vault Secret Provider
		CSIVaultSecretProvider: {"*": "docker.io/hashicorp/vault-csi-provider:1.1.0@sha256:849f224b9a7998152c242dd56f6e2ed194ed505678fc1857b31521315e0e12b9"},

		// CSI Secrets Driver
		SecretStoreCSIDriverNodeRegistrar: {"*": "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.5.1@sha256:0103eee7c35e3e0b5cd8cdca9850dc71c793cdeb6669d8be7a89440da2d06ae4"},
		SecretStoreCSIDriver:              {"*": "registry.k8s.io/csi-secrets-store/driver:v1.2.1@sha256:1fa698be6cfa7bb83c90788dfe6c1fd4bd2d72f1b09881c54920e2f81fe82c73"},
		SecretStoreCSIDriverLivenessProbe: {"*": "registry.k8s.io/sig-storage/livenessprobe:v2.7.0@sha256:933940f13b3ea0abc62e656c1aa5c5b47c04b15d71250413a6b821bd0c58b94e"},
		SecretStoreCSIDriverCRDs:          {"*": "registry.k8s.io/csi-secrets-store/driver-crds:v1.2.1@sha256:2ec9390010f4012b5104ac240b38e9b3c416aef65d65166c7f6d8cc3e52b5a54"},

		// KubeVirt's CCM
		KubeVirtCCM: {"*": "quay.io/kubevirt/kubevirt-cloud-controller-manager:v0.6.0@sha256:37ad4c47594119b79e74a5b431095ba653e09283d0dd6e924e28ebaea3b3a978"},

		// KubeVirt CSI
		//
		// kubevirt-csi-driver-operator deploys the node plugin with images hard-coded in the operator release.
		// KubeVirtCSI, KubeVirtCSINodeDriverRegistrar and KubeVirtCSILivenessProbe must match them, and are bumped
		// together with KubeVirtCSIDriverOperator. KubeVirtCSI is also used by the CSI controller.
		KubeVirtCSIDriverOperator:      {"*": "quay.io/kubermatic/kubevirt-csi-driver-operator:v0.5.3@sha256:0bb5541fe2322faa5834dfd49e8cf57ffd6b8a8103b1368a0e7871230c61021a"},
		KubeVirtCSI:                    {"*": "quay.io/kubermatic/kubevirt-csi-driver:v0.4.5@sha256:d31886d3701db91aa339b134ef612b23d9ae3ef2519ddcaac6cc91606a617e9c"},
		KubeVirtCSINodeDriverRegistrar: {"*": "quay.io/openshift/origin-csi-node-driver-registrar:4.20.0@sha256:5645a16ddc56d224343f4b1923b6ed08f5eb11ce79b9caffc4ec5cd93de5f7ea"},
		KubeVirtCSILivenessProbe:       {"*": "quay.io/openshift/origin-csi-livenessprobe:4.20.0@sha256:815d09cdf1cd3c0b8611b00df17a987a962be151e40e02873c44bb0fa90e1f43"},
		KubeVirtCSIProvisioner:         {"*": "quay.io/openshift/origin-csi-external-provisioner:4.20.0@sha256:2a82886b3625d782ccf7b33fa4042f47ec95b2ba7b007edcb549310748ec85c9"},
		KubeVirtCSIAttacher:            {"*": "quay.io/openshift/origin-csi-external-attacher:4.20.0@sha256:ce81973fa79e0699d0c86b8e0717205a4405105e46c8ff12597769456db33ccd"},
		KubeVirtCSISnapshotter:         {"*": "registry.k8s.io/sig-storage/csi-snapshotter:v8.6.0@sha256:42af0929bcd60a43499825c078a60ff0534e08af8fbeb283aa391be40feb9f3e"},
		KubeVirtCSIResizer:             {"*": "registry.k8s.io/sig-storage/csi-resizer:v2.2.1@sha256:ea1d25e23479000c7e8eeb92d827df66258df4e482ca054c5e7ce3fc0f5c41a5"},

		// Backup Restic
		BackupResticSnapshotter: {"*": "registry.k8s.io/etcd:3.5.32-0@sha256:a3bce45d2a04b24ede0d80d7246b68d7dead8840a047dbec1659a68e3f1988e2"},
		BackupResticUploader:    {"*": "ghcr.io/restic/restic:0.19.1@sha256:2f0373803493361f9304a57150d464677f69a9dad487afec202105aafb2592f2"},

		// Unattended upgrades
		UUApline: {"*": "docker.io/library/alpine:3.23@sha256:469c9ea410b9215eaee7976ba6c0639039b444be8e5d638f7abbada189971eb2"},
		UUFluo:   {"*": "ghcr.io/flatcar/flatcar-linux-update-operator:v0.10.0@sha256:461135dc71a9c296d5b1dd5f65cb984fca82fed5fe18c0bace59a5e308653414"},
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
			KubeVirtCSISnapshotter,
			KubeVirtCSIResizer,
			KubeVirtCSIDriverOperator,
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

// tagAndDigest returns the ":tag" and "@digest" suffixes of the reference.
// The tag defaults to "latest" and the digest is empty when not pinned.
func tagAndDigest(named reference.Named) string {
	suffix := ":latest"
	if tagged, ok := named.(reference.Tagged); ok {
		suffix = ":" + tagged.Tag()
	}

	if digested, ok := named.(reference.Digested); ok {
		suffix += "@" + digested.Digest().String()
	}

	return suffix
}

func WithDomain(domain string) GetOpt {
	return func(ref string) string {
		named, _ := reference.ParseNormalizedNamed(ref)

		return domain + "/" + reference.Path(named) + tagAndDigest(named)
	}
}

// WithTag replaces the tag of the reference. The digest is dropped, as it
// belongs to the original tag.
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
	reminder := reference.Path(named) + tagAndDigest(named)

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
