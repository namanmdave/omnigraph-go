package omnigraph

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type ChangesResource struct {
	c *Client
}

func (c *Client) Changes() *ChangesResource {
	return &ChangesResource{c: c}
}

func (r *ChangesResource) Poll(ctx context.Context, in PollChangesInput, opts ...CallOptions) (*ChangeFeedOutput, error) {
	q := url.Values{}
	setOptional(q, "cursor", in.Cursor)
	setOptional(q, "branch", in.Branch)
	addSlice(q, "kind", in.Kind)
	addSlice(q, "op", in.Op)
	addSlice(q, "type", in.Type)
	setOptional(q, "page_token", in.PageToken)
	if in.Limit > 0 {
		q.Set("limit", strconv.Itoa(in.Limit))
	}

	var out ChangeFeedOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/changes", transport.RequestOptions{
		Query:   q,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ChangesResource) Baseline(ctx context.Context, in ChangeBaselineInput, opts ...CallOptions) (*ChangeBaselineOutput, error) {
	var out ChangeBaselineOutput
	if err := r.c.doJSON(ctx, http.MethodPost, "/changes/baseline", transport.RequestOptions{
		Body:    in,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
