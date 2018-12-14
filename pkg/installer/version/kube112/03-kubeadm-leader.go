package kube112

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

const (
	kubeadmCertCommand = `
grep -q KUBECONFIG /etc/environment || { echo 'export KUBECONFIG=/etc/kubernetes/admin.conf' | sudo tee -a /etc/environment; }
if [[ -d ./{{ .WORK_DIR }}/pki ]]; then
       sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
       rm -rf ./{{ .WORK_DIR }}/pki
fi
sudo chown -R root:root /etc/kubernetes
sudo kubeadm alpha phase certs all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo mkdir -p /etc/kubernetes/manifests
cat <<EOF | sudo tee /etc/kubernetes/manifests/etcd.yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    scheduler.alpha.kubernetes.io/critical-pod: ""
  creationTimestamp: null
  labels:
    component: etcd
    tier: control-plane
  name: etcd
  namespace: kube-system
spec:
  containers:
  - command:
    - etcd
    - --advertise-client-urls=http://{{ .PRIVATE_ADDRESS }}:2379
    - --initial-advertise-peer-urls=http://{{ .PRIVATE_ADDRESS }}:2380
    - --initial-cluster={{ .INITIAL_CLUSTER }}
    - --initial-cluster-state=new
    - --listen-client-urls=http://127.0.0.1:2379,http://{{ .PRIVATE_ADDRESS }}:2379
    - --listen-peer-urls=http://{{ .PRIVATE_ADDRESS }}:2380
    - --data-dir=/var/lib/etcd
    - --name={{ .HOSTNAME }}
    - --snapshot-count=10000
    image: k8s.gcr.io/etcd:3.2.24
    imagePullPolicy: IfNotPresent
    livenessProbe:
      exec:
        command:
        - /bin/sh
        - -ec
        - ETCDCTL_API=3 etcdctl --endpoints=http://127.0.0.1:2379 get foo
      failureThreshold: 8
      initialDelaySeconds: 15
      timeoutSeconds: 15
    name: etcd
    resources: {}
    volumeMounts:
    - mountPath: /var/lib/etcd
      name: etcd-data
  hostNetwork: true
  priorityClassName: system-cluster-critical
  volumes:
  - hostPath:
      path: /var/lib/etcd
      type: DirectoryOrCreate
    name: etcd-data
EOF
`
	kubeadmInitCommand = `
if [[ -f /etc/kubernetes/admin.conf ]]; then exit 0; fi
idx=0
while ! curl -so /dev/null --max-time 3 --fail http://127.0.0.1:2379/health; do
    if [ $(( idx++ )) -gt 100 ]; then
        printf "Error: Timeout waiting for etcd endpoint to get healthy.\n"
        exit 1
    fi
    sleep 1s
done
sudo mv /etc/systemd/system/kubelet.service.d/10-kubeadm.conf{.disabled,}
sudo systemctl daemon-reload
sudo systemctl stop kubelet
sudo kubeadm init --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml \
  --ignore-preflight-errors=FileAvailable--etc-kubernetes-manifests-etcd.yaml
`
)

func kubeadmCertsAndEtcdOnLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Configuring certs and etcd on first controller…")
	return util.RunTaskOnLeader(ctx, kubeadmCertsExecutor)
}

func kubeadmCertsAndEtcdOnFollower(ctx *util.Context) error {
	ctx.Logger.Infoln("Configuring certs and etcd on consecutive controller…")
	return util.RunTaskOnFollowers(ctx, kubeadmCertsExecutor, true)
}

func kubeadmCertsExecutor(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	initialCluster := make([]string, len(ctx.Cluster.Hosts))
	for i, host := range ctx.Cluster.Hosts {
		initialCluster[i] = fmt.Sprintf("%s=http://%s:2380", host.Hostname, host.PrivateAddress)
	}
	initialClusterString := strings.Join(initialCluster, ",")

	ctx.Logger.Infoln("Ensuring Certificates…")
	_, _, err := ctx.Runner.Run(kubeadmCertCommand, util.TemplateVariables{
		"PRIVATE_ADDRESS": node.PrivateAddress,
		"HOSTNAME":        node.Hostname,
		"INITIAL_CLUSTER": initialClusterString,
		"WORK_DIR":        ctx.WorkDir,
		"NODE_ID":         strconv.Itoa(node.ID),
	})
	return err
}

func initKubernetesLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Initializing Kubernetes on leader…")

	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Running kubeadm…")

		_, _, err := ctx.Runner.Run(kubeadmInitCommand, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
			"NODE_ID":  strconv.Itoa(node.ID),
		})

		return err
	})
}
