package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const socketEnvPrefix = "env:"

// Connection represents an established connection to an SSH server.
type Connection interface {
	Close() error
	Closed() bool

	Exec(string) (string, string, int, error)
	Stream(string, io.Writer, io.Writer) (int, error)

	Upload(io.Reader, int64, os.FileMode, string) error
	UploadFile(string, string) error

	Download(string, io.Writer) error
	DownloadToFile(string, string) error
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

const (
	defaultTimeout = 60 * time.Second
	defaultSSHPort = 22
)

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
			return o, fmt.Errorf("failed to read keyfile '%s': %v", o.KeyFile, err)
		}

		o.PrivateKey = string(content)
		o.KeyFile = ""
	}

	if o.Port <= 0 || o.Port > 65535 {
		o.Port = defaultSSHPort
	}

	if o.Timeout == 0 {
		o.Timeout = defaultTimeout
	}

	return o, nil
}

type connection struct {
	client *ssh.Client
}

// NewConnection attempts to create a new SSH connection to the host
// specified via the given options.
func NewConnection(o Opts) (Connection, error) {
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
			return nil, fmt.Errorf("the given SSH key could not be parsed (note that password-protected keys are not supported): %v", parseErr)
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
			return nil, fmt.Errorf("could not open socket '%s': %v", addr, dialErr)
		}

		agentClient := agent.NewClient(socket)

		signers, signersErr := agentClient.Signers()
		if signersErr != nil {
			socket.Close()
			return nil, fmt.Errorf("error when creating signer for SSH agent: %v", signersErr)
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
		return nil, fmt.Errorf("could not establish connection to %s: %v", endpoint, err)
	}

	return &connection{client}, nil
}

func (c *connection) Close() error {
	var err error
	if !c.Closed() {
		err = c.client.Close()
	}

	c.client = nil
	return err
}

func (c *connection) Closed() bool {
	return c.client == nil
}

func (c *connection) Stream(cmd string, stdout io.Writer, stderr io.Writer) (int, error) {
	if c.client == nil {
		return 0, errors.New("cannot exec commands because connection was already closed")
	}

	ses, err := c.client.NewSession()
	if err != nil {
		return 0, fmt.Errorf("failed to create new SSH session: %v", err)
	}
	defer ses.Close()

	ses.Stdout = stdout
	ses.Stderr = stderr

	err = ses.Run(strings.TrimSpace(cmd))
	if err != nil {
		return 1, fmt.Errorf("failed to exec command: %v", err)
	}
	return 0, err
}

func (c *connection) Exec(cmd string) (string, string, int, error) {
	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)

	exitCode, err := c.Stream(cmd, &stdoutBuf, &stderrBuf)

	return strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()), exitCode, err
}

func cpy(source io.Reader, size int64, mode os.FileMode, destination string) func(*ssh.Session) error {
	return func(ses *ssh.Session) error {
		filename := filepath.Base(destination)

		return scp.Copy(size, mode, filename, source, destination, ses)
	}
}

func cpyPath(source string, destination string) func(*ssh.Session) error {
	return func(ses *ssh.Session) error {
		return scp.CopyPath(source, destination, ses)
	}
}

func (c *connection) upload(upload func(*ssh.Session) error) error {
	if c.client == nil {
		return errors.New("cannot transfer files because connection was already closed")
	}

	ses, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create new SSH session: %v", err)
	}
	defer ses.Close()

	err = upload(ses)

	if err != nil {
		err = fmt.Errorf("failed to transfer file: %v", err)
	}

	return err
}

func (c *connection) Upload(source io.Reader, size int64, mode os.FileMode, destination string) error {
	return c.upload(cpy(source, size, mode, destination))
}

func (c *connection) UploadFile(source string, destination string) error {
	return c.upload(cpyPath(source, destination))
}

func (c *connection) Download(source string, target io.Writer) error {
	if c.client == nil {
		return errors.New("cannot transfer files because connection was already closed")
	}

	var stderrBuf bytes.Buffer

	_, err := c.Stream(fmt.Sprintf(`cat -- "%s"`, source), target, &stderrBuf)
	if err != nil {
		err = fmt.Errorf("failed to transfer file: %v", err)
	}

	return err
}

func (c *connection) DownloadToFile(source string, target string) error {
	if c.client == nil {
		return errors.New("cannot transfer files because connection was already closed")
	}

	f, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer f.Close()

	return c.Download(source, f)
}
