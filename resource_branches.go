package omnigraph

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type BranchesResource struct {
	c *Client
}

func (c *Client) Branches() *BranchesResource {
	return &BranchesResource{c: c}
}

func (r *BranchesResource) List(ctx context.Context, opts ...CallOptions) (*BranchListOutput, error) {
	var out BranchListOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/branches", transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BranchesResource) Create(ctx context.Context, in BranchCreateInput, opts ...CallOptions) (*BranchCreateOutput, error) {
	var out BranchCreateOutput
	if err := r.c.doJSON(ctx, http.MethodPost, "/branches", transport.RequestOptions{
		Body:    in,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BranchesResource) Delete(ctx context.Context, name string, opts ...CallOptions) (*BranchDeleteOutput, error) {
	var out BranchDeleteOutput
	path := "/branches/" + url.PathEscape(name)
	if err := r.c.doJSON(ctx, http.MethodDelete, path, transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BranchesResource) Merge(ctx context.Context, in BranchMergeInput, opts ...CallOptions) (*BranchMergeOutput, error) {
	var out BranchMergeOutput
	if err := r.c.doJSON(ctx, http.MethodPost, "/branches/merge", transport.RequestOptions{
		Body:    in,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BranchesResource) Diff(ctx context.Context, in BranchDiffInput, opts ...CallOptions) (*DiffOutput, error) {
	q := url.Values{}
	setOptional(q, "source", in.Source)
	setOptional(q, "target", in.Target)
	addSlice(q, "kind", in.Kind)
	addSlice(q, "op", in.Op)
	addSlice(q, "type", in.Type)
	setOptional(q, "page_token", in.PageToken)
	if in.Limit > 0 {
		q.Set("limit", strconv.Itoa(in.Limit))
	}

	var out DiffOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/branches/diff", transport.RequestOptions{
		Query:   q,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BranchesResource) MergeBase(ctx context.Context, in MergeBaseInput, opts ...CallOptions) (*MergeBaseOutput, error) {
	q := url.Values{}
	setOptional(q, "source", in.Source)
	setOptional(q, "target", in.Target)

	var out MergeBaseOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/branches/merge-base", transport.RequestOptions{
		Query:   q,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
