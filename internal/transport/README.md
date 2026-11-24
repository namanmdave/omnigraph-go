# internal/transport

Low-level HTTP transport for the Omnigraph SDK.

## What goes here
- `baseURL` validation and `/graphs/{graph_id}` path rewriting
- Pre-flight `HEAD /healthz` contract discovery
- `Omnigraph-Http-Api` request/response header enforcement
- Public-path credential stripping (`Authorization`, `Omnigraph-If-Actor`)
- Redirect suppression and `OutcomeUnknown` write-uncertainty tracking

## What does NOT go here
- Domain resource methods (`Branches`, `Queries`, `Blobs`, etc.)
- Public Go DTO structs or user-facing `Client` options
