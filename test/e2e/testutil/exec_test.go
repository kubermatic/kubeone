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

package testutil

import (
	"context"
	"reflect"
	"testing"
)

func TestWithMapEnv(t *testing.T) {
	tests := []struct {
		name string
		args map[string]string
		want []string
	}{
		{
			name: "hash map as env source",
			args: map[string]string{
				"ENV2": "val2",
				"ENV1": "val1",
			},
			want: []string{
				"ENV1=val1",
				"ENV2=val2",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			e := Exec{}
			WithMapEnv(tt.args)(&e)
			cmd := e.BuildCmd(context.Background())

			if got := cmd.Env; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
