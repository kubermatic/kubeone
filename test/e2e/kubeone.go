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
	"time"

	"github.com/pkg/errors"

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
{{ if .CLUSTER_NETWORK_POD }}
clusterNetwork:
  podSubnet: "{{ .CLUSTER_NETWORK_POD }}"
  serviceSubnet: "{{ .CLUSTER_NETWORK_SERVICE }}"
{{ end }}
`

// Kubeone is wrapper around KubeOne CLI
type Kubeone struct {
	// KubeoneDir is a temporary directory for storing test files (e.g. tf.json)
	KubeoneDir string
	// ConfigurationFilePath is a path to the KubeOneCluster manifest
	ConfigurationFilePath string
}

// NewKubeone creates and initializes the Kubeone structure
func NewKubeone(kubeoneDir, configurationFilePath string) Kubeone {
	return Kubeone{
		KubeoneDir:            kubeoneDir,
		ConfigurationFilePath: configurationFilePath,
	}
}

// CreateConfig creates a KubeOneCluster manifest
func (p *Kubeone) CreateConfig(kubernetesVersion, providerName string,
	providerExternal bool, clusterNetworkPod string, clusterNetworkService string) error {
	variables := map[string]interface{}{
		"KUBERNETES_VERSION":      kubernetesVersion,
		"CLOUD_PROVIDER_NAME":     providerName,
		"CLOUD_PROVIDER_EXTERNAL": providerExternal,
		"CLUSTER_NETWORK_POD":     clusterNetworkPod,
		"CLUSTER_NETWORK_SERVICE": clusterNetworkService,
	}

	tpl, err := template.New("base").Parse(configurationTpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse KubeOne configuration template")
	}
	buf := bytes.Buffer{}
	if tplErr := tpl.Execute(&buf, variables); tplErr != nil {
		return errors.Wrap(tplErr, "failed to render KubeOne configuration template")
	}

	err = ioutil.WriteFile(p.ConfigurationFilePath, buf.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write KubeOne configuration manifest")
	}

	return nil
}

// Install runs 'kubeone install' command to provision the cluster
func (p *Kubeone) Install(tfJSON string, installFlags []string) error {
	// deliberate delay, to give nodes time to start
	time.Sleep(2 * time.Minute)

	err := p.storeTFJson(tfJSON)
	if err != nil {
		return err
	}

	flags := []string{"install", "--tfjson", "tf.json", p.ConfigurationFilePath}
	if len(installFlags) != 0 {
		flags = append(flags, installFlags...)
	}
	_, err = testutil.ExecuteCommand(p.KubeoneDir, "kubeone", flags, nil)
	if err != nil {
		return fmt.Errorf("k8s cluster deployment failed: %v", err)
	}
	return nil
}

// Upgrade runs 'kubeone upgrade' command to upgrade the cluster
func (p *Kubeone) Upgrade(upgradeFlags []string) error {
	flags := []string{"upgrade", "--tfjson", "tf.json", "--upgrade-machine-deployments", p.ConfigurationFilePath}
	if len(upgradeFlags) != 0 {
		flags = append(flags, upgradeFlags...)
	}
	_, err := testutil.ExecuteCommand(p.KubeoneDir, "kubeone", flags, nil)
	if err != nil {
		return fmt.Errorf("k8s cluster upgrade failed: %v", err)
	}
	return nil
}

// Kubeconfig runs 'kubeone kubeconfig' command to create and store kubeconfig file
func (p *Kubeone) Kubeconfig() ([]byte, error) {
	rawKubeconfig, err := testutil.ExecuteCommand(p.KubeoneDir, "kubeone", []string{"kubeconfig", "--tfjson", "tf.json", p.ConfigurationFilePath}, nil)
	if err != nil {
		return nil, fmt.Errorf("creating kubeconfig failed: %v", err)
	}

	homePath := os.Getenv("HOME")
	kubeconfigPath := fmt.Sprintf("%s/.kube/config", homePath)

	err = testutil.CreateFile(kubeconfigPath, rawKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("saving kubeconfig for given path %s failed: %v", kubeconfigPath, err)
	}
	return []byte(rawKubeconfig), nil
}

// Reset runs 'kubeone reset' command to destroy worker nodes and unprovision the cluster
func (p *Kubeone) Reset() error {
	_, err := testutil.ExecuteCommand(p.KubeoneDir, "kubeone", []string{"-v", "reset", "--tfjson", "tf.json", "--destroy-workers", p.ConfigurationFilePath}, nil)
	if err != nil {
		return fmt.Errorf("destroing workers failed: %v", err)
	}
	return nil
}

// storeTFJson saves tf.json in the temporary test directory
func (p *Kubeone) storeTFJson(tfJSON string) error {
	tfJSONPath := fmt.Sprintf("%s/tf.json", p.KubeoneDir)
	err := testutil.CreateFile(tfJSONPath, tfJSON)
	if err != nil {
		return fmt.Errorf("saving tf.json for given path %s failed: %v", tfJSONPath, err)
	}
	return nil
}
