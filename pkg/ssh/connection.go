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
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const socketEnvPrefix = "env:"

var (
	_ Tunneler = &connection{}
)

// Connection represents an established connection to an SSH server.
type Connection interface {
	Exec(cmd string) (stdout string, stderr string, exitCode int, err error)
	POpen(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (exitCode int, err error)
	io.Closer
}

// Tunneler interface creates net.Conn originating from the remote ssh host to
// target `addr`
type Tunneler interface {
	// `network` can be tcp, tcp4, tcp6, unix
	TunnelTo(ctx context.Context, network, addr string) (net.Conn, error)
	io.Closer
}

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
		return o, errors.New("no username specified for SSH connection")
	}

	if len(o.Hostname) == 0 {
		return o, errors.New("no hostname specified for SSH connection")
	}

	if len(o.Password) == 0 && len(o.PrivateKey) == 0 && len(o.KeyFile) == 0 && len(o.AgentSocket) == 0 {
		return o, errors.New("must specify at least one of password, private key, keyfile or agent socket")
	}

	if len(o.KeyFile) > 0 {
		content, err := ioutil.ReadFile(o.KeyFile)
		if err != nil {
			return o, errors.Wrapf(err, "failed to read keyfile %q", o.KeyFile)
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
func NewConnection(connector *Connector, o Opts) (Connection, error) {
	o, err := validateOptions(o)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate ssh connection options")
	}

	authMethods := make([]ssh.AuthMethod, 0)

	if len(o.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(o.Password))
	}

	if len(o.PrivateKey) > 0 {
		signer, parseErr := ssh.ParsePrivateKey([]byte(o.PrivateKey))
		if parseErr != nil {
			return nil, errors.Wrap(parseErr, "the given SSH key could not be parsed (note that password-protected keys are not supported)")
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
			return nil, errors.Wrapf(dialErr, "could not open socket %q", addr)
		}

		agentClient := agent.NewClient(socket)

		signers, signersErr := agentClient.Signers()
		if signersErr != nil {
			socket.Close()
			return nil, errors.Wrap(signersErr, "error when creating signer for SSH agent")
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
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpoint)
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
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpointBehindBastion)
	}

	sshConfig.User = o.Username
	ncc, chans, reqs, err := ssh.NewClientConn(conn, endpointBehindBastion, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpointBehindBastion)
	}

	sshConn.sshclient = ssh.NewClient(ncc, chans, reqs)
	return sshConn, nil
}

func (c *connection) TunnelTo(_ context.Context, network, addr string) (net.Conn, error) {
	// the voided context.Context is voided as a workaround of always Done
	// context that being passed. Please don't try to <-ctx.Done(), it will
	// always return immediately
	netconn, err := c.sshclient.Dial(network, addr)
	if err == nil {
		go func() {
			<-c.ctx.Done()
			netconn.Close()
		}()
	}
	return netconn, err
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
		return 0, errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	sess.Stdin = stdin
	sess.Stdout = stdout
	sess.Stderr = stderr

	exitCode := 0
	if err = sess.Run(cmd); err != nil {
		exitCode = -1
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}

	// preserve original error
	return exitCode, err
}

func (c *connection) Exec(cmd string) (string, string, int, error) {
	var stdoutBuf, stderrBuf strings.Builder

	exitCode, err := c.POpen(cmd, nil, &stdoutBuf, &stderrBuf)

	return strings.TrimSpace(stdoutBuf.String()), stderrBuf.String(), exitCode, err
}

func (c *connection) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}

	return c.sshclient.NewSession()
}
