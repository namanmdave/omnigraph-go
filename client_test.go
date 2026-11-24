package omnigraph_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	omnigraph "github.com/namanmdave/omnigraph-go"
)

func newMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(omnigraph.HTTPAPIContractHeader, omnigraph.HTTPAPIContract)
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","version":"0.14.0","internal_schema_version":16}`))
			return
		}
		handler(w, r)
	}))
}

func TestClientGraphScopingAndQueries(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/graphs":
			_, _ = w.Write([]byte(`{"graphs":[{"graph_id":"alpha","uri":"/graphs/alpha","state":"ready","action":"none","read_available":true,"write_available":true}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/graphs/alpha/query":
			_, _ = w.Write([]byte(`{"query_name":"find","row_count":1,"rows":[{"name":"Alice"}],"target":{"branch":"main"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/graphs/alpha/mutate/if-graph-commit":
			if r.Header.Get(omnigraph.IfGraphCommitHeader) != "c123" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"query_name":"add","branch":"main","affected_nodes":1,"affected_edges":0,"commit":{"graph_commit_id":"c124"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":"not_found","error":"missing","disposition":{"effects":"none","action":"fail"}}`))
		}
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithToken("tok"))
	if err != nil {
		t.Fatal(err)
	}

	graphs, err := client.Graphs().List(context.Background())
	if err != nil || len(graphs.Graphs) != 1 || graphs.Graphs[0].GraphID != "alpha" {
		t.Fatalf("unexpected graphs response: %+v, err=%v", graphs, err)
	}

	alpha := client.Graph("alpha")
	res, err := alpha.Query(context.Background(), omnigraph.QueryInput{
		Query:  "query find() { match { $p: Person } return { $p.name } }",
		Branch: "main",
	})
	if err != nil || res.RowCount != 1 {
		t.Fatalf("unexpected query result: %+v, err=%v", res, err)
	}

	mut, err := alpha.Mutate(context.Background(), omnigraph.MutationInput{
		Query: "mutate add { create { (:Person { name: 'Bob' }) } }",
	}, omnigraph.ConditionalCallOptions{IfGraphCommit: "c123"})
	if err != nil || mut.AffectedNodes != 1 {
		t.Fatalf("unexpected mutate result: %+v, err=%v", mut, err)
	}

	_, err = alpha.Commits().Show(context.Background(), "missing")
	if !errors.Is(err, omnigraph.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBlobsAndResources(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/graphs/alpha/blob":
			_, _ = w.Write([]byte(`{"kind":"written","branch":"main","selector":{},"commit":{}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/graphs/alpha/blob":
			w.Header().Set("Location", "https://cdn.example.com/blob")
			w.WriteHeader(http.StatusFound)
		case r.Method == http.MethodPost && r.URL.Path == "/graphs/alpha/branches":
			_, _ = w.Write([]byte(`{"name":"feat","uri":"/graphs/alpha/branches/feat","from":"main","from_commit_id":"c1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/graphs/alpha/schema":
			_, _ = w.Write([]byte(`{"schema_source":"type Person { name: String }"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	putOut, err := client.Blobs().Put(context.Background(), omnigraph.BlobWriteInput{
		Type:  "Person",
		Key:   "alice",
		Field: "avatar",
	}, strings.NewReader("png-bytes"), "image/png")
	if err != nil || putOut.Kind != "written" {
		t.Fatalf("unexpected blob put: %+v, err=%v", putOut, err)
	}

	getResp, err := client.Blobs().Get(context.Background(), omnigraph.BlobInput{
		Type:  "Person",
		Key:   "alice",
		Field: "avatar",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 redirect preserved, got %d", getResp.StatusCode)
	}

	br, err := client.Branches().Create(context.Background(), omnigraph.BranchCreateInput{Name: "feat"})
	if err != nil || br.Name != "feat" {
		t.Fatalf("unexpected branch create: %+v, err=%v", br, err)
	}

	sc, err := client.Schema().Show(context.Background())
	if err != nil || !strings.Contains(sc.SchemaSource, "Person") {
		t.Fatalf("unexpected schema: %+v, err=%v", sc, err)
	}
}

func TestOpaqueUserKeysPreserved(t *testing.T) {
	var seenParams map[string]any
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		seenParams, _ = body["params"].(map[string]any)
		_, _ = w.Write([]byte(`{"query_name":"q","row_count":1,"rows":[{"user_Name":"Alice","camelKey":42}],"target":{"branch":"main"}}`))
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.Query(context.Background(), omnigraph.QueryInput{
		Query:  "query q($userId: String) { match { $p: Person } return { $p.name } }",
		Params: map[string]any{"userId": "u-1", "snake_param": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if seenParams["userId"] != "u-1" || seenParams["snake_param"] != true {
		t.Fatalf("params keys were altered: %+v", seenParams)
	}
	rows := res.Rows.([]any)
	row := rows[0].(map[string]any)
	if row["user_Name"] != "Alice" || row["camelKey"] != float64(42) {
		t.Fatalf("response row keys were altered: %+v", row)
	}
}

func TestErrorMappingAndContextCancellation(t *testing.T) {
	cases := []struct {
		status   int
		code     string
		sentinel error
	}{
		{http.StatusBadRequest, "bad_request", omnigraph.ErrBadRequest},
		{http.StatusUnauthorized, "unauthorized", omnigraph.ErrUnauthorized},
		{http.StatusForbidden, "forbidden", omnigraph.ErrForbidden},
		{http.StatusConflict, "conflict", omnigraph.ErrConflict},
		{http.StatusGone, "change_feed_gap", omnigraph.ErrGone},
		{http.StatusTooManyRequests, "rate_limited", omnigraph.ErrTooManyRequests},
		{http.StatusServiceUnavailable, "graph_unavailable", omnigraph.ErrGraphUnavailable},
		{http.StatusInternalServerError, "internal", omnigraph.ErrInternalServer},
	}

	for _, tc := range cases {
		srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-Id", "req-99")
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(`{"code":"` + tc.code + `","error":"boom","disposition":{"effects":"none","action":"retry"}}`))
		})
		client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Schema().Show(context.Background())
		srv.Close()

		if !errors.Is(err, tc.sentinel) {
			t.Fatalf("status %d (%s): expected %v, got %v", tc.status, tc.code, tc.sentinel, err)
		}
		var apiErr *omnigraph.Error
		if !errors.As(err, &apiErr) || apiErr.RequestID != "req-99" || apiErr.RetryAfter != "5" {
			t.Fatalf("expected RequestID and RetryAfter on Error, got %+v", err)
		}
	}

	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()
	client, _ := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Schema().Show(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
