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

package dashboard

import (
	"errors"
	"net/http"
	"time"

	"github.com/angelofallars/htmx-go"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
)

func httpHandleError(handler func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		if err := handler(wr, req); err != nil {
			if uiErr, ok := errors.AsType[*uiError](err); ok {
				http.Error(wr, uiErr.Message, uiErr.Code)
			} else {
				http.Error(wr, err.Error(), http.StatusInternalServerError)
			}
		}
	})
}

func dashboardHandler() http.Handler {
	return httpHandleError(func(wr http.ResponseWriter, req *http.Request) error {
		return Layout().Render(req.Context(), wr)
	})
}

func controlPlaneNodesHandler(st *state.State) http.Handler {
	return httpHandleError(func(wr http.ResponseWriter, req *http.Request) error {
		nodes, err := getNodes(st)
		if err != nil {
			return err
		}

		return NodesTable(nodes.ControlPlaneNodes).Render(req.Context(), wr)
	})
}

func workerNodesHandler(st *state.State) http.Handler {
	return httpHandleError(func(wr http.ResponseWriter, req *http.Request) error {
		nodes, err := getNodes(st)
		if err != nil {
			return err
		}

		return WorkerNodesSection(nodes.WorkerNodes).Render(req.Context(), wr)
	})
}

func machineDeploymentsHandler(st *state.State) http.Handler {
	return httpHandleError(func(wr http.ResponseWriter, req *http.Request) error {
		mds, err := getMachineDeployments(st)
		if err != nil {
			return err
		}

		return MachineDeploymentsTable(mds).Render(req.Context(), wr)
	})
}

func scaleHandler(st *state.State) http.Handler {
	return httpHandleError(func(wr http.ResponseWriter, req *http.Request) error {
		form, err := parseAndValidateForm[scaleForm](req)
		if err != nil {
			return err
		}

		md := clusterv1alpha1.MachineDeployment{}
		key := dynclient.ObjectKey{Namespace: form.Namespace, Name: form.Name}
		if err := st.DynamicClient.Get(req.Context(), key, &md); err != nil {
			return fail.KubeClient(err, "getting MachineDeployment")
		}

		current := int32(0)
		if md.Spec.Replicas != nil {
			current = *md.Spec.Replicas
		}

		patch := dynclient.MergeFrom(md.DeepCopy())

		switch form.Direction {
		case "up":
			current++
		case "down":
			if current > 0 {
				current--
			}
		default:
			return &uiError{
				Message: "direction must be 'up' or 'down'",
				Code:    http.StatusBadRequest,
			}
		}

		md.Spec.Replicas = &current
		if err := st.DynamicClient.Patch(req.Context(), &md, patch); err != nil {
			return fail.KubeClient(err, "patching MachineDeployment")
		}

		if htmx.IsHTMX(req) {
			mds, err := getMachineDeployments(st)
			if err != nil {
				return err
			}

			return MachineDeploymentsTable(mds).Render(req.Context(), wr)
		}

		http.Redirect(wr, req, "/", http.StatusSeeOther)

		return nil
	})
}

func rolloutHandler(st *state.State) http.Handler {
	return httpHandleError(func(wr http.ResponseWriter, req *http.Request) error {
		form, err := parseAndValidateForm[namespaceNameForm](req)
		if err != nil {
			return err
		}

		md := clusterv1alpha1.MachineDeployment{}
		key := dynclient.ObjectKey{Namespace: form.Namespace, Name: form.Name}
		if err := st.DynamicClient.Get(req.Context(), key, &md); err != nil {
			return fail.KubeClient(err, "getting MachineDeployment")
		}

		patch := dynclient.MergeFrom(md.DeepCopy())

		if md.Spec.Template.Labels == nil {
			md.Spec.Template.Labels = make(map[string]string)
		}
		md.Spec.Template.Labels["forced-restart"] = time.Now().UTC().Format("2006-01-02T15.04.05Z")

		if err := st.DynamicClient.Patch(req.Context(), &md, patch); err != nil {
			return fail.KubeClient(err, "patching MachineDeployment")
		}

		if htmx.IsHTMX(req) {
			mds, err := getMachineDeployments(st)
			if err != nil {
				return err
			}

			return MachineDeploymentsTable(mds).Render(req.Context(), wr)
		}

		http.Redirect(wr, req, "/", http.StatusSeeOther)

		return nil
	})
}

func deleteMachineHandler(st *state.State) http.Handler {
	return httpHandleError(func(wr http.ResponseWriter, req *http.Request) error {
		form, err := parseAndValidateForm[namespaceNameForm](req)
		if err != nil {
			return err
		}

		m := clusterv1alpha1.Machine{}
		key := dynclient.ObjectKey{Namespace: form.Namespace, Name: form.Name}
		if err := st.DynamicClient.Get(req.Context(), key, &m); err != nil {
			return fail.KubeClient(err, "getting Machine")
		}

		if err := st.DynamicClient.Delete(req.Context(), &m); err != nil {
			return fail.KubeClient(err, "deleting Machine")
		}

		if htmx.IsHTMX(req) {
			mds, err := getMachineDeployments(st)
			if err != nil {
				return err
			}

			return MachineDeploymentsTable(mds).Render(req.Context(), wr)
		}

		http.Redirect(wr, req, "/", http.StatusSeeOther)

		return nil
	})
}
