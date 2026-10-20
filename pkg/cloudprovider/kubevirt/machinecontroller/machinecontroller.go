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

// Package machinecontroller deploys machine-controller and its webhook in the
// KubeVirt infra cluster. Both access the infra cluster with a kubeconfig of
// their infra ServiceAccount (the KubeVirt provider has no in-cluster
// fallback) and the tenant cluster with a kubeconfig of their tenant
// ServiceAccount. The tenant API server reaches the webhook through an infra
// Service.
package machinecontroller

import (
	"context"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/certificate/cabundle"
	"k8c.io/kubeone/pkg/clientutil"
	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// TenantNamespace is the tenant cluster namespace of the tenant ServiceAccounts.
	TenantNamespace = resources.MachineControllerNameSpace

	// ControllerTenantTokenSecretName is the token Secret of the machine-controller tenant ServiceAccount.
	ControllerTenantTokenSecretName = resources.MachineControllerName + "-token"
	// WebhookTenantTokenSecretName is the token Secret of the machine-controller-webhook tenant ServiceAccount.
	WebhookTenantTokenSecretName = resources.MachineControllerWebhookName + "-token"

	readyWaitInterval = 5 * time.Second
	readyWaitTimeout  = 5 * time.Minute
)

// clusterReadRules are cluster-scoped infra resources the KubeVirt provider
// reads: the StorageClass of the primary disk (topologies, on every machine
// create and in the webhook defaulting), its CDI StorageProfile (default
// access mode) and cluster instancetypes and preferences.
var clusterReadRules = []rbacv1.PolicyRule{
	{
		APIGroups: []string{"storage.k8s.io"},
		Resources: []string{"storageclasses"},
		Verbs:     []string{"get", "list", "watch"},
	},
	{
		APIGroups: []string{"cdi.kubevirt.io"},
		Resources: []string{"storageprofiles"},
		Verbs:     []string{"get", "list", "watch"},
	},
	{
		APIGroups: []string{"instancetype.kubevirt.io"},
		Resources: []string{"virtualmachineclusterinstancetypes", "virtualmachineclusterpreferences"},
		Verbs:     []string{"get", "list", "watch"},
	},
}

var instancetypeReadRule = rbacv1.PolicyRule{
	APIGroups: []string{"instancetype.kubevirt.io"},
	Resources: []string{"virtualmachineinstancetypes", "virtualmachinepreferences"},
	Verbs:     []string{"get", "list", "watch"},
}

// ControllerComponent is the infra identity of machine-controller.
var ControllerComponent = kubevirtinfra.Component{
	Name: resources.MachineControllerName,
	Rules: []rbacv1.PolicyRule{
		{
			APIGroups: []string{"kubevirt.io"},
			Resources: []string{"virtualmachines"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		},
		{
			APIGroups: []string{"kubevirt.io"},
			Resources: []string{"virtualmachineinstances"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"cdi.kubevirt.io"},
			Resources: []string{"datavolumes"},
			Verbs:     []string{"get", "list", "watch", "create", "delete"},
		},
		{
			// Cloning a PVC in the same namespace as the VM disk source
			APIGroups: []string{"cdi.kubevirt.io"},
			Resources: []string{"datavolumes/source"},
			Verbs:     []string{"create"},
		},
		{
			// cloud-init user data Secrets of the VMs, extra headers Secrets
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
		},
		instancetypeReadRule,
	},
	ClusterRules:    clusterReadRules,
	NeedsKubeconfig: true,
}

// WebhookComponent is the infra identity of the machine-controller webhook,
// which validates and defaults Machines against the infra cluster.
var WebhookComponent = kubevirtinfra.Component{
	Name: resources.MachineControllerWebhookName,
	Rules: []rbacv1.PolicyRule{
		{
			APIGroups: []string{"kubevirt.io"},
			Resources: []string{"virtualmachines", "virtualmachineinstances"},
			Verbs:     []string{"get", "list"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get"},
		},
		instancetypeReadRule,
	},
	ClusterRules:    clusterReadRules,
	NeedsKubeconfig: true,
}

// caBundleConfigMapName returns the name of the infra ConfigMap holding the
// cluster's custom CA bundle.
func caBundleConfigMapName(clusterName string) string {
	return kubevirtinfra.ObjectName(clusterName, resources.MachineControllerName) + "-ca-bundle"
}

// Ensure reconciles machine-controller and its webhook in the infra cluster,
// and the tenant MutatingWebhookConfiguration pointing at the webhook. It runs
// after the machinecontroller addon created the tenant ServiceAccounts and
// their tokens.
func Ensure(s *state.State) error {
	if s.DynamicClient == nil {
		return fail.NoKubeClient()
	}

	ctx := s.Context
	clusterName := s.Cluster.Name
	spec := s.Cluster.CloudProvider.Kubevirt

	infraClient, namespace, err := kubevirtinfra.Client(s)
	if err != nil {
		return err
	}

	tenantServer := kubevirtinfra.TenantServer(s.Cluster.APIEndpoint)

	identities := []struct {
		comp     kubevirtinfra.Component
		tokenKey ctrlrt.ObjectKey
	}{
		{ControllerComponent, ctrlrt.ObjectKey{Namespace: TenantNamespace, Name: ControllerTenantTokenSecretName}},
		{WebhookComponent, ctrlrt.ObjectKey{Namespace: TenantNamespace, Name: WebhookTenantTokenSecretName}},
	}

	for _, id := range identities {
		if err = id.comp.EnsureIdentity(ctx, infraClient, clusterName, namespace); err != nil {
			return err
		}

		if err = id.comp.EnsureTenantKubeconfig(ctx, s.DynamicClient, infraClient, clusterName, namespace, tenantServer, id.tokenKey); err != nil {
			return err
		}
	}

	caBundle := s.Cluster.CertificateAuthority.Bundle
	if caBundle != "" {
		cm := cabundle.ConfigMap(caBundle, namespace)
		cm.Name = caBundleConfigMapName(clusterName)
		cm.Labels = kubevirtinfra.Labels(clusterName, ControllerComponent.Name)

		if err = clientutil.CreateOrReplace(ctx, infraClient, cm); err != nil {
			return err
		}
	}

	svc := WebhookService(clusterName, namespace, spec)
	if err = clientutil.CreateOrUpdate(ctx, infraClient, svc); err != nil {
		return err
	}

	host, err := webhookHost(ctx, infraClient, ctrlrt.ObjectKeyFromObject(svc))
	if err != nil {
		return err
	}

	kubeCAKey, kubeCACert, err := certificate.CAKeyPair(s.Configuration)
	if err != nil {
		return err
	}

	servingCert, err := ensureServingCert(ctx, infraClient, clusterName, namespace, host, kubeCAKey, kubeCACert)
	if err != nil {
		return err
	}

	opts := deploymentOptions{
		clusterName:           clusterName,
		namespace:             namespace,
		image:                 s.Images.Get(imageMachineController),
		externalCloudProvider: s.Cluster.CloudProvider.External,
		caBundle:              caBundle,
		joinClusterTimeout:    joinClusterTimeout(s.Cluster),
		servingCertHash:       servingCert.hash,
	}

	for _, d := range []*appsv1.Deployment{ControllerDeployment(opts), WebhookDeployment(opts)} {
		if err = clientutil.CreateOrReplace(ctx, infraClient, d); err != nil {
			return err
		}
	}

	return clientutil.CreateOrReplace(ctx, s.DynamicClient, MutatingWebhookConfiguration(host, servingCert.caPEM))
}

// joinClusterTimeout returns the JOIN_TIMEOUT param of the machinecontroller
// addon, as the tenant addon manifest used to.
func joinClusterTimeout(cluster *kubeoneapi.KubeOneCluster) string {
	for _, addon := range cluster.Addons.DeclaredAddonsOnly() {
		if addon.Name == resources.AddonMachineController && addon.Params["JOIN_TIMEOUT"] != "" {
			return addon.Params["JOIN_TIMEOUT"]
		}
	}

	return "15m"
}

// WaitReady waits for the machine-controller and webhook Deployments in the
// infra cluster to become ready.
func WaitReady(s *state.State) error {
	infraClient, namespace, err := kubevirtinfra.Client(s)
	if err != nil {
		return err
	}

	for _, comp := range []kubevirtinfra.Component{WebhookComponent, ControllerComponent} {
		key := ctrlrt.ObjectKey{Namespace: namespace, Name: kubevirtinfra.ObjectName(s.Cluster.Name, comp.Name)}

		err = wait.PollUntilContextTimeout(s.Context, readyWaitInterval, readyWaitTimeout, false, func(ctx context.Context) (bool, error) {
			d := &appsv1.Deployment{}
			if getErr := infraClient.Get(ctx, key, d); getErr != nil {
				return false, ctrlrt.IgnoreNotFound(getErr)
			}

			return deploymentReady(d), nil
		})
		if err != nil {
			return fail.KubeClient(err, "waiting for %s in the infra cluster to become ready", key)
		}
	}

	return nil
}

func deploymentReady(d *appsv1.Deployment) bool {
	want := int32(1)
	if d.Spec.Replicas != nil {
		want = *d.Spec.Replicas
	}

	return d.Status.ObservedGeneration >= d.Generation &&
		d.Status.UpdatedReplicas == want &&
		d.Status.ReadyReplicas == want &&
		d.Status.Replicas == want
}

// tenantKubeconfigVolume mounts the component's tenant kubeconfig.
func tenantKubeconfigVolume(clusterName string, comp kubevirtinfra.Component) corev1.Volume {
	return corev1.Volume{
		Name:   tenantVolumeName,
		Secret: &corev1.SecretVolumeSource{SecretName: comp.TenantKubeconfigSecretName(clusterName)},
	}
}

// infraKubeconfigEnv passes the component's infra kubeconfig the way the
// KubeVirt provider expects it.
func infraKubeconfigEnv(clusterName string, comp kubevirtinfra.Component) corev1.EnvVar {
	return corev1.EnvVar{
		Name: "KUBEVIRT_KUBECONFIG",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Name: comp.KubeconfigSecretName(clusterName),
				Key:  kubevirtinfra.KubeconfigSecretKey,
			},
		},
	}
}
