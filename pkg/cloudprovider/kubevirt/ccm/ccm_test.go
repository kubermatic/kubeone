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

package ccm

import (
	"slices"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/templates/images"

	"sigs.k8s.io/yaml"
)

func fakeImage(res images.Resource, _ ...images.GetOpt) string {
	return "registry.example.com/" + res.String()
}

func allows(comp kubevirtinfra.Component, group, resource, verb string) bool {
	for _, r := range comp.Rules {
		if slices.Contains(r.APIGroups, group) && slices.Contains(r.Resources, resource) && slices.Contains(r.Verbs, verb) {
			return true
		}
	}

	return false
}

func allowsCluster(comp kubevirtinfra.Component, group, resource, verb string) bool {
	for _, r := range comp.ClusterRules {
		if slices.Contains(r.APIGroups, group) && slices.Contains(r.Resources, resource) && slices.Contains(r.Verbs, verb) {
			return true
		}
	}

	return false
}

func TestComponent(t *testing.T) {
	tests := []struct {
		name          string
		loadBalancer  bool
		zoneAndRegion bool
	}{
		{name: "load balancer disabled"},
		{name: "load balancer enabled", loadBalancer: true},
		{name: "zone and region enabled", zoneAndRegion: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			comp := Component(&kubeoneapi.KubevirtSpec{
				LoadBalancerEnabled:  tc.loadBalancer,
				ZoneAndRegionEnabled: tc.zoneAndRegion,
			})

			if comp.NeedsKubeconfig {
				t.Error("the CCM uses in-cluster config, no infra kubeconfig must be generated")
			}

			if !allows(comp, "kubevirt.io", "virtualmachineinstances", "get") || !allows(comp, "kubevirt.io", "virtualmachines", "list") {
				t.Error("Role must allow reading VMs and VMIs")
			}

			if got := allows(comp, "", "services", "create"); got != tc.loadBalancer {
				t.Errorf("services create allowed = %v, want %v", got, tc.loadBalancer)
			}

			// getRegionAndZone reads the infra Node hosting the VMI.
			if got := allowsCluster(comp, "", "nodes", "get"); got != tc.zoneAndRegion {
				t.Errorf("nodes get allowed cluster-wide = %v, want %v", got, tc.zoneAndRegion)
			}
			if !tc.zoneAndRegion && len(comp.ClusterRules) > 0 {
				t.Errorf("no ClusterRules expected without zone and region discovery, got %v", comp.ClusterRules)
			}
		})
	}
}

func TestCloudConfig(t *testing.T) {
	raw, err := CloudConfig("c1", &kubeoneapi.KubevirtSpec{
		InfraNamespace:       "infra-ns",
		ZoneAndRegionEnabled: true,
		LoadBalancerEnabled:  false,
	})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := yaml.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}

	if _, ok := got["kubeconfig"]; ok {
		t.Error("kubeconfig must not be set, the CCM uses in-cluster config for the infra cluster")
	}

	if got["namespace"] != "infra-ns" {
		t.Errorf("namespace = %v", got["namespace"])
	}

	lb, _ := got["loadBalancer"].(map[string]any)
	// The CCM defaults loadBalancer.enabled to true, so false must be explicit
	if v, ok := lb["enabled"]; !ok || v != false {
		t.Errorf("loadBalancer.enabled = %v (set=%v), want explicit false", v, ok)
	}

	inst, _ := got["instancesV2"].(map[string]any)
	if inst["enabled"] != true || inst["zoneAndRegionEnabled"] != true {
		t.Errorf("instancesV2 = %v", inst)
	}

	labels, _ := got["infraLabels"].(map[string]any)
	if labels[kubevirtinfra.LabelClusterName] != "c1" {
		t.Errorf("infraLabels = %v", labels)
	}

	if _, ok := labels[kubevirtinfra.LabelComponent]; ok {
		t.Error("infraLabels must not carry the component label, or reset would delete tenant LoadBalancer Services")
	}
}

func TestDeployment(t *testing.T) {
	d := Deployment("c1", "infra-ns", []byte("cfg"), fakeImage)

	if d.Name != "c1-kubevirt-ccm" || d.Namespace != "infra-ns" {
		t.Errorf("unexpected name %s/%s", d.Namespace, d.Name)
	}

	if d.Labels[kubevirtinfra.LabelClusterName] != "c1" || d.Labels[kubevirtinfra.LabelComponent] == "" {
		t.Errorf("missing cleanup labels: %v", d.Labels)
	}

	pod := d.Spec.Template.Spec
	if pod.ServiceAccountName != "c1-kubevirt-ccm" {
		t.Errorf("ServiceAccountName = %q", pod.ServiceAccountName)
	}

	c := pod.Containers[0]
	if c.Image != "registry.example.com/KubeVirtCCM" {
		t.Errorf("image = %q", c.Image)
	}

	for _, want := range []string{
		"--cloud-provider=kubevirt",
		"--cluster-name=c1",
		"--kubeconfig=/var/run/secrets/tenantcluster/kubeconfig",
		"--authentication-kubeconfig=/var/run/secrets/tenantcluster/kubeconfig",
		"--authorization-kubeconfig=/var/run/secrets/tenantcluster/kubeconfig",
	} {
		if !slices.Contains(c.Args, want) {
			t.Errorf("args %v missing %q", c.Args, want)
		}
	}

	if c.SecurityContext != nil && c.SecurityContext.Privileged != nil && *c.SecurityContext.Privileged {
		t.Error("CCM must not run privileged")
	}

	volumes := map[string]string{}
	for _, v := range pod.Volumes {
		switch {
		case v.ConfigMap != nil:
			volumes["configmap"] = v.ConfigMap.Name
		case v.Secret != nil:
			volumes["secret"] = v.Secret.SecretName
		}
	}

	if volumes["configmap"] != "c1-kubevirt-ccm-cloud-config" {
		t.Errorf("cloud-config ConfigMap = %q", volumes["configmap"])
	}

	if volumes["secret"] != "c1-kubevirt-ccm-tenant-kubeconfig" {
		t.Errorf("tenant kubeconfig Secret = %q", volumes["secret"])
	}

	cm := ConfigMap("c1", "infra-ns", []byte("cfg"))
	if cm.Name != volumes["configmap"] || cm.Data["cloud-config"] != "cfg" {
		t.Errorf("ConfigMap does not match the Deployment: %+v", cm)
	}
}

func TestDeploymentRollsOnCloudConfigChange(t *testing.T) {
	a := Deployment("c1", "infra-ns", []byte("a"), fakeImage).Spec.Template.Annotations[cloudConfigHashAnnotation]
	b := Deployment("c1", "infra-ns", []byte("b"), fakeImage).Spec.Template.Annotations[cloudConfigHashAnnotation]

	if a == "" || a == b {
		t.Errorf("cloud-config hash annotation must change with the cloud-config: %q vs %q", a, b)
	}
}
