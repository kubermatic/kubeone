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

package yamled

import "testing"

func assertPath(t *testing.T, value Path, expected Path) {
	t.Helper()
	if len(value) != len(expected) {
		t.Fatalf("Expected path of length %d but got %d elements.", len(expected), len(value))
	}

	for idx, step := range value {
		if step != expected[idx] {
			t.Fatalf("Path should have been %v, but is %v.", expected, value)
		}
	}
}

func TestEmptyPathParent(t *testing.T) {
	assertPath(t, Path{}.Parent(), Path{})
}

func TestPathEmptyParent(t *testing.T) {
	assertPath(t, Path{"a"}.Parent(), Path{})
}

func TestPathParent(t *testing.T) {
	assertPath(t, Path{"a", "b", "c"}.Parent(), Path{"a", "b"})
}

func TestEmptyPathTail(t *testing.T) {
	if tail := (Path{}).Tail(); tail != nil {
		t.Errorf("tail of an empty path should be nil, but is %v", tail)
	}
}

func TestPathTail(t *testing.T) {
	if tail := (Path{"a", "b", "c"}).Tail(); tail != "c" {
		t.Errorf("tail of [a b c] should be a, but is %v", tail)
	}
}
