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
	_ = x[DigitaloceanCCM-4]
	_ = x[DNSNodeCache-5]
	_ = x[Flannel-6]
	_ = x[HetznerCCM-7]
	_ = x[MachineController-8]
	_ = x[MetricsServer-9]
	_ = x[OpenstackCCM-10]
	_ = x[PacketCCM-11]
	_ = x[VsphereCCM-12]
	_ = x[WeaveNetCNIKube-13]
	_ = x[WeaveNetCNINPC-14]
}

const _Resource_name = "CalicoCNICalicoControllerCalicoNodeDigitaloceanCCMDNSNodeCacheFlannelHetznerCCMMachineControllerMetricsServerOpenstackCCMPacketCCMVsphereCCMWeaveNetCNIKubeWeaveNetCNINPC"

var _Resource_index = [...]uint8{0, 9, 25, 35, 50, 62, 69, 79, 96, 109, 121, 130, 140, 155, 169}

func (i Resource) String() string {
	i -= 1
	if i < 0 || i >= Resource(len(_Resource_index)-1) {
		return "Resource(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Resource_name[_Resource_index[i]:_Resource_index[i+1]]
}
