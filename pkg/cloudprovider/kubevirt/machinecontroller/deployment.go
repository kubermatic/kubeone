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
	"maps"
	"strconv"

	"k8c.io/kubeone/pkg/certificate/cabundle"
	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/templates/images"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	imageMachineController = images.MachineController

	tenantVolumeName    = "tenantcluster"
	tenantKubeconfigDir = "/var/run/secrets/tenantcluster"
	tenantKubeconfig    = tenantKubeconfigDir + "/" + kubevirtinfra.KubeconfigSecretKey

	servingCertVolumeName = "serving-cert"
	servingCertDir        = "/tmp/cert"

	healthPort  = 8085
	metricsPort = 8080

	servingCertHashAnnotation = "kubeone.k8c.io/serving-cert-hash"
)

type deploymentOptions struct {
	clusterName           string
	namespace             string
	image                 string
	externalCloudProvider bool
	caBundle              string
	joinClusterTimeout    string
	servingCertHash       string
}

// commonArgs are the flags machine-controller and its webhook share.
func (o deploymentOptions) commonArgs() []string {
	args := []string{"-kubeconfig=" + tenantKubeconfig}

	if o.externalCloudProvider {
		args = append(args, "-node-external-cloud-provider")
	}

	if o.caBundle != "" {
		args = append(args, "-ca-bundle="+cabundle.SSLCertFilePath)
	}

	return args
}

// env returns the environment of machine-controller and its webhook. Proxy
// settings of the tenant cluster are intentionally not passed: they are for
// the tenant network, while these pods run in the infra cluster.
func (o deploymentOptions) env(comp kubevirtinfra.Component) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "POD_NAMESPACE", Value: o.namespace},
		infraKubeconfigEnv(o.clusterName, comp),
	}
}

func (o deploymentOptions) deployment(comp kubevirtinfra.Component, podLabels, annotations map[string]string, container corev1.Container, volumes []corev1.Volume) *appsv1.Deployment {
	podTpl := corev1.PodTemplateSpec{
		Labels:      podLabels,
		Annotations: annotations,
		Spec: corev1.PodSpec{
			ServiceAccountName: comp.ServiceAccountName(o.clusterName),
			Containers:         []corev1.Container{container},
			Volumes:            append([]corev1.Volume{tenantKubeconfigVolume(o.clusterName, comp)}, volumes...),
		},
	}

	cabundle.InjectFromConfigMap(o.caBundle, caBundleConfigMapName(o.clusterName), &podTpl)

	return &appsv1.Deployment{
		Name:      kubevirtinfra.ObjectName(o.clusterName, comp.Name),
		Namespace: o.namespace,
		Labels:    kubevirtinfra.Labels(o.clusterName, comp.Name),
		Spec: appsv1.DeploymentSpec{
			Replicas: new(int32(1)),
			Selector: &metav1.LabelSelector{MatchLabels: podLabels},
			Template: podTpl,
		},
	}
}

func tenantMount() corev1.VolumeMount {
	return corev1.VolumeMount{Name: tenantVolumeName, MountPath: tenantKubeconfigDir, ReadOnly: true}
}

func requests() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
	}
}

// ControllerDeployment returns the machine-controller Deployment running in
// the infra cluster. Leader election uses the tenant cluster, so it does not
// overlap with a machine-controller still running there.
func ControllerDeployment(o deploymentOptions) *appsv1.Deployment {
	podLabels := maps.Clone(kubevirtinfra.Labels(o.clusterName, ControllerComponent.Name))
	podLabels["app.kubernetes.io/name"] = ControllerComponent.Name

	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   strconv.Itoa(metricsPort),
		"prometheus.io/path":   "/metrics",
	}

	args := append(
		o.commonArgs(),
		"-health-probe-address=0.0.0.0:"+strconv.Itoa(healthPort),
		"-metrics-address=0.0.0.0:"+strconv.Itoa(metricsPort),
		"-node-csr-approver",
		"-join-cluster-timeout="+o.joinClusterTimeout,
	)

	container := corev1.Container{
		Name:    resources.MachineControllerName,
		Image:   o.image,
		Command: []string{"/usr/local/bin/machine-controller"},
		Args:    args,
		Env:     o.env(ControllerComponent),
		Ports: []corev1.ContainerPort{
			{Name: "health", ContainerPort: healthPort, Protocol: corev1.ProtocolTCP},
			{Name: "metrics", ContainerPort: metricsPort, Protocol: corev1.ProtocolTCP},
		},
		LivenessProbe: &corev1.Probe{
			HTTPGet:             &corev1.HTTPGetAction{Path: "/readyz", Port: intstr.FromInt32(healthPort)},
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
		},
		ReadinessProbe: &corev1.Probe{
			HTTPGet:       &corev1.HTTPGetAction{Path: "/healthz", Port: intstr.FromInt32(healthPort)},
			PeriodSeconds: 5,
		},
		VolumeMounts: []corev1.VolumeMount{tenantMount()},
		Resources:    requests(),
	}

	return o.deployment(ControllerComponent, podLabels, annotations, container, nil)
}

// WebhookDeployment returns the machine-controller webhook Deployment running
// in the infra cluster.
func WebhookDeployment(o deploymentOptions) *appsv1.Deployment {
	annotations := map[string]string{
		servingCertHashAnnotation: o.servingCertHash,
	}

	args := append(
		o.commonArgs(),
		"-listen-address=0.0.0.0:"+strconv.Itoa(webhookTargetPort),
		// The namespace of the Machines in the tenant cluster
		"-namespace="+resources.MachineControllerNameSpace,
		"-tls-cert-path="+servingCertDir+"/"+resources.TLSCertName,
		"-tls-key-path="+servingCertDir+"/"+resources.TLSKeyName,
	)

	probe := func() *corev1.Probe {
		return &corev1.Probe{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   intstr.FromInt32(webhookTargetPort),
				Scheme: corev1.URISchemeHTTPS,
			},
			PeriodSeconds: 5,
		}
	}

	container := corev1.Container{
		Name:    resources.MachineControllerWebhookName,
		Image:   o.image,
		Command: []string{"/usr/local/bin/webhook"},
		Args:    args,
		Env:     o.env(WebhookComponent),
		Ports: []corev1.ContainerPort{
			{Name: "https", ContainerPort: webhookTargetPort, Protocol: corev1.ProtocolTCP},
		},
		LivenessProbe:  probe(),
		ReadinessProbe: probe(),
		VolumeMounts: []corev1.VolumeMount{
			tenantMount(),
			{Name: servingCertVolumeName, MountPath: servingCertDir, ReadOnly: true},
		},
		Resources: requests(),
	}
	container.LivenessProbe.InitialDelaySeconds = 5

	volumes := []corev1.Volume{
		{
			Name:   servingCertVolumeName,
			Secret: &corev1.SecretVolumeSource{SecretName: servingCertSecretName(o.clusterName)},
		},
	}

	return o.deployment(WebhookComponent, webhookPodLabels(o.clusterName), annotations, container, volumes)
}
