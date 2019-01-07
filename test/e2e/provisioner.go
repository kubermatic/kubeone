package e2e

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	AWS = "aws"

	tfStateFileName = "terraform.tfstate"
)

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
	// identifier aka. the build number, a unique identifier for the test run.
	idendifier string
}

// AWSProvisioner describes AWS provisioner
type AWSProvisioner struct {
	testPath  string
	terraform *terraform
}

// NewAWSProvisioner creates and initialize AWSProvisioner structure
//
// Note:
// It also stores terraform related environment variables in tf.env file,
// primarily required to automatically destroy state (resources) in the event of failure
func NewAWSProvisioner(region, testName, testPath, identifier string) (*AWSProvisioner, error) {
	terraform := &terraform{terraformDir: "../../terraform/aws/",
		envVars: map[string]string{
			"TF_VAR_ssh_public_key_file": os.Getenv("SSH_PUBLIC_KEY_FILE"),
			"TF_VAR_cluster_name":        testName,
			"TF_VAR_aws_region":          region,
		}, idendifier: identifier}

	envVarsStr := ""
	for k, v := range terraform.envVars {
		envVarsStr = fmt.Sprintf("%s%s=%s\n", envVarsStr, k, v)
	}

	envVarFilePath := strings.Join([]string{terraform.terraformDir, "tf.env"}, "/")
	err := CreateFile(envVarFilePath, envVarsStr)
	if err != nil {
		return nil, fmt.Errorf("unable to write data to file = %s", envVarFilePath)
	}

	return &AWSProvisioner{
		terraform: terraform,
		testPath:  testPath,
	}, nil
}

// Provision starts provisioning on AWS
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

// Cleanup destroys infrastructure created by terraform
func (p *AWSProvisioner) Cleanup() error {

	err := p.terraform.destroy()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	_, err = executeCommand("", "rm", []string{"-rf", p.testPath})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

// initAndApply method to initialize a terraform working directory
// and build infrastructure
func (p *terraform) initAndApply() (string, error) {

	for k, v := range p.envVars {
		os.Setenv(k, v)
	}

	initCmd := []string{"init"}
	if len(p.idendifier) > 0 {
		initCmd = append(initCmd, fmt.Sprintf("--backend-config=key=%s", p.idendifier))
	}

	_, err := executeCommand(p.terraformDir, "terraform", initCmd)
	if err != nil {
		return "", fmt.Errorf("terraform init command failed: %v", err)
	}

	_, err = executeCommand(p.terraformDir, "terraform", []string{"apply", "-auto-approve"})
	if err != nil {
		return "", fmt.Errorf("terraform apply command failed: %v", err)
	}

	return p.getTFJson()
}

// destroy method
func (p *terraform) destroy() error {
	_, err := executeCommand(p.terraformDir, "terraform", []string{"destroy", "-auto-approve"})
	if err != nil {
		return fmt.Errorf("terraform destroy command failed: %v", err)
	}
	// remove the state file to indicate that the infrastructure has been destroyed successfully
	return os.Remove(strings.Join([]string{p.terraformDir, tfStateFileName}, "/"))
}

// GetTFJson reads an output from a state file
func (p *terraform) getTFJson() (string, error) {
	tf, err := executeCommand(p.terraformDir, "terraform", []string{"output", fmt.Sprintf("-state=%v", tfStateFileName), "-json"})
	if err != nil {
		return "", fmt.Errorf("generating tf json failed: %v", err)
	}

	return tf, nil
}
