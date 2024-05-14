# Kubermatic KubeOne

<p align="center">
  <img src="docs/img/kubeone-logo-text.png#gh-light-mode-only" width="700px" />
  <img src="docs/img/kubeone-logo-text-dark.png#gh-dark-mode-only" width="700px" />
</p>

[![KubeOne Report Card](https://goreportcard.com/badge/github.com/kubermatic/kubeone)](https://goreportcard.com/report/github.com/kubermatic/kubeone)

Kubermatic KubeOne automates cluster operations on all your cloud, on-prem,
edge, and IoT environments. KubeOne can install high-available (HA) master
clusters as well single master clusters.

## Getting Started

All user documentation for the latest stable version is available at the
[KubeOne docs website][docs].

Information about the support policy (natively-supported providers, supported
Kubernetes versions, and supported operating systems) can be found in the
[Compatibility document][docs-compatibility].

For a quick start, you should check the following documents:

* [Architecture][docs-concepts] to get familiar with the concepts of KubeOne.
* [Getting KubeOne][docs-getting-kubeone] to download and install KubeOne locally.
* [Provisioning][docs-provisioning] to provision the Kubernetes cluster.

## Installing KubeOne

The fastest way to install KubeOne is to use the installation script:

```bash
curl -sfL get.kubeone.io | sh
```

The installation script downloads the release archive from GitHub, installs the
KubeOne binary in your `/usr/local/bin` directory, and unpacks the example
Terraform configs, addons, and helper scripts in your current working
directory.

For other installation methods, check the
[Getting KubeOne guide][docs-getting-kubeone] on our documentation website.

## Features

### Easily Deploy Your Highly Available Cluster On Any Infrastructure

KubeOne works on any infrastructure out of the box. All you need to do is to
provision the infrastructure and let KubeOne know about it. KubeOne will take
care of setting up a production ready Highly Available cluster!

### Native Support For The Most Popular Providers

KubeOne natively supports the most popular providers, including AWS, Azure,
DigitalOcean, GCP, Hetzner Cloud, Nutanix, OpenStack, VMware Cloud Director, and VMware vSphere.
The natively supported providers enjoy additional features such as integration
with Terraform and Kubermatic machine-controller.

### Kubernetes Conformance Certified

KubeOne is a Kubernetes Conformance Certified installer with support for
all [upstream-supported][upstream-supported-versions] Kubernetes versions.

### Declarative Cluster Definition

Define all your clusters declaratively, in a form of a YAML manifest.
You describe what features you want and KubeOne takes care of setting them up.

### Integration With Terraform

The built-in integration with Terraform, allows you to easily provision your
infrastructure using Terraform and let KubeOne take all the needed information
from the Terraform state.

### Integration With Cluster-API, Kubermatic machine-controller, and operating-system-manager

Manage your worker nodes declaratively by utilizing the [Cluster-API][cluster-api]
and [Kubermatic machine-controller][machine-controller]. Create, remove,
upgrade, or scale your worker nodes using kubectl.
[Kubermatic operating-system-manager][operating-system-manager] is responsibile for
managing user-data for worker machines in the cluster.

## Getting Involved

We very appreciate contributions! If you want to contribute or have an idea for
a new feature or improvement, please check out our
[contributing guide][contributing-guide].

If you want to get in touch with us and discuss about improvements and new
features, please create a new issue on GitHub or connect with us over Slack:

* [`#kubeone` channel][k8s-slack-kubeone] on [Kubernetes Slack][k8s-slack]

## Reporting Bugs

If you encounter issues, please [create a new issue on GitHub][github-issue] or
talk to us on the [`#kubeone` Slack channel][k8s-slack-kubeone]. When reporting
a bug please include the following information:

* KubeOne version or Git commit that you're running (`kubeone version`),
* description of the bug and logs from the relevant `kubeone` command (if
  applicable),
* steps to reproduce the issue,
* expected behavior

If you're reporting a security vulnerability, please follow
[the process for reporting security issues][security-vulnerability].

## Changelog

See [the list of releases][changelog] to find out about feature changes.

[upstream-supported-versions]: https://kubernetes.io/docs/setup/release/version-skew-policy/#supported-versions
[cluster-api]: https://github.com/kubernetes-sigs/cluster-api
[machine-controller]: https://github.com/kubermatic/machine-controller
[operating-system-manager]: https://github.com/kubermatic/operating-system-manager
[docs]: https://docs.kubermatic.com/kubeone/
[docs-architecture]: https://docs.kubermatic.com/kubeone/v1.8/architecture/
[docs-concepts]: https://docs.kubermatic.com/kubeone/v1.8/architecture/concepts/
[docs-compatibility]: https://docs.kubermatic.com/kubeone/v1.8/architecture/compatibility/
[docs-getting-kubeone]: https://docs.kubermatic.com/kubeone/v1.8/getting-kubeone/
[docs-provisioning]: https://docs.kubermatic.com/kubeone/v1.8/tutorials/creating-clusters/
[contributing-guide]: https://github.com/kubermatic/kubeone/blob/main/CONTRIBUTING.md
[k8s-slack-kubeone]: https://kubernetes.slack.com/messages/CNEV2UMT7
[k8s-slack]: http://slack.k8s.io/
[github-issue]: https://github.com/kubermatic/kubeone/issues
[security-vulnerability]: https://github.com/kubermatic/kubeone/blob/main/CONTRIBUTING.md#reporting-a-security-vulnerability
[changelog]: https://github.com/kubermatic/kubeone/releases
