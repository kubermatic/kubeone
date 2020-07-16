# KubeOne Release Process

This document explains how to release a new version of KubeOne.

## Prerequisites

Please follow these steps to prepare your environment for releasing
a new version of KubeOne.

* Make sure your Go version is up-to-date
  * On your computer run `go version` to find out what
  Go version you're using
  * If you're not using the latest Go version, go to
  [the official Go website][go] to grab the latest binaries
* Install [GoReleaser][goreleaser]
* [Generate GitHub Token][github-token] with the `repo` permissions
and export the token as the `GITHUB_TOKEN` environment variable
  * The GitHub Token is needed to create a GitHub Release and
  upload KubeOne binaries

## Preparing the release

Before releasing a new version, you need to update the documentation
to reference to the upcoming version and generate a changelog.

### Updating documentation

You need to update the following documents to point to the new release:

* The Kubernetes versions compatibility matrix should be updated if there
  are changes in supported Kubernetes versions, Terraform versions and/or
  providers. The compatibility matrix is located in two places, in the
  repo [`README.md` file][matrix-readme] and at the 
  [docs website][matrix-docs]
* Update [the quickstart guides][quickstart] to require/recommend
  the latest version if needed

### Generating the changelog

Add a new entry for the upcoming release to the [CHANGELOG.md][changelog] file.

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

### Preparing the release

Before pushing a new Git tag, ensure your `master` branch is up-to-date:

```
git checkout master
git fetch origin
git reset --hard origin/master
```

**Warning:** Before cutting the release, it's **strictly** required to reset
the `examples` directory to prevent secrets from leaking and being released.
GoReleaser ships all content from the `examples` directory including Terraform
state files and credentials if present.

The `examples` directory can be reset by moving the existing directory
**outside** of the repository and then resetting the branch:

```
mv examples ~/kubeone-examples
git reset --hard origin/master
```

After the release process is done, you can move back the old `examples`
directory.

### Creating a branch (only for RC and stable releases)

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

### Pushing a new tag

Create a tag for the new version and push it:

```
git tag -a v1.x.y -m "KubeOne version 1.x.y"
git push origin --tags
```

Now that we have a Git tag, we can proceed to releasing KubeOne binaries.

### Releasing binaries

We're using [GoReleaser][goreleaser] to build and release binaries.

Next, create a release notes file somewhere outside of the repo and
fill it with the changelog for the release. The release notes file will be
used to load custom release notes instead of using the list of commits.

The release notes should look like:

```
Check out the [documentation](https://docs.kubermatic.com/kubeone/v1.x/) for this release to find out how to get started with KubeOne.

<changelog>

### Checksums

SHA256 checksums can be found in the `kubeone_1.x.y_checksums.txt` file.
```

Once this is done, create a snapshot of the release and ensure that everything
is in the place as intended. Pay attention that there are no any leftover files,
especially in the `examples` directory.

```
goreleaser release --rm-dist --release-notes=$HOME/notes.md --snapshot
```

If you're sure that everything is as intended, you can proceed to releasing
a new version:

```
goreleaser release --rm-dist --release-notes=$HOME/notes.md
```

This command builds KubeOne, creates the archive, and uploads it to GitHub.
It's recommended to try to download to the release from GitHub after it's
available and compare the checksums, as well as, confirm that `kubeone version`
shows the correct version.

[go]: https://golang.org/dl/
[goreleaser]: https://goreleaser.com
[goreleaser-install]: https://goreleaser.com/install/
[github-token]: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line
[matrix-readme]: https://github.com/kubermatic/kubeone#kubernetes-versions-compatibility
[matrix-docs]: https://docs.kubermatic.com/kubeone/master/#kubernetes-versions-compatibility
[quickstart]: https://docs.kubermatic.com/kubeone/master/getting_started/
[changelog]: https://github.com/kubermatic/kubeone/blob/master/CHANGELOG.md
