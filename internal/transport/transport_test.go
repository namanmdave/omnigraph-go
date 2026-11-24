package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewValidation(t *testing.T) {
	cases := []string{
		"",
		"not-a-url",
		"ftp://example.com",
		"http://user:pass@example.com",
		"http://example.com?foo=bar",
		"http://example.com#frag",
	}
	for _, raw := range cases {
		if _, err := New(Config{BaseURL: raw}); err == nil {
			t.Fatalf("expected error for baseURL %q", raw)
		}
	}
}

func TestGraphScopingAndPublicCredentials(t *testing.T) {
	var seenHealthAuth string
	var seenQueryAuth string
	var seenQueryPath string
	var seenContract string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContractHeader, "0.14")
		switch r.URL.Path {
		case "/healthz":
			seenHealthAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","version":"0.14.0","internal_schema_version":16}`))
		case "/graphs/alpha/query":
			seenQueryPath = r.URL.Path
			seenQueryAuth = r.Header.Get("Authorization")
			seenContract = r.Header.Get(ContractHeader)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"query_name":"q","row_count":0,"rows":[],"target":{"branch":"main"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, Token: "secret-token"})
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	err = c.RequestJSON(context.Background(), http.MethodPost, "/query", RequestOptions{Body: map[string]any{"query": "q"}}, &out)
	var tErr *TransportError
	if !errors.As(err, &tErr) || tErr.Kind != KindConfiguration {
		t.Fatalf("expected KindConfiguration without graphID, got %v", err)
	}

	scoped := c.WithGraphID("alpha")
	if err := scoped.RequestJSON(context.Background(), http.MethodPost, "/query", RequestOptions{Body: map[string]any{"query": "q"}}, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if seenHealthAuth != "" {
		t.Fatalf("expected no Authorization on /healthz, got %q", seenHealthAuth)
	}
	if seenQueryAuth != "Bearer secret-token" {
		t.Fatalf("expected bearer token on /query, got %q", seenQueryAuth)
	}
	if seenQueryPath != "/graphs/alpha/query" {
		t.Fatalf("expected /graphs/alpha/query, got %q", seenQueryPath)
	}
	if seenContract != "0.14" {
		t.Fatalf("expected contract 0.14, got %q", seenContract)
	}
}

func TestDiscoveryContractMismatchFailsClosed(t *testing.T) {
	var dataCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.Header().Set(ContractHeader, "0.13")
			w.WriteHeader(http.StatusOK)
			return
		}
		dataCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, GraphID: "alpha"})
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	err = c.RequestJSON(context.Background(), http.MethodPost, "/mutate", RequestOptions{Body: map[string]any{"query": "m"}}, &out)
	var tErr *TransportError
	if !errors.As(err, &tErr) || tErr.Kind != KindAPIContract || tErr.RequestDispatched {
		t.Fatalf("expected fail-closed APIContract error before dispatch, got %+v", err)
	}
	if dataCalls.Load() != 0 {
		t.Fatalf("expected 0 data calls after discovery failure, got %d", dataCalls.Load())
	}
}

func TestOutcomeUnknownOnWriteDisposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContractHeader, "0.14")
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/graphs/alpha/mutate" {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"code":"conflict","error":"stale","disposition":{"effects":"none","action":"retry"}}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, GraphID: "alpha"})
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	err = c.RequestJSON(context.Background(), http.MethodPost, "/mutate", RequestOptions{Body: map[string]any{"query": "m"}}, &out)
	var tErr *TransportError
	if !errors.As(err, &tErr) || tErr.OutcomeUnknown {
		t.Fatalf("expected OutcomeUnknown=false when disposition.effects=none, got %+v", err)
	}

	err = c.RequestJSON(context.Background(), http.MethodPost, "/branches", RequestOptions{Body: map[string]any{"name": "dev"}}, &out)
	if !errors.As(err, &tErr) || !tErr.OutcomeUnknown {
		t.Fatalf("expected OutcomeUnknown=true when write fails without disposition, got %+v", err)
	}
}

func TestConcurrentDiscovery(t *testing.T) {
	var healthCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContractHeader, "0.14")
		if r.URL.Path == "/healthz" {
			healthCalls.Add(1)
			w.WriteHeader(http.StatusOK)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, GraphID: "alpha"})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var out map[string]any
			if err := c.RequestJSON(context.Background(), http.MethodGet, "/schema", RequestOptions{}, &out); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if healthCalls.Load() != 1 {
		t.Fatalf("expected exactly 1 discovery call under concurrency, got %d", healthCalls.Load())
	}
}

func TestDuplicateContractHeaderRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(ContractHeader, "0.14")
		w.Header().Add(ContractHeader, "0.14")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, GraphID: "alpha", SkipDiscovery: true})
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	err = c.RequestJSON(context.Background(), http.MethodGet, "/schema", RequestOptions{}, &out)
	var tErr *TransportError
	if !errors.As(err, &tErr) || tErr.Kind != KindAPIContract {
		t.Fatalf("expected KindAPIContract for duplicate contract header, got %v", err)
	}
}

func TestRedirectsNotFollowedAndNoTokenLeak(t *testing.T) {
	var targetHits atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContractHeader, "0.14")
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, target.URL+"/leak", http.StatusFound)
	}))
	defer srv.Close()

	c, err := New(Config{
		BaseURL:    srv.URL,
		Token:      "top-secret",
		GraphID:    "alpha",
		HTTPClient: &http.Client{},
	})
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	err = c.RequestJSON(context.Background(), http.MethodGet, "/schema", RequestOptions{}, &out)
	var tErr *TransportError
	if !errors.As(err, &tErr) || tErr.Status != http.StatusFound {
		t.Fatalf("expected unfollowed 302 error, got %v", err)
	}
	if targetHits.Load() != 0 {
		t.Fatalf("redirect target should never be hit, got %d hits", targetHits.Load())
	}
}

func TestPublicPathsStripCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContractHeader, "0.14")
		if r.Header.Get("Authorization") != "" || r.Header.Get(IfActorHeader) != "" {
			t.Errorf("credentials leaked on public path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, Token: "secret", SkipDiscovery: true})
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{"/healthz", "/readyz", "/.well-known/oauth-protected-resource", "/.well-known/oauth-protected-resource/alpha"} {
		var out map[string]any
		if err := c.RequestJSON(context.Background(), http.MethodGet, p, RequestOptions{IfActor: "alice"}, &out); err != nil {
			t.Fatalf("path %s failed: %v", p, err)
		}
	}
}

func TestDiscoveryRecoversAfterTransientFailure(t *testing.T) {
	var healthCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContractHeader, "0.14")
		if r.URL.Path == "/healthz" {
			if healthCalls.Add(1) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		_, _ = w.Write([]byte(`{"schema_source":"type A { id: String }"}`))
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, GraphID: "alpha"})
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	if err := c.RequestJSON(context.Background(), http.MethodGet, "/schema", RequestOptions{}, &out); err == nil {
		t.Fatal("expected first call to fail during transient discovery failure")
	}
	if err := c.RequestJSON(context.Background(), http.MethodGet, "/schema", RequestOptions{}, &out); err != nil {
		t.Fatalf("expected recovery on second call, got %v", err)
	}
}

func TestPathEscaping(t *testing.T) {
	var seenPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContractHeader, "0.14")
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		seenPath = r.URL.RawPath
		if seenPath == "" {
			seenPath = r.URL.Path
		}
		_, _ = w.Write([]byte(`{"schema_source":"ok"}`))
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, GraphID: "team/alpha 1#x"})
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	if err := c.RequestJSON(context.Background(), http.MethodGet, "/schema", RequestOptions{}, &out); err != nil {
		t.Fatal(err)
	}
	if seenPath != "/graphs/team%2Falpha%201%23x/schema" {
		t.Fatalf("unexpected escaped path: %q", seenPath)
	}
}
