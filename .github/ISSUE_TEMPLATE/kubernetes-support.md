---
name: Support Kubernetes 1.2x
about: Add support for the latest Kubernetes release
title: Support Kubernetes 1.2x
labels: sig/cluster-management, kind/feature, Epic

---

<!--
Update default admission controllers if needed:
To find out what admission controllers are enabled by default, you can run
kube-apiserver --help and search for the --enable-admission-plugins flag.
The easiest way to run kube-apiserver is using Docker such as:
docker run --rm registry.k8s.io/kube-apiserver:v1.2x.0 kube-apiserver -h
-->

This is a collector issue for Kubernetes 1.2x support in KubeOne. The following tasks should be taken care of:

* [ ] Check the Kubernetes changelog to ensure there are no breaking changes and removals
* [ ] Update [the `kubeone-e2e` image](https://github.com/kubermatic/kubeone/tree/main/hack/images/kubeone-e2e) to update Sonobuoy (and other dependencies if needed) <!-- (link to the PR) -->
* [ ] Update the latest supported Kubernetes version in [the API validation](https://github.com/kubermatic/kubeone/blob/main/pkg/apis/kubeone/validation/validation.go#L40-L41) <!-- (link to the PR) -->
* [ ] Update [default admission controllers](https://github.com/kubermatic/kubeone/blob/main/pkg/kubeflags/data.go) if needed <!-- (link to the PR) -->
* [ ] Update [the stable version marker in Makefile](https://github.com/kubermatic/kubeone/blob/5273f9a372736569c6b09b38f2959019d29e4d6a/Makefile#L24) <!-- (link to the PR) -->
* [ ] Add E2E tests <!-- (link to the PR) -->
* [ ] Update daily periodics to use the latest Kubernetes release
* [ ] Update [the Compatibility Matrix](https://docs.kubermatic.com/kubeone/main/architecture/compatibility/supported-versions/) <!-- (link to the PR) -->
* [ ] Create an issue to track updating [images](https://github.com/kubermatic/kubeone/blob/main/pkg/templates/images/images.go) <!-- link to the issue -->
* [ ] Run the full conformance tests suite using [Sonobuoy](https://github.com/vmware-tanzu/sonobuoy)

<!--
**Action items:**

* [ ] insert any action items here
-->
