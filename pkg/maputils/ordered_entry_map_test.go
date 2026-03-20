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

package maputils

import (
	"testing"
)

func TestOrderEntryMap_Iter(t *testing.T) {
	tests := []struct {
		name      string
		keyvalues []KeyValue[string, int]
		expected  []KeyValue[string, int]
	}{
		{
			name:      "empty map",
			keyvalues: []KeyValue[string, int]{},
			expected:  []KeyValue[string, int]{},
		},
		{
			name: "single entry",
			keyvalues: []KeyValue[string, int]{
				{"a", 1},
			},
			expected: []KeyValue[string, int]{
				{"a", 1},
			},
		},
		{
			name: "multiple entries in insertion order",
			keyvalues: []KeyValue[string, int]{
				{"a", 1},
				{"b", 2},
				{"b", 3},
				{"b", 4},
				{"c", 3},
			},
			expected: []KeyValue[string, int]{
				{"a", 1},
				{"b", 4},
				{"c", 3},
			},
		},
		{
			name: "updated entry preserves order",
			keyvalues: []KeyValue[string, int]{
				{"a", 1},
				{"b", 2},
				{"a", 10},
				{"c", 3},
			},
			expected: []KeyValue[string, int]{
				{"a", 10},
				{"b", 2},
				{"c", 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewOrderEntryMap[string, int](tt.keyvalues...)

			var result []KeyValue[string, int]
			for k, v := range m.Iter() {
				result = append(result, KeyValue[string, int]{k, v})
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d entries, got %d", len(tt.expected), len(result))
			}

			for i, kv := range result {
				if kv.Key != tt.expected[i].Key || kv.Value != tt.expected[i].Value {
					t.Errorf("entry %d: expected %v, got %v", i, tt.expected[i], kv)
				}
			}
		})
	}
}

func TestOrderEntryMap_Get(t *testing.T) {
	tests := []struct {
		name      string
		keyvalues []KeyValue[string, int]
		getKey    string
		expected  int
		exists    bool
	}{
		{
			name:      "get existing key",
			keyvalues: []KeyValue[string, int]{{"a", 1}, {"b", 2}},
			getKey:    "a",
			expected:  1,
			exists:    true,
		},
		{
			name:      "get non-existing key",
			keyvalues: []KeyValue[string, int]{{"a", 1}},
			getKey:    "b",
			expected:  0,
			exists:    false,
		},
		{
			name:      "get updated key",
			keyvalues: []KeyValue[string, int]{{"a", 1}, {"a", 10}},
			getKey:    "a",
			expected:  10,
			exists:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewOrderEntryMap[string, int](tt.keyvalues...)

			val, exists := m.Get(tt.getKey)
			if exists != tt.exists || val != tt.expected {
				t.Errorf("Get(%q): expected (%v, %v), got (%v, %v)", tt.getKey, tt.expected, tt.exists, val, exists)
			}
		})
	}
}
