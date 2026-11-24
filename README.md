# omnigraph-go

Go SDK for the [Omnigraph](https://github.com/ModernRelay/omnigraph) HTTP API (`Omnigraph-Http-Api: 0.14`).

## Install

```sh
go get github.com/namanmdave/omnigraph-go
```

## Quick Start

```go
client, err := omnigraph.NewClient(
	"http://127.0.0.1:8080",
	omnigraph.WithToken(os.Getenv("OMNIGRAPH_TOKEN")),
	omnigraph.WithGraphID("alpha"),
)
if err != nil {
	log.Fatal(err)
}

res, err := client.Query(ctx, omnigraph.QueryInput{
	Query:  "query find { match { $p: Person } return { $p.name } }",
	Branch: "main",
})
```

## Package Boundaries

### What goes here (`package omnigraph`)
- Public `Client`, functional `Option`s, and `client.Graph(id)` scoping
- Public request/response DTOs (`types.go`) and typed API errors (`errors.go`)
- Resource methods (`graphs`, `branches`, `commits`, `changes`, `schema`, `queries`, `blobs`, `cluster`)
- NDJSON streaming helpers (`stream.go`)
- Unit tests (`*_test.go`), spec coverage test (`spec_coverage_test.go`), and live integration tests (`integration_test.go`)

### What does NOT go here
- Raw HTTP transport mechanics, `HEAD /healthz` discovery, or header/redirect enforcement (lives in `internal/transport`)
- OpenAPI specification files (lives in `spec`)

## Server-Version Pin & Versioning Policy

This SDK targets `omnigraph-server` **v0.14.0** (`Omnigraph-Http-Api: 0.14` in `version.go`) and keeps `spec/openapi.json` pinned to that release.

- Major/minor versions match `omnigraph-server` (`0.14.*`); patch versions are independent.
- `internal/transport` enforces `Omnigraph-Http-Api: 0.14` via `HEAD /healthz` pre-flight discovery and on every response.

## Local Development

```sh
go vet ./...
go test -race ./...

# Live server integration tests:
OMNIGRAPH_TEST_URL=http://127.0.0.1:8080 go test -tags=integration -v ./...
```

## Collaboration & Contributing

Following the [ModernRelay/omnigraph](https://github.com/ModernRelay/omnigraph) collaboration workflow:

- **Issues First:** Open an issue or discussion before sending large code changes. Use Issues for reproducible bugs and tracked feature requests, and Discussions for general ideas or design questions.
- **Pull Requests:** Every pull request implementing a change or fix should link an accepted issue (unless it is a trivial fix), keep diffs focused, and keep `spec/openapi.json` + `TestOpenAPISpecCoverage` in sync.
- **Pre-merge Checks:** Ensure `go vet ./...` and `go test -race ./...` pass before opening a PR.

## License

MIT — see [LICENSE](LICENSE).
