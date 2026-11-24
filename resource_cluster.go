package omnigraph

import (
	"context"
	"net/http"
	"net/url"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type ClusterResource struct {
	c *Client
}

func (c *Client) Cluster() *ClusterResource {
	return &ClusterResource{c: c}
}

func (r *ClusterResource) Status(ctx context.Context, opts ...CallOptions) (*DeploymentStatusOutput, error) {
	var out DeploymentStatusOutput
	if err := r.c.doJSON(ctx, http.MethodGet, "/cluster/deployments", transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ClusterResource) Apply(ctx context.Context, in DeploymentInput, opts ...CallOptions) (*DeploymentOutput, error) {
	var out DeploymentOutput
	if err := r.c.doJSON(ctx, http.MethodPost, "/cluster/deployments", transport.RequestOptions{
		Body:    in,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ClusterResource) Lookup(ctx context.Context, deploymentID string, opts ...CallOptions) (*DeploymentOutput, error) {
	var out DeploymentOutput
	path := "/cluster/deployments/" + url.PathEscape(deploymentID)
	if err := r.c.doJSON(ctx, http.MethodGet, path, transport.RequestOptions{
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ClusterResource) Plan(ctx context.Context, in DeploymentPlanInput, opts ...CallOptions) (*DeploymentStatusOutput, error) {
	var out DeploymentStatusOutput
	if err := r.c.doJSON(ctx, http.MethodPost, "/cluster/plan", transport.RequestOptions{
		Body:    in,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
