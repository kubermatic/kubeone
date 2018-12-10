package e2e

import (
	"errors"
	"fmt"
	"os"
)

const (
	AWS          Provider = 0
	DIGITALOCEAN Provider = 1
)

type Provider int

type Provisioner interface {
	Provision() (string, error)
	Cleanup() error
}

// terraform structure
type terraform struct {
	// terraformDir the path to where your terraform code is located
	terraformDir string
	// envVars terraform environment variables
	envVars map[string]string
}

// AWSProvisioner describes AWS provisioner
type AWSProvisioner struct {
	testPath  string
	terraform *terraform
}

func NewAWSProvisioner(region, testName, testPath string) *AWSProvisioner {

	terraform := &terraform{terraformDir: "../../terraform/aws/",
		envVars: map[string]string{
			"TF_VAR_ssh_public_key_file": os.Getenv("SSH_PUBLIC_KEY_FILE"),
			"TF_VAR_cluster_name":        testName,
			"TF_VAR_aws_region":          region,
		}}

	return &AWSProvisioner{
		terraform: terraform,
		testPath:  testPath,
	}
}

func (p *AWSProvisioner) Provision() (string, error) {

	awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if len(awsKeyID) == 0 || len(awsSecret) == 0 {
		return "", errors.New("unable to run the test suite, AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY environment variables cannot be empty")
	}

	tf, err := p.terraform.initAndApply()
	if err != nil {
		return "", err
	}

	return tf, nil

}

func (p *AWSProvisioner) Cleanup() error {

	err := p.terraform.destroy()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	_, stderr, exitCode := executeCommand("", "rm", []string{"-rf", p.testPath})
	if exitCode != 0 {
		return fmt.Errorf("%s", stderr)
	}

	return nil
}

// initAndApply method to initialize a terraform working directory
// and build infrastructure
func (p *terraform) initAndApply() (string, error) {

	for k, v := range p.envVars {
		os.Setenv(k, v)
	}

	_, stderr, exitCode := executeCommand(p.terraformDir, "terraform", []string{"init"})
	if exitCode != 0 {
		return "", fmt.Errorf("terraform init command failed: %s", stderr)
	}

	_, stderr, exitCode = executeCommand(p.terraformDir, "terraform", []string{"apply", "-auto-approve"})
	if exitCode != 0 {
		return "", fmt.Errorf("terraform apply command failed: %s", stderr)
	}

	return p.getTFJson()
}

// destroy method
func (p *terraform) destroy() error {
	_, stderr, exitCode := executeCommand(p.terraformDir, "terraform", []string{"destroy", "-auto-approve"})
	if exitCode != 0 {
		return fmt.Errorf("terraform destroy command failed: %s", stderr)
	}
	return nil
}

// GetTFJson reads an output from a state file
func (p *terraform) getTFJson() (string, error) {
	tf, stderr, exitCode := executeCommand(p.terraformDir, "terraform", []string{"output", "-state=terraform.tfstate", "-json"})
	if exitCode != 0 {
		return "", fmt.Errorf("generating tf json failed: %s", stderr)
	}

	return tf, nil
}
