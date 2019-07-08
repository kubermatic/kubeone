# Environment Variables

This document lists all environment variables used by KubeOne and related components.

### Sourcing Environment Variables

In the following table you can find all configuration variables with support for sourcing using the `env:` prefix:

| Variable                 | Type   | Default Value | Description               |
|--------------------------|--------|---------------|---------------------------|
| `hosts.ssh_agent_socket` | string | ""            | Socket to be used for SSH |

## Credentials

Credentials are required for terraform plans and the [`machine-controller`](https://github.com/kubermatic/machine-controller) to create nodes and such. It needs credentials with the appropriate permissions, so it can create machines and needed infrastructure. Those credentials are deployed on the cluster created by KubeOne.

<details><summary>AWS</summary><p>

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need an [IAM account](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html) with the appropriate permissions.

Once you have the IAM account you need to set the following variables:

| Environment Variable    | Description                                                |
|-------------------------|------------------------------------------------------------|
| `AWS_ACCESS_KEY_ID`     | The AWS Access Key used for creating workers on AWS        |
| `AWS_SECRET_ACCESS_KEY` | The AWS Secret Access Key used for creating workers on AWS |

If the environment variables are not set, KubeOne will try to read the default profile from `~/.aws/credentials`.

</p></details><details><summary>Azure</summary><p>

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need to setup credentials for your Azure cluster.

For the terraform reference please take a look at [Azure provider docs][3]

The following environment variables should be set:

| Environment Variable  | Description          |
|-----------------------|----------------------|
| `ARM_CLIENT_ID`       | Azure ClientID       |
| `ARM_CLIENT_SECRET`   | Azure Client secret  |
| `ARM_TENANT_ID`       | Azure TenantID       |
| `ARM_SUBSCRIPTION_ID` | Azure SubscriptionID |

</p></details><details><summary>Digital Ocean</summary><p>

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need an API Access Token with read and write permissions. You can refer to [the official documentation](https://www.digitalocean.com/docs/api/create-personal-access-token/) for guidelines for generating the token.

Once you have the API access token you need to set the `DIGITALOCEAN_TOKEN` environment variable containing the token:

| Environment Variable | Description                                                       |
|----------------------|-------------------------------------------------------------------|
| `DIGITALOCEAN_TOKEN` | The DigitalOcean API Access Token used for creating workers on DO |

</p></details><details><summary>GCE</summary><p>

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need an [Service Account](https://cloud.google.com/iam/docs/creating-managing-service-accounts) with the appropriate permissions.

The following environment variables should be set:

| Environment Variable | Description         |
|----------------------|---------------------|
| `GOOGLE_CREDENTIALS` | GCE Service Account |

</p></details><details><summary>Hetzner</summary><p>

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need a Hetzner API Token.

The following environment variables should be set:

| Environment Variable | Description                                                       |
|----------------------|-------------------------------------------------------------------|
| `HCLOUD_TOKEN`       | The Hetzner API Access Token used for creating workers on Hetzner |

</p></details><details><summary>OpenStack</summary><p>

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need to setup credentials for your OpenStack instance.

The following environment variables should be set:

| Environment Variable | Description                           |
|----------------------|---------------------------------------|
| `OS_AUTH_URL`        | The URL of OpenStack Identity Service |
| `OS_USERNAME`        | The username of the OpenStack user    |
| `OS_PASSWORD`        | The password of the OpenStack user    |
| `OS_DOMAIN_NAME`     | The name of the OpenStack domain      |
| `OS_TENANT_NAME`     | The name of the OpenStack tenant      |

</p></details><details><summary>Packet</summary><p>

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need an API Access Token. You can refer to [the official documentation][packet_support_docs] for guidelines for generating the token.

The following environment variables should be set:

| Environment Variable | Description       |
|----------------------|-------------------|
| `PACKET_AUTH_TOKEN`  | Packet auth token |
| `PACKET_PROJECT_ID`  | Packet project ID |

</p></details><details><summary>vSphere</summary><p>

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need to setup credentials for your vSphere cluster.

For the terraform reference please take a look at [vSphere provider docs][3]

The following environment variables should be set:

| Environment Variable | Description                         |
|----------------------|-------------------------------------|
| `VSPHERE_ADDRESS`    | The address of the vSphere instance |
| `VSPHERE_USERNAME`   | The username of the vSphere user    |
| `VSPHERE_PASSWORD`   | The password of the vSphere user    |

</p></details>
