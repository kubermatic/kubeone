package ark

import (
	"bytes"
	"fmt"

	"github.com/alecthomas/template"
	"github.com/kubermatic/kubeone/pkg/config"
)

// deployment deploys Ark version 0.10.0 using default settings
func deployment(cluster *config.Cluster) (string, error) {
	const deploy = `
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  namespace: heptio-ark
  name: ark
spec:
  replicas: 1
  template:
    metadata:
      labels:
        component: ark
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8085"
        prometheus.io/path: "/metrics"
    spec:
      restartPolicy: Always
      serviceAccountName: ark
      containers:
        - name: ark
          image: gcr.io/heptio-images/ark:v0.10.0
          command:
            - /ark
          args:
            - server
          volumeMounts:
            - name: cloud-credentials
              mountPath: /credentials
            - name: plugins
              mountPath: /plugins
            - name: scratch
              mountPath: /scratch
          env:
            - name: AWS_SHARED_CREDENTIALS_FILE
              value: /credentials/cloud
            - name: ARK_SCRATCH_DIR
              value: /scratch
            - name: AWS_CLUSTER_NAME
              value: {{ .AWS_CLUSTER_NAME }}
      volumes:
        - name: cloud-credentials
          secret:
            secretName: cloud-credentials
        - name: plugins
          emptyDir: {}
        - name: scratch
          emptyDir: {}
`

	tpl, err := template.New("base").Parse(deploy)
	if err != nil {
		return "", fmt.Errorf("failed to parse ark deployment manifest: %v", err)
	}

	variables := map[string]interface{}{
		"AWS_CLUSTER_NAME": cluster.Name,
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render flannel config: %v", err)
	}

	return buf.String(), nil
}
