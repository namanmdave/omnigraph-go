//go:build integration

package omnigraph_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	omnigraph "github.com/namanmdave/omnigraph-go"
)

func TestLiveServerIntegration(t *testing.T) {
	baseURL := os.Getenv("OMNIGRAPH_TEST_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}

	var opts []omnigraph.Option
	if tok := os.Getenv("OMNIGRAPH_TOKEN"); tok != "" {
		opts = append(opts, omnigraph.WithToken(tok))
	}

	client, err := omnigraph.NewClient(baseURL, opts...)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	health, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if health.Status == "" {
		t.Fatalf("expected non-empty health status")
	}

	ready, err := client.Readiness(ctx)
	if err != nil {
		t.Fatalf("readiness check failed: %v", err)
	}
	if !ready.Ready {
		t.Fatalf("expected server ready=true, got %+v", ready)
	}

	graphs, err := client.Graphs().List(ctx)
	if err != nil {
		t.Fatalf("list graphs failed: %v", err)
	}
	if len(graphs.Graphs) == 0 {
		t.Skip("no graphs configured on live server")
	}

	g := client.Graph(graphs.Graphs[0].GraphID)

	schema, err := g.Schema().Show(ctx)
	if err != nil {
		t.Fatalf("schema show failed: %v", err)
	}
	if schema.SchemaSource == "" {
		t.Fatalf("expected non-empty schema source")
	}

	snap, err := g.Snapshot(ctx, omnigraph.SnapshotInput{Branch: "main"})
	if err != nil {
		t.Fatalf("snapshot failed: %v", err)
	}
	if snap.GraphBranch != "main" {
		t.Fatalf("expected branch main, got %q", snap.GraphBranch)
	}

	branchName := fmt.Sprintf("go-it-%d", time.Now().UnixNano())
	br, err := g.Branches().Create(ctx, omnigraph.BranchCreateInput{Name: branchName, From: "main"})
	if err != nil {
		t.Fatalf("branch create failed: %v", err)
	}
	if br.Name != branchName {
		t.Fatalf("expected branch %q, got %q", branchName, br.Name)
	}
	defer func() {
		_, _ = g.Branches().Delete(ctx, branchName)
	}()

	mut, err := g.Mutate(ctx, omnigraph.MutationInput{
		Query:  `query add_it() { insert Person { name: "GoSDKUser", age: 30 } }`,
		Branch: branchName,
	})
	if err != nil {
		t.Fatalf("mutate failed: %v", err)
	}
	if mut.AffectedNodes < 1 {
		t.Fatalf("expected affected_nodes >= 1, got %d", mut.AffectedNodes)
	}

	read, err := g.Query(ctx, omnigraph.QueryInput{
		Query:  `query find_it() { match { $p: Person { name: "GoSDKUser" } } return { $p.name as name, $p.age as age } }`,
		Branch: branchName,
	})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if read.RowCount != 1 {
		t.Fatalf("expected 1 row, got %d", read.RowCount)
	}

	stream, err := g.Export(ctx, omnigraph.ExportInput{Branch: branchName})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	defer stream.Close()

	var exported int
	for stream.Next() {
		exported++
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("export stream error: %v", err)
	}
	if exported == 0 {
		t.Fatalf("expected at least 1 exported row")
	}
}
