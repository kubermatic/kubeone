/*
Copyright 2026 The KubeOne Authors.

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
	"errors"
	"io"
	"testing"

	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/state"
)

var errRefusedByStub = errors.New("refused by stub executor")

// hostRecordingExecutor records the hosts it was asked to connect to and then
// refuses the connection, which ends the task after it selected a node but
// before it touches one.
type hostRecordingExecutor struct {
	openedHosts []string
}

func (e *hostRecordingExecutor) Open(host kubeoneapi.HostConfig) (executor.Interface, error) {
	e.openedHosts = append(e.openedHosts, host.PrivateAddress)

	return nil, errRefusedByStub
}

func (e *hostRecordingExecutor) Tunnel(_ kubeoneapi.HostConfig) (executor.Tunneler, error) {
	return nil, errRefusedByStub
}

func Test_generateUpgradeStaticWorkersTasks(t *testing.T) {
	tests := []struct {
		name             string
		staticWorkers    []kubeoneapi.HostConfig
		wantDescriptions []string
	}{
		{
			name: "no static workers",
		},
		{
			name: "single static worker",
			staticWorkers: []kubeoneapi.HostConfig{
				{PrivateAddress: "192.168.1.1"},
			},
			wantDescriptions: []string{
				"upgrading 192.168.1.1 static worker node",
			},
		},
		{
			name: "one task per static worker",
			staticWorkers: []kubeoneapi.HostConfig{
				{PrivateAddress: "192.168.1.1"},
				{PrivateAddress: "192.168.1.2"},
				{PrivateAddress: "192.168.1.3"},
			},
			wantDescriptions: []string{
				"upgrading 192.168.1.1 static worker node",
				"upgrading 192.168.1.2 static worker node",
				"upgrading 192.168.1.3 static worker node",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateUpgradeStaticWorkersTasks(tt.staticWorkers)

			if len(got) != len(tt.wantDescriptions) {
				t.Fatalf(
					"generateUpgradeStaticWorkersTasks() = %d tasks, want %d",
					len(got),
					len(tt.wantDescriptions),
				)
			}

			for idx, task := range got {
				if task.Description != tt.wantDescriptions[idx] {
					t.Errorf(
						"task %d Description = %q, want %q",
						idx,
						task.Description,
						tt.wantDescriptions[idx],
					)
				}

				if task.Operation != "upgrading static worker nodes" {
					t.Errorf(
						"task %d Operation = %q, want %q",
						idx,
						task.Operation,
						"upgrading static worker nodes",
					)
				}

				if task.Fn == nil {
					t.Errorf("task %d has no Fn", idx)
				}
			}
		})
	}
}

func Test_generateUpgradeStaticWorkersTasks_runsOnMatchingNode(t *testing.T) {
	staticWorkers := []kubeoneapi.HostConfig{
		{PrivateAddress: "192.168.1.1"},
		{PrivateAddress: "192.168.1.2"},
		{PrivateAddress: "192.168.1.3"},
	}

	got := generateUpgradeStaticWorkersTasks(staticWorkers)

	if len(got) != len(staticWorkers) {
		t.Fatalf(
			"generateUpgradeStaticWorkersTasks() = %d tasks, want %d",
			len(got),
			len(staticWorkers),
		)
	}

	for idx, task := range got {
		stubExecutor := &hostRecordingExecutor{}

		logger := logrus.New()
		logger.SetOutput(io.Discard)

		// The task is expected to fail because the stub executor refuses to
		// connect. What matters is which host it tried to connect to first.
		err := task.Fn(&state.State{
			Logger:   logger,
			Executor: stubExecutor,
		})
		if !errors.Is(err, errRefusedByStub) {
			t.Fatalf("task %d error = %v, want %v", idx, err, errRefusedByStub)
		}

		if len(stubExecutor.openedHosts) != 1 {
			t.Fatalf(
				"task %d connected to %d hosts, want 1",
				idx,
				len(stubExecutor.openedHosts),
			)
		}

		if stubExecutor.openedHosts[0] != staticWorkers[idx].PrivateAddress {
			t.Errorf(
				"task %d connected to host %q, want %q",
				idx,
				stubExecutor.openedHosts[0],
				staticWorkers[idx].PrivateAddress,
			)
		}
	}
}
