package omnigraph

import (
	"context"
	"io"
	"net/http"
)

type BranchClient struct {
	c      *Client
	branch string
}

func (c *Client) Branch(name string) *BranchClient {
	return &BranchClient{c: c, branch: name}
}

func (b *BranchClient) Name() string {
	return b.branch
}

func (b *BranchClient) Query(ctx context.Context, query string, params Params, opts ...CallOptions) (*ReadOutput, error) {
	return b.c.Query(ctx, QueryInput{
		Query:  query,
		Params: params,
		Branch: b.branch,
	}, opts...)
}

func (b *BranchClient) Mutate(ctx context.Context, query string, params Params, opts ...ConditionalCallOptions) (*ChangeOutput, error) {
	return b.c.Mutate(ctx, MutationInput{
		Query:  query,
		Params: params,
		Branch: b.branch,
	}, opts...)
}

func (b *BranchClient) InvokeQuery(ctx context.Context, name string, params Params, opts ...ConditionalCallOptions) (InvokeQueryOutput, error) {
	return b.c.Queries().Invoke(ctx, name, InvokeQueryInput{
		Params: params,
		Branch: b.branch,
	}, opts...)
}

func (b *BranchClient) Snapshot(ctx context.Context, opts ...CallOptions) (*SnapshotOutput, error) {
	return b.c.Snapshot(ctx, SnapshotInput{Branch: b.branch}, opts...)
}

func (b *BranchClient) Export(ctx context.Context, typeNames []string, opts ...CallOptions) (*Stream[ExportRow], error) {
	return b.c.Export(ctx, ExportInput{
		Branch:    b.branch,
		TypeNames: typeNames,
	}, opts...)
}

func (b *BranchClient) Load(ctx context.Context, data string, from string, mode any, opts ...CallOptions) (*IngestOutput, error) {
	return b.c.Load(ctx, IngestInput{
		Data:   data,
		Branch: b.branch,
		From:   from,
		Mode:   mode,
	}, opts...)
}

func (b *BranchClient) LoadNDJSON(ctx context.Context, ndjson []byte, opts ...CallOptions) (*GraphBatchLoadOutput, error) {
	return b.c.LoadNDJSON(ctx, ndjson, b.branch, opts...)
}

func (b *BranchClient) GetBlob(ctx context.Context, entityType, key, field string, opts ...CallOptions) (*http.Response, error) {
	return b.c.Blobs().Get(ctx, BlobInput{
		Type:   entityType,
		Key:    key,
		Field:  field,
		Branch: b.branch,
	}, opts...)
}

func (b *BranchClient) PutBlob(ctx context.Context, entityType, key, field string, body io.Reader, contentType string, opts ...CallOptions) (*BlobWriteOutput, error) {
	return b.c.Blobs().Put(ctx, BlobWriteInput{
		Type:   entityType,
		Key:    key,
		Field:  field,
		Branch: b.branch,
	}, body, contentType, opts...)
}
