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

package state

import (
	"testing"
	"time"
)

func TestCluster_CertsToExpireInLessThen90Days(t *testing.T) {
	tests := []struct {
		name  string
		hosts []Host
		want  bool
	}{
		{
			name:  "expired now",
			hosts: []Host{{EarliestCertExpiry: time.Now()}},
			want:  true,
		},
		{
			name:  "expire soon",
			hosts: []Host{{EarliestCertExpiry: time.Now().Add(time.Hour * 24)}},
			want:  true,
		},
		{
			name:  "expire after 91 days",
			hosts: []Host{{EarliestCertExpiry: time.Now().Add(time.Hour * 24 * 91)}},
			want:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				ControlPlane: tt.hosts,
			}

			if got := c.CertsToExpireInLessThen90Days(); got != tt.want {
				t.Errorf("Cluster.CertsToExpireInLessThen90Days() = %v, want %v", got, tt.want)
			}
		})
	}
}
