package features

import (
	"github.com/pkg/errors"

	kubeadmv1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta1"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/util"
)

// Activate configured features.
// Installing CRDs, creating policies and so on
func Activate(ctx *util.Context) error {
	// TODO: more features flags to go
	if err := installKubeSystemPSP(ctx.Cluster.Features.EnablePSP, ctx); err != nil {
		return errors.Wrap(err, "failed to install PodSecurityPolicy")
	}

	return nil
}

// KubeadmActivate features in cluster config
func KubeadmActivate(featuresCfg config.Features, clusterConfig *kubeadmv1beta1.ClusterConfiguration) {
	if clusterConfig.APIServer.ExtraArgs == nil {
		clusterConfig.APIServer.ExtraArgs = make(map[string]string)
	}

	// TODO: more features flags to go
	activateKubeadmPSP(featuresCfg.EnablePSP, clusterConfig)
}
