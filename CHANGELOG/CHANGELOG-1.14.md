# [v1.14.0](https://github.com/kubermatic/kubeone/releases/tag/v1.14.0) - 2026-07-27

## Changelog since v1.13.0

## Changes by Kind

### Feature

- Add support for Kubernetes v1.36
  - Remove everything related to CCM/CSI migration ([#4096](https://github.com/kubermatic/kubeone/pull/4096), [@kron4eg](https://github.com/kron4eg))
- Add support for KubeOne-managed KubeVirt control planes ([#4108](https://github.com/kubermatic/kubeone/pull/4108), [@kron4eg](https://github.com/kron4eg))
- Add openstack managed control-plane ([#4098](https://github.com/kubermatic/kubeone/pull/4098), [@kron4eg](https://github.com/kron4eg))
- Add kubevirt as a supported provider for init command ([#4110](https://github.com/kubermatic/kubeone/pull/4110), [@kron4eg](https://github.com/kron4eg))
- Add parallel image copy support for mirror-images ([#4107](https://github.com/kubermatic/kubeone/pull/4107), [@kron4eg](https://github.com/kron4eg))
- Add successfulJobsHistoryLimit param to backups-restic addon ([#4113](https://github.com/kubermatic/kubeone/pull/4113), [@kron4eg](https://github.com/kron4eg))
- Support gatewayAPI in Cilium ([#4069](https://github.com/kubermatic/kubeone/pull/4069), [@kron4eg](https://github.com/kron4eg))
- EnableLocalRedirectPolicy for Cilium ([#4070](https://github.com/kubermatic/kubeone/pull/4070), [@kron4eg](https://github.com/kron4eg))
- `kubeone certificates renew` will detect and fix the drift between apiserver certificate and APIEndpoint.AlternativeNames ([#4114](https://github.com/kubermatic/kubeone/pull/4114), [@kron4eg](https://github.com/kron4eg))

### Bug or Regression

- Fix Azure CCM and CNM image versions ([#4042](https://github.com/kubermatic/kubeone/pull/4042), [@kron4eg](https://github.com/kron4eg))
- Fix csi-azuredisk-controller pods crashing on startup. ([#4064](https://github.com/kubermatic/kubeone/pull/4064), [@bastianpaetzold](https://github.com/bastianpaetzold))
- Fixed typo in cilium configmap ([#4044](https://github.com/kubermatic/kubeone/pull/4044), [@kron4eg](https://github.com/kron4eg))
- Fix v1.kubelet-config annotations in terraform examples ([#4090](https://github.com/kubermatic/kubeone/pull/4090), [@kron4eg](https://github.com/kron4eg))
- Avoid modprobe ip_tables ([#4088](https://github.com/kubermatic/kubeone/pull/4088), [@kron4eg](https://github.com/kron4eg))
- The jump host can have a different SSH key. ([#4063](https://github.com/kubermatic/kubeone/pull/4063), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Fixed `kubeone reset` failing to delete PersistentVolumes when a PodDisruptionBudget blocked Pod eviction, and PVCs getting stuck in Terminating because workload controllers recreated the Pods during volume cleanup. ([#4100](https://github.com/kubermatic/kubeone/pull/4100), [@rajaSahil](https://github.com/rajaSahil))
- Prevent Gateway controllers from recreating LoadBalancer Services during reset. ([#4084](https://github.com/kubermatic/kubeone/pull/4084), [@rajaSahil](https://github.com/rajaSahil))
- Update Operating System Manager image version to v1.10.5 ([#4068](https://github.com/kubermatic/kubeone/pull/4068), [@kron4eg](https://github.com/kron4eg))
- Upgrade helm library to helm4 ([#4123](https://github.com/kubermatic/kubeone/pull/4123), [@kron4eg](https://github.com/kron4eg))
- Upgrade golang.org/x/crypto v0.52.0 ([#4092](https://github.com/kubermatic/kubeone/pull/4092), [@kron4eg](https://github.com/kron4eg))
- Update google.golang.org/grpc from 1.82.0 to 1.82.1 ([#4148](https://github.com/kubermatic/kubeone/pull/4148), [@dependabot[bot]](https://github.com/apps/dependabot))

### Documentation

- Docs: add managed control-plane guides for Hetzner and OpenStack ([#4143](https://github.com/kubermatic/kubeone/pull/4143), [@kron4eg](https://github.com/kron4eg))

### Chore

- Add SBOM generation and Cosign signing to release workflow. ([#4117](https://github.com/kubermatic/kubeone/pull/4117), [@rajaSahil](https://github.com/rajaSahil))
- Add govulncheck to the linter job ([#4051](https://github.com/kubermatic/kubeone/pull/4051), [@kron4eg](https://github.com/kron4eg))
- Kubeone mirror-images github action ([#4116](https://github.com/kubermatic/kubeone/pull/4116), [@kron4eg](https://github.com/kron4eg))
- Let unattended-upgrades work on RHEL9 like systems as well ([#4147](https://github.com/kubermatic/kubeone/pull/4147), [@kron4eg](https://github.com/kron4eg))
- Upgrade machine-controller to v1.66.0 ([#4142](https://github.com/kubermatic/kubeone/pull/4142), [@kron4eg](https://github.com/kron4eg))
- Update vSphere CSI to v3.7.1
  Update Hetzner CSI to v2.21.2
  Update Hetzner CCM to v1.31.1
  Update GCP CSI driver to v1.26.0
  Update GCP CCM to v36.0.7
  Update DigitalOcean CSI driver to v4.17.0
  Update DigitalOcean CCM to v0.1.67
  Update Azure File CSI driver to v1.35.3
  Update Azure Disk CSI driver to v1.34.4
  Update AWS CSI driver to v1.61.1
  Update Canal CNI to v3.32.0
  Update cilium to v1.19.4
  Update OSM to the latest main ([#4106](https://github.com/kubermatic/kubeone/pull/4106), [@kron4eg](https://github.com/kron4eg))
- Upgrade hcloud-ccm to v1.34.0 ([#4151](https://github.com/kubermatic/kubeone/pull/4151), [@rajaSahil](https://github.com/rajaSahil))
- Upgrade AzureFile CSI Driver to v1.35.2
  - Upgrade AzureDisk CSI Driver to v1.34.3 ([#4055](https://github.com/kubermatic/kubeone/pull/4055), [@kron4eg](https://github.com/kron4eg))
- Upgrade GCP CSI compute-persistent driver to v1.23.3
  Upgrade GCP CCM to v35.0.2
  Upgrade DigitalOceam CSI Driver to v4.16.0
  Upgrade DigitalOcean CCM to v0.1.66
  Upgrade AzureDisk CSI driver to 1.34.2
  Upgrade AzureFile CSI driver to v1.35.1 ([#4054](https://github.com/kubermatic/kubeone/pull/4054), [@kron4eg](https://github.com/kron4eg))
- Update restic addon images restic 0.19.1, etcd 3.5.32 ([#4134](https://github.com/kubermatic/kubeone/pull/4134), [@kron4eg](https://github.com/kron4eg))
- Update Flatcar version to 4593.2.2 for azure ([#4140](https://github.com/kubermatic/kubeone/pull/4140), [@rajaSahil](https://github.com/rajaSahil))
- Update golang.org/x/net to v0.54.0 ([#4076](https://github.com/kubermatic/kubeone/pull/4076), [@kron4eg](https://github.com/kron4eg))
- Update oras-go dependency to v2.6.1 ([#4120](https://github.com/kubermatic/kubeone/pull/4120), [@kron4eg](https://github.com/kron4eg))
- Fix goreleaser formats ([#4049](https://github.com/kubermatic/kubeone/pull/4049), [@kron4eg](https://github.com/kron4eg))
- Bump github.com/containerd/containerd from 1.7.30 to 1.7.33 ([#4111](https://github.com/kubermatic/kubeone/pull/4111), [@dependabot[bot]](https://github.com/apps/dependabot))
- Upgrade helm.sh/helm/v3 to 3.20.2 ([#4041](https://github.com/kubermatic/kubeone/pull/4041), [@dependabot[bot]](https://github.com/apps/dependabot))
