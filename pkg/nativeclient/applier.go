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

// Package nativeclient applies and deletes manifests through the Kubernetes API using server-side apply, as a
// native replacement for running "kubectl apply --prune" and "kubectl delete" on a control plane node.
package nativeclient

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/fail"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// DefaultFieldManager is the server-side apply field manager used unless overridden with WithFieldManager.
	DefaultFieldManager = "kubeone"

	// DefaultConcurrency is the default number of objects applied in parallel.
	DefaultConcurrency = 8

	crdEstablishTimeout = 1 * time.Minute
)

// DefaultPruneTypes mirrors the default prune allowlist of "kubectl apply --prune", minus the deprecated
// v1 Endpoints. Types of the applied objects are always pruned in addition to these.
var DefaultPruneTypes = []schema.GroupVersionKind{
	{Version: "v1", Kind: "ConfigMap"},
	{Version: "v1", Kind: "Namespace"},
	{Version: "v1", Kind: "PersistentVolumeClaim"},
	{Version: "v1", Kind: "PersistentVolume"},
	{Version: "v1", Kind: "Pod"},
	{Version: "v1", Kind: "ReplicationController"},
	{Version: "v1", Kind: "Secret"},
	{Version: "v1", Kind: "Service"},
	{Group: "batch", Version: "v1", Kind: "Job"},
	{Group: "batch", Version: "v1", Kind: "CronJob"},
	{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"},
	{Group: "apps", Version: "v1", Kind: "DaemonSet"},
	{Group: "apps", Version: "v1", Kind: "Deployment"},
	{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
	{Group: "apps", Version: "v1", Kind: "StatefulSet"},
}

// protectedNamespaces are never pruned, even if they carry the prune label.
var protectedNamespaces = sets.New(
	metav1.NamespaceDefault,
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
	"kube-node-lease",
)

// Applier applies and deletes sets of objects using server-side apply.
type Applier struct {
	client       client.Client
	fieldManager string
	concurrency  int
}

type Option func(*Applier)

// WithFieldManager overrides the server-side apply field manager.
func WithFieldManager(name string) Option {
	return func(a *Applier) {
		a.fieldManager = name
	}
}

// WithConcurrency sets how many objects are applied in parallel once CRDs and Namespaces are in place.
func WithConcurrency(n int) Option {
	return func(a *Applier) {
		a.concurrency = max(n, 1)
	}
}

func New(c client.Client, opts ...Option) *Applier {
	a := &Applier{
		client:       c,
		fieldManager: DefaultFieldManager,
		concurrency:  DefaultConcurrency,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

type ApplyOptions struct {
	// PruneSelector enables pruning when set: live objects matching the selector that are not part of the applied
	// set are deleted, the equivalent of "kubectl apply --prune -l <selector>".
	PruneSelector labels.Selector

	// PruneTypes are pruned in addition to DefaultPruneTypes and the types of the applied objects.
	// CustomResourceDefinitions are pruned only when listed here.
	PruneTypes []schema.GroupVersionKind

	// Inventory, when set, is a ConfigMap recording the types of the applied objects. Types recorded by a previous
	// Apply are pruned too, so objects are cleaned up even when their whole type disappears from the manifest.
	Inventory *client.ObjectKey

	// DryRun sends all requests with dryRun=All and skips managed fields migration and inventory updates.
	DryRun bool
}

type Result struct {
	Applied []ObjectRef

	// Migrated objects were previously managed by "kubectl apply" (client-side apply) and had their managed
	// fields handed over to the server-side apply field manager.
	Migrated []ObjectRef
	Pruned   []ObjectRef
}

type applyOutcome struct {
	ref      ObjectRef
	migrated bool
	ok       bool
}

// Apply server-side applies objs, migrating objects previously managed by client-side apply, and prunes stale
// objects if opts.PruneSelector is set. objs are not modified.
//
// CRDs are applied first and waited for, then Namespaces, then all remaining objects in parallel. Like kubectl,
// a failing object doesn't stop the others in its phase; all errors are returned joined and pruning is skipped.
func (a *Applier) Apply(ctx context.Context, objs []*unstructured.Unstructured, opts ApplyOptions) (*Result, error) {
	result := &Result{}

	previousKinds, err := a.prepareInventory(ctx, objs, opts)
	if err != nil {
		return result, err
	}

	crds, namespaces, rest := splitPhases(objs)

	if err = a.applyPhase(ctx, crds, 1, opts.DryRun, result); err != nil {
		return result, err
	}

	crdNames := make([]string, 0, len(crds))
	for _, crd := range crds {
		crdNames = append(crdNames, crd.GetName())
	}

	if err = a.waitForCRDs(ctx, crdNames, opts.DryRun); err != nil {
		return result, err
	}

	if err = a.applyPhase(ctx, namespaces, 1, opts.DryRun, result); err != nil {
		return result, err
	}

	if err = a.applyPhase(ctx, rest, a.concurrency, opts.DryRun, result); err != nil {
		return result, err
	}

	if opts.PruneSelector != nil {
		applied := sets.New(result.Applied...)
		if opts.Inventory != nil {
			applied.Insert(inventoryRef(*opts.Inventory))
		}

		result.Pruned, err = a.prune(ctx, a.pruneGVKs(objs, previousKinds, opts), applied, opts)
		if err != nil {
			return result, err
		}
	}

	return result, a.finalizeInventory(ctx, objs, opts)
}

// applyPhase applies objs with up to concurrency requests in flight, recording successes in result in the order
// of objs.
func (a *Applier) applyPhase(ctx context.Context, objs []*unstructured.Unstructured, concurrency int, dryRun bool, result *Result) error {
	var (
		outcomes = make([]applyOutcome, len(objs))
		errs     = make([]error, len(objs))
		group    errgroup.Group
	)

	group.SetLimit(concurrency)

	for i, obj := range objs {
		group.Go(func() error {
			desired := obj.DeepCopy()

			migrated, err := a.applyOne(ctx, desired, dryRun)
			if err != nil {
				errs[i] = err

				return nil
			}

			outcomes[i] = applyOutcome{ref: RefOf(desired), migrated: migrated, ok: true}

			return nil
		})
	}

	_ = group.Wait()

	for _, outcome := range outcomes {
		if !outcome.ok {
			continue
		}

		result.Applied = append(result.Applied, outcome.ref)
		if outcome.migrated {
			result.Migrated = append(result.Migrated, outcome.ref)
		}
	}

	return errors.Join(errs...)
}

// applyOne applies obj, defaulting its namespace, and reports whether its managed fields had to be migrated from
// client-side apply.
func (a *Applier) applyOne(ctx context.Context, obj *unstructured.Unstructured, dryRun bool) (bool, error) {
	if err := a.defaultNamespace(ctx, obj, true); err != nil {
		return false, err
	}

	applyOpts := []client.ApplyOption{client.FieldOwner(a.fieldManager), client.ForceOwnership}
	if dryRun {
		applyOpts = append(applyOpts, client.DryRunAll)
	}

	ref := RefOf(obj)

	// client.Apply decodes the response into the object it's given, so always apply a fresh copy of obj.
	live := obj.DeepCopy()
	if err := retryOnNoMatch(ctx, func() error {
		return a.client.Apply(ctx, client.ApplyConfigurationFromUnstructured(live), applyOpts...)
	}); err != nil {
		return false, fail.KubeClient(err, "applying %s", ref)
	}

	if dryRun {
		return false, nil
	}

	migrated, err := a.migrateClientSideApply(ctx, live)
	if err != nil || !migrated {
		return false, err
	}

	// Fields previously owned by client-side apply are now owned by our field manager. Apply again so fields
	// removed from the manifest are dropped (including the kubectl last-applied-configuration annotation).
	live = obj.DeepCopy()
	if err = a.client.Apply(ctx, client.ApplyConfigurationFromUnstructured(live), applyOpts...); err != nil {
		return true, fail.KubeClient(err, "re-applying %s after managed fields migration", ref)
	}

	return true, nil
}

// defaultNamespace sets the "default" namespace on namespaced objects without one, like kubectl does.
// With waitForType, it retries while the object's type is not yet known to the RESTMapper.
func (a *Applier) defaultNamespace(ctx context.Context, obj *unstructured.Unstructured, waitForType bool) error {
	if obj.GetNamespace() != "" {
		return nil
	}

	var namespaced bool

	isNamespaced := func() error {
		var err error
		namespaced, err = a.client.IsObjectNamespaced(obj)

		return err
	}

	var err error
	if waitForType {
		err = retryOnNoMatch(ctx, isNamespaced)
	} else {
		err = isNamespaced()
	}

	if meta.IsNoMatchError(err) {
		return err
	}

	if err != nil {
		return fail.KubeClient(err, "determining scope of %s", RefOf(obj))
	}

	if namespaced {
		obj.SetNamespace(metav1.NamespaceDefault)
	}

	return nil
}

func (a *Applier) waitForCRDs(ctx context.Context, names []string, dryRun bool) error {
	if len(names) == 0 || dryRun {
		return nil
	}

	condFn := clientutil.CRDsReadyCondition(ctx, a.client, names)
	err := wait.PollUntilContextTimeout(ctx, time.Second, crdEstablishTimeout, true, func(context.Context) (bool, error) {
		return condFn()
	})

	return fail.KubeClient(err, "waiting for CRDs %v to become established", names)
}

// pruneGVKs returns the types to look for stale objects in: explicitly requested types, the kubectl defaults, the
// types of objs, and the types recorded in the inventory, deduplicated by GroupKind.
func (a *Applier) pruneGVKs(objs []*unstructured.Unstructured, previousKinds sets.Set[schema.GroupKind], opts ApplyOptions) []schema.GroupVersionKind {
	var (
		seen      = sets.New[schema.GroupKind]()
		pruneGVKs []schema.GroupVersionKind
	)

	add := func(gvk schema.GroupVersionKind) {
		if !seen.Has(gvk.GroupKind()) {
			seen.Insert(gvk.GroupKind())
			pruneGVKs = append(pruneGVKs, gvk)
		}
	}

	// Explicitly requested types first, so they win over defaults sharing a GroupKind.
	for _, gvk := range opts.PruneTypes {
		add(gvk)
	}

	// Never prune CRDs implicitly; that would cascade to all their custom resources.
	seen.Insert(crdGroupKind)

	for _, gvk := range slices.Concat(DefaultPruneTypes, objectGVKs(objs)) {
		add(gvk)
	}

	// Recorded kinds have no version; use whatever version the API server prefers now. Kinds the server no longer
	// serves can't have objects left to prune.
	previous := previousKinds.Difference(seen).UnsortedList()
	slices.SortFunc(previous, func(a, b schema.GroupKind) int { return strings.Compare(a.String(), b.String()) })

	for _, gk := range previous {
		mapping, err := a.client.RESTMapper().RESTMapping(gk)
		if err != nil {
			continue
		}

		add(mapping.GroupVersionKind)
	}

	return pruneGVKs
}

func (a *Applier) prune(ctx context.Context, pruneGVKs []schema.GroupVersionKind, applied sets.Set[ObjectRef], opts ApplyOptions) ([]ObjectRef, error) {
	var pruned []ObjectRef

	for _, gvk := range pruneGVKs {
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(listGVK(gvk))

		err := a.client.List(ctx, list, client.MatchingLabelsSelector{Selector: opts.PruneSelector})
		if meta.IsNoMatchError(err) {
			continue
		}
		if err != nil {
			return pruned, fail.KubeClient(err, "listing %s for pruning", gvk.GroupKind())
		}

		for i := range list.Items {
			item := &list.Items[i]
			ref := RefOf(item)

			if applied.Has(ref) || item.GetDeletionTimestamp() != nil {
				continue
			}

			if ref.GroupKind == namespaceGroupKind && protectedNamespaces.Has(ref.Name) {
				continue
			}

			if err := a.deleteObject(ctx, item, opts.DryRun); err != nil {
				return pruned, fail.KubeClient(err, "pruning %s", ref)
			}

			pruned = append(pruned, ref)
		}
	}

	return pruned, nil
}

type DeleteOptions struct {
	// Selector restricts deletion to live objects matching it; nil matches everything.
	Selector labels.Selector

	// Inventory is the ConfigMap written by Apply, deleted once all objects are gone.
	Inventory *client.ObjectKey
}

// Delete deletes objs that exist and match opts.Selector, the equivalent of
// "kubectl delete -f - -l <selector> --ignore-not-found". Objects whose types aren't served are skipped.
func (a *Applier) Delete(ctx context.Context, objs []*unstructured.Unstructured, opts DeleteOptions) ([]ObjectRef, error) {
	var deleted []ObjectRef

	for _, obj := range sortForDelete(objs) {
		live := obj.DeepCopy()

		if err := a.defaultNamespace(ctx, live, false); err != nil {
			if meta.IsNoMatchError(err) {
				continue
			}

			return deleted, err
		}

		ref := RefOf(live)

		err := a.client.Get(ctx, client.ObjectKeyFromObject(live), live)
		switch {
		case k8serrors.IsNotFound(err), meta.IsNoMatchError(err):
			continue
		case err != nil:
			return deleted, fail.KubeClient(err, "getting %s", ref)
		}

		if opts.Selector != nil && !opts.Selector.Matches(labels.Set(live.GetLabels())) {
			continue
		}

		if err := a.deleteObject(ctx, live, false); err != nil {
			return deleted, fail.KubeClient(err, "deleting %s", ref)
		}

		deleted = append(deleted, ref)
	}

	if opts.Inventory != nil {
		return deleted, a.deleteInventory(ctx, *opts.Inventory)
	}

	return deleted, nil
}

func (a *Applier) deleteObject(ctx context.Context, obj *unstructured.Unstructured, dryRun bool) error {
	opts := []client.DeleteOption{client.PropagationPolicy(metav1.DeletePropagationBackground)}
	if dryRun {
		opts = append(opts, client.DryRunAll)
	}

	return client.IgnoreNotFound(a.client.Delete(ctx, obj, opts...))
}

func objectGVKs(objs []*unstructured.Unstructured) []schema.GroupVersionKind {
	gvks := make([]schema.GroupVersionKind, 0, len(objs))
	for _, obj := range objs {
		gvks = append(gvks, obj.GroupVersionKind())
	}

	return gvks
}

// retryOnNoMatch retries fn while the RESTMapper doesn't know the object's type yet, which happens right after
// its CRD has been established.
func retryOnNoMatch(ctx context.Context, fn func() error) error {
	backoff := wait.Backoff{Duration: 500 * time.Millisecond, Factor: 2, Steps: 6}

	var lastErr error

	err := wait.ExponentialBackoffWithContext(ctx, backoff, func(context.Context) (bool, error) {
		lastErr = fn()
		if meta.IsNoMatchError(lastErr) {
			return false, nil
		}

		return true, lastErr
	})
	if wait.Interrupted(err) {
		return lastErr
	}

	return err
}
