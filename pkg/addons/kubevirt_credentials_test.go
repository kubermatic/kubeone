/*
Copyright 2026 The KubeOne Authors.

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

package addons

import (
	"context"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubevirtmc "k8c.io/kubeone/pkg/cloudprovider/kubevirt/machinecontroller"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMigrateKubeVirtCredentials(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	ns := credentials.SecretNamespace
	legacy := []client.Object{
		&corev1.Secret{Name: credentials.SecretNameCCM, Namespace: ns},
		&corev1.Secret{Name: credentials.SecretNameMC, Namespace: ns},
		&corev1.Secret{Name: credentials.SecretNameOSM, Namespace: ns},
	}
	kept := []client.Object{
		&corev1.Secret{Name: credentials.CloudConfigSecretName, Namespace: ns},
		&corev1.Secret{Name: kubevirtmc.ControllerTenantTokenSecretName, Namespace: ns},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(append(legacy, kept...)...).Build()
	s := &state.State{Context: context.Background(), DynamicClient: c}

	for range 2 { // idempotent
		if err := migrateKubeVirtCredentials(s); err != nil {
			t.Fatal(err)
		}
	}

	for _, obj := range legacy {
		if err := c.Get(s.Context, client.ObjectKeyFromObject(obj), obj); !apierrors.IsNotFound(err) {
			t.Errorf("%T %s: expected NotFound, got %v", obj, obj.GetName(), err)
		}
	}

	for _, obj := range kept {
		if err := c.Get(s.Context, client.ObjectKeyFromObject(obj), obj); err != nil {
			t.Errorf("%T %s: expected to be kept, got %v", obj, obj.GetName(), err)
		}
	}
}

func TestMigrateKubeVirtCredentialsLegacyCSI(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	ns := credentials.SecretNamespace
	ccm := &corev1.Secret{Name: credentials.SecretNameCCM, Namespace: ns}
	mc := &corev1.Secret{Name: credentials.SecretNameMC, Namespace: ns}
	// left behind with .cloudProvider.disableBundledCSIDrivers, mounts the CCM Secret
	legacyCSI := &appsv1.Deployment{Name: kubevirtLegacyCSIControllerName, Namespace: metav1.NamespaceSystem}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ccm, mc, legacyCSI).Build()
	logger, hook := logrustest.NewNullLogger()
	s := &state.State{Context: context.Background(), DynamicClient: c, Logger: logger}

	if err := migrateKubeVirtCredentials(s); err != nil {
		t.Fatal(err)
	}

	if err := c.Get(s.Context, client.ObjectKeyFromObject(ccm), ccm); err != nil {
		t.Errorf("CCM Secret must be kept while the legacy CSI controller exists, got %v", err)
	}
	if err := c.Get(s.Context, client.ObjectKeyFromObject(mc), mc); !apierrors.IsNotFound(err) {
		t.Errorf("machine-controller Secret: expected NotFound, got %v", err)
	}
	if e := hook.LastEntry(); e == nil || e.Level != logrus.WarnLevel || !strings.Contains(e.Message, credentials.SecretNameCCM) {
		t.Errorf("expected a warning about the kept CCM Secret, got %v", e)
	}

	// once the legacy CSI controller is gone, the CCM Secret is removed too
	if err := c.Delete(s.Context, legacyCSI); err != nil {
		t.Fatal(err)
	}
	if err := migrateKubeVirtCredentials(s); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(s.Context, client.ObjectKeyFromObject(ccm), ccm); !apierrors.IsNotFound(err) {
		t.Errorf("CCM Secret: expected NotFound, got %v", err)
	}
}

func TestCredentialsTemplateData(t *testing.T) {
	t.Setenv(credentials.KubevirtKubeconfigKey, "apiVersion: v1\nkind: Config\n")
	t.Setenv(credentials.HetznerTokenKey, "token")

	tests := []struct {
		name          string
		cloudProvider kubeoneapi.CloudProviderSpec
		wantSecret    bool
	}{
		{
			name:          "kubevirt",
			cloudProvider: kubeoneapi.CloudProviderSpec{Kubevirt: &kubeoneapi.KubevirtSpec{InfraNamespace: "infra-ns"}},
		},
		{
			name:          "other provider",
			cloudProvider: kubeoneapi.CloudProviderSpec{Hetzner: &kubeoneapi.HetznerSpec{}},
			wantSecret:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &state.State{
				Cluster: &kubeoneapi.KubeOneCluster{
					CloudProvider:          tc.cloudProvider,
					MachineController:      &kubeoneapi.MachineControllerConfig{Deploy: true},
					OperatingSystemManager: &kubeoneapi.OperatingSystemManagerConfig{Deploy: true},
				},
			}

			envVarsMC, hashMC, err := mcCredentials(s)
			if err != nil {
				t.Fatal(err)
			}

			var data templateData
			if err := osmCredentials(s, &data); err != nil {
				t.Fatal(err)
			}

			checks := []struct {
				what, envVars, secretName, hash string
			}{
				{"machine-controller", string(envVarsMC), credentials.SecretNameMC, hashMC},
				{"operating-system-manager", data.OperatingSystemManagerCredentialsEnvVars, credentials.SecretNameOSM, data.OperatingSystemManagerCredentialsHash},
			}

			for _, check := range checks {
				if got := strings.Contains(check.envVars, check.secretName); got != tc.wantSecret {
					t.Errorf("%s env vars reference %s = %v, want %v", check.what, check.secretName, got, tc.wantSecret)
				}

				if got := check.hash != ""; got != tc.wantSecret {
					t.Errorf("%s credentials hash set = %v, want %v", check.what, got, tc.wantSecret)
				}
			}
		})
	}
}
