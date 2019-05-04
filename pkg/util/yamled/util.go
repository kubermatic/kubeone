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

func untypedSliceHas(haystack []interface{}, needle interface{}) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}

	return false
}

func mapSliceKeys(m yaml.MapSlice) []interface{} {
	keys := make([]interface{}, 0)
	for _, item := range m {
		keys = append(keys, item.Key)
	}

	return keys
}

func mapSliceGet(haystack yaml.MapSlice, key interface{}) (interface{}, bool) {
	for _, item := range haystack {
		if item.Key == key {
			return item.Value, true
		}
	}

	return nil, false
}
