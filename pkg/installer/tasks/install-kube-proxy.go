package tasks

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type InstallKubeProxyTask struct{}

func (t *InstallKubeProxyTask) Execute(ctx *Context) error {
	ctx.Logger.Infoln("Installing kube-proxy…")

	node := ctx.Manifest.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node":   node.PublicAddress,
		"master": true,
	})

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	logger.Infoln("Running kubectl…")
	command, err := makeShellCommand(`
set -xeu pipefail

mkdir -p ~/.kube
sudo cp /etc/kubernetes/admin.conf ~/.kube/config
sudo chown -R $(id -u):$(id -g) ~/.kube

kubectl apply -f ./{{ .WORK_DIR }}/kube-flannel.yaml

kubectl -n kube-system get configmap kube-proxy -o yaml > kube-proxy-configmap.yaml
sed -i -e 's#server:.*#server: https://{{ .IP_ADDRESS }}:6443#g' kube-proxy-configmap.yaml
kubectl delete -f kube-proxy-configmap.yaml
kubectl create -f kube-proxy-configmap.yaml
kubectl -n kube-system delete pod -l k8s-app=kube-proxy
`, templateVariables{
		"WORK_DIR":   ctx.WorkDir,
		"IP_ADDRESS": node.PublicAddress,
	})
	if err != nil {
		return fmt.Errorf("failed to construct shell script: %v", err)
	}

	_, stderr, _, err := runCommand(conn, command, ctx.Verbose)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

	return err
}
