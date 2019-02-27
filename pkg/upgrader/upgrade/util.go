package upgrade

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1types "k8s.io/client-go/kubernetes/typed/core/v1"
)

func determineHostname(ctx *util.Context) error {
	ctx.Logger.Infoln("Determine hostname…")
	return ctx.RunTaskOnAllNodes(func(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
		stdout, _, err := ctx.Runner.Run("hostname -f", nil)
		if err != nil {
			return err
		}

		node.Hostname = stdout
		return nil
	}, true)
}

func determineOS(ctx *util.Context) error {
	ctx.Logger.Infoln("Determine operating system…")
	return ctx.RunTaskOnAllNodes(func(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
		osID, _, err := ctx.Runner.Run("source /etc/os-release && echo -n $ID", nil)
		if err != nil {
			return err
		}

		node.OperatingSystem = osID
		return nil
	}, true)
}

func labelNode(nodeClient corev1types.NodeInterface, host *config.HostConfig) error {
	node, err := nodeClient.Get(host.Hostname, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var modified bool
	label := map[string]string{
		labelUpgradeLock: "",
	}
	mergeStringMap(&modified, &node.ObjectMeta.Labels, label)
	if !modified {
		return nil
	}

	_, err = nodeClient.Update(node)
	return err
}

func unlabelNode(nodeClient corev1types.NodeInterface, host *config.HostConfig) error {
	node, err := nodeClient.Get(host.Hostname, metav1.GetOptions{})
	if err != nil {
		return err
	}

	delete(node.ObjectMeta.Labels, labelUpgradeLock)
	_, err = nodeClient.Update(node)
	return err
}

// mergeStringMap merges two string maps into destination string map
func mergeStringMap(modified *bool, destination *map[string]string, required map[string]string) {
	if *destination == nil {
		*destination = map[string]string{}
	}

	for k, v := range required {
		if destinationV, ok := (*destination)[k]; !ok || destinationV != v {
			(*destination)[k] = v
			*modified = true
		}
	}
}
