package tasks

import (
	"time"
)

type ClusterSettleTask struct {
	Duration time.Duration
}

func (t *ClusterSettleTask) Execute(ctx *Context) error {
	ctx.Logger.Infoln("Letting the cluster settle down…")
	time.Sleep(t.Duration)

	return nil
}
