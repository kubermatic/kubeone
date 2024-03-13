/*
Copyright 2024 The KubeOne Authors.

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

package kubernetesconfigs

import "crypto/tls"

// This list is produces according to CIS 1.8 / 1.2.30
//
// See more: https://github.com/aquasecurity/kube-bench/blob/v0.7.2/cfg/cis-1.8/master.yaml#L768-L788
func APIServerDefaultTLSCipherSuites() []string {
	return []string{
		tls.CipherSuiteName(tls.TLS_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_AES_256_GCM_SHA384),
		tls.CipherSuiteName(tls.TLS_CHACHA20_POLY1305_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256),

		// Followin cipher suites considered insecure, however they are included in the CIS list and without them AWS LB
		// healthcheck client doesn't work
		tls.CipherSuiteName(tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_128_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_256_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_256_GCM_SHA384),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA),
	}
}

// This list is produces according to CIS 1.8 / 4.2.12
//
// TLS_RSA_WITH_AES_256_GCM_SHA384 and TLS_RSA_WITH_AES_128_GCM_SHA256 excluded from the list as insecure.
// See more: https://github.com/aquasecurity/kube-bench/blob/v0.7.2/cfg/cis-1.8/node.yaml#L420-L442
func DefaultTLSCipherSuites() []string {
	return []string{
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384),

		// Following to removed since they are actually insecure
		// tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_256_GCM_SHA384),
		// tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_128_GCM_SHA256),
	}
}

func TLSCipherSuites(cipherSuites []*tls.CipherSuite) []string {
	result := make([]string, 0, len(cipherSuites))

	for _, cs := range cipherSuites {
		result = append(result, tls.CipherSuiteName(cs.ID))
	}

	return result
}
