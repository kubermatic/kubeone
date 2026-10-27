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

package gce

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

const (
	testProject = "proj"
	testRegion  = "europe-west3"
	testZone    = "europe-west3-a"
	testLBIP    = "203.0.113.10"
	testSubnet  = "projects/proj/regions/europe-west3/subnetworks/default"
)

// fakeCompute is a minimal in-memory Compute Engine REST API. Resources are
// stored as JSON objects by their path, e.g.
// "projects/proj/regions/europe-west3/addresses/demo-kubeapi".
type fakeCompute struct {
	mu      sync.Mutex
	objects map[string]map[string]any
	calls   []string

	// pendingOps makes mutating calls return RUNNING operations, which are
	// DONE on the first wait call.
	pendingOps bool
	// opError is set as error of every operation.
	opError string
	// deleted are the paths of the deleted resources, in order.
	deleted []string
}

func newFakeCompute(t *testing.T) (*fakeCompute, *compute.Service) {
	t.Helper()

	f := &fakeCompute{objects: map[string]map[string]any{}}
	srv := httptest.NewServer(http.HandlerFunc(f.serveHTTP))
	t.Cleanup(srv.Close)

	svc, err := compute.NewService(context.Background(), option.WithEndpoint(srv.URL+"/"), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("creating compute client: %v", err)
	}

	return f, svc
}

func (f *fakeCompute) put(p string, obj map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	obj["name"] = path.Base(p)
	obj["selfLink"] = "https://www.googleapis.com/compute/v1/" + p
	f.objects[p] = obj
}

func (f *fakeCompute) get(p string) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.objects[p]
}

// mutations returns the non-GET calls, except the wait calls.
func (f *fakeCompute) mutations() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []string
	for _, c := range f.calls {
		if !strings.HasPrefix(c, "GET ") && !strings.HasSuffix(c, "/wait") && !strings.HasSuffix(c, "/listNetworkEndpoints") {
			out = append(out, c)
		}
	}

	return out
}

func (f *fakeCompute) serveHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	p := strings.TrimPrefix(r.URL.Path, "/")
	f.calls = append(f.calls, r.Method+" "+p)

	var body map[string]any
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	switch {
	case r.Method == http.MethodPost && strings.HasSuffix(p, "/wait"):
		f.writeOperation(w, path.Base(path.Dir(p)), true)
	case r.Method == http.MethodPost && strings.HasSuffix(p, "/listNetworkEndpoints"):
		neg := f.objects[path.Dir(p)]
		items, _ := neg["_endpoints"].([]any)
		writeJSON(w, map[string]any{"items": items})
	case r.Method == http.MethodPost && strings.HasSuffix(p, "/attachNetworkEndpoints"):
		neg := f.objects[path.Dir(p)]
		items, _ := neg["_endpoints"].([]any)
		for _, ep := range field[[]any](body, "networkEndpoints") {
			items = append(items, map[string]any{"networkEndpoint": ep})
		}
		neg["_endpoints"] = items
		f.writeOperation(w, "attach", false)
	case r.Method == http.MethodGet && strings.HasSuffix(p, "/aggregated/networkEndpointGroups"):
		items := map[string]any{}
		for key, obj := range f.objects {
			if !strings.Contains(key, "/networkEndpointGroups/") {
				continue
			}
			zone := "zones/" + strings.Split(key, "/")[3]
			scoped, _ := items[zone].(map[string]any)
			if scoped == nil {
				scoped = map[string]any{}
				items[zone] = scoped
			}
			negs, _ := scoped["networkEndpointGroups"].([]any)
			withZone := map[string]any{"zone": "https://www.googleapis.com/compute/v1/projects/proj/" + zone}
			for k, v := range obj {
				if k != "_endpoints" {
					withZone[k] = v
				}
			}
			scoped["networkEndpointGroups"] = append(negs, withZone)
		}
		writeJSON(w, map[string]any{"items": items})
	case r.Method == http.MethodDelete:
		obj, ok := f.objects[p]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"error": map[string]any{"code": http.StatusNotFound, "message": "not found"}})

			return
		}
		if user := f.usedBy(p, obj); user != "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]any{"error": map[string]any{"code": http.StatusBadRequest, "message": "resourceInUseByAnotherResource: " + p + " is used by " + user}})

			return
		}
		delete(f.objects, p)
		f.deleted = append(f.deleted, p)
		f.writeOperation(w, "delete-"+path.Base(p), false)
	case r.Method == http.MethodGet:
		obj, ok := f.objects[p]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"error": map[string]any{"code": http.StatusNotFound, "message": "not found"}})

			return
		}
		writeJSON(w, obj)
	case r.Method == http.MethodPost:
		name, _ := body["name"].(string)
		resource := p + "/" + name
		body["selfLink"] = "https://www.googleapis.com/compute/v1/" + resource
		if strings.Contains(p, "/addresses") {
			body["address"] = testLBIP
		}
		f.objects[resource] = body
		f.writeOperation(w, "insert-"+name, false)
	case r.Method == http.MethodPatch:
		obj, ok := f.objects[p]
		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}
		for k, v := range body {
			obj[k] = v
		}
		f.writeOperation(w, "patch-"+path.Base(p), false)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// usedBy returns the path of a resource referencing the resource at p, like
// GCE refuses to delete a backend service used by a forwarding rule.
func (f *fakeCompute) usedBy(p string, obj map[string]any) string {
	ip, _ := obj["address"].(string)

	for key, other := range f.objects {
		if key == p {
			continue
		}
		raw, err := json.Marshal(other)
		if err != nil {
			continue
		}
		if strings.Contains(string(raw), `"`+p+`"`) || strings.Contains(string(raw), "/"+p+`"`) {
			return key
		}
		if ip != "" && other["IPAddress"] == ip {
			return key
		}
	}

	return ""
}

func (f *fakeCompute) writeOperation(w http.ResponseWriter, name string, waited bool) {
	op := map[string]any{
		"name":   name,
		"status": "DONE",
		"region": "https://www.googleapis.com/compute/v1/projects/proj/regions/" + testRegion,
	}
	if f.pendingOps && !waited {
		op["status"] = "RUNNING"
	}
	if f.opError != "" && op["status"] == "DONE" {
		op["error"] = map[string]any{"errors": []any{map[string]any{"code": "QUOTA_EXCEEDED", "message": f.opError}}}
	}
	writeJSON(w, op)
}

func writeJSON(w http.ResponseWriter, v map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// field returns obj[key] as T, or the zero value.
func field[T any](obj map[string]any, key string) T {
	v, _ := obj[key].(T)

	return v
}

func testLBConfig() *lbConfig {
	return &lbConfig{
		ClusterName:         "demo",
		ProjectID:           testProject,
		Region:              testRegion,
		Name:                "demo-kubeapi",
		Network:             "global/networks/default",
		AllowedSourceRanges: []string{"0.0.0.0/0"},
	}
}

// regional returns the path of the regional load balancer resource.
func regional(collection string) string {
	return fmt.Sprintf("projects/%s/regions/%s/%s/demo-kubeapi", testProject, testRegion, collection)
}

func ensureAll(t *testing.T, svc *compute.Service, cfg *lbConfig) string {
	t.Helper()

	ctx := context.Background()
	if err := ensureFirewalls(ctx, svc, cfg); err != nil {
		t.Fatalf("ensureFirewalls: %v", err)
	}

	ip, err := ensureLoadBalancer(ctx, svc, cfg)
	if err != nil {
		t.Fatalf("ensureLoadBalancer: %v", err)
	}

	return ip
}

func TestEnsureLoadBalancerCreatesResources(t *testing.T) {
	f, svc := newFakeCompute(t)
	cfg := testLBConfig()

	if ip := ensureAll(t, svc, cfg); ip != testLBIP {
		t.Errorf("got IP %q, want %q", ip, testLBIP)
	}

	apiFW := f.get("projects/proj/global/firewalls/demo-kubeapi-6443")
	if apiFW == nil || fmt.Sprint(apiFW["sourceRanges"]) != "[0.0.0.0/0]" || fmt.Sprint(apiFW["targetTags"]) != "[demo-control-plane]" {
		t.Errorf("unexpected api firewall rule: %v", apiFW)
	}

	hcFW := f.get("projects/proj/global/firewalls/demo-kubeapi-6443-hc")
	if hcFW == nil || fmt.Sprint(hcFW["sourceRanges"]) != "[35.191.0.0/16 209.85.204.0/22]" {
		t.Errorf("unexpected health check firewall rule: %v", hcFW)
	}

	hc := f.get(regional("healthChecks"))
	if hc == nil || hc["type"] != "TCP" || fmt.Sprint(hc["tcpHealthCheck"]) != "map[port:6443]" {
		t.Errorf("unexpected health check: %v", hc)
	}

	bs := f.get(regional("backendServices"))
	if bs == nil || bs["loadBalancingScheme"] != "EXTERNAL" || bs["protocol"] != "TCP" ||
		fmt.Sprint(bs["healthChecks"]) != "["+regional("healthChecks")+"]" {
		t.Errorf("unexpected backend service: %v", bs)
	}

	fr := f.get(regional("forwardingRules"))
	if fr == nil || fr["loadBalancingScheme"] != "EXTERNAL" || fr["IPProtocol"] != "TCP" ||
		fmt.Sprint(fr["ports"]) != "[6443]" || fr["IPAddress"] != testLBIP ||
		fr["backendService"] != regional("backendServices") {
		t.Errorf("unexpected forwarding rule: %v", fr)
	}

	for _, obj := range []map[string]any{apiFW, hcFW, hc, bs, fr, f.get(regional("addresses"))} {
		if obj["description"] != "kubeone:demo" {
			t.Errorf("resource %v has no ownership description", obj["name"])
		}
	}
}

func TestEnsureLoadBalancerIsIdempotent(t *testing.T) {
	f, svc := newFakeCompute(t)
	cfg := testLBConfig()

	ensureAll(t, svc, cfg)
	created := len(f.mutations())

	ensureAll(t, svc, cfg)
	if got := f.mutations(); len(got) != created {
		t.Errorf("second run made changes: %v", got[created:])
	}
}

func TestEnsureLoadBalancerRejectsForeignResources(t *testing.T) {
	f, svc := newFakeCompute(t)
	f.put(regional("addresses"), map[string]any{"description": "someone else"})

	_, err := ensureLoadBalancer(context.Background(), svc, testLBConfig())
	if err == nil || !strings.Contains(err.Error(), "not created by KubeOne") {
		t.Fatalf("expected ownership error, got %v", err)
	}

	if got := f.mutations(); len(got) != 0 {
		t.Errorf("expected no changes, got %v", got)
	}
}

func TestEnsureFirewallsUpdatesSourceRanges(t *testing.T) {
	f, svc := newFakeCompute(t)
	cfg := testLBConfig()
	ensureAll(t, svc, cfg)

	cfg.AllowedSourceRanges = []string{"10.0.0.0/8"}
	if err := ensureFirewalls(context.Background(), svc, cfg); err != nil {
		t.Fatalf("ensureFirewalls: %v", err)
	}

	if got := fmt.Sprint(f.get("projects/proj/global/firewalls/demo-kubeapi-6443")["sourceRanges"]); got != "[10.0.0.0/8]" {
		t.Errorf("source ranges not updated, got %s", got)
	}
	if got := fmt.Sprint(f.get("projects/proj/global/firewalls/demo-kubeapi-6443-hc")["sourceRanges"]); got != "[35.191.0.0/16 209.85.204.0/22]" {
		t.Errorf("health check source ranges changed, got %s", got)
	}
}

func TestWaitForOperation(t *testing.T) {
	t.Run("pending operations are waited for", func(t *testing.T) {
		f, svc := newFakeCompute(t)
		f.pendingOps = true

		ensureAll(t, svc, testLBConfig())

		waits := 0
		for _, c := range f.calls {
			if strings.HasSuffix(c, "/wait") {
				waits++
			}
		}
		if waits == 0 {
			t.Error("expected wait calls for pending operations")
		}
	})

	t.Run("operation errors are returned", func(t *testing.T) {
		f, svc := newFakeCompute(t)
		f.pendingOps = true
		f.opError = "quota exceeded"

		err := ensureFirewalls(context.Background(), svc, testLBConfig())
		if err == nil || !strings.Contains(err.Error(), "quota exceeded") {
			t.Fatalf("expected operation error, got %v", err)
		}
	})
}

func putInstance(f *fakeCompute, zone, name, ip, subnet string) {
	f.put(fmt.Sprintf("projects/%s/zones/%s/instances/%s", testProject, zone, name), map[string]any{
		"networkInterfaces": []any{map[string]any{
			"network":    "https://www.googleapis.com/compute/v1/projects/proj/global/networks/default",
			"subnetwork": "https://www.googleapis.com/compute/v1/" + subnet,
			"networkIP":  ip,
		}},
	})
}

func ensureVMEndpoint(t *testing.T, svc *compute.Service, cfg *lbConfig, zone, name string) error {
	t.Helper()

	bs, err := svc.RegionBackendServices.Get(cfg.ProjectID, cfg.Region, cfg.Name).Do()
	if err != nil {
		t.Fatalf("getting backend service: %v", err)
	}

	return ensureEndpoint(context.Background(), svc, cfg, bs, zone, name)
}

func negEndpoints(f *fakeCompute, zone string) []string {
	neg := f.get(fmt.Sprintf("projects/%s/zones/%s/networkEndpointGroups/demo-kubeapi-%s", testProject, zone, zone))
	items, _ := neg["_endpoints"].([]any)

	var out []string
	for _, item := range items {
		itemObj, _ := item.(map[string]any)
		ep := field[map[string]any](itemObj, "networkEndpoint")
		out = append(out, fmt.Sprintf("%s/%s", ep["instance"], ep["ipAddress"]))
	}

	return out
}

func TestEnsureEndpoint(t *testing.T) {
	f, svc := newFakeCompute(t)
	cfg := testLBConfig()
	ensureAll(t, svc, cfg)

	putInstance(f, testZone, "demo-cp-a-0", "10.156.0.2", testSubnet)
	putInstance(f, testZone, "demo-cp-a-1", "10.156.0.3", testSubnet)
	putInstance(f, "europe-west3-b", "demo-cp-b-0", "10.156.0.4", testSubnet)

	for _, vm := range []struct{ zone, name string }{
		{testZone, "demo-cp-a-0"},
		{testZone, "demo-cp-a-1"},
		{"europe-west3-b", "demo-cp-b-0"},
	} {
		if err := ensureVMEndpoint(t, svc, cfg, vm.zone, vm.name); err != nil {
			t.Fatalf("ensureEndpoint %s: %v", vm.name, err)
		}
	}

	neg := f.get("projects/proj/zones/europe-west3-a/networkEndpointGroups/demo-kubeapi-europe-west3-a")
	if neg == nil || neg["networkEndpointType"] != "GCE_VM_IP" || resourcePath(field[string](neg, "subnetwork")) != testSubnet || neg["description"] != "kubeone:demo" {
		t.Errorf("unexpected network endpoint group: %v", neg)
	}

	if got := negEndpoints(f, testZone); !slices.Equal(got, []string{"demo-cp-a-0/10.156.0.2", "demo-cp-a-1/10.156.0.3"}) {
		t.Errorf("unexpected endpoints in %s: %v", testZone, got)
	}
	if got := negEndpoints(f, "europe-west3-b"); !slices.Equal(got, []string{"demo-cp-b-0/10.156.0.4"}) {
		t.Errorf("unexpected endpoints in europe-west3-b: %v", got)
	}

	backends, _ := f.get(regional("backendServices"))["backends"].([]any)
	var groups []string
	for _, b := range backends {
		b, _ := b.(map[string]any)
		if b["balancingMode"] != "CONNECTION" {
			t.Errorf("backend %v has balancing mode %v, want CONNECTION", b["group"], b["balancingMode"])
		}
		groups = append(groups, field[string](b, "group"))
	}
	want := []string{
		"projects/proj/zones/europe-west3-a/networkEndpointGroups/demo-kubeapi-europe-west3-a",
		"projects/proj/zones/europe-west3-b/networkEndpointGroups/demo-kubeapi-europe-west3-b",
	}
	if !slices.Equal(groups, want) {
		t.Errorf("got backends %v, want %v", groups, want)
	}

	// Running again changes nothing
	before := len(f.mutations())
	if err := ensureVMEndpoint(t, svc, cfg, testZone, "demo-cp-a-0"); err != nil {
		t.Fatalf("ensureEndpoint again: %v", err)
	}
	if got := f.mutations(); len(got) != before {
		t.Errorf("second run made changes: %v", got[before:])
	}
}

func TestEnsureEndpointRejectsOtherSubnetwork(t *testing.T) {
	f, svc := newFakeCompute(t)
	cfg := testLBConfig()
	ensureAll(t, svc, cfg)

	putInstance(f, testZone, "demo-cp-0", "10.156.0.2", testSubnet)
	putInstance(f, testZone, "demo-cp-1", "10.200.0.2", "projects/proj/regions/europe-west3/subnetworks/other")

	if err := ensureVMEndpoint(t, svc, cfg, testZone, "demo-cp-0"); err != nil {
		t.Fatalf("ensureEndpoint: %v", err)
	}

	err := ensureVMEndpoint(t, svc, cfg, testZone, "demo-cp-1")
	if err == nil || !strings.Contains(err.Error(), "same subnetwork") {
		t.Fatalf("expected subnetwork error, got %v", err)
	}
}

// setUpLoadBalancer creates the load balancer with control plane VMs in two
// zones, and resources of others next to it.
func setUpLoadBalancer(t *testing.T) (*fakeCompute, *compute.Service, *lbConfig) {
	t.Helper()

	f, svc := newFakeCompute(t)
	cfg := testLBConfig()
	ensureAll(t, svc, cfg)

	putInstance(f, testZone, "demo-cp-a-0", "10.156.0.2", testSubnet)
	putInstance(f, "europe-west3-b", "demo-cp-b-0", "10.156.0.3", testSubnet)
	for _, vm := range []struct{ zone, name string }{{testZone, "demo-cp-a-0"}, {"europe-west3-b", "demo-cp-b-0"}} {
		if err := ensureVMEndpoint(t, svc, cfg, vm.zone, vm.name); err != nil {
			t.Fatalf("ensureEndpoint %s: %v", vm.name, err)
		}
	}

	f.put("projects/proj/global/firewalls/allow-ssh", map[string]any{"description": "someone else"})
	f.put(regional("addresses")+"-other", map[string]any{"description": "kubeone:other"})
	f.put("projects/proj/zones/europe-west3-c/networkEndpointGroups/other-neg", map[string]any{"description": "someone else"})

	return f, svc, cfg
}

func ownedResources(f *fakeCompute) []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []string
	for key, obj := range f.objects {
		if obj["description"] == "kubeone:demo" {
			out = append(out, key)
		}
	}
	slices.Sort(out)

	return out
}

func TestCleanupLoadBalancer(t *testing.T) {
	f, svc, cfg := setUpLoadBalancer(t)

	if err := cleanupLoadBalancer(context.Background(), svc, cfg, testLBIP, logrus.New()); err != nil {
		t.Fatalf("cleanupLoadBalancer: %v", err)
	}

	if left := ownedResources(f); len(left) != 0 {
		t.Errorf("owned resources left: %v", left)
	}

	want := []string{
		regional("forwardingRules"),
		regional("backendServices"),
		"projects/proj/zones/europe-west3-a/networkEndpointGroups/demo-kubeapi-europe-west3-a",
		"projects/proj/zones/europe-west3-b/networkEndpointGroups/demo-kubeapi-europe-west3-b",
		regional("healthChecks"),
		regional("addresses"),
		"projects/proj/global/firewalls/demo-kubeapi-6443",
		"projects/proj/global/firewalls/demo-kubeapi-6443-hc",
	}
	got := slices.Clone(f.deleted)
	// The network endpoint groups are deleted in listing order
	slices.Sort(got[2:4])
	if !slices.Equal(got, want) {
		t.Errorf("deleted\n%v\nwant\n%v", got, want)
	}

	for _, kept := range []string{
		"projects/proj/global/firewalls/allow-ssh",
		regional("addresses") + "-other",
		"projects/proj/zones/europe-west3-c/networkEndpointGroups/other-neg",
		"projects/proj/zones/europe-west3-a/instances/demo-cp-a-0",
	} {
		if f.get(kept) == nil {
			t.Errorf("%s was deleted", kept)
		}
	}

	// Running again finds nothing to delete
	before := len(f.mutations())
	if err := cleanupLoadBalancer(context.Background(), svc, cfg, "", logrus.New()); err != nil {
		t.Fatalf("cleanupLoadBalancer again: %v", err)
	}
	if got := f.mutations(); len(got) != before {
		t.Errorf("second run made changes: %v", got[before:])
	}
}

func TestCleanupLoadBalancerKeepsForeignResources(t *testing.T) {
	f, svc, cfg := setUpLoadBalancer(t)

	// A health check with the name of ours, but created by someone else
	hc := regional("healthChecks")
	f.get(hc)["description"] = "someone else"

	if err := cleanupLoadBalancer(context.Background(), svc, cfg, testLBIP, logrus.New()); err != nil {
		t.Fatalf("cleanupLoadBalancer: %v", err)
	}

	if f.get(hc) == nil {
		t.Error("foreign health check was deleted")
	}
	if left := ownedResources(f); len(left) != 0 {
		t.Errorf("owned resources left: %v", left)
	}
}

func TestCleanupLoadBalancerWithUserEndpoint(t *testing.T) {
	f, svc, cfg := setUpLoadBalancer(t)

	if err := cleanupLoadBalancer(context.Background(), svc, cfg, "198.51.100.1", logrus.New()); err != nil {
		t.Fatalf("cleanupLoadBalancer: %v", err)
	}

	want := []string{"projects/proj/global/firewalls/demo-kubeapi-6443", "projects/proj/global/firewalls/demo-kubeapi-6443-hc"}
	if !slices.Equal(f.deleted, want) {
		t.Errorf("deleted %v, want only the firewall rules %v", f.deleted, want)
	}
	if f.get(regional("forwardingRules")) == nil {
		t.Error("load balancer was deleted although the kube-apiserver endpoint is not its address")
	}
}

func TestCleanupLoadBalancerPartial(t *testing.T) {
	f, svc := newFakeCompute(t)
	cfg := testLBConfig()

	// Apply failed after the firewall rules and the address
	if err := ensureFirewalls(context.Background(), svc, cfg); err != nil {
		t.Fatalf("ensureFirewalls: %v", err)
	}
	f.put(regional("addresses"), map[string]any{"description": "kubeone:demo", "address": testLBIP})

	if err := cleanupLoadBalancer(context.Background(), svc, cfg, "", logrus.New()); err != nil {
		t.Fatalf("cleanupLoadBalancer: %v", err)
	}
	if left := ownedResources(f); len(left) != 0 {
		t.Errorf("owned resources left: %v", left)
	}
}

func TestFakeRefusesDeletingUsedResources(t *testing.T) {
	_, svc, cfg := setUpLoadBalancer(t)

	// Guards the fake itself: deleting out of order must fail like on GCE
	_, err := svc.RegionBackendServices.Delete(cfg.ProjectID, cfg.Region, cfg.Name).Do()
	if err == nil || !strings.Contains(err.Error(), "resourceInUseByAnotherResource") {
		t.Fatalf("expected in-use error, got %v", err)
	}
}
