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
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc/v2"
)

const (
	cniCanal                  = "Canal (Calico + Flannel)"
	cniCanalValue             = "canal"
	cniCilium                 = "Cilium"
	cniCiliumValue            = "cilium"
	cniCiliumReplacement      = "Cilium (with kube-proxy replacement)"
	cniCiliumReplacementValue = "cilium-ebpf"
	cniExternal               = "External/Bring Your Own"
	cniExternalValue          = "external"

	featureEncryption      = "Encryption-at-rest"
	featureEncryptionValue = "encryptionProviders"
	featureCoreDNSPDB      = "CoreDNS PodDisruptionBudget"
	featureCoreDNSPDBValue = "coreDNSPDB"

	addonClusterAutoscaler = "cluster-autoscaler"
	addonBackupsRestic     = "backups-restic"
)

var (
	questionsCluster = func(defaultKubeVersion string) []*survey.Question {
		return []*survey.Question{
			{
				Name: "clusterName",
				Prompt: &survey.Input{
					Message: "Provide a name for your cluster",
					Help:    "This name will be used for various cloud and Kubernetes resources. It must be a valid DNS name.",
				},
				Validate: clusterNameValidator,
			},
			{
				Name: "cloudProvider",
				Prompt: &survey.Select{
					Message: "Select a provider where you want to deploy your cluster",
					Help:    "You'll need to provide an account and credentials yourself. If your provider is not in the list or you want to deploy on baremetal, use the 'None' option.",
					Options: cloudProviderSelectOptions(),
				},
				Validate: survey.Required,
			},
			{
				Name: "kubernetesVersion",
				Prompt: &survey.Input{
					Message: "What Kubernetes version do you want to use?",
					Help:    "See https://kubernetes.io/releases/patch-releases/ for more details about Kubernetes releases.",
					Default: defaultKubeVersion,
				},
				Validate: kubernetesVersionValidator,
			},
			{
				Name: "cni",
				Prompt: &survey.Select{
					Message: "What CNI do you want to use?",
					Help:    "Canal is the default option -- if you're not sure about this question, we recommend choosing Canal.",
					Options: []string{
						cniCanal,
						cniCilium,
						cniCiliumReplacement,
						cniExternal,
					},
					Default: cniCanal,
				},
				Validate: survey.Required,
			},
			{
				Name: "features",
				Prompt: &survey.MultiSelect{
					Message: "Do you want to enable any of optional cluster features?",
					Help: heredoc.Doc(`- Encryption-at-rest encrypts your Secrets when persisting them to etcd (https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/)
				- CoreDNS PodDisruptionBudget (PDB) creates a PDB for CoreDNS to ensure CoreDNS availability when running operations such as cluster upgrades
				`),
					Options: []string{
						featureEncryption,
						featureCoreDNSPDB,
					},
				},
			},
		}
	}

	questionAddons = func(providerNone bool) *survey.MultiSelect {
		return &survey.MultiSelect{
			Message: "Do you want to enable any of optional addons?",
			Help: heredoc.Doc(`
			- cluster-autoscaler scales up and down your worker nodes based on the resource consumption
			- backups-restic periodically backups etcd and PKI (certificates and keys) using Restic to a S3-compatible bucket
			`),
			Options: addonsSelectOptions(providerNone),
		}
	}

	questionsAddonAutoscaler = []*survey.Question{
		{
			Name: "minReplicas",
			Prompt: &survey.Input{
				Message: "Provide the minimum number of worker nodes in your cluster",
				Help:    "cluster-autoscaler will not scale down the number of your worker nodes less than value provided here.",
				Default: "1",
			},
			Validate: positiveNumberValidator,
		},
		{
			Name: "maxReplicas",
			Prompt: &survey.Input{
				Message: "Provide the maximum number of worker nodes in your cluster",
				Help:    "cluster-autoscaler will not scale up the number of your worker nodes more than value provided here.",
				Default: "1",
			},
			Validate: positiveNumberValidator,
		},
	}

	questionsAddonBackups = []*survey.Question{
		{
			Name: "resticPassword",
			Prompt: &survey.Password{
				Message: "Provide a password used to encrypt backups",
				Help:    "Restic will use this password to encrypt the etcd and PKI backups.",
			},
			Validate: survey.Required,
		},
		{
			Name: "s3Bucket",
			Prompt: &survey.Input{
				Message: "Provide a link to the S3-compatible bucket where to store backups. Leave this empty if you don't have a bucket yet",
				Help:    "Example value: 's3:///...'",
			},
		},
		{
			Name: "awsDefaultRegion",
			Prompt: &survey.Input{
				Message: "If this is an AWS S3 bucket, provide the default AWS region for your S3 bucket",
				Help:    "Example value: eu-west-3",
			},
		},
	}

	questionTerraformConfigs = &survey.Confirm{
		Message: "Do you want to use Terraform configs provided by KubeOne to manage the infrastructure?",
		Default: true,
	}

	questionsTerraformVars = func(workerPerAZ, clusterAutoscaler bool) []*survey.Question {
		qs := []*survey.Question{
			{
				Name: "sshPublicKeyPath",
				Prompt: &survey.Input{
					Message: "Provide a path to a public SSH key to be deployed on instances",
					Default: "~/.ssh/id_rsa.pub",
				},
				Validate: survey.Required,
			},
			{
				Name: "controlPlaneCount",
				Prompt: &survey.Input{
					Message: "How many control plane nodes do you want in your cluster?",
					Help:    "This number must be odd because of the etcd quorum.",
					Default: "3",
				},
				Validate: oddNumberValidator,
			},
		}

		if !clusterAutoscaler && workerPerAZ {
			qs = append(qs, &survey.Question{
				Name: "workerNodesCount",
				Prompt: &survey.Input{
					Message: "How many worker nodes do you want PER Availability Zone?",
					Help:    "KubeOne by default creates worker nodes in three availability zones. For example, if you choose 1 here, you'll have 3 worker nodes (one in each AZ). This can be any positive integer.",
					Default: "1",
				},
				Validate: positiveNumberValidator,
			})
		} else if !clusterAutoscaler {
			qs = append(qs, &survey.Question{
				Name: "workerNodesCount",
				Prompt: &survey.Input{
					Message: "How many worker nodes do you want in your cluster?",
					Help:    "This can be any positive integer. We recommend at least 2 worker nodes.",
					Default: "2",
				},
				Validate: positiveNumberValidator,
			})
		}

		return qs
	}

	questionPath = &survey.Input{
		Message: "Where do you want to save generated files?",
		Default: ".",
	}
)
