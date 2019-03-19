// +build e2e

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

package e2e

import (
	"flag"
	"testing"
)

// testRunIdentifier aka. the build number, a unique identifier for the test run.
var (
	testRunIdentifier string
	testProvider      string
)

func init() {
	flag.StringVar(&testRunIdentifier, "identifier", "", "The unique identifier for this test run")
	flag.StringVar(&testProvider, "provider", "", "Provider to run tests on")
	flag.Parse()
}

func setupTearDown(p Provisioner, k Kubeone) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("cleanup ....")

		errKubeone := k.Reset()
		errProvisioner := p.Cleanup()

		if errKubeone != nil {
			t.Errorf("%v", errKubeone)
		}
		if errProvisioner != nil {
			t.Errorf("%v", errProvisioner)
		}
	}
}
