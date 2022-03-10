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

import "errors"

const (
	defaultExitCode         = 1
	runtimeErrorExitCode    = 10
	etcdErrorExitCode       = 11
	kubeClientErrorExitCode = 12
	sshErrorExitCode        = 13
	connectionErrorExitCode = 14
	configErrorExitCode     = 15
)

type exitCoder interface {
	exitCode() int
}

var (
	_ exitCoder = RuntimeError{}
	_ exitCoder = EtcdError{}
	_ exitCoder = KubeClientError{}
	_ exitCoder = SSHError{}
	_ exitCoder = ConnectionError{}
	_ exitCoder = ConfigError{}
	_ exitCoder = CredentialsError{}
)

func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var exiter exitCoder
	if errors.As(err, &exiter) {
		return exiter.exitCode()
	}

	return defaultExitCode
}
