/*
Copyright 2019 The KubeOne Authors.

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

package scheme

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// Scheme is a runtime.Scheme object to which all KubeOne API types are registered
var Scheme = runtime.NewScheme()

// Codecs is a CodecFactory object used to provide encoding and decoding for the scheme
var Codecs = serializer.NewCodecFactory(Scheme, serializer.EnableStrict)

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	AddToScheme(Scheme)
}

// AddToScheme builds the KubeOne scheme
func AddToScheme(scheme *runtime.Scheme) {
	utilruntime.Must(kubeoneapi.AddToScheme(scheme))
	utilruntime.Must(kubeonev1beta2.AddToScheme(scheme))
	utilruntime.Must(kubeonev1beta1.AddToScheme(scheme))
	utilruntime.Must(scheme.SetVersionPriority(kubeonev1beta1.SchemeGroupVersion))
}
