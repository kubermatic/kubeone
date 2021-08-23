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
	_ = x[CSIAttacher-6]
	_ = x[CSINodeDriverRegistar-7]
	_ = x[CSIProvisioner-8]
	_ = x[CSISnapshotter-9]
	_ = x[CSIResizer-10]
	_ = x[CSILivenessProbe-11]
	_ = x[DigitaloceanCCM-12]
	_ = x[DNSNodeCache-13]
	_ = x[Flannel-14]
	_ = x[HetznerCCM-15]
	_ = x[MachineController-16]
	_ = x[MetricsServer-17]
	_ = x[OpenstackCCM-18]
	_ = x[OpenstackCSI-19]
	_ = x[PacketCCM-20]
	_ = x[VsphereCCM-21]
	_ = x[WeaveNetCNIKube-22]
	_ = x[WeaveNetCNINPC-23]
}

const _Resource_name = "AzureCCMAzureCNMCalicoCNICalicoControllerCalicoNodeCSIAttacherCSINodeDriverRegistarCSIProvisionerCSISnapshotterCSIResizerCSILivenessProbeDigitaloceanCCMDNSNodeCacheFlannelHetznerCCMMachineControllerMetricsServerOpenstackCCMOpenstackCSIPacketCCMVsphereCCMWeaveNetCNIKubeWeaveNetCNINPC"

var _Resource_index = [...]uint16{0, 8, 16, 25, 41, 51, 62, 83, 97, 111, 121, 137, 152, 164, 171, 181, 198, 211, 223, 235, 244, 254, 269, 283}

func (i Resource) String() string {
	i -= 1
	if i < 0 || i >= Resource(len(_Resource_index)-1) {
		return "Resource(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Resource_name[_Resource_index[i]:_Resource_index[i+1]]
}
