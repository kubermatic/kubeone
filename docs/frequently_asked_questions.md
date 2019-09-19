# Frequently Asked Questions

This document lists some commonly asked questions about KubeOne, what it does,
and how it works. If you have any question not covered here, please create 
[a new GitHub issue][1] or contact us on [the mailing list][2] or 
on the [`#kubeone` channel on Kubernetes Slack][3].

- [What is KubeOne?](#what-is-kubeone-)
- [What cloud providers KubeOne does support?](#what-cloud-providers-kubeone-does-support-)
- [Are on-prem and bare metal clusters supported?](#are-on-prem-and-bare-metal-clusters-supported-)
- [Does KubeOne handles the infrastructure and cloud resources?](#does-kubeone-handles-the-infrastructure-and-cloud-resources-)
- [How KubeOne works?](#how-kubeone-works-)
- [How are commands executed on nodes?](#how-are-commands-executed-on-nodes-)
- [Can I deploy other CNI plugin then Canal?](#can-i-deploy-other-cni-plugin-then-canal-)
- [Can I deploy other controller than machine-controller or decide not to deploy and machine-controller?](#can-i-deploy-other-controller-than-machine-controller-or-decide-not-to-deploy-and-machine-controller-)
- [Can I use KubeOne to create Kubernetes clusters older than 1.13?](#can-i-use-kubeone-to-create-kubernetes-clusters-older-than-113-)
- [Can I use KubeOne to upgrade Kubernetes 1.12 or older cluster to 1.13+?](#can-i-use-kubeone-to-upgrade-kubernetes-112-or-older-cluster-to-113--)
- [How many versions can I upgrade at the same time?](#how-many-versions-can-i-upgrade-at-the-same-time-)
- [I'd like to contribute to KubeOne! Where can I start?](#i-d-like-to-contribute-to-kubeone--where-can-i-start-)

## What is KubeOne?

KubeOne is a CLI and a Go library for installing, maintaining and upgrading
Kubernetes 1.13+ clusters.

## What cloud providers KubeOne does support?

KubeOne is supposed to work on any cloud provider, on-prem and bare-metal
cluster, as long as there is no need for additional configuration. However, to
utilize all features of KubeOne, such as Terraform integration and creating
worker nodes, KubeOne and [Kubermatic machine-controller][4] need to support
that provider.

Currently we support AWS, DigitalOcean, Google Compute Engine (GCE), Hetzner,
Packet, OpenStack, VMware vSphere and Azure.

## Are on-prem and bare metal clusters supported?

Yes. We officially support VMware vSphere and OpenStack.

## Does KubeOne handles the infrastructure and cloud resources?

No, it's up to the operator to setup the needed infrastructure and provide the
needed parameters to KubeOne. To make this task easier, we provide
[Terraform scripts][5] that operators can use to create the needed
infrastructure. Using the Terraform integration operator can source information
about the infrastructure from the Terraform output.

This decision was based on that fact that we didn't want to limit operators how
the infrastructure can be configured and what resources can be used. There are
many possible setups and supporting each of them in most of cases isn't
possible. Operators are free to define infrastructure how they prefer and then
use KubeOne to provision the cluster. We're open to feedback, so if you have any
suggestion or idea ping us on [the mailing list][6] or on the 
[`#kubeone` channel on Kubernetes Slack][3].

## How KubeOne works?

KubeOne uses [kubeadm][7] to provision and upgrade Kubernetes control plane
nodes. The worker nodes are provisioned and managed using the
[Kubermatic machine-controller][4]. KubeOne takes care of installing Docker and
all needed dependencies for Kubernetes and kubeadm. After the cluster is
provisioned, KubeOne deploys the [Canal CNI plugin][8] and
[Kubermatic machine-controller][4].

## How are commands executed on nodes?

All commands are executed over SSH. Because we don't take care of the
infrastructure, it makes almost impossible to use `cloud-config`. Addons (Canal
and `machine-controller`) are deployed using Go and the
[controller-runtime client][9] library.

## Can I deploy other CNI plugin then Canal?

This is currently not possible, however we're
[researching about switching to WeaveNet][10] or providing option to choose the
CNI plugin.

## Can I deploy other controller than machine-controller or decide not to deploy and machine-controller?

You can opt out deploying machine-controller by setting
`machine_controller.Deploy` to `false`. In that case you can't deploy worker
nodes using KubeOne.

## Can I use KubeOne to create Kubernetes clusters older than 1.13?

No. Due to breaking changes and the new features available only in kubeadm
1.13+, we decided to support only Kubernetes 1.13+ clusters.

## Can I use KubeOne to upgrade Kubernetes 1.12 or older cluster to 1.13+?

No. Due to breaking changes in the upgrade process, we only support upgrading
from Kubernetes 1.13 to newer.

## How many versions can I upgrade at the same time?

It is only possible to upgrade from one minor to the next minor version (`n+1`).
For example, if you want to upgrade from Kubernetes 1.13 to Kubernetes 1.15,
you'd need to upgrade to 1.14 and then to 1.15.

## I'd like to contribute to KubeOne! Where can I start?

Please check our [contributing guide](CONTRIBUTING.md).

[1]: https://github.com/kubermatic/kubeone/issues
[2]: https://groups.google.com/forum/#!forum/loodse-dev
[3]: http://slack.k8s.io/
[4]: https://github.com/kubermatic/machine-controller
[5]: http://github.com/kubermatic/kubeone/tree/master/examples/terraform
[6]: https://groups.google.com/forum/#!forum/loodse-dev
[7]: https://github.com/kubernetes/kubeadm
[8]: https://github.com/projectcalico/canal
[9]: https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client
[10]: https://github.com/kubermatic/kubeone/issues/256
