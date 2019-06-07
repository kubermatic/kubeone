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

package testutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

// IsCommandAvailable checks if command is available OS
func IsCommandAvailable(name string) bool {
	path, err := exec.LookPath(name)
	if err != nil {
		return false
	}
	if len(path) > 0 {
		return true
	}

	return false
}

// ExecuteCommand executes the given command
func ExecuteCommand(path, name string, arg []string, additionalEnv map[string]string) (string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	var errStdout, errStderr error

	cmd := exec.Command(name, arg...)
	if len(path) > 0 {
		cmd.Dir = path
	}

	if additionalEnv != nil {
		cmd.Env = os.Environ()
		for k, v := range additionalEnv {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	doneStdout := make(chan struct{})
	doneStderr := make(chan struct{})

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Start()
	if err != nil {
		return "", err
	}

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		doneStdout <- struct{}{}

	}()

	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
		doneStderr <- struct{}{}
	}()

	<-doneStdout
	<-doneStderr
	err = cmd.Wait()
	if err != nil {
		return "", err
	}
	if errStdout != nil {
		return "", errStdout
	}
	if errStderr != nil {
		return "", errStderr
	}

	outStr := string(stdoutBuf.Bytes())
	return outStr, nil
}

// CreateFile create file with given content
func CreateFile(filepath, content string) error {
	// Create directory if needed.
	basepath := path.Dir(filepath)
	filename := path.Base(filepath)

	err := os.MkdirAll(basepath, 0755)
	if err != nil {
		return fmt.Errorf("unable to create directory %s", basepath)
	}

	// Create the file.
	err = ioutil.WriteFile(strings.Join([]string{basepath, filename}, "/"), []byte(content), os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write data to file")
	}
	return nil
}

// ValidateCommon validates variables necessary to start process
func ValidateCommon() error {
	sshPublicKey := os.Getenv("SSH_PUBLIC_KEY_FILE")
	if len(sshPublicKey) == 0 {
		return errors.New("unable to run the test suite, SSH_PUBLIC_KEY_FILE environment variables cannot be empty")
	}

	if ok := IsCommandAvailable("terraform"); !ok {
		return errors.New("the terraform client is not available, please install")
	}

	if ok := IsCommandAvailable("kubetest"); !ok {
		return errors.New("the kubetest is not available, please install: 'go get -u k8s.io/test-infra/kubetest'")
	}

	return nil
}
