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

package e2e

import (
	"errors"
	"fmt"
	"strings"

	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	"k8c.io/kubeone/test/e2e/provisioner"
)

type eksdVersions struct {
	Eksd          string
	Etcd          string
	CoreDNS       string
	MetricsServer string
	CNI           string
}

func genEKSDAssetConfig(versions *eksdVersions) (*kubeonev1beta1.AssetConfiguration, error) {
	if versions == nil {
		return nil, errors.New("eks-d versions are not provided")
	}

	// eksdVersion is formatted as v1.19.8-eks-1-19-4
	// eksdRelease will return v1.19.8 and 1-19-4
	eksdRelease := strings.Split(versions.Eksd, "-eks-")
	if len(eksdRelease) != 2 {
		return nil, errors.New("wrong format of provided eksdVersion")
	}
	eksdMajorMinorPatch := strings.Split(eksdRelease[1], "-")
	if len(eksdMajorMinorPatch) != 3 {
		return nil, errors.New("wrong format of provided eksdVersion")
	}

	baseURL := fmt.Sprintf("https://distro.eks.amazonaws.com/kubernetes-%s-%s/releases/%s/artifacts", eksdMajorMinorPatch[0], eksdMajorMinorPatch[1], eksdMajorMinorPatch[2])

	return &kubeonev1beta1.AssetConfiguration{
		Kubernetes: kubeonev1beta1.ImageAsset{
			ImageRepository: "public.ecr.aws/eks-distro/kubernetes",
		},
		Pause: kubeonev1beta1.ImageAsset{
			ImageRepository: "public.ecr.aws/eks-distro/kubernetes",
			ImageTag:        versions.Eksd,
		},
		Etcd: kubeonev1beta1.ImageAsset{
			ImageRepository: "public.ecr.aws/eks-distro/etcd-io",
			ImageTag:        versions.Etcd,
		},
		CoreDNS: kubeonev1beta1.ImageAsset{
			ImageRepository: "public.ecr.aws/eks-distro/coredns",
			ImageTag:        versions.CoreDNS,
		},
		MetricsServer: kubeonev1beta1.ImageAsset{
			ImageRepository: "public.ecr.aws/eks-distro/kubernetes-sigs",
			ImageTag:        versions.MetricsServer,
		},
		CNI: kubeonev1beta1.BinaryAsset{
			URL: fmt.Sprintf("%s/plugins/%s/cni-plugins-linux-amd64-%s.tar.gz", baseURL, versions.CNI, versions.CNI),
		},
		NodeBinaries: kubeonev1beta1.BinaryAsset{
			URL: fmt.Sprintf("%s/kubernetes/%s/kubernetes-node-linux-amd64.tar.gz", baseURL, eksdRelease[0]),
		},
		Kubectl: kubeonev1beta1.BinaryAsset{
			URL: fmt.Sprintf("%s/kubernetes/%s/bin/linux/amd64/kubectl", baseURL, eksdRelease[0]),
		},
	}, nil
}

func EksdTerraformFlags(provider string) ([]string, error) {
	if provider == provisioner.AWS {
		flags := []string{
			"-var", "initial_machinedeployment_replicas=0",
			"-var", "static_workers_count=3",
		}
		return flags, nil
	}
	return nil, errors.New("EKS-D is currently supported only on AWS")
}
