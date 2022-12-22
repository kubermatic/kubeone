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

package initcmd

import (
	"errors"
	"flag"
	"fmt"
	"sort"
	"testing"

	"k8c.io/kubeone/pkg/testhelper"
)

var updateFlag = flag.Bool("update", false, "update testdata files")

type genOptsParams func(*GenerateOpts)

func withProvider(provName string) genOptsParams {
	return func(opts *GenerateOpts) {
		opts.providerName = provName
	}
}

func withCNI(cni string) genOptsParams {
	return func(opts *GenerateOpts) {
		opts.cni = cni
	}
}

func withEncryption(opts *GenerateOpts) {
	opts.enableFeatureEncryption = true
}

func withCoreDNSPDB(opts *GenerateOpts) {
	opts.enableFeatureCoreDNSPDB = true
}

func withClusterAutoscaler(opts *GenerateOpts) {
	opts.enableAddonAutoscaler = true
}

func withBackupsRestic(opts *GenerateOpts) {
	opts.enableAddonBackups = true
	opts.addonBackupsS3Bucket = "s3:///"
	opts.addonBackupsPassword = "test"
	opts.addonBackupsDefaultAWSRegion = "eu-west-3"
}

func genOpts(opts ...genOptsParams) *GenerateOpts {
	params := &GenerateOpts{
		validProviders:    ValidProviders,
		path:              "",
		clusterName:       "example",
		kubernetesVersion: "v1.24.4",
		providerName:      "aws",
	}

	for _, opt := range opts {
		opt(params)
	}

	return params
}

func TestGenKubeOneClusterYAML(t *testing.T) {
	type testArgs struct {
		name string
		opts *GenerateOpts
		err  error
	}

	tests := []testArgs{
		{
			name: "default with canal",
			opts: genOpts(withCNI(cniCanalValue)),
		},
		{
			name: "default with cilium",
			opts: genOpts(withCNI(cniCiliumValue)),
		},
		{
			name: "default with cilium replacement",
			opts: genOpts(withCNI(cniCiliumReplacementValue)),
		},
		{
			name: "default with external",
			opts: genOpts(withCNI(cniExternalValue)),
		},

		// Canal -- all combinations
		{
			name: "canal with encryption",
			opts: genOpts(
				withCNI(cniCanalValue),
				withEncryption,
			),
		},
		{
			name: "canal with coredns-pdb",
			opts: genOpts(
				withCNI(cniCanalValue),
				withCoreDNSPDB,
			),
		},
		{
			name: "canal with encryption with encryption and coredns-pdb",
			opts: genOpts(
				withCNI(cniCanalValue),
				withEncryption,
				withCoreDNSPDB,
			),
		},
		{
			name: "canal with features and autoscaler",
			opts: genOpts(
				withCNI(cniCanalValue),
				withEncryption,
				withCoreDNSPDB,
				withClusterAutoscaler,
			),
		},
		{
			name: "canal with features and backups",
			opts: genOpts(
				withCNI(cniCanalValue),
				withEncryption,
				withCoreDNSPDB,
				withBackupsRestic,
			),
		},
		{
			name: "canal with features and addons",
			opts: genOpts(
				withCNI(cniCanalValue),
				withEncryption,
				withCoreDNSPDB,
				withClusterAutoscaler,
				withBackupsRestic,
			),
		},

		// Cilium -- all combinations
		{
			name: "cilium with encryption",
			opts: genOpts(
				withCNI(cniCiliumValue),
				withEncryption,
			),
		},
		{
			name: "cilium with coredns-pdb",
			opts: genOpts(
				withCNI(cniCiliumValue),
				withCoreDNSPDB,
			),
		},
		{
			name: "cilium with encryption with encryption and coredns-pdb",
			opts: genOpts(
				withCNI(cniCiliumValue),
				withEncryption,
				withCoreDNSPDB,
			),
		},
		{
			name: "cilium with features and autoscaler",
			opts: genOpts(
				withCNI(cniCiliumValue),
				withEncryption,
				withCoreDNSPDB,
				withClusterAutoscaler,
			),
		},
		{
			name: "cilium with features and backups",
			opts: genOpts(
				withCNI(cniCiliumValue),
				withEncryption,
				withCoreDNSPDB,
				withBackupsRestic,
			),
		},
		{
			name: "cilium with features and addons",
			opts: genOpts(
				withCNI(cniCiliumValue),
				withEncryption,
				withCoreDNSPDB,
				withClusterAutoscaler,
				withBackupsRestic,
			),
		},

		// Cilium Replacement -- all combinations
		{
			name: "cilium replacement with encryption",
			opts: genOpts(
				withCNI(cniCiliumReplacementValue),
				withEncryption,
			),
		},
		{
			name: "cilium replacement with coredns-pdb",
			opts: genOpts(
				withCNI(cniCiliumReplacementValue),
				withCoreDNSPDB,
			),
		},
		{
			name: "cilium replacement with encryption with encryption and coredns-pdb",
			opts: genOpts(
				withCNI(cniCiliumReplacementValue),
				withEncryption,
				withCoreDNSPDB,
			),
		},
		{
			name: "cilium replacement with features and autoscaler",
			opts: genOpts(
				withCNI(cniCiliumReplacementValue),
				withEncryption,
				withCoreDNSPDB,
				withClusterAutoscaler,
			),
		},
		{
			name: "cilium replacement with features and backups",
			opts: genOpts(
				withCNI(cniCiliumReplacementValue),
				withEncryption,
				withCoreDNSPDB,
				withBackupsRestic,
			),
		},
		{
			name: "cilium replacement with features and addons",
			opts: genOpts(
				withCNI(cniCiliumReplacementValue),
				withEncryption,
				withCoreDNSPDB,
				withClusterAutoscaler,
				withBackupsRestic,
			),
		},
	}

	// Add default tests for other providers
	validProvidersNames := []string{}
	for provName := range ValidProviders {
		validProvidersNames = append(validProvidersNames, provName)
	}
	sort.Strings(validProvidersNames)

	for _, provName := range validProvidersNames {
		tests = append(tests, testArgs{
			name: fmt.Sprintf("default on %s", provName),
			opts: genOpts(withProvider(provName)),
		})
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := genKubeOneClusterYAML(tt.opts)
			if !errors.Is(err, tt.err) {
				t.Errorf("genKubeOneClusterYAML() unexpected error: %v, expected err %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), string(got), *updateFlag)
		})
	}
}
