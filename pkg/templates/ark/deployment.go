package ark

// arkDeployment deploys Ark version 0.10.0 using default settings
func arkDeployment() string {
	return `
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
            ## uncomment following line and specify values if needed for multiple provider snapshot locations
            # - --default-volume-snapshot-locations=<provider-1:location-1,provider-2:location-2,...>
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
            #- name: AWS_CLUSTER_NAME
            #  value: <YOUR_CLUSTER_NAME>
      volumes:
        - name: cloud-credentials
          secret:
            secretName: cloud-credentials
        - name: plugins
          emptyDir: {}
        - name: scratch
          emptyDir: {}
`
}
