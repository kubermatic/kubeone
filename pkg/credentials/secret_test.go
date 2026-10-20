/*
Copyright 2023 The KubeOne Authors.

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

package credentials

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_renderCloudConfig(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		credentials map[string]string
		want        string
		wantErr     bool
	}{
		{
			name:        "simple",
			input:       "simple {{ .Credentials.TEST }}",
			credentials: map[string]string{"TEST": "TEST1"},
			want:        "simple TEST1",
		},
		{
			name:        "broken template",
			input:       "simple {{ .Credentials.TEST ",
			credentials: map[string]string{"TEST": "TEST1"},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderCloudConfig(tt.input, tt.credentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderCloudConfig() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("renderCloudConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureKubevirt(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	legacy := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: SecretNameLegacy, Namespace: SecretNamespace}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(legacy).Build()

	s := &state.State{
		Context:       context.Background(),
		Logger:        logrus.New(),
		DynamicClient: c,
		Cluster: &kubeoneapi.KubeOneCluster{
			CloudProvider: kubeoneapi.CloudProviderSpec{
				External: true,
				Kubevirt: &kubeoneapi.KubevirtSpec{InfraNamespace: "infra-ns"},
			},
			MachineController:      &kubeoneapi.MachineControllerConfig{Deploy: true},
			OperatingSystemManager: &kubeoneapi.OperatingSystemManagerConfig{Deploy: true},
		},
	}

	if err := Ensure(s); err != nil {
		t.Fatal(err)
	}

	secrets := &corev1.SecretList{}
	if err := c.List(s.Context, secrets); err != nil {
		t.Fatal(err)
	}

	for _, secret := range secrets.Items {
		t.Errorf("expected no Secrets in the tenant cluster, found %s/%s", secret.Namespace, secret.Name)
	}
}
