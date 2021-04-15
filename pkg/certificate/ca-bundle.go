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

package certificate

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

const (
	CACertsDir            = "/etc/kubeone/certs"
	CABundleFile          = "ca-certificates.crt"
	CABundlePath          = CACertsDir + "/" + CABundleFile
	CABundleConfigMapName = "ca-bundle"

	SSLCertFileENV = "SSL_CERT_FILE"
)

func CABundleInjector(caBundle string, podTpl *corev1.PodTemplateSpec) {
	if caBundle == "" {
		return
	}

	if podTpl.Annotations == nil {
		podTpl.Annotations = map[string]string{}
	}

	hsh := sha256.New()
	hsh.Write([]byte(caBundle))
	podTpl.Annotations["caBundle-hash"] = fmt.Sprintf("%x", hsh.Sum(nil))

	podTpl.Spec.Volumes = append(podTpl.Spec.Volumes, corev1.Volume{
		Name: CABundleConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: CABundleConfigMapName,
				},
			},
		},
	})

	for idx := range podTpl.Spec.Containers {
		cont := podTpl.Spec.Containers[idx]
		cont.VolumeMounts = append(cont.VolumeMounts, corev1.VolumeMount{
			Name:      CABundleConfigMapName,
			ReadOnly:  true,
			MountPath: CACertsDir,
		})
		cont.Env = append(cont.Env, corev1.EnvVar{
			Name:  SSLCertFileENV,
			Value: filepath.Join(CACertsDir, CABundleFile),
		})
		podTpl.Spec.Containers[idx] = cont
	}
}
