package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// CreateProvisioner returns interface for specific provisioner
func CreateProvisioner(region, testName, testPath string, provider Provider) Provisioner {
	if provider == AWS {
		pr := NewAWSProvisioner(region, testName, testPath)
		return reflect.ValueOf(pr).Interface().(Provisioner)
	}

	return nil
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
func executeCommand(path, name string, arg []string) (string, string, int) {
	var stdoutBuf, stderrBuf bytes.Buffer
	var errStdout, errStderr error
	verbose := testing.Verbose()

	cmd := exec.Command(name, arg...)
	if len(path) > 0 {
		cmd.Dir = path
	}

	if !verbose {
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {

			return "", err.Error(), 1
		}
		outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
		return outStr, errStr, 0
	}

	doneStdout := make(chan struct{})
	doneStderr := make(chan struct{})

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Start()
	if err != nil {
		return "", err.Error(), 1
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
		return "", err.Error(), 1
	}
	if errStdout != nil {
		return "", errStdout.Error(), 1
	}
	if errStderr != nil {
		return "", errStderr.Error(), 1
	}

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	return outStr, errStr, 0
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

// CreateDir creates directory for given path
func CreateDir(dirName string) error {
	if os.MkdirAll(dirName, 0755) != nil {
		return fmt.Errorf("unable to create directory %s", dirName)
	}
	return nil
}

// CreateFile create file with given content
func CreateFile(filepath, content string) error {

	// Create directory if needed.
	basepath := path.Dir(filepath)
	filename := path.Base(filepath)

	if err := CreateDir(basepath); err != nil {
		return err
	}

	// Create the file.
	fileOut, err := os.Create(strings.Join([]string{basepath, filename}, "/"))

	if err != nil {
		return fmt.Errorf("unable to create tag file! %v", err)
	}

	defer fileOut.Close()

	_, err = io.WriteString(fileOut, content)

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
