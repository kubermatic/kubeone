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

package cloudspec

import (
	"encoding/json"
	"reflect"
	"testing"

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

func TestMergeCopiesAllFields(t *testing.T) {
	tests := []struct {
		name         string
		upstreamSpec func() any
	}{
		{name: "aws", upstreamSpec: func() any { return &aws.RawConfig{} }},
		{name: "azure", upstreamSpec: func() any { return &azure.RawConfig{} }},
		{name: "digitalocean", upstreamSpec: func() any { return &digitalocean.RawConfig{} }},
		{name: "equinixmetal", upstreamSpec: func() any { return &equinixmetal.RawConfig{} }},
		{name: "gce", upstreamSpec: func() any { return &gce.CloudProviderSpec{} }},
		{name: "hetzner", upstreamSpec: func() any { return &hetzner.RawConfig{} }},
		{name: "nutanix", upstreamSpec: func() any { return &nutanix.RawConfig{} }},
		{name: "openstack", upstreamSpec: func() any { return &openstack.RawConfig{} }},
		{name: "vmwareclouddirector", upstreamSpec: func() any { return &vmwareclouddirector.RawConfig{} }},
		{name: "vsphere", upstreamSpec: func() any { return &vsphere.RawConfig{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := tt.upstreamSpec()
			fillNonZero(reflect.ValueOf(spec).Elem())

			input, err := json.Marshal(spec)
			if err != nil {
				t.Fatal(err)
			}

			merger, err := NewMerger(tt.upstreamSpec())
			if err != nil {
				t.Fatal(err)
			}

			merged, err := merger.Merge(nil, input)
			if err != nil {
				t.Fatalf("merge failed: %v", err)
			}

			var want, got map[string]any
			if err = json.Unmarshal(input, &want); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(merged, &got); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(want, got) {
				t.Errorf("CloudProviderSpec mismatch\ngot:  %v\nwant: %v", got, want)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	awsProvider := &aws.RawConfig{}
	openstackProvider := &openstack.RawConfig{}
	vcdProvider := &vmwareclouddirector.RawConfig{}

	tests := []struct {
		name          string
		cloudProvider any
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
			name:          "zero values of pointer fields are copied",
			cloudProvider: vcdProvider,
			existing:      `{"cpus":2}`,
			terraform:     `{"diskIOPS":0,"diskSizeGB":0,"storageProfile":"","metadata":{},"cpus":4}`,
			want:          `{"cpus":2,"diskIOPS":0,"diskSizeGB":0,"metadata":{},"storageProfile":""}`,
		},
		{
			name:          "zero values of non-pointer fields are ignored",
			cloudProvider: vcdProvider,
			terraform:     `{"cpus":0,"memoryMB":0,"catalog":"","networks":[],"diskIOPS":null}`,
			want:          `{}`,
		},
		{
			name:          "zero values of pointer fields matched case-insensitively are copied",
			cloudProvider: awsProvider,
			terraform:     `{"diskIOPS":0}`,
			want:          `{"diskIOPS":0}`,
		},
		{
			name:          "existing values take precedence regardless of key casing",
			cloudProvider: openstackProvider,
			existing:      `{"floatingIPPool":"user-pool"}`,
			terraform:     `{"floatingIpPool":"tf-pool","network":"net"}`,
			want:          `{"floatingIPPool":"user-pool","network":"net"}`,
		},
		{
			name:          "existing pointer values take precedence regardless of key casing",
			cloudProvider: awsProvider,
			existing:      `{"diskIOPS":100}`,
			terraform:     `{"diskIops":3000}`,
			want:          `{"diskIOPS":100}`,
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
			var existing json.RawMessage
			if tt.existing != "" {
				existing = json.RawMessage(tt.existing)
			}

			merger, err := NewMerger(tt.cloudProvider)
			if err != nil {
				t.Fatal(err)
			}

			merged, err := merger.Merge(existing, json.RawMessage(tt.terraform))
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if got := string(merged); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNewMergerRejectsNonStructPointer(t *testing.T) {
	for _, spec := range []any{nil, aws.RawConfig{}, new(string)} {
		if _, err := NewMerger(spec); err == nil {
			t.Errorf("expected error for %T", spec)
		}
	}
}
