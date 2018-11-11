package kube112

import (
	"time"

	"github.com/kubermatic/kubeone/pkg/installer/util"
)

func wait(ctx *util.Context, t time.Duration) error {
	ctx.Logger.Infoln("Letting the cluster settle down…")
	time.Sleep(t)

	return nil
}
