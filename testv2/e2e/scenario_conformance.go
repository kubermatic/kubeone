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

package e2e

import (
	"io"
	"testing"
)

type scenarioConformance struct {
	name                 string
	manifestTemplatePath string
	versions             []string
	infra                Infra
}

func (scenario *scenarioConformance) SetInfra(infrastructure Infra) {
	scenario.infra = infrastructure
}

func (scenario *scenarioConformance) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *scenarioConformance) GenerateTests(output io.Writer, testType GeneratorType, cfg ProwConfig) error {
	return nil
}

func (scenario *scenarioConformance) Run(t *testing.T) {
	t.Helper()

	install := scenarioInstall{
		name:                 scenario.name,
		manifestTemplatePath: scenario.manifestTemplatePath,
		infra:                scenario.infra,
		versions:             scenario.versions,
	}

	install.install(t)
	scenario.test(t)
}

func (scenario *scenarioConformance) test(t *testing.T) {
	// TODO: add sonobuoy test launch
}
