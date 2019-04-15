# KubeOne

<p align="center"><img src="docs/img/kubeone-logo-text.png" width="700px"></p>
<p align="center">
  <a href="https://godoc.org/github.com/kubermatic/kubeone"><img src="https://godoc.org/github.com/kubermatic/kubeone?status.svg" alt="GoDoc"></img></a>
  <a href="https://goreportcard.com/report/github.com/kubermatic/kubeone"><img src="https://goreportcard.com/badge/github.com/kubermatic/kubeone" alt="Go Report Card"></img></a>
</p>

`kubeone` is a CLI tool and a Go library for installing, managing, and upgrading Kubernetes High-Available (HA) clusters. It can be used on any cloud provider, on-prem or bare-metal cluster.

## Project Status

KubeOne is currently in the alpha phase, so breaking changes can be expected in the upcoming period.
You can find more details about breaking changes and actions needed to migrate them in the [Release Notes](https://github.com/kubermatic/kubeone/releases). In the upcoming weeks we're planning to enter the beta phase and and define a backwards compatibility policy.

## KubeOne in Action

TBD

## Features

* Supports Kubernetes 1.13+ High-Available (HA) clusters
* Uses `kubeadm` to provision clusters
* Comes with a straightforward and easy to use CLI
* Choice of Linux distributions between Ubuntu, CentOS and CoreOS
* Integrates with [Cluster-API](https://github.com/kubernetes-sigs/cluster-api) and [Kubermatic machine-controller](https://github.com/kubermatic/machine-controller) to manage worker nodes
* Integrates with Terraform for sourcing data about infrastructure and control plane nodes
* Officially supports AWS, DigitalOcean, GCE, Hetzner and OpenStack

## Installing KubeOne

The easiest way to install KubeOne is using `go get`:
```bash
go get -u github.com/kubermatic/kubeone
```
However, running of the master branch introduces potential risks as the project is currently in the alpha phase and backwards incompatible changes can be expected.

Alternatively, you can obtain KubeOne via [GitHub Releases](https://github.com/kubermatic/kubeone/releases):
```bash
curl -LO https://github.com/kubermatic/kubeone/releases/download/v0.5.0/kubeone_0.5.0_linux_amd64.zip
unzip kubeone_0.5.0_linux_amd64.zip
sudo mv kubeone /usr/local/bin
```

If you already have KubeOne repository cloned, you can use Makefile to install KubeOne.
```bash
make install
```

## Kubernetes Versions Compatibility

Each KubeOne version is supposed to support and work with a set of Kubernetes minor versions. We're targeting to support at least 3 minor Kubernetes versions, however for early KubeOne releases we're supporting only one or two minor versions.

New KubeOne release will be done for each minor Kubernetes version. Usually, a new release is targeted 2-3 weeks after Kubernetes release, depending on number of changes needed to support a new version.

In the following table you can find what are supported Kubernetes versions for each KubeOne version. KubeOne versions that are crossed out are not supported. It's highly recommended to use the latest version whenever possible.

| KubeOne version | 1.14 | 1.13 | 1.12 | Supported providers |
|-----------------|------|------|------|---------------------|
| v0.5.0 | + | + | - | AWS, DigitalOcean, GCE, Hetzner, OpenStack |
| v0.4.1 | - | + | - | AWS, DigitalOcean, Hetzner, OpenStack |
| <s>v0.4.0<s> | - | + | - | AWS, DigitalOcean, Hetzner, OpenStack |
| <s>v0.3.0<s> | - | + | - | AWS, DigitalOcean, Hetzner, OpenStack |

**Note:** Versions v0.2.0 and v0.1.0-beta0 don't support upgrading to the newer Kubernetes version.

## Getting Started

We have a getting started tutorial for each cloud provider we support in our [documentation](./docs).
For example, the following document shows [how to get started with KubeOne on AWS](./docs/quickstart-aws.md).

A cluster is created using the `kubeone install` command. It takes a KubeOne configuration file and
optionally Terraform state used to source information about the infrastructure.
```bash
kubeone install config.yaml --tfjson tf.json
```
To learn more about KubeOne configuration, check out [the example configuration file](./config.yaml.dist).

For advanced use cases and other features, check the [KubeOne features](https://github.com/kubermatic/kubeone#features) document.

## Getting Involved

We very appreciate contributions! If you want to contribute or have an idea for a new feature or improvement, please check out our [contributing guide][2].

If you want to get in touch with us and discuss about improvements and new features, please create a new issue on GitHub or connect with us over the mailing list or Slack:

* [loodse-dev mailing list](https://groups.google.com/forum/#!forum/loodse-dev)
* [Kubermatic Slack](http://slack.kubermatic.io/)

## Reporting Bugs

If you encounter issues, please [create a new issue on GitHub][1] or talk to us on the [#KubeOne Slack channel][12]. When reporting a bug please include the following information:

* KubeOne version or Git commit that you're running (`kubeone version`),
* description of the bug and logs from the relevant `kubeone` command (if applicable),
* steps to reproduce the issue,
* expected behavior

If you're reporting a security vulnerability, please follow [the process for reporting security issues](https://github.com/kubermatic/kubeone/blob/master/CONTRIBUTING.md#reporting-a-security-vulnerability).

## Changelog

See [the list of releases][3] to find out about feature changes.

[1]: https://github.com/kubermatic/KubeOne/issues
[2]: https://github.com/kubermatic/KubeOne/blob/master/CONTRIBUTING.md
[3]: https://github.com/kubermatic/KubeOne/releases
[4]: https://github.com/kubermatic/KubeOne/blob/master/CODE_OF_CONDUCT.md

[11]: https://groups.google.com/forum/#!forum/projectKubeOne
[12]: https://kubermatic.slack.com/messages/KubeOne

[21]: https://kubermatic.github.io/KubeOne/
