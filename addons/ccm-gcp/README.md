# GCP Cloud Controller Manager

## Development

### Generate manifest YAML
```shell
kubectl kustomize . | yq > ccm-gcp.yaml
```
