package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
)

func installMachineController(ctx *util.Context) error {
	return ctx.RunTaskOnLeader(func(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Creating machine-controller certificate…")

		config, err := machinecontroller.WebhookConfiguration(ctx.Cluster, ctx.Configuration)
		if err != nil {
			return err
		}

		ctx.Configuration.AddFile("machine-controller-webhook.yaml", []byte(config))
		err = ctx.Configuration.UploadTo(conn, ctx.WorkDir)
		if err != nil {
			return fmt.Errorf("failed to upload: %v", err)
		}

		ctx.Logger.Infoln("Installing machine-controller…")

		_, _, err = ctx.Runner.Run(`
sudo kubectl apply -f ./{{ .WORK_DIR }}/machine-controller.yaml
sudo kubectl apply -f ./{{ .WORK_DIR }}/machine-controller-webhook.yaml
`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})

		return err
	})
}
