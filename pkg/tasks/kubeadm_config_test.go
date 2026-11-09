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

package tasks

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func kubeSystemConfigMap(name, key string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: metav1.NamespaceSystem},
		Data:       map[string]string{key: "old"},
	}
}

func Test_updateKubeadmConfigMaps(t *testing.T) {
	t.Parallel()

	kubeadmConfig := &kubeadm.Config{
		ClusterConfiguration:   "cluster",
		KubeletConfiguration:   "kubelet",
		KubeProxyConfiguration: "kube-proxy",
	}

	tests := []struct {
		name    string
		objs    []client.Object
		wantErr bool
	}{
		{
			name: "all ConfigMaps present",
			objs: []client.Object{
				kubeSystemConfigMap("kubeadm-config", "ClusterConfiguration"),
				kubeSystemConfigMap("kubelet-config", "kubelet"),
				kubeSystemConfigMap("kube-proxy", "config.conf"),
			},
		},
		{
			name: "kubeadm-config missing",
			objs: []client.Object{
				kubeSystemConfigMap("kubelet-config", "kubelet"),
				kubeSystemConfigMap("kube-proxy", "config.conf"),
			},
			wantErr: true,
		},
		{
			name: "kube-proxy missing",
			objs: []client.Object{
				kubeSystemConfigMap("kubeadm-config", "ClusterConfiguration"),
				kubeSystemConfigMap("kubelet-config", "kubelet"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &state.State{
				Context:       context.Background(),
				Logger:        logrus.New(),
				DynamicClient: fake.NewClientBuilder().WithObjects(tt.objs...).Build(),
			}

			err := updateKubeadmConfigMaps(s, kubeadmConfig)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error when a ConfigMap can't be updated, got nil")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for name, want := range map[string]struct{ key, value string }{
				"kubeadm-config": {"ClusterConfiguration", "cluster"},
				"kubelet-config": {"kubelet", "kubelet"},
				"kube-proxy":     {"config.conf", "kube-proxy"},
			} {
				cm := corev1.ConfigMap{}
				if err := s.DynamicClient.Get(context.Background(), client.ObjectKey{Name: name, Namespace: metav1.NamespaceSystem}, &cm); err != nil {
					t.Fatalf("getting %s: %v", name, err)
				}
				if got := cm.Data[want.key]; got != want.value {
					t.Errorf("%s[%s] = %q, want %q", name, want.key, got, want.value)
				}
			}
		})
	}
}
