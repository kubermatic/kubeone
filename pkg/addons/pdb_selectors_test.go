/*
Copyright 2026 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package addons

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"

	embeddedaddons "k8c.io/kubeone/addons"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"
)

// pdbSelectorExceptions lists PodDisruptionBudgets (keyed by "namespace/name") that intentionally select pods
// not shipped by any addon. Every entry must explain where the selected pods come from.
var pdbSelectorExceptions = map[string]string{
	// CoreDNS is deployed by kubeadm, not by an addon.
	"kube-system/coredns": "CoreDNS pods are deployed by kubeadm",
}

// pdbSelectorSkippedAddons lists addons excluded from the PodDisruptionBudget selector check.
var pdbSelectorSkippedAddons = map[string]string{
	"ccm-equinixmetal": "community supported",
}

var (
	yamlDocSeparator   = regexp.MustCompile(`(?m)^---\s*$`)
	templateActionLine = regexp.MustCompile(`(?m)^\s*\{\{.*\}\}\s*$`)
	templateAction     = regexp.MustCompile(`\{\{.*?\}\}`)
)

type pdbTestObject struct {
	Kind     string `json:"kind"`
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Selector *metav1.LabelSelector `json:"selector"`
		Template struct {
			Metadata struct {
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
		} `json:"template"`
	} `json:"spec"`
}

type pdbTestWorkload struct {
	source    string
	namespace string
	labels    labels.Set
}

type pdbTestPDB struct {
	source string
	obj    pdbTestObject
}

var podTemplateKinds = map[string]bool{
	"Deployment":  true,
	"StatefulSet": true,
	"DaemonSet":   true,
	"ReplicaSet":  true,
}

// TestPodDisruptionBudgetSelectorsMatchWorkloads ensures every PodDisruptionBudget shipped in the embedded addons
// selects the pod template of at least one workload shipped in the embedded addons. Upstream manifest updates
// regularly change pod labels or drop workloads, leaving behind hand-written PDBs that select nothing.
//
// Templates are not rendered; template actions are stripped instead, as PDB selectors and pod labels are literal.
// Workloads from all addons are considered, since a PDB may protect pods deployed by a companion addon (e.g.
// the Azure CSI addons protect snapshot-controller from csi-external-snapshotter).
func TestPodDisruptionBudgetSelectorsMatchWorkloads(t *testing.T) {
	t.Parallel()

	var (
		pdbs      []pdbTestPDB
		workloads []pdbTestWorkload
	)

	addonDirs, err := fs.ReadDir(embeddedaddons.FS, ".")
	if err != nil {
		t.Fatalf("unable to read embedded addons: %v", err)
	}

	for _, addonDir := range addonDirs {
		if !addonDir.IsDir() {
			continue
		}

		if _, skip := pdbSelectorSkippedAddons[addonDir.Name()]; skip {
			continue
		}

		// Mirror loadAddonsManifests: only top-level manifest files of each addon are applied.
		files, err := fs.ReadDir(embeddedaddons.FS, addonDir.Name())
		if err != nil {
			t.Fatalf("unable to read addon %q: %v", addonDir.Name(), err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			switch strings.ToLower(path.Ext(file.Name())) {
			case ".yaml", ".yml", ".json":
			default:
				continue
			}

			filePath := path.Join(addonDir.Name(), file.Name())

			content, err := fs.ReadFile(embeddedaddons.FS, filePath)
			if err != nil {
				t.Fatalf("unable to read %q: %v", filePath, err)
			}

			for _, doc := range yamlDocSeparator.Split(string(content), -1) {
				if !strings.Contains(doc, "PodDisruptionBudget") && !strings.Contains(doc, "template:") {
					continue
				}

				doc = templateActionLine.ReplaceAllString(doc, "")
				doc = templateAction.ReplaceAllString(doc, "templated")

				var obj pdbTestObject
				if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
					if strings.Contains(doc, "kind: PodDisruptionBudget") {
						t.Errorf("unable to parse PodDisruptionBudget in %q: %v", filePath, err)
					}

					continue
				}

				switch {
				case obj.Kind == "PodDisruptionBudget":
					pdbs = append(pdbs, pdbTestPDB{source: filePath, obj: obj})
				case podTemplateKinds[obj.Kind]:
					workloads = append(workloads, pdbTestWorkload{
						source:    filePath,
						namespace: obj.Metadata.Namespace,
						labels:    obj.Spec.Template.Metadata.Labels,
					})
				}
			}
		}
	}

	if len(pdbs) == 0 {
		t.Fatal("no PodDisruptionBudgets found in embedded addons")
	}

	for _, pdb := range pdbs {
		key := pdb.obj.Metadata.Namespace + "/" + pdb.obj.Metadata.Name

		if pdb.obj.Spec.Selector == nil {
			t.Errorf("PodDisruptionBudget %s (%s) has no selector", key, pdb.source)

			continue
		}

		selector, err := metav1.LabelSelectorAsSelector(pdb.obj.Spec.Selector)
		if err != nil {
			t.Errorf("PodDisruptionBudget %s (%s) has invalid selector: %v", key, pdb.source, err)

			continue
		}

		matched := false
		for _, w := range workloads {
			if w.namespace == pdb.obj.Metadata.Namespace && selector.Matches(w.labels) {
				matched = true

				break
			}
		}

		_, excepted := pdbSelectorExceptions[key]

		switch {
		case !matched && !excepted:
			t.Errorf("PodDisruptionBudget %s (%s) selector %q matches no workload pod template in the addons",
				key, pdb.source, selector)
		case matched && excepted:
			t.Errorf("PodDisruptionBudget %s (%s) now matches a workload; remove it from pdbSelectorExceptions",
				key, pdb.source)
		}
	}
}
