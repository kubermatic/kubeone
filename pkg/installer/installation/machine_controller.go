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

package installation

import (
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"
)

func installMachineController(ctx *util.Context) error {
	if !*ctx.Cluster.MachineController.Deploy {
		ctx.Logger.Info("Skipping machine-controller deployment because it was disabled in configuration.")
		return nil
	}

	ctx.Logger.Infoln("Installing machine-controller…")
	if err := machinecontroller.Deploy(ctx); err != nil {
		return err
	}

	ctx.Logger.Infoln("Installing machine-controller webhooks…")
	if err := machinecontroller.DeployWebhookConfiguration(ctx); err != nil {
		return err
	}

	return nil
}
