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

	"github.com/pkg/errors"
)

// ConfigValidation is a shortcut to quickly construct ConfigError
func ConfigValidation(err error) error {
	return Config(err, "validation")
}

func NewConfigError(op string, format string, args ...interface{}) error {
	return ConfigError{
		Op:  op,
		Err: errors.Errorf(format, args...),
	}
}

func Config(err error, op string) error {
	if err == nil {
		return nil
	}

	return ConfigError{
		Err: errors.WithStack(err),
		Op:  op,
	}
}

// SSH is a shortcut to quickly construct SSHError
func SSH(err error, op string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return SSHError{
		Op:  fmt.Sprintf(op, args...),
		Err: errors.WithStack(err),
	}
}

// Connection is a shortcut to quickly construct ConnectionError
func Connection(err error, target string) error {
	if err == nil {
		return nil
	}

	return ConnectionError{
		Target: target,
		Err:    errors.WithStack(err),
	}
}

// KubeClient is a shortcut to quickly construct KubeClientError
func KubeClient(err error, op string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return KubeClientError{
		Op:  fmt.Sprintf(op, args...),
		Err: errors.WithStack(err),
	}
}

// NoKubeClient is a shortcut to quickly construct KubeClientError with predefined not initialized error
func NoKubeClient() error {
	return KubeClientError{
		Op:  "kubernetes dynamic client",
		Err: errors.New("not initialized"),
	}
}

// Etcd is a shortcut to quickly construct EtcdError
func Etcd(err error, op string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return EtcdError{
		Op:  fmt.Sprintf(op, args...),
		Err: errors.WithStack(err),
	}
}

// Runtime is a shortcut to quickly construct RuntimeError
func Runtime(err error, op string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return RuntimeError{
		Op:  fmt.Sprintf(op, args...),
		Err: errors.WithStack(err),
	}
}

func NewRuntimeError(op string, format string, args ...interface{}) error {
	return RuntimeError{
		Op:  op,
		Err: errors.Errorf(format, args...),
	}
}
