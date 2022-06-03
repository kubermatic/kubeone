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

package v1beta2

import (
	"fmt"

	"k8c.io/kubeone/pkg/fail"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetCloudProvider parses the string representation of the provider
// name and sets the appropriate CloudProviderSpec field.
func SetCloudProvider(cp *CloudProviderSpec, name string) error {
	switch name {
	case "aws":
		cp.AWS = &AWSSpec{}
	case "azure":
		cp.Azure = &AzureSpec{}
	case "digitalocean":
		cp.DigitalOcean = &DigitalOceanSpec{}
	case "gce":
		cp.GCE = &GCESpec{}
	case "hetzner":
		cp.Hetzner = &HetznerSpec{}
	case "nutanix":
		cp.Nutanix = &NutanixSpec{}
	case "openstack":
		cp.Openstack = &OpenstackSpec{}
	case "equinixmetal", "packet":
		cp.EquinixMetal = &EquinixMetalSpec{}
	case "vmwareCloudDirector":
		cp.VMwareCloudDirector = &VMwareCloudDirectorSpec{}
	case "vsphere":
		cp.Vsphere = &VsphereSpec{}
	case "none":
		cp.None = &NoneSpec{}
	default:
		return fail.ConfigValidation(fmt.Errorf("provider %q is not supported", name))
	}

	return nil
}

// NewKubeOneCluster initialize KubeOneCluster with correct typeMeta
func NewKubeOneCluster() *KubeOneCluster {
	return &KubeOneCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeOneCluster",
			APIVersion: SchemeGroupVersion.String(),
		},
	}
}
