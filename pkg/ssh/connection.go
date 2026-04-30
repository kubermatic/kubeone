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

var _ executor.Tunneler = &connection{}

// Opts represents all the possible options for connecting to
// a remote server via SSH.
type Opts struct {
	Context               context.Context
	Username              string
	Password              string
	Hostname              string
	Port                  int
	PrivateKey            []byte
	KeyFile               string
	SSHCert               string
	SSHCertFile           string
	HostPublicKey         []byte
	AgentSocket           string
	Timeout               time.Duration
	Bastion               string
	BastionPort           int
	BastionUser           string
	BastionHostPublicKey  []byte
	BastionPrivateKeyFile string
	BastionPrivateKey     []byte
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

		o.PrivateKey = content
		o.KeyFile = ""
	}

	if len(o.BastionPrivateKeyFile) > 0 {
		content, err := os.ReadFile(o.BastionPrivateKeyFile)
		if err != nil {
			return o, fail.Config(err, "reading bastion SSH private key")
		}

		o.BastionPrivateKey = content
		o.BastionPrivateKeyFile = ""
	}

	if len(o.SSHCertFile) > 0 {
		content, err := os.ReadFile(o.SSHCertFile)
		if err != nil {
			return o, fail.Config(err, "reading SSH signed public key")
		}

		o.SSHCert = string(content)
		o.SSHCertFile = ""
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
func NewConnection(connector *Connector, opts Opts) (executor.Interface, error) {
	opts, err := validateOptions(opts)
	if err != nil {
		return nil, err
	}

	// nodeAuthMethods are used to authenticate to the target node.
	nodeAuthMethods := make([]ssh.AuthMethod, 0)

	if len(opts.Password) > 0 {
		nodeAuthMethods = append(nodeAuthMethods, ssh.Password(opts.Password))
	}

	if len(opts.PrivateKey) > 0 {
		signer, parseErr := ssh.ParsePrivateKey(opts.PrivateKey)
		if parseErr != nil {
			return nil, fail.SSHError{
				Op:  "parsing private key",
				Err: errors.Wrap(parseErr, "SSH key could not be parsed (note that password-protected keys are not supported)"),
			}
		}

		if len(opts.SSHCert) > 0 {
			cert, _, _, _, certParseErr := ssh.ParseAuthorizedKey([]byte(opts.SSHCert))
			if certParseErr != nil {
				return nil, fail.SSHError{
					Op:  "parsing certificate",
					Err: errors.Wrapf(certParseErr, "SSH certificate could not be parsed"),
				}
			}

			sshCert, ok := cert.(*ssh.Certificate)
			if !ok {
				return nil, fail.SSHError{
					Op:  "cert.(*ssh.Certificate) type asserting",
					Err: errors.New("type asserting failed"),
				}
			}
			// create a signer using both the certificate and the private key:
			certSigner, signersErr := ssh.NewCertSigner(sshCert, signer)
			if signersErr != nil {
				return nil, fail.SSHError{
					Op:  "creating new signer with private key and certificate",
					Err: signersErr,
				}
			}

			nodeAuthMethods = append(nodeAuthMethods, ssh.PublicKeys(certSigner))
		} else {
			nodeAuthMethods = append(nodeAuthMethods, ssh.PublicKeys(signer))
		}
	}

	// bastionAuthMethods are used exclusively for the bastion hop. When a
	// dedicated bastion key is provided we use only that key so we never waste
	// MaxAuthTries attempts on the node key (which the bastion does not know).
	// If no bastion key is configured we fall back to the node auth methods.
	bastionAuthMethods := nodeAuthMethods

	if len(opts.BastionPrivateKey) > 0 {
		signer, parseErr := ssh.ParsePrivateKey(opts.BastionPrivateKey)
		if parseErr != nil {
			return nil, fail.SSHError{
				Op:  "parsing private key",
				Err: errors.Wrap(parseErr, "bastion SSH key could not be parsed (note that password-protected keys are not supported)"),
			}
		}
		bastionAuthMethods = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

	if len(opts.AgentSocket) > 0 {
		addr := opts.AgentSocket

		if after, ok := strings.CutPrefix(opts.AgentSocket, socketEnvPrefix); ok {
			envName := after

			if envAddr := os.Getenv(envName); len(envAddr) > 0 {
				addr = envAddr
			}
		}

		var dialer net.Dialer
		dialCtx, dialCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dialCancel()

		socket, dialErr := dialer.DialContext(dialCtx, "unix", addr)
		if dialErr != nil {
			return nil, fail.SSHError{
				Op:  "agent unix dialing",
				Err: errors.Wrapf(dialErr, "could not open socket %q", addr),
			}
		}

		agentClient := agent.NewClient(socket)

		signers, signersErr := agentClient.Signers()
		switch {
		case signersErr != nil:
			socket.Close()

			return nil, fail.SSHError{
				Op:  "creating signer for SSH agent",
				Err: signersErr,
			}
		case len(signers) == 0:
			socket.Close()
		default:
			nodeAuthMethods = append(nodeAuthMethods, ssh.PublicKeys(signers...))
			// only add agent keys to bastion if no dedicated bastion key was set
			if len(opts.BastionPrivateKey) == 0 {
				bastionAuthMethods = nodeAuthMethods
			}
		}
	}

	nodeConfig := &ssh.ClientConfig{
		User:            opts.Username,
		Timeout:         opts.Timeout,
		Auth:            nodeAuthMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	if opts.HostPublicKey != nil {
		nodeConfig.HostKeyCallback = hostKeyCallback(opts.HostPublicKey)
	}

	// No bastion: connect directly to the node.
	if opts.Bastion == "" {
		endpoint := net.JoinHostPort(opts.Hostname, strconv.Itoa(opts.Port))

		client, dialErr := ssh.Dial("tcp", endpoint, nodeConfig)
		if dialErr != nil {
			return nil, fail.SSH(fail.Connection(dialErr, endpoint), "dialing")
		}

		ctx, cancelFn := context.WithCancel(connector.ctx)

		return &connection{
			sshclient: client,
			connector: connector,
			ctx:       ctx,
			cancel:    cancelFn,
		}, nil
	}

	// Connect to the bastion with its own dedicated auth methods.
	bastionConfig := &ssh.ClientConfig{
		User:            opts.BastionUser,
		Timeout:         opts.Timeout,
		Auth:            bastionAuthMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	if opts.BastionHostPublicKey != nil {
		bastionConfig.HostKeyCallback = hostKeyCallback(opts.BastionHostPublicKey)
	}

	bastionEndpoint := net.JoinHostPort(opts.Bastion, strconv.Itoa(opts.BastionPort))

	bastionClient, err := ssh.Dial("tcp", bastionEndpoint, bastionConfig)
	if err != nil {
		return nil, fail.SSH(fail.Connection(err, bastionEndpoint), "dialing")
	}

	// Dial a connection to the service host, from the bastion
	endpointBehindBastion := net.JoinHostPort(opts.Hostname, strconv.Itoa(opts.Port))

	conn, err := bastionClient.Dial("tcp", endpointBehindBastion)
	if err != nil {
		return nil, fail.SSH(fail.Connection(err, endpointBehindBastion), "dialing behind the bastion")
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, endpointBehindBastion, nodeConfig)
	if err != nil {
		return nil, fail.SSH(fail.Connection(err, endpointBehindBastion), "new client")
	}

	ctx, cancelFn := context.WithCancel(connector.ctx)

	return &connection{
		sshclient: ssh.NewClient(ncc, chans, reqs),
		connector: connector,
		ctx:       ctx,
		cancel:    cancelFn,
	}, nil
}

func hostKeyCallback(knownKey []byte) ssh.HostKeyCallback {
	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		if !bytes.Equal(key.Marshal(), knownKey) {
			return fmt.Errorf("ssh: host key mismatch")
		}

		return nil
	}
}

func (c *connection) TunnelTo(_ context.Context, network, addr string) (net.Conn, error) {
	// the voided context.Context is voided as a workaround of always Done
	// context that being passed. Please don't try to <-ctx.Done(), it will
	// always return immediately
	if c.sshclient == nil {
		return nil, fail.SSHError{
			Err: fail.Connection(errors.New("no SSH connection established"), addr),
			Op:  "tunneling",
		}
	}

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
		c.Close()

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
