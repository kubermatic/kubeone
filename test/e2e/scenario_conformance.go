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
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

type scenarioConformance struct {
	Name                 string
	HumanReadableName    string
	ManifestTemplatePath string
	SkipTests            string

	versions []string
	infra    Infra
}

func (scenario scenarioConformance) GetHumanReadableName() string {
	return scenario.HumanReadableName
}

func (scenario *scenarioConformance) SetInfra(infrastructure Infra) {
	scenario.infra = infrastructure
}

func (scenario *scenarioConformance) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *scenarioConformance) FetchVersions() error {
	for i := range scenario.versions {
		latestVer, err := latestUpstreamVersion(scenario.versions[i])
		if err != nil {
			return err
		}

		scenario.versions[i] = latestVer
	}

	return nil
}

func (scenario *scenarioConformance) GenerateTests(wr io.Writer, generatorType GeneratorType, cfg ProwConfig) error {
	install := scenarioInstall{
		Name:                 scenario.Name,
		ManifestTemplatePath: scenario.ManifestTemplatePath,
		infra:                scenario.infra,
		versions:             scenario.versions,
	}

	return install.GenerateTests(wr, generatorType, cfg)
}

func (scenario *scenarioConformance) Run(ctx context.Context, t *testing.T) {
	if err := makeBin("build").Run(); err != nil {
		t.Fatalf("building kubeone: %v", err)
	}

	t.Logf("Testing Kubernetes version(s): %s", strings.Join(scenario.versions, ","))

	install := scenarioInstall{
		Name:                 scenario.Name,
		ManifestTemplatePath: scenario.ManifestTemplatePath,
		infra:                scenario.infra,
		versions:             scenario.versions,
	}

	install.install(ctx, t)
	scenario.test(ctx, t)
}

func (scenario *scenarioConformance) test(ctx context.Context, t *testing.T) {
	var k1Opts []kubeoneBinOpts

	if *kubeoneVerboseFlag {
		k1Opts = append(k1Opts, withKubeoneVerbose)
	}

	if *credentialsFlag != "" {
		k1Opts = append(k1Opts, withKubeoneCredentials(*credentialsFlag))
	}

	k1 := newKubeoneBin(
		scenario.infra.terraform.path,
		renderManifest(t,
			scenario.ManifestTemplatePath,
			manifestData{
				VERSION: scenario.versions[0],
				INFRA:   scenario.infra.name,
			},
		),
		k1Opts...,
	)

	// launch kubeone proxy, to have a HTTPS proxy through the SSH tunnel
	// to open access to the kubeapi behind the bastion host
	proxyCtx, killProxy := context.WithCancel(ctx)
	proxyURL, waitK1, err := k1.AsyncProxy(proxyCtx)
	if err != nil {
		t.Fatalf("starting kubeone proxy: %v", err)
	}
	defer func() {
		waitErr := waitK1()
		if waitErr != nil {
			t.Logf("wait kubeone proxy: %v", waitErr)
		}
	}()
	defer killProxy()

	// let kubeone proxy start and open the port
	time.Sleep(5 * time.Second)
	t.Logf("kubeone proxy is running on %s", proxyURL)

	kubeconfigPath, err := k1.kubeconfigPath(t.TempDir())
	if err != nil {
		t.Fatalf("fetching kubeconfig: %v", err)
	}

	stopProtokol, err := scenario.infra.protokol.Start(ctx, kubeconfigPath, proxyURL)
	if err != nil {
		t.Fatalf("starting protokol: %v", err)
	}
	defer stopProtokol()

	waitKubeOneNodesReady(ctx, t, k1)

	client := dynamicClientRetriable(t, k1)
	cpTests := newCloudProviderTests(client, scenario.infra.Provider())
	cpTests.runWithCleanup(t)

	sonobuoyRun(ctx, t, k1, sonobuoyConformance, scenario.SkipTests, proxyURL)
}
