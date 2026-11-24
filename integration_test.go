//go:build integration

package omnigraph_test

import (
	"context"
	"os"
	"testing"

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

	graphs, err := client.Graphs().List(ctx)
	if err != nil {
		t.Fatalf("list graphs failed: %v", err)
	}
	if len(graphs.Graphs) == 0 {
		t.Skip("no graphs configured on live server")
	}

	graphClient := client.Graph(graphs.Graphs[0].GraphID)
	if _, err := graphClient.Schema().Show(ctx); err != nil {
		t.Fatalf("schema show failed: %v", err)
	}
}
