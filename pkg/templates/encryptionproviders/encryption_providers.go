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

package encryptionproviders

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"k8c.io/kubeone/pkg/state"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	apiserverconfigv1 "k8s.io/apiserver/pkg/apis/config/v1"
)

func generateAESCBCSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Reader.Read(buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

func NewEncyrptionProvidersConfig(s *state.State) (*apiserverconfigv1.EncryptionConfiguration, error) {
	secret, err := generateAESCBCSecret()
	if err != nil {
		return nil, err
	}
	return &apiserverconfigv1.EncryptionConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiserver.config.k8s.io/v1",
			Kind:       "EncryptionConfiguration",
		},
		Resources: []apiserverconfigv1.ResourceConfiguration{
			{
				Resources: []string{"secrets"},
				Providers: []apiserverconfigv1.ProviderConfiguration{
					{
						AESCBC: &apiserverconfigv1.AESConfiguration{
							Keys: []apiserverconfigv1.Key{
								{
									Name:   fmt.Sprintf("kubeone-%s", utilrand.String(6)),
									Secret: secret,
								},
							},
						},
					},
					{
						Identity: &apiserverconfigv1.IdentityConfiguration{},
					},
				},
			},
		},
	}, nil
}

func UpdateEncryptionConfigDecryptOnly(config *apiserverconfigv1.EncryptionConfiguration) error {
	if config.Resources[0].Providers[0].AESCBC == nil {
		return errors.New("empty AESCBC key configuration")
	}

	config.Resources[0].Providers = []apiserverconfigv1.ProviderConfiguration{
		{
			Identity: &apiserverconfigv1.IdentityConfiguration{},
		},
		{
			AESCBC: config.Resources[0].Providers[0].AESCBC,
		},
	}
	return nil
}

func UpdateEncryptionConfigWithNewKey(config *apiserverconfigv1.EncryptionConfiguration) error {
	secret, err := generateAESCBCSecret()
	if err != nil {
		return err
	}
	config.Resources[0].Providers = []apiserverconfigv1.ProviderConfiguration{
		{
			AESCBC: &apiserverconfigv1.AESConfiguration{
				Keys: []apiserverconfigv1.Key{
					{
						Name:   fmt.Sprintf("kubeone-%s", utilrand.String(6)),
						Secret: secret,
					},

					config.Resources[0].Providers[0].AESCBC.Keys[len(config.Resources[0].Providers[0].AESCBC.Keys)-1],
				},
			},
		},
		{
			Identity: &apiserverconfigv1.IdentityConfiguration{},
		},
	}
	return nil
}

func UpdateEncryptionConfigRemoveOldKey(config *apiserverconfigv1.EncryptionConfiguration) {
	config.Resources[0].Providers = []apiserverconfigv1.ProviderConfiguration{
		{
			AESCBC: &apiserverconfigv1.AESConfiguration{
				Keys: []apiserverconfigv1.Key{
					config.Resources[0].Providers[0].AESCBC.Keys[0],
				},
			},
		},
		{
			Identity: &apiserverconfigv1.IdentityConfiguration{},
		},
	}
}
