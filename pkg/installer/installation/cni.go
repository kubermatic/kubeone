package installation

import (
	"github.com/kubermatic/kubeone/pkg/templates/canal"
	"github.com/kubermatic/kubeone/pkg/util"
)

func applyCanalCNI(ctx *util.Context) error {
	ctx.Logger.Infoln("Applying canal CNI plugin…")
	return canal.Deploy(ctx)
}
