package containerruntime

import (
	"encoding/json"
	"testing"

	"k8c.io/kubeone/pkg/apis/kubeone"
)

func Test_marshalDockerConfig(t *testing.T) {
	tests := []struct {
		name    string
		cluster *kubeone.KubeOneCluster
		want    string
	}{
		{
			name:    "Should be convert 100Mi to 100m",
			cluster: genCluster(withContainerLogMaxSize("100Mi")),
			want:    "100m",
		},
		{
			name:    "Should be convert 100Ki to 100k",
			cluster: genCluster(withContainerLogMaxSize("100Ki")),
			want:    "100k",
		},
		{
			name:    "Should be convert 100Gi to 100g",
			cluster: genCluster(withContainerLogMaxSize("100Gi")),
			want:    "100g",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalDockerConfig(tt.cluster)
			if err != nil {
				t.Errorf("marshalDockerConfig() error = %v,", err)
				return
			}
			cfg := dockerConfig{}
			err = json.Unmarshal([]byte(got), &cfg)
			gotLogSize := cfg.LogOpts["max-size"]

			if err != nil {
				t.Errorf("failed to unmarshal docker config: %v", err)
			}
			if gotLogSize != tt.want {
				t.Errorf("marshalDockerConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func withContainerLogMaxSize(logSize string) clusterOpts {
	return func(cls *kubeone.KubeOneCluster) {
		cls.LoggingConfig.ContainerLogMaxSize = logSize
	}
}
