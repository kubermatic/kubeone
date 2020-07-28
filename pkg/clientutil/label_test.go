/*
Copyright 2020 The KubeOne Authors.

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

package clientutil

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAddLabels(t *testing.T) {
	type args struct {
		labels  map[string]string
		objects []runtime.Object
	}

	tests := []struct {
		name string
		args args
		want []runtime.Object
	}{
		{
			name: "simple",
			args: args{
				labels: map[string]string{"testkey": "testval"},
				objects: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "canal-config",
							Namespace: metav1.NamespaceSystem,
							Labels: map[string]string{
								"existing": "label",
							},
						},
					},
				},
			},
			want: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "canal-config",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							"existing": "label",
							"testkey":  "testval",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := AddLabels(tt.args.labels, tt.args.objects...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
