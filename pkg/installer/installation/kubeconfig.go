package installation

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func copyKubeconfig(ctx *util.Context) error {
	return ctx.RunTaskOnAllNodes(func(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Copying Kubeconfig to home directory…")

		_, _, err := ctx.Runner.Run(`
mkdir -p $HOME/.kube/
sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -u) $HOME/.kube/config
`, util.TemplateVariables{})
		if err != nil {
			return err
		}

		return nil
	}, true)
}

func buildKubernetesClientset(ctx *util.Context) error {
	ctx.Logger.Infoln("Building Kubernetes clientset…")

	// connect to leader
	leader, err := ctx.Cluster.Leader()
	if err != nil {
		return err
	}
	connector := ssh.NewConnector()

	conn, err := connector.Connect(*leader)
	if err != nil {
		return fmt.Errorf("failed to connect to leader: %v", err)
	}
	defer conn.Close()

	// get the kubeconfig
	kubeconfig, _, _, err := conn.Exec("sudo cat /etc/kubernetes/admin.conf")
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %v", err)
	}

	c, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("unable to build config from kubeconfig bytes: %v", err)
	}

	ctx.Clientset, err = kubernetes.NewForConfig(c)
	if err != nil {
		return fmt.Errorf("unable to build kubernetes clientset: %v", err)
	}

	ctx.APIExtensionClientset, err = apiextensionsclientset.NewForConfig(c)
	if err != nil {
		return fmt.Errorf("unable to build apiextension-apiserver clientset: %v", err)
	}

	return nil
}
