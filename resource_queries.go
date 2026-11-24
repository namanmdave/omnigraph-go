package omnigraph

import (
	"context"
	"net/http"
	"net/url"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type QueriesResource struct {
	c *Client
}

func (c *Client) Queries() *QueriesResource {
	return &QueriesResource{c: c}
}

func (r *QueriesResource) List(ctx context.Context, opts ...CallOptions) (*QueriesCatalogOutput, error) {
	var out QueriesCatalogOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/queries", transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *QueriesResource) Invoke(ctx context.Context, name string, in InvokeQueryInput, opts ...ConditionalCallOptions) (InvokeQueryOutput, error) {
	path := "/queries/" + url.PathEscape(name)
	headers := http.Header{}
	if len(opts) > 0 && opts[0].IfGraphCommit != "" {
		path += "/if-graph-commit"
		headers.Set(IfGraphCommitHeader, opts[0].IfGraphCommit)
	}

	var out InvokeQueryOutput
	if err := r.c.doJSON(ctx, http.MethodPost, path, transport.RequestOptions{
		Body:    in,
		Headers: headers,
		IfActor: firstCondIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
