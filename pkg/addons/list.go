/*
Copyright 2020 The KubeOne Authors.

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
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"

	embeddedaddons "k8c.io/kubeone/addons"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tabwriter"
)

const (
	AddonStatusInstall  = AddonStatus("install")
	AddonStatusInactive = AddonStatus("inactive")
	AddonStatusDelete   = AddonStatus("delete")

	// AddonSourceEmbedded is an addon embedded in the KubeOne binary.
	AddonSourceEmbedded = AddonSource("embedded")
	// AddonSourceOverridden is an embedded addon that is overridden by a
	// custom addon of the same name from the addons directory.
	AddonSourceOverridden = AddonSource("overridden")
	// AddonSourceCustom is an addon that exists only in the addons directory.
	AddonSourceCustom = AddonSource("custom")
)

type (
	AddonStatus string
	AddonSource string
)

type AddonItem struct {
	Name   string      `json:"name"`
	Status AddonStatus `json:"status"`
	// Source is empty for an addon that is declared in the manifest, but is
	// neither embedded nor present in the addons directory.
	Source AddonSource `json:"source"`
}

// ResolveAddons returns all known addons sorted by name: the embedded ones,
// the ones found in the addons directory and the ones declared in the
// manifest, together with their status and the source they are applied from.
func ResolveAddons(s *state.State) ([]AddonItem, error) {
	embeddedEntries, err := fs.ReadDir(embeddedaddons.FS, ".")
	if err != nil {
		return nil, fail.Runtime(err, "reading embedded addons directory")
	}

	embeddedNames := map[string]struct{}{}
	combinedAddons := map[string]AddonItem{}

	for _, addon := range embeddedEntries {
		if !addon.IsDir() {
			continue
		}

		embeddedNames[addon.Name()] = struct{}{}
		combinedAddons[addon.Name()] = AddonItem{
			Name:   addon.Name(),
			Status: AddonStatusInactive,
		}
	}

	for _, activeAddon := range collectAddons(s) {
		add, ok := combinedAddons[activeAddon.name]
		if !ok {
			continue
		}

		add.Status = AddonStatusInstall
		combinedAddons[activeAddon.name] = add
	}

	localNames := map[string]struct{}{}

	if s.Cluster.Addons.Enabled() {
		localFS, err := addonsLocalFS(s.Cluster.Addons, s.ManifestFilePath)
		if err != nil {
			return nil, err
		}

		if localFS != nil {
			customAddons, err := fs.ReadDir(localFS, ".")
			if err != nil {
				return nil, fail.Runtime(err, "reading local addons directory")
			}

			for _, useraddon := range customAddons {
				if !useraddon.IsDir() {
					continue
				}

				localNames[useraddon.Name()] = struct{}{}
				combinedAddons[useraddon.Name()] = AddonItem{
					Name:   useraddon.Name(),
					Status: AddonStatusInstall,
				}
			}
		}

		for _, declaredAddon := range s.Cluster.Addons.DeclaredAddonsOnly() {
			combinedAddons[declaredAddon.Name] = AddonItem{
				Name:   declaredAddon.Name,
				Status: AddonStatusInstall,
			}

			if declaredAddon.Delete {
				combinedAddons[declaredAddon.Name] = AddonItem{
					Name:   declaredAddon.Name,
					Status: AddonStatusDelete,
				}
			}
		}
	}

	items := make([]AddonItem, 0, len(combinedAddons))

	for name, item := range combinedAddons {
		_, isEmbedded := embeddedNames[name]
		_, isLocal := localNames[name]

		switch {
		case isEmbedded && isLocal:
			item.Source = AddonSourceOverridden
		case isEmbedded:
			item.Source = AddonSourceEmbedded
		case isLocal:
			item.Source = AddonSourceCustom
		}

		items = append(items, item)
	}

	slices.SortFunc(items, func(a, b AddonItem) int {
		return strings.Compare(a.Name, b.Name)
	})

	return items, nil
}

func List(s *state.State, outputFormat string) error {
	switch outputFormat {
	case "table", "json":
	default:
		return fail.RuntimeError{
			Op:  "validating output format",
			Err: errors.Errorf("wrong format: %q", outputFormat),
		}
	}

	items, err := ResolveAddons(s)
	if err != nil {
		return err
	}

	omap := orderedmap.New()

	for _, item := range items {
		omap.Set(item.Name, item)
	}

	switch outputFormat {
	case "json":
		buf, err := json.Marshal(omap)
		if err != nil {
			return fail.Runtime(err, "marshalling addons list")
		}

		fmt.Printf("%s\n", buf)
	case "table":
		tab := tabwriter.New(os.Stdout)
		defer tab.Flush()

		fmt.Fprintf(tab, "Name\tStatus\tSource\t")
		fmt.Fprintln(tab, "")

		for _, addon := range items {
			fmt.Fprintf(tab, "%s\t%s\t%s\t", addon.Name, addon.Status, addon.Source)
			fmt.Fprintln(tab, "")
		}
	}

	return nil
}
