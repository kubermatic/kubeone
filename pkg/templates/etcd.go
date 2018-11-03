package templates

import (
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

func hostPathTypePtr(s corev1.HostPathType) *corev1.HostPathType {
	return &s
}

func EtcdConfig(manifest *manifest.Manifest, instance int) (string, error) {
	masterNodes := manifest.Hosts
	if len(masterNodes) < (instance - 1) {
		return "", fmt.Errorf("manifest does not contain node #%d", instance)
	}

	token, err := manifest.EtcdClusterToken()
	if err != nil {
		return "", errors.New("failed to generate new secure etcd cluster token")
	}

	node := masterNodes[instance]
	name := fmt.Sprintf("etcd-%d", instance)
	etcdRing := make([]string, 0)

	for i, node := range masterNodes {
		etcdRing = append(etcdRing, fmt.Sprintf("etcd-%d=%s", i, node.EtcdPeerUrl()))
	}

	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd",
			Namespace: "kube-system",
			Annotations: map[string]string{
				"scheduler.alpha.kubernetes.io/critical-pod": "",
			},
			Labels: map[string]string{
				"component": "etcd",
				"tier":      "control-plane",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "etcd",
					Command: []string{
						"etcd",
						"--data-dir=/var/lib/etcd",
						fmt.Sprintf("--name=%s", name),
						fmt.Sprintf("--advertise-client-urls=%s", node.EtcdUrl()),
						fmt.Sprintf("--listen-client-urls=%s", node.EtcdUrl()),
						fmt.Sprintf("--listen-peer-urls=%s", node.EtcdPeerUrl()),

						"--initial-cluster-state=new",
						fmt.Sprintf("--initial-advertise-peer-urls=%s", node.EtcdPeerUrl()),
						fmt.Sprintf("--initial-cluster=%s", strings.Join(etcdRing, ",")),
						fmt.Sprintf("--initial-cluster-token=%s", token),

						"--client-cert-auth=true",
						"--cert-file=/etc/kubernetes/pki/etcd/server.crt",
						"--key-file=/etc/kubernetes/pki/etcd/server.key",
						"--trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",

						"--peer-client-cert-auth=true",
						"--peer-cert-file=/etc/kubernetes/pki/etcd/peer.crt",
						"--peer-key-file=/etc/kubernetes/pki/etcd/peer.key",
						"--peer-trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",
					},
					Image: fmt.Sprintf("k8s.gcr.io/etcd-amd64:%s", manifest.Versions.Etcd()),
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/var/lib/etcd",
							Name:      "etcd-data",
						},
						{
							MountPath: "/etc/kubernetes/pki/etcd",
							Name:      "etcd-certs",
						},
					},
				},
			},
			HostNetwork: true,
			Volumes: []corev1.Volume{
				{
					Name: "etcd-data",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/etcd",
							Type: hostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "etcd-certs",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/kubernetes/pki/etcd",
							Type: hostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
			},
		},
	}

	return kubernetesToYAML([]interface{}{pod})
}
