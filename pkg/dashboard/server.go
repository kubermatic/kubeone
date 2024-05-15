package dashboard

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"

	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed index.html
var indexTemplate string

type dbData struct {
	Nodes              *[]node
	MachineDeployments *[]machineDeployment
}

type node struct {
	Name    string
	Status  string
	Roles   string
	Age     string
	Version string
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
}

func Serve(state *state.State) error {

	htmlTemplate, err := template.New("mainPage").Parse(indexTemplate)
	if err != nil {
		return err
	}

	dbData, err := getDBData(state)
	if err != nil {
		return err
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		err := htmlTemplate.Execute(writer, dbData)
		if err != nil {
			fmt.Printf("Error on serving dashboard %s", err)
		}
	})


	http.Handle(
        "/assets",
        http.StripPrefix(
            "/assets",
            http.FileServer(http.Dir("./assets")),
        ),
    )

	state.Logger.Infoln("Visit http://localhost:8080")
	http.ListenAndServe(":8080", nil)
	return nil
	// TODO
}

func getDBData(state *state.State) (*dbData, error) {

	err := kubeconfig.BuildKubernetesClientset(state)
	if err != nil {
		return nil, err
	}

	nodes, err := getNodes(state)
	if err != nil {
		return nil, err
	}

	machineDeployments, err := getMachineDeployments(state)
	if err != nil {
		return nil, err
	}

	state.Logger.Infof("MDs %v", machineDeployments)

	dbData := dbData{
		Nodes:              &nodes,
		MachineDeployments: &machineDeployments,
	}
	return &dbData, nil
}

func getNodes(s *state.State) ([]node, error) {
	if s.DynamicClient == nil {
		return nil, fail.NoKubeClient()
	}

	// Get node list
	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{}
	if err := s.DynamicClient.List(s.Context, &nodes, &nodeListOpts); err != nil {
		return nil, fail.KubeClient(err, "listing nodes")
	}

	result := []node{}
	for _, currNode := range nodes.Items {
		// TODO is this safe
		// TODO does this deliver really the last state?
		// TODO nil checks
		role, err := getNodeRole(&currNode)
		if err != nil {
			return nil, err
		}

		lastCondition := currNode.Status.Conditions[len(currNode.Status.Conditions)-1]
		result = append(result, node{
			Name:    currNode.Name,
			Status:  string(lastCondition.Type),
			Roles:   role,
			Age:     lastCondition.LastHeartbeatTime.Format("2006-01-02 15:04:05"),
			Version: currNode.Status.NodeInfo.KubeletVersion,
		})
	}

	return result, nil
}

func getNodeRole(node *corev1.Node) (string, error) {
	// TODO is this smart enough?
	_, ok := node.ObjectMeta.Labels["node-role.kubernetes.io/control-plane"]
	if ok {
		return "control-plane", nil
	}
	return "<none>", nil
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
		state.Logger.Infof("MD %v", md)
		result = append(result, machineDeployment{
			Namespace:         md.Namespace,
			Name:              md.Name,
			Replicas:          int(*md.Spec.Replicas),
			AvailableReplicas: int(md.Status.AvailableReplicas),
			Provider:          "TODO",
			OS:                "TODO",
			Kubelet:           "TODO",
			Age:               "TODO",
		},
		)
	}
	return result, nil
}
