# Adding Support For A New Provider

KubeOne is supposed to work on any cluster out-of-box as long as there is no need for additional and potentially complex configuration. However, if additional configuration is needed and in order to utilize all features of KubeOne, KubeOne and [Kubermatic machine-controller](https://github.com/kubermatic/machine-controller) need to support that provider.

This walkthrough shows what are prerequisites and how to implement support for a provider.

## Prerequisites

Before KubeOne can support a provider, [Kubermatic machine-controller](https://github.com/kubermatic/machine-controller) needs to support creating worker nodes for that provider. For details how to implement support for the provider in the machine-controller, please check the [machine-controller repository](https://github.com/kubermatic/machine-controller).

Once there is support for the provider in the machine-controller, KubeOne support is implemented by:

* adding example Terraform scripts,
* modifying the KubeOne API to support the provider,
* implementing Terraform integration for worker nodes,
* (optionally) adding provider-specific configuration for provisioning the cluster,
* adding E2E tests verifying the cluster creation and cluster upgrade processes.

## Adding Example Terraform Scripts

In the [`./examples/terraform` directory](https://github.com/kubermatic/kubeone/tree/master/examples/terraform) you can find example Terraform scripts for each supported provider. You can check them for the reference how scripts should look like and what variables you should expose and return as the output.

The [Terraform documentation](https://www.terraform.io/docs/providers/) has a document showing how to provision the infrastructure for each supported provider. If there is no Terraform support for such provider, please create an issue in the KubeOne repository or contact us over [Slack](https://kubermatic.slack.com/messages/KubeOne) to discuss about potential alternatives.

Once Terraform scripts are done, you should proceed to modifying the KubeOne API to support the provider.

## Modifying The KubeOne API To Support The Provider

The KubeOne API has references for each supported provider used for various tasks including validating the provider specific settings and grabbing credentials for the machine-controller. The API is defined in the [`pkg/config` package](https://github.com/kubermatic/kubeone/tree/master/pkg/config).

The first change you need to make is to add a [`ProviderName` value](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/config/cluster.go#L174-L182) for the provider. Once the `ProviderName` value is in the place, you need to add validation for provider-specific configuration [to the Validate function](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/config/cluster.go#L191-L208). Usually in validation we check are all [ProviderConfig](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/config/cluster.go#L184-L189) fields required by the provider populated. For example, if provider requires `cloud_config` to be present, we can check it here. If the provider has an [in-tree cloud controller manager](https://github.com/kubernetes/kubernetes/tree/master/pkg/cloudprovider) please add an entry to the [`CloudProviderInTree` function](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/config/cluster.go#L210-L219).

Finally, you need to configure how credentials for machine-controller are obtained by modifying the [`ProviderCredentials` function](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/config/cluster.go#L210-L219). Credentials are usually obtained by reading them from the environment variables, but custom workflows are supposed as well. For example AWS uses the AWS SDK to fetch credentials.

With API modified to support the new provider, you need to implement Terraform integration for worker nodes, so it's possible to easily create worker nodes.

## Implementing Terraform Integration for Worker Nodes

The Terraform Integration for Worker Nodes is used to source common information used to create the worker nodes. For example, information such as region, instance size, VPC IDs, are usually same as for the control plane nodes. If infrastructure is provisioned using Terraform, those information can be easily sourced from the Terraform state, so operator doesn't have to manually provide them upon provisioning the cluster. The Terraform integration is implemented in the [`pkg/terraform` package](https://github.com/kubermatic/kubeone/tree/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/terraform).

For the reference what fields are part of the WorkerSet, check the `CloudProviderSpec` API implemented in the machine-controller. You can find example manifests with all available options in the [examples directory of machine-controller](https://github.com/kubermatic/machine-controller/tree/master/examples) or check out the [API implementation](https://github.com/kubermatic/machine-controller/tree/master/pkg/cloudprovider/provider) for more details.

First, you need to implement a `WorkerConfig` structure in the `pkg/terraform` package with fields that will be sourced from the Terraform state. You can refer to the [`awsWorkerConfig` structure](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/terraform/config.go#L66-L76) as an example.

**Note:** Using types other than string and integer in the `WorkerConfig` structure may not work as expected.

Then, you need to implement a `updateProviderNameWorkerset` function that replaces values in the KubeOne WorkerSet configuration with values from the Terraform state. See the [`updateAWSWorkerset` function](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/terraform/config.go#L220-L271) as an reference how the function should look like.

Finally, update the [`Apply` function](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/terraform/config.go#L192-L205) to utilize the newly-added functions.

With those changes in the place, KubeOne will automatically source information about worker nodes from the Terraform output if it's provided.

## Adding Provider-Specific Configuration For Provisioning The Cluster

If the provider you're implementing requires provider-specific steps for provisioning or upgrading the cluster make sure to add them to the [`installer`](https://github.com/kubermatic/kubeone/tree/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/installer) or [`upgrader`](https://github.com/kubermatic/kubeone/tree/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/pkg/upgrader) package.

## Adding E2E tests

End-to-End (E2E) tests are used to verify is KubeOne correctly provisioning clusters on the provider. This is done by creating the needed infrastructure, provisioning the cluster, and running Kubernetes conformance tests.

First, you should add a test case to the [`TestClusterConformance` function](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/test/e2e/conformance_test.go#L26). This function creates a cluster using example Terraform scripts, provisions the cluster, and then runs the conformance tests on the cluster. The manifest used to provision the cluster should be stored in the [`testdata`](https://github.com/kubermatic/kubeone/tree/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/test/e2e/testdata) directory.

Then, add a test case to the [`TestClusterUpgrade` function](https://github.com/kubermatic/kubeone/blob/8b35b17876dd2f547205a3ab8468cf3b5d37d95c/test/e2e/upgrade_test.go) which provisions the cluster with older Kubernetes version and then upgrades to the newer version.

After the PR is merged, the KubeOne maintainers will create a ProwJob that runs E2E tests for a newly-added provider in the CI pipeline.