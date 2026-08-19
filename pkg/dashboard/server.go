/*
Copyright 2026 The KubeOne Authors.

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

//go:generate go tool github.com/a-h/templ/cmd/templ generate

package dashboard

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
)

//go:embed assets/*
var assetsFS embed.FS

func Serve(st *state.State, port int) error {
	if err := kubeconfig.BuildKubernetesClientset(st); err != nil {
		return err
	}

	http.Handle("/", dashboardHandler())
	http.Handle("/assets/", http.FileServerFS(assetsFS))
	http.Handle("GET /nodes/control-plane", controlPlaneNodesHandler(st))
	http.Handle("GET /nodes/worker", workerNodesHandler(st))
	http.Handle("GET /machine-deployments", machineDeploymentsHandler(st))
	http.Handle("POST /scale", scaleHandler(st))
	http.Handle("POST /rollout", rolloutHandler(st))
	http.Handle("POST /delete-machine", deleteMachineHandler(st))

	st.Logger.Infoln(fmt.Sprintf("Visit http://localhost:%d to access UI", port))
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}
