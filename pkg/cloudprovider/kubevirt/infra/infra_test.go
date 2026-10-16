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

package infra

import (
	"context"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testCluster   = "c1"
	testNamespace = "infra-ns"
)

var testComponent = Component{
	Name: "machine-controller",
	Rules: []rbacv1.PolicyRule{
		{APIGroups: []string{"kubevirt.io"}, Resources: []string{"virtualmachines"}, Verbs: []string{"get"}},
	},
	NeedsKubeconfig: true,
}

func newFakeClient(t *testing.T, objs ...ctrlrt.Object) ctrlrt.Client {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func populatedToken(namespace, name, saName string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{corev1.ServiceAccountNameKey: saName},
		},
		Type: corev1.SecretTypeServiceAccountToken,
		Data: map[string][]byte{
			corev1.ServiceAccountTokenKey:  []byte("token-" + name),
			corev1.ServiceAccountRootCAKey: []byte("ca-" + name),
		},
	}
}

func TestComponentObjects(t *testing.T) {
	wantLabels := map[string]string{LabelClusterName: testCluster, LabelComponent: "machine-controller"}

	sa := testComponent.ServiceAccount(testCluster, testNamespace)
	role := testComponent.Role(testCluster, testNamespace)
	rb := testComponent.RoleBinding(testCluster, testNamespace)
	token := testComponent.TokenSecret(testCluster, testNamespace)

	for _, obj := range []ctrlrt.Object{sa, role, rb, token} {
		if obj.GetNamespace() != testNamespace {
			t.Errorf("%T: namespace = %q, want %q", obj, obj.GetNamespace(), testNamespace)
		}

		for k, v := range wantLabels {
			if obj.GetLabels()[k] != v {
				t.Errorf("%T: label %s = %q, want %q", obj, k, obj.GetLabels()[k], v)
			}
		}
	}

	if sa.Name != "c1-machine-controller" {
		t.Errorf("ServiceAccount name = %q", sa.Name)
	}

	if rb.RoleRef.Name != role.Name || rb.Subjects[0].Name != sa.Name || rb.Subjects[0].Namespace != testNamespace {
		t.Errorf("RoleBinding does not bind Role %q to ServiceAccount %q: %+v", role.Name, sa.Name, rb)
	}

	if token.Type != corev1.SecretTypeServiceAccountToken || token.Annotations[corev1.ServiceAccountNameKey] != sa.Name {
		t.Errorf("token Secret is not bound to ServiceAccount %q: %+v", sa.Name, token)
	}
}

func TestInClusterKubeconfig(t *testing.T) {
	token := populatedToken(testNamespace, "c1-machine-controller-token", "c1-machine-controller")

	raw, err := InClusterKubeconfig(testCluster, testNamespace, token)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := clientcmd.Load(raw)
	if err != nil {
		t.Fatal(err)
	}

	ctx := cfg.Contexts[cfg.CurrentContext]
	if ctx == nil {
		t.Fatalf("current context %q not found", cfg.CurrentContext)
	}

	if ctx.Namespace != testNamespace {
		t.Errorf("namespace = %q, want %q", ctx.Namespace, testNamespace)
	}

	cluster := cfg.Clusters[ctx.Cluster]
	if cluster.Server != "https://kubernetes.default.svc" {
		t.Errorf("server = %q", cluster.Server)
	}

	if string(cluster.CertificateAuthorityData) != "ca-c1-machine-controller-token" {
		t.Errorf("CA = %q", cluster.CertificateAuthorityData)
	}

	if cfg.AuthInfos[ctx.AuthInfo].Token != "token-c1-machine-controller-token" {
		t.Errorf("token = %q", cfg.AuthInfos[ctx.AuthInfo].Token)
	}
}

func TestTenantKubeconfig(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   kubeoneapi.APIEndpoint
		wantServer string
	}{
		{
			name:       "hostname",
			endpoint:   kubeoneapi.APIEndpoint{Host: "c1-kubeapi.infra-ns.svc.cluster.local", Port: 6443},
			wantServer: "https://c1-kubeapi.infra-ns.svc.cluster.local:6443",
		},
		{
			name:       "ipv6",
			endpoint:   kubeoneapi.APIEndpoint{Host: "fd00::1", Port: 6443},
			wantServer: "https://[fd00::1]:6443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := populatedToken("kube-system", "machine-controller-token", "machine-controller")

			raw, err := TenantKubeconfig(testCluster, TenantServer(tt.endpoint), token)
			if err != nil {
				t.Fatal(err)
			}

			cfg, err := clientcmd.Load(raw)
			if err != nil {
				t.Fatal(err)
			}

			ctx := cfg.Contexts[cfg.CurrentContext]
			if got := cfg.Clusters[ctx.Cluster].Server; got != tt.wantServer {
				t.Errorf("server = %q, want %q", got, tt.wantServer)
			}

			if ctx.Namespace != "" {
				t.Errorf("namespace = %q, want empty", ctx.Namespace)
			}
		})
	}
}

func TestKubeconfigRequiresPopulatedToken(t *testing.T) {
	token := testComponent.TokenSecret(testCluster, testNamespace)

	if _, err := InClusterKubeconfig(testCluster, testNamespace, token); err == nil {
		t.Error("expected an error for an unpopulated token Secret")
	}
}

func TestEnsureIdentity(t *testing.T) {
	ctx := context.Background()
	// The token controller does not run in the fake client, so the token
	// Secret is pre-populated.
	c := newFakeClient(t, populatedToken(testNamespace, testComponent.TokenSecretName(testCluster), testComponent.ServiceAccountName(testCluster)))

	if err := testComponent.EnsureIdentity(ctx, c, testCluster, testNamespace); err != nil {
		t.Fatal(err)
	}

	for _, obj := range []ctrlrt.Object{&corev1.ServiceAccount{}, &rbacv1.Role{}, &rbacv1.RoleBinding{}} {
		if err := c.Get(ctx, ctrlrt.ObjectKey{Namespace: testNamespace, Name: "c1-machine-controller"}, obj); err != nil {
			t.Errorf("getting %T: %v", obj, err)
		}
	}

	token := &corev1.Secret{}
	if err := c.Get(ctx, ctrlrt.ObjectKey{Namespace: testNamespace, Name: testComponent.TokenSecretName(testCluster)}, token); err != nil {
		t.Fatal(err)
	}

	if len(token.Data[corev1.ServiceAccountTokenKey]) == 0 {
		t.Error("reconciling the token Secret dropped the populated token")
	}

	kubeconfig := &corev1.Secret{}
	if err := c.Get(ctx, ctrlrt.ObjectKey{Namespace: testNamespace, Name: "c1-machine-controller-kubeconfig"}, kubeconfig); err != nil {
		t.Fatal(err)
	}

	if _, err := clientcmd.Load(kubeconfig.Data[KubeconfigSecretKey]); err != nil {
		t.Errorf("invalid kubeconfig: %v", err)
	}
}

func TestEnsureIdentityWithoutKubeconfig(t *testing.T) {
	ctx := context.Background()
	c := newFakeClient(t)
	comp := Component{Name: "kubevirt-csi-controller"}

	if err := comp.EnsureIdentity(ctx, c, testCluster, testNamespace); err != nil {
		t.Fatal(err)
	}

	secrets := &corev1.SecretList{}
	if err := c.List(ctx, secrets); err != nil {
		t.Fatal(err)
	}

	if len(secrets.Items) != 0 {
		t.Errorf("expected no Secrets, got %d", len(secrets.Items))
	}
}

func TestEnsureTenantKubeconfig(t *testing.T) {
	ctx := context.Background()
	tenantToken := populatedToken("kube-system", "machine-controller-token", "machine-controller")
	tenant := newFakeClient(t, tenantToken)
	infra := newFakeClient(t)

	err := testComponent.EnsureTenantKubeconfig(ctx, tenant, infra, testCluster, testNamespace, "https://10.0.0.1:6443", ctrlrt.ObjectKeyFromObject(tenantToken))
	if err != nil {
		t.Fatal(err)
	}

	secret := &corev1.Secret{}
	if err := infra.Get(ctx, ctrlrt.ObjectKey{Namespace: testNamespace, Name: "c1-machine-controller-tenant-kubeconfig"}, secret); err != nil {
		t.Fatal(err)
	}

	if secret.Labels[LabelClusterName] != testCluster {
		t.Errorf("missing cluster label: %v", secret.Labels)
	}
}

func TestCleanup(t *testing.T) {
	ctx := context.Background()
	meta := func(name, cluster string, withComponent bool) metav1.ObjectMeta {
		l := map[string]string{LabelClusterName: cluster}
		if withComponent {
			l[LabelComponent] = "x"
		}

		return metav1.ObjectMeta{Name: name, Namespace: testNamespace, Labels: l}
	}

	c := newFakeClient(t,
		&appsv1.Deployment{ObjectMeta: meta("own-deploy", testCluster, true)},
		&corev1.Secret{ObjectMeta: meta("own-secret", testCluster, true)},
		&corev1.ServiceAccount{ObjectMeta: meta("own-sa", testCluster, true)},
		&rbacv1.Role{ObjectMeta: meta("own-role", testCluster, true)},
		// control plane API server Service: same cluster, no component label
		&corev1.Service{ObjectMeta: meta("c1-kubeapi", testCluster, false)},
		// another cluster sharing the namespace
		&corev1.Secret{ObjectMeta: meta("other-secret", "c2", true)},
	)

	if err := Cleanup(ctx, c, testCluster, testNamespace); err != nil {
		t.Fatal(err)
	}

	gone := []ctrlrt.Object{
		&appsv1.Deployment{ObjectMeta: meta("own-deploy", testCluster, true)},
		&corev1.Secret{ObjectMeta: meta("own-secret", testCluster, true)},
		&corev1.ServiceAccount{ObjectMeta: meta("own-sa", testCluster, true)},
		&rbacv1.Role{ObjectMeta: meta("own-role", testCluster, true)},
	}
	for _, obj := range gone {
		if err := c.Get(ctx, ctrlrt.ObjectKeyFromObject(obj), obj); !apierrors.IsNotFound(err) {
			t.Errorf("%T %s: expected NotFound, got %v", obj, obj.GetName(), err)
		}
	}

	kept := []ctrlrt.Object{
		&corev1.Service{ObjectMeta: meta("c1-kubeapi", testCluster, false)},
		&corev1.Secret{ObjectMeta: meta("other-secret", "c2", true)},
	}
	for _, obj := range kept {
		if err := c.Get(ctx, ctrlrt.ObjectKeyFromObject(obj), obj); err != nil {
			t.Errorf("%T %s: expected to be kept, got %v", obj, obj.GetName(), err)
		}
	}
}
