package templates

import (
	corev1 "k8s.io/api/core/v1"
	rbac_v1beta1 "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

//  Downloaded from https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml

func FlannelConfiguration(manifest *manifest.Manifest, instance int) (string, error) {
	items := []interface{}{
		flannelServiceAccount(),
		flannelClusterRole(),
		flannelClusterRoleBinding(),
	}

	return kubernetesToYAML(items)
}

func flannelServiceAccount() corev1.ServiceAccount {
	account := corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flannel",
			Namespace: "kube-system",
		},
	}

	return account
}

func flannelClusterRole() rbac_v1beta1.ClusterRole {
	role := rbac_v1beta1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "flannel",
		},
		Rules: []rbac_v1beta1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes/status"},
				Verbs:     []string{"patch"},
			},
		},
	}

	return role
}

func flannelClusterRoleBinding() rbac_v1beta1.ClusterRoleBinding {
	role := rbac_v1beta1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "flannel",
		},
		RoleRef: rbac_v1beta1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "flannel",
		},
		Subjects: []rbac_v1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "flannel",
				Namespace: "kube-system",
			},
		},
	}

	return role
}

const flannel = `
# Downloaded from https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
# and 'quay.io' replaced with the mirror template stub
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: kube-flannel-cfg
  namespace: kube-system
  labels:
    tier: node
    app: flannel
data:
  cni-conf.json: |
    {
      "name": "cbr0",
      "plugins": [
        {
          "type": "flannel",
          "delegate": {
            "hairpinMode": true,
            "isDefaultGateway": true
          }
        },
        {
          "type": "portmap",
          "capabilities": {
            "portMappings": true
          }
        }
      ]
    }
  net-conf.json: |
    {
      "Network": "{{ .PodSubnet }}",
      "Backend": {
        "Type": "vxlan"
      }
    }
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: kube-flannel-ds
  namespace: kube-system
  labels:
    tier: node
    app: flannel
spec:
  template:
    metadata:
      labels:
        tier: node
        app: flannel
    spec:
      hostNetwork: true
      nodeSelector:
        beta.kubernetes.io/arch: amd64
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      serviceAccountName: flannel
      initContainers:
      - name: install-cni
        image: {{ .QuayMirror }}/coreos/flannel:v0.10.0-amd64
        command:
        - cp
        args:
        - -f
        - /etc/kube-flannel/cni-conf.json
        - /etc/cni/net.d/10-flannel.conflist
        volumeMounts:
        - name: cni
          mountPath: /etc/cni/net.d
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      containers:
      - name: kube-flannel
        image: {{ .QuayMirror }}/coreos/flannel:v0.10.0-amd64
        command:
        - /opt/bin/flanneld
        args:
        - --ip-masq
        - --kube-subnet-mgr
        resources:
          requests:
            cpu: "100m"
            memory: "50Mi"
          limits:
            cpu: "100m"
            memory: "50Mi"
        securityContext:
          privileged: true
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        volumeMounts:
        - name: run
          mountPath: /run
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      volumes:
        - name: run
          hostPath:
            path: /run
        - name: cni
          hostPath:
            path: /etc/cni/net.d
        - name: flannel-cfg
          configMap:
            name: kube-flannel-cfg
`
