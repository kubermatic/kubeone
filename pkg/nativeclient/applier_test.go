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
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const addonLabel = "kubeone.io/addon"

func newFakeClient(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithRESTMapper(testrestmapper.TestOnlyStaticRESTMapper(clientgoscheme.Scheme)).
		WithObjects(objs...).
		Build()
}

func labelledConfigMap(name, addon string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
			Labels:    map[string]string{addonLabel: addon},
		},
	}
}

func TestApplyAndPrune(t *testing.T) {
	ctx := context.Background()
	c := newFakeClient(
		labelledConfigMap("stale", "foo"),
		labelledConfigMap("other-addon", "bar"),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system", Labels: map[string]string{addonLabel: "foo"}}},
	)

	objs, err := ParseManifest(strings.NewReader(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: defaulted
  labels:
    kubeone.io/addon: foo
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept
  namespace: kube-system
  labels:
    kubeone.io/addon: foo
`))
	if err != nil {
		t.Fatal(err)
	}

	applier := New(c)

	result, err := applier.Apply(ctx, objs, ApplyOptions{
		PruneSelector: labels.SelectorFromSet(labels.Set{addonLabel: "foo"}),
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if len(result.Applied) != 2 {
		t.Errorf("Applied = %v, want 2 objects", result.Applied)
	}

	if len(result.Pruned) != 1 || result.Pruned[0].String() != "ConfigMap/kube-system/stale" {
		t.Errorf("Pruned = %v, want [ConfigMap/kube-system/stale]", result.Pruned)
	}

	var cm corev1.ConfigMap
	if err := c.Get(ctx, client.ObjectKey{Namespace: "default", Name: "defaulted"}, &cm); err != nil {
		t.Errorf("expected ConfigMap in default namespace: %v", err)
	} else if cm.Data["key"] != "value" {
		t.Errorf("ConfigMap data = %v, want key=value", cm.Data)
	}

	for _, key := range []client.ObjectKey{{Namespace: "kube-system", Name: "kept"}, {Namespace: "kube-system", Name: "other-addon"}} {
		if err := c.Get(ctx, key, &corev1.ConfigMap{}); err != nil {
			t.Errorf("expected ConfigMap %s to exist: %v", key, err)
		}
	}

	if err := c.Get(ctx, client.ObjectKey{Name: "kube-system"}, &corev1.Namespace{}); err != nil {
		t.Errorf("protected namespace must not be pruned: %v", err)
	}

	// objs must not be mutated by Apply.
	if objs[0].GetNamespace() != "" {
		t.Errorf("Apply() mutated its input, namespace = %q", objs[0].GetNamespace())
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	c := newFakeClient(
		labelledConfigMap("match", "foo"),
		labelledConfigMap("nomatch", "bar"),
	)

	objs, err := ParseManifest(strings.NewReader(`
apiVersion: v1
kind: ConfigMap
metadata: {name: match, namespace: kube-system}
---
apiVersion: v1
kind: ConfigMap
metadata: {name: nomatch, namespace: kube-system}
---
apiVersion: v1
kind: ConfigMap
metadata: {name: missing, namespace: kube-system}
---
apiVersion: example.com/v1
kind: Unknown
metadata: {name: unserved}
`))
	if err != nil {
		t.Fatal(err)
	}

	deleted, err := New(c).Delete(ctx, objs, labels.SelectorFromSet(labels.Set{addonLabel: "foo"}))
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if len(deleted) != 1 || deleted[0].String() != "ConfigMap/kube-system/match" {
		t.Errorf("Delete() = %v, want [ConfigMap/kube-system/match]", deleted)
	}

	err = c.Get(ctx, client.ObjectKey{Namespace: "kube-system", Name: "match"}, &corev1.ConfigMap{})
	if !k8serrors.IsNotFound(err) {
		t.Errorf("expected ConfigMap match to be deleted, got %v", err)
	}

	if err := c.Get(ctx, client.ObjectKey{Namespace: "kube-system", Name: "nomatch"}, &corev1.ConfigMap{}); err != nil {
		t.Errorf("expected ConfigMap nomatch to be kept: %v", err)
	}
}
