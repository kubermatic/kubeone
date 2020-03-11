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

package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/kubermatic/kubeone/test/e2e/testutil"
)

const configurationTpl = `
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
versions:
  kubernetes: {{ .KUBERNETES_VERSION }}
cloudProvider:
  name: {{ .CLOUD_PROVIDER_NAME }}
  external: {{ .CLOUD_PROVIDER_EXTERNAL }}
{{ with .CLOUD_CONFIG }}
  cloudConfig: |
{{ . }}
{{ end }}
{{ if .CLUSTER_NETWORK_POD }}
clusterNetwork:
  podSubnet: "{{ .CLUSTER_NETWORK_POD }}"
  serviceSubnet: "{{ .CLUSTER_NETWORK_SERVICE }}"
{{ end }}
`

// Kubeone is wrapper around KubeOne CLI
type Kubeone struct {
	// Dir is a temporary directory for storing test files (e.g. tf.json)
	Dir string
	// ConfigurationFilePath is a path to the KubeOneCluster manifest
	ConfigurationFilePath string
}

// NewKubeone creates and initializes the Kubeone structure
func NewKubeone(kubeoneDir, configurationFilePath string) *Kubeone {
	return &Kubeone{
		Dir:                   kubeoneDir,
		ConfigurationFilePath: configurationFilePath,
	}
}

// CreateConfig creates a KubeOneCluster manifest
func (k1 *Kubeone) CreateConfig(
	kubernetesVersion string,
	providerName string,
	providerExternal bool,
	clusterNetworkPod string,
	clusterNetworkService string,
	credentialsFile string,
) error {
	data := map[string]interface{}{
		"KUBERNETES_VERSION":      kubernetesVersion,
		"CLOUD_PROVIDER_NAME":     providerName,
		"CLOUD_PROVIDER_EXTERNAL": providerExternal,
		"CLUSTER_NETWORK_POD":     clusterNetworkPod,
		"CLUSTER_NETWORK_SERVICE": clusterNetworkService,
		"CLOUD_CONFIG":            "",
	}

	if credentialsFile != "" {
		ymlbuf, err := ioutil.ReadFile(credentialsFile)
		if err != nil {
			return errors.Wrap(err, "unable to read credentials file")
		}

		credentials := make(map[string]string)
		if err = yaml.Unmarshal(ymlbuf, &credentials); err != nil {
			return errors.Wrap(err, "unable to unmarshal credentials file from yaml")
		}

		data["CLOUD_CONFIG"] = credentials["cloudConfig"]
	}

	tpl, err := template.New("base").Parse(configurationTpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse KubeOne configuration template")
	}

	var buf bytes.Buffer

	if tplErr := tpl.Execute(&buf, data); tplErr != nil {
		return errors.Wrap(tplErr, "failed to render KubeOne configuration template")
	}

	err = ioutil.WriteFile(k1.ConfigurationFilePath, buf.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write KubeOne configuration manifest")
	}

	return nil
}

// Install runs 'kubeone install' command to provision the cluster
func (k1 *Kubeone) Install(tfJSON string, installFlags []string) error {
	err := k1.storeTFJson(tfJSON)
	if err != nil {
		return err
	}

	flags := []string{"install", "--tfjson", "tf.json", k1.ConfigurationFilePath}
	if len(installFlags) != 0 {
		flags = append(flags, installFlags...)
	}

	err = k1.run(flags...)
	if err != nil {
		return fmt.Errorf("k8s cluster deployment failed: %w", err)
	}

	return nil
}

// Upgrade runs 'kubeone upgrade' command to upgrade the cluster
func (k1 *Kubeone) Upgrade(upgradeFlags []string) error {
	flags := []string{"upgrade", "--tfjson", "tf.json", "--upgrade-machine-deployments", k1.ConfigurationFilePath}
	if len(upgradeFlags) != 0 {
		flags = append(flags, upgradeFlags...)
	}

	err := k1.run(flags...)
	if err != nil {
		return fmt.Errorf("k8s cluster upgrade failed: %w", err)
	}

	return nil
}

// Kubeconfig runs 'kubeone kubeconfig' command to create and store kubeconfig file
func (k1 *Kubeone) Kubeconfig() ([]byte, error) {
	var kubeconfigBuf bytes.Buffer

	exe := k1.build("kubeconfig", "--tfjson", "tf.json", k1.ConfigurationFilePath)
	testutil.StdoutTo(&kubeconfigBuf)(exe)

	if err := exe.Run(); err != nil {
		return nil, fmt.Errorf("creating kubeconfig failed: %w", err)
	}

	rawKubeconfig := kubeconfigBuf.String()
	homePath := os.Getenv("HOME")
	kubeconfigPath := fmt.Sprintf("%s/.kube/config", homePath)

	err := testutil.CreateFile(kubeconfigPath, rawKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("saving kubeconfig for given path %s failed: %w", kubeconfigPath, err)
	}

	return []byte(rawKubeconfig), nil
}

// Reset runs 'kubeone reset' command to destroy worker nodes and unprovision the cluster
func (k1 *Kubeone) Reset() error {
	err := k1.run("-v", "reset", "--tfjson", "tf.json", "--destroy-workers", k1.ConfigurationFilePath)
	if err != nil {
		return fmt.Errorf("destroing workers failed: %w", err)
	}

	return nil
}

// storeTFJson saves tf.json in the temporary test directory
func (k1 *Kubeone) storeTFJson(tfJSON string) error {
	tfJSONPath := fmt.Sprintf("%s/tf.json", k1.Dir)

	err := testutil.CreateFile(tfJSONPath, tfJSON)
	if err != nil {
		return fmt.Errorf("saving tf.json for given path %s failed: %w", tfJSONPath, err)
	}

	return nil
}

func (k1 *Kubeone) build(args ...string) *testutil.Exec {
	return testutil.NewExec("kubeone",
		testutil.WithArgs(args...),
		testutil.WithEnv(os.Environ()),
		testutil.InDir(k1.Dir),
	)
}

func (k1 *Kubeone) run(args ...string) error {
	return k1.build(args...).Run()
}
