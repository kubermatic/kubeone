/*
Copyright 2020 The KubeOne Authors.

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
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"text/template"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

var testManifests = []string{
	`kind: ConfigMap
apiVersion: v1
metadata:
  name: test1
  namespace: kube-system
  labels:
    app: test
data:
  foo: bar
`,

	`kind: ConfigMap
apiVersion: v1
metadata:
  name: test2
  namespace: kube-system
  labels:
    app: test
data:
  foo: bar`,

	`kind: ConfigMap
apiVersion: v1
metadata:
  name: test3
  namespace: kube-system
  labels:
    app: test
data:
  foo: bar
`}

const (
	testManifest1WithoutLabel = `apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    app: test
    cluster: {{ .Config.Name }}
  name: test1
  namespace: kube-system
`

	testManifest1WithEmptyLabel = `apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: ""
  name: test1
  namespace: kube-system
`

	testManifest1WithNamedLabel = `apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: test-addon
  name: test1
  namespace: kube-system
`
)

const (
	testManifest1WithImage = `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: ""
  name: test1
  namespace: kube-system
spec:
  containers:
  - image: {{ Registry "k8s.gcr.io" }}/kube-apiserver:v1.19.3
`

	testManifest1WithImageParsed = `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: ""
  name: test1
  namespace: kube-system
spec:
  containers:
  - image: k8s.gcr.io/kube-apiserver:v1.19.3
`

	testManifest1WithCustomImageParsed = `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: ""
  name: test1
  namespace: kube-system
spec:
  containers:
  - image: 127.0.0.1:5000/kube-apiserver:v1.19.3
`
)

var (
	// testManifest1 & testManifest3 have a linebreak at the end, testManifest2 not
	combinedTestManifest = fmt.Sprintf("%s---\n%s\n---\n%s", testManifests[0], testManifests[1], testManifests[2])
)

func TestEnsureAddonsLabelsOnResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		addonName        string
		addonManifest    string
		expectedManifest string
	}{
		{
			name:             "addon with no name (root directory addons)",
			addonName:        "",
			addonManifest:    testManifest1WithoutLabel,
			expectedManifest: testManifest1WithEmptyLabel,
		},
		{
			name:             "addon with name",
			addonName:        "test-addon",
			addonManifest:    testManifest1WithoutLabel,
			expectedManifest: testManifest1WithNamedLabel,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			addonsDir := t.TempDir()

			if writeErr := os.WriteFile(path.Join(addonsDir, "testManifest.yaml"), []byte(tc.addonManifest), 0600); writeErr != nil {
				t.Fatalf("unable to create temporary addon manifest: %v", writeErr)
			}

			td := templateData{
				Config: &kubeoneapi.KubeOneCluster{
					Name: "kubeone-test",
				},
			}

			applier := &applier{
				TemplateData: td,
				LocalFS:      os.DirFS(addonsDir),
			}

			manifests, err := applier.loadAddonsManifests(applier.LocalFS, ".", nil, nil, false, "")
			if err != nil {
				t.Fatalf("unable to load manifests: %v", err)
			}
			if len(manifests) != 1 {
				t.Fatalf("expected to load 1 manifest, got %d", len(manifests))
			}

			b, err := ensureAddonsLabelsOnResources(manifests, tc.addonName)
			if err != nil {
				t.Fatalf("unable to ensure labels: %v", err)
			}
			manifest := b[0].String()

			if manifest != tc.expectedManifest {
				t.Fatalf("invalid manifest returned. expected \n%s, got \n%s", tc.expectedManifest, manifest)
			}
		})
	}
}

func TestCombineManifests(t *testing.T) {
	var manifests []*bytes.Buffer
	for _, m := range testManifests {
		manifests = append(manifests, bytes.NewBufferString(m))
	}

	manifest := combineManifests(manifests)

	if manifest.String() != combinedTestManifest {
		t.Fatalf("invalid combined manifest returned. expected \n%s, got \n%s", combinedTestManifest, manifest.String())
	}
}

func TestImageRegistryParsing(t *testing.T) {
	testCases := []struct {
		name                 string
		registryConfigurtion *kubeoneapi.RegistryConfiguration
		inputManifest        string
		expectedManifest     string
	}{
		{
			name:                 "default registry configuration",
			registryConfigurtion: nil,
			inputManifest:        testManifest1WithImage,
			expectedManifest:     testManifest1WithImageParsed,
		},
		{
			name: "custom registry",
			registryConfigurtion: &kubeoneapi.RegistryConfiguration{
				OverwriteRegistry: "127.0.0.1:5000",
			},
			inputManifest:    testManifest1WithImage,
			expectedManifest: testManifest1WithCustomImageParsed,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			addonsDir := t.TempDir()

			if writeErr := os.WriteFile(path.Join(addonsDir, "testManifest.yaml"), []byte(tc.inputManifest), 0600); writeErr != nil {
				t.Fatalf("unable to create temporary addon manifest: %v", writeErr)
			}

			td := templateData{
				Config: &kubeoneapi.KubeOneCluster{
					Name:                  "kubeone-test",
					RegistryConfiguration: tc.registryConfigurtion,
				},
			}

			overwriteRegistry := ""
			if tc.registryConfigurtion != nil && tc.registryConfigurtion.OverwriteRegistry != "" {
				overwriteRegistry = tc.registryConfigurtion.OverwriteRegistry
			}

			applier := &applier{
				TemplateData: td,
				LocalFS:      os.DirFS(addonsDir),
			}

			manifests, err := applier.loadAddonsManifests(applier.LocalFS, ".", nil, nil, false, overwriteRegistry)
			if err != nil {
				t.Fatalf("unable to load manifests: %v", err)
			}
			if len(manifests) != 1 {
				t.Fatalf("expected to load 1 manifest, got %d", len(manifests))
			}

			b, err := ensureAddonsLabelsOnResources(manifests, "")
			if err != nil {
				t.Fatalf("unable to ensure labels: %v", err)
			}
			manifest := b[0].String()

			if manifest != tc.expectedManifest {
				t.Fatalf("invalid manifest returned. expected \n%s, got \n%s", tc.expectedManifest, manifest)
			}
		})
	}
}

func TestCABundleFuncs(t *testing.T) {
	tests := []string{
		"caBundleEnvVar",
		"caBundleVolume",
		"caBundleVolumeMount",
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt, func(t *testing.T) {
			tpl, err := template.New("addons-base").Funcs(txtFuncMap("")).Parse(fmt.Sprintf(`{{ %s }}`, tt))

			if err != nil {
				t.Errorf("failed to parse template: %v", err)
				t.FailNow()
			}

			var out strings.Builder
			if err := tpl.Execute(&out, nil); err != nil {
				t.Errorf("failed to parse template: %v", err)
			}
			t.Logf("\n%s", out.String())
		})
	}
}
