# cmd/sync-spec

CLI tool to sync `spec/openapi.json` with a pinned `omnigraph-server` release tag.

## What goes here
- Fetching and validating `openapi.json` from `ModernRelay/omnigraph` for `omnigraph.ServerVersion`

## What does NOT go here
- Runtime SDK client logic or HTTP transport
