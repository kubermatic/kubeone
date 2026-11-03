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

// Package ccm deploys the KubeVirt cloud-controller-manager in the infra
// cluster, following the upstream "isolated" topology: the CCM accesses the
// infra cluster with its in-cluster ServiceAccount and the tenant cluster with
// a kubeconfig built from a tenant ServiceAccount token.
package ccm

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	// TenantServiceAccountName is the tenant cluster ServiceAccount of the CCM.
	TenantServiceAccountName = "cloud-controller-manager"
	// TenantTokenSecretName is the token Secret of TenantServiceAccountName.
	TenantTokenSecretName = "cloud-controller-manager-token"
	// TenantNamespace is the tenant cluster namespace of TenantServiceAccountName.
	TenantNamespace = "kube-system"

	componentName = "kubevirt-ccm"

	cloudConfigKey        = "cloud-config"
	cloudConfigDir        = "/etc/cloud"
	cloudConfigVolumeName = "cloud-config"
	tenantKubeconfigDir   = "/var/run/secrets/tenantcluster"
	tenantVolumeName      = "tenantcluster"
	tenantKubeconfig      = tenantKubeconfigDir + "/" + kubevirtinfra.KubeconfigSecretKey

	cloudConfigHashAnnotation = "kubeone.k8c.io/cloud-config-hash"
)

// ImageGetter resolves an image reference, applying the overwrite registry.
type ImageGetter func(images.Resource, ...images.GetOpt) string

// Component returns the infra cluster identity of the CCM. The Role follows
// the upstream "isolated" overlay. Services are only granted when the CCM
// manages load balancers. Nodes are only granted when the CCM discovers zones
// and regions, which it reads from the labels of the infra Node running the
// VMI. The CCM uses in-cluster configuration, so no infra kubeconfig is
// generated.
func Component(spec *kubeoneapi.KubevirtSpec) kubevirtinfra.Component {
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{"kubevirt.io"},
			Resources: []string{"virtualmachines"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"kubevirt.io"},
			Resources: []string{"virtualmachineinstances"},
			Verbs:     []string{"get", "list", "watch", "update"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}

	if spec.LoadBalancerEnabled {
		rules = append(rules, rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"services"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		})
	}

	var clusterRules []rbacv1.PolicyRule
	if spec.ZoneAndRegionEnabled {
		clusterRules = append(clusterRules, rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"nodes"},
			Verbs:     []string{"get"},
		})
	}

	return kubevirtinfra.Component{
		Name:         componentName,
		Rules:        rules,
		ClusterRules: clusterRules,
	}
}

// cloudConfig is the subset of the KubeVirt CCM cloud-config KubeOne sets.
// "kubeconfig" is intentionally empty: the CCM then uses in-cluster
// configuration for the infra cluster.
type cloudConfig struct {
	Namespace    string            `json:"namespace"`
	InfraLabels  map[string]string `json:"infraLabels"`
	InstancesV2  instancesV2       `json:"instancesV2"`
	LoadBalancer loadBalancer      `json:"loadBalancer"`
}

type instancesV2 struct {
	Enabled              bool `json:"enabled"`
	ZoneAndRegionEnabled bool `json:"zoneAndRegionEnabled"`
}

type loadBalancer struct {
	Enabled bool `json:"enabled"`
}

// CloudConfig renders the CCM cloud-config.
func CloudConfig(clusterName string, spec *kubeoneapi.KubevirtSpec) ([]byte, error) {
	cfg := cloudConfig{
		Namespace: spec.InfraNamespace,
		// Set on the LoadBalancer Services the CCM creates in the infra
		// cluster. They intentionally lack kubevirtinfra.LabelComponent, so
		// kubevirtinfra.Cleanup does not delete them; they belong to tenant
		// Services.
		InfraLabels: map[string]string{kubevirtinfra.LabelClusterName: kubevirtinfra.ClusterLabelValue(clusterName)},
		InstancesV2: instancesV2{
			Enabled:              true,
			ZoneAndRegionEnabled: spec.ZoneAndRegionEnabled,
		},
		LoadBalancer: loadBalancer{Enabled: spec.LoadBalancerEnabled},
	}

	out, err := yaml.Marshal(cfg)

	return out, fail.Runtime(err, "marshalling kubevirt CCM cloud-config")
}

// ConfigMap returns the ConfigMap holding the CCM cloud-config.
func ConfigMap(clusterName, namespace string, cloudCfg []byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		Name:      configMapName(clusterName),
		Namespace: namespace,
		Labels:    kubevirtinfra.Labels(clusterName, componentName),
		Data: map[string]string{
			cloudConfigKey: string(cloudCfg),
		},
	}
}

func configMapName(clusterName string) string {
	return kubevirtinfra.ObjectName(clusterName, componentName) + "-cloud-config"
}

// Deployment returns the CCM Deployment running in the infra cluster.
func Deployment(clusterName, namespace string, cloudCfg []byte, image ImageGetter) *appsv1.Deployment {
	comp := kubevirtinfra.Component{Name: componentName}
	labels := kubevirtinfra.Labels(clusterName, componentName)
	podLabels := maps.Clone(labels)
	podLabels["app.kubernetes.io/name"] = componentName

	cfgHash := sha256.Sum256(cloudCfg)

	return &appsv1.Deployment{
		Name:      kubevirtinfra.ObjectName(clusterName, componentName),
		Namespace: namespace,
		Labels:    labels,
		Spec: appsv1.DeploymentSpec{
			Replicas: new(int32(1)),
			Selector: &metav1.LabelSelector{MatchLabels: podLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
					Annotations: map[string]string{
						cloudConfigHashAnnotation: hex.EncodeToString(cfgHash[:]),
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: comp.ServiceAccountName(clusterName),
					Containers: []corev1.Container{
						{
							Name:    "kubevirt-cloud-controller-manager",
							Image:   image(images.KubeVirtCCM),
							Command: []string{"/bin/kubevirt-cloud-controller-manager"},
							Args: []string{
								"--cloud-provider=kubevirt",
								"--cloud-config=" + cloudConfigDir + "/" + cloudConfigKey,
								"--cluster-name=" + clusterName,
								"--kubeconfig=" + tenantKubeconfig,
								"--authentication-kubeconfig=" + tenantKubeconfig,
								"--authorization-kubeconfig=" + tenantKubeconfig,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: new(false),
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: cloudConfigVolumeName, MountPath: cloudConfigDir, ReadOnly: true},
								{Name: tenantVolumeName, MountPath: tenantKubeconfigDir, ReadOnly: true},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: cloudConfigVolumeName,
							ConfigMap: &corev1.ConfigMapVolumeSource{
								Name: configMapName(clusterName),
							},
						},
						{
							Name: tenantVolumeName,
							Secret: &corev1.SecretVolumeSource{
								SecretName: comp.TenantKubeconfigSecretName(clusterName),
							},
						},
					},
				},
			},
		},
	}
}

// Ensure reconciles the CCM in the infra cluster. It runs after the
// ccm-kubevirt addon created the tenant ServiceAccount and its token.
func Ensure(s *state.State) error {
	if s.DynamicClient == nil {
		return fail.NoKubeClient()
	}

	ctx := s.Context
	spec := s.Cluster.CloudProvider.Kubevirt
	comp := Component(spec)

	infraClient, namespace, err := kubevirtinfra.Client(s)
	if err != nil {
		return err
	}

	if err = comp.EnsureIdentity(ctx, infraClient, s.Cluster.Name, namespace); err != nil {
		return err
	}

	// EnsureIdentity doesn't remove the cluster-wide objects once
	// ClusterRules is empty, so they are deleted here when zone and region
	// discovery is disabled.
	if len(comp.ClusterRules) == 0 {
		for _, obj := range []ctrlrt.Object{
			comp.ClusterRoleBinding(s.Cluster.Name, namespace),
			comp.ClusterRole(s.Cluster.Name, namespace),
		} {
			if err = clientutil.DeleteIfExists(ctx, infraClient, obj); err != nil {
				return err
			}
		}
	}

	err = comp.EnsureTenantKubeconfig(
		ctx,
		s.DynamicClient,
		infraClient,
		s.Cluster.Name,
		namespace,
		kubevirtinfra.TenantServer(s.Cluster.APIEndpoint),
		ctrlrt.ObjectKey{Namespace: TenantNamespace, Name: TenantTokenSecretName},
	)
	if err != nil {
		return err
	}

	cloudCfg, err := CloudConfig(s.Cluster.Name, spec)
	if err != nil {
		return err
	}

	if err = clientutil.CreateOrReplace(ctx, infraClient, ConfigMap(s.Cluster.Name, namespace, cloudCfg)); err != nil {
		return err
	}

	return clientutil.CreateOrReplace(ctx, infraClient, Deployment(s.Cluster.Name, namespace, cloudCfg, s.Images.Get))
}
