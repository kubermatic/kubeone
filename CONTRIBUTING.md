# KubeOne Contributing Guide <!-- omit in toc -->

This documents explains how to contribute to KubeOne, how to find what to work on, and what practices are we following along the way. Kubermatic projects are [Apache 2.0 licensed](LICENSE) and accept contributions via GitHub Pull Requests.

## Table of Contents <!-- omit in toc -->

- [Getting Started](#getting-started)
  - [Setting Up The Environment](#setting-up-the-environment)
  - [Code of Conduct](#code-of-conduct)
  - [Project Structure and Frequently Asked Questions](#project-structure-and-frequently-asked-questions)
  - [Certificate of Origin](#certificate-of-origin)
- [Contributing Process](#contributing-process)
  - [Finding What To Work On](#finding-what-to-work-on)
  - [Submitting Feature Requests](#submitting-feature-requests)
  - [Creating a New Pull Request](#creating-a-new-pull-request)
  - [Prow](#prow)
  - [Release Notes Block](#release-notes-block)
- [Code Style Guide](#code-style-guide)
  - [Linting](#linting)
  - [Guidelines](#guidelines)
  - [Import Order](#import-order)
- [Contact](#contact)
  - [Reporting a Security Vulnerability](#reporting-a-security-vulnerability)

## Getting Started

This part of the document explains what you need to do in order to get started with contributing to KubeOne.

### Setting Up The Environment

Before you start contributing, you need to fork the KubeOne repository. To make KubeOne work properly, you should clone the repository in the `$(go env GOPATH)/src/k8c.io/kubeone` directory and then set up the repository to sync with your fork.

### Code of Conduct

Please make sure to read our [Code of Conduct](code-of-conduct.md).

### Project Structure and Frequently Asked Questions

Check out [the project structure document](docs/project_structure.md) to learn more about how the project is structured and what are responsibilities of the each package. We have a [frequently asked questions document](docs/frequently_asked_questions.md) which explains how KubeOne works and some of decisions we made.

### Certificate of Origin

By contributing to this project you agree to the Developer Certificate of Origin (DCO).
This document was created by the Linux Kernel community and is a simple statement that you, as a contributor, have the legal right to make the contribution. See the [DCO](DCO) file for details.

Any copyright notices in this repository should specify the authors as "The KubeOne Authors".

To sign your work, just add a line like this at the end of your commit message:

```
Signed-off-by: Joe Example <joe@example.com>
```

This can easily be done with the `--signoff` option to `git commit`.

Note that we're requiring all commits in a PR to be signed-off. If you already created a PR, you can sign-off all existing commits by rebasing with the `--signoff` flag.

```
git rebase --signoff origin/main
```

By doing this you state that you can certify the following (from https://developercertificate.org/).

## Contributing Process

This part of the document explains how to get started with contributing, including finding what to work on and how to submit a PR.

### Finding What To Work On

One of the most challenging part of contributing to the open source projects is finding what to work on. To make it easier, we're trying to label all issues that may be good for first time contributors using the [`good first issue`](https://github.com/kubermatic/kubeone/labels/good%20first%20issue) label. Another label you can take a look at is the [`help wanted`](https://github.com/kubermatic/kubeone/labels/help%20wanted) label, however there may be issue that can be harder to solve if you're not experienced with the code base.

If you have any questions or need assistance, please feel free to comment on the appropriate issue, create a new issue or ping us on [`#kubeone` channel on Kubernetes Slack](http://slack.k8s.io/) or [Kubermatic forums](http://forum.kubermatic.com/)!

### Submitting Feature Requests

If you have an idea for a new feature or improvement, we'd love to hear it! Before getting started working on the new feature, please [create a new issue](https://github.com/kubermatic/kubeone/issues/new) so we can discuss  and agree how the feature should look like. Include details such as how you expect the feature to look like and what are the potential use cases.

Once we agree on the idea, we'd love you to work on it! You can create a WIP PR, so maintainers and users can give you early feedback and potentially assist you.

### Creating a New Pull Request

When you want to create a pull request, you'll be asked to input the basic information about changes made in that pull request. For example, you can include a quick summary what changes the PR brings and are there any breaking changes.

You'll also be asked to write a release note. Without the release note, the PR will not be merged by the bot. You can find more details about how a release note should look like in the [release note block](#release-notes-block) section of the document.

If you want to annotate PR as a WIP (Work-in-Progress) PR, you can add `WIP` in the PR title. That blocks PR from being accidentally merged until `WIP` is not removed from the PR title.

The CI tests will not start until a maintainer doesn't approve the PR using the `/ok-to-test` command. This ensures that potentially malicious and/or spam PRs can't effect our CI pipeline and the project.

### Prow

We use automation called [Prow](https://github.com/kubernetes/test-infra/tree/master/prow), which takes care of many tasks including running tests, merging PRs and labeling PRs and issues.

### Release Notes Block

Prow expects a release note block to be present in the each PR. Release Notes are used when generating a changelog, however we're currently doing that manually.

Release Note block looks such as:
```
'''release-note
'''
```
Where `'` represents the back quote character.

Release Notes should be targeted at the end users. They should contain a summary of changes that are going to effect the user.

## Code Style Guide

### Linting

In order for CI checks to pass, the code must pass linting checks. We're using [`golangci/golintci-lint`](https://github.com/golangci/golangci-lint) linter. In the repository you can find instructions for installing the linter. The lint tests can be run by using the `make lint` command.

### Guidelines

This is mostly copied from [Kubernetes Code Conventions](https://github.com/kubernetes/community/blob/master/contributors/guide/coding-conventions.md#code-conventions).

* Bash
  * [Google Bash Style guide](https://google.github.io/styleguide/shell.xml)
  * Ensure that build, release, test, and cluster-management scripts run on macOS
* Go
  * [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
  * [Effective Go](https://golang.org/doc/effective_go.html)
  * Know and avoid [Go landmines](https://gist.github.com/lavalamp/4bd23295a9f32706a48f)
  * Comment your code.
  * [Go's commenting conventions](http://blog.golang.org/godoc-documenting-go-code)
    * If reviewers ask questions about why the code is the way it is, that's a sign that comments might be helpful.
  * Command-line flags should use dashes, not underscores
  * Naming
    * Please consider package name when selecting an interface name, and avoid redundancy.
      * e.g.: `storage.Interface` is better than `storage.StorageInterface`.
    * Do not use uppercase characters, underscores, or dashes in package names.
    * Please consider parent directory name when choosing a package name.
      * so `pkg/controllers/autoscaler/foo.go` should say `package autoscaler` not `package autoscalercontroller`.
      * Unless there's a good reason, the `package foo` line should match the name of the directory in which the .go file exists.
      * Importers can use a different name if they need to disambiguate.
    * Locks should be called `lock` and should never be embedded (always `lock sync.Mutex`). When multiple locks are present, give each lock a distinct name following Go conventions - `stateLock`, `mapLock` etc.

### Import Order

We group imports the following way:

* Go SDK
* external packages which do not apply to other rules (Like `github.com/golang/glog`, etc.)
* `github.com/kubermatic/*`
* `k8s.io/*`

```go
import (
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/installer/util"
	"k8c.io/kubeone/pkg/ssh"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)
```

Depending of the number of packages we import from a individual repository, those packages can be grouped as well:

```go
import (
	"errors"
	"fmt"

	kubermaticv1 "github.com/kubermatic/kubermatic/api/pkg/crd/kubermatic/v1"
	kuberneteshelper "github.com/kubermatic/kubermatic/api/pkg/kubernetes"
	"github.com/kubermatic/kubermatic/api/pkg/provider"

	"github.com/golang/glog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)
```

## Contact

The KubeOne project currently uses the [Kubermatic forums](http://forum.kubermatic.com/) and [`#kubeone` Slack channel](https://kubernetes.slack.com/messages/CNEV2UMT7) on [Kubernetes Slack](http://slack.k8s.io/). You can also ask questions by creating a new issue on GitHub, but using mailing list or Slack is preferred.

Please avoid emailing maintainers found in the MAINTAINERS file directly. They are very busy, but they often read the mailing list and the Slack channel.

### Reporting a Security Vulnerability

Due to their public nature, GitHub and mailing lists are not appropriate places for reporting vulnerabilities. If you suspect you have found a security vulnerability in KubeOne, please do not file a GitHub issue, but instead email security@kubermatic.com with the full details, including steps to reproduce the issue.
