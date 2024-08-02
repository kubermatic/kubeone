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

package features

import (
	"fmt"
	"strconv"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"
)

const (
	auditWebhookConfigFileFlag           = "audit-webhook-config-file"
	auditWebhookBatchBufferSizeFlag      = "audit-webhook-batch-buffer-size"
	auditWebhookBatchMaxSizeFlag         = "audit-webhook-batch-max-size"
	auditWebhookBatchMaxWaitFlag         = "audit-webhook-batch-max-wait"
	auditWebhookThrottleBurstFlag        = "audit-webhook-batch-throttle-burst"
	auditWebhookThrottleEnableFlag       = "audit-webhook-batch-throttle-enable"
	auditWebhookThrottleQPSFlag          = "audit-webhook-batch-throttle-qps"
	auditWebhookInitialBackoffFlag       = "audit-webhook-initial-backoff"
	auditWebhookModeFlag                 = "audit-webhook-mode"
	auditWebhookTruncateEnabledFlag      = "audit-webhook-truncate-enabled"
	auditWebhookTruncateMaxBatchSizeFlag = "audit-webhook-truncate-max-batch-size"
	auditWebhookTruncateMaxEventSizeFlag = "audit-webhook-truncate-max-event-size"
	auditWebhookWebhookVersionFlag       = "audit-webhook-version"
)

func activateKubeadmWebhookAuditLogs(feature *kubeoneapi.WebhookAuditLog, args *kubeadmargs.Args) {
	if feature == nil || !feature.Enable {
		return
	}

	args.APIServer.ExtraArgs[auditPolicyFileFlag] = "/etc/kubernetes/audit/policy.yaml"
	args.APIServer.ExtraArgs[auditWebhookConfigFileFlag] = "/etc/kubernetes/audit/webhook-config.yaml"

	if feature.Config.Mode != "" {
		args.APIServer.ExtraArgs[auditWebhookModeFlag] = string(feature.Config.Mode)
	}

	if feature.Config.Version != "" {
		args.APIServer.ExtraArgs[auditWebhookWebhookVersionFlag] = feature.Config.Version
	}

	// exit early if no batch options are configured
	if b := feature.Config.Batch; b != (kubeoneapi.WebHookAuditLogBatchConfig{}) {
		if b.BufferSize != 0 {
			args.APIServer.ExtraArgs[auditWebhookBatchBufferSizeFlag] = strconv.Itoa(b.BufferSize)
		}
		if b.MaxSize != 0 {
			args.APIServer.ExtraArgs[auditWebhookBatchMaxSizeFlag] = strconv.Itoa(b.MaxSize)
		}
		if b.MaxWait.Duration != 0 {
			args.APIServer.ExtraArgs[auditWebhookBatchMaxWaitFlag] = b.MaxWait.Duration.String()
		}

		if t := b.Throttle; t != (kubeoneapi.WebHookAuditLogThrottleConfig{}) {
			if !t.Disable { // default is enable=true, so we need to check for the inverse https://github.com/kubernetes/kubernetes/blob/2a1d4172e22abb6759b3d2ad21bb09a04eef596d/staging/src/k8s.io/apiserver/pkg/server/options/audit.go#L593
				args.APIServer.ExtraArgs[auditWebhookThrottleEnableFlag] = strconv.FormatBool(t.Disable)
			}
			if t.Burst != 0 {
				args.APIServer.ExtraArgs[auditWebhookThrottleBurstFlag] = strconv.Itoa(t.Burst)
			}
			if t.QPS != 0 {
				args.APIServer.ExtraArgs[auditWebhookThrottleQPSFlag] = fmt.Sprintf("%f", t.QPS)
			}
		}
	}

	if t := feature.Config.Truncate; t != (kubeoneapi.WebHookAuditLogTruncateConfig{}) {
		if t.Enable { // default is false https://github.com/kubernetes/kubernetes/blob/2a1d4172e22abb6759b3d2ad21bb09a04eef596d/staging/src/k8s.io/apiserver/pkg/server/options/audit.go#L170
			args.APIServer.ExtraArgs[auditWebhookTruncateEnabledFlag] = strconv.FormatBool(t.Enable)
		}
		if t.MaxBatchSize != 0 {
			args.APIServer.ExtraArgs[auditWebhookTruncateMaxBatchSizeFlag] = strconv.Itoa(t.MaxBatchSize)
		}
		if t.MaxEventSize != 0 {
			args.APIServer.ExtraArgs[auditWebhookTruncateMaxEventSizeFlag] = strconv.Itoa(t.MaxEventSize)
		}
	}
}
