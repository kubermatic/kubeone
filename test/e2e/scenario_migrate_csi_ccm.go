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
	"testing"
	"time"

	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type scenarioMigrateCSIAndCCM struct {
	Name                    string
	OldManifestTemplatePath string
	NewManifestTemplatePath string

	versions []string
	infra    Infra
}

func (scenario scenarioMigrateCSIAndCCM) Title() string { return titleize(scenario.Name) }

func (scenario *scenarioMigrateCSIAndCCM) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *scenarioMigrateCSIAndCCM) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *scenarioMigrateCSIAndCCM) GenerateTests(wr io.Writer, generatorType GeneratorType, cfg ProwConfig) error {
	install := scenarioInstall{
		Name:     scenario.Name,
		infra:    scenario.infra,
		versions: scenario.versions,
	}

	return install.GenerateTests(wr, generatorType, cfg)
}

func (scenario *scenarioMigrateCSIAndCCM) Run(ctx context.Context, t *testing.T) {
	if err := makeBin("build").Run(); err != nil {
		t.Fatalf("building kubeone: %v", err)
	}

	install := scenarioInstall{
		Name:                 scenario.Name,
		ManifestTemplatePath: scenario.OldManifestTemplatePath,
		infra:                scenario.infra,
		versions:             scenario.versions,
	}

	install.install(ctx, t)
	k1Old := install.kubeone(t)
	waitKubeOneNodesReady(ctx, t, k1Old)

	client := dynamicClientRetriable(t, k1Old)
	cpTests := newCloudProviderTests(client, scenario.infra.Provider())
	defer cpTests.cleanUp(t)
	cpTests.run(t)

	// change to new manifest
	install.ManifestTemplatePath = scenario.NewManifestTemplatePath
	k1New := install.kubeone(t)

	scenario.migrate(ctx, t, k1New, false)
	waitKubeOneNodesReady(ctx, t, k1New)

	labelNodesSkipEviction(t, client)
	scenario.forceRolloutMachinedeployments(t, client)
	waitMachinesHasNodes(t, k1New, client)
	waitKubeOneNodesReady(ctx, t, k1New)

	scenario.migrate(ctx, t, k1New, true)
	waitKubeOneNodesReady(ctx, t, k1New)

	cpTests.validateStatefulSetReadiness(t)
	cpTests.validateLoadBalancerReadiness(t)
}

func (scenario *scenarioMigrateCSIAndCCM) migrate(ctx context.Context, t *testing.T, k1 *kubeoneBin, complete bool) {
	args := []string{"migrate", "to-ccm-csi", "--auto-approve"}
	if complete {
		args = append(args, "--complete")
	}

	if err := k1.build(args...).BuildCmd(ctx).Run(); err != nil {
		t.Fatalf("migrating CCM/CSI: %v", err)
	}
}

func (scenario *scenarioMigrateCSIAndCCM) forceRolloutMachinedeployments(t *testing.T, client ctrlruntimeclient.Client) {
	var machinedeployments clusterv1alpha1.MachineDeploymentList
	if err := client.List(context.Background(), &machinedeployments, ctrlruntimeclient.InNamespace(metav1.NamespaceSystem)); err != nil {
		t.Error(err)
	}

	for _, md := range machinedeployments.Items {
		mdOld := md.DeepCopy()
		mdNew := md

		if mdNew.Spec.Template.Spec.ObjectMeta.Annotations == nil {
			mdNew.Spec.Template.Spec.ObjectMeta.Annotations = map[string]string{}
		}

		mdNew.Spec.Template.Spec.ObjectMeta.Annotations["forceRestart"] = time.Now().String()

		err := retryFn(func() error {
			return client.Patch(context.Background(), &mdNew, ctrlruntimeclient.MergeFrom(mdOld))
		})
		if err != nil {
			t.Fatalf("forcing machineDeployment %q to rollout: %v", ctrlruntimeclient.ObjectKeyFromObject(&mdNew), err)
		}
	}

	delay := 10 * time.Second
	t.Logf("Waiting %s to give machine-controller time to start rolling-out MachineDeployments", delay)
	time.Sleep(delay)
}
