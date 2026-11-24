package omnigraph

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type BlobsResource struct {
	c *Client
}

func (c *Client) Blobs() *BlobsResource {
	return &BlobsResource{c: c}
}

func (r *BlobsResource) Get(ctx context.Context, in BlobInput, opts ...CallOptions) (*http.Response, error) {
	q := blobQuery(in.Type, in.Key, in.Field, in.Branch)
	headers := http.Header{}
	if in.Range != "" {
		headers.Set("Range", in.Range)
	}
	resp, err := r.c.t.Send(ctx, http.MethodGet, "/blob", transport.RequestOptions{
		Query:            q,
		Headers:          headers,
		IfActor:          firstIfActor(opts),
		AcceptedStatuses: []int{http.StatusFound, http.StatusNotModified},
	})
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

func (r *BlobsResource) Stat(ctx context.Context, in BlobInput, opts ...CallOptions) (*http.Response, error) {
	q := blobQuery(in.Type, in.Key, in.Field, in.Branch)
	resp, err := r.c.t.Send(ctx, http.MethodHead, "/blob", transport.RequestOptions{
		Query:            q,
		IfActor:          firstIfActor(opts),
		AcceptedStatuses: []int{http.StatusFound, http.StatusNotModified},
	})
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

func (r *BlobsResource) Put(ctx context.Context, in BlobWriteInput, body io.Reader, contentType string, opts ...CallOptions) (*BlobWriteOutput, error) {
	q := blobQuery(in.Type, in.Key, in.Field, in.Branch)
	headers := http.Header{}
	if in.IfMatch != "" {
		headers.Set("If-Match", in.IfMatch)
	}
	if in.IfNoneMatch != "" {
		headers.Set("If-None-Match", in.IfNoneMatch)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var out BlobWriteOutput
	if err := r.c.doJSON(ctx, http.MethodPut, "/blob", transport.RequestOptions{
		Query:          q,
		Headers:        headers,
		RawBody:        body,
		RawContentType: contentType,
		IfActor:        firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BlobsResource) Clear(ctx context.Context, in BlobWriteInput, opts ...CallOptions) (*BlobWriteOutput, error) {
	q := blobQuery(in.Type, in.Key, in.Field, in.Branch)
	headers := http.Header{}
	if in.IfMatch != "" {
		headers.Set("If-Match", in.IfMatch)
	}

	var out BlobWriteOutput
	if err := r.c.doJSON(ctx, http.MethodDelete, "/blob", transport.RequestOptions{
		Query:   q,
		Headers: headers,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func blobQuery(entityType, key, field, branch string) url.Values {
	q := url.Values{}
	setOptional(q, "type", entityType)
	setOptional(q, "key", key)
	setOptional(q, "field", field)
	setOptional(q, "branch", branch)
	return q
}
