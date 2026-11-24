package omnigraph

import (
	"context"
	"iter"
)

func (r *CommitsResource) All(ctx context.Context, in ListCommitsInput, opts ...CallOptions) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		req := in
		for {
			page, err := r.List(ctx, req, opts...)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, item := range page.Commits {
				if !yield(item, nil) {
					return
				}
			}
			if page.NextPageToken == nil || *page.NextPageToken == "" {
				return
			}
			req.PageToken = *page.NextPageToken
		}
	}
}

func (c *Client) DiffAll(ctx context.Context, in DiffInput, opts ...CallOptions) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		req := in
		for {
			page, err := c.Diff(ctx, req, opts...)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, item := range page.Changes {
				if !yield(item, nil) {
					return
				}
			}
			if page.NextPageToken == nil || *page.NextPageToken == "" {
				return
			}
			req.PageToken = *page.NextPageToken
		}
	}
}

func (r *ChangesResource) PollAll(ctx context.Context, in PollChangesInput, opts ...CallOptions) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		req := in
		for {
			page, err := r.Poll(ctx, req, opts...)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, item := range page.Blocks {
				if !yield(item, nil) {
					return
				}
			}
			if page.NextPageToken == nil || *page.NextPageToken == "" {
				return
			}
			req.PageToken = *page.NextPageToken
		}
	}
}
