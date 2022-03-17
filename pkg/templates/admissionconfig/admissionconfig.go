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

package admissionconfig

import (
	"github.com/Masterminds/semver/v3"

	apiserverv1 "k8c.io/kubeone/pkg/apis/apiserver/v1"
	apiserverv1alpha1 "k8c.io/kubeone/pkg/apis/apiserver/v1alpha1"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewAdmissionConfig generates the AdmissionConfiguration manifest
func NewAdmissionConfig(k8sVersion string, podNodeSelectorFeature *kubeoneapi.PodNodeSelector) (string, error) {
	sver, err := semver.NewVersion(k8sVersion)
	if err != nil {
		return "", fail.Runtime(err, "parsing kubernetes semver")
	}
	c, err := semver.NewConstraint("< 1.17.0")
	if err != nil {
		return "", fail.Runtime(err, "parsing semver constraint")
	}

	var admissionCfg []runtime.Object
	switch {
	case c.Check(sver):
		admissionCfg = admissionConfigV1alpha1(podNodeSelectorFeature)
	default:
		admissionCfg = admissionConfigV1(podNodeSelectorFeature)
	}

	return templates.KubernetesToYAML(admissionCfg)
}

func admissionConfigV1(podNodeSelectorFeature *kubeoneapi.PodNodeSelector) []runtime.Object {
	admissionConfig := &apiserverv1.AdmissionConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiserver.config.k8s.io/v1",
			Kind:       "AdmissionConfiguration",
		},
	}

	if podNodeSelectorFeature != nil && podNodeSelectorFeature.Enable {
		pnsPlugin := apiserverv1.AdmissionPluginConfiguration{
			Name: "PodNodeSelector",
			Path: "/etc/kubernetes/admission/podnodeselector.yaml",
		}
		admissionConfig.Plugins = append(admissionConfig.Plugins, pnsPlugin)
	}

	return []runtime.Object{admissionConfig}
}

func admissionConfigV1alpha1(podNodeSelectorFeature *kubeoneapi.PodNodeSelector) []runtime.Object {
	admissionConfig := &apiserverv1alpha1.AdmissionConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiserver.k8s.io/v1alpha1",
			Kind:       "AdmissionConfiguration",
		},
	}

	if podNodeSelectorFeature != nil && podNodeSelectorFeature.Enable {
		pnsPlugin := apiserverv1alpha1.AdmissionPluginConfiguration{
			Name: "PodNodeSelector",
			Path: "/etc/kubernetes/admission/podnodeselector.yaml",
		}
		admissionConfig.Plugins = append(admissionConfig.Plugins, pnsPlugin)
	}

	return []runtime.Object{admissionConfig}
}
