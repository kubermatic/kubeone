package machinecontroller

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/certificate"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/cert/triple"
)

const (
	WebbhookName         = "machine-controller-webhook"
	WebhookAppLabelKey   = "app"
	WebhookAppLabelValue = WebbhookName
	WebhookTag           = "v0.10.0"
	WebhookNamespace     = "kube-system"

	WebhookCredentialsSecretName = "machine-controller-credentials"
)

func WebhookConfiguration(cluster *config.Cluster, runtimeConfig *util.Configuration) (string, error) {
	caKeyPair, err := certificate.CAKeyPair(runtimeConfig)
	if err != nil {
		return "", fmt.Errorf("failed to load CA keypair: %v", err)
	}

	deployment, err := WebhookDeployment(cluster)
	if err != nil {
		return "", err
	}

	service, err := Service()
	if err != nil {
		return "", err
	}

	servingCert, err := TLSServingCertificate(caKeyPair)
	if err != nil {
		return "", err
	}

	config, err := MutatingwebhookConfiguration(caKeyPair)
	if err != nil {
		return "", err
	}

	return templates.KubernetesToYAML([]interface{}{
		deployment,
		service,
		servingCert,
		config,
	})
}

// WebhookDeployment returns the deployment for the machine-controllers MutatignAdmissionWebhook
func WebhookDeployment(cluster *config.Cluster) (*appsv1.Deployment, error) {
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

	return dep, nil
}

// Service returns the internal service for the machine-controller webhook
func Service() (*corev1.Service, error) {
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

	return se, nil
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

// TLSServingCertificate returns a secret with the machine-controller-webhook tls certificate
func TLSServingCertificate(ca *triple.KeyPair) (*corev1.Secret, error) {
	se := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
	}

	se.Name = "machinecontroller-webhook-serving-cert"
	se.Namespace = WebhookNamespace
	se.Data = map[string][]byte{}

	commonName := fmt.Sprintf("%s.%s.svc.cluster.local.", WebbhookName, WebhookNamespace)

	newKP, err := triple.NewServerKeyPair(ca,
		commonName,
		WebbhookName,
		WebhookNamespace,
		"",
		nil,
		// For some reason the name the APIServer validates against must be in the SANs, having it as CN is not enough
		[]string{commonName})
	if err != nil {
		return nil, fmt.Errorf("failed to generate serving cert: %v", err)
	}
	se.Data["cert.pem"] = certutil.EncodeCertPEM(newKP.Cert)
	se.Data["key.pem"] = certutil.EncodePrivateKeyPEM(newKP.Key)
	// Include the CA for simplicity
	se.Data["ca.crt"] = certutil.EncodeCertPEM(ca.Cert)

	return se, nil
}

// MutatingwebhookConfiguration returns the MutatingwebhookConfiguration for the machine controler
func MutatingwebhookConfiguration(ca *triple.KeyPair) (*admissionregistrationv1beta1.MutatingWebhookConfiguration, error) {
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
					Name:      WebbhookName,
					Namespace: WebhookNamespace,
					Path:      strPtr("/machinedeployments"),
				},
				CABundle: certutil.EncodeCertPEM(ca.Cert),
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
					Name:      WebbhookName,
					Namespace: WebhookNamespace,
					Path:      strPtr("/machines"),
				},
				CABundle: certutil.EncodeCertPEM(ca.Cert),
			},
		},
	}

	return cfg, nil
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
