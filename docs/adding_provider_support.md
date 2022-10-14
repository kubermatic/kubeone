# Adding Support For A New Provider

KubeOne is supposed to work on any cluster out-of-box as long as there is no
need for additional and potentially complex configuration. However, in order to
utilize all features of KubeOne or if additional configuration is needed to
provision the cluster, KubeOne and Kubermatic [machine-controller][1] need to
support the provider.

This walkthrough shows what are prerequisites and how to implement support for a
provider.

## Prerequisites

Before KubeOne can support a provider, Kubermatic [machine-controller][1] needs
to support creating worker nodes for that provider. For details how to implement
support for the provider in the `machine-controller`, please check the
[`machine-controller` repository][1].

Once there is support for the provider in the machine-controller, KubeOne
support is implemented by:

* adding example Terraform scripts
* adding the needed API types for a new provider to the KubeOneCluster API
* implementing Terraform integration for worker nodes
* (optionally) adding provider-specific configuration for provisioning the
  cluster
* adding E2E tests verifying the cluster creation and the cluster upgrade
  processes

## Adding Example Terraform Scripts

In the [`./examples/terraform` directory][2] you can find example Terraform
scripts for each supported provider. You can check them for the reference how
scripts should look like and what variables should be exposed and returned as
the output.

The [Terraform documentation][3] has a document showing how to provision the
infrastructure for each Terraform-supported provider. If there is no Terraform
support for such provider, please create an issue in the KubeOne repository or
contact us on the [`#kubeone` channel on Kubernetes Slack][4] to discuss about
potential alternatives.

Once Terraform scripts are done, you should proceed to adding the needed API
types to the KubeOneCluster API.

## Adding The Needed API Types

The [KubeOneCluster API][5] has references for each supported provider used for
various tasks including validating the provider specific settings and grabbing
credentials for the `machine-controller`. The API is defined in the
[`pkg/apis/kubeone` package][5].

The first change you need to make is to add a `CloudProviderName` value for the
provider to the [internal API][6] and to the versioned APIs (e.g.
[`v1alpha1`][7]). Once the `CloudProviderName` value is in the place, you need
to add validation for provider-specific configuration to the
[`ValidateCloudProviderSpec` function][8]. In validation we usually check are
all [CloudProviderSpec][9] fields required by the provider populated. For
example, if provider requires `cloudConfig` to be present, we can check it here.
If the provider has an [in-tree cloud controller manager][10] please add an
entry to the [`CloudProviderInTree` function][11].

Finally, you need to configure how credentials for `machine-controller` and the
external CCM (if applicable) are obtained. First, you need to add a constant
with the environment variable name used by the `machine-controller` to the
[`pkg/util/credentials`][12] package. Then, you need to add the logic for
obtaining credentials to the [`ProviderCredentials` function][13]. Credentials
are usually obtained by reading them from the environment variables, but custom
workflows are supported as well. For example AWS uses the AWS SDK to fetch the
credentials.

With the API types added to support the new provider, you need to implement
Terraform integration for worker nodes.

## Implementing Terraform Integration for Worker Nodes

The Terraform Integration for worker nodes is used to source common information
used to create the worker nodes. For example, information such as region,
instance size, VPC IDs, are usually same as for the control plane nodes. If
infrastructure is provisioned using Terraform, those information can be sourced
from the Terraform state, so operator doesn't have to manually provide them upon
provisioning the cluster. The Terraform integration is implemented in the
[`pkg/terraform` package][14].

For the reference what fields are part of the WorkerSet, check the
`CloudProviderSpec` API implemented in the machine-controller. You can find
example manifests with all available options in the `examples` directory in the
[`machine-controller` repository][15] or check out the [API implementation][16]
for more details.

First, you need to implement a `WorkerConfig` structure in the
`pkg/templates/machinecontroller/cloudprovider_specs.go` file with fields that
will be sourced from the Terraform state. You can refer to the [`AWSSpec`][17]
as an example.

Then, you need to implement a `update<CloudProviderName>Workerset` function that
replaces values in the KubeOne WorkerSet configuration with values from the
Terraform state. See the [`updateAWSWorkerset` function][18] as an reference how
the function should look like.

Finally, update the [`Apply` function][19] to utilize the newly-added functions.

With those changes in the place, KubeOne will automatically source information
about worker nodes from the Terraform output if it's provided.

## Adding Provider-Specific Configuration For Provisioning The Cluster

If the provider you're implementing requires provider-specific steps for
provisioning or upgrading the cluster make sure to add them to the
[`installer`][20] and/or [`upgrader`][21] packages.

## Adding E2E tests

End-to-End (E2E) tests are used to verify is KubeOne correctly provisioning
clusters on the provider. This is done by creating the needed infrastructure,
provisioning the cluster, and running Kubernetes conformance tests.

First, you need to implement the `Provisioner` interface for the provider in
[`test/e2e/provisioner.go`][22]. You can see the [`AWSProvisioner`][23]
implementation as an example.

The `Provisioner` should implement three functions:

* [`New<CloudProviderName>Provisioner`][24] that builds the provisioner.
* [`Provision`][25] that verifies the credentials and provisions the cluster
  using Terraform.
* [`Cleanup`][26] that cleanups the cluster after tests are done.

Once the `Provisioner` interface implementation is in the place, you need to add
a testcase to the [`TestClusterConformance` function][27]. This function creates
a cluster using example Terraform scripts, provisions the cluster, and then runs
the conformance tests on the cluster. The manifests used to provision the
cluster are stored in the [`testdata`][28] directory.

Then, add a test case to the [`TestClusterUpgrade` function][29] which
provisions the cluster with older Kubernetes version, then upgrades to the newer
version and at the end runs the conformance tests.

After the PR is merged, the KubeOne maintainers will create a ProwJob that runs
E2E tests for a newly-added provider in the CI pipeline.

[1]: https://github.com/kubermatic/machine-controller
[2]: https://github.com/kubermatic/kubeone/tree/19e5e6bf792ae47d65bd8adf75f390c74159e3de/examples/terraform
[3]: https://www.terraform.io/docs/providers/
[4]: http://slack.k8s.io/
[5]: https://github.com/kubermatic/kubeone/tree/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/apis/kubeone
[6]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/apis/kubeone/types.go#L83-L92
[7]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/apis/kubeone/v1alpha1/types.go#L83-L92
[8]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/apis/kubeone/validation/validation.go#L57-L77
[9]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/apis/kubeone/types.go#L94-L99
[10]: https://github.com/kubernetes/kubernetes/tree/release-1.14/pkg/cloudprovider/providers
[11]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/apis/kubeone/helpers.go#L57-L66
[12]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/util/credentials/credentials.go#L30-L48
[13]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/util/credentials/credentials.go#L57-L128
[14]: https://github.com/kubermatic/kubeone/tree/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/terraform
[15]: https://github.com/kubermatic/machine-controller/tree/main/examples
[16]: https://github.com/kubermatic/machine-controller/tree/main/pkg/cloudprovider/provider
[17]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/templates/machinecontroller/cloudprovider_specs.go#L19-L31
[18]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/terraform/config.go#L177-L219
[19]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/terraform/config.go#L145-L162
[20]: https://github.com/kubermatic/kubeone/tree/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/installer
[21]: https://github.com/kubermatic/kubeone/tree/19e5e6bf792ae47d65bd8adf75f390c74159e3de/pkg/upgrader
[22]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/test/e2e/provisioner.go
[23]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/test/e2e/provisioner.go#L50-L54
[24]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/test/e2e/provisioner.go#L56-L67
[25]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/test/e2e/provisioner.go#L69-L83
[26]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/test/e2e/provisioner.go#L85-L98
[27]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/test/e2e/conformance_test.go#L30-L172
[28]: https://github.com/kubermatic/kubeone/tree/19e5e6bf792ae47d65bd8adf75f390c74159e3de/test/e2e/testdata
[29]: https://github.com/kubermatic/kubeone/blob/19e5e6bf792ae47d65bd8adf75f390c74159e3de/test/e2e/upgrade_test.go#L42-L199
