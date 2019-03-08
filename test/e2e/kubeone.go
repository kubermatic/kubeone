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
