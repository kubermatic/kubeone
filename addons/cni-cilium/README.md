# Cilium CNI addon

This addon is used to deploy [Cilium CNI](https://cilium.io/).

## Available parameters

This section what [addon parameters][params] can be used with this addon.

[params]: https://docs.kubermatic.com/kubeone/v1.8/guides/addons/#parameters

* `HubbleUI` - used to enable/disable Hubble UI listening on an IPv6 interface
  * `true` (default): enable listening on IPv6
  * `false`: disable listening on IPv6
  * Note: enabling this parameter requires having IPv6 support enabled in kernel

## Development

```shell
kubectl kustomize --enable-helm . | yq > cilium.yaml
```
