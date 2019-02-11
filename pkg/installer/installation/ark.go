package installation

import (
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates/ark"
)

func deployArk(ctx *util.Context) error {
	if !ctx.Cluster.Backup.Enabled() {
		ctx.Logger.Info("Skipping Ark deployment because no backup provider was configured.")
		return nil
	}

	ctx.Logger.Infoln("Deploying Ark…")

	return ark.Deploy(ctx)
}
