// +build e2e

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
