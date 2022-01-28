/*
Copyright 2019 The KubeOne Authors.

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

package yamled

import yaml "gopkg.in/yaml.v2"

func unifyMapType(thing interface{}) (yaml.MapSlice, bool) {
	if m, ok := thing.(map[string]interface{}); ok {
		slice := makeMapSlice(m)

		return slice, true
	}

	// dereference pointer
	if m, ok := thing.(*yaml.MapSlice); ok {
		return *m, true
	}

	// handle any kind of map type
	if m, ok := thing.(yaml.MapSlice); ok {
		return m, true
	}

	return nil, false
}

func makeMapSlice(m map[string]interface{}) yaml.MapSlice {
	result := make(yaml.MapSlice, 0)

	for k, v := range m {
		result = append(result, yaml.MapItem{
			Key:   k,
			Value: v,
		})
	}

	return result
}

func setValueInMapSlice(m yaml.MapSlice, key interface{}, value interface{}) yaml.MapSlice {
	for idx, item := range m {
		if item.Key == key {
			m[idx].Value = value

			return m
		}
	}

	return append(m, yaml.MapItem{
		Key:   key,
		Value: value,
	})
}

func removeArrayItem(array []interface{}, pos int) []interface{} {
	if pos >= 0 && pos < len(array) {
		array = append(array[:pos], array[pos+1:]...)
	}

	return array
}

func removeKeyFromMapSlice(m yaml.MapSlice, key interface{}) yaml.MapSlice {
	for idx, item := range m {
		if item.Key == key {
			return append(m[:idx], m[idx+1:]...)
		}
	}

	return m
}

func mapSliceGet(haystack yaml.MapSlice, key interface{}) (interface{}, bool) {
	for _, item := range haystack {
		if item.Key == key {
			return item.Value, true
		}
	}

	return nil, false
}
