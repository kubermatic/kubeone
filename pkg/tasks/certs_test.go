/*
Copyright 2021 The KubeOne Authors.

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
	"testing"
	"time"
)

func Test_timeBefore(t *testing.T) {
	type args struct {
		t1 time.Time
		t2 time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "t2 zero",
			args: args{
				t1: time.Now(),
			},
			want: true,
		},
		{
			name: "simple",
			args: args{
				t1: time.Now(),
				t2: time.Now().Add(time.Minute),
			},
			want: true,
		},
		{
			name: "simple",
			args: args{
				t1: time.Now().Add(time.Minute),
				t2: time.Now(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := timeBefore(tt.args.t1, tt.args.t2); got != tt.want {
				t.Errorf("timeBefore() = %v, want %v", got, tt.want)
			}
		})
	}
}
