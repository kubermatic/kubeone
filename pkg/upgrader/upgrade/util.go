package upgrade

import (
	"context"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

func labelNode(client dynclient.Client, host *config.HostConfig) error {
	retErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node := corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: host.Hostname},
		}

		_, err := controllerutil.CreateOrUpdate(context.Background(), client, &node, func(runtime.Object) error {
			if node.ObjectMeta.CreationTimestamp.IsZero() {
				return errors.New("node not found")
			}
			node.Labels[labelUpgradeLock] = ""
			return nil
		})
		return err
	})

	return errors.Wrapf(retErr, "failed to label node %q with label %q", host.Hostname, labelUpgradeLock)
}

func unlabelNode(client dynclient.Client, host *config.HostConfig) error {
	retErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node := corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: host.Hostname},
		}

		_, err := controllerutil.CreateOrUpdate(context.Background(), client, &node, func(runtime.Object) error {
			if node.ObjectMeta.CreationTimestamp.IsZero() {
				return errors.New("node not found")
			}
			delete(node.ObjectMeta.Labels, labelUpgradeLock)
			return nil
		})
		return err
	})

	return errors.Wrapf(retErr, "failed to remove label %s from node %s", labelUpgradeLock, host.Hostname)
}
