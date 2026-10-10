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
	"fmt"
	"strings"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
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
	AddonCCMKubeVirt            = "ccm-kubevirt"
	AddonCCMGCP                 = "ccm-gcp"
	AddonCCMNutanix             = "ccm-nutanix"
	AddonCCMOpenStack           = "ccm-openstack"
	AddonCCMPacket              = "ccm-packet" // TODO: Remove after deprecation period.
	AddonCCMVsphere             = "ccm-vsphere"
	AddonClusterAutoscaler      = "cluster-autoscaler"
	AddonCNICanal               = "cni-canal"
	AddonCNICilium              = "cni-cilium"
	AddonCNIWeavenet            = "cni-weavenet"
	AddonCoreDNSPDB             = "coredns-pdb"
	AddonCSIExternalSnapshotter = "csi-external-snapshotter"
	AddonCSIAwsEBS              = "csi-aws-ebs"
	AddonCSIAzureDisk           = "csi-azuredisk"
	AddonCSIAzureFile           = "csi-azurefile"
	AddonCSIDigitalOcean        = "csi-digitalocean"
	AddonCSIGCPComputePD        = "csi-gcp-compute-persistent"
	AddonCSIHetzner             = "csi-hetzner"
	AddonCSIKubeVirt            = "csi-kubevirt"
	AddonCSINutanix             = "csi-nutanix"
	AddonCSIOpenStackCinder     = "csi-openstack-cinder"
	AddonCSIVMwareCloudDirector = "csi-vmware-cloud-director"
	AddonCSIVsphere             = "csi-vsphere"
	AddonMachineController      = "machinecontroller"
	AddonMetricsServer          = "metrics-server"
	AddonNodeLocalDNS           = "nodelocaldns"
	AddonNodeLocalDNSCilium     = "nodelocaldns-cilium"
	AddonOperatingSystemManager = "operating-system-manager"
	AddonBackupsRestic          = "backups-restic"
	AddonEtcdDefrag             = "etcd-defrag"
)

func CloudAddons() []string {
	return []string{
		AddonCCMAws,
		AddonCCMAzure,
		AddonCCMGCP,
		AddonCCMDigitalOcean,
		AddonCCMEquinixMetal,
		AddonCCMHetzner,
		AddonCCMOpenStack,
		AddonCCMPacket,
		AddonCCMVsphere,
		AddonCSIExternalSnapshotter,
		AddonCSIAwsEBS,
		AddonCSIAzureDisk,
		AddonCSIAzureFile,
		AddonCSIDigitalOcean,
		AddonCSIGCPComputePD,
		AddonCSIHetzner,
		AddonCSIKubeVirt,
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

// ClusterDNSIPs returns the DNS server IP addresses that kubelets should
// configure for pods. It's the single source of truth for both kubeadm-managed
// nodes and nodes provisioned by machine-controller/OSM.
func ClusterDNSIPs(cluster *kubeoneapi.KubeOneCluster) []string {
	if len(cluster.ClusterNetwork.ClusterDNS) > 0 {
		return cluster.ClusterNetwork.ClusterDNS
	}

	dnsServiceIP := cluster.ClusterNetwork.NthServiceSubnetIP(10)

	switch {
	case cluster.Features.NodeLocalDNS != nil && cluster.Features.NodeLocalDNS.Deploy:
		return []string{NodeLocalDNSVirtualIP}
	case cluster.ClusterNetwork.CNI != nil && cluster.ClusterNetwork.CNI.Cilium != nil && cluster.ClusterNetwork.CNI.Cilium.EnableLocalRedirectPolicy:
		return []string{NodeLocalDNSVirtualIP, dnsServiceIP}
	default:
		return []string{dnsServiceIP}
	}
}

// ClusterDNS returns ClusterDNSIPs as a comma-separated list.
func ClusterDNS(cluster *kubeoneapi.KubeOneCluster) string {
	return strings.Join(ClusterDNSIPs(cluster), ",")
}

func ciliumNodeLocalDNSVirtualIP(dnsServiceIP string) string {
	return fmt.Sprintf("%s,%s", NodeLocalDNSVirtualIP, dnsServiceIP)
}

func All(dnsServiceIP, clusterDNS string) map[string]string {
	return map[string]string{
		"MachineControllerName":             MachineControllerName,
		"MachineControllerNameSpace":        MachineControllerNameSpace,
		"MachineControllerWebhookName":      MachineControllerWebhookName,
		"OperatingSystemManagerName":        OperatingSystemManagerName,
		"OperatingSystemManagerNamespace":   OperatingSystemManagerNamespace,
		"OperatingSystemManagerWebhookName": OperatingSystemManagerWebhookName,
		"KubeletImageRepository":            KubeletImageRepository,
		"NodeLocalDNSVirtualIP":             NodeLocalDNSVirtualIP,
		"ClusterDNS":                        clusterDNS,
		"CiliumNodeLocalDNSVirtualIP":       ciliumNodeLocalDNSVirtualIP(dnsServiceIP),
		"CABundleSSLCertFilePath":           cabundle.SSLCertFilePath,
	}
}
