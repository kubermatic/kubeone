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

package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
)

const socketEnvPrefix = "env:"

var (
	_ executor.Tunneler = &connection{}
)

// Opts represents all the possible options for connecting to
// a remote server via SSH.
type Opts struct {
	Context     context.Context
	Username    string
	Password    string
	Hostname    string
	Port        int
	PrivateKey  string
	KeyFile     string
	AgentSocket string
	Timeout     time.Duration
	Bastion     string
	BastionPort int
	BastionUser string
}

func validateOptions(o Opts) (Opts, error) {
	if len(o.Username) == 0 {
		return o, fail.ConfigValidation(errors.New("no username specified for SSH connection"))
	}

	if len(o.Hostname) == 0 {
		return o, fail.ConfigValidation(errors.New("no hostname specified for SSH connection"))
	}

	if len(o.Password) == 0 && len(o.PrivateKey) == 0 && len(o.KeyFile) == 0 && len(o.AgentSocket) == 0 {
		return o, fail.ConfigValidation(errors.New("must specify at least one of password, private key, keyfile or agent socket"))
	}

	if len(o.KeyFile) > 0 {
		content, err := os.ReadFile(o.KeyFile)
		if err != nil {
			return o, fail.Config(err, "reading SSH private key")
		}

		o.PrivateKey = string(content)
		o.KeyFile = ""
	}

	if o.Port <= 0 {
		o.Port = 22
	}

	if o.BastionPort <= 0 {
		o.BastionPort = 22
	}

	if o.BastionUser == "" {
		o.BastionUser = o.Username
	}

	if o.Timeout == 0 {
		o.Timeout = 60 * time.Second
	}

	return o, nil
}

type connection struct {
	mu        sync.Mutex
	sshclient *ssh.Client
	connector *Connector
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewConnection attempts to create a new SSH connection to the host
// specified via the given options.
func NewConnection(connector *Connector, o Opts) (executor.Interface, error) {
	o, err := validateOptions(o)
	if err != nil {
		return nil, err
	}

	authMethods := make([]ssh.AuthMethod, 0)

	if len(o.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(o.Password))
	}

	if len(o.PrivateKey) > 0 {
		signer, parseErr := ssh.ParsePrivateKey([]byte(o.PrivateKey))
		if parseErr != nil {
			return nil, fail.SSHError{
				Op:  "parsing private key",
				Err: errors.Wrap(parseErr, "SSH key could not be parsed (note that password-protected keys are not supported)"),
			}
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(o.AgentSocket) > 0 {
		addr := o.AgentSocket

		if strings.HasPrefix(o.AgentSocket, socketEnvPrefix) {
			envName := strings.TrimPrefix(o.AgentSocket, socketEnvPrefix)

			if envAddr := os.Getenv(envName); len(envAddr) > 0 {
				addr = envAddr
			}
		}

		socket, dialErr := net.Dial("unix", addr)
		if dialErr != nil {
			return nil, fail.SSHError{
				Op:  "agent unix dialing",
				Err: errors.Wrapf(dialErr, "could not open socket %q", addr),
			}
		}

		agentClient := agent.NewClient(socket)

		signers, signersErr := agentClient.Signers()
		if signersErr != nil {
			socket.Close()

			return nil, fail.SSHError{
				Op:  "creating signer for SSH agent",
				Err: signersErr,
			}
		} else if len(signers) == 0 {
			socket.Close()

			return nil, fail.SSHError{
				Err: errors.New("could not retrieve signers"),
				Op:  "creating signer for SSH agent",
			}
		}

		authMethods = append(authMethods, ssh.PublicKeys(signers...))
	}

	sshConfig := &ssh.ClientConfig{
		User:            o.Username,
		Timeout:         o.Timeout,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	targetHost := o.Hostname
	targetPort := strconv.Itoa(o.Port)

	if o.Bastion != "" {
		targetHost = o.Bastion
		targetPort = strconv.Itoa(o.BastionPort)
		sshConfig.User = o.BastionUser
	}

	// do not use fmt.Sprintf() to allow proper IPv6 handling if hostname is an IP address
	endpoint := net.JoinHostPort(targetHost, targetPort)

	client, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return nil, fail.SSH(fail.Connection(err, endpoint), "dialing")
	}

	ctx, cancelFn := context.WithCancel(connector.ctx)
	sshConn := &connection{
		connector: connector,
		ctx:       ctx,
		cancel:    cancelFn,
	}

	if o.Bastion == "" {
		sshConn.sshclient = client
		// connection established
		return sshConn, nil
	}

	// continue to setup if we are running over bastion
	endpointBehindBastion := net.JoinHostPort(o.Hostname, strconv.Itoa(o.Port))

	// Dial a connection to the service host, from the bastion
	conn, err := client.Dial("tcp", endpointBehindBastion)
	if err != nil {
		return nil, fail.SSH(fail.Connection(err, endpointBehindBastion), "dialing behind the bastion")
	}

	sshConfig.User = o.Username
	ncc, chans, reqs, err := ssh.NewClientConn(conn, endpointBehindBastion, sshConfig)
	if err != nil {
		return nil, fail.SSH(fail.Connection(err, endpointBehindBastion), "new client")
	}

	sshConn.sshclient = ssh.NewClient(ncc, chans, reqs)

	return sshConn, nil
}

func (c *connection) TunnelTo(_ context.Context, network, addr string) (net.Conn, error) {
	// the voided context.Context is voided as a workaround of always Done
	// context that being passed. Please don't try to <-ctx.Done(), it will
	// always return immediately
	netconn, err := c.sshclient.Dial(network, addr)
	if err != nil {
		return nil, fail.SSH(fail.Connection(err, addr), "tunneling")
	}

	go func() {
		<-c.ctx.Done()
		netconn.Close()
	}()

	return netconn, nil
}

func (c *connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil
	}
	c.cancel()

	defer func() { c.sshclient = nil }()
	defer c.connector.forgetConnection(c)

	return c.sshclient.Close()
}

func (c *connection) POpen(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	sess, err := c.session()
	if err != nil {
		return 0, err
	}
	defer sess.Close()

	sess.Stdin = stdin
	sess.Stdout = stdout
	sess.Stderr = stderr

	exitCode := 0
	if err = sess.Run(cmd); err != nil {
		exitCode = -1
		var errSSH *ssh.ExitError
		if errors.As(err, &errSSH) {
			exitCode = errSSH.ExitStatus()
		}
	}

	// preserve original error
	return exitCode, fail.SSH(err, "popen")
}

func (c *connection) Exec(cmd string) (string, string, int, error) {
	var (
		stdoutBuf, stderrBuf strings.Builder
		returnErr            error
	)

	exitCode, err := c.POpen(cmd, nil, &stdoutBuf, &stderrBuf)

	stdout := strings.TrimSpace(stdoutBuf.String())
	stderr := stderrBuf.String()

	if err != nil {
		returnErr = fail.SSHError{
			Err:    err,
			Op:     "exec",
			Stderr: stderr,
			Cmd:    cmd,
		}
	}

	return stdout, stderr, exitCode, returnErr
}

func (c *connection) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, fail.SSH(fmt.Errorf("connection closed"), "session")
	}

	sess, err := c.sshclient.NewSession()
	if err != nil {
		return nil, fail.SSH(err, "new session")
	}

	return sess, nil
}
