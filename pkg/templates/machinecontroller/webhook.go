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
	"crypto"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/certificate"
	"github.com/kubermatic/kubeone/pkg/clientutil"
	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/state"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
func DeployWebhookConfiguration(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes clientset not initialized")
	}

	// Generate Webhook certificate
	caPrivateKey, caCert, err := certificate.CAKeyPair(s.Configuration)
	if err != nil {
		return errors.Wrap(err, "failed to load CA keypair")
	}

	// Generate serving certificate secret
	servingCert, err := tlsServingCertificate(caPrivateKey, caCert)
	if err != nil {
		return errors.Wrap(err, "failed to generate machine-controller webhook TLS secret")
	}

	deployment, err := webhookDeployment(s.Cluster, s.CredentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to generate machine-controller webhook deployment")
	}

	ctx := context.Background()

	k8sobjects := []runtime.Object{
		deployment,
		service(),
		servingCert,
		mutatingwebhookConfiguration(caCert),
	}

	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj); err != nil {
			return errors.Wrapf(err, "failed to ensure machine-controller webhook %T", obj)
		}
	}

	return nil
}

// WaitForWebhook waits for machine-controller-webhook to become running
func WaitForWebhook(client dynclient.Client) error {
	listOpts := dynclient.ListOptions{
		Namespace: WebhookNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			WebhookAppLabelKey: WebhookAppLabelValue,
		}),
	}

	return wait.Poll(5*time.Second, 3*time.Minute, func() (bool, error) {
		webhookPods := corev1.PodList{}
		err := client.List(context.Background(), &webhookPods, &listOpts)
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
func webhookDeployment(cluster *kubeoneapi.KubeOneCluster, credentialsFilePath string) (*appsv1.Deployment, error) {
	var replicas int32 = 1

	envVar, err := credentials.EnvVarBindings(cluster.CloudProvider.Name, credentialsFilePath)
	envVar = append(envVar,
		corev1.EnvVar{
			Name:  "HTTPS_PROXY",
			Value: cluster.Proxy.HTTPS,
		},
		corev1.EnvVar{
			Name:  "NO_PROXY",
			Value: cluster.Proxy.NoProxy,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get env var bindings for a secret")
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller-webhook",
			Namespace: WebhookNamespace,
			Labels: map[string]string{
				WebhookAppLabelKey: WebhookAppLabelValue,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					WebhookAppLabelKey: WebhookAppLabelValue,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 0,
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						WebhookAppLabelKey: WebhookAppLabelValue,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
					Volumes: []corev1.Volume{
						getServingCertVolume(),
					},
					Tolerations: []corev1.Toleration{
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
					},
					Containers: []corev1.Container{
						{
							Name:            "machine-controller-webhook",
							Image:           "kubermatic/machine-controller:" + WebhookTag,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/usr/local/bin/webhook"},
							Args: []string{
								"-logtostderr",
								"-v", "4",
								"-listen-address", "0.0.0.0:9876",
							},
							Env:                      envVar,
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
					},
				},
			},
		},
	}, nil
}

// service returns the internal service for the machine-controller webhook
func service() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller-webhook",
			Namespace: WebhookNamespace,
			Labels: map[string]string{
				WebhookAppLabelKey: WebhookAppLabelValue,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				WebhookAppLabelKey: WebhookAppLabelValue,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "",
					Port:       443,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(9876),
				},
			},
		},
	}
}

func getServingCertVolume() corev1.Volume {
	var mode int32 = 0444

	return corev1.Volume{
		Name: "machinecontroller-webhook-serving-cert",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  "machinecontroller-webhook-serving-cert",
				DefaultMode: &mode,
			},
		},
	}
}

// tlsServingCertificate returns a secret with the machine-controller-webhook tls certificate
// func tlsServingCertificate(ca *triple.KeyPair) (*corev1.Secret, error) {
func tlsServingCertificate(caKey crypto.Signer, caCert *x509.Certificate) (*corev1.Secret, error) {
	commonName := fmt.Sprintf("%s.%s.svc.cluster.local.", WebhookName, WebhookNamespace)
	altdnsNames := []string{
		commonName,
		fmt.Sprintf("%s.%s.svc", WebhookName, WebhookNamespace),
	}

	newKPKey, err := certificate.NewPrivateKey()
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

	newKPCert, err := certificate.NewSignedCert(&certCfg, newKPKey, caCert, caKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate certificate")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machinecontroller-webhook-serving-cert",
			Namespace: WebhookNamespace,
		},
		Data: map[string][]byte{
			"cert.pem": certificate.EncodeCertPEM(newKPCert),
			"key.pem":  certificate.EncodePrivateKeyPEM(newKPKey),
			"ca.crt":   certificate.EncodeCertPEM(caCert),
		},
	}, nil
}

// mutatingwebhookConfiguration returns the MutatingwebhookConfiguration for the machine controler
func mutatingwebhookConfiguration(caCert *x509.Certificate) *admissionregistrationv1beta1.MutatingWebhookConfiguration {
	return &admissionregistrationv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller.kubermatic.io",
			Namespace: WebhookNamespace,
		},
		Webhooks: []admissionregistrationv1beta1.MutatingWebhook{
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
					CABundle: certificate.EncodeCertPEM(caCert),
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
					CABundle: certificate.EncodeCertPEM(caCert),
				},
			},
		},
	}
}

func strPtr(a string) *string {
	return &a
}

func failurePolicyPtr(a admissionregistrationv1beta1.FailurePolicyType) *admissionregistrationv1beta1.FailurePolicyType {
	return &a
}
