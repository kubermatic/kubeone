# Unattended Upgrades

This addon will automate upgrading system packages of the distro of your choice.

## Requirements

Since KubeOne 1.3+ we automatically label control-plane nodes with
`v1.kubeone.io/operating-system` and worker nodes with
`v1.machine-controller.kubermatic.io/operating-system` and use those labels as
nodeAffinity in this addon manifests.

## What's included

This addon provides bunch of DaemonSets and operators:

* **Debian/Ubuntu**
  DaemonSet that will install `unattended-upgrades`
* **RHEL/CentOS/Rocky Linux/Amazon Linux 2**
  DaemonSet that will install and configure `yum-cron`/`dnf-automatic`
* **Flatcar Linux**
  [Flatcar Linux Update Operator](https://github.com/kinvolk/flatcar-linux-update-operator)

## Deployment instructions

In `kubeone.yaml` config add:
```yaml
addons:
  enable: true
  addons:
    - name: "unattended-upgrades"
```

## Information about permissions

Since daemonSets provided by this addon are making changes on the nodes
themselves they require elevated permissions like full root access to the host
machine.

## Development

### Generate Flatcar Linux Update Operator YAML
```shell
kubectl kustomize ./fluo > fluo.yaml
```
