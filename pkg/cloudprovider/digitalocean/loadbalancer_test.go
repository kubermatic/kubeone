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

package digitalocean

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

// fakeLBService is a scripted, in-memory godo.LoadBalancersService. The
// embedded interface is nil, calling an unimplemented method panics.
type fakeLBService struct {
	godo.LoadBalancersService

	// pages returned by List, one slice per page
	pages [][]godo.LoadBalancer
	// getResponses are returned by Get in order, the last one is repeated
	getResponses []*godo.LoadBalancer
	getErr       error
	getStatus    int
	// notFoundGets is the number of Get calls that return 404 first, like a
	// load balancer that isn't visible yet right after it's created
	notFoundGets int

	createResult *godo.LoadBalancer
	updateErr    error

	created      []*godo.LoadBalancerRequest
	updated      []*godo.LoadBalancerRequest
	deleted      []string
	getCalls     int
	listRequests []int
}

func (f *fakeLBService) List(_ context.Context, opts *godo.ListOptions) ([]godo.LoadBalancer, *godo.Response, error) {
	page := 1
	if opts != nil && opts.Page > 0 {
		page = opts.Page
	}
	f.listRequests = append(f.listRequests, page)

	if len(f.pages) == 0 {
		return nil, &godo.Response{Links: &godo.Links{}}, nil
	}

	pages := &godo.Pages{}
	if page > 1 {
		pages.Prev = "https://api.digitalocean.com/v2/load_balancers?page=" + strconv.Itoa(page-1)
	}
	if page < len(f.pages) {
		pages.Next = "https://api.digitalocean.com/v2/load_balancers?page=" + strconv.Itoa(page+1)
	}

	return f.pages[page-1], &godo.Response{Links: &godo.Links{Pages: pages}}, nil
}

func (f *fakeLBService) Get(_ context.Context, _ string) (*godo.LoadBalancer, *godo.Response, error) {
	f.getCalls++
	if f.getCalls <= f.notFoundGets {
		return nil, &godo.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
	}
	if f.getErr != nil {
		return nil, &godo.Response{Response: &http.Response{StatusCode: f.getStatus}}, f.getErr
	}

	idx := min(f.getCalls-1, len(f.getResponses)-1)
	lb := *f.getResponses[idx]

	return &lb, &godo.Response{}, nil
}

func (f *fakeLBService) Create(_ context.Context, req *godo.LoadBalancerRequest) (*godo.LoadBalancer, *godo.Response, error) {
	f.created = append(f.created, req)

	return f.createResult, &godo.Response{}, nil
}

func (f *fakeLBService) Update(_ context.Context, id string, req *godo.LoadBalancerRequest) (*godo.LoadBalancer, *godo.Response, error) {
	f.updated = append(f.updated, req)
	if f.updateErr != nil {
		// Like godo, return no load balancer on errors
		return nil, &godo.Response{Response: &http.Response{StatusCode: http.StatusUnprocessableEntity}}, f.updateErr
	}

	return &godo.LoadBalancer{ID: id, Name: req.Name, Tag: req.Tag, ForwardingRules: req.ForwardingRules, HealthCheck: req.HealthCheck, Status: lbStatusUp, IP: "203.0.113.20"}, &godo.Response{}, nil
}

func (f *fakeLBService) Delete(_ context.Context, id string) (*godo.Response, error) {
	f.deleted = append(f.deleted, id)

	return &godo.Response{}, nil
}

// fakeVPCService returns the scripted VPCs by ID.
type fakeVPCService struct {
	vpcs     map[string]*godo.VPC
	getCalls int
}

func (f *fakeVPCService) Get(_ context.Context, id string) (*godo.VPC, *godo.Response, error) {
	f.getCalls++

	vpc, ok := f.vpcs[id]
	if !ok {
		return nil, &godo.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
	}

	return vpc, &godo.Response{}, nil
}

var fra1Region = &godo.Region{Slug: "fra1"}

func testVPCs() *fakeVPCService {
	return &fakeVPCService{vpcs: map[string]*godo.VPC{
		"vpc-default": {ID: "vpc-default", Name: "default-fra1", RegionSlug: "fra1", Default: true},
		"vpc-custom":  {ID: "vpc-custom", Name: "custom", RegionSlug: "fra1"},
	}}
}

func fastPolling(t *testing.T, timeout time.Duration) {
	t.Helper()

	interval, oldTimeout := lbPollInterval, lbPollTimeout
	lbPollInterval, lbPollTimeout = time.Millisecond, timeout
	t.Cleanup(func() { lbPollInterval, lbPollTimeout = interval, oldTimeout })
}

func testCluster(lb kubeoneapi.DigitalOceanLoadBalancer) *kubeoneapi.KubeOneCluster {
	return &kubeoneapi.KubeOneCluster{
		Name: "demo",
		CloudProvider: kubeoneapi.CloudProviderSpec{
			DigitalOcean: &kubeoneapi.DigitalOceanSpec{
				ControlPlane: &kubeoneapi.DigitalOceanControlPlane{LoadBalancer: lb},
			},
		},
	}
}

func createModeLB() kubeoneapi.DigitalOceanLoadBalancer {
	return kubeoneapi.DigitalOceanLoadBalancer{Name: "demo-kubeapi", Region: "fra1", SizeUnit: 1}
}

func testLogger() logrus.FieldLogger {
	l := logrus.New()
	l.SetOutput(io.Discard)

	return l
}

func TestEnsureLoadBalancerCreates(t *testing.T) {
	fastPolling(t, time.Second)

	fake := &fakeLBService{
		createResult: &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: "new"},
		getResponses: []*godo.LoadBalancer{
			{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: "new"},
			{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: lbStatusUp},
			{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: lbStatusUp, IP: "203.0.113.10"},
		},
	}

	vpcs := testVPCs()
	lb, err := ensureLoadBalancer(context.Background(), fake, vpcs, testCluster(createModeLB()), testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lb.IP != "203.0.113.10" {
		t.Errorf("expected IP of the active load balancer, got %q", lb.IP)
	}
	if len(fake.created) != 1 {
		t.Fatalf("expected one Create call, got %d", len(fake.created))
	}
	if vpcs.getCalls != 0 {
		t.Errorf("create mode must not check VPCs, got %d calls", vpcs.getCalls)
	}
	if fake.getCalls != 3 {
		t.Errorf("expected polling until active with an IP (3 Get calls), got %d", fake.getCalls)
	}

	req := fake.created[0]
	if req.Name != "demo-kubeapi" || req.Region != "fra1" || req.SizeUnit != 1 {
		t.Errorf("unexpected create request: %+v", req)
	}
	if req.Tag != "kubeone-cluster:demo" {
		t.Errorf("expected backends selected by cluster tag, got %q", req.Tag)
	}
	if len(req.ForwardingRules) != 1 || req.ForwardingRules[0] != kubeAPIForwardingRule() {
		t.Errorf("unexpected forwarding rules %+v", req.ForwardingRules)
	}
	if req.HealthCheck == nil || req.HealthCheck.Protocol != "tcp" || req.HealthCheck.Port != kubeAPIPort {
		t.Errorf("unexpected health check %+v", req.HealthCheck)
	}
}

func TestEnsureLoadBalancerIdempotent(t *testing.T) {
	fake := &fakeLBService{
		pages: [][]godo.LoadBalancer{
			{{ID: "other", Name: "other-lb"}},
			{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", ForwardingRules: []godo.ForwardingRule{kubeAPIForwardingRule()}, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.10"}},
		},
	}

	vpcs := testVPCs()
	lb, err := ensureLoadBalancer(context.Background(), fake, vpcs, testCluster(createModeLB()), testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.created) != 0 || len(fake.updated) != 0 {
		t.Errorf("expected no Create or Update call for a configured load balancer, got %d/%d", len(fake.created), len(fake.updated))
	}
	if lb.ID != "lb-1" {
		t.Errorf("expected load balancer from the second page, got %q", lb.ID)
	}
	if !slices.Equal(fake.listRequests, []int{1, 2}) {
		t.Errorf("expected pages 1 and 2 to be listed, got %v", fake.listRequests)
	}
}

// A load balancer found by name in create mode must target the cluster tag,
// otherwise it's not KubeOne's (or was modified) and apply must fail.
func TestEnsureLoadBalancerRejectsForeignSelector(t *testing.T) {
	tests := []struct {
		name string
		lb   godo.LoadBalancer
		want string
	}{
		{
			name: "other tag",
			lb:   godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "my-app", Status: lbStatusUp, IP: "203.0.113.10"},
			want: `droplets tagged "my-app"`,
		},
		{
			name: "droplet ids",
			lb:   godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, DropletIDs: []int{1, 2}, Status: lbStatusUp, IP: "203.0.113.10"},
			want: "2 droplets by ID",
		},
		{
			name: "no backends",
			lb:   godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: lbStatusUp, IP: "203.0.113.10"},
			want: "no droplets",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeLBService{pages: [][]godo.LoadBalancer{{tc.lb}}}

			vpcs := testVPCs()
			_, err := ensureLoadBalancer(context.Background(), fake, vpcs, testCluster(createModeLB()), testLogger())
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, want := range []string{tc.want, `"kubeone-cluster:demo"`, "loadBalancer.name"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("expected error to mention %q, got: %v", want, err)
				}
			}
			if len(fake.created) != 0 || len(fake.updated) != 0 {
				t.Error("must not create or modify the load balancer")
			}
		})
	}
}

// Load balancer names are only unique per region, a cluster with the same name
// in another region must not be used.
func TestEnsureLoadBalancerIgnoresOtherRegion(t *testing.T) {
	fastPolling(t, time.Second)

	fake := &fakeLBService{
		pages:        [][]godo.LoadBalancer{{{ID: "lb-nyc1", Name: "demo-kubeapi", Tag: "kubeone-cluster:demo", Region: &godo.Region{Slug: "nyc1"}, Status: lbStatusUp, IP: "203.0.113.99"}}},
		createResult: &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Status: "new"},
		getResponses: []*godo.LoadBalancer{{ID: "lb-1", Name: "demo-kubeapi", Status: lbStatusUp, IP: "203.0.113.10"}},
	}

	lb, err := ensureLoadBalancer(context.Background(), fake, testVPCs(), testCluster(createModeLB()), testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.created) != 1 || lb.ID != "lb-1" {
		t.Errorf("expected a new load balancer in fra1, got %q (%d creates)", lb.ID, len(fake.created))
	}
}

// A load balancer created by KubeOne is restored if its kube-apiserver rule or
// health check was changed.
func TestEnsureLoadBalancerRepairsOwned(t *testing.T) {
	fake := &fakeLBService{pages: [][]godo.LoadBalancer{{{
		ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", SizeSlug: "lb-small", SizeUnit: 1,
		HealthCheck: &godo.HealthCheck{Protocol: "http", Port: 80, Path: "/"}, DropletIDs: []int{1, 2, 3},
		Status: lbStatusUp, IP: "203.0.113.10",
	}}}}

	if _, err := ensureLoadBalancer(context.Background(), fake, testVPCs(), testCluster(createModeLB()), testLogger()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.created) != 0 || len(fake.updated) != 1 {
		t.Fatalf("expected one Update and no Create, got %d/%d", len(fake.updated), len(fake.created))
	}

	req := fake.updated[0]
	if req.Tag != "kubeone-cluster:demo" || len(req.DropletIDs) != 0 || req.SizeSlug != "" {
		t.Errorf("unexpected update request %+v", req)
	}
	if !isKubeAPIHealthCheck(req.HealthCheck) || !slices.Contains(req.ForwardingRules, kubeAPIForwardingRule()) {
		t.Errorf("expected kube-apiserver rule and health check, got %+v / %+v", req.ForwardingRules, req.HealthCheck)
	}
}

// Right after Create, a load balancer might not be visible yet.
func TestEnsureLoadBalancerCreateRetriesNotFound(t *testing.T) {
	fastPolling(t, time.Second)

	fake := &fakeLBService{
		createResult: &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Status: "new"},
		notFoundGets: 2,
		getResponses: []*godo.LoadBalancer{{ID: "lb-1", Name: "demo-kubeapi", Status: lbStatusUp, IP: "203.0.113.10"}},
	}

	lb, err := ensureLoadBalancer(context.Background(), fake, testVPCs(), testCluster(createModeLB()), testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lb.IP != "203.0.113.10" || fake.getCalls != 3 {
		t.Errorf("expected two 404s to be retried, got IP %q after %d Get calls", lb.IP, fake.getCalls)
	}
}

// An owned load balancer found by name that's still provisioning is waited for
// before it's updated.
func TestEnsureLoadBalancerWaitsBeforeRepair(t *testing.T) {
	fastPolling(t, time.Second)

	fake := &fakeLBService{
		pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", Status: "new"}}},
		getResponses: []*godo.LoadBalancer{
			{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", Status: lbStatusUp, IP: "203.0.113.10"},
		},
	}

	if _, err := ensureLoadBalancer(context.Background(), fake, testVPCs(), testCluster(createModeLB()), testLogger()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.getCalls == 0 {
		t.Error("expected to wait for the load balancer before updating it")
	}
	if len(fake.updated) != 1 {
		t.Fatalf("expected the missing rule and health check to be restored in one Update, got %d", len(fake.updated))
	}
}

// DigitalOcean allows only one rule per entry port, a 6443 rule forwarding
// elsewhere isn't replaced.
func TestEnsureLoadBalancerRejectsChangedRule(t *testing.T) {
	fake := &fakeLBService{pages: [][]godo.LoadBalancer{{{
		ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", Status: lbStatusUp, IP: "203.0.113.10",
		ForwardingRules: []godo.ForwardingRule{{EntryProtocol: "tcp", EntryPort: kubeAPIPort, TargetProtocol: "tcp", TargetPort: 8443}},
	}}}}

	_, err := ensureLoadBalancer(context.Background(), fake, testVPCs(), testCluster(createModeLB()), testLogger())
	if err == nil || !strings.Contains(err.Error(), "kube-apiserver requires tcp/6443 to tcp/6443") {
		t.Fatalf("expected an error about the 6443 rule, got: %v", err)
	}
	if len(fake.updated) != 0 {
		t.Error("must not update the load balancer")
	}
}

// Errors that might be temporary are retried until the timeout.
func TestWaitForLoadBalancerRetriesTemporaryErrors(t *testing.T) {
	fastPolling(t, 20*time.Millisecond)

	for _, status := range []int{http.StatusRequestTimeout, http.StatusConflict, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			fake := &fakeLBService{getErr: errors.New(http.StatusText(status)), getStatus: status}
			if _, err := waitForLoadBalancer(context.Background(), fake, &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Status: "new"}, false, testLogger()); err == nil {
				t.Fatal("expected a timeout error")
			}
			if fake.getCalls < 2 {
				t.Errorf("expected %d to be retried, got %d Get calls", status, fake.getCalls)
			}
		})
	}
}

// godo returns a nil load balancer when Update fails, that must not panic.
func TestReconcileLoadBalancerUpdateError(t *testing.T) {
	fake := &fakeLBService{updateErr: errors.New("load balancer is being updated")}
	lb := &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", Status: lbStatusUp, IP: "203.0.113.10"}

	_, err := reconcileLoadBalancer(context.Background(), fake, lb, "", true, testLogger())
	if err == nil || !strings.Contains(err.Error(), "lb-1") || !strings.Contains(err.Error(), "being updated") {
		t.Fatalf("expected update error for lb-1, got: %v", err)
	}
}

// DigitalOcean doesn't enforce unique load balancer names, the one owned by
// the cluster must win over a foreign one with the same name.
func TestFindLoadBalancerByNamePrefersOwned(t *testing.T) {
	fake := &fakeLBService{pages: [][]godo.LoadBalancer{
		{{ID: "foreign", Name: "demo-kubeapi", Region: fra1Region, Tag: "my-app"}},
		{{ID: "owned", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo"}},
	}}

	lb, err := findLoadBalancerByName(context.Background(), fake, "demo-kubeapi", "fra1", "demo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lb.ID != "owned" {
		t.Errorf("expected the owned load balancer, got %q", lb.ID)
	}

	fake = &fakeLBService{pages: [][]godo.LoadBalancer{{{ID: "foreign", Name: "demo-kubeapi", Region: fra1Region, Tag: "my-app"}}}}
	if lb, err = findLoadBalancerByName(context.Background(), fake, "demo-kubeapi", "fra1", "demo"); err != nil || lb.ID != "foreign" {
		t.Errorf("expected the only match when none is owned, got %v, %v", lb, err)
	}
}

func TestWaitForLoadBalancerRetriesNotFound(t *testing.T) {
	fastPolling(t, 20*time.Millisecond)

	fake := &fakeLBService{getErr: errors.New("not found"), getStatus: http.StatusNotFound}
	_, err := waitForLoadBalancer(context.Background(), fake, &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: "new"}, true, testLogger())
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if fake.getCalls < 2 {
		t.Errorf("expected 404 to be retried right after create, got %d Get calls", fake.getCalls)
	}
}

func TestEnsureLoadBalancerTimeout(t *testing.T) {
	fastPolling(t, 20*time.Millisecond)

	fake := &fakeLBService{
		createResult: &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: "new"},
		getResponses: []*godo.LoadBalancer{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: "new"}},
	}

	vpcs := testVPCs()
	if _, err := ensureLoadBalancer(context.Background(), fake, vpcs, testCluster(createModeLB()), testLogger()); err == nil {
		t.Fatal("expected a timeout error")
	}
}

func TestEnsureExistingLoadBalancer(t *testing.T) {
	kubeAPIRules := []godo.ForwardingRule{kubeAPIForwardingRule()}
	defaultHealthCheck := &godo.HealthCheck{Protocol: "http", Port: 80, Path: "/"}
	readyzHealthCheck := &godo.HealthCheck{Protocol: "https", Port: kubeAPIPort, Path: "/readyz", CheckIntervalSeconds: 10}
	fra1 := &godo.Region{Slug: "fra1"}

	tests := []struct {
		name         string
		existing     godo.LoadBalancer
		wantErr      string
		wantTag      bool
		wantAddRules bool
	}{
		{
			name:     "already configured",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, Tag: "kubeone-cluster:demo", ForwardingRules: kubeAPIRules, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.20"},
		},
		{
			name:         "no backends, no rule",
			existing:     godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, ForwardingRules: []godo.ForwardingRule{{EntryProtocol: "http", EntryPort: 80, TargetProtocol: "http", TargetPort: 80}}, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.20"},
			wantTag:      true,
			wantAddRules: true,
		},
		{
			name:     "https readyz health check is kept",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, ForwardingRules: kubeAPIRules, HealthCheck: readyzHealthCheck, Status: lbStatusUp, IP: "203.0.113.20"},
			wantTag:  true,
		},
		{
			name:     "default health check",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, ForwardingRules: kubeAPIRules, HealthCheck: defaultHealthCheck, Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "health checks http on port 80 with path /",
		},
		{
			name:     "tagged but custom health check",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, Tag: "kubeone-cluster:demo", ForwardingRules: kubeAPIRules, HealthCheck: &godo.HealthCheck{Protocol: "http", Port: 8080, Path: "/status"}, Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "must health check kube-apiserver",
		},
		{
			name:     "https health check on another path",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, ForwardingRules: kubeAPIRules, HealthCheck: &godo.HealthCheck{Protocol: "https", Port: kubeAPIPort, Path: "/"}, Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "must health check kube-apiserver",
		},
		{
			name:     "other backends",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, Tag: "something-else", Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "already has backends",
		},
		{
			name:     "droplet backends",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, DropletIDs: []int{1}, Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "already has backends",
		},
		{
			name:     "6443 forwarded to another port",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, ForwardingRules: []godo.ForwardingRule{{EntryProtocol: "tcp", EntryPort: kubeAPIPort, TargetProtocol: "tcp", TargetPort: 443}}, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "kube-apiserver requires tcp/6443 to tcp/6443",
		},
		{
			name:     "default VPC",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, VPCUUID: "vpc-default", Tag: "kubeone-cluster:demo", ForwardingRules: kubeAPIRules, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.20"},
		},
		{
			name:     "custom VPC",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, VPCUUID: "vpc-custom", Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "default VPC",
		},
		{
			name:     "size slug and size unit",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, SizeSlug: "lb-small", SizeUnit: 1, ForwardingRules: kubeAPIRules, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.20"},
			wantTag:  true,
		},
		{
			// DigitalOcean resolves the tag to droplet IDs in GET responses
			name:         "tagged with resolved droplets",
			existing:     godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, Tag: "kubeone-cluster:demo", DropletIDs: []int{1, 2, 3}, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.20"},
			wantAddRules: true,
		},
		{
			name:     "global load balancer",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Type: godo.LoadBalancerTypeGlobal, Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "global load balancer",
		},
		{
			name:     "other region",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: &godo.Region{Slug: "nyc1"}, Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  `region "nyc1"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeLBService{getResponses: []*godo.LoadBalancer{&tc.existing}}

			vpcs := testVPCs()
			lb, err := ensureLoadBalancer(context.Background(), fake, vpcs, testCluster(kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x", Region: "fra1"}), testLogger())
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tc.wantErr, err)
				}
				if len(fake.updated) != 0 {
					t.Error("must not modify the load balancer when failing")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(fake.created) != 0 {
				t.Error("must never create a load balancer in existing mode")
			}
			if tc.existing.VPCUUID == "" && vpcs.getCalls != 0 {
				t.Error("must not look up a VPC for a load balancer without one")
			}
			wantUpdate := tc.wantTag || tc.wantAddRules
			if got := len(fake.updated) == 1; got != wantUpdate {
				t.Fatalf("update called = %v, want %v", got, wantUpdate)
			}
			if wantUpdate {
				req := fake.updated[0]
				if req.Tag != "kubeone-cluster:demo" {
					t.Errorf("expected update to target the cluster tag, got %q", req.Tag)
				}
				if req.HealthCheck == nil || *req.HealthCheck != *tc.existing.HealthCheck {
					t.Errorf("update must keep the health check %+v, got %+v", tc.existing.HealthCheck, req.HealthCheck)
				}
			}
			for _, req := range fake.updated {
				if req.SizeSlug != "" && req.SizeUnit != 0 {
					t.Errorf("update must not set both size %q and size_unit %d", req.SizeSlug, req.SizeUnit)
				}
				if req.Tag != "" && len(req.DropletIDs) != 0 {
					t.Errorf("update must not set both tag %q and droplet_ids %v", req.Tag, req.DropletIDs)
				}
			}
			if tc.wantAddRules && !slices.Contains(fake.updated[0].ForwardingRules, kubeAPIForwardingRule()) {
				t.Errorf("expected update to add the kube-apiserver forwarding rule, got %+v", fake.updated[0].ForwardingRules)
			}
			if lb.IP != "203.0.113.20" {
				t.Errorf("unexpected IP %q", lb.IP)
			}
		})
	}
}

// DigitalOcean rejects updates while a load balancer is still provisioning.
func TestEnsureExistingLoadBalancerWaitsBeforeUpdate(t *testing.T) {
	fastPolling(t, time.Second)

	fake := &fakeLBService{getResponses: []*godo.LoadBalancer{
		{ID: "lb-x", Name: "byo", Region: &godo.Region{Slug: "fra1"}, Status: "new"},
		{ID: "lb-x", Name: "byo", Region: &godo.Region{Slug: "fra1"}, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.20"},
	}}

	if _, err := ensureLoadBalancer(context.Background(), fake, testVPCs(), testCluster(kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x", Region: "fra1"}), testLogger()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.getCalls < 2 {
		t.Errorf("expected to wait for the load balancer before updating, got %d Get calls", fake.getCalls)
	}
	if len(fake.updated) != 1 {
		t.Fatalf("expected one Update call, got %d", len(fake.updated))
	}
}

func TestWaitForLoadBalancerFailsFast(t *testing.T) {
	fastPolling(t, time.Minute)

	tests := []struct {
		name string
		fake *fakeLBService
		want string
	}{
		{
			name: "errored",
			fake: &fakeLBService{getResponses: []*godo.LoadBalancer{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: lbStatusErr}}},
			want: `"errored" state`,
		},
		{
			name: "forbidden",
			fake: &fakeLBService{getErr: errors.New("forbidden"), getStatus: http.StatusForbidden},
			want: "forbidden",
		},
		{
			name: "bad request",
			fake: &fakeLBService{getErr: errors.New("bad request"), getStatus: http.StatusBadRequest},
			want: "bad request",
		},
		{
			// Only retried right after create
			name: "not found",
			fake: &fakeLBService{getErr: errors.New("not found"), getStatus: http.StatusNotFound},
			want: "not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := waitForLoadBalancer(context.Background(), tc.fake, &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Status: "new"}, false, testLogger())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got: %v", tc.want, err)
			}
			if tc.fake.getCalls != 1 {
				t.Errorf("expected to fail on the first Get, got %d calls", tc.fake.getCalls)
			}
		})
	}
}

func fakeClient(t *testing.T, fake *fakeLBService) {
	t.Helper()

	t.Setenv("DIGITALOCEAN_TOKEN", "test-token")
	old := newClients
	newClients = func(string) (godo.LoadBalancersService, vpcGetter) { return fake, testVPCs() }
	t.Cleanup(func() { newClients = old })
}

// With .apiEndpoint.host set the load balancer is managed outside of KubeOne.
func TestAPIEndpointHostSkipsLoadBalancer(t *testing.T) {
	p := &provider{}
	fake := &fakeLBService{}
	fakeClient(t, fake)

	s := testState(&kubeoneapi.DigitalOceanLoadBalancer{Name: "demo-kubeapi"}, testNodeSets(validSpec, 1))
	s.Logger = testLogger()
	s.Cluster.APIEndpoint.Host = "api.example.com"

	if err := p.EnsureLoadBalancer(s); err != nil {
		t.Fatalf("EnsureLoadBalancer: unexpected error: %v", err)
	}
	if err := p.LookupLoadBalancer(s); err != nil {
		t.Fatalf("LookupLoadBalancer: unexpected error: %v", err)
	}
	if len(fake.listRequests) != 0 || fake.getCalls != 0 || len(fake.created) != 0 {
		t.Error("expected no API calls")
	}
	if s.Cluster.APIEndpoint.Host != "api.example.com" {
		t.Errorf("host must not change, got %q", s.Cluster.APIEndpoint.Host)
	}
}

func TestLookupLoadBalancerSetsEndpoint(t *testing.T) {
	p := &provider{}
	fake := &fakeLBService{pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", IP: "203.0.113.10"}}}}
	fakeClient(t, fake)

	s := testState(&kubeoneapi.DigitalOceanLoadBalancer{Name: "demo-kubeapi"}, testNodeSets(validSpec, 1))
	s.Logger = testLogger()
	s.Cluster.APIEndpoint.Port = 443

	if err := p.LookupLoadBalancer(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Cluster.APIEndpoint.Host != "203.0.113.10" || s.Cluster.APIEndpoint.Port != kubeAPIPort {
		t.Errorf("unexpected endpoint %s:%d", s.Cluster.APIEndpoint.Host, s.Cluster.APIEndpoint.Port)
	}
}

func TestLookupLoadBalancer(t *testing.T) {
	t.Run("by name", func(t *testing.T) {
		fake := &fakeLBService{pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, IP: "203.0.113.10"}}}}

		lb, err := lookupLoadBalancer(context.Background(), fake, testCluster(createModeLB()))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if lb.ID != "lb-1" {
			t.Errorf("unexpected load balancer %q", lb.ID)
		}
	})

	t.Run("by name not found", func(t *testing.T) {
		fake := &fakeLBService{pages: [][]godo.LoadBalancer{{{ID: "other", Name: "other"}}}}

		if _, err := lookupLoadBalancer(context.Background(), fake, testCluster(createModeLB())); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("by id not found", func(t *testing.T) {
		fake := &fakeLBService{getErr: errors.New("not found"), getStatus: http.StatusNotFound}

		if _, err := lookupLoadBalancer(context.Background(), fake, testCluster(kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x"})); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestCleanupLoadBalancer(t *testing.T) {
	tests := []struct {
		name        string
		pages       [][]godo.LoadBalancer
		host        string
		wantDeleted []string
	}{
		{
			name:        "host set by lookup to the load balancer IP",
			pages:       [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", IP: "203.0.113.10"}}},
			host:        "203.0.113.10",
			wantDeleted: []string{"lb-1"},
		},
		{
			name:  "user provided apiEndpoint.host",
			pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo", IP: "203.0.113.10"}}},
			host:  "api.example.com",
		},
		{
			name:        "owned load balancer is deleted",
			pages:       [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "kubeone-cluster:demo"}}},
			wantDeleted: []string{"lb-1"},
		},
		{
			name:  "unowned load balancer is kept",
			pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Region: fra1Region, Tag: "someone-else"}}},
		},
		{
			// DigitalOcean doesn't persist tags on load balancers, only the
			// backend selector tag can be relied upon
			name:  "tags are not an ownership marker",
			pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Tags: []string{"kubeone-owned:demo"}}}},
		},
		{
			name:  "same name in another region is kept",
			pages: [][]godo.LoadBalancer{{{ID: "lb-nyc1", Name: "demo-kubeapi", Tag: "kubeone-cluster:demo", Region: &godo.Region{Slug: "nyc1"}, IP: "203.0.113.99"}}},
		},
		{
			name:  "missing load balancer is a no-op",
			pages: [][]godo.LoadBalancer{{{ID: "other", Name: "other"}}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeLBService{pages: tc.pages}
			cluster := testCluster(createModeLB())
			cluster.APIEndpoint.Host = tc.host

			if err := cleanupLoadBalancer(context.Background(), fake, cluster, testLogger()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(fake.deleted, tc.wantDeleted) {
				t.Errorf("deleted %v, want %v", fake.deleted, tc.wantDeleted)
			}
		})
	}
}

// A load balancer referenced by ID is never deleted, reset only removes the
// cluster tag backend selector KubeOne set when configuring it.
func TestCleanupLoadBalancerReleasesExistingMode(t *testing.T) {
	readyzHealthCheck := &godo.HealthCheck{Protocol: "https", Port: kubeAPIPort, Path: "/readyz"}
	rules := []godo.ForwardingRule{kubeAPIForwardingRule()}

	tests := []struct {
		name        string
		existing    *godo.LoadBalancer
		getStatus   int
		wantRelease bool
	}{
		{
			name:        "cluster tag is removed",
			existing:    &godo.LoadBalancer{ID: "lb-x", Name: "byo", SizeSlug: "lb-small", SizeUnit: 1, Tag: "kubeone-cluster:demo", DropletIDs: []int{1, 2}, ForwardingRules: rules, HealthCheck: readyzHealthCheck, Status: lbStatusUp, IP: "203.0.113.20"},
			wantRelease: true,
		},
		{
			name:     "other tag is kept",
			existing: &godo.LoadBalancer{ID: "lb-x", Name: "byo", Tag: "someone-else", Status: lbStatusUp, IP: "203.0.113.20"},
		},
		{
			name:     "no backends",
			existing: &godo.LoadBalancer{ID: "lb-x", Name: "byo", Status: lbStatusUp, IP: "203.0.113.20"},
		},
		{
			name:      "missing load balancer is a no-op",
			getStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &provider{}

			fake := &fakeLBService{getResponses: []*godo.LoadBalancer{tc.existing}}
			if tc.getStatus != 0 {
				fake.getErr, fake.getStatus = errors.New(http.StatusText(tc.getStatus)), tc.getStatus
			}
			fakeClient(t, fake)

			s := testState(&kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x"}, testNodeSets(validSpec, 1))
			s.Logger = testLogger()
			if err := p.CleanupLoadBalancer(s); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fake.deleted) != 0 || len(fake.listRequests) != 0 {
				t.Error("must never list or delete a load balancer referenced by ID")
			}
			if got := len(fake.updated) == 1; got != tc.wantRelease {
				t.Fatalf("update called = %v, want %v", got, tc.wantRelease)
			}
			if !tc.wantRelease {
				return
			}

			req := fake.updated[0]
			if req.Tag != "" || len(req.DropletIDs) != 0 {
				t.Errorf("expected update to remove the backends, got tag %q and droplets %v", req.Tag, req.DropletIDs)
			}
			if req.HealthCheck == nil || *req.HealthCheck != *readyzHealthCheck || !slices.Equal(req.ForwardingRules, rules) {
				t.Errorf("update must keep the health check and forwarding rules, got %+v and %+v", req.HealthCheck, req.ForwardingRules)
			}
			if req.SizeSlug != "" && req.SizeUnit != 0 {
				t.Errorf("update must not set both size %q and size_unit %d", req.SizeSlug, req.SizeUnit)
			}
		})
	}
}

// TestReleaseLoadBalancerWire checks the update request body sent to the
// DigitalOcean API when releasing a load balancer referenced by ID.
func TestReleaseLoadBalancerWire(t *testing.T) {
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/load_balancers/lb-x":
			// DigitalOcean returns the droplets resolved from the tag
			_, _ = w.Write([]byte(`{"load_balancer":{"id":"lb-x","name":"byo","status":"active","ip":"203.0.113.20",` +
				`"size_unit":1,"region":{"slug":"fra1"},"tag":"kubeone-cluster:demo","droplet_ids":[1,2],` +
				`"forwarding_rules":[{"entry_protocol":"tcp","entry_port":6443,"target_protocol":"tcp","target_port":6443}],` +
				`"health_check":{"protocol":"tcp","port":6443}}}`))
		case r.Method == http.MethodPut && r.URL.Path == "/v2/load_balancers/lb-x":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
			_, _ = w.Write([]byte(`{"load_balancer":{"id":"lb-x","name":"byo","status":"active","ip":"203.0.113.20"}}`))
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	client, err := godo.New(srv.Client(), godo.SetBaseURL(srv.URL+"/"))
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	cluster := testCluster(kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x", Region: "fra1"})
	if err = releaseLoadBalancer(context.Background(), client.LoadBalancers, cluster, testLogger()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body == nil {
		t.Fatal("expected an update request")
	}
	for _, key := range []string{"tag", "droplet_ids"} {
		if _, ok := body[key]; ok {
			t.Errorf("update must not send %s, got %v", key, body)
		}
	}
	hc, _ := body["health_check"].(map[string]any)
	if hc["protocol"] != "tcp" || hc["port"] != float64(kubeAPIPort) {
		t.Errorf("unexpected health check %v", hc)
	}
}

// TestCreateRequestWire checks the request body sent to the DigitalOcean API.
func TestCreateRequestWire(t *testing.T) {
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/load_balancers" {
			http.Error(w, "unexpected request", http.StatusBadRequest)

			return
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"load_balancer":{"id":"lb-1","name":"demo-kubeapi","status":"new"}}`))
	}))
	defer srv.Close()

	client, err := godo.New(srv.Client(), godo.SetBaseURL(srv.URL+"/"))
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	if _, _, err = client.LoadBalancers.Create(context.Background(), createRequest(testCluster(createModeLB()))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body["tag"] != "kubeone-cluster:demo" || body["region"] != "fra1" || body["name"] != "demo-kubeapi" {
		t.Errorf("unexpected request body %v", body)
	}
	rules, ok := body["forwarding_rules"].([]any)
	if !ok || len(rules) != 1 {
		t.Fatalf("unexpected forwarding rules %v", body["forwarding_rules"])
	}
	rule, _ := rules[0].(map[string]any)
	if rule["entry_port"] != float64(kubeAPIPort) || rule["entry_protocol"] != "tcp" {
		t.Errorf("unexpected forwarding rule %v", rule)
	}
}

// TestExistingLoadBalancerUpdateWire checks the update request body sent to
// the DigitalOcean API for a load balancer referenced by ID.
func TestExistingLoadBalancerUpdateWire(t *testing.T) {
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/load_balancers/lb-x":
			// DigitalOcean returns both size fields and the tagged droplets
			_, _ = w.Write([]byte(`{"load_balancer":{"id":"lb-x","name":"byo","status":"active","ip":"203.0.113.20",` +
				`"size":"lb-small","size_unit":1,"region":{"slug":"fra1"},"droplet_ids":[],` +
				`"forwarding_rules":[{"entry_protocol":"http","entry_port":80,"target_protocol":"http","target_port":80}],` +
				`"health_check":{"protocol":"https","port":6443,"path":"/readyz"}}}`))
		case r.Method == http.MethodPut && r.URL.Path == "/v2/load_balancers/lb-x":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
			_, _ = w.Write([]byte(`{"load_balancer":{"id":"lb-x","name":"byo","status":"active","ip":"203.0.113.20"}}`))
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	client, err := godo.New(srv.Client(), godo.SetBaseURL(srv.URL+"/"))
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	cluster := testCluster(kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x", Region: "fra1"})
	if _, err = ensureLoadBalancer(context.Background(), client.LoadBalancers, testVPCs(), cluster, testLogger()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body == nil {
		t.Fatal("expected an update request")
	}
	if _, ok := body["size"]; ok {
		t.Errorf("update must not send size together with size_unit, got %v", body)
	}
	if _, ok := body["droplet_ids"]; ok {
		t.Errorf("update must not send droplet_ids together with tag, got %v", body)
	}
	if body["tag"] != "kubeone-cluster:demo" || body["size_unit"] != float64(1) {
		t.Errorf("unexpected update body %v", body)
	}
	hc, _ := body["health_check"].(map[string]any)
	if hc["protocol"] != "https" || hc["port"] != float64(kubeAPIPort) || hc["path"] != "/readyz" {
		t.Errorf("update must keep the health check, got %v", hc)
	}
}
