package e2e

import (
	"errors"
	"fmt"
	"os"
)

const (
	AWS          = "aws"
	DigitalOcean = "digitalocean"

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
	// identifier aka. the build number, a unique identifier for the test run.
	idendifier string
}

// AWSProvisioner describes AWS provisioner
type AWSProvisioner struct {
	testPath  string
	terraform *terraform
}

// NewAWSProvisioner creates and initialize AWSProvisioner structure
func NewAWSProvisioner(testPath, identifier string) (*AWSProvisioner, error) {
	terraform := &terraform{
		terraformDir: "../../examples/terraform/aws/",
		idendifier:   identifier,
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

	_, err = executeCommand("", "rm", []string{"-rf", p.testPath}, nil)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

// DOProvisioner describes DigitalOcean provisioner
type DOProvisioner struct {
	testPath  string
	terraform *terraform
}

// NewDOProvisioner creates and initialize DOProvisioner structure
func NewDOProvisioner(testPath, identifier string) (*DOProvisioner, error) {
	terraform := &terraform{
		terraformDir: "../../examples/terraform/digitalocean/",
		idendifier:   identifier,
	}

	return &DOProvisioner{
		terraform: terraform,
		testPath:  testPath,
	}, nil
}

// Provision starts provisioning on DigitalOcean
func (p *DOProvisioner) Provision() (string, error) {
	doToken := os.Getenv("DIGITALOCEAN_TOKEN")
	if len(doToken) == 0 {
		return "", errors.New("unable to run the test suite, DIGITALOCEAN_TOKEN environment variable cannot be empty")
	}

	tf, err := p.terraform.initAndApply()
	if err != nil {
		return "", err
	}

	return tf, nil
}

// Cleanup destroys infrastructure created by terraform
func (p *DOProvisioner) Cleanup() error {
	err := p.terraform.destroy()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	_, err = executeCommand("", "rm", []string{"-rf", p.testPath}, nil)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

// initAndApply method to initialize a terraform working directory
// and build infrastructure
func (p *terraform) initAndApply() (string, error) {
	initCmd := []string{"init"}
	if len(p.idendifier) > 0 {
		initCmd = append(initCmd, fmt.Sprintf("--backend-config=key=%s", p.idendifier))
	}

	_, err := executeCommand(p.terraformDir, "terraform", initCmd, nil)
	if err != nil {
		return "", fmt.Errorf("terraform init command failed: %v", err)
	}

	_, err = executeCommand(p.terraformDir, "terraform", []string{"apply", "-auto-approve"}, nil)
	if err != nil {
		return "", fmt.Errorf("terraform apply command failed: %v", err)
	}

	return p.getTFJson()
}

// destroy method
func (p *terraform) destroy() error {
	_, err := executeCommand(p.terraformDir, "terraform", []string{"destroy", "-auto-approve"}, nil)
	if err != nil {
		return fmt.Errorf("terraform destroy command failed: %v", err)
	}
	return nil
}

// GetTFJson reads an output from a state file
func (p *terraform) getTFJson() (string, error) {
	tf, err := executeCommand(p.terraformDir, "terraform", []string{"output", fmt.Sprintf("-state=%v", tfStateFileName), "-json"}, nil)
	if err != nil {
		return "", fmt.Errorf("generating tf json failed: %v", err)
	}

	return tf, nil
}
