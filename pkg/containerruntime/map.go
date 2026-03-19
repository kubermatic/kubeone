/*
Copyright 2021 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF string KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package containerruntime

import (
	"iter"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func UpdateDataMap(cluster *kubeoneapi.KubeOneCluster, inputMap map[string]any) error {
	containerd2Configs, err := marshalContainerdConfigs(cluster)
	if err != nil {
		return err
	}

	inputMap["CONTAINER_RUNTIME_CONFIGS"] = containerd2Configs
	inputMap["CONTAINER_RUNTIME_SOCKET"] = cluster.ContainerRuntime.CRISocket()

	return nil
}

// orderedMapEntry represents a single key-value pair in an ordered map.
type orderedMapEntry struct {
	key   string
	value string
}

// orderedStringMap is a map that preserves insertion order.
type orderedStringMap struct {
	entries []orderedMapEntry
	index   map[string]int
}

// Iter returns an iterator over all key-value pairs in insertion order.
func (m *orderedStringMap) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, e := range m.entries {
			if !yield(e.key, e.value) {
				return
			}
		}
	}
}

// newOrderedMap creates a new empty orderedMap.
func newOrderedMap() *orderedStringMap {
	return &orderedStringMap{
		index: make(map[string]int),
	}
}

// set inserts or updates a key-value pair, preserving insertion order for new keys.
func (m *orderedStringMap) set(key string, value string) {
	if i, ok := m.index[key]; ok {
		m.entries[i].value = value

		return
	}

	m.index[key] = len(m.entries)
	m.entries = append(m.entries, orderedMapEntry{key: key, value: value})
}
