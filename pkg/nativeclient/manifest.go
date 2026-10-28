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
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"k8c.io/kubeone/pkg/fail"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

var (
	crdGroupKind       = schema.GroupKind{Group: "apiextensions.k8s.io", Kind: "CustomResourceDefinition"}
	namespaceGroupKind = schema.GroupKind{Kind: "Namespace"}
)

// ObjectRef identifies an object independently of the API version it was read or written with.
type ObjectRef struct {
	schema.GroupKind

	Namespace string
	Name      string
}

// RefOf returns the ObjectRef of the given object.
func RefOf(obj *unstructured.Unstructured) ObjectRef {
	return ObjectRef{
		GroupKind: obj.GroupVersionKind().GroupKind(),
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func (r ObjectRef) String() string {
	if r.Namespace == "" {
		return fmt.Sprintf("%s/%s", r.GroupKind, r.Name)
	}

	return fmt.Sprintf("%s/%s/%s", r.GroupKind, r.Namespace, r.Name)
}

// ParseManifest decodes a multi-document YAML or JSON stream into unstructured objects. Empty documents are
// skipped and List objects are flattened into their items.
func ParseManifest(r io.Reader) ([]*unstructured.Unstructured, error) {
	var (
		decoder = yamlutil.NewYAMLOrJSONDecoder(r, 4096)
		objs    []*unstructured.Unstructured
	)

	for i := 0; ; i++ {
		var raw runtime.RawExtension
		if err := decoder.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fail.Runtime(err, "decoding manifest document #%d", i)
		}

		trimmed := bytes.TrimSpace(raw.Raw)
		if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
			continue
		}

		obj := &unstructured.Unstructured{}
		if err := obj.UnmarshalJSON(trimmed); err != nil {
			return nil, fail.Runtime(err, "parsing manifest document #%d", i)
		}

		if !obj.IsList() {
			objs = append(objs, obj)

			continue
		}

		list, err := obj.ToList()
		if err != nil {
			return nil, fail.Runtime(err, "parsing list in manifest document #%d", i)
		}

		for j := range list.Items {
			objs = append(objs, &list.Items[j])
		}
	}

	for _, obj := range objs {
		if obj.GetName() == "" {
			return nil, fail.Runtime(fmt.Errorf("%s object has no metadata.name", obj.GroupVersionKind()), "validating manifest")
		}
	}

	return objs, nil
}

// applyPriority orders objects so that the types other objects depend on are created first.
func applyPriority(obj *unstructured.Unstructured) int {
	switch obj.GroupVersionKind().GroupKind() {
	case crdGroupKind:
		return 0
	case namespaceGroupKind:
		return 1
	default:
		return 2
	}
}

// splitPhases splits objs into CRDs, Namespaces, and everything else, keeping the original order within each.
func splitPhases(objs []*unstructured.Unstructured) (crds, namespaces, rest []*unstructured.Unstructured) {
	for _, obj := range objs {
		switch obj.GroupVersionKind().GroupKind() {
		case crdGroupKind:
			crds = append(crds, obj)
		case namespaceGroupKind:
			namespaces = append(namespaces, obj)
		default:
			rest = append(rest, obj)
		}
	}

	return crds, namespaces, rest
}

// sortForApply returns a copy of objs with CRDs first, Namespaces second, and everything else afterwards, keeping
// the original order within each group.
func sortForApply(objs []*unstructured.Unstructured) []*unstructured.Unstructured {
	sorted := slices.Clone(objs)
	slices.SortStableFunc(sorted, func(a, b *unstructured.Unstructured) int {
		return applyPriority(a) - applyPriority(b)
	})

	return sorted
}

// sortForDelete returns objects in the reverse of the apply order, so custom resources are removed before their
// CRDs and Namespaces are removed last.
func sortForDelete(objs []*unstructured.Unstructured) []*unstructured.Unstructured {
	sorted := sortForApply(objs)
	slices.Reverse(sorted)

	return sorted
}

// listGVK returns the List kind for the given GVK, as expected by controller-runtime for unstructured lists.
func listGVK(gvk schema.GroupVersionKind) schema.GroupVersionKind {
	if strings.HasSuffix(gvk.Kind, "List") {
		return gvk
	}

	return gvk.GroupVersion().WithKind(gvk.Kind + "List")
}
