package etcd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kubermatic/kubeone/pkg/config"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	DataVolumeName   = "etcd-data"
	BackupVolumeName = "etcd-backup"
	ContainerImage   = "k8s.gcr.io/etcd-amd64"
)

func hostPathTypePtr(s corev1.HostPathType) *corev1.HostPathType {
	return &s
}

// Pod returns static etcd manifest to YAML
func Pod(cluster *config.Cluster, node *config.HostConfig) (*corev1.Pod, error) {
	initialCluster := make([]string, len(cluster.Hosts))
	for i, host := range cluster.Hosts {
		initialCluster[i] = fmt.Sprintf("%s=http://%s:2380", host.Hostname, host.PrivateAddress)
	}
	initialClusterString := strings.Join(initialCluster, ",")

	image := fmt.Sprintf("%s:%s", ContainerImage, cluster.ETCD.Version)

	backupCmd, err := backupCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to create backup command: %v", err)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd",
			Namespace: "kube-system",
			Annotations: map[string]string{
				"scheduler.alpha.kubernetes.io/critical-pod": "",

				// configure backup hook
				"backup.ark.heptio.com/backup-volumes":     BackupVolumeName,
				"pre.hook.backup.ark.heptio.com/container": "etcd",
				"pre.hook.backup.ark.heptio.com/timeout":   "10m",
				"pre.hook.backup.ark.heptio.com/command":   backupCmd,
			},
			Labels: map[string]string{
				"component": "etcd",
				"tier":      "control-plane",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "etcd",
					Image: image,
					Command: []string{
						"etcd",
						fmt.Sprintf("--advertise-client-urls=http://%s:2379", node.PrivateAddress),
						fmt.Sprintf("--initial-advertise-peer-urls=http://%s:2380", node.PrivateAddress),
						fmt.Sprintf("--initial-cluster=%s", initialClusterString),
						"--initial-cluster-state=new",
						fmt.Sprintf("--listen-client-urls=http://127.0.0.1:2379,http://%s:2379", node.PrivateAddress),
						fmt.Sprintf("--listen-peer-urls=http://%s:2380", node.PrivateAddress),
						"--data-dir=/var/lib/etcd",
						fmt.Sprintf("--name=%s", node.Hostname),
						"--snapshot-count=10000",
						fmt.Sprintf("--initial-cluster-token=%s", cluster.EtcdClusterToken()),
					},
					ImagePullPolicy: corev1.PullIfNotPresent,
					LivenessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							Exec: &corev1.ExecAction{
								Command: []string{
									"/bin/sh",
									"-c",
									"ETCDCTL_API=3 etcdctl --endpoints=http://127.0.0.1:2379 get foo",
								},
							},
						},
						FailureThreshold:    8,
						InitialDelaySeconds: 15,
						TimeoutSeconds:      15,
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      DataVolumeName,
							MountPath: "/var/lib/etcd",
						},
						{
							Name:      BackupVolumeName,
							MountPath: "/backup",
						},
					},
				},
			},
			HostNetwork:       true,
			PriorityClassName: "system-cluster-critical",
			Volumes: []corev1.Volume{
				{
					Name: DataVolumeName,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/etcd",
							Type: hostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: BackupVolumeName,
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	gvkList, _, err := scheme.Scheme.ObjectKinds(pod)
	if err != nil {
		return nil, fmt.Errorf("failed to get GroupVersionKind: %v", err)
	}
	if len(gvkList) < 1 {
		return nil, fmt.Errorf("no GroupVersionKind found: %v", err)
	}
	apiVersion, kind := gvkList[0].ToAPIVersionAndKind()
	pod.APIVersion = apiVersion
	pod.Kind = kind
	return pod, nil
}

func backupCommand() (string, error) {
	backup := strings.Join([]string{
		"ETCDCTL_API=3",
		"etcdctl",
		"--endpoints=http://127.0.0.1:2379",
		"snapshot", "save", "/backup/snapshot.db",
	}, " ")

	command := []string{
		"/bin/sh",
		"-c",
		backup,
	}

	encoded, err := json.Marshal(command)

	return string(encoded), err
}
