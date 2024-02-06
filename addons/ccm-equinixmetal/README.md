# Equinix Metal CCM

## Development

### Generate manifest YAML
```shell
kubectl kustomize . | yq > ccm-equinixmetal.yaml
```

Please note the `yq`, its usage is necessary.
