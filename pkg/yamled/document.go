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

import (
	"io"

	yaml "gopkg.in/yaml.v2"

	"k8c.io/kubeone/pkg/fail"
)

type Document struct {
	root yaml.MapSlice
}

func Load(r io.Reader) (*Document, error) {
	var data yaml.MapSlice
	if err := yaml.NewDecoder(r).Decode(&data); err != nil {
		return nil, fail.Runtime(err, "unmarshal YAML input")
	}

	return NewFromMapSlice(data)
}

func NewFromMapSlice(m yaml.MapSlice) (*Document, error) {
	return &Document{
		root: m,
	}, nil
}

func (d *Document) MarshalYAML() (any, error) {
	return d.root, nil
}

func (d *Document) Root() yaml.MapSlice {
	return d.root
}

func (d *Document) Has(path Path) bool {
	_, exists := d.Get(path)

	return exists
}

func (d *Document) Get(path Path) (any, bool) {
	result := any(d.root)

	for _, step := range path {
		stepFound := false

		switch tstep := step.(type) {
		case string:
			// step is string => try descending down a map
			m, ok := result.(map[string]any)
			if ok {
				result, stepFound = m[tstep]
			} else {
				node, ok := result.(yaml.MapSlice)
				if !ok {
					return nil, false
				}

				for _, item := range node {
					if realItem, _ := item.Key.(string); realItem == tstep {
						stepFound = true
						result = item.Value

						break
					}
				}
			}
		case int:
			// step is int => try getting Nth element of list
			node, ok := result.([]any)
			if !ok {
				return nil, false
			}

			if tstep < 0 || tstep >= len(node) {
				return nil, false
			}

			stepFound = true
			result = node[tstep]
		}

		if !stepFound {
			return nil, false
		}
	}

	return result, true
}

func (d *Document) GetString(path Path) (string, bool) {
	val, exists := d.Get(path)
	if !exists {
		return "", exists
	}

	asserted, ok := val.(string)

	return asserted, ok
}

func (d *Document) GetInt(path Path) (int, bool) {
	val, exists := d.Get(path)
	if !exists {
		return 0, exists
	}

	asserted, ok := val.(int)

	return asserted, ok
}

func (d *Document) GetBool(path Path) (bool, bool) {
	val, exists := d.Get(path)
	if !exists {
		return false, exists
	}

	asserted, ok := val.(bool)

	return asserted, ok
}

func (d *Document) GetArray(path Path) ([]any, bool) {
	val, exists := d.Get(path)
	if !exists {
		return nil, exists
	}

	asserted, ok := val.([]any)

	return asserted, ok
}

func (d *Document) Set(path Path, newValue any) bool {
	// we always need a key or array position to work with
	if len(path) == 0 {
		return false
	}

	return d.setInternal(path, newValue)
}

func (d *Document) setInternal(path Path, newValue any) bool {
	// when we have reached the root level,
	// replace our root element with the new data structure
	if len(path) == 0 {
		return d.setRoot(newValue)
	}

	leafKey := path.Tail()
	parentPath := path.Parent()
	target := any(d.root)

	// check if the parent element exists;
	// create parent if missing

	if len(parentPath) > 0 {
		var exists bool

		target, exists = d.Get(parentPath)
		if !exists {
			switch leafKey.(type) {
			case int:
				if !d.setInternal(parentPath, []any{}) {
					return false
				}
			case string:
				if !d.setInternal(parentPath, map[string]any{}) {
					return false
				}
			default:
				return false
			}

			target, _ = d.Get(parentPath)
		}
	}

	// Now we know that the parent element exists.

	switch value := leafKey.(type) {
	case int:
		// check if we are really in an array
		if array, ok := target.([]any); ok {
			for i := len(array); i <= value; i++ {
				array = append(array, nil)
			}

			array[value] = newValue

			return d.setInternal(parentPath, array)
		}
	case string:
		// check if we are really in a map
		switch m := target.(type) {
		case map[string]any:
			m[value] = newValue

			return d.setInternal(parentPath, m)
		case yaml.MapSlice:
			return d.setInternal(parentPath, setValueInMapSlice(m, value, newValue))
		}
	}

	return false
}

func (d *Document) setRoot(newValue any) bool {
	switch asserted := newValue.(type) {
	case map[string]any:
		d.root = makeMapSlice(asserted)
	case yaml.MapSlice:
		d.root = asserted
	case *yaml.MapSlice:
		d.root = *asserted
	default:
		// attempted to set something that's not a map
		return false
	}

	return true
}

func (d *Document) Append(path Path, newValue any) bool {
	// we require maps at the root level, so the path cannot be empty
	if len(path) == 0 {
		return false
	}

	node, ok := d.Get(path)
	if !ok {
		return d.Set(path, []any{newValue})
	}

	array, ok := node.([]any)
	if !ok {
		return false
	}

	return d.Set(path, append(array, newValue))
}

func (d *Document) Remove(path Path) bool {
	// nuke everything
	if len(path) == 0 {
		return d.setRoot(yaml.MapSlice{})
	}

	leafKey := path.Tail()
	parentPath := path.Parent()

	parent, exists := d.Get(parentPath)
	if !exists {
		return true
	}

	switch value := leafKey.(type) {
	case int:
		if array, ok := parent.([]any); ok {
			return d.setInternal(parentPath, removeArrayItem(array, value))
		}
	case string:
		switch m := parent.(type) {
		case map[string]any:
			delete(m, value)

			return d.setInternal(parentPath, m)
		case yaml.MapSlice:
			return d.setInternal(parentPath, removeKeyFromMapSlice(m, value))
		}
	}

	return false
}

// Fill will set the value at the path to the newValue, but keeps any existing
// sub values intact.
func (d *Document) Fill(path Path, newValue any) bool {
	node, exists := d.Get(path)
	if !exists {
		// exit early if there is nothing fancy to do
		return d.Set(path, newValue)
	}

	if source, ok := unifyMapType(node); ok {
		if newMap, ok := unifyMapType(newValue); ok {
			node = d.fillMap(source, newMap)
		}
	}

	// persist changes to the node
	return d.setInternal(path, node)
}

func (d *Document) fillMap(source yaml.MapSlice, newMap yaml.MapSlice) yaml.MapSlice {
	for _, newItem := range newMap {
		key := newItem.Key
		newValue := newItem.Value
		existingValue, existed := mapSliceGet(source, key)

		if existed {
			if subSource, ok := unifyMapType(existingValue); ok {
				if newSubMap, ok := unifyMapType(newValue); ok {
					source = setValueInMapSlice(source, key, d.fillMap(subSource, newSubMap))
				}
			}
		} else {
			source = setValueInMapSlice(source, key, newValue)
		}
	}

	return source
}
