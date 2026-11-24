package omnigraph

import (
	"context"
	"net/http"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type GraphsResource struct {
	c *Client
}

func (c *Client) Graphs() *GraphsResource {
	return &GraphsResource{c: c}
}

func (r *GraphsResource) List(ctx context.Context, opts ...CallOptions) (*GraphListOutput, error) {
	var out GraphListOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/graphs", transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *GraphsResource) Discover(ctx context.Context, opts ...CallOptions) (*GraphDiscoveryOutput, error) {
	var out GraphDiscoveryOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/graphs/discovery", transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
