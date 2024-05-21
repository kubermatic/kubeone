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

	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"

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
	OS                string
	Kubelet           string
	Age               time.Duration
	Machines          *[]machine
}

type machine struct {
	Namespace string
	Name      string
	OS        string
	Node      string
	Kubelet   string
	Address   string
	Age       time.Duration
	Deleted   bool
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
	http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil)

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
	return httpHandleError(func(wr http.ResponseWriter, req *http.Request) error {
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

	dashboardData := dashboardData{
		ControlPlaneNodes:  nodes.ControlPlaneNodes,
		WorkerNodes:        nodes.WorkerNodes,
		MachineDeployments: machineDeployments,
	}

	return &dashboardData, nil
}

type nodesResult struct {
	ControlPlaneNodes []node
	WorkerNodes       []node
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
			LastHeartbeatTime: time.Now().Sub(lastCondition.LastHeartbeatTime.Time).Truncate(time.Second),
			Version:           currNode.Status.NodeInfo.KubeletVersion,
			EtcdOK:            findNodeEtcd(controlPlaneStatus, currNode.Name),
			APIServerOK:       findNodeApiServer(controlPlaneStatus, currNode.Name),
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
	for _, md := range machineDeployments.Items {
		machines, err := getMachines(state, &md)
		if err != nil {
			return nil, err
		}
		result = append(result, machineDeployment{
			Namespace:         md.Namespace,
			Name:              md.Name,
			Replicas:          int(*md.Spec.Replicas),
			AvailableReplicas: int(md.Status.AvailableReplicas),
			OS:                "TODO",
			Kubelet:           md.Spec.Template.Spec.Versions.Kubelet,
			Age:               time.Now().Sub(md.CreationTimestamp.Time).Truncate(time.Second),
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
	for _, currMachine := range filteredMachines {

		address, err := getExternalIp(&currMachine)
		if err != nil {
			return nil, err
		}

		result = append(result, machine{
			Namespace: currMachine.Namespace,
			Name:      currMachine.Name,
			OS:        "TODO",
			Node:      currMachine.Status.NodeRef.Name,
			Kubelet:   currMachine.Spec.Versions.Kubelet,
			Address:   address,
			Age:       time.Now().Sub(currMachine.CreationTimestamp.Time).Truncate(time.Second),
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

func findNodeApiServer(nodes []clusterstatus.NodeStatus, search string) bool {
	for _, cp := range nodes {
		if cp.NodeName == search {
			return cp.APIServer
		}
	}

	return false
}

func getExternalIp(machine *clusterv1alpha1.Machine) (string, error) {
	addressIndex := slices.IndexFunc(machine.Status.Addresses, func(a corev1.NodeAddress) bool { return a.Type == "ExternalIP" })
	if addressIndex >= 0 {
		return machine.Status.Addresses[addressIndex].Address, nil
	}
	// TODO what if no external ip address
	return "", nil
}
