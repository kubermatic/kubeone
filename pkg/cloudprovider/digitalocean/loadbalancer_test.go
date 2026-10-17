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

	createResult *godo.LoadBalancer

	created      []*godo.LoadBalancerRequest
	updated      []*godo.LoadBalancerRequest
	deleted      []string
	addedRules   []godo.ForwardingRule
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
		pages.Prev = "https://api.digitalocean.com/v2/load_balancers?page=" + itoa(page-1)
	}
	if page < len(f.pages) {
		pages.Next = "https://api.digitalocean.com/v2/load_balancers?page=" + itoa(page+1)
	}

	return f.pages[page-1], &godo.Response{Links: &godo.Links{Pages: pages}}, nil
}

func (f *fakeLBService) Get(_ context.Context, _ string) (*godo.LoadBalancer, *godo.Response, error) {
	f.getCalls++
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

	return &godo.LoadBalancer{ID: id, Name: req.Name, Tag: req.Tag, ForwardingRules: req.ForwardingRules, HealthCheck: req.HealthCheck, Status: lbStatusUp, IP: "203.0.113.20"}, &godo.Response{}, nil
}

func (f *fakeLBService) Delete(_ context.Context, id string) (*godo.Response, error) {
	f.deleted = append(f.deleted, id)

	return &godo.Response{}, nil
}

func (f *fakeLBService) AddForwardingRules(_ context.Context, _ string, rules ...godo.ForwardingRule) (*godo.Response, error) {
	f.addedRules = append(f.addedRules, rules...)

	return &godo.Response{}, nil
}

func itoa(i int) string {
	b, _ := json.Marshal(i)

	return string(b)
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
		createResult: &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Status: "new"},
		getResponses: []*godo.LoadBalancer{
			{ID: "lb-1", Name: "demo-kubeapi", Status: "new"},
			{ID: "lb-1", Name: "demo-kubeapi", Status: lbStatusUp},
			{ID: "lb-1", Name: "demo-kubeapi", Status: lbStatusUp, IP: "203.0.113.10"},
		},
	}

	lb, err := ensureLoadBalancer(context.Background(), fake, testCluster(createModeLB()), testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lb.IP != "203.0.113.10" {
		t.Errorf("expected IP of the active load balancer, got %q", lb.IP)
	}
	if len(fake.created) != 1 {
		t.Fatalf("expected one Create call, got %d", len(fake.created))
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
			{{ID: "lb-1", Name: "demo-kubeapi", Tag: "kubeone-cluster:demo", Status: lbStatusUp, IP: "203.0.113.10"}},
		},
	}

	lb, err := ensureLoadBalancer(context.Background(), fake, testCluster(createModeLB()), testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.created) != 0 {
		t.Errorf("expected no Create call for an existing load balancer, got %d", len(fake.created))
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
			lb:   godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Tag: "my-app", Status: lbStatusUp, IP: "203.0.113.10"},
			want: `droplets tagged "my-app"`,
		},
		{
			name: "droplet ids",
			lb:   godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", DropletIDs: []int{1, 2}, Status: lbStatusUp, IP: "203.0.113.10"},
			want: "2 droplets by ID",
		},
		{
			name: "no backends",
			lb:   godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Status: lbStatusUp, IP: "203.0.113.10"},
			want: "no droplets",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeLBService{pages: [][]godo.LoadBalancer{{tc.lb}}}

			_, err := ensureLoadBalancer(context.Background(), fake, testCluster(createModeLB()), testLogger())
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

func TestEnsureLoadBalancerRejectsOtherRegion(t *testing.T) {
	fake := &fakeLBService{pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Tag: "kubeone-cluster:demo", Region: &godo.Region{Slug: "nyc1"}, Status: lbStatusUp, IP: "203.0.113.10"}}}}

	_, err := ensureLoadBalancer(context.Background(), fake, testCluster(createModeLB()), testLogger())
	if err == nil || !strings.Contains(err.Error(), `region "nyc1"`) {
		t.Fatalf("expected region mismatch error, got: %v", err)
	}
}

func TestWaitForLoadBalancerRetriesNotFound(t *testing.T) {
	fastPolling(t, 20*time.Millisecond)

	fake := &fakeLBService{getErr: errors.New("not found"), getStatus: http.StatusNotFound}
	_, err := waitForLoadBalancer(context.Background(), fake, &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Status: "new"}, testLogger())
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if fake.getCalls < 2 {
		t.Errorf("expected 404 to be retried, got %d Get calls", fake.getCalls)
	}
}

func TestEnsureLoadBalancerTimeout(t *testing.T) {
	fastPolling(t, 20*time.Millisecond)

	fake := &fakeLBService{
		createResult: &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Status: "new"},
		getResponses: []*godo.LoadBalancer{{ID: "lb-1", Name: "demo-kubeapi", Status: "new"}},
	}

	if _, err := ensureLoadBalancer(context.Background(), fake, testCluster(createModeLB()), testLogger()); err == nil {
		t.Fatal("expected a timeout error")
	}
}

func TestEnsureExistingLoadBalancer(t *testing.T) {
	kubeAPIRules := []godo.ForwardingRule{kubeAPIForwardingRule()}
	defaultHealthCheck := &godo.HealthCheck{Protocol: "http", Port: 80, Path: "/"}
	fra1 := &godo.Region{Slug: "fra1"}

	tests := []struct {
		name            string
		existing        godo.LoadBalancer
		wantErr         string
		wantTag         bool
		wantHealthCheck bool
		wantAddRules    bool
	}{
		{
			name:     "already configured",
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, Tag: "kubeone-cluster:demo", ForwardingRules: kubeAPIRules, HealthCheck: kubeAPIHealthCheck(), Status: lbStatusUp, IP: "203.0.113.20"},
		},
		{
			name:            "no backends, no rule, default health check",
			existing:        godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, ForwardingRules: []godo.ForwardingRule{{EntryProtocol: "http", EntryPort: 80, TargetProtocol: "http", TargetPort: 80}}, HealthCheck: defaultHealthCheck, Status: lbStatusUp, IP: "203.0.113.20"},
			wantTag:         true,
			wantHealthCheck: true,
			wantAddRules:    true,
		},
		{
			name:            "tagged but default health check",
			existing:        godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, Tag: "kubeone-cluster:demo", ForwardingRules: kubeAPIRules, HealthCheck: defaultHealthCheck, Status: lbStatusUp, IP: "203.0.113.20"},
			wantHealthCheck: true,
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
			existing: godo.LoadBalancer{ID: "lb-x", Name: "byo", Region: fra1, ForwardingRules: []godo.ForwardingRule{{EntryProtocol: "tcp", EntryPort: kubeAPIPort, TargetProtocol: "tcp", TargetPort: 443}}, Status: lbStatusUp, IP: "203.0.113.20"},
			wantErr:  "kube-apiserver requires tcp/6443 to tcp/6443",
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

			lb, err := ensureLoadBalancer(context.Background(), fake, testCluster(kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x", Region: "fra1"}), testLogger())
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tc.wantErr, err)
				}
				if len(fake.updated) != 0 || len(fake.addedRules) != 0 {
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
			wantUpdate := tc.wantTag || tc.wantHealthCheck || tc.wantAddRules
			if got := len(fake.updated) == 1; got != wantUpdate {
				t.Fatalf("update called = %v, want %v", got, wantUpdate)
			}
			if wantUpdate {
				req := fake.updated[0]
				if req.Tag != "kubeone-cluster:demo" {
					t.Errorf("expected update to target the cluster tag, got %q", req.Tag)
				}
				if !isKubeAPIHealthCheck(req.HealthCheck) {
					t.Errorf("expected update to health check tcp/6443, got %+v", req.HealthCheck)
				}
			}
			if len(fake.addedRules) != 0 {
				t.Error("forwarding rule must be added in the same Update, not with AddForwardingRules")
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

func TestWaitForLoadBalancerFailsFast(t *testing.T) {
	fastPolling(t, time.Minute)

	tests := []struct {
		name string
		fake *fakeLBService
		want string
	}{
		{
			name: "errored",
			fake: &fakeLBService{getResponses: []*godo.LoadBalancer{{ID: "lb-1", Name: "demo-kubeapi", Status: lbStatusErr}}},
			want: `"errored" state`,
		},
		{
			name: "forbidden",
			fake: &fakeLBService{getErr: errors.New("forbidden"), getStatus: http.StatusForbidden},
			want: "forbidden",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := waitForLoadBalancer(context.Background(), tc.fake, &godo.LoadBalancer{ID: "lb-1", Name: "demo-kubeapi", Status: "new"}, testLogger())
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
	old := newLBClient
	newLBClient = func(string) godo.LoadBalancersService { return fake }
	t.Cleanup(func() { newLBClient = old })
}

// With .apiEndpoint.host set the load balancer is managed outside of KubeOne.
func TestAPIEndpointHostSkipsLoadBalancer(t *testing.T) {
	p := &Provider{}
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
	p := &Provider{}
	fake := &fakeLBService{pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Tag: "kubeone-cluster:demo", IP: "203.0.113.10"}}}}
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
		fake := &fakeLBService{pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", IP: "203.0.113.10"}}}}

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
			pages:       [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Tag: "kubeone-cluster:demo", IP: "203.0.113.10"}}},
			host:        "203.0.113.10",
			wantDeleted: []string{"lb-1"},
		},
		{
			name:  "user provided apiEndpoint.host",
			pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Tag: "kubeone-cluster:demo", IP: "203.0.113.10"}}},
			host:  "api.example.com",
		},
		{
			name:        "owned load balancer is deleted",
			pages:       [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Tag: "kubeone-cluster:demo"}}},
			wantDeleted: []string{"lb-1"},
		},
		{
			name:  "unowned load balancer is kept",
			pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Tag: "someone-else"}}},
		},
		{
			// DigitalOcean doesn't persist tags on load balancers, only the
			// backend selector tag can be relied upon
			name:  "tags are not an ownership marker",
			pages: [][]godo.LoadBalancer{{{ID: "lb-1", Name: "demo-kubeapi", Tags: []string{"kubeone-owned:demo"}}}},
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

func TestCleanupLoadBalancerSkipsExistingMode(t *testing.T) {
	p := &Provider{}

	fake := &fakeLBService{}
	fakeClient(t, fake)

	s := testState(&kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x"}, testNodeSets(validSpec, 1))
	s.Logger = testLogger()
	if err := p.CleanupLoadBalancer(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.deleted) != 0 || len(fake.listRequests) != 0 || fake.getCalls != 0 {
		t.Error("expected no API calls for a load balancer referenced by ID")
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
