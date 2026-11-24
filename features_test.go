package omnigraph_test

import (
	"context"
	"net/http"
	"testing"

	omnigraph "github.com/namanmdave/omnigraph-go"
)

func TestConnectAndFromEnv(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"schema_source":"ok"}`))
	})
	defer srv.Close()

	c, err := omnigraph.Connect(omnigraph.Config{
		BaseURL: srv.URL,
		Token:   "tok",
		GraphID: "alpha",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Schema().Show(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Setenv("OMNIGRAPH_URL", srv.URL)
	t.Setenv("OMNIGRAPH_TOKEN", "tok")
	t.Setenv("OMNIGRAPH_GRAPH", "alpha")

	envClient, err := omnigraph.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := envClient.Schema().Show(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestBranchClient(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/graphs/alpha/query":
			_, _ = w.Write([]byte(`{"query_name":"q","row_count":1,"rows":[{"name":"Alice"}],"target":{"branch":"feat"}}`))
		case "/graphs/alpha/snapshot":
			if r.URL.Query().Get("branch") != "feat" {
				t.Fatalf("expected branch=feat, got %q", r.URL.Query().Get("branch"))
			}
			_, _ = w.Write([]byte(`{"graph_branch":"feat","graph_manifest_version":1,"internal_schema_version":16,"datasets":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	feat := client.Branch("feat")
	if feat.Name() != "feat" {
		t.Fatalf("expected feat, got %q", feat.Name())
	}
	if _, err := feat.Query(context.Background(), `query q() { match { $p: Person } return { $p.name } }`, nil); err != nil {
		t.Fatal(err)
	}
	snap, err := feat.Snapshot(context.Background())
	if err != nil || snap.GraphBranch != "feat" {
		t.Fatalf("unexpected snapshot: %+v, err=%v", snap, err)
	}
}

func TestPaginationIterators(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphs/alpha/commits" {
			switch r.URL.Query().Get("page_token") {
			case "":
				_, _ = w.Write([]byte(`{"commits":[{"graph_commit_id":"c1"},{"graph_commit_id":"c2"}],"next_page_token":"p2"}`))
			case "p2":
				_, _ = w.Write([]byte(`{"commits":[{"graph_commit_id":"c3"}]}`))
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	var count int
	for _, err := range client.Commits().All(context.Background(), omnigraph.ListCommitsInput{Branch: "main"}) {
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 3 {
		t.Fatalf("expected 3 commits across 2 pages, got %d", count)
	}
}

func TestInvokeAsAndExportAs(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/graphs/alpha/queries/find_person":
			_, _ = w.Write([]byte(`{"query_name":"find_person","row_count":1,"rows":[{"user_name":"Alice","age":30}],"target":{"branch":"main"}}`))
		case "/graphs/alpha/export":
			w.Header().Set("Content-Type", "application/x-ndjson")
			_, _ = w.Write([]byte("{\"type\":\"Person\",\"data\":{\"user_name\":\"Alice\",\"age\":30}}\n{\"type\":\"Person\",\"data\":{\"user_name\":\"Bob\",\"age\":25}}\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	type person struct {
		Name string `omnigraph:"user_name"`
		Age  int    `json:"age"`
	}

	people, _, err := omnigraph.InvokeAs[person](context.Background(), client, "find_person", omnigraph.InvokeQueryInput{
		Params: omnigraph.Params{"name": "Alice"},
	})
	if err != nil || len(people) != 1 || people[0].Name != "Alice" {
		t.Fatalf("unexpected InvokeAs output: %+v, err=%v", people, err)
	}

	stream, err := omnigraph.ExportAs[person](context.Background(), client, omnigraph.ExportInput{Branch: "main"})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	var exported []person
	for stream.Next() {
		exported = append(exported, stream.Item())
	}
	if err := stream.Err(); err != nil {
		t.Fatal(err)
	}
	if len(exported) != 2 || exported[1].Name != "Bob" || exported[1].Age != 25 {
		t.Fatalf("unexpected ExportAs output: %+v", exported)
	}
}
