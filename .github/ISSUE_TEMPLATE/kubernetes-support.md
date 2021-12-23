---
name: Support Kubernetes 1.2x
about: Add support for the latest Kubernetes release
labels: sig/cluster-management, kind/feature, Epic

---

<!--
Update default admission controllers if needed:
To find out what admission controllers are enabled by default, you can run
kube-apiserver --help and search for the --enable-admission-plugins flag.
The easiest way to run kube-apiserver is using Docker such as:
docker run --rm k8s.gcr.io/kube-apiserver:v1.2x.0 kube-apiserver -h
-->

This is a collector issue for Kubernetes 1.2x support in KubeOne. The following tasks should be taken care of:

* [ ] Check the Kubernetes changelog to ensure there are no breaking changes and removals
* [ ] Update [the `kubeone-e2e` image](https://github.com/kubermatic/kubeone/tree/master/hack/images/kubeone-e2e) to add the needed Kubernetes test binaries <!-- link to the PR -->
* [ ] Update [default admission controllers](https://github.com/kubermatic/kubeone/blob/master/pkg/kubeflags/data.go) if needed <!-- link to the PR -->
* [ ] Add E2E tests <!-- link to the PR -->
* [ ] Update daily periodics to use the latest Kubernetes release
* [ ] Update [the Compatibility Matrix](https://docs.kubermatic.com/kubeone/master/architecture/compatibility/) <!-- link to the PR -->
* [ ] Create an issue to track updating [images](https://github.com/kubermatic/kubeone/blob/master/pkg/templates/images/images.go) <!-- link to the issue -->
* [ ] Run the full conformance tests suite using [Sonobuoy](https://github.com/vmware-tanzu/sonobuoy)

<!--
**Action items:**

* [ ] insert any action items here
-->
