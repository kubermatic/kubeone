package installer

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/installer/version/kube112"
	"github.com/kubermatic/kubeone/pkg/pkigen"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/sirupsen/logrus"
	k8scert "k8s.io/client-go/util/cert"
)

// Options groups the various possible options for running
// the Kubernetes installation.
type Options struct {
	Verbose        bool
	BackupFile     string
	DestroyWorkers bool
}

// Installer is entrypoint for installation process
type Installer struct {
	cluster *config.Cluster
	logger  *logrus.Logger
}

// NewInstaller returns a new installer, responsible for dispatching
// between the different supported Kubernetes versions and running the
func NewInstaller(cluster *config.Cluster, logger *logrus.Logger) *Installer {
	return &Installer{
		cluster: cluster,
		logger:  logger,
	}
}

// Install run the installation process
func (i *Installer) Install(options *Options) error {
	var err error

	ctx := i.createContext(options)

	v, err := semver.NewVersion(i.cluster.Versions.Kubernetes)
	if err != nil {
		return fmt.Errorf("can't parse kubernetes version: %v", err)
	}

	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	switch majorMinor {
	case "1.12":
		err = kube112.Install(ctx)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	return err
}

// Reset resets cluster:
// * destroys all the worker machines
// * kubeadm reset masters
func (i *Installer) Reset(options *Options) error {
	var err error

	ctx := i.createContext(options)
	if err = generatePKI(ctx.Configuration); err != nil {
		return fmt.Errorf("can't generate CA: %v", err)
	}

	v := semver.MustParse(i.cluster.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	switch majorMinor {
	case "1.12":
		err = kube112.Reset(ctx)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	return err
}

// createContext creates a basic, non-host bound context with
// all relevant information, but *no* Runner yet. The various
// task helper functions will take care of setting up Runner
// structs for each task individually.
func (i *Installer) createContext(options *Options) *util.Context {
	return &util.Context{
		Cluster:        i.cluster,
		Connector:      ssh.NewConnector(),
		Configuration:  util.NewConfiguration(),
		WorkDir:        "kubeone",
		Logger:         i.logger,
		Verbose:        options.Verbose,
		BackupFile:     options.BackupFile,
		DestroyWorkers: options.DestroyWorkers,
	}
}

func generatePKI(cfg *util.Configuration) error {
	caparams := []struct {
		cn, cert, key string
	}{
		{cn: "kubernetes", cert: "pki/ca.crt", key: "pki/ca.key"},
		{cn: "front-proxy-ca", cert: "pki/front-proxy-ca.crt", key: "pki/front-proxy-ca.key"},
		{cn: "etcd-ca", cert: "pki/etcd/ca.crt", key: "pki/etcd/ca.key"},
	}

	for _, caparam := range caparams {
		key, err := k8scert.NewPrivateKey()
		if err != nil {
			return err
		}
		ca, err := pkigen.NewCA(caparam.cn, key)
		if err != nil {
			return err
		}
		cfg.AddFile(caparam.key, ca.Key())
		cfg.AddFile(caparam.cert, ca.Certificate())
	}

	saKey, err := k8scert.NewPrivateKey()
	if err != nil {
		return err
	}
	saPubPem, err := k8scert.EncodePublicKeyPEM(&saKey.PublicKey)
	if err != nil {
		return err
	}
	cfg.AddFile("pki/sa.key", string(k8scert.EncodePrivateKeyPEM(saKey)))
	cfg.AddFile("pki/sa.pub", string(saPubPem))

	return nil
}
