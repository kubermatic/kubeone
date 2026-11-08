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

	"github.com/pkg/errors"

	embeddedaddons "k8c.io/kubeone/addons"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tabwriter"
)

const (
	addonStatusInstall  = addonStatus("install")
	addonStatusInactive = addonStatus("inactive")
	addonStatusDelete   = addonStatus("delete")

	// addonSourceEmbedded is an addon embedded in the KubeOne binary.
	addonSourceEmbedded = addonSource("embedded")
	// addonSourceOverridden is an embedded addon that is overridden by a
	// custom addon of the same name from the addons directory.
	addonSourceOverridden = addonSource("overridden")
	// addonSourceCustom is an addon that exists only in the addons directory.
	addonSourceCustom = addonSource("custom")
)

type (
	addonStatus string
	addonSource string
)

type addonItem struct {
	Name   string      `json:"name"`
	Status addonStatus `json:"status"`
	// Source is empty for an addon that is declared in the manifest, but is
	// neither embedded nor present in the addons directory.
	Source addonSource `json:"source"`
}

// resolveAddons returns all known addons sorted by name: the embedded ones,
// the ones found in the addons directory and the ones declared in the
// manifest, together with their status and the source they are applied from.
// The status is derived from the same functions that decide what
// 'kubeone apply' deploys (collectAddons) and what it deploys or deletes as
// user addons (userAddons).
func resolveAddons(s *state.State) ([]addonItem, error) {
	embeddedNames, err := addonDirNames(embeddedaddons.FS)
	if err != nil {
		return nil, fail.Runtime(err, "reading embedded addons directory")
	}

	localNames := map[string]struct{}{}

	localFS, err := addonsLocalFS(s.Cluster.Addons, s.ManifestFilePath)
	if err != nil {
		return nil, err
	}

	if localFS != nil {
		localNames, err = addonDirNames(localFS)
		if err != nil {
			return nil, fail.Runtime(err, "reading local addons directory")
		}
	}

	combinedAddons := map[string]addonItem{}

	for name := range embeddedNames {
		combinedAddons[name] = addonItem{Name: name, Status: addonStatusInactive}
	}

	// every local addon starts as inactive: a directory that overrides a managed
	// embedded addon is applied only if collectAddons selects that addon, while
	// all other local addons are user addons and are set below
	for name := range localNames {
		combinedAddons[name] = addonItem{Name: name, Status: addonStatusInactive}
	}

	for _, activeAddon := range collectAddons(s) {
		if _, ok := combinedAddons[activeAddon.name]; ok {
			combinedAddons[activeAddon.name] = addonItem{Name: activeAddon.name, Status: addonStatusInstall}
		}
	}

	deployed, deleted, err := userAddons(s)
	if err != nil {
		return nil, err
	}

	for name := range deployed {
		combinedAddons[name] = addonItem{Name: name, Status: addonStatusInstall}
	}

	for name := range deleted {
		combinedAddons[name] = addonItem{Name: name, Status: addonStatusDelete}
	}

	items := make([]addonItem, 0, len(combinedAddons))

	for name, item := range combinedAddons {
		_, isEmbedded := embeddedNames[name]
		_, isLocal := localNames[name]

		switch {
		case isEmbedded && isLocal:
			item.Source = addonSourceOverridden
		case isEmbedded:
			item.Source = addonSourceEmbedded
		case isLocal:
			item.Source = addonSourceCustom
		}

		items = append(items, item)
	}

	slices.SortFunc(items, func(a, b addonItem) int {
		return strings.Compare(a.Name, b.Name)
	})

	return items, nil
}

// addonDirNames returns the names of the addon directories in the root of fsys.
func addonDirNames(fsys fs.FS) (map[string]struct{}, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	names := map[string]struct{}{}

	for _, entry := range entries {
		if entry.IsDir() {
			names[entry.Name()] = struct{}{}
		}
	}

	return names, nil
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

	items, err := resolveAddons(s)
	if err != nil {
		return err
	}

	itemsByName := make(map[string]addonItem, len(items))

	for _, item := range items {
		itemsByName[item.Name] = item
	}

	switch outputFormat {
	case "json":
		buf, err := json.Marshal(itemsByName)
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
