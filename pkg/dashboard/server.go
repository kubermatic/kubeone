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

package dashboard

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"time"

	"k8c.io/kubeone/pkg/clusterstatus"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/pkg/apis/cluster/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed index.html
var indexTemplate string

//go:embed assets/*
var assetsFS embed.FS

type dashboardData struct {
	ControlPlaneNodes  []node
	WorkerNodes        []node
	MachineDeployments []machineDeployment
}

type node struct {
	Name              string
	Status            string
	IsControlPlane    bool
	LastHeartbeatTime time.Duration
	Version           string
	EtcdOK            bool
	APIServerOK       bool
}

type machineDeployment struct {
	Namespace         string
	Name              string
	Replicas          int
	AvailableReplicas int
	Kubelet           string
	Age               time.Duration
	Machines          *[]machine
}

type machine struct {
	Namespace string
	Name      string
	Node      string
	Kubelet   string
	Address   string
	Age       time.Duration
	Deleted   bool
}

type nodesResult struct {
	ControlPlaneNodes []node
	WorkerNodes       []node
}

func Serve(st *state.State, port int) error {
	htmlTemplate, err := template.New("mainPage").Parse(indexTemplate)
	if err != nil {
		return err
	}

	if err := kubeconfig.BuildKubernetesClientset(st); err != nil {
		return err
	}

	http.Handle("/", dashboardHandler(st, htmlTemplate))
	http.Handle("/assets/", http.FileServerFS(assetsFS))

	st.Logger.Infoln(fmt.Sprintf("Visit http://localhost:%d to access UI", port))
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func httpHandleError(handler func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		if err := handler(wr, req); err != nil {
			http.Error(wr, err.Error(), 500)
		}
	})
}

func dashboardHandler(st *state.State, htmlTemplate *template.Template) http.Handler {
	return httpHandleError(func(wr http.ResponseWriter, _ *http.Request) error {
		dashboardData, err := getDashboardData(st)
		if err != nil {
			return err
		}

		if err = htmlTemplate.Execute(wr, dashboardData); err != nil {
			return err
		}

		return nil
	})
}

func getDashboardData(state *state.State) (*dashboardData, error) {
	nodes, err := getNodes(state)
	if err != nil {
		return nil, err
	}

	machineDeployments, err := getMachineDeployments(state)
	if err != nil {
		return nil, err
	}

	result := dashboardData{
		ControlPlaneNodes:  nodes.ControlPlaneNodes,
		WorkerNodes:        nodes.WorkerNodes,
		MachineDeployments: machineDeployments,
	}

	return &result, nil
}

func getNodes(s *state.State) (*nodesResult, error) {
	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{}
	if err := s.DynamicClient.List(s.Context, &nodes, &nodeListOpts); err != nil {
		return nil, fail.KubeClient(err, "listing nodes")
	}

	controlPlaneStatus, err := clusterstatus.Fetch(s, false)
	if err != nil {
		return nil, err
	}

	var result nodesResult

	for _, currNode := range nodes.Items {
		_, isControlPlane := currNode.ObjectMeta.Labels["node-role.kubernetes.io/control-plane"]
		lastCondition := currNode.Status.Conditions[len(currNode.Status.Conditions)-1]

		aNode := node{
			Name:              currNode.Name,
			IsControlPlane:    isControlPlane,
			Status:            string(lastCondition.Type),
			LastHeartbeatTime: time.Since(lastCondition.LastHeartbeatTime.Time).Truncate(time.Second),
			Version:           currNode.Status.NodeInfo.KubeletVersion,
			EtcdOK:            findNodeEtcd(controlPlaneStatus, currNode.Name),
			APIServerOK:       findNodeAPIServer(controlPlaneStatus, currNode.Name),
		}

		if isControlPlane {
			result.ControlPlaneNodes = append(result.ControlPlaneNodes, aNode)
		} else {
			result.WorkerNodes = append(result.WorkerNodes, aNode)
		}
	}

	return &result, nil
}

func getMachineDeployments(state *state.State) ([]machineDeployment, error) {
	if state.DynamicClient == nil {
		return nil, fail.NoKubeClient()
	}

	machineDeployments := clusterv1alpha1.MachineDeploymentList{}
	err := state.DynamicClient.List(
		state.Context,
		&machineDeployments,
		dynclient.InNamespace(metav1.NamespaceSystem),
	)
	if err != nil {
		return nil, err
	}

	result := []machineDeployment{}
	for i := range machineDeployments.Items {
		currMachineDeployment := &machineDeployments.Items[i]
		machines, err := getMachines(state, currMachineDeployment)
		if err != nil {
			return nil, err
		}

		result = append(result, machineDeployment{
			Namespace:         currMachineDeployment.Namespace,
			Name:              currMachineDeployment.Name,
			Replicas:          int(*currMachineDeployment.Spec.Replicas),
			AvailableReplicas: int(currMachineDeployment.Status.AvailableReplicas),
			Kubelet:           currMachineDeployment.Spec.Template.Spec.Versions.Kubelet,
			Age:               time.Since(currMachineDeployment.CreationTimestamp.Time).Truncate(time.Second),
			Machines:          &machines,
		},
		)
	}

	return result, nil
}

func getMachines(state *state.State, md *clusterv1alpha1.MachineDeployment) ([]machine, error) {
	// filter MachineSets owned by the MachineDeployment
	machineSets := clusterv1alpha1.MachineSetList{}
	if err := state.DynamicClient.List(state.Context, &machineSets); err != nil {
		return nil, err
	}

	filteredMachineSets := []clusterv1alpha1.MachineSet{}
	for _, currMS := range machineSets.Items {
		for _, currMSOR := range currMS.OwnerReferences {
			if md.UID == currMSOR.UID {
				filteredMachineSets = append(filteredMachineSets, currMS)
			}
		}
	}

	// filter Machines owned by one of the MachineSets owned by the MachineDeployment
	machines := clusterv1alpha1.MachineList{}
	if err := state.DynamicClient.List(state.Context, &machines); err != nil {
		return nil, err
	}

	filteredMachines := []clusterv1alpha1.Machine{}
	for _, currMachine := range machines.Items {
		for _, currMachineOR := range currMachine.OwnerReferences {
			for _, currMS := range filteredMachineSets {
				if currMachineOR.UID == currMS.UID {
					filteredMachines = append(filteredMachines, currMachine)
				}
			}
		}
	}

	result := []machine{}
	for i := range filteredMachines {
		currMachine := filteredMachines[i]
		address := getExternalIP(&currMachine)

		var nodeName string
		if noderef := currMachine.Status.NodeRef; noderef != nil {
			nodeName = noderef.Name
		}

		result = append(result, machine{
			Namespace: currMachine.Namespace,
			Name:      currMachine.Name,
			Node:      nodeName,
			Kubelet:   currMachine.Spec.Versions.Kubelet,
			Address:   address,
			Age:       time.Since(currMachine.CreationTimestamp.Time).Truncate(time.Second),
			Deleted:   !currMachine.ObjectMeta.DeletionTimestamp.IsZero(),
		})
	}

	return result, nil
}

func findNodeEtcd(nodes []clusterstatus.NodeStatus, search string) bool {
	for _, cp := range nodes {
		if cp.NodeName == search {
			return cp.Etcd
		}
	}

	return false
}

func findNodeAPIServer(nodes []clusterstatus.NodeStatus, search string) bool {
	for _, cp := range nodes {
		if cp.NodeName == search {
			return cp.APIServer
		}
	}

	return false
}

func getExternalIP(machine *clusterv1alpha1.Machine) string {
	addressIndex := slices.IndexFunc(machine.Status.Addresses, func(a corev1.NodeAddress) bool { return a.Type == "ExternalIP" })
	if addressIndex >= 0 {
		return machine.Status.Addresses[addressIndex].Address
	}

	// TODO what if no external ip address
	return ""
}
