# KubeOne Release Process

This document explains how to release a new version of KubeOne.

## Preparing the release

Before releasing a new version, you need to update the documentation
to reference to the upcoming version and generate a changelog.

### Updating documentation

The [Compatibility document][docs-compatibility] should be updated if there
are changes in supported Kubernetes versions, Terraform versions, operating
systems, and/or providers.

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
https://github.com/kubermatic/kubeone/compare/<commit>...main
```

## Releasing KubeOne

### Preparing the release

Before pushing a new Git tag, ensure your `main` branch is up-to-date.
If you're pushing to a release branch, switch to the appropriate branch and
make sure it's up-to-date.

```
git checkout main
git fetch origin
git reset --hard origin/main
```

### Creating a branch (only for RC and stable releases)

In case you are releasing a stable or RC release, create a release
branch and push it:

```
git checkout -b release/v0.x
git push origin release/v0.x
```

The alpha and beta releases are cut directly from the main branch,
without creating the release branch. This ensures we don't have to
cherry-pick each PR to the release branch.

Once the branch is created, make sure to update branch-protector to add a new
protection rule for the newly-created release branch.

### Creating a new release (tagging the release)

Create a tag for the new version and push it:

```
git tag -a v1.x.y -m "KubeOne version 1.x.y"
git push origin --tags
```

The binaries are built and uploaded by CI once a new tag is pushed.
Once the binaries are uploaded, update the release note to match the format
we use:

```
<changelog>

### Checksums

SHA256 checksums can be found in the `kubeone_1.x.y_checksums.txt` file.
```

It's recommended to try to download to the release from GitHub after it's
available and compare the checksums, as well as, confirm that `kubeone version`
shows the correct version.

[docs-compatibility]: https://docs.kubermatic.com/kubeone/main/architecture/compatibility/
[changelog]: https://github.com/kubermatic/kubeone/blob/main/CHANGELOG.md
