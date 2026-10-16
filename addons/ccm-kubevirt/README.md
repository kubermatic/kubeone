# KubeVirt CCM

The KubeVirt cloud-controller-manager runs in the infra cluster's
`InfraNamespace`, not in the tenant cluster, so the infra credentials never
reach the tenant cluster. This addon only contains its tenant cluster identity
(ServiceAccount, token Secret and RBAC).

The infra side (ServiceAccount, Role, `cloud-config` ConfigMap, tenant
kubeconfig Secret and Deployment) is built in Go, see
`pkg/cloudprovider/kubevirt/ccm`, and follows the upstream `config/isolated`
overlay of https://github.com/kubevirt/cloud-provider-kubevirt.
