// Code generated by "stringer -type=Resource"; DO NOT EDIT.

package images

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AzureCCM-1]
	_ = x[AzureCNM-2]
	_ = x[CalicoCNI-3]
	_ = x[CalicoController-4]
	_ = x[CalicoNode-5]
	_ = x[Cilium-6]
	_ = x[CiliumOperator-7]
	_ = x[ClusterAutoscaler-8]
	_ = x[CSIAttacher-9]
	_ = x[CSINodeDriverRegistar-10]
	_ = x[CSIProvisioner-11]
	_ = x[CSISnapshotter-12]
	_ = x[CSIResizer-13]
	_ = x[CSILivenessProbe-14]
	_ = x[DigitaloceanCCM-15]
	_ = x[DNSNodeCache-16]
	_ = x[Flannel-17]
	_ = x[HetznerCCM-18]
	_ = x[HetznerCSI-19]
	_ = x[HubbleUi-20]
	_ = x[HubbleUiBackend-21]
	_ = x[HubbleRelay-22]
	_ = x[HubbleProxy-23]
	_ = x[MachineController-24]
	_ = x[MetricsServer-25]
	_ = x[OpenstackCCM-26]
	_ = x[OpenstackCSI-27]
	_ = x[PacketCCM-28]
	_ = x[VsphereCCM-29]
	_ = x[VsphereCSIDriver-30]
	_ = x[VsphereCSISyncer-31]
	_ = x[WeaveNetCNIKube-32]
	_ = x[WeaveNetCNINPC-33]
}

const _Resource_name = "AzureCCMAzureCNMCalicoCNICalicoControllerCalicoNodeCiliumCiliumOperatorClusterAutoscalerCSIAttacherCSINodeDriverRegistarCSIProvisionerCSISnapshotterCSIResizerCSILivenessProbeDigitaloceanCCMDNSNodeCacheFlannelHetznerCCMHetznerCSIHubbleUiHubbleUiBackendHubbleRelayHubbleProxyMachineControllerMetricsServerOpenstackCCMOpenstackCSIPacketCCMVsphereCCMVsphereCSIDriverVsphereCSISyncerWeaveNetCNIKubeWeaveNetCNINPC"

var _Resource_index = [...]uint16{0, 8, 16, 25, 41, 51, 57, 71, 88, 99, 120, 134, 148, 158, 174, 189, 201, 208, 218, 228, 236, 251, 262, 273, 290, 303, 315, 327, 336, 346, 362, 378, 393, 407}

func (i Resource) String() string {
	i -= 1
	if i < 0 || i >= Resource(len(_Resource_index)-1) {
		return "Resource(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Resource_name[_Resource_index[i]:_Resource_index[i+1]]
}
