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

- [ ] [AWS CCM](https://github.com/kubernetes/cloud-provider-aws)
- [ ] [AWS CSI](https://github.com/kubernetes-sigs/aws-ebs-csi-driver)
- [ ] [Azure CCM](https://github.com/kubernetes-sigs/cloud-provider-azure)
- [ ] [AzureDisk CSI](https://github.com/kubernetes-sigs/azuredisk-csi-driver)
- [ ] [AzureFile CSI](https://github.com/kubernetes-sigs/azurefile-csi-driver)
- [ ] [OpenStack CCM](https://github.com/kubernetes/cloud-provider-openstack)
- [ ] [OpenStack CSI](https://github.com/kubernetes/cloud-provider-openstack)
- [ ] [vSphere CCM](https://github.com/kubernetes/cloud-provider-vsphere)
- [ ] [vSphere CSI](https://github.com/kubernetes-sigs/vsphere-csi-driver)
- [ ] [Cluster Autoscaler](https://github.com/kubernetes/autoscaler)
