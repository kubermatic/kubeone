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

package v1beta2

import (
	"reflect"
	"testing"

	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/machine-controller/sdk/cloudprovider/aws"
	"k8c.io/machine-controller/sdk/cloudprovider/azure"
	"k8c.io/machine-controller/sdk/cloudprovider/digitalocean"
	"k8c.io/machine-controller/sdk/cloudprovider/equinixmetal"
	"k8c.io/machine-controller/sdk/cloudprovider/gce"
	"k8c.io/machine-controller/sdk/cloudprovider/hetzner"
	"k8c.io/machine-controller/sdk/cloudprovider/nutanix"
	"k8c.io/machine-controller/sdk/cloudprovider/openstack"
	"k8c.io/machine-controller/sdk/cloudprovider/vmwareclouddirector"
	"k8c.io/machine-controller/sdk/cloudprovider/vsphere"
)

func TestUpstreamCloudProviderSpec(t *testing.T) {
	tests := []struct {
		name          string
		cloudProvider kubeonev1beta2.CloudProviderSpec
		want          any
	}{
		{name: "aws", cloudProvider: kubeonev1beta2.CloudProviderSpec{AWS: &kubeonev1beta2.AWSSpec{}}, want: &aws.RawConfig{}},
		{name: "azure", cloudProvider: kubeonev1beta2.CloudProviderSpec{Azure: &kubeonev1beta2.AzureSpec{}}, want: &azure.RawConfig{}},
		{name: "digitalocean", cloudProvider: kubeonev1beta2.CloudProviderSpec{DigitalOcean: &kubeonev1beta2.DigitalOceanSpec{}}, want: &digitalocean.RawConfig{}},
		{name: "equinixmetal", cloudProvider: kubeonev1beta2.CloudProviderSpec{EquinixMetal: &kubeonev1beta2.EquinixMetalSpec{}}, want: &equinixmetal.RawConfig{}},
		{name: "gce", cloudProvider: kubeonev1beta2.CloudProviderSpec{GCE: &kubeonev1beta2.GCESpec{}}, want: &gce.CloudProviderSpec{}},
		{name: "hetzner", cloudProvider: kubeonev1beta2.CloudProviderSpec{Hetzner: &kubeonev1beta2.HetznerSpec{}}, want: &hetzner.RawConfig{}},
		{name: "nutanix", cloudProvider: kubeonev1beta2.CloudProviderSpec{Nutanix: &kubeonev1beta2.NutanixSpec{}}, want: &nutanix.RawConfig{}},
		{name: "openstack", cloudProvider: kubeonev1beta2.CloudProviderSpec{Openstack: &kubeonev1beta2.OpenstackSpec{}}, want: &openstack.RawConfig{}},
		{name: "vmwareclouddirector", cloudProvider: kubeonev1beta2.CloudProviderSpec{VMwareCloudDirector: &kubeonev1beta2.VMwareCloudDirectorSpec{}}, want: &vmwareclouddirector.RawConfig{}},
		{name: "vsphere", cloudProvider: kubeonev1beta2.CloudProviderSpec{Vsphere: &kubeonev1beta2.VsphereSpec{}}, want: &vsphere.RawConfig{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := upstreamCloudProviderSpec(tt.cloudProvider)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("got %T, want %T", got, tt.want)
			}
		})
	}

	if _, err := upstreamCloudProviderSpec(kubeonev1beta2.CloudProviderSpec{}); err == nil {
		t.Error("expected error for unknown cloud provider")
	}
}
