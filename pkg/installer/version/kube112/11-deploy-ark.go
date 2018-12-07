package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/ark"
)

func deployArk(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ config.HostConfig, conn ssh.Connection) error {
		arkConfig, err := ark.ArkManifest(ctx.Cluster)
		if err != nil {
			return fmt.Errorf("failed to create ark configuration: %v", err)
		}
		ctx.Configuration.AddFile("ark.yaml", arkConfig)
		err = ctx.Configuration.UploadTo(conn, fmt.Sprintf("%s/ark", ctx.WorkDir))
		if err != nil {
			return fmt.Errorf("failed to upload ark configuration: %v", err)
		}

		ctx.Logger.Infoln("Deploying Ark…")

		cmd, err := util.MakeShellCommand(`
# Create Ark prerequisites, configuration, and deploy Ark
sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f ./{{ .WORK_DIR }}/ark/ark.yaml
`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})
		if err != nil {
			return err
		}

		_, _, _, err = util.RunCommand(conn, cmd, ctx.Verbose)
		return err
	})
}
