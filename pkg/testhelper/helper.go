/*
Copyright 2019 The KubeOne Authors.

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

// originally copied from https://github.com/kubermatic/machine-controller/blob/main/pkg/test/helper.go
package testhelper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

func FSGoldenName(t *testing.T) string {
	t.Helper()

	return strings.ReplaceAll(t.Name(), "/", "-") + ".golden"
}

func DiffOutput(t *testing.T, name, output string, update bool) {
	t.Helper()
	golden, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to get absolute path to testdata file: %v", err)
	}

	if update {
		if errw := os.WriteFile(golden, []byte(output), 0600); errw != nil {
			t.Fatalf("failed to write updated fixture: %v", errw)
		}
	}

	expected, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("failed to read testdata file: %v", err)
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(expected)),
		B:        difflib.SplitLines(output),
		FromFile: "Fixture",
		ToFile:   "Current",
		Context:  3,
	}

	diffStr, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		t.Fatal(err)
	}

	if diffStr != "" {
		t.Fatalf("got diff between expected and actual result: \n%s\n", diffStr)
	}
}
