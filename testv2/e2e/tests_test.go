// Code generated by e2e/generator, DO NOT EDIT.

package e2e

import (
	"testing"
)

func TestStub(t *testing.T) {
	t.Skip("stub is skipped")
}
func TestAwsDefaultsInstallContainerdV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsCentosInstallContainerdV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsRhelInstallContainerdV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsFlatcarInstallContainerdV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsAmznInstallContainerdV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsDefaultsInstallContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsCentosInstallContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsRhelInstallContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsFlatcarInstallContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsAmznInstallContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsDefaultsInstallContainerdV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.23.6")
	scenario.Run(t)
}

func TestAwsCentosInstallContainerdV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.23.6")
	scenario.Run(t)
}

func TestAwsRhelInstallContainerdV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.23.6")
	scenario.Run(t)
}

func TestAwsFlatcarInstallContainerdV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.23.6")
	scenario.Run(t)
}

func TestAwsAmznInstallContainerdV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["install_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.23.6")
	scenario.Run(t)
}

func TestAwsDefaultsInstallDockerV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsCentosInstallDockerV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsRhelInstallDockerV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsFlatcarInstallDockerV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsAmznInstallDockerV1_21_12(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12")
	scenario.Run(t)
}

func TestAwsDefaultsInstallDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsCentosInstallDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsRhelInstallDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsFlatcarInstallDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsAmznInstallDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["install_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsDefaultsUpgradeContainerdFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsCentosUpgradeContainerdFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsRhelUpgradeContainerdFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsFlatcarUpgradeContainerdFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsAmznUpgradeContainerdFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsDefaultsUpgradeContainerdFromV1_22_9_ToV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9", "v1.23.6")
	scenario.Run(t)
}

func TestAwsCentosUpgradeContainerdFromV1_22_9_ToV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9", "v1.23.6")
	scenario.Run(t)
}

func TestAwsRhelUpgradeContainerdFromV1_22_9_ToV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9", "v1.23.6")
	scenario.Run(t)
}

func TestAwsFlatcarUpgradeContainerdFromV1_22_9_ToV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9", "v1.23.6")
	scenario.Run(t)
}

func TestAwsAmznUpgradeContainerdFromV1_22_9_ToV1_23_6(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["upgrade_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9", "v1.23.6")
	scenario.Run(t)
}

func TestAwsDefaultsUpgradeDockerFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["upgrade_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsCentosUpgradeDockerFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["upgrade_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsRhelUpgradeDockerFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["upgrade_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsFlatcarUpgradeDockerFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["upgrade_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsAmznUpgradeDockerFromV1_21_12_ToV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["upgrade_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.21.12", "v1.22.9")
	scenario.Run(t)
}

func TestAwsDefaultsCalicoContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["calico_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsCentosCalicoContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["calico_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsRhelCalicoContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["calico_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsFlatcarCalicoContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["calico_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsAmznCalicoContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["calico_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsDefaultsCalicoDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["calico_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsCentosCalicoDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["calico_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsRhelCalicoDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["calico_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsFlatcarCalicoDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["calico_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsAmznCalicoDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["calico_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsDefaultsWeaveContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["weave_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsCentosWeaveContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["weave_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsRhelWeaveContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["weave_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsFlatcarWeaveContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["weave_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsAmznWeaveContainerdV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["weave_containerd"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsDefaultsWeaveDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_defaults"]
	scenario := Scenarios["weave_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsCentosWeaveDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_centos"]
	scenario := Scenarios["weave_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsRhelWeaveDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_rhel"]
	scenario := Scenarios["weave_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsFlatcarWeaveDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_flatcar"]
	scenario := Scenarios["weave_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}

func TestAwsAmznWeaveDockerV1_22_9(t *testing.T) {
	infra := Infrastructures["aws_amzn"]
	scenario := Scenarios["weave_docker"]
	scenario.SetInfra(infra)
	scenario.SetVersions("v1.22.9")
	scenario.Run(t)
}
