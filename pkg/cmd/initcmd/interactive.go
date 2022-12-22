/*
Copyright 2022 The KubeOne Authors.

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

package initcmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/common-nighthawk/go-figure"
)

type interactiveOpts struct {
	cluster               *modelCluster
	addons                []string
	addonAutoscaler       *modelAddonAutoscaler
	addonBackups          *modelAddonBackups
	generateTerraform     bool
	terraformVars         *modelTerraformVars
	terraformProviderVars map[string]interface{}
}

func newInteractiveOpts() *interactiveOpts {
	return &interactiveOpts{
		cluster:               &modelCluster{},
		addons:                []string{},
		addonAutoscaler:       &modelAddonAutoscaler{},
		addonBackups:          &modelAddonBackups{},
		generateTerraform:     false,
		terraformVars:         &modelTerraformVars{},
		terraformProviderVars: map[string]interface{}{},
	}
}

var (
	messageWelcome = heredoc.Doc(`

	Welcome to KubeOne!

	This is an interactive walkthrough to prepare you for creating a Kubernetes cluster using KubeOne.

	At the end of this walkthrough, you'll have:
	    - KubeOneCluster manifest ('kubeone.yaml') that instructs KubeOne how to provision your cluster
	    - Terraform configurations that can be used to create the infrastructure needed for your cluster
	
	You'll be asked to provide some basic information about your desired cluster. Depending on your answers, you might be asked some follow-up questions for additional parameters.

	Note: you can't get back to previous question, but you can modify the generated files before provisioning the cluster.
	
	Help is provided for some questions by typing in '?'. You can cancel this walkthrough any time by pressing CTRL+C.
	`)

	messageAddonAutoscaler = heredoc.Doc(`

	The selected cluster-autoscaler addon requires additional configuration.
	
	You need to provide the minimum and the maximum number of worker nodes. Those values must be positive integers. Provided values can be changed anytime by modifying the 'terraform.tfvars' file before provisioning the cluster or by modifying MachineDeployment object in the kube-system namespace after provisioning the cluster.
	`)

	messageAddonBackups = heredoc.Doc(`
		
	The selected backups-restic addon requires additional configuration.
	
	You'll be asked to provide a path to the bucket for storing backups. If you don't have a bucket, you can leave this question empty and provide it manually by modifying the generated 'kubeone.yaml' file. The default AWS region doesn't have to be provided if the provided bucket is not an AWS S3 bucket. All values can be changed by modifying the generated 'kubeone.yaml' file.
	`)

	messageTerraform = heredoc.Doc(`

	The next step is to decide how do you want to manage your infrastructure. The infrastructure in this case is virtual machines for the Kubernetes control plane and other resources needed for Kubernetes to function properly (e.g. VPCs, Firewalls, Load Balancers...).

	Note: the worker nodes are managed by machine-controller component which is deployed by KubeOne. Worker nodes are defined in 'output.tf' in the 'kubeone_workers' section.
	
	There are two options:
	    - Manage infrastructure using Terraform
	    - Manage infrastructure manually or using some other tool
	
	KubeOne provides:
	    - Example Terraform configs for all supported providers that can be used to manage the needed infrastructure
	    - Terraform integration in a way that KubeOne can read exported Terraform state to determine information about the infrastructure
	
	Using provided example Terraform configs is recommended for getting started with KubeOne. Terraform integration can be used with any other Terraform configs, but in this case, the 'output.tf' file must follow the structure mandated by KubeOne.

	If you don't want to use Terraform, you can manually extend your generated KubeOneCluster manifest ('kubeone.yaml') with information about your infrastructure.
	`)
)

func InitInteractive(defaultKubeVersion string) (*GenerateOpts, error) {
	iOpts := newInteractiveOpts()

	k1Figure := figure.NewFigure("KUBEONE", "", true)
	k1Figure.Print()

	// Ask initial cluster questions
	fmt.Println(messageWelcome)
	if err := survey.Ask(questionsCluster(defaultKubeVersion), iOpts.cluster); err != nil {
		return nil, err
	}

	var providerNone bool
	if iOpts.cluster.CloudProvider == ValidProviders["none"].title {
		providerNone = true
	}

	// Ask questions about addons
	if err := survey.AskOne(questionAddons(providerNone), &iOpts.addons); err != nil {
		return nil, err
	}

	var autoscaler bool
	if stringSliceIncludes(iOpts.addons, addonClusterAutoscaler) {
		fmt.Println(messageAddonAutoscaler)

		if err := survey.Ask(questionsAddonAutoscaler, iOpts.addonAutoscaler); err != nil {
			return nil, err
		}

		autoscaler = true
	}

	if stringSliceIncludes(iOpts.addons, addonBackupsRestic) {
		fmt.Println(messageAddonBackups)

		if err := survey.Ask(questionsAddonBackups, iOpts.addonBackups); err != nil {
			return nil, err
		}
	}

	// None cloud provider doesn't have Terraform support, so we skip straight to config generation.
	if providerNone {
		return iOpts.parseInteractiveOpts()
	}

	// Ask questions about Terraform
	fmt.Println(messageTerraform)
	if err := survey.AskOne(questionTerraformConfigs, &iOpts.generateTerraform, nil); err != nil {
		return nil, err
	}

	if iOpts.generateTerraform {
		_, cp := cloudProviderForSelectedOption(iOpts.cluster.CloudProvider)

		if err := survey.Ask(questionsTerraformVars(cp.workerPerAZ, autoscaler), iOpts.terraformVars, nil); err != nil {
			return nil, err
		}
		if err := survey.Ask(providerTFVarsQuestions(iOpts.cluster.CloudProvider), &iOpts.terraformProviderVars, nil); err != nil {
			return nil, err
		}
	}

	return iOpts.parseInteractiveOpts()
}

func (opts *interactiveOpts) parseInteractiveOpts() (*GenerateOpts, error) {
	var path string
	if err := survey.AskOne(questionPath, &path, nil); err != nil {
		return nil, err
	}

	cpKey, cp := cloudProviderForSelectedOption(opts.cluster.CloudProvider)

	providerName := cp.alternativeName
	if providerName == "" {
		providerName = cpKey
	}

	gOpts := NewGenerateOpts(path, providerName, opts.cluster.ClusterName, opts.cluster.KubernetesVersion, opts.generateTerraform)

	// CNI
	switch opts.cluster.CNI {
	case cniCanal:
		gOpts.cni = cniCanalValue
	case cniCilium:
		gOpts.cni = cniCiliumValue
	case cniCiliumReplacement:
		gOpts.cni = cniCiliumReplacementValue
	case cniExternal:
		gOpts.cni = cniExternalValue
	}

	// Features
	if stringSliceIncludes(opts.cluster.Features, featureEncryption) {
		gOpts.enableFeatureEncryption = true
	}
	if stringSliceIncludes(opts.cluster.Features, featureCoreDNSPDB) {
		gOpts.enableFeatureCoreDNSPDB = true
	}

	// Addons
	if stringSliceIncludes(opts.addons, addonClusterAutoscaler) {
		gOpts.enableAddonAutoscaler = true
		gOpts.terraformVars[tfvarName(opts.addonAutoscaler, "MinReplicas")] = opts.addonAutoscaler.MinReplicas
		gOpts.terraformVars[tfvarName(opts.addonAutoscaler, "MaxReplicas")] = opts.addonAutoscaler.MaxReplicas

		// We don't ask for number of worker nodes if cluster-autoscaler is enabled, instead we take minimum replicas
		opts.terraformVars.WorkerNodesCount = opts.addonAutoscaler.MinReplicas
	}
	if stringSliceIncludes(opts.addons, addonBackupsRestic) {
		gOpts.enableAddonBackups = true
		gOpts.addonBackupsPassword = opts.addonBackups.ResticPassword
		gOpts.addonBackupsS3Bucket = opts.addonBackups.S3Bucket
		gOpts.addonBackupsDefaultAWSRegion = opts.addonBackups.AWSDefaultRegion
	}

	if opts.terraformVars != nil {
		gOpts.terraformVars[tfvarName(opts.terraformVars, "SSHPublicKeyPath")] = opts.terraformVars.SSHPublicKeyPath
		gOpts.terraformVars[tfvarName(opts.terraformVars, "ControlPlaneCount")] = opts.terraformVars.ControlPlaneCount
		gOpts.terraformVars[tfvarName(opts.terraformVars, "WorkerNodesCount")] = opts.terraformVars.WorkerNodesCount
	}
	if opts.terraformProviderVars != nil {
		for k, v := range opts.terraformProviderVars {
			switch ans := v.(type) {
			case survey.OptionAnswer:
				if k == "os" || k == "worker_os" {
					gOpts.terraformVars[k] = osForSelectedOption(cp, ans.Index)
				} else {
					gOpts.terraformVars[k] = ans.Value
				}
			case string:
				gOpts.terraformVars[k] = ans
			default:
				return nil, fmt.Errorf("unknown answer type: %v", ans)
			}
		}
	}

	return gOpts, nil
}
