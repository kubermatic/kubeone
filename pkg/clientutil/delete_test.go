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

package clientutil

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestDeleteAndWait(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	daemonSet := func(name string, finalizers ...string) *appsv1.DaemonSet {
		return &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "kube-system", Finalizers: finalizers}}
	}

	var propagation *metav1.DeletionPropagation
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		// the finalizer stands in for pods that are still terminating
		WithObjects(daemonSet("gone"), daemonSet("stuck", "example.com/pods")).
		WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				do := &client.DeleteOptions{}
				do.ApplyOptions(opts)
				propagation = do.PropagationPolicy

				return c.Delete(ctx, obj, opts...)
			},
		}).
		Build()

	ctx := context.Background()

	if err := DeleteAndWait(ctx, c, daemonSet("gone"), time.Second); err != nil {
		t.Fatalf("existing object: %v", err)
	}
	if propagation == nil || *propagation != metav1.DeletePropagationForeground {
		t.Errorf("propagation = %v, want Foreground", propagation)
	}
	if err := c.Get(ctx, client.ObjectKey{Namespace: "kube-system", Name: "gone"}, &appsv1.DaemonSet{}); !k8serrors.IsNotFound(err) {
		t.Errorf("expected the object to be deleted, got %v", err)
	}

	if err := DeleteAndWait(ctx, c, daemonSet("missing"), time.Second); err != nil {
		t.Errorf("missing object: %v", err)
	}

	if err := DeleteAndWait(ctx, c, daemonSet("stuck"), 100*time.Millisecond); err == nil {
		t.Error("object that is not removed in time: expected an error")
	}
}
