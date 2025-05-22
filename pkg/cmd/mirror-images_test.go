package cmd

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestRetagImage(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		registry string
		want     string
		wantErr  bool
	}{
		{
			name:     "coredns special case",
			source:   "registry.k8s.io/coredns/coredns:v1.8.6",
			registry: "myregistry",
			want:     "myregistry/coredns:v1.8.6",
		},
		{
			name:     "regular image",
			source:   "nginx:latest",
			registry: "myregistry",
			want:     "myregistry/library/nginx:latest",
		},
		{
			name:     "Default kube-api-server image",
			source:   "registry.k8s.io/api-server:tag",
			registry: "myregistry",
			want:     "myregistry/api-server:tag",
		},
		{
			name:     "invalid image",
			source:   "invalid_image%%%_ref",
			registry: "myregistry",
			wantErr:  true,
		},
	}

	log := logrus.New()
	log.Out = io.Discard

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retagImage(log, tt.source, tt.registry)

			if (err != nil) != tt.wantErr {
				t.Errorf("retagImage() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("retagImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
