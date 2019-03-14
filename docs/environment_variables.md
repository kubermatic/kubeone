# Environment Variables

This document lists all environment variables used by KubeOne and related components.

## Sourcing Environment Variables

Some configuration variables can be sourced from an environment variable by specifying name of the environment variable with the `env:` prefix.

For example, the following snippet is used to source the AWS Access Key and Secret Access Key to be used with Ark from `BACKUP_AWS_ACCESS_KEY_ID` and `BACKUP_AWS_SECRET_ACCESS_KEY` environment variables:

```yaml
backup:
  provider: 'aws'
  s3_access_key: 'env:BACKUP_AWS_ACCESS_KEY_ID'
  s3_secret_access_key: 'env:BACKUP_AWS_SECRET_ACCESS_KEY'
```

In the following table you can find all configuration variables with support for sourcing using the `env:` prefix:

| Variable | Type | Default Value | Description |
|----------|------|---------------|-------------|
| `backup.s3_access_key` | string | "" | The AWS Access Key to be used with [Ark (Velero)](https://github.com/heptio/velero/) |
| `backup.s3_secret_access_key` | string | "" | The AWS Secret Access Key to be used with [Ark (Velero)](https://github.com/heptio/velero/) |
| `hosts.ssh_agent_socket` | string | "" | Socket to be used for SSH |

## `machine-controller` Environment Variables

[`machine-controller`](https://github.com/kubermatic/machine-controller) is used to create worker nodes. It needs credentials with the appropriate permissions, so it can create machines and needed infrastructure. Those credentials are deployed on the cluster.

| Environment Variable | Description |
|---|---|
| `AWS_ACCESS_KEY_ID` | The AWS Access Key used for creating workers on AWS |
| `AWS_SECRET_ACCESS_KEY` | The AWS Secret Access Key used for creating workers on AWS |
| | |
| `DIGITALOCEAN_TOKEN` | The DigitalOcean API Access Token used for creating workers on DO |
| | |
| `HCLOUD_TOKEN` | The Hetzner API Access Token used for creating workers on Hetzner |
| | |
| `OS_AUTH_URL` | The URL of OpenStack Identity Service |
| `OS_USERNAME` | The username of the OpenStack user |
| `OS_PASSWORD` | The password of the OpenStack user |
| `OS_DOMAIN_NAME` | The name of the OpenStack domain |
| `OS_TENANT_NAME` | The name of the OpenStack tenant |
| | |
| `VSPHERE_ADDRESS` | The address of the vSphere instance |
| `VSPHERE_USERNAME` | The username of the vSphere user |
| `VSPHERE_PASSWORD` | The password of the vSphere user |
