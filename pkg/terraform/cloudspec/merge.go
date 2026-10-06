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

// Package cloudspec merges machine-controller CloudProviderSpecs coming from
// the terraform output into CloudProviderSpecs of existing workersets.
package cloudspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"k8c.io/kubeone/pkg/fail"
)

// Merger merges terraform output CloudProviderSpecs of a single upstream
// machine-controller cloud provider spec type.
type Merger struct {
	specType reflect.Type
	// fields maps JSON field names to their Go types.
	fields map[string]reflect.Type
	// foldedFields maps lowercased JSON field names to their Go types, used
	// when there is no exact match, like encoding/json does.
	foldedFields map[string]reflect.Type
}

// NewMerger returns a Merger for the type of the given upstream
// machine-controller cloud provider spec, which must be a pointer to a struct.
func NewMerger(upstreamSpec any) (*Merger, error) {
	t := reflect.TypeOf(upstreamSpec)
	if t == nil || t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct {
		return nil, fail.Runtime(fmt.Errorf("expected pointer to struct, got %T", upstreamSpec), "creating CloudProviderSpec merger")
	}

	m := &Merger{
		specType:     t.Elem(),
		fields:       map[string]reflect.Type{},
		foldedFields: map[string]reflect.Type{},
	}
	m.indexFields(m.specType)

	return m, nil
}

// indexFields records the JSON field names of the given struct type, following
// the encoding/json rules: untagged embedded structs are flattened, fields
// tagged "-" are skipped and untagged fields use their Go name.
func (m *Merger) indexFields(t reflect.Type) {
	for field := range t.Fields() {
		tag := field.Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")

		if field.Anonymous && name == "" {
			ft := field.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				m.indexFields(ft)

				continue
			}
		}

		if !field.IsExported() || tag == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}

		if _, ok := m.fields[name]; !ok {
			m.fields[name] = field.Type
		}
		if _, ok := m.foldedFields[strings.ToLower(name)]; !ok {
			m.foldedFields[strings.ToLower(name)] = field.Type
		}
	}
}

// fieldType returns the Go type of the struct field that encoding/json would
// decode the given key into, or nil if there is no such field.
func (m *Merger) fieldType(key string) reflect.Type {
	if t, ok := m.fields[key]; ok {
		return t
	}

	return m.foldedFields[strings.ToLower(key)]
}

// Merge copies values from the terraform output CloudProviderSpec into the
// existing CloudProviderSpec and returns the result. Values already set in the
// existing spec take precedence, also when they are spelled in a different
// casing, since encoding/json matches keys case-insensitively. Empty terraform
// values are ignored (see isEmptyJSON). The terraform output is validated
// against the upstream machine-controller spec type.
func (m *Merger) Merge(existing, terraform json.RawMessage) (json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(terraform))
	dec.DisallowUnknownFields()
	if err := dec.Decode(reflect.New(m.specType).Interface()); err != nil {
		return nil, fail.Config(err, "unmarshalling DynamicWorkerConfig cloud provider spec")
	}

	var tfSpec map[string]json.RawMessage
	if err := json.Unmarshal(terraform, &tfSpec); err != nil {
		return nil, fail.Config(err, "reading terraform CloudProviderSpec")
	}

	spec := make(map[string]json.RawMessage)
	if existing != nil {
		if err := json.Unmarshal(existing, &spec); err != nil {
			return nil, fail.Config(err, "reading CloudProviderSpec")
		}
	}

	existingKeys := make(map[string]struct{}, len(spec))
	for key := range spec {
		existingKeys[strings.ToLower(key)] = struct{}{}
	}

	for key, value := range tfSpec {
		if _, ok := existingKeys[strings.ToLower(key)]; ok || isEmptyJSON(value, m.fieldType(key)) {
			continue
		}
		spec[key] = value
	}

	merged, err := json.Marshal(spec)
	if err != nil {
		return nil, fail.Config(err, "updating cloud provider spec")
	}

	return merged, nil
}

// isEmptyJSON reports whether the given JSON value is considered not set in
// the terraform output. null is always empty. For a value decoded into a
// pointer field of the upstream spec, any non-null value is an explicit
// setting (e.g. diskIOPS: 0), so it is never empty. Otherwise zero values
// (empty string, zero number, empty array or empty object) are empty.
// Booleans are never considered empty.
func isEmptyJSON(value json.RawMessage, fieldType reflect.Type) bool {
	var v any
	if err := json.Unmarshal(value, &v); err != nil {
		return false
	}

	if v == nil {
		return true
	}
	if fieldType != nil && fieldType.Kind() == reflect.Pointer {
		return false
	}

	switch s := v.(type) {
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
