/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package externalccm

import (
	"context"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Ensure external CCM deployen if Provider.External
func Ensure(ctx *util.Context) error {
	if !ctx.Cluster.CloudProvider.External {
		return nil
	}

	ctx.Logger.Info("Ensure external CCM is up to date")

	switch ctx.Cluster.CloudProvider.Name {
	case kubeoneapi.CloudProviderNameHetzner:
		return ensureHetzner(ctx)
	case kubeoneapi.CloudProviderNameDigitalOcean:
		return ensureDigitalOcean(ctx)
	case kubeoneapi.CloudProviderNamePacket:
		return ensurePacket(ctx)
	default:
		ctx.Logger.Infof("External CCM for %q not yet supported, skipping", ctx.Cluster.CloudProvider.Name)
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
