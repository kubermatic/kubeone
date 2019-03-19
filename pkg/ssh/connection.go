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
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const socketEnvPrefix = "env:"

// Connection represents an established connection to an SSH server.
type Connection interface {
	Exec(cmd string) (stdout string, stderr string, exitCode int, err error)
	File(filename string, flags int) (io.ReadWriteCloser, error)
	Stream(cmd string, stdout io.Writer, stderr io.Writer) (exitCode int, err error)
	io.Closer
}

// Opts represents all the possible options for connecting to
// a remote server via SSH.
type Opts struct {
	Username    string
	Password    string
	Hostname    string
	Port        int
	PrivateKey  string
	KeyFile     string
	AgentSocket string
	Timeout     time.Duration
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

	if o.Timeout == 0 {
		o.Timeout = 60 * time.Second
	}

	return o, nil
}

type connection struct {
	mu         sync.Mutex
	sftpclient *sftp.Client
	sshclient  *ssh.Client
}

// NewConnection attempts to create a new SSH connection to the host
// specified via the given options.
func NewConnection(o Opts) (Connection, error) {
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
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// do not use fmt.Sprintf() to allow proper IPv6 handling if hostname is an IP address
	endpoint := net.JoinHostPort(o.Hostname, strconv.Itoa(o.Port))

	client, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpoint)
	}

	return &connection{sshclient: client}, nil
}

// File return remote file (as an io.ReadWriteCloser).
//
// mode is os package file modes: https://golang.org/pkg/os/#pkg-constants
// returned file optionally implement
func (c *connection) File(filename string, flags int) (io.ReadWriteCloser, error) {
	sftpClient, err := c.sftp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open SFTP")
	}

	return sftpClient.OpenFile(filename, flags)
}

func (c *connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil
	}
	defer func() { c.sshclient = nil }()
	defer func() { c.sftpclient = nil }()

	return c.sshclient.Close()
}

func (c *connection) Stream(cmd string, stdout io.Writer, stderr io.Writer) (int, error) {
	sess, err := c.session()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	sess.Stdout = stdout
	sess.Stderr = stderr

	exitCode := 0
	err = sess.Run(strings.TrimSpace(cmd))
	if err != nil {
		exitCode = 1
	}

	return exitCode, errors.Wrapf(err, "failed to exec command: %s", cmd)
}

func (c *connection) Exec(cmd string) (string, string, int, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	exitCode, err := c.Stream(cmd, &stdoutBuf, &stderrBuf)

	return strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()), exitCode, err
}

func (c *connection) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}

	return c.sshclient.NewSession()
}

func (c *connection) sftp() (*sftp.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}

	if c.sftpclient == nil {
		s, err := sftp.NewClient(c.sshclient)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get sftp.Client")
		}
		c.sftpclient = s
	}

	return c.sftpclient, nil
}
