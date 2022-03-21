/*
Copyright 2021 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodeutils

import (
	"context"

	"github.com/sirupsen/logrus"

	"k8c.io/kubeone/pkg/fail"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/drain"
)

type Drainer interface {
	Drain(ctx context.Context, nodeName string) error
	Cordon(ctx context.Context, nodeName string, state bool) error
}

func NewDrainer(restconfig *rest.Config, logger logrus.FieldLogger) Drainer {
	return &drainer{
		logger:     logger,
		restconfig: restconfig,
	}
}

type drainer struct {
	logger     logrus.FieldLogger
	restconfig *rest.Config
}

func (dr *drainer) Drain(ctx context.Context, nodeName string) error {
	drainerHelper, err := dr.drainHelper(ctx)
	if err != nil {
		return err
	}

	return fail.KubeClient(drain.RunNodeDrain(drainerHelper, nodeName), "draining %q node", nodeName)
}

func (dr *drainer) Cordon(ctx context.Context, nodeName string, desired bool) error {
	drainerHelper, err := dr.drainHelper(ctx)
	if err != nil {
		return err
	}

	node, err := drainerHelper.Client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fail.KubeClient(err, "getting Node %s", nodeName)
	}

	op := "cordon Node %s"
	if !desired {
		op = "uncordon Node %s"
	}

	return fail.KubeClient(drain.RunCordonOrUncordon(drainerHelper, node, desired), op, nodeName)
}

func (dr *drainer) drainHelper(ctx context.Context) (*drain.Helper, error) {
	kubeClinet, err := kubernetes.NewForConfig(dr.restconfig)
	if err != nil {
		return nil, fail.KubeClient(err, "initializing new kubernetes clientset")
	}

	return &drain.Helper{
		Ctx:    ctx,
		Client: kubeClinet,
		// Force is used to force deleting standalone pods (i.e. not managed by
		// ReplicaSet)
		Force:               true,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
		Out:                 loggerIoWriter(dr.logger.Infof),
		ErrOut:              loggerIoWriter(dr.logger.Errorf),
		OnPodDeletedOrEvicted: func(pod *corev1.Pod, usingEviction bool) {
			evicted := "evicted"
			if !usingEviction {
				evicted = "deleted"
			}
			dr.logger.Infof("pod %q/%q is %s", pod.GetNamespace(), pod.GetName(), evicted)
		},
	}, nil
}

type loggerIoWriter func(format string, args ...interface{})

func (lw loggerIoWriter) Write(p []byte) (n int, err error) {
	lw("%s", p)

	return len(p), nil
}
