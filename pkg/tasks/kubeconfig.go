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
	"fmt"
	"os"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
)

const (
	superAdminConfPath = "/etc/kubernetes/super-admin.conf"
)

func saveKubeconfig(st *state.State) error {
	st.Logger.Info("Downloading kubeconfig...")

	kc, err := kubeconfig.Download(st)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s-kubeconfig", st.Cluster.Name)
	err = os.WriteFile(fileName, kc, 0o600)

	return fail.Runtime(err, "saving kubeconfig file")
}

func removeSuperKubeconfig(st *state.State) error {
	st.Logger.Info("Removing %s...", superAdminConfPath)

	host, err := st.Cluster.Leader()
	if err != nil {
		return err
	}

	conn, err := st.Executor.Open(host)
	if err != nil {
		return err
	}

	// The constant superAdminConfPath is NOT used here for safety reasons, to avoid
	// accidental and catastrophic changes in the future.
	//
	// We don't care if file doesn't exist.
	conn.Exec("sudo rm -rf /etc/kubernetes/super-admin.conf") //nolint:errcheck

	return nil
}
