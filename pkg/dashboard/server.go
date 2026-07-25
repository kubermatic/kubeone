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
	"context"
	"embed"
	"fmt"
	"net/http"
	"time"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed assets/*
var assetsFS embed.FS

func Serve(st *state.State, port int) error {
	if err := kubeconfig.BuildKubernetesClientset(st); err != nil {
		return err
	}

	http.Handle("/", dashboardHandler(st))
	http.Handle("/assets/", http.FileServerFS(assetsFS))
	http.Handle("POST /scale", scaleHandler(st))
	http.Handle("POST /rollout", rolloutHandler(st))
	http.Handle("POST /delete-machine", deleteMachineHandler(st))
	http.Handle("POST /pods", podsHandler(st))
	http.Handle("POST /delete-pod", deletePodHandler(st))

	st.Logger.Infoln(fmt.Sprintf("Visit http://localhost:%d to access UI", port))
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}

func getPodsForNode(ctx context.Context, dynClient dynclient.Client, nodeName string) ([]pod, error) {
	podList := corev1.PodList{}
	listOpts := dynclient.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.nodeName", nodeName),
	}
	if err := dynClient.List(ctx, &podList, &listOpts); err != nil {
		return nil, fail.KubeClient(err, "listing pods")
	}

	pods := make([]pod, 0, len(podList.Items))
	for i := range podList.Items {
		p := &podList.Items[i]
		readyContainers := 0
		totalRestarts := int32(0)
		for _, cs := range p.Status.ContainerStatuses {
			if cs.Ready {
				readyContainers++
			}
			totalRestarts += cs.RestartCount
		}

		pods = append(pods, pod{
			Name:      p.Name,
			Namespace: p.Namespace,
			Status:    string(p.Status.Phase),
			Ready:     fmt.Sprintf("%d/%d", readyContainers, len(p.Spec.Containers)),
			Restarts:  totalRestarts,
			Age:       time.Since(p.CreationTimestamp.Time).Truncate(time.Second),
		})
	}

	return pods, nil
}
