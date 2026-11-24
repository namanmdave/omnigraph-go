package omnigraph

import (
	"context"
	"net/http"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type SchemaResource struct {
	c *Client
}

func (c *Client) Schema() *SchemaResource {
	return &SchemaResource{c: c}
}

func (r *SchemaResource) Show(ctx context.Context, opts ...CallOptions) (*SchemaOutput, error) {
	var out SchemaOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/schema", transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
