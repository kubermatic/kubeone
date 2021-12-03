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

package v1beta1

import (
	kubeone "k8c.io/kubeone/pkg/apis/kubeone"

	conversion "k8s.io/apimachinery/pkg/conversion"
)

// Convert_v1beta1_Features_To_kubeone_Features is an autogenerated conversion function.
func Convert_v1beta1_Features_To_kubeone_Features(in *Features, out *kubeone.Features, s conversion.Scope) error {
	if err := autoConvert_v1beta1_Features_To_kubeone_Features(in, out, s); err != nil {
		return err
	}
	// The PodPresets field has been dropped from v1beta2 API.
	return nil
}
