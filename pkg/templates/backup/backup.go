/*
Copyright 2020 The KubeOne Authors.

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

package backup

import (
	"context"

	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/clientutil"
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	backupCronJobInterval = "@every 30m"

	snapshoterEtcdImage = "gcr.io/etcd-development/etcd:v3.4.3"
	uploaderResticImage = "docker.io/restic/restic:0.9.6"

	resticConfigSecretName        = "restic-config"
	resticConfigSecretNamespace   = "kube-system"
	resticConfigSecretPasswordKey = "password"
)

func Deploy(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes dynamic client is not initialized")
	}

	ctx := context.Background()

	k8sobjects := []runtime.Object{
		resticPasswordSecret(s.Cluster.Features.Backup.Config.ResticPassword),
		cronJob(s.Cluster.Features.Backup.Config),
	}

	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj); err != nil {
			return errors.WithStack(err)
		}
	}

	// HACK: re-init dynamic client in order to re-init RestMapper, to drop caches
	err := kubeconfig.HackIssue321InitDynamicClient(s)
	return errors.Wrap(err, "failed to re-init dynamic client")
}

func cronJob(backupConfig kubeone.BackupConfig) *batchv1beta1.CronJob {
	suspend := false
	var successfulJobsHistoryLimit int32 = 0
	var failedJobsHistoryLimit int32 = 1
	uploaderEnv := []corev1.EnvVar{
		{
			Name: "ETCD_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
				},
			},
		},
		{
			Name:  "RESTIC_REPOSITORY",
			Value: backupConfig.ResticRepository,
		},
		{
			Name: "RESTIC_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: resticConfigSecretName,
					},
					Key: resticConfigSecretPasswordKey,
				},
			},
		},
	}
	for envName, envVal := range backupConfig.Env {
		envVar := corev1.EnvVar{
			Name:  envName,
			Value: envVal,
		}
		uploaderEnv = append(uploaderEnv, envVar)
	}
	uploaderEnv = append(uploaderEnv, backupConfig.RepositoryCredentials...)

	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-s3-backup",
			Namespace: "kube-system",
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:                   backupCronJobInterval,
			ConcurrencyPolicy:          batchv1beta1.ForbidConcurrent,
			Suspend:                    &suspend,
			SuccessfulJobsHistoryLimit: &successfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     &failedJobsHistoryLimit,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							HostNetwork: true,
							NodeSelector: map[string]string{
								"node-role.kubernetes.io/master": "",
							},
							Tolerations: []corev1.Toleration{
								{
									Key:      "node-role.kubernetes.io/master",
									Effect:   corev1.TaintEffectNoSchedule,
									Operator: corev1.TolerationOpExists,
								},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure,
							Volumes: []corev1.Volume{
								{
									Name: "etcd-backup",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{},
									},
								},
								{
									Name: "host-etcd-pki",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/etc/kubernetes/pki/etcd",
										},
									},
								},
							},
							InitContainers: []corev1.Container{
								{
									Name:            "snapshoter",
									Image:           snapshoterEtcdImage,
									ImagePullPolicy: corev1.PullIfNotPresent,
									Command:         []string{"etcdctl"},
									Args:            []string{"snapshot", "save", "/backup/$(ETCD_HOSTNAME)-snapshot.db"},
									Env: []corev1.EnvVar{
										{
											Name:  "ETCDCTL_API",
											Value: "3",
										},
										{
											Name:  "ETCDCTL_DIAL_TIMEOUT",
											Value: "3s",
										},
										{
											Name:  "ETCDCTL_ENDPOINTS",
											Value: "127.0.0.1",
										},
										{
											Name:  "ETCDCTL_CACERT",
											Value: "/etc/kubernetes/pki/etcd/ca.crt",
										},
										{
											Name:  "ETCDCTL_CERT",
											Value: "/etc/kubernetes/pki/etcd/healthcheck-client.crt",
										},
										{
											Name:  "ETCDCTL_KEY",
											Value: "/etc/kubernetes/pki/etcd/healthcheck-client.key",
										},
										{
											Name: "ETCD_HOSTNAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "spec.nodeName",
												},
											},
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "etcd-backup",
											MountPath: "/backup",
										},
										{
											Name:      "host-etcd-pki",
											MountPath: "/etc/kubernetes/pki/etcd",
											ReadOnly:  true,
										},
									},
								},
							},
							Containers: []corev1.Container{
								{
									Name:            "uploader",
									Image:           uploaderResticImage,
									ImagePullPolicy: corev1.PullIfNotPresent,
									Command: []string{
										"/bin/sh",
										"-c",
										"|-",
										"set -euf",
										"restic snapshots -q || restic init -q",
										"restic backup --tag=etcd --host=${ETCD_HOSTNAME} /backup",
										"restic forget --prune --keep-last 48",
									},
									Env: uploaderEnv,
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "etcd-backup",
											MountPath: "/backup",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resticPasswordSecret(password string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resticConfigSecretName,
			Namespace: resticConfigSecretNamespace,
		},
		StringData: map[string]string{
			resticConfigSecretPasswordKey: password,
		},
		Type: corev1.SecretTypeOpaque,
	}
}
