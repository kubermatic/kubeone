# Nodelocal DNS Cache

Upstream: https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dns/nodelocaldns

## Development

```
kubectl kustomize . |
yq |
sed 's/__PILLAR__DNS__DOMAIN__/{{ .Config.ClusterNetwork.ServiceDomainName }}/' \
> nodelocaldns.yaml
```
