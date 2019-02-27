package util

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

// TemplateVariables is a render context for templates
type TemplateVariables map[string]interface{}

// MakeShellCommand render text template with given `variables` render-context
func MakeShellCommand(cmd string, variables TemplateVariables) (string, error) {
	tpl, err := template.New("base").Parse(cmd)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse shell script")
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "failed to render shell script")
	}

	return buf.String(), nil
}
