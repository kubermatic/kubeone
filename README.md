# KubeOne

<p align="center"><img src="docs/img/kubeone-logo-text.png" width="700px" /></p>
<p align="center">
  <a href="https://godoc.org/github.com/kubermatic/kubeone">
    <img src="https://godoc.org/github.com/kubermatic/kubeone?status.svg" alt="GoDoc" />
  </a>
  <a href="https://goreportcard.com/report/github.com/kubermatic/kubeone">
    <img src="https://goreportcard.com/badge/github.com/kubermatic/kubeone" alt="Go Report Card" />
  </a>
  <a href="https://bestpractices.coreinfrastructure.org/projects/2934"><img src="https://bestpractices.coreinfrastructure.org/projects/2934/badge">
  </a>
</p>

`kubeone` is a CLI tool and a Go library for installing, managing, and upgrading
Kubernetes High-Available (HA) clusters. It can be used on any cloud provider,
on-prem or bare-metal cluster.

## Project Status

As of v0.6.0, KubeOne is in the beta phase. Check out the
[Backwards Compatibility Policy][6] for more details on
backwards compatibility, KubeOne versioning, and maturity of each KubeOne
component.

Versions earlier than v0.6.0 are considered alpha and it's strongly advised to
upgrade to the v0.6.0 or newer as soon as possible.

## KubeOne in Action

[![KubeOne Demo asciicast](https://asciinema.org/a/244104.svg)](https://asciinema.org/a/244104)

## Features

* Supports Kubernetes 1.13+ High-Available (HA) clusters
* Uses `kubeadm` to provision clusters
* Comes with a straightforward and easy to use CLI
* Choice of Linux distributions between Ubuntu, CentOS and CoreOS
* Integrates with [Cluster-API][7] and [Kubermatic machine-controller][8] to
  manage worker nodes
* Integrates with Terraform for sourcing data about infrastructure and control
  plane nodes
* Officially supports AWS, DigitalOcean, GCE, Hetzner, Packet, OpenStack, VMware
  vSphere and Azure

## Installing KubeOne

### Downloading a binary from GitHub Releases

The recommended way to obtain KubeOne is to grab the
binary from the [GitHub Releases][9] page. On the
releases page, find the binary for your operating system
and architecture and download it or grab the URL and use
`wget` or `curl` to download the binary.

```bash
curl -LO https://github.com/kubermatic/kubeone/releases/download/v<version>/kubeone_<version>_linux_amd64.zip
```

Extract the binary. On Linux and macOS, you can use `unzip`.

```bash
unzip kubeone_0.8.0_linux_amd64.zip
```

Move the `kubeone` binary to your path, so you can easily
invoke it from your terminal.

```bash
sudo mv kubeone /usr/local/bin
```

### Building KubeOne

The alternative way to install KubeOne is using `go get`.

```bash
go get -u github.com/kubermatic/kubeone
```

While running of the master branch is a great way to peak at and test
the new features before they are released, note that master branch can
break at any time or may contain bugs. Official releases are considered
stable and recommended for the production usage.

If you already have KubeOne repository cloned, you can use `make`
to install it.

```bash
make install
```

### Using package managers

Support for packages managers is still work in progress and expected
to be finished for one of the upcoming release. For details about the
progress follow the [issue #471][17]

#### Arch Linux

We have a package in the AUR [here](https://aur.archlinux.org/packages/kubeone).
Use your favorite method to build it on your system, for example by using
`aurutils`:
```bash
aur sync kubeone && pacman -S kubeone
```

### Shell completion and generating documentation

KubeOne comes with commands for generating scripts for the shell
completion and for the documentation in format of man pages
and more.

To activate completions for `bash` (or `zsh`), run or put this command
into your `.bashrc` file:

```bash
. <(kubeone completion bash)
```

To put changes in the effect, source your `.bashrc` file.

```bash
source ~/.bashrc
```

To generate documentation (man pages for example, more available), run:

```bash
kubeone document man -o /tmp/man
```

## Kubernetes Versions Compatibility

Each KubeOne version is supposed to support and work with a set of Kubernetes
minor versions. We're targeting to support at least 3 minor Kubernetes versions,
however for early KubeOne releases we're supporting only one or two minor
versions.

New KubeOne release will be done for each minor Kubernetes version. Usually, a
new release is targeted 2-3 weeks after Kubernetes release, depending on number
of changes needed to support a new version.

Since some terraform releases introduces incompatibilities to previuos versions,
only a specific version range is supported with each KubeOne release.

In the following table you can find what are supported Kubernetes and Terraform
versions for each KubeOne version. KubeOne versions that are crossed out are not
supported. It's highly recommended to use the latest version whenever possible.

| KubeOne version | 1.15 | 1.14 | 1.13 | Terraform | Supported providers                                                |
|-----------------|------|------|------|-----------|--------------------------------------------------------------------|
| v0.9.0+         | +    | +    | +    | v0.12+    | AWS, DigitalOcean, GCE, Hetzner, Packet, OpenStack, vSphere, Azure |
| v0.8.0+         | -    | +    | +    | v0.11     | AWS, DigitalOcean, GCE, Hetzner, Packet, OpenStack, vSphere        |
| v0.6.0+         | -    | +    | +    | v0.11     | AWS, DigitalOcean, GCE, Hetzner, Packet, OpenStack                 |
| v0.5.0          | -    | +    | +    | v0.11     | AWS, DigitalOcean, GCE, Hetzner, OpenStack                         |

## Getting Started

We have a getting started tutorial for each cloud provider we support in our
[documentation][10]. For example, the following document shows
[how to get started with KubeOne on AWS][11].

A cluster is created using the `kubeone install` command. It takes a KubeOne
configuration file and optionally Terraform state used to source information
about the infrastructure.

```bash
kubeone install config.yaml --tfjson tf.json
```

To learn more about KubeOne configuration, please run `kubeone config print --full`.

For advanced use cases and other features, check the [KubeOne features][13]
document.

## Getting Involved

We very appreciate contributions! If you want to contribute or have an idea for
a new feature or improvement, please check out our [contributing guide][2].

If you want to get in touch with us and discuss about improvements and new
features, please create a new issue on GitHub or connect with us over the
mailing list or Slack:

* [loodse-dev mailing list][14]
* [Kubermatic Slack][15]

## Reporting Bugs

If you encounter issues, please [create a new issue on GitHub][1] or talk to us
on the [#KubeOne Slack channel][5]. When reporting a bug please include the
following information:

* KubeOne version or Git commit that you're running (`kubeone version`),
* description of the bug and logs from the relevant `kubeone` command (if
  applicable),
* steps to reproduce the issue,
* expected behavior

If you're reporting a security vulnerability, please follow
[the process for reporting security issues][16].

## Changelog

See [the list of releases][3] to find out about feature changes.

[1]: https://github.com/kubermatic/KubeOne/issues
[2]: https://github.com/kubermatic/KubeOne/blob/master/CONTRIBUTING.md
[3]: https://github.com/kubermatic/KubeOne/releases
[4]: https://github.com/kubermatic/KubeOne/blob/master/CODE_OF_CONDUCT.md
[5]: https://kubermatic.slack.com/messages/KubeOne
[6]: ./docs/backwards_compatibility_policy.md
[7]: https://github.com/kubernetes-sigs/cluster-api
[8]: https://github.com/kubermatic/machine-controller
[9]: https://github.com/kubermatic/kubeone/releases
[10]: ./docs
[11]: ./docs/quickstart-aws.md
[13]: https://github.com/kubermatic/kubeone#features
[14]: https://groups.google.com/forum/#!forum/loodse-dev
[15]: http://slack.kubermatic.io/
[16]: https://github.com/kubermatic/kubeone/blob/master/CONTRIBUTING.md#reporting-a-security-vulnerability
[17]: https://github.com/kubermatic/kubeone/issues/471
