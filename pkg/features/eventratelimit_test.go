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

package features

import (
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"
)

func TestActivateKubeadmEventRateLimit(t *testing.T) {
	args := kubeadmargs.New()

	activateKubeadmEventRateLimit(&kubeoneapi.EventRateLimit{Enable: true}, args)

	if got := args.APIServer.ExtraArgs[apiServerAdmissionPluginsFlag]; got != eventRateLimitAdmissionPlugin {
		t.Fatalf("unexpected admission plugin flag: got %q, want %q", got, eventRateLimitAdmissionPlugin)
	}
	if got := args.APIServer.ExtraArgs[apiServerAdmissionControlConfigFlag]; got != apiServerAdmissionControlConfigPath {
		t.Fatalf("unexpected admission config flag: got %q, want %q", got, apiServerAdmissionControlConfigPath)
	}
}

func TestActivateKubeadmEventRateLimitNil(t *testing.T) {
	args := kubeadmargs.New()

	activateKubeadmEventRateLimit(nil, args)

	if got, ok := args.APIServer.ExtraArgs[apiServerAdmissionPluginsFlag]; ok {
		t.Fatalf("expected no admission plugin flag, got %q", got)
	}
	if got, ok := args.APIServer.ExtraArgs[apiServerAdmissionControlConfigFlag]; ok {
		t.Fatalf("expected no admission config flag, got %q", got)
	}
}

func TestActivateKubeadmEventRateLimitDisabled(t *testing.T) {
	args := kubeadmargs.New()

	activateKubeadmEventRateLimit(&kubeoneapi.EventRateLimit{Enable: false}, args)

	if got, ok := args.APIServer.ExtraArgs[apiServerAdmissionPluginsFlag]; ok {
		t.Fatalf("expected no admission plugin flag, got %q", got)
	}
	if got, ok := args.APIServer.ExtraArgs[apiServerAdmissionControlConfigFlag]; ok {
		t.Fatalf("expected no admission config flag, got %q", got)
	}
}
