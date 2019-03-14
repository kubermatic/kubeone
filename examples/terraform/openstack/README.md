# OpenStack

## Environment variables

So Terraform can communicate with the OpenStack API it is necessary that the required environment variables are set correctly.
Example
```bash
export OS_AUTH_URL="https://some-keystone-endpoint:5000/v3"
export OS_IDENTITY_API_VERSION=3
export OS_USERNAME="some-username"
export OS_PASSWORD="some-password"
export OS_REGION_NAME="region1"
export OS_INTERFACE=public
export OS_ENDPOINT_TYPE=public
export OS_USER_DOMAIN_NAME="Default"
export OS_PROJECT_ID="some-project-id"
```

## Terraform

```
$ terraform init
$ terraform plan
$ terraform apply
$ terraform output -json > tf.json
```

## KubeOne

Create a config.yaml and make sure you specify `.provider.cloud_config`.

```bash
kubeone install --tfjson tf.json config.yaml
```
