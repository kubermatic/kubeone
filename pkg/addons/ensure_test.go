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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
)

func Test_validateUserAddonNames(t *testing.T) {
	deployed := map[string]string{"foo": "", "bar": ""}
	deleted := map[string]struct{}{"old": {}}

	tests := []struct {
		name    string
		names   []string
		wantErr bool
	}{
		{name: "no names selected", names: nil},
		{name: "deployed addon", names: []string{"foo"}},
		{name: "multiple addons", names: []string{"foo", "bar"}},
		{name: "addon marked for deletion", names: []string{"old"}},
		{name: "unknown addon", names: []string{"unknown"}, wantErr: true},
		{name: "one unknown among known", names: []string{"foo", "unknown"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserAddonNames(tt.names, deployed, deleted)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUserAddonNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_userAddons(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"local-only", "to-delete"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	s := &state.State{
		Cluster: &kubeoneapi.KubeOneCluster{
			Addons: &kubeoneapi.Addons{
				Path: dir,
				Addons: []kubeoneapi.AddonRef{
					{Addon: &kubeoneapi.Addon{Name: "declared"}},
					{Addon: &kubeoneapi.Addon{Name: "to-delete", Delete: true}},
					// embedded addons are managed by kubeone apply
					{Addon: &kubeoneapi.Addon{Name: "metrics-server"}},
				},
			},
		},
	}

	deployed, deleted, err := userAddons(s)
	if err != nil {
		t.Fatalf("userAddons() error = %v", err)
	}

	wantDeployed := map[string]string{"local-only": "", "declared": ""}
	wantDeleted := map[string]struct{}{"to-delete": {}}

	if !reflect.DeepEqual(deployed, wantDeployed) {
		t.Errorf("deployed = %v, want %v", deployed, wantDeployed)
	}

	if !reflect.DeepEqual(deleted, wantDeleted) {
		t.Errorf("deleted = %v, want %v", deleted, wantDeleted)
	}
}
