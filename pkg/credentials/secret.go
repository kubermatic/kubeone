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
	// SecretNameCCM is name of the secret which contains the cloud provider credentials for CCM
	SecretNameCCM = "kubeone-ccm-credentials" //nolint:gosec
	// SecretNameMC is name of the secret which contains the cloud provider credentials for machine-controller
	SecretNameMC = "kubeone-machine-controller-credentials"
	// SecretNameOSM is name of the secret which contains the cloud provider credentials for operating-system-manager
	SecretNameOSM = "kubeone-operating-system-manager-credentials"
	// SecretNameLegacy is name of the secret created by earlier KubeOne versions, but not used anymore
	// This secret will be removed for all clusters when running kubeone apply the next time
	SecretNameLegacy = "cloud-provider-credentials"
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
	if !s.Cluster.MachineController.Deploy && !s.Cluster.CloudProvider.External && !s.Cluster.OperatingSystemManagerEnabled() {
		s.Logger.Info("Skipping creating credentials secret because both machine-controller and external CCM are disabled.")

		return nil
	}

	oldSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretNameLegacy,
			Namespace: SecretNamespace,
		},
	}
	if err := clientutil.DeleteIfExists(s.Context, s.DynamicClient, oldSecret); err != nil {
		return errors.Wrap(err, "unable to remove cloud-provider-credentials secret")
	}

	// Ensure that we remove credentials secret for OSM if it's queued for deletion
	if s.Cluster.OperatingSystemManagerQueuedForDeletion() {
		osmSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      SecretNameOSM,
				Namespace: SecretNamespace,
			},
		}
		if err := clientutil.DeleteIfExists(s.Context, s.DynamicClient, osmSecret); err != nil {
			return errors.Wrapf(err, "unable to remove %v secret", SecretNameOSM)
		}
	}

	s.Logger.Infoln("Creating machine-controller credentials secret...")

	providerCreds, err := ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, TypeMC)
	if err != nil {
		return errors.Wrap(err, "unable to fetch cloud provider credentials")
	}

	mcSecret := credentialsSecret(SecretNameMC, providerCreds)
	if createErr := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, mcSecret); createErr != nil {
		return errors.Wrap(createErr, "failed to ensure credentials secret for machine-controller")
	}

	if s.Cluster.OperatingSystemManagerEnabled() {
		osmSecret := credentialsSecret(SecretNameOSM, providerCreds)
		if createErr := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, osmSecret); createErr != nil {
			return errors.Wrap(createErr, "failed to ensure credentials secret for operating-system-manager")
		}
	}

	if s.Cluster.CloudProvider.CloudConfig != "" {
		cloudCfgSecret := cloudConfigSecret(s.Cluster.CloudProvider.CloudConfig)
		if err := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, cloudCfgSecret); err != nil {
			return errors.Wrap(err, "failed to ensure cloud-config secret")
		}
	}

	if s.Cluster.CloudProvider.External && s.Cluster.CloudProvider.Vsphere == nil {
		s.Logger.Infoln("Creating CCM credentials secret...")

		ccmCreds, err := ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, TypeCCM)
		if err != nil {
			return errors.Wrap(err, "unable to fetch cloud provider credentials")
		}

		ccmSecret := credentialsSecret(SecretNameCCM, ccmCreds)
		if createErr := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, ccmSecret); createErr != nil {
			return errors.Wrap(createErr, "failed to ensure credentials secret")
		}
	} else if s.Cluster.CloudProvider.Vsphere != nil {
		s.Logger.Infoln("Creating vSphere CCM credentials secret...")

		ccmCreds, err := ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, TypeCCM)
		if err != nil {
			return errors.Wrap(err, "unable to fetch cloud provider credentials")
		}

		vsecret := vsphereSecret(ccmCreds)
		if err := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, vsecret); err != nil {
			return errors.Wrap(err, "failed to ensure vsphere credentials secret")
		}
	}

	return nil
}

func EnvVarBindings(cloudProviderSpec kubeoneapi.CloudProviderSpec, credentialsFilePath, secretName string, credentialsType Type) ([]corev1.EnvVar, error) {
	creds, err := ProviderCredentials(cloudProviderSpec, credentialsFilePath, credentialsType)
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
						Name: secretName,
					},
					Key: k,
				},
			},
		})
	}

	return env, nil
}

func credentialsSecret(secretName string, credentials map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
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
