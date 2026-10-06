/*
Copyright 2022 The KubeOne Authors.

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

package v1beta3

import (
	"bytes"
	"encoding/json"
	"fmt"

	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/fail"
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

func unmarshalStrict(buf []byte, obj any) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.DisallowUnknownFields()

	return fail.Runtime(dec.Decode(obj), "strict unmarshal of %T", obj)
}

// upstreamCloudProviderSpec returns an empty machine-controller cloud provider
// spec for the cloud provider configured in the given cluster, used to
// validate CloudProviderSpecs coming from the terraform output.
func upstreamCloudProviderSpec(cloudProvider kubeonev1beta3.CloudProviderSpec) (any, error) {
	switch {
	case cloudProvider.AWS != nil:
		return &aws.RawConfig{}, nil
	case cloudProvider.Azure != nil:
		return &azure.RawConfig{}, nil
	case cloudProvider.DigitalOcean != nil:
		return &digitalocean.RawConfig{}, nil
	case cloudProvider.GCE != nil:
		return &gce.CloudProviderSpec{}, nil
	case cloudProvider.Hetzner != nil:
		return &hetzner.RawConfig{}, nil
	case cloudProvider.Nutanix != nil:
		return &nutanix.RawConfig{}, nil
	case cloudProvider.Openstack != nil:
		return &openstack.RawConfig{}, nil
	case cloudProvider.EquinixMetal != nil:
		return &equinixmetal.RawConfig{}, nil
	case cloudProvider.VMwareCloudDirector != nil:
		return &vmwareclouddirector.RawConfig{}, nil
	case cloudProvider.Vsphere != nil:
		return &vsphere.RawConfig{}, nil
	default:
		return nil, fail.Runtime(fmt.Errorf("unknown"), "checking provider")
	}
}
