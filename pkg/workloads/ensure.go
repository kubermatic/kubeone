package workloads

import (
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/workloads/addons"
	"k8c.io/kubeone/pkg/workloads/localhelm"
)

func Ensure(st *state.State) error {
	if len(st.Cluster.Workloads) == 0 {
		return nil
	}

	tmpKubeConf, cleanupFn, err := kubeconfig.File(st)
	if err != nil {
		return err
	}
	defer cleanupFn()

	for _, wk := range st.Cluster.Workloads {
		switch {
		case wk.Addon != nil:
			if err := addons.EnsureAddonByName(st, wk.Addon.Name); err != nil {
				return err
			}
		case wk.HelmRelease != nil:
			helmSettings := localhelm.NewHelmSettings(st.Verbose)
			helmCfg, err := localhelm.NewActionConfiguration(helmSettings.Debug)
			if err != nil {
				return err
			}

			if err := localhelm.DeployRelease(st, *wk.HelmRelease, helmSettings, tmpKubeConf, helmCfg); err != nil {
				return err
			}
		}
	}

	return nil
}
