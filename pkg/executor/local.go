/*
Copyright 2022 The KubeOne Authors.

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

package executor

import (
	"context"
	"io"
	"net"
	"os/exec"
	"strings"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
)

func NewLocal(ctx context.Context) Adapter {
	return &localAdapter{
		ctx: ctx,
	}
}

type localAdapter struct {
	ctx context.Context
}

func (la *localAdapter) Open(host kubeoneapi.HostConfig) (Interface, error) {
	ctx, cancel := context.WithCancel(la.ctx)

	return &localExec{ctx: ctx, cancelFn: cancel}, nil
}

func (la *localAdapter) Tunnel(host kubeoneapi.HostConfig) (Tunneler, error) {
	ctx, cancel := context.WithCancel(la.ctx)

	return &localExec{ctx: ctx, cancelFn: cancel}, nil
}

type localExec struct {
	ctx      context.Context
	cancelFn func()
}

func (le *localExec) Exec(cmd string) (string, string, int, error) {
	var (
		stdout, stderr strings.Builder
		returnErr      error
	)

	exitcode, err := le.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		returnErr = fail.ExecError{
			Err:    err,
			Op:     "exec",
			Cmd:    cmd,
			Stderr: stderr.String(),
		}
	}

	return stdout.String(), stderr.String(), exitcode, returnErr
}

func (le *localExec) POpen(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	command := exec.CommandContext(le.ctx, "bash", "-c", cmd)
	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = stderr
	err := command.Run()

	return command.ProcessState.ExitCode(), err
}

func (le *localExec) TunnelTo(ctx context.Context, network, addr string) (net.Conn, error) {
	var d net.Dialer

	return d.DialContext(ctx, network, addr)
}

func (le *localExec) Close() error {
	le.cancelFn()

	return nil
}
