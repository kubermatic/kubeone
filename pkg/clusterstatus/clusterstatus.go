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
	"fmt"
	"os"
	"strings"

	"k8c.io/kubeone/pkg/clusterstatus/apiserverstatus"
	"k8c.io/kubeone/pkg/clusterstatus/etcdstatus"
	"k8c.io/kubeone/pkg/clusterstatus/preflightstatus"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tabwriter"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type nodeStatus struct {
	NodeName  string `json:"nodeName,omitempty"`
	Version   string `json:"version,omitempty"`
	APIServer bool   `json:"apiServer,omitempty"`
	Etcd      bool   `json:"etcd,omitempty"`
}

func Print(s *state.State) error {
	status, err := getClusterStatus(s)
	if err != nil {
		return err
	}

	printer := tabwriter.New(os.Stdout)
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

func getClusterStatus(s *state.State) ([]nodeStatus, error) {
	if s.DynamicClient == nil {
		return nil, fail.NoKubeClient()
	}

	// Get node list
	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{preflightstatus.LabelControlPlaneNode: ""}),
	}

	if err := s.DynamicClient.List(s.Context, &nodes, &nodeListOpts); err != nil {
		return nil, fail.KubeClient(err, "listing nodes")
	}

	// Run preflight checks
	if err := preflightstatus.Run(s, nodes); err != nil {
		return nil, err
	}

	status := []nodeStatus{}
	errs := []error{}

	etcdRing, err := etcdstatus.MemberList(s)
	if err != nil {
		return nil, err
	}

	for _, host := range s.Cluster.ControlPlane.Hosts {
		etcdStatus, err := etcdstatus.Get(s, host, etcdRing)
		if err != nil {
			errs = append(errs, err)
		}

		apiserverStatus, err := apiserverstatus.Get(s, host)
		if err != nil {
			errs = append(errs, err)
		}

		var kubeletVersion string
		for _, node := range nodes.Items {
			if node.ObjectMeta.Name == host.Hostname {
				kubeletVersion = node.Status.NodeInfo.KubeletVersion
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

		status = append(status, nodeStatus{
			NodeName:  host.Hostname,
			Version:   kubeletVersion,
			Etcd:      eStatus,
			APIServer: aStatus,
		})
	}

	if len(errs) > 0 {
		for _, e := range errs {
			s.Logger.Errorf("failed to obtain cluster status: %v", e)
		}
	}

	return status, nil
}
