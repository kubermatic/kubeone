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

package sshtunnel

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/ssh"
)

// NewGRPCDialer initialize gRPC dialer that will use ssh tunnel as
// transport
func NewGRPCDialer(connector *ssh.Connector, target kubeoneapi.HostConfig) (grpc.DialOption, error) {
	tunnel, err := connector.Tunnel(target)
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish SSH tunnel")
	}

	return grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return tunnel.TunnelTo(ctx, "tcp4", addr)
	}), nil
}
