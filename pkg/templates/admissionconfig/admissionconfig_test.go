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

package admissionconfig

import (
	"strings"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func TestNewAdmissionConfigEventRateLimit(t *testing.T) {
	cfg, err := NewAdmissionConfig("1.33.4", nil, &kubeoneapi.EventRateLimit{
		Enable: true,
		Config: kubeoneapi.EventRateLimitConfig{ConfigFilePath: "./eventratelimit.yaml"},
	})
	if err != nil {
		t.Fatalf("NewAdmissionConfig returned error: %v", err)
	}

	if !strings.Contains(cfg, "name: EventRateLimit") {
		t.Fatalf("generated admission config is missing EventRateLimit plugin: %s", cfg)
	}
	if !strings.Contains(cfg, "path: "+eventRateLimitAdmissionConfigPath) {
		t.Fatalf("generated admission config is missing EventRateLimit path: %s", cfg)
	}
}

func TestNewAdmissionConfigEventRateLimitDisabled(t *testing.T) {
	cfg, err := NewAdmissionConfig("1.33.4", nil, &kubeoneapi.EventRateLimit{
		Enable: false,
		Config: kubeoneapi.EventRateLimitConfig{ConfigFilePath: "./eventratelimit.yaml"},
	})
	if err != nil {
		t.Fatalf("NewAdmissionConfig returned error: %v", err)
	}

	if strings.Contains(cfg, "EventRateLimit") {
		t.Fatalf("disabled EventRateLimit should not appear in admission config: %s", cfg)
	}
}

func TestNewAdmissionConfigMultiplePlugins(t *testing.T) {
	cfg, err := NewAdmissionConfig("1.33.4", &kubeoneapi.PodNodeSelector{
		Enable: true,
		Config: kubeoneapi.PodNodeSelectorConfig{ConfigFilePath: "./podnodeselector.yaml"},
	}, &kubeoneapi.EventRateLimit{
		Enable: true,
		Config: kubeoneapi.EventRateLimitConfig{ConfigFilePath: "./eventratelimit.yaml"},
	})
	if err != nil {
		t.Fatalf("NewAdmissionConfig returned error: %v", err)
	}

	for _, expected := range []string{
		"name: PodNodeSelector",
		"path: " + podNodeSelectorAdmissionConfigPath,
		"name: EventRateLimit",
		"path: " + eventRateLimitAdmissionConfigPath,
	} {
		if !strings.Contains(cfg, expected) {
			t.Fatalf("generated admission config is missing %q: %s", expected, cfg)
		}
	}
}
