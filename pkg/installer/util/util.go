package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/template"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// Tee mimics the unix `tee` command by piping its
// input through to the upstream writer and also
// capturing it in a buffer.
type Tee struct {
	buffer   bytes.Buffer
	upstream io.Writer
}

func (t *Tee) Write(p []byte) (int, error) {
	t.buffer.Write(p)

	return t.upstream.Write(p)
}

func (t *Tee) String() string {
	return strings.TrimSpace(t.buffer.String())
}

// RunCommand on remove machine over SSH
func RunCommand(conn ssh.Connection, cmd string, verbose bool) (string, string, int, error) {
	if !verbose {
		return conn.Exec(cmd)
	}

	stdout := &Tee{
		upstream: os.Stdout,
	}

	stderr := &Tee{
		upstream: os.Stdout,
	}

	// ensure we fail early
	cmd = fmt.Sprintf("set -xeu pipefail\n\n%s", cmd)

	// ensure sudo works on exotic distros
	cmd = fmt.Sprintf("export \"PATH=$PATH:/sbin:/usr/local/bin:/opt/bin\"\n\n%s", cmd)

	exitCode, err := conn.Stream(cmd, stdout, os.Stderr)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

	return stdout.String(), stderr.String(), exitCode, err
}

// TemplateVariables is a render context for templates
type TemplateVariables map[string]interface{}

// MakeShellCommand render text template with given `variables` render-context
func MakeShellCommand(cmd string, variables TemplateVariables) (string, error) {
	tpl, err := template.New("base").Parse(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to parse shell script: %v", err)
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render shell script: %v", err)
	}

	return buf.String(), nil
}

// RunShellCommand combines MakeShellCommand and RunCommand.
func RunShellCommand(conn ssh.Connection, verbose bool, cmd string, variables TemplateVariables) (string, string, int, error) {
	command, err := MakeShellCommand(cmd, variables)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to construct shell script: %v", err)
	}

	stdout, stderr, exitCode, err := RunCommand(conn, command, verbose)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

	return stdout, stderr, exitCode, err
}

// WaitForPod waits for the availability of the given Kubernetes element.
func WaitForPod(conn ssh.Connection, verbose bool, namespace string, name string, timeout time.Duration) error {
	cmd := fmt.Sprintf(`sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf -n "%s" get pod "%s" -o jsonpath='{.status.phase}' --ignore-not-found`, namespace, name)
	if !WaitForCondition(conn, verbose, cmd, timeout, IsRunning) {
		return fmt.Errorf("timed out while waiting for %s/%s to come up for %v", namespace, name, timeout)
	}

	return nil
}

type validatorFunc func(stdout string) bool

// IsRunning checks if the given output represents the "Running" status of a Kubernetes pod.
func IsRunning(stdout string) bool {
	return strings.ToLower(stdout) == "running"
}

// WaitForCondition waits for something to be true.
func WaitForCondition(conn ssh.Connection, verbose bool, cmd string, timeout time.Duration, validator validatorFunc) bool {
	cutoff := time.Now().Add(timeout)

	for time.Now().Before(cutoff) {
		stdout, _, _, _ := RunCommand(conn, cmd, verbose)
		if validator(stdout) {
			return true
		}

		time.Sleep(1 * time.Second)
	}

	return false
}
