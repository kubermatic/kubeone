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
	if err := installKubeSystemPSP(ctx.Cluster.Features.EnablePodSecurityPolicy, ctx); err != nil {
		return errors.Wrap(err, "failed to install PodSecurityPolicy")
	}

	return nil
}

// UpdateKubeadmClusterConfiguration update additional config options in the kubeadm's
// v1beta1.ClusterConfiguration according to enabled features
func UpdateKubeadmClusterConfiguration(featuresCfg config.Features, clusterConfig *kubeadmv1beta1.ClusterConfiguration) {
	if clusterConfig.APIServer.ExtraArgs == nil {
		clusterConfig.APIServer.ExtraArgs = make(map[string]string)
	}

	activateKubeadmPSP(featuresCfg.EnablePodSecurityPolicy, clusterConfig)
}
