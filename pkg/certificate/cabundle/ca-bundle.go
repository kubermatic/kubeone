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

package cabundle

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	OriginalCertsDir = "/etc/ssl/certs"
	CustomCertsDir   = "/etc/kubeone/certs"
	FileName         = "ca-certificates.crt"
	SSLCertFilePath  = CustomCertsDir + "/" + FileName
	ConfigMapName    = "ca-bundle"

	SSLCertFileENV = "SSL_CERT_FILE"
)

func Inject(caBundle string, podTpl *corev1.PodTemplateSpec) {
	if caBundle == "" {
		return
	}

	if podTpl.Annotations == nil {
		podTpl.Annotations = map[string]string{}
	}

	hsh := sha256.New()
	hsh.Write([]byte(caBundle))
	podTpl.Annotations["caBundle-hash"] = fmt.Sprintf("%x", hsh.Sum(nil))
	podTpl.Spec.Volumes = append(podTpl.Spec.Volumes, Volume())

	for idx := range podTpl.Spec.Containers {
		cont := podTpl.Spec.Containers[idx]
		cont.VolumeMounts = append(cont.VolumeMounts, VolumeMount())
		cont.Env = append(cont.Env, EnvVar())
		podTpl.Spec.Containers[idx] = cont
	}
}

func ConfigMap(caBundle string) *corev1.ConfigMap {
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			FileName: caBundle,
		},
	}

	return &cm
}

func VolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      ConfigMapName,
		ReadOnly:  true,
		MountPath: CustomCertsDir,
	}
}

func Volume() corev1.Volume {
	return corev1.Volume{
		Name: ConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ConfigMapName,
				},
			},
		},
	}
}

func EnvVar() corev1.EnvVar {
	return corev1.EnvVar{
		Name:  SSLCertFileENV,
		Value: filepath.Join(CustomCertsDir, FileName),
	}
}
