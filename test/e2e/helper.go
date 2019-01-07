package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// CreateProvisioner returns interface for specific provisioner
func CreateProvisioner(region, testPath string, identifier string, provider string) (Provisioner, error) {
	if provider == AWS {
		return NewAWSProvisioner(region, testPath, identifier)
	}

	return nil, fmt.Errorf("unsuported provider %v", provider)
}

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

// executeCommand executes given command
func executeCommand(path, name string, arg []string) (string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	var errStdout, errStderr error

	cmd := exec.Command(name, arg...)
	if len(path) > 0 {
		cmd.Dir = path
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

// RandomString generates random string
func RandomString(length int) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
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
