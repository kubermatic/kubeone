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

package credentials

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SecretName is name of the secret which contains the cloud provider credentials
	SecretName = "cloud-provider-credentials"
	// SecretNamespace is namespace of the credentials secret
	SecretNamespace = "kube-system"
	// VsphereSecretName is name of the secret which contains the vSphere credentials
	// used by the cloud provider integrations (CCM, CSI)
	VsphereSecretName = "vsphere-ccm-credentials" //nolint:gosec
	// VsphereSecretNamespace is namespace of the vSphere credentials secret
	VsphereSecretNamespace = "kube-system"
	// CloudConfigSecretName is name of the secret which contains the cloud-config file
	CloudConfigSecretName = "cloud-config" //nolint:gosec
	// CloudConfigSecretNamespace is namespace of the cloud-config secret
	CloudConfigSecretNamespace = "kube-system"
)

// Ensure creates/updates the credentials secret
func Ensure(s *state.State) error {
	if s.Cluster.CloudProvider.None != nil {
		s.Logger.Info("Skipping creating credentials secret because cloud provider is none.")
		return nil
	}
	if !s.Cluster.MachineController.Deploy && !s.Cluster.CloudProvider.External {
		s.Logger.Info("Skipping creating credentials secret because both machine-controller and external CCM are disabled.")
		return nil
	}

	s.Logger.Infoln("Creating credentials secret...")

	creds, err := ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "unable to fetch cloud provider credentials")
	}

	secret := credentialsSecret(creds)
	if err := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, secret); err != nil {
		return errors.Wrap(err, "failed to ensure credentials secret")
	}

	if s.Cluster.CloudProvider.CloudConfig != "" {
		cloudCfgSecret := cloudConfigSecret(s.Cluster.CloudProvider.CloudConfig)
		if err := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, cloudCfgSecret); err != nil {
			return errors.Wrap(err, "failed to ensure cloud-config secret")
		}
	}

	if s.Cluster.CloudProvider.Vsphere != nil {
		vsecret := vsphereSecret(creds)
		if err := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, vsecret); err != nil {
			return errors.Wrap(err, "failed to ensure vsphere credentials secret")
		}
	}

	return nil
}

func EnvVarBindings(cloudProviderSpec kubeoneapi.CloudProviderSpec, credentialsFilePath string) ([]corev1.EnvVar, error) {
	creds, err := ProviderCredentials(cloudProviderSpec, credentialsFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch cloud provider credentials")
	}

	env := make([]corev1.EnvVar, 0)

	for k := range creds {
		env = append(env, corev1.EnvVar{
			Name: k,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: SecretName,
					},
					Key: k,
				},
			},
		})
	}

	return env, nil
}

func credentialsSecret(credentials map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName,
			Namespace: SecretNamespace,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: credentials,
	}
}

func vsphereSecret(credentials map[string]string) *corev1.Secret {
	vscreds := map[string]string{}

	vcenterPrefix := strings.ReplaceAll(credentials[VSphereAddressMC], "https://", "")
	// Save credentials in Secret and configure vSphere cloud controller
	// manager to read it, in replace of storing those in /etc/kubernates/cloud-config
	// see more: https://vmware.github.io/vsphere-storage-for-kubernetes/documentation/k8s-secret.html
	vscreds[fmt.Sprintf("%s.username", vcenterPrefix)] = credentials[VSphereUsernameMC]
	vscreds[fmt.Sprintf("%s.password", vcenterPrefix)] = credentials[VSpherePassword]

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VsphereSecretName,
			Namespace: VsphereSecretNamespace,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: vscreds,
	}
}

func cloudConfigSecret(cloudConfig string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CloudConfigSecretName,
			Namespace: CloudConfigSecretNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"cloud-config": cloudConfig,
		},
	}
}
