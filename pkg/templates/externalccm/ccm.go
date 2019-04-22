package externalccm

import (
	"context"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/util"

	"k8s.io/apimachinery/pkg/runtime"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Ensure external CCM when
func Ensure(ctx *util.Context) error {
	if !ctx.Cluster.Provider.External {
		return nil
	}

	ctx.Logger.Info("Ensure external CCM is up to date")

	switch ctx.Cluster.Provider.Name {
	case config.ProviderNameHetzner:
		return ensureHetzner(ctx)
	default:
		ctx.Logger.Infof("External CCM for %q not yet supported, skipping", ctx.Cluster.Provider.Name)
		return nil
	}
}

func simpleCreateOrUpdate(ctx context.Context, client dynclient.Client, obj runtime.Object) error {
	okFunc := func(runtime.Object) error { return nil }
	_, err := controllerutil.CreateOrUpdate(ctx, client, obj, okFunc)
	return err
}
