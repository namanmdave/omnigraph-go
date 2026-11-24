package omnigraph

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type CommitsResource struct {
	c *Client
}

func (c *Client) Commits() *CommitsResource {
	return &CommitsResource{c: c}
}

func (r *CommitsResource) List(ctx context.Context, in ListCommitsInput, opts ...CallOptions) (*CommitListOutput, error) {
	q := url.Values{}
	setOptional(q, "branch", in.Branch)
	setOptional(q, "page_token", in.PageToken)
	if in.Limit > 0 {
		q.Set("limit", strconv.Itoa(in.Limit))
	}

	var out CommitListOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/commits", transport.RequestOptions{
		Query:   q,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CommitsResource) Show(ctx context.Context, commitID string, opts ...CallOptions) (*CommitOutput, error) {
	var out CommitOutput
	path := "/commits/" + url.PathEscape(commitID)
	if err := r.c.doJSON(ctx, http.MethodGet, path, transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CommitsResource) Drop(ctx context.Context, commitID string, opts ...CallOptions) error {
	path := "/commits/" + url.PathEscape(commitID)
	return r.c.doJSON(ctx, http.MethodDelete, path, transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, nil)
}

func (r *CommitsResource) Changes(ctx context.Context, commitID string, pageToken string, limit int, opts ...CallOptions) (*CommitChangesOutput, error) {
	q := url.Values{}
	setOptional(q, "page_token", pageToken)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	var out CommitChangesOutput
	path := "/commits/" + url.PathEscape(commitID) + "/changes"
	if err := r.c.doJSON(ctx, http.MethodGet, path, transport.RequestOptions{
		Query:   q,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CommitsResource) Publish(ctx context.Context, commitID string, opts ...CallOptions) (*CommitOutput, error) {
	var out CommitOutput
	path := "/commits/" + url.PathEscape(commitID) + "/publish"
	if err := r.c.doJSON(ctx, http.MethodPost, path, transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
