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

package tasks

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"
)

func setupKubernetesBinaries(s *state.State, node kubeoneapi.HostConfig, params scripts.Params) error {
	return runOnOS(s, node.OperatingSystem, map[kubeoneapi.OperatingSystemName]runOnOSFn{
		kubeoneapi.OperatingSystemNameAmazon:     kubernetesBinariesAmazonLinux(params),
		kubeoneapi.OperatingSystemNameCentOS:     kubernetesBinariesRHELLike(params),
		kubeoneapi.OperatingSystemNameDebian:     kubernetesBinariesDeb(params),
		kubeoneapi.OperatingSystemNameFlatcar:    kubernetesBinariesFlatcar(params),
		kubeoneapi.OperatingSystemNameRHEL:       kubernetesBinariesRHELLike(params),
		kubeoneapi.OperatingSystemNameRockyLinux: kubernetesBinariesRHELLike(params),
		kubeoneapi.OperatingSystemNameUbuntu:     kubernetesBinariesDeb(params),
	})
}

func upgradeKubernetesBinaries(s *state.State, node kubeoneapi.HostConfig, params scripts.Params) error {
	params.Upgrade = true

	return setupKubernetesBinaries(s, node, params)
}

func kubernetesBinariesDeb(params scripts.Params) func(*state.State) error {
	return func(s *state.State) error {
		cmd, err := scripts.DebScript(s.Cluster, params)
		if err != nil {
			return err
		}

		_, _, err = s.Runner.RunRaw(cmd)

		return fail.SSH(err, "%s", params.String())
	}
}

func kubernetesBinariesFlatcar(params scripts.Params) func(*state.State) error {
	return func(s *state.State) error {
		cmd, err := scripts.FlatcarScript(s.Cluster, params)
		if err != nil {
			return err
		}

		_, _, err = s.Runner.RunRaw(cmd)

		return fail.SSH(err, "%s", params.String())
	}
}

func kubernetesBinariesRHELLike(params scripts.Params) func(*state.State) error {
	return func(s *state.State) error {
		cmd, err := scripts.RHELLikeScript(s.Cluster, params)
		if err != nil {
			return err
		}

		_, _, err = s.Runner.RunRaw(cmd)

		return fail.SSH(err, "%s", params.String())
	}
}

func kubernetesBinariesAmazonLinux(params scripts.Params) func(*state.State) error {
	return func(s *state.State) error {
		cmd, err := scripts.AmazonLinuxScript(s.Cluster, params)
		if err != nil {
			return err
		}

		_, _, err = s.Runner.RunRaw(cmd)

		return fail.SSH(err, "%s", params.String())
	}
}
