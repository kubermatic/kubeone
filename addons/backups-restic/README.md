# Backups Addon

The backups addon ([`backups-restic.yaml`][backups-addon]) can be used to backup
the most important parts of a cluster, including:
* `etcd`
* `etcd` PKI (certificates and keys used by Kubernetes to access the `etcd` cluster)
* Kubernetes PKI (certificates and keys used by Kubernetes and clients)

The addon uses [Restic][restic] to upload backups, encrypt them, and handle backup
rotation. By default, backups are done every 30 minutes and are kept for 48 hours.

## Prerequisites

In order to use this addon, you need an S3 bucket or Restic-compatible repository for
storing backups.

## Using The Addon

You need to replace the following values with the actual ones:
* `<<RESTIC_PASSWORD>>` - a password used to encrypt the backups
* `<<S3_BUCKET>>` - the restic-style path of the repository to be used for backups (e.g. `s3:s3.amazonaws.com/<backup-bucket-name>`)
* `<<AWS_DEFAULT_REGION>>` - default AWS region

Credentials are fetched automatically if you are deploying on AWS. If you want to use
non-default credentials or you're not deploying on AWS, update the `kubeone-backups-credentials`
secret (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` keys).

[backups-addon]: (./backups-restic.yaml)
[restic]: (https://restic.net/)
