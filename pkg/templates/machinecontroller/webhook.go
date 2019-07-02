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

package machinecontroller

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/certificate"
	kubeonecontext "github.com/kubermatic/kubeone/pkg/util/context"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	certutil "k8s.io/client-go/util/cert"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineController Webhook related constants
const (
	WebhookName          = "machine-controller-webhook"
	WebhookAppLabelKey   = "app"
	WebhookAppLabelValue = WebhookName
	WebhookTag           = MachineControllerTag
	WebhookNamespace     = metav1.NamespaceSystem
)

// DeployWebhookConfiguration deploys MachineController webhook deployment on the cluster
func DeployWebhookConfiguration(ctx *kubeonecontext.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes clientset not initialized")
	}

	// Generate Webhook certificate
	caPrivateKey, caCert, err := certificate.CAKeyPair(ctx.Configuration)
	if err != nil {
		return errors.Wrap(err, "failed to load CA keypair")
	}

	bgCtx := context.Background()

	// Deploy Webhook
	err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, webhookDeployment(ctx.Cluster))
	if err != nil {
		return errors.Wrap(err, "failed to ensure machine-controller webhook deployment")
	}

	// Deploy Webhook service
	err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, service())
	if err != nil {
		return errors.Wrap(err, "failed to ensure machine-controller webhook service")
	}

	// Deploy serving certificate secret
	servingCert, err := tlsServingCertificate(caPrivateKey, caCert)
	if err != nil {
		return errors.Wrap(err, "failed to generate machine-controller webhook TLS secret")
	}

	err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, servingCert)
	if err != nil {
		return errors.Wrap(err, "failed to ensure machine-controller webhook secret")
	}

	err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, mutatingwebhookConfiguration(caCert))
	if err != nil {
		return errors.Wrap(err, "failed to ensure machine-controller mutating webhook")
	}

	return nil
}

// WaitForWebhook waits for machine-controller-webhook to become running
func WaitForWebhook(client dynclient.Client) error {
	listOpts := dynclient.ListOptions{
		Namespace: WebhookNamespace,
	}
	err := listOpts.SetLabelSelector(fmt.Sprintf("%s=%s", WebhookAppLabelKey, WebhookAppLabelValue))
	if err != nil {
		return errors.Wrap(err, "failed to parse machine-controller labels")
	}

	return wait.Poll(5*time.Second, 3*time.Minute, func() (bool, error) {
		webhookPods := corev1.PodList{}
		err = client.List(context.Background(), &listOpts, &webhookPods)
		if err != nil {
			return false, errors.Wrap(err, "failed to list machine-controller's webhook pods")
		}

		if len(webhookPods.Items) == 0 {
			return false, nil
		}

		whpod := webhookPods.Items[0]

		if whpod.Status.Phase == corev1.PodRunning {
			for _, podcond := range whpod.Status.Conditions {
				if podcond.Type == corev1.PodReady && podcond.Status == corev1.ConditionTrue {
					return true, nil
				}
			}
		}

		return false, nil
	})
}

// webhookDeployment returns the deployment for the machine-controllers MutatignAdmissionWebhook
func webhookDeployment(cluster *kubeoneapi.KubeOneCluster) *appsv1.Deployment {
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
	}

	dep.Name = "machine-controller-webhook"
	dep.Namespace = WebhookNamespace
	dep.Labels = map[string]string{
		WebhookAppLabelKey: WebhookAppLabelValue,
	}
	dep.Spec.Replicas = int32Ptr(1)
	dep.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			WebhookAppLabelKey: WebhookAppLabelValue,
		},
	}
	dep.Spec.Strategy.Type = appsv1.RollingUpdateStatefulSetStrategyType
	dep.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
		MaxSurge: &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: 1,
		},
		MaxUnavailable: &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: 0,
		},
	}

	// TODO: Why whould we need this?
	// dep.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: resources.ImagePullSecretName}}

	volumes := []corev1.Volume{getServingCertVolume()}
	dep.Spec.Template.Spec.Volumes = volumes
	dep.Spec.Template.ObjectMeta = metav1.ObjectMeta{
		Labels: map[string]string{
			WebhookAppLabelKey: WebhookAppLabelValue,
		},
	}

	dep.Spec.Template.Spec.Tolerations = []corev1.Toleration{
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoSchedule,
		},
		{
			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
			Value:  "true",
			Effect: corev1.TaintEffectNoSchedule,
		},
		{
			Key:      "CriticalAddonsOnly",
			Operator: corev1.TolerationOpExists,
		},
	}

	dep.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:            "machine-controller-webhook",
			Image:           "kubermatic/machine-controller:" + WebhookTag,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/usr/local/bin/webhook"},
			Args: []string{
				// "-kubeconfig", "/etc/kubernetes/kubeconfig/kubeconfig",
				"-logtostderr",
				"-v", "4",
				"-listen-address", "0.0.0.0:9876",
			},
			Env:                      getEnvVarCredentials(cluster),
			TerminationMessagePath:   corev1.TerminationMessagePathDefault,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/healthz",
						Port:   intstr.FromInt(9876),
						Scheme: corev1.URISchemeHTTPS,
					},
				},
				FailureThreshold: 3,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				TimeoutSeconds:   15,
			},
			LivenessProbe: &corev1.Probe{
				FailureThreshold: 8,
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/healthz",
						Port:   intstr.FromInt(9876),
						Scheme: corev1.URISchemeHTTPS,
					},
				},
				InitialDelaySeconds: 15,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      15,
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "machinecontroller-webhook-serving-cert",
					MountPath: "/tmp/cert",
					ReadOnly:  true,
				},
			},
		},
	}

	return dep
}

// service returns the internal service for the machine-controller webhook
func service() *corev1.Service {
	se := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
	}

	se.Name = "machine-controller-webhook"
	se.Namespace = WebhookNamespace
	se.Labels = map[string]string{
		WebhookAppLabelKey: WebhookAppLabelValue,
	}
	se.Spec.Type = corev1.ServiceTypeClusterIP
	se.Spec.Selector = map[string]string{
		WebhookAppLabelKey: WebhookAppLabelValue,
	}
	se.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "",
			Port:       443,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromInt(9876),
		},
	}

	return se
}

func getServingCertVolume() corev1.Volume {
	return corev1.Volume{
		Name: "machinecontroller-webhook-serving-cert",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  "machinecontroller-webhook-serving-cert",
				DefaultMode: int32Ptr(0444),
			},
		},
	}
}

// tlsServingCertificate returns a secret with the machine-controller-webhook tls certificate
// func tlsServingCertificate(ca *triple.KeyPair) (*corev1.Secret, error) {
func tlsServingCertificate(caKey *rsa.PrivateKey, caCert *x509.Certificate) (*corev1.Secret, error) {
	se := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
	}

	se.Name = "machinecontroller-webhook-serving-cert"
	se.Namespace = WebhookNamespace
	se.Data = map[string][]byte{}

	commonName := fmt.Sprintf("%s.%s.svc.cluster.local.", WebhookName, WebhookNamespace)
	altdnsNames := []string{
		commonName,
		fmt.Sprintf("%s.%s.svc", WebhookName, WebhookNamespace),
	}

	newKPKey, err := certutil.NewPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate private key")
	}

	certCfg := certutil.Config{
		AltNames: certutil.AltNames{
			DNSNames: altdnsNames,
		},
		CommonName: commonName,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	newKPCert, err := certutil.NewSignedCert(certCfg, newKPKey, caCert, caKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate certificate")
	}

	se.Data["cert.pem"] = certutil.EncodeCertPEM(newKPCert)
	se.Data["key.pem"] = certutil.EncodePrivateKeyPEM(newKPKey)
	// Include the CA for simplicity
	se.Data["ca.crt"] = certutil.EncodeCertPEM(caCert)

	return se, nil
}

// mutatingwebhookConfiguration returns the MutatingwebhookConfiguration for the machine controler
func mutatingwebhookConfiguration(caCert *x509.Certificate) *admissionregistrationv1beta1.MutatingWebhookConfiguration {
	cfg := &admissionregistrationv1beta1.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1beta1",
			Kind:       "MutatingWebhookConfiguration",
		},
	}

	cfg.Name = "machine-controller.kubermatic.io"
	cfg.Namespace = WebhookNamespace

	cfg.Webhooks = []admissionregistrationv1beta1.Webhook{
		{
			Name:              "machine-controller.kubermatic.io-machinedeployments",
			NamespaceSelector: &metav1.LabelSelector{},
			FailurePolicy:     failurePolicyPtr(admissionregistrationv1beta1.Fail),
			Rules: []admissionregistrationv1beta1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1beta1.OperationType{
						admissionregistrationv1beta1.Create,
						admissionregistrationv1beta1.Update,
					},
					Rule: admissionregistrationv1beta1.Rule{
						APIGroups:   []string{"cluster.k8s.io"},
						APIVersions: []string{"v1alpha1"},
						Resources:   []string{"machinedeployments"},
					},
				},
			},
			ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
				Service: &admissionregistrationv1beta1.ServiceReference{
					Name:      WebhookName,
					Namespace: WebhookNamespace,
					Path:      strPtr("/machinedeployments"),
				},
				CABundle: certutil.EncodeCertPEM(caCert),
			},
		},
		{
			Name:              "machine-controller.kubermatic.io-machines",
			NamespaceSelector: &metav1.LabelSelector{},
			FailurePolicy:     failurePolicyPtr(admissionregistrationv1beta1.Fail),
			Rules: []admissionregistrationv1beta1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1beta1.OperationType{
						admissionregistrationv1beta1.Create,
						admissionregistrationv1beta1.Update,
					},
					Rule: admissionregistrationv1beta1.Rule{
						APIGroups:   []string{"cluster.k8s.io"},
						APIVersions: []string{"v1alpha1"},
						Resources:   []string{"machines"},
					},
				},
			},
			ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
				Service: &admissionregistrationv1beta1.ServiceReference{
					Name:      WebhookName,
					Namespace: WebhookNamespace,
					Path:      strPtr("/machines"),
				},
				CABundle: certutil.EncodeCertPEM(caCert),
			},
		},
	}

	return cfg
}

func int32Ptr(i int32) *int32 {
	return &i
}

func strPtr(a string) *string {
	return &a
}

func failurePolicyPtr(a admissionregistrationv1beta1.FailurePolicyType) *admissionregistrationv1beta1.FailurePolicyType {
	return &a
}
