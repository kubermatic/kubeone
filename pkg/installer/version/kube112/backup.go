package kube112

import (
	"github.com/kubermatic/kubeone/pkg/installer/util"
)

func backup(ctx *util.Context) error {
	if ctx.BackupFile != "" {
		ctx.Logger.Infoln("Creating local backup…")
		if err := ctx.Configuration.Backup(ctx.BackupFile); err != nil {
			// do not stop in case of failed backups, the user can
			// always create the backup themselves if needed
			ctx.Logger.Warnf("Failed to create backup: %v", err)
		}
	}
	return nil
}
