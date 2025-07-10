# v1.11.0

## Changelog since v1.10.0

## Changes by Kind

### Feature

- Automate caBundle injection to the PodSpec of addons ([#3683](https://github.com/kubermatic/kubeone/pull/3683), [@kron4eg](https://github.com/kron4eg))
- Add CA Config API
  - Deprecate CABundle field ([#3647](https://github.com/kubermatic/kubeone/pull/3647), [@kron4eg](https://github.com/kron4eg))
- Add `--insecure`  flag to mirror-images command to bypass TLS verification ([#3657](https://github.com/kubermatic/kubeone/pull/3657), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Add support for kubernetes version 1.33. ([#3651](https://github.com/kubermatic/kubeone/pull/3651), [@soer3n](https://github.com/soer3n))
- Add the ability to override containerd sandbox image ([#3646](https://github.com/kubermatic/kubeone/pull/3646), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Allow control-plane and static workers nodes annotation ([#3658](https://github.com/kubermatic/kubeone/pull/3658), [@kron4eg](https://github.com/kron4eg))
- Introduce `mirror-images` command to mirror images used by kubeone into another registry ([#3631](https://github.com/kubermatic/kubeone/pull/3631), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Support KubeVirt CCM Deployment ([#3661](https://github.com/kubermatic/kubeone/pull/3661), [@moadqassem](https://github.com/moadqassem))

### Bug or Regression

- Bump helm.sh/helm/v3 to v3.18.4 ([#3744](https://github.com/kubermatic/kubeone/pull/3744), [@dependabot[bot]](https://github.com/apps/dependabot))
- Controlplane nodes will now have the fs.inotify.max_user_instances limit increased aswell ([#3649](https://github.com/kubermatic/kubeone/pull/3649), [@4ch3los](https://github.com/4ch3los))
- Drop kubevirt infraClusterKubeconfig API fields ([#3674](https://github.com/kubermatic/kubeone/pull/3674), [@kron4eg](https://github.com/kron4eg))
- Fix CABundle flag for OSM ([#3642](https://github.com/kubermatic/kubeone/pull/3642), [@kron4eg](https://github.com/kron4eg))
- Fix Canal dualstack setup ([#3747](https://github.com/kubermatic/kubeone/pull/3747), [@kron4eg](https://github.com/kron4eg))
- Fixes the vsphere-config-secret name misalignment across vSphere CSI driver components. ([#3745](https://github.com/kubermatic/kubeone/pull/3745), [@rajaSahil](https://github.com/rajaSahil))

### Chore

- Update GCP CSI manifests reference to v1.20.0
- Update OpenStack CSI driver to v2.33.0
- Update OpenStack CCM to v2.33.0
- Update Hetzner CSI driver to v2.16.0
- Update Hetzner CCM to v1.25.1
- Update vSphere CSI driver to v3.5.0
- Update Azure Disk CSI drive to v1.33.1
- Update CSI Azure File driver to v1.33.2
- Update Azure CCM to v1.33.1
- Update AWS CSI to v1.45.0
- Update AWS CCM
- Update calico CNI v3.30.2 ([#3683](https://github.com/kubermatic/kubeone/pull/3683), [@kron4eg](https://github.com/kron4eg))
- Bump Go version to 1.24.4 ([#3682](https://github.com/kubermatic/kubeone/pull/3682), [@archups](https://github.com/archups))
- Bump OSM to [v1.7.2](https://github.com/kubermatic/operating-system-manager/releases/tag/v1.7.2) and MC to [v1.62.0](https://github.com/kubermatic/machine-controller/releases/tag/v1.62.0) ([#3725](https://github.com/kubermatic/kubeone/pull/3725), [@archups](https://github.com/archups))
- Bump operating-system-manager to [v1.7.4](https://github.com/kubermatic/operating-system-manager/releases/tag/v1.7.4) ([#3738](https://github.com/kubermatic/kubeone/pull/3738), [@archups](https://github.com/archups))
- Images updates:
  - Calico / Canal v3.30.0
  - Cilium: v1.17.3
  - cluster-autoscaler: v1.30.4
  - cluster-autoscaler: v1.31.2
  - cluster-autoscaler: v1.32.1
  - AWS CCM: v1.33.0
  - AWS Ebs Csi: v1.43.0
  - AWS Ebs Csi Attacher: v4.8.1-eks-1-33-2
  - AWS Ebs Csi Livenessprobe: v2.15.0-eks-1-33-3
  - AWS Ebs Csi Node Driver Registrar: v2.13.0-eks-1-33-2
  - AWS Ebs Csi External Provisioner: v5.2.0-eks-1-33-3
  - AWS Ebs Csi External Resizer: v1.13.2-eks-1-33-3
  - Azure CCM: v1.30.12
  - Azure CCM: v1.31.6
  - Azure CCM: v1.32.5
  - * Azure CCM: v1.33.0
  - * Azure CNM: v1.30.12
  - Azure CNM: v1.31.6
  - Azure CNM: v1.32.5
  - * Azure CNM: v1.33.0
  - Azure Disk CSI: v1.33.0
  - Azure File CSI: v1.33.0
  - Digitalocean CSI Attacher: vv4.8.1
  - Digitalocean CSI Provisioner: v5.2.0
  - Digitalocean CSI Resizer: v1.13.2
  - Digitalocean CSI Snapshotter: v8.2.1
  - Hetzner CCM: v1.24.
  - OpenstackCCM: v1.30.3
  - OpenstackCCM: v1.31.2
  - OpenStack Cinder CSI: v1.30.3
  - OpenStack Cinder CSI: v1.31.3
  - Vsphere CCM; v1.30.2
  - Vsphere CCM; v1.31.1
  - Vsphere CCM; v1.32.1
  - Vsphere CCM; v1.33.0
  - Vsphere CSI Driver: v3.4.0
  - Vsphere CSI Syncer: v3.4.0
  - Vsphere CSI Attacher: v4.8.1
  - Vsphere CSI Livenessprobe: v2.15.0
  - Vsphere CSI Node Driver Registrar: v2.13.0
  - Vsphere CSI Provisioner; v5.2.0
  - Vsphere CSI Resizer: v1.13.2
  - Vsphere CSI Snapshotter: v8.2.1
  - GCP CCM: v32.2.2
  - GCP Compute CSI Driver: v1.15.4
  - GCP Compute CSI Provisioner: v5.2.0
  - GCP Compute CSI Attacher: v4.8.1
  - GCP Compute CSI Resizer v1.13.2
  - GCP Compute CSI Snapshotter: v8.2.1
  - GCP Compute CSI Node Driver Registrar: v2.13.0
  - Restic: v0.17.3
  - CSI external snapshot: v8.2.1
  - Cilium Hubble: v0.13.2 ([#3653](https://github.com/kubermatic/kubeone/pull/3653), [@soer3n](https://github.com/soer3n))
- Images updates:
  - OperatingSystemManager v1.6.5
  - DigitalOcean CCM: v1.59.0
  - DigitalOcean CSI Plugin: v4.14.0 ([#3660](https://github.com/kubermatic/kubeone/pull/3660), [@soer3n](https://github.com/soer3n))
- Update Hetzner CCM to the v1.26.0 ([#3732](https://github.com/kubermatic/kubeone/pull/3732), [@kron4eg](https://github.com/kron4eg))
- Update Flannel CNI to v0.24.4 ([#3736](https://github.com/kubermatic/kubeone/pull/3736), [@kron4eg](https://github.com/kron4eg))
- Update MachineController to v1.61.3 ([#3672](https://github.com/kubermatic/kubeone/pull/3672), [@xmudrii](https://github.com/xmudrii))
- Update to Go 1.24.2 ([#3618](https://github.com/kubermatic/kubeone/pull/3618), [@xrstf](https://github.com/xrstf))

### Other (Cleanup or Flake)

- Make kubeone go installa-ble again ([#3648](https://github.com/kubermatic/kubeone/pull/3648), [@kron4eg](https://github.com/kron4eg))
