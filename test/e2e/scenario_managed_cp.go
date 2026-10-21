/*
Copyright 2026 The KubeOne Authors.

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
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"text/template"

	"sigs.k8s.io/yaml"
)

const (
	// envClusterName and envSSHPublicKeyFile are set by test/go-test-e2e.sh
	envClusterName      = "TF_VAR_cluster_name"
	envSSHPublicKeyFile = "TF_VAR_ssh_public_key_file"
)

// scenarioManagedControlPlane provisions a cluster whose control plane VMs
// (and load balancer) are created by KubeOne from .controlPlane.nodeSets,
// without Terraform.
type scenarioManagedControlPlane struct {
	Name              string
	HumanReadableName string
	// ManifestTemplatePaths maps the provider name to the manifest template
	ManifestTemplatePaths map[string]string

	versions []string
	infra    Infra
}

func (scenario scenarioManagedControlPlane) GetHumanReadableName() string {
	return scenario.HumanReadableName
}

func (scenario scenarioManagedControlPlane) Title() string { return titleize(scenario.Name) }

func (scenario *scenarioManagedControlPlane) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *scenarioManagedControlPlane) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *scenarioManagedControlPlane) FetchVersions() error {
	ctx := context.Background()

	for i := range scenario.versions {
		latestVer, err := latestUpstreamVersion(ctx, scenario.versions[i])
		if err != nil {
			return err
		}

		scenario.versions[i] = latestVer
	}

	return nil
}

func (scenario *scenarioManagedControlPlane) Run(ctx context.Context, t *testing.T) {
	if !scenario.infra.managedControlPlane {
		t.Fatalf("infrastructure %q doesn't support managed control plane", scenario.infra.name)
	}

	if len(scenario.versions) != 1 {
		t.Fatalf("only 1 version is expected to be set, got %v", scenario.versions)
	}

	if err := makeBin("build").Run(); err != nil {
		t.Fatalf("building kubeone: %v", err)
	}

	t.Logf("Testing Kubernetes version(s): %s", strings.Join(scenario.versions, ","))

	clusterName := os.Getenv(envClusterName)
	if clusterName == "" {
		t.Fatalf("%s is required", envClusterName)
	}

	cloud, err := newManagedControlPlaneCloud(scenario.infra.Provider(), clusterName)
	if err != nil {
		t.Fatal(err)
	}

	k1 := scenario.kubeone(t, clusterName)

	// Cleanups run in reverse order: the janitor runs after kubeone reset, to
	// delete anything that reset left behind (e.g. if apply failed midway)
	t.Cleanup(func() {
		if err := retryFn(func() error {
			return cloud.Janitor(context.Background())
		}); err != nil {
			t.Errorf("janitor failed, resources of cluster %q might be leaked: %v", clusterName, err)
		}
	})

	t.Cleanup(func() {
		// Force only if the test already failed, kube-apiserver might be
		// unreachable then (the janitor deletes the workers). Otherwise reset
		// must work as users run it, so its failures fail the test.
		force := t.Failed()
		if err := retryFn(func() error { return k1.ResetManagedControlPlane(force) }); err != nil {
			t.Errorf("kubeone reset failed: %v", err)
		}
	})

	if err := k1.Apply(ctx); err != nil {
		t.Fatalf("kubeone apply failed: %v", err)
	}

	scenario.verifyControlPlane(ctx, t, k1, cloud)

	// apply must be idempotent, no new VMs or load balancers
	if err := k1.Apply(ctx); err != nil {
		t.Fatalf("second kubeone apply failed: %v", err)
	}

	scenario.verifyControlPlane(ctx, t, k1, cloud)

	runClusterTests(ctx, t, k1, scenario.infra)
}

func (scenario *scenarioManagedControlPlane) verifyControlPlane(ctx context.Context, t *testing.T, k1 *kubeoneBin, cloud managedControlPlaneCloud) {
	t.Helper()

	manifest, err := k1.ClusterManifest()
	if err != nil {
		t.Fatalf("rendering cluster manifest: %v", err)
	}

	expectedVMs := 0
	for _, ns := range manifest.ControlPlane.NodeSets {
		expectedVMs += ns.Replicas
	}

	if len(manifest.ControlPlane.Hosts) != expectedVMs {
		t.Fatalf("expected %d control plane hosts, got %d", expectedVMs, len(manifest.ControlPlane.Hosts))
	}

	vms, err := cloud.ControlPlaneVMs(ctx)
	if err != nil {
		t.Fatalf("counting control plane VMs: %v", err)
	}
	if vms != expectedVMs {
		t.Fatalf("expected %d control plane VMs at the cloud provider, got %d", expectedVMs, vms)
	}

	lbIPs, err := cloud.LoadBalancers(ctx)
	if err != nil {
		t.Fatalf("listing load balancers: %v", err)
	}

	endpoint := manifest.APIEndpoint.Host

	if len(lbIPs) != 1 {
		t.Fatalf("expected exactly one load balancer, got %v", lbIPs)
	}
	if !slices.Contains(lbIPs, endpoint) {
		t.Fatalf("expected endpoint %q to be the load balancer IP %v", endpoint, lbIPs)
	}
}

func (scenario *scenarioManagedControlPlane) kubeone(t *testing.T, clusterName string) *kubeoneBin {
	templatePath, ok := scenario.ManifestTemplatePaths[scenario.infra.Provider()]
	if !ok {
		t.Fatalf("no manifest template for provider %q", scenario.infra.Provider())
	}

	keyPath := os.Getenv(envSSHPublicKeyFile)
	if keyPath == "" {
		t.Fatalf("%s is required", envSSHPublicKeyFile)
	}

	sshPublicKey, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("reading SSH public key: %v", err)
	}

	k1Opts := []kubeoneBinOpts{
		withKubeoneBin(getKubeoneDistPath()),
	}

	if *kubeoneVerboseFlag {
		k1Opts = append(k1Opts, withKubeoneVerbose)
	}

	if *credentialsFlag != "" {
		k1Opts = append(k1Opts, withKubeoneCredentials(*credentialsFlag))
	}

	k1 := newKubeoneBin(
		t.TempDir(),
		renderManifest(t,
			templatePath,
			manifestData{
				VERSION:      scenario.versions[0],
				INFRA:        scenario.infra.name,
				ClusterName:  clusterName,
				SSHPublicKey: strings.TrimSpace(string(sshPublicKey)),
			},
		),
		k1Opts...,
	)

	// no Terraform output, hosts are provisioned by KubeOne
	k1.tfjsonPath = ""

	return k1
}

func (scenario *scenarioManagedControlPlane) GenerateTests(wr io.Writer, generatorType GeneratorType, cfg ProwConfig) error {
	type templateData struct {
		TestTitle string
		Infra     string
		Scenario  string
		Version   string
	}

	version := scenario.versions[0]
	testTitle := fmt.Sprintf("Test%s%s%s",
		titleize(scenario.infra.name),
		scenario.Title(),
		titleize(version),
	)

	data := []templateData{
		{
			TestTitle: testTitle,
			Infra:     scenario.infra.name,
			Scenario:  scenario.Name,
			Version:   version,
		},
	}

	cfg.Environ = scenario.infra.environ

	prowJobs := []ProwJob{
		newProwJob(
			PullProwJobName(scenario.infra.name, scenario.Name, version),
			scenario.infra.labels,
			testTitle,
			cfg,
			nil,
		),
	}

	switch generatorType {
	case GeneratorTypeGo:
		tpl, err := template.New("").Parse(installScenarioTemplate)
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
