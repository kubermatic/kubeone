# KubeOne Release Process

This document explains how to release a new version of KubeOne.

## Prerequisites

Please follow these steps to prepare your environment for releasing
a new version of KubeOne.

* Make sure your Go version is up-to-date
  * On your computer, run `go version` to find out what
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

The changelog file contains only changes that are directly affecting
the end-users, such as a new feature, a bug or security fix, a new
version of a dependency and more. Changes such as changes to tests,
refactors (as long as there are no behavior changes) are not documented.

**Note:** Currently, we're generating the changelog manually.
To find what changes have been made since a previous release
you can use the GitHub Compare feature. Grab the latest
commit from the previous release and put it in the
following link:
```
https://github.com/kubermatic/kubeone/compare/<commit>...master
```

### Creating a PR

**Warning:** This part is still work in progress!

Once you have finished updating the documentation and the
changelog, go ahead and create a PR with all changes.

We're using this PR to run all tests and ensure that
the upcoming version works correctly on all providers.

Once the PR is created, hold it so it's not merged by
Prow/Tide as soon as the required tests pass:
```
/hold
```

Then, run optional tests that tests KubeOne on providers
other than AWS:
```
TBD
```

Finally, go to the Prow Dashboard to monitor the progress.
Once all tests are green, unhold the PR:
```
/hold cancel
```

At this point, the documentation is updated and we ensured that
KubeOne works correctly on all providers. You're ready
to create a new KubeOne release.

## Releasing KubeOne

Before running the release process, ensure your
`master` branch is up-to-date:

```
git checkout master
git fetch origin
git reset --hard origin/master
```

Create a release branch and push it:

```
git checkout -b release/v0.x
git push origin release/v0.x
```

**Note:** Currently there's a bug with branch-protector that
requires a new protection to be created for each release branch.

Create a tag for the new version and push it:

```
git tag -a v0.x.y -m "KubeOne version 0.x.y"
git push origin --tags
```

Once the tag is in the place, run GoReleaser from the KubeOne
root directory which will create a GitHub Release and generate
and upload binaries:

```
goreleaser
```

As GoReleaser generates the release notes based on the list of
all commits since the last release, we're going to change it
to the changelog we generated earlier.

At the top, add the following links:

```
Check out the [documentation](https://github.com/kubermatic/kubeone/tree/v0.x.y/docs) for this release to find out how to get started with KubeOne.
```

Then, add the changelog and after the changelog add information
about checksums:

```
### Checksums

SHA256 checksums can be found in the `kubeone_0.x.y_checksums.txt` file.
```

[1]: https://golang.org/dl/
[2]: https://goreleaser.com/install/
[3]: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line
[4]: https://github.com/kubermatic/kubeone#installing-kubeone
[5]: https://github.com/kubermatic/kubeone#kubernetes-versions-compatibility
[6]: https://github.com/kubermatic/kubeone/tree/master/docs
[7]: https://github.com/kubermatic/kubeone/blob/master/CHANGELOG.md
