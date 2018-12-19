package util

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/koron-go/prefixw"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// Runner bundles a connection to a host with the verbosity and
// other options for running commands via SSH.
type Runner struct {
	Conn    ssh.Connection
	Prefix  string
	OS      string
	Verbose bool
}

// Run executes a given command/script, optionally printing its output to
// stdout/stderr.
func (r *Runner) Run(cmd string, variables TemplateVariables) (string, string, error) {
	if r.Conn == nil {
		return "", "", errors.New("runner is not tied to an opened SSH connection")
	}

	cmd, err := MakeShellCommand(cmd, variables)
	if err != nil {
		return "", "", err
	}

	cmd = r.prepareShell(cmd)

	if !r.Verbose {
		var stdout, stderr string

		stdout, stderr, _, err = r.Conn.Exec(cmd)
		if err != nil {
			err = fmt.Errorf("%v: %s", err, stderr)
		}

		return stdout, stderr, err
	}

	stdout := &Tee{
		upstream: prefixw.New(os.Stdout, r.Prefix),
	}

	stderr := &Tee{
		upstream: prefixw.New(os.Stderr, r.Prefix),
	}

	// run the command
	_, err = r.Conn.Stream(cmd, stdout, stderr)

	stdout.Close()
	stderr.Close()

	return stdout.String(), stderr.String(), err
}

// WaitForPod waits for the availability of the given Kubernetes element.
func (r *Runner) WaitForPod(namespace string, name string, timeout time.Duration) error {
	cmd := fmt.Sprintf(`sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf -n "%s" get pod "%s" -o jsonpath='{.status.phase}' --ignore-not-found`, namespace, name)
	if !r.WaitForCondition(cmd, timeout, IsRunning) {
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
func (r *Runner) WaitForCondition(cmd string, timeout time.Duration, validator validatorFunc) bool {
	cutoff := time.Now().Add(timeout)

	for time.Now().Before(cutoff) {
		stdout, _, _ := r.Run(cmd, nil)
		if validator(stdout) {
			return true
		}

		time.Sleep(1 * time.Second)
	}

	return false
}

// prepareShell sets up the shell depending on the OS it's running on.
func (r *Runner) prepareShell(cmd string) string {
	// ensure we fail early
	cmd = fmt.Sprintf("set -xeu pipefail\n\n%s", cmd)

	// ensure sudo works on exotic distros
	cmd = fmt.Sprintf("export \"PATH=$PATH:/sbin:/usr/local/bin:/opt/bin\"\n\n%s", cmd)

	return cmd
}
