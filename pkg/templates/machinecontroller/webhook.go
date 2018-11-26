package machinecontroller

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/certificate"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"
	kubermaticv1 "github.com/kubermatic/kubermatic/api/pkg/crd/kubermatic/v1"
	"github.com/kubermatic/kubermatic/api/pkg/resources"

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

	return templates.KubernetesToYAML([]interface{}{
		deployment,
		service,
		servingCert,
	})
}

// WebhookDeployment returns the deployment for the machine-controllers MutatignAdmissionWebhook
func WebhookDeployment(cluster *config.Cluster) (*appsv1.Deployment, error) {
	dep := &appsv1.Deployment{}

	dep.Name = "machine-controller-webhook"
	dep.Labels = map[string]string{
		WebhookAppLabelKey: WebhookAppLabelValue,
	}
	dep.Spec.Replicas = resources.Int32(1)
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
	dep.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: resources.ImagePullSecretName}}

	volumes := []corev1.Volume{getServingCertVolume()}
	dep.Spec.Template.Spec.Volumes = volumes
	dep.Spec.Template.ObjectMeta = metav1.ObjectMeta{
		Labels: map[string]string{
			WebhookAppLabelKey: WebhookAppLabelValue,
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
				// {
				// 	Name:      resources.MachineControllerKubeconfigSecretName,
				// 	MountPath: "/etc/kubernetes/kubeconfig",
				// 	ReadOnly:  true,
				// },
				{
					Name:      resources.MachineControllerWebhookServingCertSecretName,
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
	se := &corev1.Service{}

	se.Name = "machine-controller-webhook"
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
		Name: resources.MachineControllerWebhookServingCertSecretName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  resources.MachineControllerWebhookServingCertSecretName,
				DefaultMode: resources.Int32(0444),
			},
		},
	}
}

// TLSServingCertificate returns a secret with the machine-controller-webhook tls certificate
func TLSServingCertificate(ca *triple.KeyPair) (*corev1.Secret, error) {
	se := &corev1.Secret{}

	se.Name = "machinecontroller-webhook-serving-cert"
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
func MutatingwebhookConfiguration(c *kubermaticv1.Cluster, data *resources.TemplateData, existing *admissionregistrationv1beta1.MutatingWebhookConfiguration) (*admissionregistrationv1beta1.MutatingWebhookConfiguration, error) {
	mutatingWebhookConfiguration := &admissionregistrationv1beta1.MutatingWebhookConfiguration{}
	mutatingWebhookConfiguration.Name = resources.MachineControllerMutatingWebhookConfigurationName

	ca, err := data.GetRootCA()
	if err != nil {
		return nil, fmt.Errorf("failed to get root ca: %v", err)
	}

	mutatingWebhookConfiguration.Webhooks = []admissionregistrationv1beta1.Webhook{
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
				URL:      strPtr(fmt.Sprintf("https://%s.%s.svc.cluster.local./machinedeployments", WebbhookName, WebhookNamespace)),
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
				URL:      strPtr(fmt.Sprintf("https://%s.%s.svc.cluster.local./machines", WebbhookName, WebhookNamespace)),
				CABundle: certutil.EncodeCertPEM(ca.Cert),
			},
		},
	}

	return mutatingWebhookConfiguration, nil
}

func strPtr(a string) *string {
	return &a
}

func failurePolicyPtr(a admissionregistrationv1beta1.FailurePolicyType) *admissionregistrationv1beta1.FailurePolicyType {
	return &a
}
