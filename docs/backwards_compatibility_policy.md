# Backwards Compatibility Policy

KubeOne is trying to be backwards compatible as much as possible. Users upgrading from the older
to the newer KubeOne minor or patch release can expect behavior to stay the same.

As of KubeOne v0.6.0, KubeOne's overall maturity is beta. However, some components are still considered alpha,
but we'll ensure that changes to those components will not affect end-users or that there will be a
migration path.

Versions earlier than v0.6.0 are considered alpha and it's strongly advised to upgrade to the v0.6.0 or newer as soon as possible.

This backwards compatibility policy is in effect as of version 0.6.0.

## Versioning

KubeOne follows the [semantic versioning 2.0.0](https://semver.org/). The minor or patch releases will **not** include
backwards incompatible changes or there will be an automatic migration path. A new major release includes the backwards
incompatible changes and we can't guarantee an automatic migration or migration without user's interaction.

## CLI maturity

The KubeOne CLI maturity is considered to be **beta**.

Commands, command arguments and commands behavior are expected not to be changed without prior notice
of at least 3 releases before the change.

Alpha commands can be introduced at any time. All alpha commands will live under the `alpha` subcommand.

## API and Configuration File maturity

The KubeOneCluster API and KubeOne configuration file are considered to be **alpha**.

Changes to the KubeOneCluster API and KubeOne configuration file will be done by introducing a new API version.
Previous alpha API version will be supported for at least two more release or three months (whichever is longer).

An automatic migration path to the newer API version is not guaranteed while the API is alpha, but we'll
do our best to provide one.

## Cluster Installation/Upgrade Process

The steps used to install and upgrade the cluster are considered to be **beta**.

The backwards incompatible changes in the installation and upgrade process will have a prior notice of
at least 3 releases before the change.

## Library

Using KubeOne as a Go library is considered **alpha** and experimental. Breaking changes to Go structs and functions,
including behavior changes can happen at any time without any prior notice.

If using KubeOne as a Go library it's strongly advised to vendor and pin it to the current commit.
