/*
Copyright 2019 The KubeOne Authors.

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

package clusterstatus

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/clusterstatus/apiserver"
	"github.com/kubermatic/kubeone/pkg/clusterstatus/etcd"
	"github.com/kubermatic/kubeone/pkg/clusterstatus/preflight"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/tabwriter"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Status struct {
	NodeName  string `json:"nodeName,omitempty"`
	Version   string `json:"version,omitempty"`
	APIServer bool   `json:"apiServer,omitempty"`
	Etcd      bool   `json:"etcd,omitempty"`
}

func PrintClusterStatus(s *state.State) error {
	status, err := GetClusterStatus(s)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster status")
	}

	printer := tabwriter.GetNewTabWriter(os.Stdout)
	defer printer.Flush()

	headers := clusterStatusHeader()
	for _, h := range headers {
		fmt.Fprintf(printer, "%s\t", strings.ToUpper(h))
	}
	fmt.Fprintln(printer, "")
	for _, s := range status {
		fmt.Fprintf(printer, "%s\t", s.NodeName)
		fmt.Fprintf(printer, "%s\t", s.Version)
		if s.APIServer {
			fmt.Fprintf(printer, "healthy\t")
		} else {
			fmt.Fprintf(printer, "unhealthy\t")
		}
		if s.Etcd {
			fmt.Fprintf(printer, "healthy\t")
		} else {
			fmt.Fprintf(printer, "unhealthy\t")
		}
		fmt.Fprintln(printer, "")
	}

	return nil
}

func clusterStatusHeader() []string {
	return []string{
		"Node",
		"Version",
		"APIServer",
		"Etcd",
	}
}

func GetClusterStatus(s *state.State) ([]Status, error) {
	if s.DynamicClient == nil {
		return nil, errors.New("kubernetes client not initialized")
	}

	// Get node list
	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{preflight.LabelControlPlaneNode: ""}),
	}
	err := s.DynamicClient.List(context.Background(), &nodes, &nodeListOpts)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list nodes")
	}

	// Run preflight checks
	if preflightErr := preflight.RunPreflightChecks(s, nodes); err != nil {
		return nil, preflightErr
	}

	tunn, err := s.Connector.Tunnel(s.Cluster.RandomHost())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get SSH tunnel")
	}

	status := []Status{}
	errs := []error{}
	for _, host := range s.Cluster.Hosts {
		etcdStatus, err := etcd.GetStatus(s, host, tunn)
		if err != nil {
			errs = append(errs, err)
		}
		apiserverStatus, err := apiserver.GetStatus(host, tunn)
		if err != nil {
			errs = append(errs, err)
		}
		var version string
		for _, node := range nodes.Items {
			if node.ObjectMeta.Name == host.Hostname {
				version = node.Status.NodeInfo.KubeletVersion
			}
		}

		eStatus := false
		if etcdStatus != nil && etcdStatus.Health && etcdStatus.Member {
			eStatus = true
		}
		aStatus := false
		if apiserverStatus != nil && apiserverStatus.Health {
			aStatus = true
		}

		status = append(status, Status{
			NodeName:  host.Hostname,
			Version:   version,
			Etcd:      eStatus,
			APIServer: aStatus,
		})
	}
	if len(errs) > 0 {
		return nil, utilerrors.NewAggregate(errs)
	}

	return status, nil
}
