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

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate/cabundle"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/scripts"
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

	renewCmd := "sudo kubeadm certs renew all"
	err := s.RunTaskOnControlPlane(
		func(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
			_, _, err := s.Runner.RunRaw(renewCmd)

			return fail.SSH(err, "running %q on %s node", renewCmd, node.PublicAddress)
		},
		state.RunParallel,
	)
	if err != nil {
		return err
	}

	return kubeconfig.BuildKubernetesClientset(s)
}

func fetchCert(virtfs fs.FS, filename string) (*x509.Certificate, error) {
	buf, err := fs.ReadFile(virtfs, filename)
	if err != nil {
		return nil, err
	}

	pemBlock, rest := pem.Decode(buf)
	if len(rest) != 0 {
		return nil, fail.Runtime(fmt.Errorf("returned non-zero rest"), "PEM decoding %q certificate", filename)
	}

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, fail.Runtime(err, "parsing %q certificate", filename)
	}

	return cert, nil
}

func timeBefore(t1 time.Time, t2 time.Time) bool {
	if t2.IsZero() {
		return true
	}

	return t1.Before(t2)
}

func earliestCertExpiry(conn executor.Interface) (time.Time, error) {
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

	virtfs := executorfs.New(conn)
	for _, certName := range certsToCheck {
		cert, err := fetchCert(virtfs, certName)
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
		return fail.NoKubeClient()
	}

	s.Logger.Infoln("Creating ca-bundle configMap...")
	cm := cabundle.ConfigMap(s.Cluster.CABundle)

	return clientutil.CreateOrUpdate(s.Context, s.DynamicClient, cm)
}

func saveCABundle(s *state.State) error {
	s.Configuration.AddFile("ca-certs/"+cabundle.FileName, s.Cluster.CABundle)

	return s.RunTaskOnControlPlane(saveCABundleOnControlPlane, state.RunParallel)
}

func saveCABundleOnControlPlane(s *state.State, _ *kubeoneapi.HostConfig, conn executor.Interface) error {
	if err := s.Configuration.UploadTo(conn, s.WorkDir); err != nil {
		return err
	}

	cmd, err := scripts.SaveCABundle(s.WorkDir)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return fail.SSH(err, "save CABundle")
}

func restartKubelet(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
	s.Logger.WithField("node", node.PublicAddress).Debug("Restarting Kubelet to force regenerating CSRs...")

	_, _, err := s.Runner.RunRaw(scripts.RestartKubelet())

	return fail.SSH(err, "restart Kubelet")
}

func restartKubeletOnControlPlane(s *state.State) error {
	s.Logger.Infof("Restarting Kubelet on control plane nodes to force Kubelet to generate correct CSRs...")

	// Restart Kubelet on all control plane nodes to force CSRs to be regenerated
	if err := s.RunTaskOnControlPlane(restartKubelet, state.RunParallel); err != nil {
		return err
	}

	// Wait 40 seconds to give Kubelet time to come up and generate correct CSRs.
	// NB: We'll wait 20 seconds on the next step, so that's one minute in total
	// which should be enough.
	sleepTime := 40 * time.Second
	s.Logger.Infof("Waiting %s to give Kubelet time to regenerate CSRs...", sleepTime)
	time.Sleep(sleepTime)

	return nil
}

func approvePendingCSR(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
	var csrFound bool
	sleepTime := 20 * time.Second
	s.Logger.Infof("Waiting %s for CSRs to approve...", sleepTime)
	time.Sleep(sleepTime)

	csrList := certificatesv1.CertificateSigningRequestList{}
	if err := s.DynamicClient.List(s.Context, &csrList); err != nil {
		return fail.KubeClient(err, "getting %T", csrList)
	}

	certv1Client, err := certificatesv1client.NewForConfig(s.RESTConfig)
	if err != nil {
		return fail.KubeClient(err, "creating certificates v1client")
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
			// CSR matched but it's already approved
			csrFound = true

			continue
		}

		if err := validateCSR(csr.Spec); err != nil {
			return err
		}
		csrFound = true

		csr := csr.DeepCopy()
		csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
			Type:   certificatesv1.CertificateApproved,
			Reason: "kubeone approved node serving cert",
			Status: corev1.ConditionTrue,
		})

		s.Logger.Infof("Approve pending CSR %q for username %q", csr.Name, csr.Spec.Username)
		_, err := certClient.UpdateApproval(s.Context, csr.Name, csr, metav1.UpdateOptions{})
		if err != nil {
			return fail.KubeClient(err, "approving CSR %q", csr.Name)
		}
	}

	if !csrFound {
		s.Logger.Infof("No CSR found for node %q, assuming it was garbage-collected", node.Hostname)
	}

	return nil
}

func validateCSR(spec certificatesv1.CertificateSigningRequestSpec) error {
	if !sets.NewString(spec.Groups...).HasAll(groupNodes, groupAuthenticated) {
		return fail.Runtime(errors.New("CSR groups is expecter to be an authenticated node"), "")
	}

	for _, usage := range spec.Usages {
		if !isUsageInUsageList(usage, allowedUsages) {
			return fail.Runtime(errors.New("CSR usages is invalid"), "")
		}
	}

	csrBlock, rest := pem.Decode(spec.Request)
	if csrBlock == nil {
		return fail.Runtime(errors.New("no certificate request found for the given CSR"), "")
	}

	if len(rest) != 0 {
		return fail.Runtime(errors.New("found more than one PEM encoded block in the result"), "")
	}

	certReq, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return fail.Runtime(err, "parsing kubelet %q CSR", spec.Username)
	}

	if certReq.Subject.CommonName != spec.Username {
		return fail.RuntimeError{
			Op:  "checking match between CSR subject CN and CSR username",
			Err: errors.Errorf("commonName %q is different then CSR username %q", certReq.Subject.CommonName, spec.Username),
		}
	}

	if len(certReq.Subject.Organization) != 1 {
		return fail.RuntimeError{
			Op:  "checking match between CSR subject CN and CSR username",
			Err: errors.Errorf("expected only one organization but got %d instead", len(certReq.Subject.Organization)),
		}
	}

	if certReq.Subject.Organization[0] != groupNodes {
		return fail.RuntimeError{
			Op:  "checking match between CSR subject CN and CSR username",
			Err: errors.Errorf("organization %q doesn't match node group %q", certReq.Subject.Organization[0], groupNodes),
		}
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
