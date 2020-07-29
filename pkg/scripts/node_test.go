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

package scripts

import (
	"testing"

	"k8c.io/kubeone/pkg/testhelper"
)

func TestDrainNode(t *testing.T) {
	got, err := DrainNode("testNode1")
	if err != nil {
		t.Errorf("DrainNode() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUncordonNode(t *testing.T) {
	got, err := UncordonNode("testNode2")
	if err != nil {
		t.Errorf("UncordonNode() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
