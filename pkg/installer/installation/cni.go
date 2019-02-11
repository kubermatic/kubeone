package installation

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates/canal"
)

func applyCNI(ctx *util.Context, cni string) error {
	switch cni {
	case "canal":
		return applyCanalCNI(ctx)
	default:
		return fmt.Errorf("unknown CNI plugin selected")
	}
}

func applyCanalCNI(ctx *util.Context) error {
	ctx.Logger.Infoln("Applying canal CNI plugin…")
	return canal.Deploy(ctx)
}
