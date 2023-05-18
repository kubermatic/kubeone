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

import "testing"

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
