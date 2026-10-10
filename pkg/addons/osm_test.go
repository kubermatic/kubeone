/*
Copyright 2026 The KubeOne Authors.

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

package addons

import (
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/templates/images"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

func TestOperatingSystemManagerClusterDNS(t *testing.T) {
	t.Parallel()

	const deploymentFile = "deployment-controller.yaml"

	manifest, err := fs.ReadFile(embeddedaddons.FS, resources.AddonOperatingSystemManager+"/"+deploymentFile)
	if err != nil {
		t.Fatalf("unable to read embedded OSM deployment manifest: %v", err)
	}

	tests := []struct {
		name               string
		nodeLocalDNS       bool
		ciliumLRP          bool
		clusterDNS         []string
		serviceSubnet      string
		expectedClusterDNS string
	}{
		{
			name:               "nodeLocalDNS enabled",
			nodeLocalDNS:       true,
			serviceSubnet:      "10.96.0.0/12",
			expectedClusterDNS: resources.NodeLocalDNSVirtualIP,
		},
		{
			name:               "nodeLocalDNS disabled, default service subnet",
			nodeLocalDNS:       false,
			serviceSubnet:      "10.96.0.0/12",
			expectedClusterDNS: "10.96.0.10",
		},
		{
			name:               "nodeLocalDNS disabled, custom service subnet",
			nodeLocalDNS:       false,
			serviceSubnet:      "10.224.0.0/12",
			expectedClusterDNS: "10.224.0.10",
		},
		{
			name:               "cilium local redirect policy enabled",
			nodeLocalDNS:       false,
			ciliumLRP:          true,
			serviceSubnet:      "10.96.0.0/12",
			expectedClusterDNS: resources.NodeLocalDNSVirtualIP + ",10.96.0.10",
		},
		{
			name:               "BYO cilium with local redirect policy enabled",
			clusterDNS:         []string{resources.NodeLocalDNSVirtualIP, "10.96.0.10"},
			serviceSubnet:      "10.96.0.0/12",
			expectedClusterDNS: resources.NodeLocalDNSVirtualIP + ",10.96.0.10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cluster := &kubeoneapi.KubeOneCluster{
				Name: "kubeone-test",
				ClusterNetwork: kubeoneapi.ClusterNetworkConfig{
					ServiceSubnet: tc.serviceSubnet,
					ClusterDNS:    tc.clusterDNS,
					CNI:           &kubeoneapi.CNI{},
				},
				ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
					Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
				},
				Features: kubeoneapi.Features{
					NodeLocalDNS: &kubeoneapi.NodeLocalDNS{Deploy: tc.nodeLocalDNS},
				},
				OperatingSystemManager: &kubeoneapi.OperatingSystemManagerConfig{Deploy: true},
				RegistryConfiguration:  &kubeoneapi.RegistryConfiguration{},
			}

			if tc.ciliumLRP {
				cluster.ClusterNetwork.CNI.Cilium = &kubeoneapi.CiliumSpec{EnableLocalRedirectPolicy: true}
			}

			applier := &applier{
				TemplateData: templateData{
					Config: cluster,
					InternalImages: &internalImages{
						pauseImage: "registry.k8s.io/pause:test",
						resolver: func(images.Resource, ...images.GetOpt) string {
							return "quay.io/kubermatic/operating-system-manager:test"
						},
					},
					Resources: resources.All(cluster.ClusterNetwork.NthServiceSubnetIP(10), resources.ClusterDNS(cluster)),
				},
			}

			fsys := fstest.MapFS{
				resources.AddonOperatingSystemManager + "/" + deploymentFile: &fstest.MapFile{Data: manifest},
			}

			manifests, err := applier.loadAddonsManifests(fsys, resources.AddonOperatingSystemManager, nil, nil, false, cluster, false)
			if err != nil {
				t.Fatalf("unable to render OSM manifests: %v", err)
			}

			var deployment *appsv1.Deployment
			for _, m := range manifests {
				var d appsv1.Deployment
				if err := yaml.Unmarshal(m.Raw, &d); err != nil {
					t.Fatalf("unable to unmarshal manifest: %v", err)
				}
				if d.Kind == "Deployment" && d.Name == resources.OperatingSystemManagerName {
					deployment = &d

					break
				}
			}

			if deployment == nil {
				t.Fatal("OSM deployment not found in rendered manifests")
			}

			args := deployment.Spec.Template.Spec.Containers[0].Args
			clusterDNSArgs := slices.DeleteFunc(slices.Clone(args), func(arg string) bool {
				return !strings.HasPrefix(arg, "-cluster-dns=")
			})

			expected := []string{"-cluster-dns=" + tc.expectedClusterDNS}
			if !slices.Equal(clusterDNSArgs, expected) {
				t.Fatalf("expected %v, got %v (all args: %v)", expected, clusterDNSArgs, args)
			}
		})
	}
}
