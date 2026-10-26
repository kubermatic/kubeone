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

package machinecontroller

import (
	"context"
	"crypto"
	"crypto/x509"
	"net"
	"slices"
	"strings"
	"testing"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/certificate/cabundle"
	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	certutil "k8s.io/client-go/util/cert"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func allows(rules []rbacv1.PolicyRule, group, resource, verb string) bool {
	for _, r := range rules {
		if slices.Contains(r.APIGroups, group) && slices.Contains(r.Resources, resource) && slices.Contains(r.Verbs, verb) {
			return true
		}
	}

	return false
}

func TestComponents(t *testing.T) {
	for _, comp := range []kubevirtinfra.Component{ControllerComponent, WebhookComponent} {
		if !comp.NeedsKubeconfig {
			t.Errorf("%s: the KubeVirt provider has no in-cluster fallback, an infra kubeconfig is required", comp.Name)
		}

		// Read on every machine create and in the webhook defaulting
		if !allows(comp.ClusterRules, "storage.k8s.io", "storageclasses", "get") {
			t.Errorf("%s: ClusterRole must allow reading StorageClasses", comp.Name)
		}

		if !allows(comp.ClusterRules, "cdi.kubevirt.io", "storageprofiles", "get") {
			t.Errorf("%s: ClusterRole must allow reading StorageProfiles", comp.Name)
		}

		for _, r := range comp.ClusterRules {
			for _, v := range r.Verbs {
				if !slices.Contains([]string{"get", "list", "watch"}, v) {
					t.Errorf("%s: cluster-scoped rules must be read-only, got %q on %v", comp.Name, v, r.Resources)
				}
			}
		}

		if !allows(comp.Rules, "kubevirt.io", "virtualmachineinstances", "get") {
			t.Errorf("%s: Role must allow reading VMIs", comp.Name)
		}
	}

	for _, verb := range []string{"create", "delete"} {
		if !allows(ControllerComponent.Rules, "kubevirt.io", "virtualmachines", verb) {
			t.Errorf("machine-controller Role must allow %s on VirtualMachines", verb)
		}

		if !allows(ControllerComponent.Rules, "", "secrets", verb) {
			t.Errorf("machine-controller Role must allow %s on Secrets (cloud-init user data)", verb)
		}

		if allows(WebhookComponent.Rules, "kubevirt.io", "virtualmachines", verb) {
			t.Errorf("the webhook must not be allowed to %s VirtualMachines", verb)
		}
	}
}

func TestWebhookServiceType(t *testing.T) {
	tests := []struct {
		name string
		spec kubeoneapi.KubevirtSpec
		want corev1.ServiceType
	}{
		{
			name: "control plane on KubeVirt",
			spec: kubeoneapi.KubevirtSpec{ControlPlane: &kubeoneapi.KubevirtControlPlane{}},
			want: corev1.ServiceTypeClusterIP,
		},
		{
			name: "external control plane",
			want: corev1.ServiceTypeLoadBalancer,
		},
		{
			name: "explicit",
			spec: kubeoneapi.KubevirtSpec{
				ControlPlane: &kubeoneapi.KubevirtControlPlane{},
				MachineController: &kubeoneapi.KubevirtMachineController{
					Webhook: kubeoneapi.KubevirtMachineControllerWebhook{
						ServiceType: corev1.ServiceTypeLoadBalancer,
						Annotations: map[string]string{"a": "b"},
					},
				},
			},
			want: corev1.ServiceTypeLoadBalancer,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := WebhookService("c1", "infra-ns", &tc.spec)

			if svc.Spec.Type != tc.want {
				t.Errorf("type = %s, want %s", svc.Spec.Type, tc.want)
			}

			if svc.Spec.Ports[0].Port != 443 || svc.Spec.Ports[0].TargetPort.IntValue() != 9876 {
				t.Errorf("ports = %+v", svc.Spec.Ports)
			}

			if tc.spec.MachineController != nil && svc.Annotations["a"] != "b" {
				t.Errorf("annotations = %v", svc.Annotations)
			}

			// The Service must select the webhook pods
			d := WebhookDeployment(deploymentOptions{clusterName: "c1", namespace: "infra-ns"})
			for k, v := range svc.Spec.Selector {
				if d.Spec.Template.Labels[k] != v {
					t.Errorf("selector %s=%s does not match the webhook pods", k, v)
				}
			}
		})
	}
}

func TestServiceHost(t *testing.T) {
	tests := []struct {
		name string
		svc  corev1.Service
		want string
	}{
		{
			name: "ClusterIP",
			svc:  corev1.Service{Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.0.0.5"}},
			want: "10.0.0.5",
		},
		{
			name: "LoadBalancer without ingress",
			svc:  corev1.Service{Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer, ClusterIP: "10.0.0.5"}},
			want: "",
		},
		{
			name: "LoadBalancer hostname",
			svc: corev1.Service{
				Spec:   corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer},
				Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{Hostname: "lb.example.com"}}}},
			},
			want: "lb.example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := serviceHost(&tc.svc); got != tc.want {
				t.Errorf("serviceHost = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMutatingWebhookConfiguration(t *testing.T) {
	mwc := MutatingWebhookConfiguration("fd00::5", []byte("ca"))

	if mwc.Name != "machinedeployments.machine-controller.kubermatic.io" {
		t.Errorf("name = %q, must match the machinecontroller addon to replace it", mwc.Name)
	}

	want := map[string]string{
		"machinedeployments.machine-controller.kubermatic.io": "https://[fd00::5]:443/machinedeployments",
		"machines.machine-controller.kubermatic.io":           "https://[fd00::5]:443/machines",
	}

	for _, wh := range mwc.Webhooks {
		if wh.ClientConfig.Service != nil {
			t.Errorf("%s: must use a URL, the webhook Service is in the infra cluster", wh.Name)
		}

		if wh.ClientConfig.URL == nil || *wh.ClientConfig.URL != want[wh.Name] {
			t.Errorf("%s: URL = %v, want %s", wh.Name, wh.ClientConfig.URL, want[wh.Name])
		}

		if string(wh.ClientConfig.CABundle) != "ca" {
			t.Errorf("%s: caBundle not set", wh.Name)
		}
	}
}

func testCA(t *testing.T) (*x509.Certificate, crypto.Signer) {
	t.Helper()

	key, err := certificate.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	cert, err := certutil.NewSelfSignedCACert(certutil.Config{CommonName: "kubernetes"}, key)
	if err != nil {
		t.Fatal(err)
	}

	return cert, key
}

func TestEnsureServingCert(t *testing.T) {
	ctx := context.Background()
	caCert, caKey := testCA(t)

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	first, err := ensureServingCert(ctx, c, "c1", "infra-ns", "10.0.0.5", caKey, caCert)
	if err != nil {
		t.Fatal(err)
	}

	secret := &corev1.Secret{}
	if err = c.Get(ctx, ctrlrt.ObjectKey{Namespace: "infra-ns", Name: "c1-machine-controller-webhook-serving-cert"}, secret); err != nil {
		t.Fatal(err)
	}

	if !servingCertValid(secret.Data[resources.TLSCertName], "10.0.0.5", caCert, time.Now()) {
		t.Error("generated certificate is not valid for the Service IP")
	}

	if secret.Labels[kubevirtinfra.LabelClusterName] != "c1" {
		t.Errorf("missing cleanup labels: %v", secret.Labels)
	}

	again, err := ensureServingCert(ctx, c, "c1", "infra-ns", "10.0.0.5", caKey, caCert)
	if err != nil {
		t.Fatal(err)
	}

	if again.hash != first.hash {
		t.Error("a valid certificate must be reused, or the webhook restarts on every apply")
	}

	moved, err := ensureServingCert(ctx, c, "c1", "infra-ns", "lb.example.com", caKey, caCert)
	if err != nil {
		t.Fatal(err)
	}

	if moved.hash == first.hash {
		t.Error("the certificate must be regenerated when the webhook address changes")
	}

	otherCA, otherKey := testCA(t)
	rotated, err := ensureServingCert(ctx, c, "c1", "infra-ns", "lb.example.com", otherKey, otherCA)
	if err != nil {
		t.Fatal(err)
	}

	if rotated.hash == moved.hash {
		t.Error("the certificate must be regenerated when the CA changes")
	}
}

func TestServingCertValid(t *testing.T) {
	caCert, caKey := testCA(t)

	certs, err := certificate.NewSignedTLSCert("10.0.0.5", nil, []net.IP{net.ParseIP("10.0.0.5")}, caKey, caCert)
	if err != nil {
		t.Fatal(err)
	}

	pemCert := []byte(certs[resources.TLSCertName])
	now := time.Now()

	if !servingCertValid(pemCert, "10.0.0.5", caCert, now) {
		t.Error("expected valid")
	}

	if servingCertValid(pemCert, "10.0.0.6", caCert, now) {
		t.Error("expected invalid for another IP")
	}

	if servingCertValid(pemCert, "10.0.0.5", caCert, now.Add(360*24*time.Hour)) {
		t.Error("expected invalid close to expiry")
	}

	if servingCertValid([]byte("garbage"), "10.0.0.5", caCert, now) {
		t.Error("expected invalid for garbage")
	}
}

func TestDeployments(t *testing.T) {
	o := deploymentOptions{
		clusterName:           "c1",
		namespace:             "infra-ns",
		image:                 "quay.io/kubermatic/machine-controller:test",
		externalCloudProvider: true,
		caBundle:              "bundle",
		joinClusterTimeout:    "20m",
		servingCertHash:       "abc",
	}

	tests := []struct {
		d        *appsv1.Deployment
		comp     kubevirtinfra.Component
		wantArgs []string
	}{
		{
			d:    ControllerDeployment(o),
			comp: ControllerComponent,
			wantArgs: []string{
				"-kubeconfig=/var/run/secrets/tenantcluster/kubeconfig",
				"-node-external-cloud-provider",
				"-ca-bundle=" + cabundle.SSLCertFilePath,
				"-node-csr-approver",
				"-join-cluster-timeout=20m",
			},
		},
		{
			d:    WebhookDeployment(o),
			comp: WebhookComponent,
			wantArgs: []string{
				"-kubeconfig=/var/run/secrets/tenantcluster/kubeconfig",
				"-node-external-cloud-provider",
				"-namespace=kube-system",
				"-listen-address=0.0.0.0:9876",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.comp.Name, func(t *testing.T) {
			if tc.d.Name != "c1-"+tc.comp.Name || tc.d.Namespace != "infra-ns" {
				t.Errorf("unexpected name %s/%s", tc.d.Namespace, tc.d.Name)
			}

			if tc.d.Labels[kubevirtinfra.LabelComponent] != tc.comp.Name {
				t.Errorf("missing cleanup labels: %v", tc.d.Labels)
			}

			pod := tc.d.Spec.Template.Spec
			if pod.ServiceAccountName != "c1-"+tc.comp.Name {
				t.Errorf("ServiceAccountName = %q", pod.ServiceAccountName)
			}

			c := pod.Containers[0]
			for _, want := range tc.wantArgs {
				if !slices.Contains(c.Args, want) {
					t.Errorf("args %v missing %q", c.Args, want)
				}
			}

			env := map[string]corev1.EnvVar{}
			for _, e := range c.Env {
				env[e.Name] = e
			}

			if env["POD_NAMESPACE"].Value != "infra-ns" {
				t.Errorf("POD_NAMESPACE = %q", env["POD_NAMESPACE"].Value)
			}

			kv := env["KUBEVIRT_KUBECONFIG"].ValueFrom
			if kv == nil || kv.SecretKeyRef == nil || kv.SecretKeyRef.Name != "c1-"+tc.comp.Name+"-kubeconfig" {
				t.Errorf("KUBEVIRT_KUBECONFIG must come from the infra kubeconfig Secret, got %+v", kv)
			}

			if _, ok := env["HTTPS_PROXY"]; ok {
				t.Error("tenant proxy settings must not be passed to pods in the infra cluster")
			}

			if _, ok := env[cabundle.SSLCertFileENV]; !ok {
				t.Error("CA bundle not injected")
			}

			volumes := map[string]string{}
			for _, v := range pod.Volumes {
				switch {
				case v.Secret != nil:
					volumes[v.Secret.SecretName] = v.Name
				case v.ConfigMap != nil:
					volumes[v.ConfigMap.Name] = v.Name
				}
			}

			if _, ok := volumes["c1-"+tc.comp.Name+"-tenant-kubeconfig"]; !ok {
				t.Errorf("tenant kubeconfig not mounted: %v", volumes)
			}

			if _, ok := volumes["c1-machine-controller-ca-bundle"]; !ok {
				t.Errorf("CA bundle must come from the per-cluster infra ConfigMap: %v", volumes)
			}
		})
	}

	webhook := tests[1].d.Spec.Template
	if webhook.Annotations[servingCertHashAnnotation] != "abc" {
		t.Error("the webhook must roll when the serving certificate changes")
	}

	var certMounted bool
	for _, v := range webhook.Spec.Volumes {
		if v.Secret != nil && v.Secret.SecretName == servingCertSecretName("c1") {
			certMounted = true
		}
	}

	if !certMounted {
		t.Error("serving certificate not mounted")
	}

	for _, arg := range ControllerDeployment(deploymentOptions{clusterName: "c1", namespace: "infra-ns"}).Spec.Template.Spec.Containers[0].Args {
		if strings.HasPrefix(arg, "-ca-bundle") || arg == "-node-external-cloud-provider" {
			t.Errorf("unexpected arg %q without CA bundle and external cloud provider", arg)
		}
	}
}

func TestJoinClusterTimeout(t *testing.T) {
	cluster := &kubeoneapi.KubeOneCluster{}
	if got := joinClusterTimeout(cluster); got != "15m" {
		t.Errorf("default = %q, want 15m", got)
	}

	cluster.Addons = &kubeoneapi.Addons{
		Addons: []kubeoneapi.AddonRef{{Addon: &kubeoneapi.Addon{
			Name:   resources.AddonMachineController,
			Params: map[string]string{"JOIN_TIMEOUT": "30m"},
		}}},
	}
	if got := joinClusterTimeout(cluster); got != "30m" {
		t.Errorf("from addon params = %q, want 30m", got)
	}
}

func TestDeploymentReady(t *testing.T) {
	d := &appsv1.Deployment{}
	d.Generation = 2
	d.Status = appsv1.DeploymentStatus{ObservedGeneration: 2, Replicas: 1, UpdatedReplicas: 1, ReadyReplicas: 1}

	if !deploymentReady(d) {
		t.Error("expected ready")
	}

	d.Status.ObservedGeneration = 1
	if deploymentReady(d) {
		t.Error("expected not ready while the new generation is not observed")
	}

	d.Status.ObservedGeneration = 2
	d.Status.Replicas = 2
	if deploymentReady(d) {
		t.Error("expected not ready while an old pod is still around")
	}
}

func TestWaitDeploymentReady(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	newDeployment := func(name string, ready int32) *appsv1.Deployment {
		d := &appsv1.Deployment{}
		d.Name = name
		d.Namespace = "infra-ns"
		d.Status = appsv1.DeploymentStatus{Replicas: 1, UpdatedReplicas: 1, ReadyReplicas: ready}

		return d
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(newDeployment("ready", 1), newDeployment("pending", 0)).
		WithStatusSubresource(&appsv1.Deployment{}).
		Build()

	if err := waitDeploymentReady(context.Background(), c, ctrlrt.ObjectKey{Namespace: "infra-ns", Name: "ready"}); err != nil {
		t.Errorf("ready Deployment: %v", err)
	}

	for _, name := range []string{"pending", "missing"} {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		err := waitDeploymentReady(ctx, c, ctrlrt.ObjectKey{Namespace: "infra-ns", Name: name})
		cancel()

		if err == nil {
			t.Errorf("%s Deployment: expected an error", name)
		}
	}
}
