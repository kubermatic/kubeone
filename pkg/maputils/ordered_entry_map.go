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

import "iter"

type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

// OrderEntryMap is a hashmap that preserves insertion order at the iteration.
type OrderEntryMap[K comparable, V any] struct {
	entries []KeyValue[K, V]
	index   map[K]int
}

// Iter returns an iterator over all key-value pairs in insertion order.
func (oem *OrderEntryMap[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, e := range oem.entries {
			if !yield(e.Key, e.Value) {
				return
			}
		}
	}
}

// NewOrderEntryMap creates a new empty orderedMap.
func NewOrderEntryMap[K comparable, V any](kv ...KeyValue[K, V]) *OrderEntryMap[K, V] {
	oem := &OrderEntryMap[K, V]{
		index: make(map[K]int),
	}

	for _, entry := range kv {
		oem.Set(entry.Key, entry.Value)
	}

	return oem
}

// Set inserts or updates a key-value pair, preserving insertion order for new keys.
func (oem *OrderEntryMap[K, V]) Set(key K, value V) {
	if i, ok := oem.index[key]; ok {
		oem.entries[i].Value = value

		return
	}

	oem.index[key] = len(oem.entries)
	oem.entries = append(oem.entries, KeyValue[K, V]{Key: key, Value: value})
}

func (oem *OrderEntryMap[K, V]) Get(key K) (V, bool) {
	if i, ok := oem.index[key]; ok {
		return oem.entries[i].Value, true
	}

	var zero V
	return zero, false
}
