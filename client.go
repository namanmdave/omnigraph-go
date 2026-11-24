package omnigraph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/namanmdave/omnigraph-go/internal/transport"
)

type Client struct {
	t *transport.Client
}

func NewClient(baseURL string, opts ...Option) (*Client, error) {
	cfg := clientConfig{
		contractVersion: HTTPAPIContract,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	t, err := transport.New(transport.Config{
		BaseURL:         baseURL,
		Token:           cfg.token,
		GraphID:         cfg.graphID,
		ContractVersion: cfg.contractVersion,
		HTTPClient:      cfg.httpClient,
		SkipDiscovery:   cfg.skipDiscovery,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &Client{t: t}, nil
}

func (c *Client) Graph(graphID string) *Client {
	return &Client{t: c.t.WithGraphID(graphID)}
}

func (c *Client) Health(ctx context.Context) (*HealthOutput, error) {
	var out HealthOutput
	if err := c.doJSON(ctx, http.MethodGet, "/healthz", transport.RequestOptions{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Readiness(ctx context.Context) (*ReadinessOutput, error) {
	var out ReadinessOutput
	if err := c.doJSON(ctx, http.MethodGet, "/readyz", transport.RequestOptions{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) OAuthProtectedResourceMetadata(ctx context.Context, resourcePath string) (map[string]any, error) {
	path := "/.well-known/oauth-protected-resource"
	if trimmed := strings.Trim(resourcePath, "/"); trimmed != "" {
		segments := strings.Split(trimmed, "/")
		for i, seg := range segments {
			if seg == "" || seg == "." || seg == ".." {
				return nil, &Error{
					Code:    "configuration_error",
					Message: "invalid resourcePath segment",
					Method:  http.MethodGet,
					Err:     ErrConfiguration,
				}
			}
			segments[i] = url.PathEscape(seg)
		}
		path += "/" + strings.Join(segments, "/")
	}
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodGet, path, transport.RequestOptions{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Access(ctx context.Context, in AccessInput, opts ...CallOptions) (*AccessOutput, error) {
	q := url.Values{}
	setOptional(q, "branch", in.Branch)
	setOptional(q, "target_branch", in.TargetBranch)

	var out AccessOutput
	if err := c.doJSON(ctx, http.MethodGet, "/access", transport.RequestOptions{
		Query:   q,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Snapshot(ctx context.Context, in SnapshotInput, opts ...CallOptions) (*SnapshotOutput, error) {
	q := url.Values{}
	setOptional(q, "branch", in.Branch)

	var out SnapshotOutput
	if err := c.doJSON(ctx, http.MethodGet, "/snapshot", transport.RequestOptions{
		Query:   q,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Query(ctx context.Context, in QueryInput, opts ...CallOptions) (*ReadOutput, error) {
	var out ReadOutput
	if err := c.doJSON(ctx, http.MethodPost, "/query", transport.RequestOptions{
		Body:    in,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Mutate(ctx context.Context, in MutationInput, opts ...ConditionalCallOptions) (*ChangeOutput, error) {
	path := "/mutate"
	headers := http.Header{}
	if len(opts) > 0 && opts[0].IfGraphCommit != "" {
		path = "/mutate/if-graph-commit"
		headers.Set(IfGraphCommitHeader, opts[0].IfGraphCommit)
	}

	var out ChangeOutput
	if err := c.doJSON(ctx, http.MethodPost, path, transport.RequestOptions{
		Body:    in,
		Headers: headers,
		IfActor: firstCondIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	if !validMutationReceipt(&out) {
		return nil, &Error{
			Status:            http.StatusOK,
			Code:              "api_contract_mismatch",
			Message:           "invalid mutation receipt from server; reconcile before retrying",
			Method:            http.MethodPost,
			RequestDispatched: true,
			OutcomeUnknown:    true,
			Err:               ErrAPIContract,
		}
	}
	return &out, nil
}

func (c *Client) Load(ctx context.Context, in IngestInput, opts ...CallOptions) (*IngestOutput, error) {
	var out IngestOutput
	if err := c.doJSON(ctx, http.MethodPost, "/load", transport.RequestOptions{
		Body:    in,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) LoadNDJSON(ctx context.Context, ndjson []byte, branch string, opts ...CallOptions) (*GraphBatchLoadOutput, error) {
	q := url.Values{}
	setOptional(q, "branch", branch)

	var out GraphBatchLoadOutput
	if err := c.doJSON(ctx, http.MethodPost, "/load/ndjson", transport.RequestOptions{
		Query:          q,
		RawBody:        bytes.NewReader(ndjson),
		RawContentType: "application/x-ndjson",
		IfActor:        firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Export(ctx context.Context, in ExportInput, opts ...CallOptions) (*Stream[ExportRow], error) {
	resp, err := c.t.Send(ctx, http.MethodPost, "/export", transport.RequestOptions{
		Body:    in,
		IfActor: firstIfActor(opts),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return newStream[ExportRow](resp.Body), nil
}

func (c *Client) Diff(ctx context.Context, in DiffInput, opts ...CallOptions) (*DiffOutput, error) {
	q := url.Values{}
	setOptional(q, "from", in.From)
	setOptional(q, "to", in.To)
	addSlice(q, "kind", in.Kind)
	addSlice(q, "op", in.Op)
	addSlice(q, "type", in.Type)
	setOptional(q, "page_token", in.PageToken)
	if in.Limit > 0 {
		q.Set("limit", strconv.Itoa(in.Limit))
	}

	var out DiffOutput
	if err := c.doJSON(ctx, http.MethodGet, "/diff", transport.RequestOptions{
		Query:   q,
		IfActor: firstIfActor(opts),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, opts transport.RequestOptions, out any) error {
	if err := c.t.RequestJSON(ctx, method, path, opts, out); err != nil {
		return mapError(err)
	}
	return nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var tErr *transport.TransportError
	if !errors.As(err, &tErr) {
		return err
	}

	var sentinel error
	switch tErr.Kind {
	case transport.KindConfiguration:
		sentinel = ErrConfiguration
	case transport.KindAPIContract:
		sentinel = ErrAPIContract
	case transport.KindNetwork:
		sentinel = ErrNetwork
	default:
		sentinel = sentinelForStatus(tErr.Status, tErr.Code)
	}

	var body *ErrorOutput
	if len(tErr.RawBody) > 0 {
		var parsed ErrorOutput
		if json.Unmarshal(tErr.RawBody, &parsed) == nil {
			body = &parsed
		}
	}

	var disp *DispositionOutput
	if body != nil {
		disp = body.Disposition
	}

	return &Error{
		Status:            tErr.Status,
		Code:              tErr.Code,
		Message:           tErr.Message,
		RequestID:         tErr.RequestID,
		Method:            tErr.Method,
		URL:               tErr.URL,
		RequestDispatched: tErr.RequestDispatched,
		OutcomeUnknown:    tErr.OutcomeUnknown,
		Disposition:       disp,
		RetryAfter:        tErr.RetryAfter,
		Body:              body,
		Err:               sentinel,
	}
}

func setOptional(q url.Values, key, val string) {
	if val != "" {
		q.Set(key, val)
	}
}

func addSlice(q url.Values, key string, vals []string) {
	for _, v := range vals {
		q.Add(key, v)
	}
}

func firstIfActor(opts []CallOptions) string {
	if len(opts) > 0 {
		return opts[0].IfActor
	}
	return ""
}

func firstCondIfActor(opts []ConditionalCallOptions) string {
	if len(opts) > 0 {
		return opts[0].IfActor
	}
	return ""
}
