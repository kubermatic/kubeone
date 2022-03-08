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
)

// RuntimeError wraps kubernetes client related errors
type RuntimeError struct {
	Err error
	Op  string
}

func (e RuntimeError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %s", e.Op, e.Err)
	}

	return e.Err.Error()
}

func (e RuntimeError) Unwrap() error { return e.Err }

// EtcdError wraps etcd client related errors
type EtcdError struct {
	Err error
	Op  string
}

func (e EtcdError) Error() string { return fmt.Sprintf("%s: %s", e.Op, e.Err) }
func (e EtcdError) Unwrap() error { return e.Err }

// KubeClientError wraps kubernetes client related errors
type KubeClientError struct {
	Err error
	Op  string
}

func (e KubeClientError) Error() string { return fmt.Sprintf("%s: %s", e.Op, e.Err) }
func (e KubeClientError) Unwrap() error { return e.Err }

// SSHError wraps SSH related errors
type SSHError struct {
	Err error
	Op  string
}

func (e SSHError) Error() string { return fmt.Sprintf("%s: %s", e.Op, e.Err) }
func (e SSHError) Unwrap() error { return e.Err }

// ConnectionError wraps connections related errors
type ConnectionError struct {
	Err    error
	Target string
}

func (e ConnectionError) Error() string { return fmt.Sprintf("connecting to %s: %s", e.Target, e.Err) }
func (e ConnectionError) Unwrap() error { return e.Err }

// ConfigError wraps configuration related errors
type ConfigError struct {
	Err error
	Op  string
}

func (e ConfigError) Error() string { return fmt.Sprintf("%s: %s", e.Op, e.Err) }
func (e ConfigError) Unwrap() error { return e.Err }
