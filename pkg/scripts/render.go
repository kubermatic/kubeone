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

package scripts

import (
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(cmd string, variables map[string]interface{}) (string, error) {
	tpl, err := template.New("base").Parse(cmd)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse script template")
	}

	var buf strings.Builder
	buf.WriteString(`set -xeu pipefail`)
	buf.WriteString("\n\n")
	buf.WriteString(`export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"`)
	buf.WriteString("\n\n")

	if err := tpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "failed to render script template")
	}

	return buf.String(), nil
}
