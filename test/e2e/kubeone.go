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
	"fmt"
	"os"
)

// Kubeone structure
type Kubeone struct {
	// KubeoneDir temporary directory for test purpose
	KubeoneDir string
	// ConfigurationFile for Kubeone
	ConfigurationFile string
}

func NewKubeone(kubeoneDir, configurationFile string) Kubeone {
	return Kubeone{
		KubeoneDir:        kubeoneDir,
		ConfigurationFile: configurationFile,
	}
}

// Install starts k8s cluster deployment
func (p *Kubeone) Install(tfJSON string) error {

	err := p.storeTFJson(tfJSON)
	if err != nil {
		return err
	}
	_, err = executeCommand(p.KubeoneDir, "kubeone", []string{"install", "--tfjson", "tf.json", p.ConfigurationFile}, nil)
	if err != nil {
		return fmt.Errorf("k8s cluster deployment failed: %v", err)
	}
	return nil
}

func (p *Kubeone) Upgrade() error {
	_, err := executeCommand(p.KubeoneDir, "kubeone", []string{"upgrade", "--tfjson", "tf.json", "--upgrade-machine-deployments", p.ConfigurationFile}, nil)
	if err != nil {
		return fmt.Errorf("k8s cluster upgrade failed: %v", err)
	}
	return nil
}

// CreateKubeconfig creates and store kubeconfig
func (p *Kubeone) CreateKubeconfig() ([]byte, error) {
	rawKubeconfig, err := executeCommand(p.KubeoneDir, "kubeone", []string{"kubeconfig", "--tfjson", "tf.json", p.ConfigurationFile}, nil)
	if err != nil {
		return nil, fmt.Errorf("creating kubeconfig failed: %v", err)
	}

	homePath := os.Getenv("HOME")
	kubeconfigPath := fmt.Sprintf("%s/.kube/config", homePath)

	err = CreateFile(kubeconfigPath, rawKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("saving kubeconfig for given path %s failed: %v", kubeconfigPath, err)
	}
	return []byte(rawKubeconfig), nil
}

// DestroyWorkers cleanup method
func (p *Kubeone) Reset() error {
	_, err := executeCommand(p.KubeoneDir, "kubeone", []string{"-v", "reset", "--tfjson", "tf.json", "--destroy-workers", p.ConfigurationFile}, nil)
	if err != nil {
		return fmt.Errorf("destroing workers failed: %v", err)
	}
	return nil
}

// StoreTFJson saves tf.json in temporary test directory
func (p *Kubeone) storeTFJson(tfJSON string) error {
	tfJSONPath := fmt.Sprintf("%s/tf.json", p.KubeoneDir)
	err := CreateFile(tfJSONPath, tfJSON)
	if err != nil {
		return fmt.Errorf("saving tf.json for given path %s failed: %v", tfJSONPath, err)
	}
	return nil
}
