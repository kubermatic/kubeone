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
	"io/ioutil"
	"os"
	"path"
	"testing"

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

	testManifest1WithLabel = `apiVersion: v1
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
)

var (
	// testManifest1 & testManifest3 have a linebreak at the end, testManifest2 not
	combinedTestManifest = fmt.Sprintf("%s---\n%s\n---\n%s", testManifests[0], testManifests[1], testManifests[2])
)

func TestEnsureAddonsLabelsOnResources(t *testing.T) {
	addonsDir, err := ioutil.TempDir("/tmp", "kubeone")
	if err != nil {
		t.Fatalf("unable to create temporary addons directory: %v", err)
	}
	defer os.RemoveAll(addonsDir)

	if writeErr := ioutil.WriteFile(path.Join(addonsDir, "testManifest.yaml"), []byte(testManifest1WithoutLabel), 0600); writeErr != nil {
		t.Fatalf("unable to create temporary addon manifest: %v", err)
	}

	templateData := TemplateData{
		Config: &kubeoneapi.KubeOneCluster{
			Name: "kubeone-test",
		},
	}
	manifests, err := loadAddonsManifests(addonsDir, nil, false, templateData)
	if err != nil {
		t.Fatalf("unable to load manifests: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected to load 1 manifest, got %d", len(manifests))
	}

	b, err := ensureAddonsLabelsOnResources(manifests)
	if err != nil {
		t.Fatalf("unable to ensure labels: %v", err)
	}
	manifest := b[0].String()

	if manifest != testManifest1WithLabel {
		t.Fatalf("invalid manifest returned. expected \n%s, got \n%s", testManifest1WithLabel, manifest)
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
