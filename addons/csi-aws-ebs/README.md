# AWS EBS CSI driver chart

The AWS EBS CSI manifest is generated from the [official Helm chart][helm-chart]
via Kustomize (see `Kustomization`), which also patches images to
`{{ .InternalImages.Get "..." }}`.

```shell
make update
```

To bump the driver, update the chart `version` in `Kustomization` and the
appropriate entries in the images list.

[helm-chart]: https://github.com/kubernetes-sigs/aws-ebs-csi-driver/tree/master/charts/aws-ebs-csi-driver
