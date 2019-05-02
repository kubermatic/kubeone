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

package config

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/kubermatic/kubeone/pkg/yamled"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// MigrateToKubeOneClusterAPI migrates the old API to the new KubeOneCluster API
func MigrateToKubeOneClusterAPI(oldConfigPath string) (interface{}, error) {
	oldConfig, err := loadClusterConfig(oldConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse the old config")
	}

	oldConfig.Set(yamled.Path{"apiVersion"}, "kubeone.io/v1alpha1")
	oldConfig.Set(yamled.Path{"kind"}, "KubeOneCluster")

	// basic root-level renames
	rename(oldConfig, yamled.Path{}, "apiserver", "apiEndpoints")
	rename(oldConfig, yamled.Path{}, "provider", "cloudProvider")
	rename(oldConfig, yamled.Path{}, "network", "clusterNetwork")
	rename(oldConfig, yamled.Path{}, "machine_controller", "machineController")

	// camel-casing host fields
	hosts, exists := oldConfig.GetArray(yamled.Path{"hosts"})
	if exists {
		total := len(hosts)

		for i := 0; i < total; i++ {
			path := yamled.Path{"hosts", i}

			rename(oldConfig, path, "public_address", "publicAddress")
			rename(oldConfig, path, "private_address", "privateAddress")
			rename(oldConfig, path, "ssh_port", "sshPort")
			rename(oldConfig, path, "ssh_username", "sshUsername")
			rename(oldConfig, path, "ssh_private_key_file", "sshPrivateKeyFile")
			rename(oldConfig, path, "ssh_agent_socket", "sshAgentSocket")
		}
	}

	// separating host and port for api endpoints, turn it into an array
	apiserver, exists := oldConfig.GetString(yamled.Path{"apiEndpoints", "address"})
	if exists {
		host, sport, err := net.SplitHostPort(apiserver)
		if err != nil {
			host = apiserver
			sport = "6443"
		}

		port, err := strconv.Atoi(sport)
		if err != nil {
			return yaml.MapSlice{}, fmt.Errorf("invalid port specified for API server: %s", port)
		}

		oldConfig.Remove(yamled.Path{"apiEndpoints"})
		oldConfig.Set(yamled.Path{"apiEndpoints"}, []map[string]interface{}{
			map[string]interface{}{
				"host": host,
				"port": port,
			},
		})
	}

	// camel-casing cloudConfig
	rename(oldConfig, yamled.Path{"cloudProvider"}, "cloud_config", "cloudConfig")

	// camel-casing clusterNetwork
	path := yamled.Path{"clusterNetwork"}
	rename(oldConfig, path, "pod_subnet", "podSubnet")
	rename(oldConfig, path, "service_subnet", "serviceSubnet")
	rename(oldConfig, path, "node_port_range", "nodePortRange")

	// camel-casing proxy
	path = yamled.Path{"proxy"}
	rename(oldConfig, path, "http_proxy", "http")
	rename(oldConfig, path, "https_proxy", "https")
	rename(oldConfig, path, "no_proxy", "noProxy")

	// move machine-controller credentials to root level
	credentials, exists := oldConfig.Get(yamled.Path{"machineController", "credentials"})
	if exists {
		oldConfig.Remove(yamled.Path{"machineController", "credentials"})
		oldConfig.Set(yamled.Path{"credentials"}, credentials)
	}

	// camel-casing features
	path = yamled.Path{"features"}
	rename(oldConfig, path, "pod_security_policy", "podSecurityPolicy")
	rename(oldConfig, path, "dynamic_audit_log", "dynamicAuditLog")
	rename(oldConfig, path, "metrics_server", "metricsServer")
	rename(oldConfig, path, "openid_connect", "openidConnect")

	// camel-casing openidConnect
	path = yamled.Path{"features", "openidConnect", "config"}
	rename(oldConfig, path, "issuer_url", "issuerUrl")
	rename(oldConfig, path, "client_id", "clientId")
	rename(oldConfig, path, "username_claim", "usernameClaim")
	rename(oldConfig, path, "username_prefix", "usernamePrefix")
	rename(oldConfig, path, "groups_claim", "groupsClaim")
	rename(oldConfig, path, "groups_prefix", "groupsPrefix")
	rename(oldConfig, path, "signing_algs", "signingAlgs")
	rename(oldConfig, path, "required_claim", "requiredClaim")
	rename(oldConfig, path, "ca_file", "caFile")

	// rename workers.config to providerSpec
	workers, exists := oldConfig.GetArray(yamled.Path{"workers"})
	if exists {
		total := len(workers)

		for i := 0; i < total; i++ {
			rename(oldConfig, yamled.Path{"workers", i}, "config", "providerSpec")
		}
	}

	return oldConfig.Root(), nil
}

func loadClusterConfig(oldConfigPath string) (*yamled.Document, error) {
	f, err := os.Open(oldConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer f.Close()

	return yamled.Load(f)
}

func rename(doc *yamled.Document, basePath yamled.Path, oldKey string, newKey string) {
	oldPath := append(basePath, oldKey)
	newPath := append(basePath, newKey)

	data, exists := doc.Get(oldPath)
	if exists {
		doc.Remove(oldPath)
		doc.Set(newPath, data)
	}
}
