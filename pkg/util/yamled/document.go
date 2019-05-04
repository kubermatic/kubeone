package yamled

import (
	"fmt"
	"io"

	yaml "gopkg.in/yaml.v2"
)

type Document struct {
	root yaml.MapSlice
}

func Load(r io.Reader) (*Document, error) {
	var data yaml.MapSlice
	if err := yaml.NewDecoder(r).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode input YAML: %v", err)
	}

	return NewFromMapSlice(data)
}

func NewFromMapSlice(m yaml.MapSlice) (*Document, error) {
	return &Document{
		root: m,
	}, nil
}

func (d *Document) MarshalYAML() (interface{}, error) {
	return d.root, nil
}

func (d *Document) Root() yaml.MapSlice {
	return d.root
}

func (d *Document) Has(path Path) bool {
	_, exists := d.Get(path)

	return exists
}

func (d *Document) Get(path Path) (interface{}, bool) {
	result := interface{}(d.root)

	for _, step := range path {
		stepFound := false

		if sstep, ok := step.(string); ok {
			// step is string => try descending down a map
			m, ok := result.(map[string]interface{})
			if ok {
				result, stepFound = m[sstep]
			} else {
				node, ok := result.(yaml.MapSlice)
				if !ok {
					return nil, false
				}

				for _, item := range node {
					if item.Key.(string) == sstep {
						stepFound = true
						result = item.Value
						break
					}
				}
			}
		} else if istep, ok := step.(int); ok {
			// step is int => try getting Nth element of list
			node, ok := result.([]interface{})
			if !ok {
				return nil, false
			}

			if istep < 0 || istep >= len(node) {
				return nil, false
			}

			stepFound = true
			result = node[istep]
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

func (d *Document) GetArray(path Path) ([]interface{}, bool) {
	val, exists := d.Get(path)
	if !exists {
		return nil, exists
	}

	asserted, ok := val.([]interface{})

	return asserted, ok
}

func (d *Document) Set(path Path, newValue interface{}) bool {
	// we always need a key or array position to work with
	if len(path) == 0 {
		return false
	}

	return d.setInternal(path, newValue)
}

func (d *Document) setInternal(path Path, newValue interface{}) bool {
	// when we have reached the root level,
	// replace our root element with the new data structure
	if len(path) == 0 {
		return d.setRoot(newValue)
	}

	leafKey := path.Tail()
	parentPath := path.Parent()
	target := interface{}(d.root)

	// check if the parent element exists;
	// create parent if missing

	if len(parentPath) > 0 {
		var exists bool

		target, exists = d.Get(parentPath)
		if !exists {
			if _, ok := leafKey.(int); ok {
				// this slice can be empty for now because we will extend it later
				if !d.setInternal(parentPath, []interface{}{}) {
					return false
				}
			} else if _, ok := leafKey.(string); ok {
				if !d.setInternal(parentPath, map[string]interface{}{}) {
					return false
				}
			} else {
				return false
			}

			target, _ = d.Get(parentPath)
		}
	}

	// Now we know that the parent element exists.

	if pos, ok := leafKey.(int); ok {
		// check if we are really in an array
		if array, ok := target.([]interface{}); ok {
			for i := len(array); i <= pos; i++ {
				array = append(array, nil)
			}

			array[pos] = newValue

			return d.setInternal(parentPath, array)
		}
	} else if key, ok := leafKey.(string); ok {
		// check if we are really in a map
		if m, ok := target.(map[string]interface{}); ok {
			m[key] = newValue
			return d.setInternal(parentPath, m)
		}

		if m, ok := target.(*yaml.MapSlice); ok {
			target = *m
		}

		if m, ok := target.(yaml.MapSlice); ok {
			return d.setInternal(parentPath, setValueInMapSlice(m, key, newValue))
		}
	}

	return false
}

func (d *Document) setRoot(newValue interface{}) bool {
	if asserted, ok := newValue.(yaml.MapSlice); ok {
		d.root = asserted
		return true
	}

	if asserted, ok := newValue.(*yaml.MapSlice); ok {
		d.root = *asserted
		return true
	}

	if asserted, ok := newValue.(map[string]interface{}); ok {
		d.root = makeMapSlice(asserted)
		return true
	}

	// attempted to set something that's not a map
	return false
}

func (d *Document) Append(path Path, newValue interface{}) bool {
	// we require maps at the root level, so the path cannot be empty
	if len(path) == 0 {
		return false
	}

	node, ok := d.Get(path)
	if !ok {
		return d.Set(path, []interface{}{newValue})
	}

	array, ok := node.([]interface{})
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

	if pos, ok := leafKey.(int); ok {
		if array, ok := parent.([]interface{}); ok {
			return d.setInternal(parentPath, removeArrayItem(array, pos))
		}
	} else if key, ok := leafKey.(string); ok {
		// check if we are really in a map
		if m, ok := parent.(map[string]interface{}); ok {
			delete(m, key)
			return d.setInternal(parentPath, m)
		}

		if m, ok := parent.(*yaml.MapSlice); ok {
			parent = *m
		}

		if m, ok := parent.(yaml.MapSlice); ok {
			return d.setInternal(parentPath, removeKeyFromMapSlice(m, key))
		}
	}

	return false
}

// Fill will set the value at the path to the newValue, but keeps any existing
// sub values intact.
func (d *Document) Fill(path Path, newValue interface{}) bool {
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
