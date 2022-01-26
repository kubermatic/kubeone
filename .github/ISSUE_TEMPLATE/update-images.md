---
name: Update images to support Kubernetes 1.2x
about: Update components to use versions that support the latest Kubernetes release
labels: sig/cluster-management, kind/feature, Epic

---

<!--
This issue template is supposed to be used a starting point and is mostly like
NOT up-to-date!

You should first check images.go file and update the list below as appropriate.
https://github.com/kubermatic/kubeone/blob/master/pkg/templates/images/images.go
-->

Action items:

- [ ] Update the issue template to add/remove [images](https://github.com/kubermatic/kubeone/blob/master/pkg/templates/images/images.go) as appropriate

The following components/images should be updated:

- [ ] [AWS CCM](https://github.com/kubernetes/cloud-provider-aws) <!-- (PR reference|already the latest) -->
- [ ] [AWS CSI](https://github.com/kubernetes-sigs/aws-ebs-csi-driver) <!-- (PR reference|already the latest) -->
- [ ] [Azure CCM](https://github.com/kubernetes-sigs/cloud-provider-azure) <!-- (PR reference|already the latest) -->
- [ ] [AzureDisk CSI](https://github.com/kubernetes-sigs/azuredisk-csi-driver) <!-- (PR reference|already the latest) -->
- [ ] [AzureFile CSI](https://github.com/kubernetes-sigs/azurefile-csi-driver) <!-- (PR reference|already the latest) -->
- [ ] [DigitalOcean CCM](https://github.com/digitalocean/digitalocean-cloud-controller-manager) <!-- (PR reference|already the latest) -->
- [ ] [DigitalOcean CSI](https://github.com/digitalocean/csi-digitalocean) <!-- (PR reference|already the latest) -->
- [ ] [EquinixMetal CCM](https://github.com/equinix/cloud-provider-equinix-metal) <!-- (PR reference|already the latest) -->
- [ ] [Hetzner CCM](https://github.com/hetznercloud/hcloud-cloud-controller-manager) <!-- (PR reference|already the latest) -->
- [ ] [Hetzner CSI](https://github.com/hetznercloud/csi-driver) <!-- (PR reference|already the latest) -->
- [ ] [Nutanix CSI](https://github.com/nutanix/helm) <!-- (PR reference|already the latest) --> <!-- We intentionally use Helm charts repo, because the nutanix-csi repo is not up-to-date -->
- [ ] [OpenStack CCM](https://github.com/kubernetes/cloud-provider-openstack) <!-- (PR reference|already the latest) -->
- [ ] [OpenStack CSI](https://github.com/kubernetes/cloud-provider-openstack) <!-- (PR reference|already the latest) -->
- [ ] [vSphere CCM](https://github.com/kubernetes/cloud-provider-vsphere) <!-- (PR reference|already the latest) -->
- [ ] [vSphere CSI](https://github.com/kubernetes-sigs/vsphere-csi-driver) <!-- (PR reference|already the latest) -->
- [ ] [Cluster Autoscaler](https://github.com/kubernetes/autoscaler) <!-- (PR reference|already the latest) -->

Relevant to <!-- epic number -->
