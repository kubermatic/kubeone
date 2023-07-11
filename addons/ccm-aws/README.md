# AWS Cloud Controller Manager (CCM) chart

The AWS CCM manifest is generated from the [official Helm chart][helm-chart].

```shell
helm repo add aws-cloud-controller-manager https://kubernetes.github.io/cloud-provider-aws
helm repo update

helm template \
    --namespace="kube-system" \
    --values="generated-values-ccm" \
    --skip-tests \
    aws-cloud-controller-manager aws-cloud-controller-manager/aws-cloud-controller-manager
```

Required manual modifications include:

* Image must be changed to `{{ .InternalImages.Get "AwsCCM" }}`
  * Make sure that you update the `AwsCCM` entry in images list to include the new CCM version

[helm-chart]: https://github.com/kubernetes/cloud-provider-aws/tree/master/charts/aws-cloud-controller-manager
