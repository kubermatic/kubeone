package kube111

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func installKubeProxy(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, node config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Installing kube-proxy…")

		_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `
mkdir -p ~/.kube
sudo cp /etc/kubernetes/admin.conf ~/.kube/config
sudo chown -R $(id -u):$(id -g) ~/.kube

kubectl apply -f ./{{ .WORK_DIR }}/kube-flannel.yaml

kubectl -n kube-system get configmap kube-proxy -o yaml > kube-proxy-configmap.yaml
sed -i -e 's#server:.*#server: https://{{ .IP_ADDRESS }}:6443#g' kube-proxy-configmap.yaml
kubectl delete -f kube-proxy-configmap.yaml
kubectl create -f kube-proxy-configmap.yaml
kubectl -n kube-system delete pod -l k8s-app=kube-proxy
`, util.TemplateVariables{
			"WORK_DIR":   ctx.WorkDir,
			"IP_ADDRESS": node.PublicAddress,
		})

		return err
	})
}
