package externalccm

import (
	"context"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Ensure external CCM deployen if Provider.External
func Ensure(ctx *util.Context) error {
	if !ctx.Cluster.Provider.External {
		return nil
	}

	ctx.Logger.Info("Ensure external CCM is up to date")

	switch ctx.Cluster.Provider.Name {
	case config.ProviderNameHetzner:
		return ensureHetzner(ctx)
	case config.ProviderNameDigitalOcean:
		return ensureDigitalOcean(ctx)
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

func mutateDeploymentWithVersionCheck(want *semver.Constraints) func(obj runtime.Object) error {
	return func(obj runtime.Object) error {
		dep, ok := obj.(*appsv1.Deployment)
		if !ok {
			return errors.Errorf("unknown object type %T passed", obj)
		}

		if dep.ObjectMeta.CreationTimestamp.IsZero() {
			// let it create deployment
			return nil
		}

		if len(dep.Spec.Template.Spec.Containers) != 1 {
			return errors.New("unable to choose a CCM container, as number of containers > 1")
		}

		imageSpec := strings.SplitN(dep.Spec.Template.Spec.Containers[0].Image, ":", 2)
		if len(imageSpec) != 2 {
			return errors.New("unable to grab CCM image version")
		}

		existing, err := semver.NewVersion(imageSpec[1])
		if err != nil {
			return errors.Wrap(err, "failed to parse deployed CCM version")
		}

		if !want.Check(existing) {
			return errors.New("newer version deployed, skipping")
		}

		// OK to update the deployment
		return nil
	}
}
