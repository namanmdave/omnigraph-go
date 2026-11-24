package omnigraph_test

import (
	"context"
	"net/http"
	"testing"

	omnigraph "github.com/namanmdave/omnigraph-go"
)

func TestQueryAsAndStructTags(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"query_name":"find",
			"row_count":2,
			"rows":[
				{"user_name":"Alice","age":30,"ignored":"x"},
				{"user_name":"Bob","age":25,"ignored":"y"}
			],
			"target":{"branch":"main"}
		}`))
	})
	defer srv.Close()

	client, err := omnigraph.NewClient(srv.URL, omnigraph.WithGraphID("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	type person struct {
		Name    string `omnigraph:"user_name"`
		Age     int    `json:"age"`
		Skipped string `omnigraph:"-"`
	}

	people, res, err := omnigraph.QueryAs[person](context.Background(), client, omnigraph.QueryInput{
		Query:  `query find() { match { $p: Person } return { $p.name as user_name, $p.age as age } }`,
		Branch: "main",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 2 || len(people) != 2 {
		t.Fatalf("expected 2 people, got len=%d rowCount=%d", len(people), res.RowCount)
	}
	if people[0].Name != "Alice" || people[0].Age != 30 || people[0].Skipped != "" {
		t.Fatalf("unexpected decoded person[0]: %+v", people[0])
	}
	if people[1].Name != "Bob" || people[1].Age != 25 {
		t.Fatalf("unexpected decoded person[1]: %+v", people[1])
	}
}
