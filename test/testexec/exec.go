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

package testexec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"syscall"
)

type Opt func(*Exec) *Exec

func NewExec(command string, opts ...Opt) *Exec {
	e := &Exec{
		Command: command,
		Stderr:  os.Stderr,
		Stdout:  os.Stdout,
	}

	StderrDebug(e)

	for _, o := range opts {
		o(e)
	}

	return e
}

type Exec struct {
	Logf    func(string, ...interface{})
	Args    []string
	Command string
	Cwd     string
	Env     []string
	Stderr  io.Writer
	Stdout  io.Writer

	dryRun bool
}

func (e *Exec) Run() error {
	cmd := e.BuildCmd(context.Background())

	if e.dryRun {
		return nil
	}

	return cmd.Run()
}

func (e *Exec) BuildCmd(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, e.Command, e.Args...) //nolint:gosec
	// getach child processes from OS signals from outside
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	cmd.Dir = e.Cwd

	if len(e.Env) != 0 {
		cmd.Env = make([]string, len(e.Env))
		copy(cmd.Env, e.Env)
	}

	if e.Stdout != nil {
		cmd.Stdout = e.Stdout
	}

	if e.Stderr != nil {
		cmd.Stderr = e.Stderr
	}

	if e.Logf != nil {
		e.Logf("in dir: %s, %v", e.Cwd, cmd.Args)
	}

	return cmd
}

func WithArgs(args ...string) Opt {
	return func(e *Exec) *Exec {
		e.Args = args

		return e
	}
}

func StdoutTo(stdout io.Writer) Opt {
	return func(e *Exec) *Exec {
		e.Stdout = stdout

		return e
	}
}

func StderrTo(stderr io.Writer) Opt {
	return func(e *Exec) *Exec {
		e.Stderr = stderr

		return e
	}
}

func InDir(dir string) Opt {
	return func(e *Exec) *Exec {
		e.Cwd = dir

		return e
	}
}

func WithMapEnv(env map[string]string) Opt {
	return func(e *Exec) *Exec {
		var maptosliceEnv []string

		for k, v := range env {
			maptosliceEnv = append(maptosliceEnv, fmt.Sprintf("%s=%s", k, v))
		}

		sort.Strings(maptosliceEnv)
		e.Env = append(e.Env, maptosliceEnv...)

		return e
	}
}

func WithEnv(env []string) Opt {
	return func(e *Exec) *Exec {
		e.Env = append(e.Env, env...)

		return e
	}
}

func WithEnvs(envs ...string) Opt {
	return func(e *Exec) *Exec {
		e.Env = append(e.Env, envs...)

		return e
	}
}

func WithDryRun() Opt {
	return func(e *Exec) *Exec {
		e.dryRun = true

		return e
	}
}

func LogFunc(logf func(string, ...interface{})) Opt {
	return func(e *Exec) *Exec {
		e.Logf = logf

		return e
	}
}

func DebugTo(w io.Writer) Opt {
	return LogFunc(func(format string, a ...interface{}) {
		fmt.Fprintf(w, "\n +"+format+"\n", a...)
	})
}

var (
	StdoutDebug = DebugTo(os.Stdout)
	StderrDebug = DebugTo(os.Stderr)
)
