---
name: Cut a new KubeOne release
about: Create a tracking issue for cutting a new KubeOne release
title: Release KubeOne 1.x.y-{alpha,beta,rc}.z
labels: sig/cluster-management, kind/documentation, priority/high
---

<!--
* Make sure to uncomment the appropriate section depending on the type of
  release you're cutting.
* Update the issue with any problems encountered during the release process.
* You can add/remove items as needed.
* The Action Items section can be used for any follow-up tasks.
-->

## Scheduled to happen: <!-- Wednesday, 2021-03-10 -->

### Before Release

* [ ] Write and publish the changelog <!-- (reference to the changelog PR) -->
<!-- UNCOMMENT FOR RELEASE CANDIDATES (RCs)
* [ ] Create and push the release branch (`release/1.x`)
* [ ] Create the milestone for the next release (`KubeOne 1.x+1`)
* [ ] Update the Prow config
  * [ ] Update the `branchprotector` rules
  * [ ] Update the `milestone_applier` rules
  * [ ] Enable the Code Freeze
* [ ] Create docs for the release based on docs for the main branch (copy
  `content/kubeone/main` to `content/kubeone/v1.x` in the
  [docs repo](https://github.com/kubermatic/docs)) (link to the docs PR)
-->

### Release

* [ ] Create and push a new tag
* [ ] Ensure that the release job has succeeded
  (watch https://public-prow.loodse.com/?job=post-kubeone-release)

### After Release

* [ ] Update the release's description on the GitHub Releases page to replace
  the automatically generated description with the changelog
* [ ] Download the binaries from GitHub and make sure that:
  * [ ] Checksums are matching
  * [ ] `kubeone version` returns the expected version
<!-- UNCOMMENT FOR RELEASE CANDIDATES (RCs)
* [ ] Run manual tests
-->
<!-- UNCOMMENT FOR FINAL RELEASES
* [ ] Disable the Code Freeze
-->
<!-- UNCOMMENT IF RELEASE INTRODUCES SUPPORT FOR A NEW KUBERNETES VERSION
* [ ] If the release introduces support for a new Kubernetes version, submit
  conformance results to https://github.com/cncf/k8s-conformance/
-->

### Action Items

<!--
This section can be used for any follow-up items/tasks.

* [ ] Item 1
-->

/assign <!-- insert GitHub handle here -->
