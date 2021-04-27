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

package tasks

import (
	"context"
	"io/fs"
	"path"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/ssh/sshiofs"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates"
	encryptionproviders "k8c.io/kubeone/pkg/templates/encryptionproviders"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	apiserverconfigv1 "k8s.io/apiserver/pkg/apis/config/v1"
	"k8s.io/client-go/util/retry"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	kyaml "sigs.k8s.io/yaml"
)

// download the configuration from leader
func fetchEncryptionProvidersFile(s *state.State) error {
	s.Logger.Infof("Downloading EncryptionProviders configuration file...")
	host, err := s.Cluster.Leader()
	if err != nil {
		return err
	}

	conn, err := s.Connector.Connect(host)
	if err != nil {
		return err
	}

	sshfs := sshiofs.New(conn)
	fileName := s.GetEncryptionProviderConfigName()

	config, err := fs.ReadFile(sshfs, path.Join("/etc/kubernetes/encryption-providers", fileName))
	if err != nil {
		return err
	}

	s.LiveCluster.Lock.Lock()
	s.LiveCluster.EncryptionConfiguration.Config = &apiserverconfigv1.EncryptionConfiguration{}
	err = kyaml.UnmarshalStrict(config, s.LiveCluster.EncryptionConfiguration.Config)
	s.LiveCluster.Lock.Unlock()

	return err
}

func uploadIdentityFirstEncryptionConfiguration(s *state.State) error {
	s.Logger.Infof("Uploading EncryptionProviders configuration file...")

	if s.LiveCluster.EncryptionConfiguration == nil ||
		s.LiveCluster.EncryptionConfiguration.Config == nil {
		return errors.New("failed to read live cluster encryption providers configuration")
	}

	if s.LiveCluster.EncryptionConfiguration.Custom {
		return errors.New("overriding custom encryption providers configuration is not supported")
	}

	oldConfig := s.LiveCluster.EncryptionConfiguration.Config.DeepCopy()

	if err := encryptionproviders.UpdateEncryptionConfigDecryptOnly(oldConfig); err != nil {
		return err
	}

	config, err := templates.KubernetesToYAML([]runtime.Object{oldConfig})
	if err != nil {
		return err
	}
	s.Configuration.AddFile("cfg/encryption-providers.yaml", config)
	return s.RunTaskOnControlPlane(pushEncryptionConfigurationOnNode, state.RunParallel)
}

func uploadEncryptionConfigurationWithNewKey(s *state.State) error {
	s.Logger.Infof("Uploading EncryptionProviders configuration file...")

	if s.LiveCluster.EncryptionConfiguration == nil ||
		s.LiveCluster.EncryptionConfiguration.Config == nil {
		return errors.New("failed to read live cluster encryption providers configuration")
	}

	if err := encryptionproviders.UpdateEncryptionConfigWithNewKey(s.LiveCluster.EncryptionConfiguration.Config); err != nil {
		return err
	}

	config, err := templates.KubernetesToYAML([]runtime.Object{s.LiveCluster.EncryptionConfiguration.Config})
	if err != nil {
		return err
	}

	s.Configuration.AddFile("cfg/encryption-providers.yaml", config)
	return s.RunTaskOnControlPlane(pushEncryptionConfigurationOnNode, state.RunParallel)
}

func uploadEncryptionConfigurationWithoutOldKey(s *state.State) error {
	s.Logger.Infof("Uploading EncryptionProviders configuration file...")

	if s.LiveCluster.EncryptionConfiguration == nil ||
		s.LiveCluster.EncryptionConfiguration.Config == nil {
		return errors.New("failed to read live cluster encryption providers configuration")
	}

	encryptionproviders.UpdateEncryptionConfigRemoveOldKey(s.LiveCluster.EncryptionConfiguration.Config)

	config, err := templates.KubernetesToYAML([]runtime.Object{s.LiveCluster.EncryptionConfiguration.Config})
	if err != nil {
		return err
	}
	s.Configuration.AddFile("cfg/encryption-providers.yaml", config)
	return s.RunTaskOnControlPlane(pushEncryptionConfigurationOnNode, state.RunParallel)
}

func pushEncryptionConfigurationOnNode(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	err := s.Configuration.UploadTo(conn, s.WorkDir)
	if err != nil {
		return err
	}
	cmd, err := scripts.SaveEncryptionProvidersConfig(s.WorkDir, s.GetEncryptionProviderConfigName())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return err
}

func rewriteClusterSecrets(s *state.State) error {
	s.Logger.Infof("Rewriting cluster secrets...")
	secrets := corev1.SecretList{}
	err := s.DynamicClient.List(context.Background(), &secrets, &dynclient.ListOptions{})
	if err != nil {
		return err
	}

	for _, secret := range secrets.Items {
		name := secret.Name
		namespace := secret.Namespace

		updateErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return rewriteSecret(s, name, namespace)
		})
		if updateErr != nil {
			return errors.WithStack(updateErr)
		}
	}
	return nil
}

func removeEncryptionProviderFile(s *state.State) error {
	s.Logger.Infof("Removing EncryptionProviders configuration file...")
	return s.RunTaskOnControlPlane(func(s *state.State, _ *kubeoneapi.HostConfig, _ ssh.Connection) error {
		cmd := scripts.DeleteEncryptionProvidersConfig(s.GetEncryptionProviderConfigName())

		_, _, err := s.Runner.RunRaw(cmd)
		return err
	}, state.RunParallel)
}

func rewriteSecret(s *state.State, name, namespace string) error {
	secret := corev1.Secret{}

	if err := s.DynamicClient.Get(s.Context, types.NamespacedName{Name: name, Namespace: namespace}, &secret); err != nil {
		return err
	}
	return s.DynamicClient.Update(s.Context, &secret, &dynclient.UpdateOptions{})
}
