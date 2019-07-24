# KubeOne Release Process

This document explains how to release a new version of KubeOne.

## Prerequisites

Please follow these steps to prepare your environment for releasing
a new version of KubeOne.

* Make sure your Go version is up-to-date
  * On your computer run `go version` to find out what
  Go version you're using
  * If you're not using the latest Go version, go to
  [the official Go website][1] to grab the latest binaries
* Install [GoReleaser][2]
* [Generate GitHub Token][3] with the `repo` permissions
and export the token as the `GITHUB_TOKEN` environment variable
  * The GitHub Token is needed to create a GitHub Release and
  upload KubeOne binaries

## Preparing the release

Before releasing a new version, you need to update the documentation
to reference to the upcoming version and generate a changelog.

### Updating documentation

You need to update the following documents to point to the new release:

* The [Kubernetes versions compatibility][5] section of the
`README.md` file should be updated if there are changes in supported
Kubernetes versions, Terraform versions and/or providers
* [Optional] Update [the quickstart guides][6] to require/recommend
the latest version

### Generating the changelog

Add a new entry for the upcoming release to the [CHANGELOG.md][7] file.

The changelog file only lists changes that are directly affecting
the end-users, such as a new feature, a bug or a security fix, a new
version of a component and more. Changes such as changes to tests and
refactors (as long as there are no behavior changes) are not documented.

**Note:** Currently, we're generating the changelog manually.
To find what changes have been made since a previous release
you can use the GitHub Compare feature. Grab the latest
commit from the previous release and put it in the
following link:
```
https://github.com/kubermatic/kubeone/compare/<commit>...master
```

## Releasing KubeOne

### Pushing a Git tag

Before pushing a new Git tag, ensure your `master` branch is up-to-date:

```
git checkout master
git fetch origin
git reset --hard origin/master
```

In case you are releasing a stable or RC release, create a release
branch and push it:

```
git checkout -b release/v0.x
git push origin release/v0.x
```

The alpha and beta releases are cut directly from the master branch,
without creating the release branch. This ensures we don't have to
cherry-pick each PR to the release branch.

**Note:** Currently there's a bug with branch-protector that
requires a new protection rule to be created for each release branch.

Create a tag for the new version and push it:

```
git tag -a v0.x.y -m "KubeOne version 0.x.y"
git push origin --tags
```

Now that we have a Git tag, we can proceed to releasing KubeOne binaries.

### Releasing binaries

We're using [GoReleaser][8] to build and release binaries.

As of v0.10.0-alpha.0, we're shipping the `examples` directory along with
the binary. It is **strictly** required to reset the `examples` directory
to prevent secrets from leaking and being released. You can do that such as:

```
mv examples ~/kubeone-examples
git reset --hard origin/master
```

After the release process is done, you can move back the old `examples` directory
back.

Next, create a release notes file somewhere outside of the repo and
fill it with the changelog for the release. The release notes file will be
used to load custom release notes instead of using the list of commits.

The release notes should look like:

```
Check out the [documentation](https://github.com/kubermatic/kubeone/tree/v0.x.y/docs) for this release to find out how to get started with KubeOne.

<changelog>

### Checksums

SHA256 checksums can be found in the `kubeone_0.x.y_checksums.txt` file.
```

Once this is done, create a snapshot of the release and ensure that everything
is in the place as intended. Pay attention that there are no any leftover files,
especially in the `examples` directory.

```
goreleaser release --rm-dist --release-notes=~/notes.md --snapshot
```

If you're sure that everything is as intended, you can proceed to releasing
a new version:

```
goreleaser release --rm-dist --release-notes=~/notes.md
```

[1]: https://golang.org/dl/
[2]: https://goreleaser.com/install/
[3]: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line
[4]: https://github.com/kubermatic/kubeone#installing-kubeone
[5]: https://github.com/kubermatic/kubeone#kubernetes-versions-compatibility
[6]: https://github.com/kubermatic/kubeone/tree/master/docs
[7]: https://github.com/kubermatic/kubeone/blob/master/CHANGELOG.md
[8]: https://goreleaser.com
