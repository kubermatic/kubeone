package installation

import (
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates/canal"
)

func applyCanalCNI(ctx *util.Context) error {
	ctx.Logger.Infoln("Applying canal CNI plugin…")
	return canal.Deploy(ctx)
}
