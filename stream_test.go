package omnigraph_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	omnigraph "github.com/namanmdave/omnigraph-go"
)

func TestExportStreamAndLoadNDJSON(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/graphs/alpha/export":
			w.Header().Set("Content-Type", "application/x-ndjson")
			_, _ = w.Write([]byte("{\"type\":\"Person\",\"data\":{\"name\":\"Alice\"}}\n{\"type\":\"Person\",\"data\":{\"name\":\"Bob\"}}\n"))
		case r.Method == http.MethodPost && r.URL.Path == "/graphs/alpha/load/ndjson":
			_, _ = w.Write([]byte(`{"branch":"main","branch_created":false,"mode":"upsert","nodes":["n1"],"edges":[],"total_entities":1}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	stream, err := client.Export(context.Background(), omnigraph.ExportInput{Branch: "main"})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	var names []string
	for stream.Next() {
		names = append(names, stream.Item().Data["name"].(string))
	}
	if err := stream.Err(); err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "Alice" || names[1] != "Bob" {
		t.Fatalf("unexpected streamed names: %v", names)
	}

	loaded, err := client.LoadNDJSON(context.Background(), []byte("{\"type\":\"Person\",\"data\":{\"name\":\"Alice\"}}\n"), "main")
	if err != nil || loaded.TotalEntities != 1 {
		t.Fatalf("unexpected load ndjson output: %+v, err=%v", loaded, err)
	}
}

func TestStreamEdgeCases(t *testing.T) {
	largeVal := strings.Repeat("x", 100*1024)
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/graphs/alpha/export":
			w.Header().Set("Content-Type", "application/x-ndjson")
			_, _ = w.Write([]byte("\n{\"type\":\"Person\",\"data\":{\"bio\":\"" + largeVal + "\"}}\n\n{\"bad-json\"\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	stream, err := client.Export(context.Background(), omnigraph.ExportInput{Branch: "main"})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	if !stream.Next() {
		t.Fatalf("expected large row to succeed, err=%v", stream.Err())
	}
	if got := stream.Item().Data["bio"].(string); len(got) != len(largeVal) {
		t.Fatalf("expected %d bytes, got %d", len(largeVal), len(got))
	}

	if stream.Next() {
		t.Fatal("expected malformed line to stop stream")
	}
	if stream.Err() == nil {
		t.Fatal("expected non-nil error on malformed JSON line")
	}
}
