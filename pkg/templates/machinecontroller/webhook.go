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

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/state"

	admissionregistration "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	certutil "k8s.io/client-go/util/cert"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineController Webhook related constants
const (
	mcWebhookName          = "machine-controller-webhook"
	mcWebhookAppLabelKey   = "app"
	mcWebhookAppLabelValue = mcWebhookName
	mcWebhookNamespace     = metav1.NamespaceSystem
	mcWebhookPort          = 9876
)

// deployMachineControllerWebhook deploys MachineController webhook deployment on the cluster
func deployMachineControllerWebhook(s *state.State) error {
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

	image := s.Cluster.RegistryConfiguration.ImageRegistry(mcImageRegistry) + mcImage + Tag

	deployment, err := webhookDeployment(s.Cluster, s.CredentialsFilePath, image)
	if err != nil {
		return errors.Wrap(err, "failed to generate machine-controller webhook deployment")
	}

	ctx := context.Background()

	k8sobjects := []dynclient.Object{
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
func waitForWebhook(ctx context.Context, client dynclient.Client) error {
	condFn := clientutil.PodsReadyCondition(ctx, client, dynclient.ListOptions{
		Namespace: mcWebhookNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			mcWebhookAppLabelKey: mcWebhookAppLabelValue,
		}),
	})

	return wait.Poll(5*time.Second, 3*time.Minute, condFn)
}

// webhookDeployment returns the deployment for the machine-controllers MutatignAdmissionWebhook
func webhookDeployment(cluster *kubeoneapi.KubeOneCluster, credentialsFilePath, image string) (*appsv1.Deployment, error) {
	var replicas int32 = 1

	envVar, err := credentials.EnvVarBindings(cluster.CloudProvider, credentialsFilePath)
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

	args := []string{
		"-logtostderr",
		"-v", "4",
		"-listen-address", fmt.Sprintf("0.0.0.0:%d", mcWebhookPort),
	}

	if cluster.CloudProvider.External {
		args = append(args, "-node-external-cloud-provider")
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller-webhook",
			Namespace: mcWebhookNamespace,
			Labels: map[string]string{
				mcWebhookAppLabelKey: mcWebhookAppLabelValue,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					mcWebhookAppLabelKey: mcWebhookAppLabelValue,
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
						mcWebhookAppLabelKey: mcWebhookAppLabelValue,
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
							Name:                     "machine-controller-webhook",
							Image:                    image,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							Command:                  []string{"/usr/local/bin/webhook"},
							Args:                     args,
							Env:                      envVar,
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(mcWebhookPort),
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
										Port:   intstr.FromInt(mcWebhookPort),
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
	}

	certificate.CABundleInjector(cluster.CABundle, &dep.Spec.Template)
	if cluster.CABundle != "" {
		dep.Spec.Template.Spec.Containers[0].Args = append(
			dep.Spec.Template.Spec.Containers[0].Args, "-ca-bundle", certificate.CABundlePath,
		)
	}

	return dep, nil
}

// service returns the internal service for the machine-controller webhook
func service() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller-webhook",
			Namespace: mcWebhookNamespace,
			Labels: map[string]string{
				mcWebhookAppLabelKey: mcWebhookAppLabelValue,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				mcWebhookAppLabelKey: mcWebhookAppLabelValue,
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
	commonName := fmt.Sprintf("%s.%s.svc.cluster.local.", mcWebhookName, mcWebhookNamespace)
	altdnsNames := []string{
		commonName,
		fmt.Sprintf("%s.%s.svc", mcWebhookName, mcWebhookNamespace),
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
			Namespace: mcWebhookNamespace,
		},
		Data: map[string][]byte{
			"cert.pem": certificate.EncodeCertPEM(newKPCert),
			"key.pem":  certificate.EncodePrivateKeyPEM(newKPKey),
			"ca.crt":   certificate.EncodeCertPEM(caCert),
		},
	}, nil
}

// mutatingwebhookConfiguration returns the MutatingwebhookConfiguration for the machine controler
func mutatingwebhookConfiguration(caCert *x509.Certificate) *admissionregistration.MutatingWebhookConfiguration {
	sideEffectsNone := admissionregistration.SideEffectClassNone

	return &admissionregistration.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller.kubermatic.io",
			Namespace: mcWebhookNamespace,
		},
		Webhooks: []admissionregistration.MutatingWebhook{
			{
				Name:                    "machine-controller.kubermatic.io-machinedeployments",
				NamespaceSelector:       &metav1.LabelSelector{},
				FailurePolicy:           failurePolicyPtr(admissionregistration.Fail),
				SideEffects:             &sideEffectsNone,
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				Rules: []admissionregistration.RuleWithOperations{
					{
						Operations: []admissionregistration.OperationType{
							admissionregistration.Create,
							admissionregistration.Update,
						},
						Rule: admissionregistration.Rule{
							APIGroups:   []string{"cluster.k8s.io"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"machinedeployments"},
						},
					},
				},
				ClientConfig: admissionregistration.WebhookClientConfig{
					Service: &admissionregistration.ServiceReference{
						Name:      mcWebhookName,
						Namespace: mcWebhookNamespace,
						Path:      strPtr("/machinedeployments"),
					},
					CABundle: certificate.EncodeCertPEM(caCert),
				},
			},
			{
				Name:                    "machine-controller.kubermatic.io-machines",
				NamespaceSelector:       &metav1.LabelSelector{},
				FailurePolicy:           failurePolicyPtr(admissionregistration.Fail),
				SideEffects:             &sideEffectsNone,
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				Rules: []admissionregistration.RuleWithOperations{
					{
						Operations: []admissionregistration.OperationType{
							admissionregistration.Create,
							admissionregistration.Update,
						},
						Rule: admissionregistration.Rule{
							APIGroups:   []string{"cluster.k8s.io"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"machines"},
						},
					},
				},
				ClientConfig: admissionregistration.WebhookClientConfig{
					Service: &admissionregistration.ServiceReference{
						Name:      mcWebhookName,
						Namespace: mcWebhookNamespace,
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

func failurePolicyPtr(a admissionregistration.FailurePolicyType) *admissionregistration.FailurePolicyType {
	return &a
}
