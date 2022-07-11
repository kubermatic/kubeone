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
	AddonCCMHetzner             = "ccm-hetzner"
	AddonCCMOpenStack           = "ccm-openstack"
	AddonCCMEquinixMetal        = "ccm-equinixmetal"
	AddonCCMPacket              = "ccm-packet" // TODO: Remove after deprecation period.
	AddonCCMVsphere             = "ccm-vsphere"
	AddonCSIAwsEBS              = "csi-aws-ebs"
	AddonCSIAzureDisk           = "csi-azuredisk"
	AddonCSIAzureFile           = "csi-azurefile"
	AddonCSIDigitalOcean        = "csi-digitalocean"
	AddonCSIHetzner             = "csi-hetzner"
	AddonCSIGCPComputePD        = "csi-gcp-compute-persistent"
	AddonCSINutanix             = "csi-nutanix"
	AddonCSIOpenStackCinder     = "csi-openstack-cinder"
	AddonCSIVMwareCloudDirector = "csi-vmware-cloud-director"
	AddonCSIVsphere             = "csi-vsphere"
	AddonCNICanal               = "cni-canal"
	AddonCNICilium              = "cni-cilium"
	AddonCNIWeavenet            = "cni-weavenet"
	AddonMachineController      = "machinecontroller"
	AddonOperatingSystemManager = "operating-system-manager"
	AddonMetricsServer          = "metrics-server"
	AddonNodeLocalDNS           = "nodelocaldns"
)

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

	VsphereCSIWebhookName      = "vsphere-webhook-svc"
	VsphereCSIWebhookNamespace = metav1.NamespaceSystem

	VsphereCSISnapshotValidatingWebhookName      = "snapshot-validation-service"
	VsphereCSISnapshotValidatingWebhookNamespace = metav1.NamespaceSystem

	NutanixCSIWebhookName      = "snapshot-validation-service"
	NutanixCSIWebhookNamespace = metav1.NamespaceSystem

	GCEComputeCSIWebhookName      = "snapshot-validation-service"
	GCEComputeCSIWebhookNamespace = metav1.NamespaceSystem

	DigitalOceanCSIWebhookName      = "snapshot-validation-service"
	DigitalOceanCSIWebhookNamespace = metav1.NamespaceSystem
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
