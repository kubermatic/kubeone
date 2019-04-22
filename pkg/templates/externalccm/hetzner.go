package externalccm

import (
	"context"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	hetznerCCMVersion     = "v1.3.0"
	hetznerSAName         = "cloud-controller-manager"
	hetznerDeploymentName = "hcloud-cloud-controller-manager"
)

func ensureHetzner(ctx *util.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	bgctx := context.Background()

	sa := hetznerServiceAccount()
	if err := simpleCreateOrUpdate(bgctx, ctx.DynamicClient, sa); err != nil {
		return errors.Wrap(err, "failed to ensure hetzner CCM ServiceAccount")
	}

	crb := hetznerClusterRoleBinding()
	if err := simpleCreateOrUpdate(bgctx, ctx.DynamicClient, crb); err != nil {
		return errors.Wrap(err, "failed to ensure hetzner CCM ClusterRoleBinding")
	}

	dep := hetznerDeployment()
	_, err := controllerutil.CreateOrUpdate(bgctx, ctx.DynamicClient, dep, func(runtime.Object) error {
		if dep.ObjectMeta.CreationTimestamp.IsZero() {
			// let it create deployment
			return nil
		}

		if len(dep.Spec.Template.Spec.Containers) != 1 {
			return errors.New("unable to choose a CCM container, as number of containers > 1")
		}

		want, err := semver.NewConstraint("<= " + hetznerCCMVersion)
		if err != nil {
			return errors.Wrap(err, "failed to parse hetzner CCM version constraint")
		}

		imageSpec := strings.SplitN(dep.Spec.Template.Spec.Containers[0].Image, ":", 2)
		if len(imageSpec) != 2 {
			return errors.New("unable to greb hetzner CCM image version")
		}

		existing, err := semver.NewVersion(imageSpec[1])
		if err != nil {
			return errors.Wrap(err, "failed to parse deployed hetzner CCM version")
		}

		if !want.Check(existing) {
			return errors.New("newer version deployed, skipping")
		}

		// let it update deployment
		return nil
	})

	if err != nil {
		ctx.Logger.Warnf("unable to ensure hetzner CCM Deployment: %v, skipping", err)
	}

	return nil
}

func hetznerServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hetznerSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func hetznerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "cluster-admin",
			Kind:     "ClusterRole",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      hetznerSAName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func hetznerDeployment() *appsv1.Deployment {
	var (
		replicas  int32 = 1
		revisions int32 = 2
	)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hetznerDeploymentName,
			Namespace: metav1.NamespaceSystem,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &revisions,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "hcloud-cloud-controller-manager",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
					Labels: map[string]string{
						"app": "hcloud-cloud-controller-manager",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: hetznerSAName,
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
							Name:  "hcloud-cloud-controller-manager",
							Image: "hetznercloud/hcloud-cloud-controller-manager:" + hetznerCCMVersion,
							Command: []string{
								"/bin/hcloud-cloud-controller-manager",
								"--cloud-provider=hcloud",
								"--leader-elect=false",
								"--allow-untagged-cloud",
							},
							Env: []corev1.EnvVar{
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name: "HCLOUD_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: machinecontroller.MachineControllerCredentialsSecretName,
											},
											Key: config.HetznerTokenKey,
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("50Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}
