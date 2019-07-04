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

package templates

import (
	"bytes"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
)

// KubernetesToYAML properly encodes a list of resources as YAML.
// Straight up encoding as YAML leaves us with a non-standard data
// structure. Going through JSON eliminates the extra fields and
// keys and results in what you would expect to see.
// This function takes a slice of items to support creating a
// multi-document YAML string (separated with "---" between each
// item).
func KubernetesToYAML(data []runtime.Object) (string, error) {
	var buffer bytes.Buffer

	for _, item := range data {
		var (
			encodedItem []byte
			err         error
		)

		encodedItem, err = yaml.Marshal(item)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal item")
		}

		if _, err := buffer.Write(encodedItem); err != nil {
			return "", errors.Wrap(err, "failed to write into buffer")
		}

		if _, err := buffer.WriteString("\n---\n"); err != nil {
			return "", errors.Wrap(err, "failed to write into buffer")
		}
	}

	return buffer.String(), nil
}
