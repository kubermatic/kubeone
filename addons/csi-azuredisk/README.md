# Azure CSI driver for azuredisk

See more: https://github.com/kubernetes-sigs/azuredisk-csi-driver/tree/master/charts

## Development

```shell
kubectl kustomize --enable-helm . | yq > csi-azuredisk.yaml
```
