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
	"encoding/base64"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func Test_kubevirtSecret(t *testing.T) {
	plainKubeconfig := "apiVersion: v1\nkind: Config\n"
	b64Kubeconfig := base64.StdEncoding.EncodeToString([]byte(plainKubeconfig))

	makeSecret := func(data map[string][]byte) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      SecretNameCCM,
				Namespace: SecretNamespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: data,
		}
	}

	tests := []struct {
		name        string
		credentials map[string]string
		want        *corev1.Secret
	}{
		{
			name:        "plain text kubeconfig stored as-is",
			credentials: map[string]string{KubevirtKubeconfigKey: plainKubeconfig},
			want: makeSecret(map[string][]byte{
				KubevirtKubeconfigKey: []byte(plainKubeconfig),
			}),
		},
		{
			name:        "base64-encoded kubeconfig is decoded",
			credentials: map[string]string{KubevirtKubeconfigKey: b64Kubeconfig},
			want: makeSecret(map[string][]byte{
				// base64.StdEncoding.DecodedLen over-allocates; the slice includes trailing null bytes
				KubevirtKubeconfigKey: func() []byte {
					decoded := make([]byte, base64.StdEncoding.DecodedLen(len(b64Kubeconfig)))
					base64.StdEncoding.Decode(decoded, []byte(b64Kubeconfig)) //nolint:errcheck

					return decoded
				}(),
			}),
		},
		{
			name:        "invalid base64 falls back to raw bytes",
			credentials: map[string]string{KubevirtKubeconfigKey: "not!!valid==base64"},
			want: makeSecret(map[string][]byte{
				KubevirtKubeconfigKey: []byte("not!!valid==base64"),
			}),
		},
		{
			name:        "missing kubeconfig key results in empty data",
			credentials: map[string]string{},
			want: makeSecret(map[string][]byte{
				KubevirtKubeconfigKey: []byte(""),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kubevirtSecret(SecretNameCCM, tt.credentials)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("kubevirtSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}
