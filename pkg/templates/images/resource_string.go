// Code generated by "stringer -type=Resource"; DO NOT EDIT.

package images

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CalicoCNI-1]
	_ = x[CalicoController-2]
	_ = x[CalicoNode-3]
	_ = x[Flannel-4]
	_ = x[Cilium-5]
	_ = x[CiliumOperator-6]
	_ = x[CiliumEnvoy-7]
	_ = x[HubbleRelay-8]
	_ = x[HubbleUI-9]
	_ = x[HubbleUIBackend-10]
	_ = x[CiliumCertGen-11]
	_ = x[WeaveNetCNIKube-12]
	_ = x[WeaveNetCNINPC-13]
	_ = x[DNSNodeCache-14]
	_ = x[MachineController-15]
	_ = x[MetricsServer-16]
	_ = x[OperatingSystemManager-17]
	_ = x[ClusterAutoscaler-18]
	_ = x[AwsCCM-19]
	_ = x[AzureCCM-20]
	_ = x[AzureCNM-21]
	_ = x[CSISnapshotController-22]
	_ = x[AwsEbsCSI-23]
	_ = x[AwsEbsCSIAttacher-24]
	_ = x[AwsEbsCSILivenessProbe-25]
	_ = x[AwsEbsCSINodeDriverRegistrar-26]
	_ = x[AwsEbsCSIProvisioner-27]
	_ = x[AwsEbsCSIResizer-28]
	_ = x[AwsEbsCSISnapshotter-29]
	_ = x[AzureFileCSI-30]
	_ = x[AzureFileCSIAttacher-31]
	_ = x[AzureFileCSILivenessProbe-32]
	_ = x[AzureFileCSINodeDriverRegistar-33]
	_ = x[AzureFileCSIProvisioner-34]
	_ = x[AzureFileCSIResizer-35]
	_ = x[AzureFileCSISnapshotter-36]
	_ = x[AzureDiskCSI-37]
	_ = x[AzureDiskCSIAttacher-38]
	_ = x[AzureDiskCSILivenessProbe-39]
	_ = x[AzureDiskCSINodeDriverRegistar-40]
	_ = x[AzureDiskCSIProvisioner-41]
	_ = x[AzureDiskCSIResizer-42]
	_ = x[AzureDiskCSISnapshotter-43]
	_ = x[NutanixCSILivenessProbe-44]
	_ = x[NutanixCSI-45]
	_ = x[NutanixCSIProvisioner-46]
	_ = x[NutanixCSIRegistrar-47]
	_ = x[NutanixCSIResizer-48]
	_ = x[NutanixCSISnapshotter-49]
	_ = x[DigitalOceanCSI-50]
	_ = x[DigitalOceanCSIAlpine-51]
	_ = x[DigitalOceanCSIAttacher-52]
	_ = x[DigitalOceanCSINodeDriverRegistar-53]
	_ = x[DigitalOceanCSIProvisioner-54]
	_ = x[DigitalOceanCSIResizer-55]
	_ = x[DigitalOceanCSISnapshotter-56]
	_ = x[OpenstackCSI-57]
	_ = x[OpenstackCSINodeDriverRegistar-58]
	_ = x[OpenstackCSILivenessProbe-59]
	_ = x[OpenstackCSIAttacher-60]
	_ = x[OpenstackCSIProvisioner-61]
	_ = x[OpenstackCSIResizer-62]
	_ = x[OpenstackCSISnapshotter-63]
	_ = x[HetznerCSI-64]
	_ = x[HetznerCSIAttacher-65]
	_ = x[HetznerCSIResizer-66]
	_ = x[HetznerCSIProvisioner-67]
	_ = x[HetznerCSILivenessProbe-68]
	_ = x[HetznerCSINodeDriverRegistar-69]
	_ = x[DigitaloceanCCM-70]
	_ = x[EquinixMetalCCM-71]
	_ = x[HetznerCCM-72]
	_ = x[GCPCCM-73]
	_ = x[NutanixCCM-74]
	_ = x[OpenstackCCM-75]
	_ = x[VsphereCCM-76]
	_ = x[CSIVaultSecretProvider-77]
	_ = x[SecretStoreCSIDriverNodeRegistrar-78]
	_ = x[SecretStoreCSIDriver-79]
	_ = x[SecretStoreCSIDriverLivenessProbe-80]
	_ = x[SecretStoreCSIDriverCRDs-81]
	_ = x[VMwareCloudDirectorCSI-82]
	_ = x[VMwareCloudDirectorCSIAttacher-83]
	_ = x[VMwareCloudDirectorCSIProvisioner-84]
	_ = x[VMwareCloudDirectorCSIResizer-85]
	_ = x[VMwareCloudDirectorCSINodeDriverRegistrar-86]
	_ = x[VsphereCSIDriver-87]
	_ = x[VsphereCSISyncer-88]
	_ = x[VsphereCSIAttacher-89]
	_ = x[VsphereCSILivenessProbe-90]
	_ = x[VsphereCSINodeDriverRegistar-91]
	_ = x[VsphereCSIProvisioner-92]
	_ = x[VsphereCSIResizer-93]
	_ = x[VsphereCSISnapshotter-94]
	_ = x[GCPComputeCSIDriver-95]
	_ = x[GCPComputeCSIProvisioner-96]
	_ = x[GCPComputeCSIAttacher-97]
	_ = x[GCPComputeCSIResizer-98]
	_ = x[GCPComputeCSISnapshotter-99]
	_ = x[GCPComputeCSINodeDriverRegistrar-100]
	_ = x[CalicoVXLANCNI-101]
	_ = x[CalicoVXLANController-102]
	_ = x[CalicoVXLANNode-103]
	_ = x[KubeVirtCSI-104]
	_ = x[KubeVirtCSINodeDriverRegistrar-105]
	_ = x[KubeVirtCSILivenessProbe-106]
	_ = x[KubeVirtCSIProvisioner-107]
	_ = x[KubeVirtCSIAttacher-108]
}

const _Resource_name = "CalicoCNICalicoControllerCalicoNodeFlannelCiliumCiliumOperatorCiliumEnvoyHubbleRelayHubbleUIHubbleUIBackendCiliumCertGenWeaveNetCNIKubeWeaveNetCNINPCDNSNodeCacheMachineControllerMetricsServerOperatingSystemManagerClusterAutoscalerAwsCCMAzureCCMAzureCNMCSISnapshotControllerAwsEbsCSIAwsEbsCSIAttacherAwsEbsCSILivenessProbeAwsEbsCSINodeDriverRegistrarAwsEbsCSIProvisionerAwsEbsCSIResizerAwsEbsCSISnapshotterAzureFileCSIAzureFileCSIAttacherAzureFileCSILivenessProbeAzureFileCSINodeDriverRegistarAzureFileCSIProvisionerAzureFileCSIResizerAzureFileCSISnapshotterAzureDiskCSIAzureDiskCSIAttacherAzureDiskCSILivenessProbeAzureDiskCSINodeDriverRegistarAzureDiskCSIProvisionerAzureDiskCSIResizerAzureDiskCSISnapshotterNutanixCSILivenessProbeNutanixCSINutanixCSIProvisionerNutanixCSIRegistrarNutanixCSIResizerNutanixCSISnapshotterDigitalOceanCSIDigitalOceanCSIAlpineDigitalOceanCSIAttacherDigitalOceanCSINodeDriverRegistarDigitalOceanCSIProvisionerDigitalOceanCSIResizerDigitalOceanCSISnapshotterOpenstackCSIOpenstackCSINodeDriverRegistarOpenstackCSILivenessProbeOpenstackCSIAttacherOpenstackCSIProvisionerOpenstackCSIResizerOpenstackCSISnapshotterHetznerCSIHetznerCSIAttacherHetznerCSIResizerHetznerCSIProvisionerHetznerCSILivenessProbeHetznerCSINodeDriverRegistarDigitaloceanCCMEquinixMetalCCMHetznerCCMGCPCCMNutanixCCMOpenstackCCMVsphereCCMCSIVaultSecretProviderSecretStoreCSIDriverNodeRegistrarSecretStoreCSIDriverSecretStoreCSIDriverLivenessProbeSecretStoreCSIDriverCRDsVMwareCloudDirectorCSIVMwareCloudDirectorCSIAttacherVMwareCloudDirectorCSIProvisionerVMwareCloudDirectorCSIResizerVMwareCloudDirectorCSINodeDriverRegistrarVsphereCSIDriverVsphereCSISyncerVsphereCSIAttacherVsphereCSILivenessProbeVsphereCSINodeDriverRegistarVsphereCSIProvisionerVsphereCSIResizerVsphereCSISnapshotterGCPComputeCSIDriverGCPComputeCSIProvisionerGCPComputeCSIAttacherGCPComputeCSIResizerGCPComputeCSISnapshotterGCPComputeCSINodeDriverRegistrarCalicoVXLANCNICalicoVXLANControllerCalicoVXLANNodeKubeVirtCSIKubeVirtCSINodeDriverRegistrarKubeVirtCSILivenessProbeKubeVirtCSIProvisionerKubeVirtCSIAttacher"

var _Resource_index = [...]uint16{0, 9, 25, 35, 42, 48, 62, 73, 84, 92, 107, 120, 135, 149, 161, 178, 191, 213, 230, 236, 244, 252, 273, 282, 299, 321, 349, 369, 385, 405, 417, 437, 462, 492, 515, 534, 557, 569, 589, 614, 644, 667, 686, 709, 732, 742, 763, 782, 799, 820, 835, 856, 879, 912, 938, 960, 986, 998, 1028, 1053, 1073, 1096, 1115, 1138, 1148, 1166, 1183, 1204, 1227, 1255, 1270, 1285, 1295, 1301, 1311, 1323, 1333, 1355, 1388, 1408, 1441, 1465, 1487, 1517, 1550, 1579, 1620, 1636, 1652, 1670, 1693, 1721, 1742, 1759, 1780, 1799, 1823, 1844, 1864, 1888, 1920, 1934, 1955, 1970, 1981, 2011, 2035, 2057, 2076}

func (i Resource) String() string {
	i -= 1
	if i < 0 || i >= Resource(len(_Resource_index)-1) {
		return "Resource(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Resource_name[_Resource_index[i]:_Resource_index[i+1]]
}
