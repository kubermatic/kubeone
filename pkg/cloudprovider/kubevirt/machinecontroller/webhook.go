/*
Copyright 2026 The KubeOne Authors.

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

package machinecontroller

import (
	"bytes"
	"context"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"maps"
	"net"
	"slices"
	"strconv"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/clientutil"
	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates/resources"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	webhookPort       = 443
	webhookTargetPort = 9876

	// webhookConfigurationName is the name of the tenant MutatingWebhookConfiguration,
	// the same as in the machinecontroller addon.
	webhookConfigurationName = "machinedeployments.machine-controller.kubermatic.io"

	// servingCertMinValidity is how long a serving certificate has to stay
	// valid to be reused.
	servingCertMinValidity = 30 * 24 * time.Hour

	lbWaitInterval = 5 * time.Second
	lbWaitTimeout  = 5 * time.Minute
)

// webhookServiceType returns the configured webhook Service type. By default,
// a ClusterIP Service is used when the control plane runs as VMs in the infra
// cluster, as the API server can then reach infra ClusterIPs. Otherwise, a
// LoadBalancer is needed.
func webhookServiceType(spec *kubeoneapi.KubevirtSpec) corev1.ServiceType {
	if spec.MachineController != nil && spec.MachineController.Webhook.ServiceType != "" {
		return spec.MachineController.Webhook.ServiceType
	}

	if spec.ControlPlane != nil {
		return corev1.ServiceTypeClusterIP
	}

	return corev1.ServiceTypeLoadBalancer
}

func webhookPodLabels(clusterName string) map[string]string {
	l := kubevirtinfra.Labels(clusterName, WebhookComponent.Name)
	l["app.kubernetes.io/name"] = WebhookComponent.Name

	return l
}

// WebhookService returns the infra Service exposing the webhook to the tenant
// API server.
func WebhookService(clusterName, namespace string, spec *kubeoneapi.KubevirtSpec) *corev1.Service {
	var annotations map[string]string
	if spec.MachineController != nil {
		annotations = maps.Clone(spec.MachineController.Webhook.Annotations)
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        kubevirtinfra.ObjectName(clusterName, WebhookComponent.Name),
			Namespace:   namespace,
			Labels:      kubevirtinfra.Labels(clusterName, WebhookComponent.Name),
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     webhookServiceType(spec),
			Selector: webhookPodLabels(clusterName),
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Port:       webhookPort,
					TargetPort: intstr.FromInt32(webhookTargetPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

// webhookHost returns the address of the webhook Service as seen from the
// tenant API server: the ClusterIP, or the LoadBalancer ingress address.
func webhookHost(ctx context.Context, c ctrlrt.Client, key ctrlrt.ObjectKey) (string, error) {
	var host string

	err := wait.PollUntilContextTimeout(ctx, lbWaitInterval, lbWaitTimeout, true, func(ctx context.Context) (bool, error) {
		svc := &corev1.Service{}
		if err := c.Get(ctx, key, svc); err != nil {
			return false, err
		}

		host = serviceHost(svc)

		return host != "", nil
	})

	return host, fail.KubeClient(err, "waiting for the address of the machine-controller webhook Service %s", key)
}

func serviceHost(svc *corev1.Service) string {
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				return ingress.IP
			}

			if ingress.Hostname != "" {
				return ingress.Hostname
			}
		}

		return ""
	}

	if svc.Spec.ClusterIP == corev1.ClusterIPNone {
		return ""
	}

	return svc.Spec.ClusterIP
}

// webhookURL returns the URL of a webhook path.
func webhookURL(host, path string) string {
	return "https://" + net.JoinHostPort(host, strconv.Itoa(webhookPort)) + path
}

func servingCertSecretName(clusterName string) string {
	return kubevirtinfra.ObjectName(clusterName, WebhookComponent.Name) + "-serving-cert"
}

type servingCert struct {
	// hash of the serving certificate, set on the webhook pods so they restart
	// when it changes.
	hash string
	// caPEM is the CA to put into the tenant MutatingWebhookConfiguration.
	caPEM []byte
}

// ensureServingCert ensures the webhook serving certificate Secret holds a
// certificate for host signed by the tenant cluster CA. A valid existing
// certificate is reused, so the webhook does not restart on every apply.
func ensureServingCert(
	ctx context.Context,
	c ctrlrt.Client,
	clusterName, namespace, host string,
	caKey crypto.Signer,
	caCert *x509.Certificate,
) (servingCert, error) {
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw})
	key := ctrlrt.ObjectKey{Namespace: namespace, Name: servingCertSecretName(clusterName)}

	existing := &corev1.Secret{}
	if err := c.Get(ctx, key, existing); ctrlrt.IgnoreNotFound(err) != nil {
		return servingCert{}, fail.KubeClient(err, "getting %s", key)
	}

	certPEM := existing.Data[resources.TLSCertName]
	keyPEM := existing.Data[resources.TLSKeyName]

	if !servingCertValid(certPEM, host, caCert, time.Now()) || len(keyPEM) == 0 {
		var (
			dnsNames []string
			ips      []net.IP
		)

		if ip := net.ParseIP(host); ip != nil {
			ips = []net.IP{ip}
		} else {
			dnsNames = []string{host}
		}

		certs, err := certificate.NewSignedTLSCert(host, dnsNames, ips, caKey, caCert)
		if err != nil {
			return servingCert{}, err
		}

		certPEM = []byte(certs[resources.TLSCertName])
		keyPEM = []byte(certs[resources.TLSKeyName])

		secret := &corev1.Secret{
			Name:      key.Name,
			Namespace: key.Namespace,
			Labels:    kubevirtinfra.Labels(clusterName, WebhookComponent.Name),
			Type:      corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				resources.TLSCertName: certPEM,
				resources.TLSKeyName:  keyPEM,
			},
		}

		if err = clientutil.CreateOrReplace(ctx, c, secret); err != nil {
			return servingCert{}, err
		}
	}

	sum := sha256.Sum256(bytes.Join([][]byte{certPEM, keyPEM}, nil))

	return servingCert{hash: hex.EncodeToString(sum[:]), caPEM: caPEM}, nil
}

// servingCertValid reports whether certPEM is a certificate for host, signed
// by caCert and valid for at least servingCertMinValidity after now.
func servingCertValid(certPEM []byte, host string, caCert *x509.Certificate, now time.Time) bool {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	if cert.CheckSignatureFrom(caCert) != nil {
		return false
	}

	if now.Add(servingCertMinValidity).After(cert.NotAfter) {
		return false
	}

	if ip := net.ParseIP(host); ip != nil {
		return slices.ContainsFunc(cert.IPAddresses, ip.Equal)
	}

	return slices.Contains(cert.DNSNames, host)
}

// MutatingWebhookConfiguration returns the tenant MutatingWebhookConfiguration
// pointing at the webhook in the infra cluster by URL.
func MutatingWebhookConfiguration(host string, caPEM []byte) *admissionregistrationv1.MutatingWebhookConfiguration {
	webhook := func(resource, path string) admissionregistrationv1.MutatingWebhook {
		return admissionregistrationv1.MutatingWebhook{
			Name:                    fmt.Sprintf("%s.machine-controller.kubermatic.io", resource),
			FailurePolicy:           new(admissionregistrationv1.Fail),
			SideEffects:             new(admissionregistrationv1.SideEffectClassNone),
			AdmissionReviewVersions: []string{"v1"},
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations:  []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					APIGroups:   []string{"cluster.k8s.io"},
					APIVersions: []string{"v1alpha1"},
					Resources:   []string{resource},
				},
			},
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				URL:      new(webhookURL(host, path)),
				CABundle: caPEM,
			},
		}
	}

	return &admissionregistrationv1.MutatingWebhookConfiguration{
		Name: webhookConfigurationName,
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			webhook("machinedeployments", "/machinedeployments"),
			webhook("machines", "/machines"),
		},
	}
}
