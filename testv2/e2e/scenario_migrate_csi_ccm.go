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

func (scenario *scenarioMigrateCSIAndCCM) Run(t *testing.T) {
	install := scenarioInstall{
		Name:                 scenario.Name,
		ManifestTemplatePath: scenario.OldManifestTemplatePath,
		infra:                scenario.infra,
		versions:             scenario.versions,
	}

	install.install(t)
	k1Old := install.kubeone(t)
	waitKubeOneNodesReady(t, k1Old)

	// change to new manifest
	install.ManifestTemplatePath = scenario.NewManifestTemplatePath
	k1New := install.kubeone(t)

	scenario.migrate(t, k1New, false)
	waitKubeOneNodesReady(t, k1New)

	var (
		client ctrlruntimeclient.Client
		err    error
	)

	err = retryFn(func() error {
		client, err = k1New.DynamicClient()

		return err
	})
	if err != nil {
		t.Fatalf("initializing dynamic client: %v", err)
	}

	scenario.forceRolloutMachinedeployments(t, client)
	waitKubeOneNodesReady(t, k1New)

	scenario.migrate(t, k1New, true)
	waitKubeOneNodesReady(t, k1New)
}

func (scenario *scenarioMigrateCSIAndCCM) migrate(t *testing.T, k1 *kubeoneBin, complete bool) {
	args := []string{"migrate", "to-ccm-csi"}
	if complete {
		args = append(args, "--complete")
	}

	if err := k1.run(args...); err != nil {
		t.Fatalf("migrating CCM/CSI: %v", err)
	}
}

func (scenario *scenarioMigrateCSIAndCCM) forceRolloutMachinedeployments(t *testing.T, client ctrlruntimeclient.Client) {
	var machinedeployments clusterv1alpha1.MachineDeploymentList
	if err := client.List(context.Background(), &machinedeployments); err != nil {
		t.Error(err)
	}

	for _, md := range machinedeployments.Items {
		mdOrig := md.DeepCopy()
		md := md

		md.Spec.Template.Spec.ObjectMeta.Annotations["forceRestart"] = time.Now().String()

		patch := ctrlruntimeclient.MergeFrom(&md)

		err := retryFn(func() error {
			return client.Patch(context.Background(), mdOrig, patch)
		})
		if err != nil {
			t.Fatalf("forcing machineDeployment %q to rollout: %v", ctrlruntimeclient.ObjectKeyFromObject(&md), err)
		}
	}
}