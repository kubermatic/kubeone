package util

import (
	"bytes"
	"fmt"
	"text/template"
)

// TemplateVariables is a render context for templates
type TemplateVariables map[string]interface{}

// MakeShellCommand render text template with given `variables` render-context
func MakeShellCommand(cmd string, variables TemplateVariables) (string, error) {
	tpl, err := template.New("base").Parse(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to parse shell script: %v", err)
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render shell script: %v", err)
	}

	return buf.String(), nil
}
