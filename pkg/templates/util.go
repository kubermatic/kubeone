package templates

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
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

	for _, item := range data {
		var (
			encodedItem []byte
			err         error
		)

		if str, ok := item.(string); ok {
			encodedItem = []byte(strings.TrimSpace(str))
		} else {
			encodedItem, err = yaml.Marshal(item)
		}

		if err != nil {
			return "", fmt.Errorf("failed to marshal item: %v", err)
		}
		if _, err := buffer.Write(encodedItem); err != nil {
			return "", fmt.Errorf("failed to write into buffer: %v", err)
		}
		if _, err := buffer.WriteString("\n---\n"); err != nil {
			return "", fmt.Errorf("failed to write into buffer: %v", err)
		}
	}

	return buffer.String(), nil
}

// MergeStringMap merges two string maps into destination string map
func MergeStringMap(modified *bool, destination *map[string]string, required map[string]string) {
	if *destination == nil {
		*destination = map[string]string{}
	}

	for k, v := range required {
		if destinationV, ok := (*destination)[k]; !ok || destinationV != v {
			(*destination)[k] = v
			*modified = true
		}
	}
}
