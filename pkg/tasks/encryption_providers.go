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
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/scripts"
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

	conn, err := s.Executor.Open(host)
	if err != nil {
		return err
	}

	virtfs := executorfs.New(conn)
	fileName := s.GetEncryptionProviderConfigName()

	config, err := fs.ReadFile(virtfs, path.Join("/etc/kubernetes/encryption-providers", fileName))
	if err != nil {
		return err
	}

	s.LiveCluster.Lock.Lock()
	s.LiveCluster.EncryptionConfiguration.Config = &apiserverconfigv1.EncryptionConfiguration{}
	err = kyaml.UnmarshalStrict(config, s.LiveCluster.EncryptionConfiguration.Config)
	s.LiveCluster.Lock.Unlock()

	return fail.Runtime(err, "unmarshalling EncryptionConfiguration")
}

func uploadIdentityFirstEncryptionConfiguration(s *state.State) error {
	s.Logger.Infof("Uploading EncryptionProviders configuration file...")

	if ec := s.LiveCluster.EncryptionConfiguration; ec == nil || ec.Config == nil {
		return fail.RuntimeError{
			Op:  "validating live cluster encryption providers configuration",
			Err: errors.New("failed to read"),
		}
	}

	if s.LiveCluster.EncryptionConfiguration.Custom {
		return fail.RuntimeError{
			Op:  "validating custom encryption providers configuration",
			Err: errors.New("overriding is not supported"),
		}
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

	if ec := s.LiveCluster.EncryptionConfiguration; ec == nil || ec.Config == nil {
		return fail.RuntimeError{
			Op:  "validating live cluster encryption providers configuration",
			Err: errors.New("failed to read"),
		}
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

	if ec := s.LiveCluster.EncryptionConfiguration; ec == nil || ec.Config == nil {
		return fail.RuntimeError{
			Op:  "validating live cluster encryption providers configuration",
			Err: errors.New("failed to read"),
		}
	}

	encryptionproviders.UpdateEncryptionConfigRemoveOldKey(s.LiveCluster.EncryptionConfiguration.Config)

	config, err := templates.KubernetesToYAML([]runtime.Object{s.LiveCluster.EncryptionConfiguration.Config})
	if err != nil {
		return err
	}
	s.Configuration.AddFile("cfg/encryption-providers.yaml", config)

	return s.RunTaskOnControlPlane(pushEncryptionConfigurationOnNode, state.RunParallel)
}

func pushEncryptionConfigurationOnNode(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
	err := s.Configuration.UploadTo(conn, s.WorkDir)
	if err != nil {
		return err
	}
	cmd, err := scripts.SaveEncryptionProvidersConfig(s.WorkDir, s.GetEncryptionProviderConfigName())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return fail.SSH(err, "saving encryption providers config")
}

func rewriteClusterSecrets(s *state.State) error {
	s.Logger.Infof("Rewriting cluster secrets...")
	secrets := corev1.SecretList{}
	err := s.DynamicClient.List(context.Background(), &secrets, &dynclient.ListOptions{})
	if err != nil {
		return fail.KubeClient(err, "getting %T", secrets)
	}

	for _, secret := range secrets.Items {
		name := secret.Name
		namespace := secret.Namespace
		key := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}

		updateErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return rewriteSecret(s, name, namespace)
		})
		if updateErr != nil {
			return fail.KubeClient(updateErr, "updating %T %s", secret, key)
		}
	}

	return nil
}

func removeEncryptionProviderFile(s *state.State) error {
	s.Logger.Infof("Removing EncryptionProviders configuration file...")

	return s.RunTaskOnControlPlane(func(s *state.State, _ *kubeoneapi.HostConfig, _ executor.Interface) error {
		cmd := scripts.DeleteEncryptionProvidersConfig(s.GetEncryptionProviderConfigName())

		_, _, err := s.Runner.RunRaw(cmd)

		return fail.SSH(err, "deleting encryption providers config")
	}, state.RunParallel)
}

func rewriteSecret(s *state.State, name, namespace string) error {
	secret := corev1.Secret{}

	if err := s.DynamicClient.Get(s.Context, types.NamespacedName{Name: name, Namespace: namespace}, &secret); err != nil {
		return err
	}

	return s.DynamicClient.Update(s.Context, &secret)
}
