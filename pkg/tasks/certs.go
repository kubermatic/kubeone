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
	"fmt"
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

	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	certificatesv1client "k8s.io/client-go/kubernetes/typed/certificates/v1"
)

const (
	nodeUser = "system:node"

	groupNodes         = "system:nodes"
	groupAuthenticated = "system:authenticated"
)

var (
	allowedUsages = []certificatesv1.KeyUsage{
		certificatesv1.UsageDigitalSignature,
		certificatesv1.UsageKeyEncipherment,
		certificatesv1.UsageServerAuth,
	}
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

func approvePendingCSR(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	approveErr := errors.Errorf("no CSR found for node %q", node.Hostname)

	sleepTime := 20 * time.Second
	s.Logger.Infof("Waiting %s for CSRs to approve...", sleepTime)
	time.Sleep(sleepTime)

	csrList := certificatesv1.CertificateSigningRequestList{}
	if err := s.DynamicClient.List(s.Context, &csrList); err != nil {
		return err
	}

	certv1Client, err := certificatesv1client.NewForConfig(s.RESTConfig)
	if err != nil {
		return err
	}
	certClient := certv1Client.CertificateSigningRequests()

	for _, csr := range csrList.Items {
		if csr.Spec.SignerName != certificatesv1.KubeletServingSignerName {
			continue
		}

		if fmt.Sprintf("%s:%s", nodeUser, node.Hostname) != csr.Spec.Username {
			// that's not the CSR we are looking for
			continue
		}

		var approved bool
		for _, cond := range csr.Status.Conditions {
			if cond.Type == certificatesv1.CertificateApproved && cond.Status == corev1.ConditionTrue {
				approved = true
			}
		}
		if approved {
			// CSR matched but it's already approved, no need to raise an error
			approveErr = nil

			continue
		}

		if err := validateCSR(csr.Spec); err != nil {
			return fmt.Errorf("failed to validate CSR: %w", err)
		}
		approveErr = nil

		csr := csr.DeepCopy()
		csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
			Type:   certificatesv1.CertificateApproved,
			Reason: "kubeone approved node serving cert",
			Status: corev1.ConditionTrue,
		})

		s.Logger.Infof("Approve pending CSR %q for username %q", csr.Name, csr.Spec.Username)
		_, err := certClient.UpdateApproval(s.Context, csr.Name, csr, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to approve CSR %q: %w", csr.Name, err)
		}
	}

	return approveErr
}

func validateCSR(spec certificatesv1.CertificateSigningRequestSpec) error {
	if !sets.NewString(spec.Groups...).HasAll(groupNodes, groupAuthenticated) {
		return errors.New("CSR groups is expecter to be an authenticated node")
	}

	for _, usage := range spec.Usages {
		if !isUsageInUsageList(usage, allowedUsages) {
			return errors.New("CSR usages is invalid")
		}
	}

	csrBlock, rest := pem.Decode(spec.Request)
	if csrBlock == nil {
		return errors.New("no certificate request found for the given CSR")
	}

	if len(rest) != 0 {
		return errors.New("found more than one PEM encoded block in the result")
	}

	certReq, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return err
	}

	if certReq.Subject.CommonName != spec.Username {
		return fmt.Errorf("commonName %q is different then CSR username %q", certReq.Subject.CommonName, spec.Username)
	}

	if len(certReq.Subject.Organization) != 1 {
		return fmt.Errorf("expected only one organization but got %d instead", len(certReq.Subject.Organization))
	}

	if certReq.Subject.Organization[0] != groupNodes {
		return fmt.Errorf("organization %q doesn't match node group %q", certReq.Subject.Organization[0], groupNodes)
	}

	return nil
}

func isUsageInUsageList(usage certificatesv1.KeyUsage, usageList []certificatesv1.KeyUsage) bool {
	for _, usageListItem := range usageList {
		if usage == usageListItem {
			return true
		}
	}

	return false
}
