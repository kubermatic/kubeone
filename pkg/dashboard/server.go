package dashboard

import (
	"embed"
	"html/template"
	"net/http"
	"time"

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

type dbData struct {
	ControlPlaneNodes  []node
	WorkerNodes        []node
	MachineDeployments []machineDeployment
}

type node struct {
	Name              string
	Status            string
	LastHeartbeatTime time.Duration
	Version           string
}

type machineDeployment struct {
	Namespace         string
	Name              string
	Replicas          int
	AvailableReplicas int
	Provider          string
	OS                string
	Kubelet           string
	Age               string
	Machines          *[]machine
}

type machine struct {
	Namespace string
	Name      string
	Provider  string
	OS        string
	Node      string
	Kubelet   string
	Address   string
	Age       string
}

func Serve(st *state.State) error {
	htmlTemplate, err := template.New("mainPage").Parse(indexTemplate)
	if err != nil {
		return err
	}

	if err := kubeconfig.BuildKubernetesClientset(st); err != nil {
		return err
	}

	http.Handle("/", dashboardHandler(st, htmlTemplate))
	http.Handle("/assets/", http.FileServerFS(assetsFS))

	st.Logger.Infoln("Visit http://localhost:8080")
	http.ListenAndServe("localhost:8080", nil)

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
		dbData, err := getDBData(st)
		if err != nil {
			return err
		}

		if err = htmlTemplate.Execute(wr, dbData); err != nil {
			return err
		}

		return nil
	})
}

func getDBData(state *state.State) (*dbData, error) {
	nodes, err := getNodes(state)
	if err != nil {
		return nil, err
	}

	machineDeployments, err := getMachineDeployments(state)
	if err != nil {
		return nil, err
	}

	dbData := dbData{
		ControlPlaneNodes:  nodes.ControlPlaneNodes,
		WorkerNodes:        nodes.WorkerNodes,
		MachineDeployments: machineDeployments,
	}

	return &dbData, nil
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

	var result nodesResult

	for _, currNode := range nodes.Items {
		_, isControlPlane := currNode.ObjectMeta.Labels["node-role.kubernetes.io/control-plane"]
		lastCondition := currNode.Status.Conditions[len(currNode.Status.Conditions)-1]

		aNode := node{
			Name:              currNode.Name,
			Status:            string(lastCondition.Type),
			LastHeartbeatTime: time.Now().Sub(lastCondition.LastHeartbeatTime.Time).Truncate(time.Second),
			Version:           currNode.Status.NodeInfo.KubeletVersion,
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
			Provider:          "TODO",
			OS:                "TODO",
			Kubelet:           "TODO",
			Age:               "TODO",
			Machines:          &machines,
		},
		)
	}
	return result, nil
}

func getMachines(state *state.State, md *clusterv1alpha1.MachineDeployment) ([]machine, error) {
	machines := clusterv1alpha1.MachineList{}
	if err := state.DynamicClient.List(state.Context, &machines); err != nil {
		return nil, err
	}

	result := []machine{}
	for _, currMachine := range machines.Items {
		result = append(result, machine{
			Namespace: currMachine.Namespace,
			Name:      currMachine.Name,
			Provider:  "TODO",
			OS:        "TODO",
			Node:      "TODO",
			Kubelet:   "TODO",
			Address:   "TODO",
			Age:       "TODO",
		})
	}

	return result, nil
}
