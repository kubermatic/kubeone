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

package fail

import (
	"fmt"
	"strings"
)

// RuntimeError wraps kubernetes client related errors
type RuntimeError struct {
	Err error
	Op  string
}

func (e RuntimeError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("runtime: %s\n%s", e.Op, e.Err)
	}

	return e.Err.Error()
}

func (e RuntimeError) Unwrap() error { return e.Err }
func (e RuntimeError) exitCode() int { return RuntimeErrorExitCode }

// EtcdError wraps etcd client related errors
type EtcdError struct {
	Err error
	Op  string
}

func (e EtcdError) Error() string { return fmt.Sprintf("etcd: %s\n%s", e.Op, e.Err) }
func (e EtcdError) Unwrap() error { return e.Err }
func (e EtcdError) exitCode() int { return EtcdErrorExitCode }

// KubeClientError wraps kubernetes client related errors
type KubeClientError struct {
	Err error
	Op  string
}

func (e KubeClientError) Error() string { return fmt.Sprintf("kubernetes: %s\n%s", e.Op, e.Err) }
func (e KubeClientError) Unwrap() error { return e.Err }
func (e KubeClientError) exitCode() int { return KubeClientErrorExitCode }

// SSHError wraps SSH related errors
type SSHError struct {
	Err    error
	Op     string
	Cmd    string
	Stderr string
}

func (e SSHError) Error() string {
	var (
		format = "ssh: %s\n%s"
		args   = []interface{}{e.Op, e.Err}
	)
	if e.Cmd != "" {
		format += "\n%s"
		args = append(args, e.Cmd)
	}

	if e.Stderr != "" {
		format += "\nstderr: %s"
		args = append(args, e.Stderr)
	}

	return fmt.Sprintf(format, args...)
}

func (e SSHError) Unwrap() error { return e.Err }
func (e SSHError) exitCode() int { return SSHErrorExitCode }

// ExecError wraps SSH related errors
type ExecError struct {
	Err    error
	Op     string
	Cmd    string
	Stderr string
}

func (e ExecError) Error() string {
	var (
		format = "exec: %s\n%s"
		args   = []interface{}{e.Op, e.Err}
	)
	if e.Cmd != "" {
		format += "\n%s"
		args = append(args, e.Cmd)
	}

	if e.Stderr != "" {
		format += "\nstderr: %s"
		args = append(args, e.Stderr)
	}

	return fmt.Sprintf(format, args...)
}

func (e ExecError) Unwrap() error { return e.Err }
func (e ExecError) exitCode() int { return ExecErrorExitCode }

// ConnectionError wraps connections related errors
type ConnectionError struct {
	Err    error
	Target string
}

func (e ConnectionError) Error() string {
	return fmt.Sprintf("connection to: %s\n%s", e.Target, e.Err)
}

func (e ConnectionError) Unwrap() error { return e.Err }
func (e ConnectionError) exitCode() int { return ConnectionErrorExitCode }

// ConfigError wraps configuration related errors
type ConfigError struct {
	Err error
	Op  string
}

func (e ConfigError) Error() string { return fmt.Sprintf("configuration %s\n%s", e.Op, e.Err) }
func (e ConfigError) Unwrap() error { return e.Err }
func (e ConfigError) exitCode() int { return ConfigErrorExitCode }

// CredentialsError wraps cloud provider credentials related errors
type CredentialsError struct {
	Err      error
	Op       string
	Provider string
}

func (e CredentialsError) Error() string {
	var res strings.Builder
	fmt.Fprintf(&res, "credentials:\n")

	if e.Op != "" {
		fmt.Fprintf(&res, "%s:\n", e.Op)
	}

	if e.Provider != "" {
		fmt.Fprintf(&res, "provider: %s\n", e.Provider)
	}

	res.WriteString(e.Err.Error())

	return res.String()
}

func (e CredentialsError) Unwrap() error { return e.Err }
func (e CredentialsError) exitCode() int { return ConfigErrorExitCode }
