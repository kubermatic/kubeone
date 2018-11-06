package tasks

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

type InitKubernetesTask struct{}

func (t *InitKubernetesTask) Execute(ctx *Context) error {
	var err error

	ctx.Logger.Infoln("Initializing Kubernetes…")

	for _, node := range ctx.Manifest.Hosts {
		logger := ctx.Logger.WithFields(logrus.Fields{
			"node": node.PublicAddress,
		})

		err = t.init(ctx, node, logger)
		if err != nil {
			break
		}

		err = t.waitForApiserver(ctx, node, logger)
		if err != nil {
			break
		}
	}

	return err
}

func (t *InitKubernetesTask) init(ctx *Context, node manifest.HostManifest, logger logrus.FieldLogger) error {
	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	logger.Infoln("Running kubeadm…")
	command, err := makeShellCommand(`
set -xeu
export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"
sudo kubeadm init \
     --config=./{{ .WORK_DIR }}/cfg/master.yaml \
     --ignore-preflight-errors=Port-10250,FileAvailable--etc-kubernetes-manifests-etcd.yaml,FileExisting-crictl
`, templateVariables{
		"WORK_DIR": ctx.WorkDir,
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

func (t *InitKubernetesTask) waitForApiserver(ctx *Context, node manifest.HostManifest, logger logrus.FieldLogger) error {
	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	command := fmt.Sprintf(
		`curl --max-time 3 --fail --cacert /etc/kubernetes/pki/ca.crt https://%s:6443/healthz`,
		node.PublicAddress)

	logger.Infoln("Waiting for apiserver…")
	for remaining := 20; remaining >= 0; remaining-- {
		_, _, _, err = runCommand(conn, command, ctx.Verbose)
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("Kubernetes apiserver did not come up, giving up")
	}

	return nil
}
