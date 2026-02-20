/*
Copyright 2022 The KubeOne Authors.

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

package initcmd

import (
	"reflect"
	"strings"
)

type modelCluster struct {
	ClusterName       string
	CloudProvider     string
	KubernetesVersion string
	CNI               string
	Features          []string
}

type modelAddonAutoscaler struct {
	MinReplicas string `tfvar:"cluster_autoscaler_min_replicas"`
	MaxReplicas string `tfvar:"cluster_autoscaler_max_replicas"`
}

type modelAddonBackups struct {
	ResticPassword   string
	S3Bucket         string
	AWSDefaultRegion string
}

type modelTerraformVars struct {
	SSHPublicKeyPath  string `tfvar:"ssh_public_key_file"`
	ControlPlaneCount string `tfvar:"control_plane_vm_count"`
	WorkerNodesCount  string `tfvar:"initial_machinedeployment_replicas"`
}

func tfvarName(obj any, fieldName string) string {
	elem := reflect.TypeOf(obj).Elem()
	field, ok := elem.FieldByName(fieldName)
	if !ok {
		return strings.ToLower(fieldName)
	}

	return field.Tag.Get("tfvar")
}
