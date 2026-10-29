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

package nativeclient

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const testManifest = `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  key: value
---
# comment only
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  namespace: kube-system
spec:
  replicas: 2
---
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Namespace
  metadata:
    name: ns
- apiVersion: apiextensions.k8s.io/v1
  kind: CustomResourceDefinition
  metadata:
    name: foos.example.com
`

func TestParseManifest(t *testing.T) {
	objs, err := ParseManifest(strings.NewReader(testManifest))
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}

	got := refs(objs)
	want := []string{
		"ConfigMap/cm",
		"Deployment.apps/kube-system/deploy",
		"Namespace/ns",
		"CustomResourceDefinition.apiextensions.k8s.io/foos.example.com",
	}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ParseManifest() = %v, want %v", got, want)
	}

	replicas, found, err := unstructured.NestedInt64(objs[1].Object, "spec", "replicas")
	if err != nil || !found || replicas != 2 {
		t.Errorf("spec.replicas = %v (found=%v, err=%v), want int64 2", replicas, found, err)
	}
}

func TestParseManifestErrors(t *testing.T) {
	tests := map[string]string{
		"missing name":   "apiVersion: v1\nkind: ConfigMap\n",
		"missing kind":   "apiVersion: v1\nmetadata:\n  name: x\n",
		"malformed yaml": "apiVersion: v1\nkind: [\n",
	}

	for name, manifest := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseManifest(strings.NewReader(manifest)); err == nil {
				t.Error("ParseManifest() expected an error")
			}
		})
	}
}

func TestSortOrder(t *testing.T) {
	objs, err := ParseManifest(strings.NewReader(testManifest))
	if err != nil {
		t.Fatal(err)
	}

	wantApply := []string{
		"CustomResourceDefinition.apiextensions.k8s.io/foos.example.com",
		"Namespace/ns",
		"ConfigMap/cm",
		"Deployment.apps/kube-system/deploy",
	}
	if got := refs(sortForApply(objs)); strings.Join(got, ",") != strings.Join(wantApply, ",") {
		t.Errorf("sortForApply() = %v, want %v", got, wantApply)
	}

	wantDelete := []string{
		"Deployment.apps/kube-system/deploy",
		"ConfigMap/cm",
		"Namespace/ns",
		"CustomResourceDefinition.apiextensions.k8s.io/foos.example.com",
	}
	if got := refs(sortForDelete(objs)); strings.Join(got, ",") != strings.Join(wantDelete, ",") {
		t.Errorf("sortForDelete() = %v, want %v", got, wantDelete)
	}

	if RefOf(objs[0]).String() != "ConfigMap/cm" {
		t.Errorf("input slice was reordered")
	}
}

func refs(objs []*unstructured.Unstructured) []string {
	out := make([]string, 0, len(objs))
	for _, obj := range objs {
		out = append(out, RefOf(obj).String())
	}

	return out
}
