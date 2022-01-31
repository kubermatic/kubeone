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
	"sort"

	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"

	embeddedaddons "k8c.io/kubeone/addons"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tabwriter"
)

const (
	addonStatusInstall  = addonStatus("install")
	addonStatusInactive = addonStatus("inactive")
	addonStatusDelete   = addonStatus("delete")
)

type addonStatus string

type addonItem struct {
	Name   string      `json:"name"`
	Status addonStatus `json:"status"`
}

func List(s *state.State, outputFormat string) error {
	switch outputFormat {
	case "table", "json":
	default:
		return errors.Errorf("wrong format: %q", outputFormat)
	}

	combinedAddons := map[string]addonItem{}

	embeddedEntries, err := fs.ReadDir(embeddedaddons.FS, ".")
	if err != nil {
		return err
	}

	for _, addon := range embeddedEntries {
		if !addon.IsDir() {
			continue
		}

		combinedAddons[addon.Name()] = addonItem{
			Name:   addon.Name(),
			Status: addonStatusInactive,
		}
	}

	activeAddons := collectAddons(s)
	for _, activeAddon := range activeAddons {
		add, ok := combinedAddons[activeAddon.name]
		if !ok {
			continue
		}

		add.Status = addonStatusInstall
		combinedAddons[activeAddon.name] = add
	}

	if s.Cluster.Addons.Enabled() {
		addonsPath, err := s.Cluster.Addons.RelativePath(s.ManifestFilePath)
		if err != nil {
			return errors.Wrap(err, "failed to get addons path")
		}

		localFS := os.DirFS(addonsPath)
		customAddons, err := fs.ReadDir(localFS, ".")
		if err != nil {
			return errors.Wrap(err, "failed to read addons directory")
		}

		for _, useraddon := range customAddons {
			if !useraddon.IsDir() {
				continue
			}

			combinedAddons[useraddon.Name()] = addonItem{
				Name:   useraddon.Name(),
				Status: addonStatusInstall,
			}
		}

		for _, embeddedAddon := range s.Cluster.Addons.Addons {
			combinedAddons[embeddedAddon.Name] = addonItem{
				Name:   embeddedAddon.Name,
				Status: addonStatusInstall,
			}

			if embeddedAddon.Delete {
				combinedAddons[embeddedAddon.Name] = addonItem{
					Name:   embeddedAddon.Name,
					Status: addonStatusDelete,
				}
			}
		}
	}

	omap := orderedmap.New()

	for k, v := range combinedAddons {
		omap.Set(k, v)
	}
	omap.SortKeys(sort.Strings)

	switch outputFormat {
	case "json":
		buf, err := json.Marshal(omap)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", buf)
	case "table":
		tab := tabwriter.New(os.Stdout)
		defer tab.Flush()

		fmt.Fprintf(tab, "Name\tStatus\t")
		fmt.Fprintln(tab, "")

		for _, k := range omap.Keys() {
			v, _ := omap.Get(k)
			addon, _ := v.(addonItem)
			fmt.Fprintf(tab, "%s\t%s\t", addon.Name, addon.Status)
			fmt.Fprintln(tab, "")
		}
	}

	return nil
}
