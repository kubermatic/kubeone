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

package weave

import (
	"crypto/rand"
	"encoding/base64"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnsureSecret ensure weave-net Secret with password exists
func EnsureSecret(s *state.State) error {
	pass, err := genPassword()
	if err != nil {
		return err
	}

	sec := weaveSecret(pass)
	key := client.ObjectKey{
		Name:      sec.GetName(),
		Namespace: sec.GetNamespace(),
	}
	secCopy := sec.DeepCopy()

	err = s.DynamicClient.Get(s.Context, key, secCopy)
	if k8serrors.IsNotFound(err) {
		err = s.DynamicClient.Create(s.Context, sec)

		return fail.KubeClient(err, "creating %T %s", sec, key)
	}

	return fail.KubeClient(err, "getting %T %s", sec, key)
}

func genPassword() (string, error) {
	pi := make([]byte, 32)
	_, err := rand.Reader.Read(pi)
	if err != nil {
		return "", fail.Runtime(err, "reading random bytes")
	}

	return base64.StdEncoding.EncodeToString(pi), nil
}

func weaveSecret(pass string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-passwd",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"name": "weave-net",
			},
		},
		StringData: map[string]string{
			"weave-passwd": pass,
		},
	}
}
