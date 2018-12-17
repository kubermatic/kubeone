package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/ark"
)

func deployArk(ctx *util.Context) error {
	if !ctx.Cluster.Backup.Enabled() {
		ctx.Logger.Info("Skipping Ark deployment because no backup provider was configured.")
		return nil
	}

	return ctx.RunTaskOnLeader(func(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
		arkConfig, err := ark.Manifest(ctx.Cluster)
		if err != nil {
			return fmt.Errorf("failed to create Ark configuration: %v", err)
		}

		ctx.Configuration.AddFile("ark.yaml", arkConfig)
		err = ctx.Configuration.UploadTo(conn, fmt.Sprintf("%s/ark", ctx.WorkDir))
		if err != nil {
			return fmt.Errorf("failed to upload Ark configuration: %v", err)
		}

		ctx.Logger.Infoln("Deploying Ark…")

		_, _, err = ctx.Runner.Run(`sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f "{{ .WORK_DIR }}/ark/ark.yaml"`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})

		return err
	})
}
