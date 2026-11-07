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

package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitBackupOptIn(t *testing.T) {
	t.Run("no path requested creates nothing", func(t *testing.T) {
		wd := t.TempDir()
		t.Chdir(wd)

		if err := initBackup(""); err != nil {
			t.Fatalf("initBackup(\"\") returned error: %v", err)
		}

		entries, err := os.ReadDir(wd)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected no files to be created, got %d", len(entries))
		}
	})

	t.Run("explicit path is prepared", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "cluster.tar.gz")

		if err := initBackup(p); err != nil {
			t.Fatalf("initBackup(%q) returned error: %v", p, err)
		}
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected backup file to exist: %v", err)
		}
	})
}
