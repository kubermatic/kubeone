package installer

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/installer/tasks"
	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

type installer struct {
	manifest *manifest.Manifest
	logger   *logrus.Logger
}

func NewInstaller(manifest *manifest.Manifest, logger *logrus.Logger) *installer {
	return &installer{
		manifest: manifest,
		logger:   logger,
	}
}

func (i *installer) Run() (*Result, error) {
	ctx := tasks.Context{
		Manifest:      i.manifest,
		Connector:     ssh.NewConnector(),
		Configuration: tasks.NewConfiguration(),
		WorkDir:       "kubermatic-installer",
	}

	// install prerequisites
	ctx.Logger = i.logger.WithField("task", "prerequisites")
	err := (&tasks.InstallPrerequisitesTask{}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("installation of prerequisites failed: %v", err)
	}

	// generate CA
	ctx.Logger = i.logger.WithField("task", "generate-ca")
	err = (&tasks.GenerateCATask{}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("CA generation failed: %v", err)
	}

	// deploy CA
	ctx.Logger = i.logger.WithField("task", "deploy-ca")
	err = (&tasks.DeployCATask{}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("CA deployment failed: %v", err)
	}

	// wait for etcd
	ctx.Logger = i.logger.WithField("task", "wait-for-etcd")
	err = (&tasks.WaitForEtcdTask{}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("etcd cluster initialization failed: %v", err)
	}

	// init Kubernetes
	ctx.Logger = i.logger.WithField("task", "init-kubernetes")
	err = (&tasks.InitKubernetesTask{}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("Kubernetes initialization failed: %v", err)
	}

	// let the cluster settle down
	ctx.Logger = i.logger.WithField("task", "coffee-break")
	err = (&tasks.ClusterSettleTask{Duration: 30 * time.Second}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to let the cluster settle down: %v", err)
	}

	// init kube-proxy
	ctx.Logger = i.logger.WithField("task", "init-kube-proxy")
	err = (&tasks.InstallKubeProxyTask{}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("kube-proxy initialization failed: %v", err)
	}

	// let the cluster settle down
	ctx.Logger = i.logger.WithField("task", "coffee-break")
	err = (&tasks.ClusterSettleTask{Duration: 10 * time.Second}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to let the cluster settle down: %v", err)
	}

	// create join token and command
	ctx.Logger = i.logger.WithField("task", "create-join-token")
	err = (&tasks.CreateJoinTokenTask{}).Execute(&ctx)
	if err != nil {
		return nil, fmt.Errorf("join token initialization failed: %v", err)
	}

	return nil, nil
}
