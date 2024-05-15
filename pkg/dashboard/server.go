package dashboard

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"

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
	Version string
}

type machineDeployment struct {
	Name     string
	Replicas int
}

func Serve(state *state.State) error {

	state.Logger.Infoln("HUST: dasboard serve")

	mainPage, err := template.New("mainPage").Parse(indexTemplate)
	if err != nil {
		return err
	}
	state.Logger.Infoln("HUST: parsed templ")

	dbData, err := getDBData(state)
	if err != nil {
		return err
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		err := mainPage.Execute(writer, dbData)
		if err != nil {
			fmt.Printf("Error on serving dashboard %s", err)
		}
	})

	state.Logger.Infoln("Visit http://localhost:8080")
	http.ListenAndServe(":8080", nil)
	return nil
}

func getDBData(state *state.State) (*dbData, error) {

	state.Logger.Infoln("HUST: get DB State")
	err := kubeconfig.BuildKubernetesClientset(state)
	if err != nil {
		return nil, err
	}

	state.Logger.Infoln("HUST: build state")

	nodes, _ := getNodes(state)
	state.Logger.Infoln("HUST: got nodes")

	state.Logger.Warnln(nodes)

	dbData := dbData{
		Nodes:              &[]node{{Name: "n1", Version: "v1"}, {Name: "n2", Version: "v2"}},
		MachineDeployments: &[]machineDeployment{},
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

	// result := []node{}

	// for _, currNode := range nodes.Items {
	// 	result = append(result, node{
	// 		NodeName: currNode.Name,
	// 		Version:  currNode.Status.NodeInfo.KubeletVersion,
	// 	})
	// 	s.Logger.Infof("HUST, GETNODES: %v", currNode.Name)
	// }

	// s.Logger.Infof("HUST, GOT NODES: %v", result)

	return nil, nil

}
