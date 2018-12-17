package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/etcd"

	"github.com/ghodss/yaml"
)

func setupEtcd(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	pod := etcd.Pod(ctx.Cluster, node)
	manifest, err := yaml.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to marhsal etcd manifest: %v", err)
	}
	_, _, err = ctx.Runner.Run(`
sudo mkdir -p /etc/kubernetes/manifests
cat <<EOF | sudo tee /etc/kubernetes/manifests/etcd.yaml
{{ .ETCD_POD}}
EOF
idx=0
while ! curl -so /dev/null --max-time 3 --fail http://127.0.0.1:2379/health; do
    if [[ $(( idx++ )) -gt 100 ]]; then
        printf "Error: Timeout waiting for etcd endpoint to get healthy.\n"
        exit 1
    fi
    sleep 1s
done
	`, util.TemplateVariables{"ETCD_POD": string(manifest)})
	return err
}
