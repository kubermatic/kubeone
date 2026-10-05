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
	"encoding/json"
	"reflect"
	"testing"

	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
)

// fillNonZero recursively sets every field reachable from v to a non-zero
// value, so that each one is present in the JSON encoding.
func fillNonZero(v reflect.Value) {
	switch v.Kind() { //nolint:exhaustive
	case reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
		fillNonZero(v.Elem())
	case reflect.Struct:
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				fillNonZero(v.Field(i))
			}
		}
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		fillNonZero(v.Index(0))
	case reflect.Map:
		v.Set(reflect.MakeMap(v.Type()))
		key := reflect.New(v.Type().Key()).Elem()
		elem := reflect.New(v.Type().Elem()).Elem()
		fillNonZero(key)
		fillNonZero(elem)
		v.SetMapIndex(key, elem)
	case reflect.String:
		v.SetString("value")
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(1)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(1)
	default:
		panic("fillNonZero: unsupported kind " + v.Kind().String())
	}
}

func TestUpdateWorkersetCopiesAllFields(t *testing.T) {
	tests := []struct {
		name          string
		cloudProvider kubeonev1beta2.CloudProviderSpec
	}{
		{name: "aws", cloudProvider: kubeonev1beta2.CloudProviderSpec{AWS: &kubeonev1beta2.AWSSpec{}}},
		{name: "azure", cloudProvider: kubeonev1beta2.CloudProviderSpec{Azure: &kubeonev1beta2.AzureSpec{}}},
		{name: "digitalocean", cloudProvider: kubeonev1beta2.CloudProviderSpec{DigitalOcean: &kubeonev1beta2.DigitalOceanSpec{}}},
		{name: "equinixmetal", cloudProvider: kubeonev1beta2.CloudProviderSpec{EquinixMetal: &kubeonev1beta2.EquinixMetalSpec{}}},
		{name: "gce", cloudProvider: kubeonev1beta2.CloudProviderSpec{GCE: &kubeonev1beta2.GCESpec{}}},
		{name: "hetzner", cloudProvider: kubeonev1beta2.CloudProviderSpec{Hetzner: &kubeonev1beta2.HetznerSpec{}}},
		{name: "nutanix", cloudProvider: kubeonev1beta2.CloudProviderSpec{Nutanix: &kubeonev1beta2.NutanixSpec{}}},
		{name: "openstack", cloudProvider: kubeonev1beta2.CloudProviderSpec{Openstack: &kubeonev1beta2.OpenstackSpec{}}},
		{name: "vmwareclouddirector", cloudProvider: kubeonev1beta2.CloudProviderSpec{VMwareCloudDirector: &kubeonev1beta2.VMwareCloudDirectorSpec{}}},
		{name: "vsphere", cloudProvider: kubeonev1beta2.CloudProviderSpec{Vsphere: &kubeonev1beta2.VsphereSpec{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := upstreamCloudProviderSpec(tt.cloudProvider)
			if err != nil {
				t.Fatal(err)
			}
			fillNonZero(reflect.ValueOf(spec).Elem())

			input, err := json.Marshal(spec)
			if err != nil {
				t.Fatal(err)
			}

			upstreamSpec, err := upstreamCloudProviderSpec(tt.cloudProvider)
			if err != nil {
				t.Fatal(err)
			}

			workerset := &kubeonev1beta2.DynamicWorkerConfig{}
			if err = updateWorkerset(workerset, input, upstreamSpec); err != nil {
				t.Fatalf("update failed: %v", err)
			}

			var want, got map[string]any
			if err = json.Unmarshal(input, &want); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(workerset.Config.CloudProviderSpec, &got); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(want, got) {
				t.Errorf("CloudProviderSpec mismatch\ngot:  %v\nwant: %v", got, want)
			}
		})
	}
}

func TestUpdateWorkerset(t *testing.T) {
	awsProvider := kubeonev1beta2.CloudProviderSpec{AWS: &kubeonev1beta2.AWSSpec{}}
	openstackProvider := kubeonev1beta2.CloudProviderSpec{Openstack: &kubeonev1beta2.OpenstackSpec{}}

	tests := []struct {
		name          string
		cloudProvider kubeonev1beta2.CloudProviderSpec
		existing      string
		terraform     string
		want          string
		wantErr       bool
	}{
		{
			name:          "existing values take precedence",
			cloudProvider: awsProvider,
			existing:      `{"diskSize":50,"region":"eu-west-1"}`,
			terraform:     `{"diskSize":100,"region":"eu-central-1","vpcId":"vpc-1"}`,
			want:          `{"diskSize":50,"region":"eu-west-1","vpcId":"vpc-1"}`,
		},
		{
			name:          "empty terraform values are ignored",
			cloudProvider: awsProvider,
			existing:      `{"diskSize":50}`,
			terraform:     `{"region":"","securityGroupIDs":[],"tags":{},"diskIops":null,"diskType":"gp3"}`,
			want:          `{"diskSize":50,"diskType":"gp3"}`,
		},
		{
			name:          "false booleans are copied",
			cloudProvider: awsProvider,
			terraform:     `{"assignPublicIP":false,"ebsOptimized":false}`,
			want:          `{"assignPublicIP":false,"ebsOptimized":false}`,
		},
		{
			name:          "config var references are copied verbatim",
			cloudProvider: awsProvider,
			terraform:     `{"region":{"secretKeyRef":{"namespace":"kube-system","name":"aws","key":"region"}}}`,
			want:          `{"region":{"secretKeyRef":{"namespace":"kube-system","name":"aws","key":"region"}}}`,
		},
		{
			name:          "nil existing spec",
			cloudProvider: awsProvider,
			terraform:     `{"vpcId":"vpc-1"}`,
			want:          `{"vpcId":"vpc-1"}`,
		},
		{
			name:          "keys are copied with the casing used by terraform",
			cloudProvider: openstackProvider,
			terraform:     `{"floatingIpPool":"public"}`,
			want:          `{"floatingIpPool":"public"}`,
		},
		{
			name:          "unknown terraform keys are rejected",
			cloudProvider: awsProvider,
			terraform:     `{"notAField":"value"}`,
			wantErr:       true,
		},
		{
			name:          "mistyped terraform values are rejected",
			cloudProvider: awsProvider,
			terraform:     `{"diskSize":"large"}`,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workerset := &kubeonev1beta2.DynamicWorkerConfig{}
			if tt.existing != "" {
				workerset.Config.CloudProviderSpec = json.RawMessage(tt.existing)
			}

			upstreamSpec, err := upstreamCloudProviderSpec(tt.cloudProvider)
			if err != nil {
				t.Fatal(err)
			}

			err = updateWorkerset(workerset, json.RawMessage(tt.terraform), upstreamSpec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if got := string(workerset.Config.CloudProviderSpec); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
