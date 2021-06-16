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

package tasks

import (
	"net/url"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"

	"k8c.io/kubeone/pkg/clusterstatus/preflightstatus"
	"k8c.io/kubeone/pkg/etcdutil"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func repairClusterIfNeeded(s *state.State) error {
	s.Logger.Info("Check if cluster needs any repairs...")

	leader, err := s.Cluster.Leader()
	if err != nil {
		return errors.WithStack(err)
	}

	etcdcfg, err := etcdutil.NewClientConfig(s, leader)
	if err != nil {
		return errors.WithStack(err)
	}

	etcdcli, err := clientv3.New(*etcdcfg)
	if err != nil {
		return errors.WithStack(err)
	}
	defer etcdcli.Close()

	ctx := s.Context

	etcdRing, err := etcdcli.MemberList(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	knownHostsIdentities := sets.NewString()
	knownEtcdMembersIdentities := sets.NewString()

	for _, host := range s.Cluster.ControlPlane.Hosts {
		knownHostsIdentities.Insert(host.Hostname, host.PublicAddress, host.PrivateAddress)
	}

	membersToDelete := make(map[string]uint64)

	for _, peer := range etcdRing.Members {
		knownEtcdMembersIdentities.Insert(peer.Name)
		peerIdentities := []string{peer.Name}

		for _, endpoint := range peer.ClientURLs {
			endpointURL, uerr := url.Parse(endpoint)
			if uerr != nil {
				s.Logger.Errorf("failed to parse etcd clientURL: %v", uerr)
				continue
			}

			peerIdentities = append(peerIdentities, endpointURL.Hostname())
			endpointStatus, serr := etcdcli.Status(ctx, endpointURL.Host)
			if serr != nil {
				s.Logger.Errorf("failed etcd member %v endpoint status, error: %v", peer, serr)
				s.Logger.Warnf("scheduling etcd member %v to delete", peer)
				membersToDelete[peer.Name] = peer.ID
			}

			if endpointStatus != nil && len(endpointStatus.Errors) > 0 {
				s.Logger.Errorf("etcd peer experience errors: %v", endpointStatus.Errors)
			}
		}

		if !knownHostsIdentities.HasAny(peerIdentities...) {
			s.Logger.Warnf("scheduling etcd member %v to delete", peer)
			membersToDelete[peer.Name] = peer.ID
		}
	}

	for memberName, memberID := range membersToDelete {
		knownEtcdMembersIdentities.Delete(memberName)
		s.Logger.Warnf("removing etcd member %q, for it's not alive", memberName)
		if _, err = etcdcli.MemberRemove(ctx, memberID); err != nil {
			return errors.WithStack(err)
		}
	}

	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{preflightstatus.LabelControlPlaneNode: ""}),
	}

	if err = s.DynamicClient.List(ctx, &nodes, &nodeListOpts); err != nil {
		return errors.WithStack(err)
	}

	for _, node := range nodes.Items {
		var deleteThisNode bool

		if _, ok := membersToDelete[node.Name]; ok {
			deleteThisNode = true
		}

		if !knownEtcdMembersIdentities.Has(node.Name) {
			deleteThisNode = true
		}

		if deleteThisNode {
			s.Logger.Warnf("Removing kubernets Node object %q, for it's not alive", node.Name)
			if err = s.DynamicClient.Delete(ctx, node.DeepCopyObject().(dynclient.Object)); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
