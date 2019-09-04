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

package features

import (
	"strconv"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/kubeadmargs"
)

const (
	auditPolicyFileFlag   = "audit-policy-file"
	auditLogPathFlag      = "audit-log-path"
	auditLogMaxAgeFlag    = "audit-log-maxage"
	auditLogMaxBackupFlag = "audit-log-maxbackup"
	auditLogMaxSizeFlag   = "audit-log-maxsize"
)

func activateKubeadmStaticAuditLogs(feature *kubeoneapi.StaticAuditLog, args *kubeadmargs.Args) {
	if feature == nil || !feature.Enable {
		return
	}

	args.APIServer.ExtraArgs[auditPolicyFileFlag] = "/etc/kubernetes/audit/policy.yaml"
	args.APIServer.ExtraArgs[auditLogPathFlag] = feature.Config.LogPath
	args.APIServer.ExtraArgs[auditLogMaxAgeFlag] = strconv.Itoa(feature.Config.LogMaxAge)
	args.APIServer.ExtraArgs[auditLogMaxBackupFlag] = strconv.Itoa(feature.Config.LogMaxBackup)
	args.APIServer.ExtraArgs[auditLogMaxSizeFlag] = strconv.Itoa(feature.Config.LogMaxSize)
}
