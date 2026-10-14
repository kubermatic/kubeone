# AWS EBS CSI driver chart

The AWS EBS CSI manifest is generated from the [official Helm chart][helm-chart].

```shell
make update
```

Required manual modifications include:

* Images must be changed to `{{ .InternalImages.Get "..." }}` as appropriate
  * Make sure that you update the appropriate entries in images list

[helm-chart]: https://github.com/kubernetes-sigs/aws-ebs-csi-driver/tree/master/charts/aws-ebs-csi-driver
