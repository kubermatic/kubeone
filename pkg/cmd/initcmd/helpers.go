/*
Copyright 2022 The KubeOne Authors.

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

package initcmd

import (
	"sort"

	"github.com/AlecAivazis/survey/v2"
)

func cloudProviderSelectOptions() []string {
	cp := []string{}

	for _, v := range ValidProviders {
		cp = append(cp, v.title)
	}

	sort.Strings(cp)

	return cp
}

func cloudProviderForSelectedOption(name string) (string, initProvider) {
	for k, v := range ValidProviders {
		if v.title == name {
			return k, v
		}
	}

	// This should never happen
	return "", initProvider{}
}

func addonsSelectOptions(providerNone bool) []string {
	addons := []string{}
	if !providerNone {
		addons = append(addons, addonClusterAutoscaler)
	}
	addons = append(addons, addonBackupsRestic)

	return addons
}

func osForSelectedOption(provider initProvider, index int) string {
	for _, v := range provider.optionalTFVars {
		if v.Name == "os" || v.Name == "worker_os" {
			return v.Choices[index].Value
		}
	}

	// This should never happen
	return ""
}

func providerTFVarsQuestions(cloudProvider string) []*survey.Question {
	_, provider := cloudProviderForSelectedOption(cloudProvider)

	qs := []*survey.Question{}
	for _, v := range provider.requiredTFVars {
		qs = append(qs, terraformVariableQuestion(v, true))
	}
	for _, v := range provider.optionalTFVars {
		qs = append(qs, terraformVariableQuestion(v, false))
	}

	return qs
}

func terraformVariableQuestion(tfVar terraformVariable, required bool) *survey.Question {
	q := &survey.Question{
		Name: tfVar.Name,
	}
	if required {
		q.Validate = survey.Required
	}

	switch {
	case tfVar.Choices != nil:
		choices := []string{}
		for _, c := range tfVar.Choices {
			choices = append(choices, c.Name)
		}
		q.Prompt = &survey.Select{
			Message: tfVar.Description,
			Default: tfVar.DefaultValue,
			Options: choices,
		}
	default:
		q.Prompt = &survey.Input{
			Message: tfVar.Description,
			Default: tfVar.DefaultValue,
		}
	}

	return q
}
