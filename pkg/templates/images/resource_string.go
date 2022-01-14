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
	_ = x[HubbleRelay-7]
	_ = x[HubbleUI-8]
	_ = x[HubbleUIBackend-9]
	_ = x[HubbleProxy-10]
	_ = x[WeaveNetCNIKube-11]
	_ = x[WeaveNetCNINPC-12]
	_ = x[DNSNodeCache-13]
	_ = x[MachineController-14]
	_ = x[MetricsServer-15]
	_ = x[ClusterAutoscaler-16]
	_ = x[CSIAttacher-17]
	_ = x[CSINodeDriverRegistar-18]
	_ = x[CSIProvisioner-19]
	_ = x[CSISnapshotter-20]
	_ = x[CSIResizer-21]
	_ = x[CSILivenessProbe-22]
	_ = x[AwsCCM-23]
	_ = x[AzureCCM-24]
	_ = x[AzureCNM-25]
	_ = x[AwsEbsCSI-26]
	_ = x[AwsEbsCSIAttacher-27]
	_ = x[AwsEbsCSILivenessProbe-28]
	_ = x[AwsEbsCSINodeDriverRegistrar-29]
	_ = x[AwsEbsCSIProvisioner-30]
	_ = x[AwsEbsCSIResizer-31]
	_ = x[AwsEbsCSISnapshotter-32]
	_ = x[AwsEbsCSISnapshotController-33]
	_ = x[AzureFileCSI-34]
	_ = x[AzureFileCSIAttacher-35]
	_ = x[AzureFileCSILivenessProbe-36]
	_ = x[AzureFileCSINodeDriverRegistar-37]
	_ = x[AzureFileCSIProvisioner-38]
	_ = x[AzureFileCSIResizer-39]
	_ = x[AzureFileCSISnapshotter-40]
	_ = x[AzureFileCSISnapshotterController-41]
	_ = x[NutanixCSISnapshotController-42]
	_ = x[NutanixCSISnapshotValidationWebhook-43]
	_ = x[AzureDiskCSI-44]
	_ = x[AzureDiskCSIAttacher-45]
	_ = x[AzureDiskCSILivenessProbe-46]
	_ = x[AzureDiskCSINodeDriverRegistar-47]
	_ = x[AzureDiskCSIProvisioner-48]
	_ = x[AzureDiskCSIResizer-49]
	_ = x[AzureDiskCSISnapshotter-50]
	_ = x[AzureDiskCSISnapshotterController-51]
	_ = x[NutanixCSILivenessProbe-52]
	_ = x[NutanixCSI-53]
	_ = x[NutanixCSIProvisioner-54]
	_ = x[NutanixCSIRegistrar-55]
	_ = x[NutanixCSIResizer-56]
	_ = x[NutanixCSISnapshotter-57]
	_ = x[DigitaloceanCCM-58]
	_ = x[HetznerCCM-59]
	_ = x[HetznerCSI-60]
	_ = x[OpenstackCCM-61]
	_ = x[OpenstackCSI-62]
	_ = x[EquinixMetalCCM-63]
	_ = x[VsphereCCM-64]
	_ = x[VsphereCSIDriver-65]
	_ = x[VsphereCSISyncer-66]
	_ = x[VsphereCSIProvisioner-67]
}

const _Resource_name = "CalicoCNICalicoControllerCalicoNodeFlannelCiliumCiliumOperatorHubbleRelayHubbleUIHubbleUIBackendHubbleProxyWeaveNetCNIKubeWeaveNetCNINPCDNSNodeCacheMachineControllerMetricsServerClusterAutoscalerCSIAttacherCSINodeDriverRegistarCSIProvisionerCSISnapshotterCSIResizerCSILivenessProbeAwsCCMAzureCCMAzureCNMAwsEbsCSIAwsEbsCSIAttacherAwsEbsCSILivenessProbeAwsEbsCSINodeDriverRegistrarAwsEbsCSIProvisionerAwsEbsCSIResizerAwsEbsCSISnapshotterAwsEbsCSISnapshotControllerAzureFileCSIAzureFileCSIAttacherAzureFileCSILivenessProbeAzureFileCSINodeDriverRegistarAzureFileCSIProvisionerAzureFileCSIResizerAzureFileCSISnapshotterAzureFileCSISnapshotterControllerNutanixCSISnapshotControllerNutanixCSISnapshotValidationWebhookAzureDiskCSIAzureDiskCSIAttacherAzureDiskCSILivenessProbeAzureDiskCSINodeDriverRegistarAzureDiskCSIProvisionerAzureDiskCSIResizerAzureDiskCSISnapshotterAzureDiskCSISnapshotterControllerNutanixCSILivenessProbeNutanixCSINutanixCSIProvisionerNutanixCSIRegistrarNutanixCSIResizerNutanixCSISnapshotterDigitaloceanCCMHetznerCCMHetznerCSIOpenstackCCMOpenstackCSIEquinixMetalCCMVsphereCCMVsphereCSIDriverVsphereCSISyncerVsphereCSIProvisioner"

var _Resource_index = [...]uint16{0, 9, 25, 35, 42, 48, 62, 73, 81, 96, 107, 122, 136, 148, 165, 178, 195, 206, 227, 241, 255, 265, 281, 287, 295, 303, 312, 329, 351, 379, 399, 415, 435, 462, 474, 494, 519, 549, 572, 591, 614, 647, 675, 710, 722, 742, 767, 797, 820, 839, 862, 895, 918, 928, 949, 968, 985, 1006, 1021, 1031, 1041, 1053, 1065, 1080, 1090, 1106, 1122, 1143}

func (i Resource) String() string {
	i -= 1
	if i < 0 || i >= Resource(len(_Resource_index)-1) {
		return "Resource(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Resource_name[_Resource_index[i]:_Resource_index[i+1]]
}
