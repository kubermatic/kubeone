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

package v1beta2

import (
	"bytes"
	"encoding/json"
	"fmt"

	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
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
func upstreamCloudProviderSpec(cloudProvider kubeonev1beta2.CloudProviderSpec) (any, error) {
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

// updateWorkerset copies values from the terraform output CloudProviderSpec
// into the CloudProviderSpec of the existing workerset. Values already set in
// the existing workerset take precedence, and empty terraform values are
// ignored. The terraform output is validated against the given upstream
// machine-controller spec type.
func updateWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage, upstreamSpec any) error {
	if err := unmarshalStrict(cfg, upstreamSpec); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig cloud provider spec")
	}

	var tfSpec map[string]json.RawMessage
	if err := json.Unmarshal(cfg, &tfSpec); err != nil {
		return fail.Config(err, "reading terraform CloudProviderSpec")
	}

	spec := make(map[string]json.RawMessage)
	if existingWorkerSet.Config.CloudProviderSpec != nil {
		if err := json.Unmarshal(existingWorkerSet.Config.CloudProviderSpec, &spec); err != nil {
			return fail.Config(err, "reading CloudProviderSpec")
		}
	}

	for key, value := range tfSpec {
		if _, ok := spec[key]; ok || isEmptyJSON(value) {
			continue
		}
		spec[key] = value
	}

	var err error
	existingWorkerSet.Config.CloudProviderSpec, err = json.Marshal(spec)
	if err != nil {
		return fail.Config(err, "updating cloud provider spec")
	}

	return nil
}

// isEmptyJSON reports whether the given JSON value is null or a zero value
// (empty string, zero number, empty array or empty object), i.e. not set in
// the terraform output. Booleans are never considered empty.
func isEmptyJSON(value json.RawMessage) bool {
	var v any
	if err := json.Unmarshal(value, &v); err != nil {
		return false
	}

	switch s := v.(type) {
	case nil:
		return true
	case string:
		return s == ""
	case float64:
		return s == 0
	case []any:
		return len(s) == 0
	case map[string]any:
		return len(s) == 0
	default:
		return false
	}
}
