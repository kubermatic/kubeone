# Kubermatic KubeOne

<p align="center"><img src="docs/img/kubeone-logo-text.png" width="700px" /></p>

[![KubeOne Report Card](https://goreportcard.com/badge/github.com/kubermatic/kubeone)](https://goreportcard.com/report/github.com/kubermatic/kubeone)

`kubeone` is a CLI tool and a Go library for installing, managing, and upgrading
Kubernetes clusters. It can be used on on all your cloud, on-prem, edge, and IoT environments.  

KubeOne can install high-available (HA) master cluster as well single master cluster.

## Project Status

As of v0.6.0, KubeOne is in the beta phase. Check out the
[Backwards Compatibility Policy][6] for more details on
backwards compatibility, KubeOne versioning, and maturity of each KubeOne
component.

Versions earlier than v0.6.0 are considered alpha and it's strongly advised to
upgrade to the v0.6.0 or newer as soon as possible.

## Getting Started

All user documentation is available at the [KubeOne docs website](https://docs.kubermatic.com/kubeone/master/).

We have a getting started tutorial for each provider we support in the
[Getting Started section][10]. For example, the following document shows
[how to get started with KubeOne on AWS][11].

For advanced use cases and other features, check out the
[Using KubeOne section][13].

## KubeOne in Action

[![KubeOne Demo asciicast](https://asciinema.org/a/244104.svg)](https://asciinema.org/a/244104)

## Features

* Supports Kubernetes 1.13+ High-Available (HA) clusters
* Uses `kubeadm` to provision clusters
* Comes with a straightforward and easy to use CLI
* Choice of Linux distributions between Ubuntu, CentOS and CoreOS/Flatcar
* Integrates with [Cluster-API][7] and [Kubermatic machine-controller][8] to
  manage worker nodes
* Integrates with Terraform for sourcing data about infrastructure and control
  plane nodes
* Officially supports AWS, DigitalOcean, GCE, Hetzner, Packet, OpenStack, VMware
  vSphere and Azure

## Installing KubeOne

### Downloading a binary from GitHub Releases

The fastest way to get KubeOne:
```bash
curl -sfL get.kubeone.io | sh
```

If you want to have more control over how KubeOne is installed, download the
binary from the [GitHub Releases][9] page. 

On the releases page, you can find the binary for your operating system
and architecture 

Download it or grab the URL and use `wget` or `curl` to download the binary.

Extract the binary to the KubeOne directory. On Linux and macOS, you can use `unzip`.

Move the `kubeone` binary to your path, so you can easily invoke it from your terminal.

```bash
OS=$(uname)
VERSION=$(curl -w '%{url_effective}' -I -L -s -S https://github.com/kubermatic/kubeone/releases/latest -o /dev/null | sed -e 's|.*/v||')
curl -LO "https://github.com/kubermatic/kubeone/releases/download/v${VERSION}/kubeone_${VERSION}_${OS}_amd64.zip"
unzip kubeone_${VERSION}_${OS}_amd64.zip -d kubeone_${VERSION}_${OS}_amd64
sudo mv kubeone_${VERSION}_${OS}_amd64/kubeone /usr/local/bin
```

### Building KubeOne

The alternative way to install KubeOne is using `go get`.

To get latest stable release:
```bash
GO111MODULE=on go get github.com/kubermatic/kubeone
```

To get latest beta release (for example v0.11.0-beta.0 tag):
```bash
GO111MODULE=on go get github.com/kubermatic/kubeone@v0.11.0-beta.0
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

`kubeone` can be found in the official Arch Linux repositories:

[https://www.archlinux.org/packages/community/x86_64/kubeone/](https://www.archlinux.org/packages/community/x86_64/kubeone/)

Install it via:

```bash
pacman -S kubeone
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

Since some Terraform releases introduces incompatibilities to previuos versions,
only a specific version range is supported with each KubeOne release.

In the following table you can find what are supported Kubernetes and Terraform
versions for each KubeOne version. KubeOne versions that are crossed out are not
supported. It's highly recommended to use the latest version whenever possible.

| KubeOne version | 1.18 | 1.17 | 1.16 | 1.15 | 1.14 | Terraform | Supported providers                                                |
| --------------- | ---- | ---- | ---- | ---- | ---- | --------- | ------------------------------------------------------------------ |
| v0.11.0+        | +    | +    | +    | +    | -    | v0.12+    | AWS, DigitalOcean, GCE, Hetzner, Packet, OpenStack, vSphere, Azure |
| v0.10.0+        | -    | -    | +    | +    | +    | v0.12+    | AWS, DigitalOcean, GCE, Hetzner, Packet, OpenStack, vSphere, Azure |

## Getting Involved

We very appreciate contributions! If you want to contribute or have an idea for
a new feature or improvement, please check out our [contributing guide][2].

If you want to get in touch with us and discuss about improvements and new
features, please create a new issue on GitHub or connect with us over the
forums or Slack:

* [`#kubeone` channel][5] on [Kubernetes Slack][15]
* [Kubermatic forums][9]

## Reporting Bugs

If you encounter issues, please [create a new issue on GitHub][1] or talk to us
on the [`#kubeone` Slack channel][5]. When reporting a bug please include the
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
[5]: https://kubernetes.slack.com/messages/CNEV2UMT7
[6]: ./docs/backwards_compatibility_policy.md
[7]: https://github.com/kubernetes-sigs/cluster-api
[8]: https://github.com/kubermatic/machine-controller
[9]: https://github.com/kubermatic/kubeone/releases
[10]: https://docs.kubermatic.com/kubeone/master/getting_started/
[11]: https://docs.kubermatic.com/kubeone/master/getting_started/aws/
[13]: https://docs.kubermatic.com/kubeone/master/using_kubeone/
[14]: https://forum.kubermatic.com/
[15]: http://slack.k8s.io/
[16]: https://github.com/kubermatic/kubeone/blob/master/CONTRIBUTING.md#reporting-a-security-vulnerability
[17]: https://github.com/kubermatic/kubeone/issues/471
