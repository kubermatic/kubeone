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

package tasks

import (
	"crypto/x509"
	"encoding/pem"
	"io/fs"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate/cabundle"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/ssh/sshiofs"
	"k8c.io/kubeone/pkg/state"
)

func renewControlPlaneCerts(s *state.State) error {
	if !s.ForceUpgrade {
		s.Logger.Warn("Your control-plane certificates are about to expire in less then 90 days")
		s.Logger.Warn("To renew them without changing kubernetes version run `kubeone apply --force-upgrade`")
		return nil
	}
	s.Logger.Warn("Your control-plane certificates are about to expire in less then 90 days")
	s.Logger.Warn("Force renewing Kubernetes certificates")

	// /etc/kubernetes/admin.conf will be changed after certificates renew, so we have to initialize client again
	s.Logger.Infoln("Resetting Kubernetes clientset...")
	s.DynamicClient = nil

	renewCmd := "sudo kubeadm alpha certs renew all"
	greaterThen120, _ := semver.NewConstraint(">=1.20")
	if greaterThen120.Check(s.LiveCluster.ExpectedVersion) {
		renewCmd = "sudo kubeadm certs renew all"
	}

	err := s.RunTaskOnControlPlane(
		func(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
			_, _, err := s.Runner.RunRaw(renewCmd)
			return err
		},
		state.RunParallel,
	)
	if err != nil {
		return err
	}

	return kubeconfig.BuildKubernetesClientset(s)
}

func fetchCert(sshfs fs.FS, filename string) (*x509.Certificate, error) {
	buf, err := fs.ReadFile(sshfs, filename)
	if err != nil {
		return nil, err
	}

	pemBlock, rest := pem.Decode(buf)
	if len(rest) != 0 {
		return nil, errors.New("returned non-zero rest")
	}

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func timeBefore(t1 time.Time, t2 time.Time) bool {
	if t2.IsZero() {
		return true
	}

	return t1.Before(t2)
}

func earliestCertExpiry(conn ssh.Connection) (time.Time, error) {
	var (
		earliestCertExpirationTime time.Time

		certsToCheck = []string{
			"/etc/kubernetes/pki/apiserver-etcd-client.crt",
			"/etc/kubernetes/pki/apiserver-kubelet-client.crt",
			"/etc/kubernetes/pki/apiserver.crt",
			"/etc/kubernetes/pki/ca.crt",
			"/etc/kubernetes/pki/etcd/ca.crt",
			"/etc/kubernetes/pki/etcd/healthcheck-client.crt",
			"/etc/kubernetes/pki/etcd/peer.crt",
			"/etc/kubernetes/pki/etcd/server.crt",
			"/etc/kubernetes/pki/front-proxy-ca.crt",
			"/etc/kubernetes/pki/front-proxy-client.crt",
		}
	)

	sshfs := sshiofs.New(conn)
	for _, certName := range certsToCheck {
		cert, err := fetchCert(sshfs, certName)
		if err != nil {
			return earliestCertExpirationTime, err
		}

		if timeBefore(cert.NotAfter, earliestCertExpirationTime) {
			earliestCertExpirationTime = cert.NotAfter
		}
	}

	return earliestCertExpirationTime, nil
}

func ensureCABundleConfigMap(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	s.Logger.Infoln("Creating ca-bundle configMap...")

	cm := cabundle.ConfigMap(s.Cluster.CABundle)
	return clientutil.CreateOrUpdate(s.Context, s.DynamicClient, cm)
}

func saveCABundle(s *state.State) error {
	s.Configuration.AddFile("ca-certs/"+cabundle.FileName, s.Cluster.CABundle)

	return s.RunTaskOnControlPlane(saveCABundleOnControlPlane, state.RunParallel)
}

func saveCABundleOnControlPlane(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	if err := s.Configuration.UploadTo(conn, s.WorkDir); err != nil {
		return errors.Wrap(err, "failed to upload")
	}

	cmd, err := scripts.SaveCABundle(s.WorkDir)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return err
}
