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
	"sort"
	"strings"

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

	oldSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretNameLegacy,
			Namespace: SecretNamespace,
		},
	}
	if err := clientutil.DeleteIfExists(s.Context, s.DynamicClient, oldSecret); err != nil {
		return err
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
			return err
		}
	}

	if s.Cluster.MachineController.Deploy {
		s.Logger.Infoln("Creating machine-controller credentials secret...")

		providerCreds, err := ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, TypeMC)
		if err != nil {
			return err
		}

		mcSecret := credentialsSecret(SecretNameMC, providerCreds)
		if err = clientutil.CreateOrReplace(context.Background(), s.DynamicClient, mcSecret); err != nil {
			return err
		}
	}

	if s.Cluster.OperatingSystemManagerEnabled() {
		osmCreds, err := ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, TypeOSM)
		if err != nil {
			return err
		}

		osmSecret := credentialsSecret(SecretNameOSM, osmCreds)
		if err := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, osmSecret); err != nil {
			return err
		}
	}

	if s.Cluster.CloudProvider.CloudConfig != "" {
		cloudCfgSecret := cloudConfigSecret(s.Cluster.CloudProvider.CloudConfig)
		if err := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, cloudCfgSecret); err != nil {
			return err
		}
	}

	if (s.Cluster.CloudProvider.External && s.Cluster.CloudProvider.Vsphere == nil) ||
		s.Cluster.CloudProvider.GCE != nil {
		s.Logger.Infoln("Creating CCM credentials secret...")

		ccmCreds, err := ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, TypeCCM)
		if err != nil {
			return err
		}

		ccmSecret := credentialsSecret(SecretNameCCM, ccmCreds)
		if createErr := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, ccmSecret); createErr != nil {
			return createErr
		}
	} else if s.Cluster.CloudProvider.Vsphere != nil {
		s.Logger.Infoln("Creating vSphere CCM credentials secret...")

		ccmCreds, err := ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, TypeCCM)
		if err != nil {
			return err
		}

		vsecret := vsphereSecret(ccmCreds)
		if err := clientutil.CreateOrReplace(context.Background(), s.DynamicClient, vsecret); err != nil {
			return err
		}
	}

	return nil
}

func EnvVarBindings(secretName string, creds map[string]string) []corev1.EnvVar {
	var (
		envVars   []corev1.EnvVar
		credsKeys []string
	)

	for k := range creds {
		credsKeys = append(credsKeys, k)
	}

	sort.Slice(credsKeys, func(i, j int) bool {
		return credsKeys[i] < credsKeys[j]
	})

	for _, key := range credsKeys {
		envVars = append(envVars, corev1.EnvVar{
			Name: key,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: key,
				},
			},
		})
	}

	return envVars
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
