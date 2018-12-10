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

// DeployCluster starts k8s cluster deployment
func (p *Kubeone) Install(tfJSON string) error {

	err := p.storeTFJson(tfJSON)
	if err != nil {
		return err
	}
	_, stderr, exitCode := executeCommand(p.KubeoneDir, "kubeone", []string{"install", "--tfjson", "tf.json", p.ConfigurationFile})
	if exitCode != 0 {
		return fmt.Errorf("k8s cluster deployment failed: %s", stderr)
	}
	return nil
}

// CreateKubeconfig creates and store kubeconfig
func (p *Kubeone) CreateKubeconfig() error {
	rawKubeconfig, stderr, exitCode := executeCommand(p.KubeoneDir, "kubeone", []string{"kubeconfig", "--tfjson", "tf.json", p.ConfigurationFile})
	if exitCode != 0 {
		return fmt.Errorf("creating kubeconfig failed: %s", stderr)
	}

	homePath := os.Getenv("HOME")
	kubeconfigPath := fmt.Sprintf("%s/.kube/config", homePath)

	err := CreateFile(kubeconfigPath, rawKubeconfig)
	if err != nil {
		return fmt.Errorf("saving kubeconfig failed: %s", err)
	}
	return nil
}

// DestroyWorkers cleanup method
func (p *Kubeone) Reset() error {
	_, stderr, exitCode := executeCommand(p.KubeoneDir, "kubeone", []string{"-v", "reset", "--tfjson", "tf.json", "--destroy-workers", p.ConfigurationFile})
	if exitCode != 0 {
		return fmt.Errorf("destroing workers failed: %s", stderr)
	}
	return nil
}

// StoreTFJson saves tf.json in temporary test directory
func (p *Kubeone) storeTFJson(tfJSON string) error {
	tfJSONPath := fmt.Sprintf("%s/tf.json", p.KubeoneDir)
	err := CreateFile(tfJSONPath, tfJSON)
	if err != nil {
		return fmt.Errorf("saving tf.json failed: %s", err)
	}
	return nil
}
