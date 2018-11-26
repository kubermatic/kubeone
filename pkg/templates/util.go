package templates

import (
	"bytes"
	"encoding/json"
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// KubernetesToYAML properly encodes a list of resources as YAML.
// Straight up encoding as YAML leaves us with a non-standard data
// structure. Going through JSON eliminates the extra fields and
// keys and results in what you would expect to see.
// This function takes a slice of items to support creating a
// multi-document YAML string (separated with "---" between each
// item).
func KubernetesToYAML(data []interface{}) (string, error) {
	var buffer bytes.Buffer

	yamlEncoder := yaml.NewEncoder(&buffer)

	for _, item := range data {
		var (
			encoded []byte
			tmp     interface{}
			err     error
		)

		if str, ok := item.(string); ok {
			err = yaml.Unmarshal([]byte(str), &tmp)
			if err != nil {
				return "", fmt.Errorf("failed to decode object as YAML: %v", err)
			}
		} else {
			encoded, err = json.Marshal(item)
			if err != nil {
				return "", fmt.Errorf("failed to encode object as JSON: %v", err)
			}

			err = json.Unmarshal(encoded, &tmp)
			if err != nil {
				return "", fmt.Errorf("failed to read JSON: %v", err)
			}
		}

		err = yamlEncoder.Encode(tmp)
		if err != nil {
			return "", fmt.Errorf("failed to encode object as YAML: %v", err)
		}
	}

	return buffer.String(), nil
}
