# KubeOne

<!--[![GoDoc](https://godoc.org/github.com/kubermatic/kubeone?status.svg)](https://godoc.org/github.com/kubermatic/kubeone) [![Go Report Card](https://goreportcard.com/badge/github.com/kubermatic/kubeone)](https://goreportcard.com/report/github.com/kubermatic/kubeone)-->

`kubeone` is a CLI tool and a Go library for installing, managing, and upgrading Kubernetes High-Available (HA) clusters. It can be used on any cloud provider, on-perm or bare-metal cluster.

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
* Integrates with Cluster-API and [Kubermatic machine-controller](https://github.com/kubermatic/machine-controller) to manage worker nodes
* Integrates with Terraform for sourcing data about infrastructure and control plane nodes
* Officially supports AWS, DigitalOcean, Hetzner and OpenStack

## Installing KubeOne

The easiest way to install KubeOne is using `go get`:
```bash
go get -u github.com/kubermatic/kubeone
```
However, running of the master branch introduces potential risks as the project is currently in the alpha phase and backwards incompatible changes can be expected.

Alternatively, you can obtain KubeOne via [GitHub Releases](https://github.com/kubermatic/kubeone/releases):
```bash
curl -LO https://github.com/kubermatic/kubeone/releases/download/v0.3.0/kubeone_0.3.0_linux_amd64.zip
unzip kubeone_0.3.0_linux_amd64.zip
sudo mv kubeone /usr/local/bin
```

If you already have KubeOne repository cloned, you can use Makefile to install KubeOne.
```bash
make install
```

## Getting Started

We have a getting started tutorial for each cloud provider we support in our [documentation](./docs).
For example, the following document shows [how to get started with KubeOne on AWS](./docs/quickstart-aws.md).

A cluster is created using the `kubeone install` command. It takes a KubeOne configuration file and
optionally Terraform state used to source information about the infrastructure.
```bash
kubeone install config.yaml --tfjson tf.json
```
To learn more about KubeOne configuration, check out [the example configuration file](./config.yaml.dist).

For advanced use cases and other features, check the [KubeOne features]() document.

## Getting Involved

We very appreciate contributions! If you want to get in touch with us and discuss about improvements and new features, please create a new issue on GitHub. You can contact us also via the general Loodse email list and Slack channel:
- Email: [loodse-dev](https://groups.google.com/forum/#!forum/loodse-dev)
- Slack: #[Slack](http://slack.kubermatic.io/) on Slack

### Troubleshooting

If you encounter issues [file an issue][1] or talk to us on the [#KubeOne channel][12] on the Loodse Slack server. Please include the following information:

* KubeOne version or Git commit that you're running (`kubeone version`),
* description of the bug and logs from the relevant `kubeone` command (if applicable),
* steps to reproduce the issue,
* expected behavior

## Contributing

Thanks for taking the time to join our community and start contributing!

Feedback and discussion are available on [the mailing list][11].

### Before you start

* Please familiarize yourself with the [Code of Conduct][4] before contributing.
* See [CONTRIBUTING.md][2] for instructions on the developer certificate of origin that we require.

### Proposing a New Feature

To propose a new feature, please [create a new issue](https://github.com/kubermatic/kubeone/issues/new) and include details about what do you expect from the feature and potential use cases. If the feature is approved by the project maintainers, we'd love help coding it! You can go ahead a create a Work-in-Progress (**WIP**) PR and start coding! In [the contributing guidelines]() you can find information about practices we're following, so make sure to check it out.

For upcoming features please check our [issue trakcer](https://github.com/kubermatic/kubeone/issues) and [milestones](https://github.com/kubermatic/kubeone/milestones). We use milestones as a way to track what features will be added in the upcoming releases.


### Pull requests

* We welcome pull requests. Feel free to dig through the [issues][1] and jump in.

## Changelog

See [the list of releases][3] to find out about feature changes.

[1]: https://github.com/kubermatic/KubeOne/issues
[2]: https://github.com/kubermatic/KubeOne/blob/master/CONTRIBUTING.md
[3]: https://github.com/kubermatic/KubeOne/releases
[4]: https://github.com/kubermatic/KubeOne/blob/master/CODE_OF_CONDUCT.md

[11]: https://groups.google.com/forum/#!forum/projectKubeOne
[12]: https://kubermatic.slack.com/messages/KubeOne

[21]: https://kubermatic.github.io/KubeOne/
