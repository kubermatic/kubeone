# Code style

## Linting

Our existing linting tries to enforce common rules, but cannot cover everything.
Linting can be executed via `make lint`

## Guidelines

*This is mostly copied from
[Kubernetes](https://github.com/kubernetes/community/blob/master/contributors/guide/coding-conventions.md#code-conventions)*

  - Bash
    - https://google.github.io/styleguide/shell.xml
    - Ensure that build, release, test, and cluster-management scripts run on
      macOS
  - Go
    - [Go Code Review
      Comments](https://github.com/golang/go/wiki/CodeReviewComments)
    - [Effective Go](https://golang.org/doc/effective_go.html)
    - Know and avoid [Go
      landmines](https://gist.github.com/lavalamp/4bd23295a9f32706a48f)
    - Comment your code.
      - [Go's commenting
        conventions](http://blog.golang.org/godoc-documenting-go-code)
      - If reviewers ask questions about why the code is the way it is, that's a
        sign that comments might be helpful.
    - Command-line flags should use dashes, not underscores
    - Naming
      - Please consider package name when selecting an interface name, and avoid
        redundancy.
          - e.g.: `storage.Interface` is better than `storage.StorageInterface`.
      - Do not use uppercase characters, underscores, or dashes in package
        names.
      - Please consider parent directory name when choosing a package name.
          - so pkg/controllers/autoscaler/foo.go should say `package autoscaler`
            not `package autoscalercontroller`.
          - Unless there's a good reason, the `package foo` line should match
            the name of the directory in which the .go file exists.
          - Importers can use a different name if they need to disambiguate.
      - Locks should be called `lock` and should never be embedded (always `lock
        sync.Mutex`). When multiple locks are present, give each lock a distinct
        name following Go conventions - `stateLock`, `mapLock` etc.

## Special cases

### Imports

We group imports the following way:
- Go SDK
- external packages which do not apply to other rules (Like
  `github.com/golang/glog`, etc.)
- github.com/kubermatic
- k8s.io/*

For example:
```go
package main

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

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)
```

Depending of the number of packages we import from a individual repository,
those packages can be grouped as well.

Like:
```go
package aws

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
