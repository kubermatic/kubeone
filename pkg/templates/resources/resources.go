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

package resources

import (
	"k8c.io/kubeone/pkg/certificate/cabundle"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Names of the internal addons
const (
	AddonCCMAws                 = "ccm-aws"
	AddonCCMAzure               = "ccm-azure"
	AddonCCMDigitalOcean        = "ccm-digitalocean"
	AddonCCMEquinixMetal        = "ccm-equinixmetal"
	AddonCCMHetzner             = "ccm-hetzner"
	AddonCCMOpenStack           = "ccm-openstack"
	AddonCCMPacket              = "ccm-packet" // TODO: Remove after deprecation period.
	AddonCCMVsphere             = "ccm-vsphere"
	AddonCNICanal               = "cni-canal"
	AddonCNICilium              = "cni-cilium"
	AddonCNIWeavenet            = "cni-weavenet"
	AddonCoreDNSPDB             = "coredns-pdb"
	AddonCSIAwsEBS              = "csi-aws-ebs"
	AddonCSIAzureDisk           = "csi-azuredisk"
	AddonCSIAzureFile           = "csi-azurefile"
	AddonCSIDigitalOcean        = "csi-digitalocean"
	AddonCSIGCPComputePD        = "csi-gcp-compute-persistent"
	AddonCSIHetzner             = "csi-hetzner"
	AddonCSINutanix             = "csi-nutanix"
	AddonCSIOpenStackCinder     = "csi-openstack-cinder"
	AddonCSIVMwareCloudDirector = "csi-vmware-cloud-director"
	AddonCSIVsphere             = "csi-vsphere"
	// AddonCSIVsphereKubeSystem represents the CSI driver deployed to Kube-System Namespace.
	AddonCSIVsphereKubeSystem   = "csi-vsphere-ks"
	AddonMachineController      = "machinecontroller"
	AddonMetricsServer          = "metrics-server"
	AddonNodeLocalDNS           = "nodelocaldns"
	AddonOperatingSystemManager = "operating-system-manager"
)

func CloudAddons() []string {
	return []string{
		AddonCCMAws,
		AddonCCMAzure,
		AddonCCMDigitalOcean,
		AddonCCMEquinixMetal,
		AddonCCMHetzner,
		AddonCCMOpenStack,
		AddonCCMPacket,
		AddonCCMVsphere,
		AddonCSIAwsEBS,
		AddonCSIAzureDisk,
		AddonCSIAzureFile,
		AddonCSIDigitalOcean,
		AddonCSIGCPComputePD,
		AddonCSIHetzner,
		AddonCSINutanix,
		AddonCSIOpenStackCinder,
		AddonCSIVMwareCloudDirector,
		AddonCSIVsphere,
	}
}

const (
	NodeLocalDNSVirtualIP = "169.254.20.10"
)

const (
	// names used for deployments/labels/etc
	MachineControllerName        = "machine-controller"
	MachineControllerNameSpace   = metav1.NamespaceSystem
	MachineControllerWebhookName = "machine-controller-webhook"

	OperatingSystemManagerName        = "operating-system-manager"
	OperatingSystemManagerNamespace   = metav1.NamespaceSystem
	OperatingSystemManagerWebhookName = "operating-system-manager-webhook"

	MetricsServerName      = "metrics-server"
	MetricsServerNamespace = metav1.NamespaceSystem

	VsphereCSINamespace        = "vmware-system-csi"
	VsphereCSIWebhookName      = "vsphere-webhook-svc"
	VsphereCSIWebhookNamespace = "vmware-system-csi"
	NutanixCSIWebhookName      = "csi-snapshot-webhook"
	GenericCSIWebhookName      = "snapshot-validation-service"
	GenericCSIWebhookNamespace = metav1.NamespaceSystem
)

const (
	TLSCertName          = "cert.pem"
	TLSKeyName           = "key.pem"
	KubernetesCACertName = "ca.pem"
)

const (
	KubeletImageRepository = "quay.io/kubermatic/kubelet"
)

func All() map[string]string {
	return map[string]string{
		"MachineControllerName":             MachineControllerName,
		"MachineControllerNameSpace":        MachineControllerNameSpace,
		"MachineControllerWebhookName":      MachineControllerWebhookName,
		"OperatingSystemManagerName":        OperatingSystemManagerName,
		"OperatingSystemManagerNamespace":   OperatingSystemManagerNamespace,
		"OperatingSystemManagerWebhookName": OperatingSystemManagerWebhookName,
		"KubeletImageRepository":            KubeletImageRepository,
		"NodeLocalDNSVirtualIP":             NodeLocalDNSVirtualIP,
		"CABundleSSLCertFilePath":           cabundle.SSLCertFilePath,
	}
}
