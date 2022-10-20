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
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"
	"github.com/kubermatic/machine-controller/pkg/jsonutil"
	providerconfigtypes "github.com/kubermatic/machine-controller/pkg/providerconfig/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	kubeoneStableVersion = "1.5.1" //nolint:deadcode,varcheck
	kubeoneStableBaseRef = "release/v1.5"
)

type scenarioUpgrade struct {
	Name                 string
	ManifestTemplatePath string

	versions []string
	infra    Infra
}

func (scenario scenarioUpgrade) Title() string { return titleize(scenario.Name) }

func (scenario *scenarioUpgrade) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *scenarioUpgrade) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *scenarioUpgrade) Run(ctx context.Context, t *testing.T) {
	if err := makeBin("build").Run(); err != nil {
		t.Fatalf("building kubeone: %v", err)
	}
	if err := makeBinWithPath(filepath.Clean("../../../kubeone-stable/"), "build").Run(); err != nil {
		t.Fatalf("building kubeone-stable: %v", err)
	}

	install := &scenarioInstall{
		Name:                 scenario.Name,
		ManifestTemplatePath: scenario.ManifestTemplatePath,
		infra:                scenario.infra,
		versions:             []string{scenario.versions[0]},
		kubeonePath:          mustAbsolutePath("../../../kubeone-stable/dist/kubeone"),
	}

	install.install(ctx, t)
	scenario.upgrade(ctx, t)
	scenario.test(ctx, t)
}

func (scenario *scenarioUpgrade) kubeone(t *testing.T, version string) *kubeoneBin {
	var k1Opts []kubeoneBinOpts

	if *kubeoneVerboseFlag {
		k1Opts = append(k1Opts, withKubeoneVerbose)
	}

	if *credentialsFlag != "" {
		k1Opts = append(k1Opts, withKubeoneCredentials(*credentialsFlag))
	}

	return newKubeoneBin(
		scenario.infra.terraform.path,
		renderManifest(t,
			scenario.ManifestTemplatePath,
			manifestData{
				VERSION: version,
			},
		),
		k1Opts...,
	)
}

func (scenario *scenarioUpgrade) upgrade(ctx context.Context, t *testing.T) {
	// NB: Due to changed node selectors between Kubernetes 1.23 and 1.24, it's
	// important to run apply with KubeOne 1.5 before upgrading the cluster,
	// otherwise upgrade might get stuck due to pods not able to get
	// rescheduled.
	k1 := scenario.kubeone(t, scenario.versions[0])

	waitKubeOneNodesReady(ctx, t, k1)

	if err := k1.Apply(ctx); err != nil {
		t.Fatalf("kubeone apply failed: %v", err)
	}

	k1 = scenario.kubeone(t, scenario.versions[1])
	if err := k1.Apply(ctx); err != nil {
		t.Fatalf("kubeone apply failed: %v", err)
	}
}

func (scenario *scenarioUpgrade) test(ctx context.Context, t *testing.T) {
	k1 := scenario.kubeone(t, scenario.versions[1])

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

	client := dynamicClientRetriable(t, k1)

	labelNodesSkipEviction(t, client)
	scenario.upgradeMachineDeployments(t, client, scenario.versions[1])
	waitMachinesHasNodes(t, k1, client)
	waitKubeOneNodesReady(ctx, t, k1)

	cpTests := newCloudProviderTests(client, scenario.infra.Provider())
	cpTests.runWithCleanup(t)

	// sonobuoyRun(t, k1, sonobuoyConformanceLite, proxyURL)
	sonobuoyRun(ctx, t, k1, sonobuoyQuick, proxyURL)
}

func (scenario *scenarioUpgrade) GenerateTests(wr io.Writer, generatorType GeneratorType, cfg ProwConfig) error {
	if len(scenario.versions) != 2 {
		return fmt.Errorf("expected only 2 versions")
	}

	type upgradeFromTo struct {
		From string
		To   string
	}

	up := upgradeFromTo{
		From: scenario.versions[0],
		To:   scenario.versions[1],
	}

	type templateData struct {
		Infra       string
		Scenario    string
		FromVersion string
		ToVersion   string
		TestTitle   string
	}

	var (
		data     []templateData
		prowJobs []ProwJob
	)

	testTitle := fmt.Sprintf("Test%s%sFrom%s_To%s",
		titleize(scenario.infra.name),
		scenario.Title(),
		titleize(up.From),
		titleize(up.To),
	)

	data = append(data, templateData{
		TestTitle:   testTitle,
		Infra:       scenario.infra.name,
		Scenario:    scenario.Name,
		FromVersion: up.From,
		ToVersion:   up.To,
	})

	cfg.Environ = scenario.infra.environ

	prowJobs = append(prowJobs,
		newProwJob(
			pullProwJobName(scenario.infra.name, scenario.Name, "from", up.From, "to", up.To),
			scenario.infra.labels,
			testTitle,
			cfg,
			kubeoneStableProwExtraRefs(kubeoneStableBaseRef),
		),
	)

	switch generatorType {
	case GeneratorTypeGo:
		tpl, err := template.New("").Parse(upgradeScenarioTemplate)
		if err != nil {
			return err
		}

		return tpl.Execute(wr, data)
	case GeneratorTypeYAML:
		buf, err := yaml.Marshal(prowJobs)
		if err != nil {
			return err
		}

		n, err := wr.Write(buf)
		if err != nil {
			return err
		}

		if n != len(buf) {
			return fmt.Errorf("wrong number of bytes written, expected %d, wrote %d", len(buf), n)
		}

		return nil
	}

	return fmt.Errorf("unknown generator type %d", generatorType)
}

func (scenario *scenarioUpgrade) upgradeMachineDeployments(t *testing.T, client ctrlruntimeclient.Client, kubeletVersion string) {
	var machinedeployments clusterv1alpha1.MachineDeploymentList
	if err := client.List(context.Background(), &machinedeployments, ctrlruntimeclient.InNamespace(metav1.NamespaceSystem)); err != nil {
		t.Error(err)
	}

	for _, md := range machinedeployments.Items {
		mdOld := md.DeepCopy()
		mdNew := md

		mdNew.Spec.Template.Spec.Versions.Kubelet = kubeletVersion

		providerConfig := providerconfigtypes.Config{}
		if err := jsonutil.StrictUnmarshal(mdNew.Spec.Template.Spec.ProviderSpec.Value.Raw, &providerConfig); err != nil {
			t.Fatalf("decoding provider config: %v", err)
		}

		// KubeOne 1.4 has been using cloud-init for Flatcar, but it doesn't work
		// with OSM, so we have to switch to Ignition.
		if providerConfig.OperatingSystem == providerconfigtypes.OperatingSystemFlatcar {
			var osConfig map[string]interface{}
			if err := json.Unmarshal(providerConfig.OperatingSystemSpec.Raw, &osConfig); err != nil {
				t.Fatalf("decoding operating system config: %v", err)
			}

			if v, ok := osConfig["provisioningUtility"]; ok && v == "cloud-init" {
				osConfig["provisioningUtility"] = "ignition"

				var err error
				providerConfig.OperatingSystemSpec.Raw, err = json.Marshal(osConfig)
				if err != nil {
					t.Fatalf("updating operating system config: %v", err)
				}

				b, err := json.Marshal(providerConfig)
				if err != nil {
					t.Fatalf("marshalling new provider config")
				}

				mdNew.Spec.Template.Spec.ProviderSpec.Value.Raw = b
			}
		}

		err := retryFn(func() error {
			return client.Patch(context.Background(), &mdNew, ctrlruntimeclient.MergeFrom(mdOld))
		})
		if err != nil {
			t.Fatalf("upgrading machineDeployment %q: %v", ctrlruntimeclient.ObjectKeyFromObject(&mdNew), err)
		}
	}

	delay := 10 * time.Second
	t.Logf("Waiting %s to give machine-controller time to start rolling-out MachineDeployments", delay)
	time.Sleep(delay)
}

const upgradeScenarioTemplate = `
{{- range . }}
func {{ .TestTitle }}(t *testing.T) {
	ctx := NewSignalContext()
	infra := Infrastructures["{{ .Infra }}"]
	scenario := Scenarios["{{ .Scenario }}"]
	scenario.SetInfra(infra)
	scenario.SetVersions("{{ .FromVersion }}", "{{ .ToVersion }}")
	scenario.Run(ctx, t)
}
{{ end -}}
`
