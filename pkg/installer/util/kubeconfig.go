package util

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/pkg/errors"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// DownloadKubeconfig downloads Kubeconfig over SSH
func DownloadKubeconfig(cluster *config.Cluster) ([]byte, error) {
	// connect to leader
	leader, err := cluster.Leader()
	if err != nil {
		return nil, err
	}
	connector := ssh.NewConnector()

	conn, err := connector.Connect(*leader)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// get the kubeconfig
	kubeconfig, _, _, err := conn.Exec("sudo cat /etc/kubernetes/admin.conf")
	if err != nil {
		return nil, err
	}

	return []byte(kubeconfig), nil
}

// BuildKubernetesClientset builds core kubernetes and apiextensions clientsets
func BuildKubernetesClientset(ctx *Context) error {
	ctx.Logger.Infoln("Building Kubernetes clientset…")

	kubeconfig, err := DownloadKubeconfig(ctx.Cluster)
	if err != nil {
		return errors.Wrap(err, "unable to download kubeconfig")
	}

	ctx.RESTConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "unable to build config from kubeconfig bytes")
	}

	ctx.Clientset, err = kubernetes.NewForConfig(ctx.RESTConfig)
	if err != nil {
		return errors.Wrap(err, "unable to build kubernetes clientset")
	}

	ctx.APIExtensionClientset, err = apiextensionsclientset.NewForConfig(ctx.RESTConfig)
	if err != nil {
		return errors.Wrap(err, "unable to build apiextension-apiserver clientset")
	}

	return nil
}
