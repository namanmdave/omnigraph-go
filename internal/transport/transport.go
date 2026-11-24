package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	ContractHeader = "Omnigraph-Http-Api"
	IfActorHeader  = "Omnigraph-If-Actor"
)

type ErrorKind int

const (
	KindHTTP ErrorKind = iota
	KindConfiguration
	KindAPIContract
	KindNetwork
)

type TransportError struct {
	Kind              ErrorKind
	Status            int
	Code              string
	Message           string
	RequestID         string
	Method            string
	URL               string
	RequestDispatched bool
	OutcomeUnknown    bool
	RetryAfter        string
	RawBody           []byte
	Cause             error
}

func (e *TransportError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s %s (%d %s): %s", e.Method, e.URL, e.Status, e.Code, e.Message)
	}
	if e.Status > 0 {
		return fmt.Sprintf("%s %s (%d): %s", e.Method, e.URL, e.Status, e.Message)
	}
	return fmt.Sprintf("%s %s: %s", e.Method, e.URL, e.Message)
}

func (e *TransportError) Unwrap() error {
	return e.Cause
}

type Config struct {
	BaseURL         string
	Token           string
	GraphID         string
	ContractVersion string
	HTTPClient      *http.Client
	SkipDiscovery   bool
}

type RequestOptions struct {
	Headers          http.Header
	IfActor          string
	Query            url.Values
	Body             any
	RawBody          io.Reader
	RawContentType   string
	AcceptedStatuses []int
}

type Client struct {
	baseURL         string
	token           string
	graphID         string
	contractVersion string
	httpClient      *http.Client
	skipDiscovery   bool

	discoverMu   sync.Mutex
	discoveredOK bool
}

func New(cfg Config) (*Client, error) {
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil || !parsed.IsAbs() {
		return nil, &TransportError{
			Kind:    KindConfiguration,
			Method:  http.MethodHead,
			URL:     "<invalid-server-root>",
			Message: "baseURL must be an absolute HTTP(S) server root",
		}
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return nil, &TransportError{
			Kind:    KindConfiguration,
			Method:  http.MethodHead,
			URL:     parsed.Scheme + "://" + parsed.Host,
			Message: "baseURL must be an HTTP(S) server root without credentials, query, or fragment",
		}
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	contract := cfg.ContractVersion
	if contract == "" {
		contract = "0.14"
	}

	return &Client{
		baseURL:         strings.TrimRight(parsed.String(), "/"),
		token:           cfg.Token,
		graphID:         cfg.GraphID,
		contractVersion: contract,
		httpClient:      httpClient,
		skipDiscovery:   cfg.SkipDiscovery,
	}, nil
}

func (c *Client) WithGraphID(graphID string) *Client {
	return &Client{
		baseURL:         c.baseURL,
		token:           c.token,
		graphID:         graphID,
		contractVersion: c.contractVersion,
		httpClient:      c.httpClient,
		skipDiscovery:   c.skipDiscovery,
	}
}

func (c *Client) GraphID() string {
	return c.graphID
}

func (c *Client) RequestJSON(ctx context.Context, method, path string, opts RequestOptions, out any) error {
	resp, err := c.Send(ctx, method, path, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil || len(bytes.TrimSpace(raw)) == 0 {
		return &TransportError{
			Kind:              KindNetwork,
			Status:            resp.StatusCode,
			Method:            method,
			URL:               resp.Request.URL.String(),
			Message:           "response body was interrupted or empty; do not automatically repeat writes",
			RequestDispatched: true,
			OutcomeUnknown:    canWrite(method, path),
			Cause:             err,
		}
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return &TransportError{
			Kind:              KindNetwork,
			Status:            resp.StatusCode,
			Method:            method,
			URL:               resp.Request.URL.String(),
			Message:           "response body was invalid JSON; do not automatically repeat writes",
			RequestDispatched: true,
			OutcomeUnknown:    canWrite(method, path),
			Cause:             err,
		}
	}

	return nil
}

func (c *Client) Send(ctx context.Context, method, path string, opts RequestOptions) (*http.Response, error) {
	if c.graphID == "" && !isRootPath(path) {
		return nil, &TransportError{
			Kind:    KindConfiguration,
			Method:  method,
			URL:     c.baseURL + path,
			Message: fmt.Sprintf("graphID is required for graph-scoped operation %s %s", method, path),
		}
	}
	if opts.Body != nil && opts.RawBody != nil {
		return nil, &TransportError{
			Kind:    KindConfiguration,
			Method:  method,
			URL:     c.baseURL + path,
			Message: "Body and RawBody are mutually exclusive",
		}
	}

	if !c.skipDiscovery && !isPublicPath(path) {
		if err := c.discover(ctx); err != nil {
			return nil, err
		}
	}

	fullURL, err := c.buildURL(path, opts.Query)
	if err != nil {
		return nil, &TransportError{
			Kind:    KindConfiguration,
			Method:  method,
			URL:     c.baseURL + path,
			Message: err.Error(),
			Cause:   err,
		}
	}

	var bodyReader io.Reader
	var contentType string
	if opts.RawBody != nil {
		bodyReader = opts.RawBody
		contentType = opts.RawContentType
	} else if opts.Body != nil {
		encoded, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, &TransportError{
				Kind:    KindConfiguration,
				Method:  method,
				URL:     fullURL,
				Message: "failed to marshal request body",
				Cause:   err,
			}
		}
		bodyReader = bytes.NewReader(encoded)
		contentType = "application/json"
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, &TransportError{
			Kind:    KindConfiguration,
			Method:  method,
			URL:     fullURL,
			Message: err.Error(),
			Cause:   err,
		}
	}

	for k, vals := range opts.Headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set(ContractHeader, c.contractVersion)
	if opts.IfActor != "" && !isPublicPath(path) {
		req.Header.Set(IfActorHeader, opts.IfActor)
	}
	if c.token != "" && !isPublicPath(path) {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json, application/x-ndjson")
	}
	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if !canWrite(method, path) && errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, &TransportError{
			Kind:              KindNetwork,
			Method:            method,
			URL:               fullURL,
			Message:           "HTTP transport failed after dispatch; reconcile writes before retrying",
			RequestDispatched: true,
			OutcomeUnknown:    canWrite(method, path),
			Cause:             err,
		}
	}

	if !strings.HasPrefix(path, "/.well-known/oauth-protected-resource") {
		got := resp.Header.Values(ContractHeader)
		if len(got) != 1 || got[0] != c.contractVersion {
			_ = resp.Body.Close()
			return nil, &TransportError{
				Kind:              KindAPIContract,
				Status:            resp.StatusCode,
				Code:              "api_contract_mismatch",
				Method:            method,
				URL:               fullURL,
				Message:           fmt.Sprintf("response must carry exactly one %s: %s", ContractHeader, c.contractVersion),
				RequestDispatched: true,
				OutcomeUnknown:    canWrite(method, path),
			}
		}
	}

	if (resp.StatusCode < 200 || resp.StatusCode >= 300) && !slices.Contains(opts.AcceptedStatuses, resp.StatusCode) {
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, parseHTTPError(method, fullURL, resp, raw, canWrite(method, path))
	}

	return resp, nil
}

func (c *Client) discover(ctx context.Context) error {
	c.discoverMu.Lock()
	if c.discoveredOK {
		c.discoverMu.Unlock()
		return nil
	}
	c.discoverMu.Unlock()

	discoverCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	savedSkip := c.skipDiscovery
	c.discoverMu.Lock()
	c.skipDiscovery = true
	c.discoverMu.Unlock()

	resp, err := c.Send(discoverCtx, http.MethodHead, "/healthz", RequestOptions{})

	c.discoverMu.Lock()
	c.skipDiscovery = savedSkip
	if err == nil {
		c.discoveredOK = true
		_ = resp.Body.Close()
		c.discoverMu.Unlock()
		return nil
	}
	c.discoverMu.Unlock()

	var tErr *TransportError
	status := 0
	if errors.As(err, &tErr) {
		status = tErr.Status
	}
	return &TransportError{
		Kind:              KindAPIContract,
		Status:            status,
		Code:              "api_contract_mismatch",
		Method:            http.MethodHead,
		URL:               c.baseURL + "/healthz",
		Message:           fmt.Sprintf("server discovery failed; expected %s: %s; data request was not sent", ContractHeader, c.contractVersion),
		RequestDispatched: false,
		OutcomeUnknown:    false,
		Cause:             err,
	}
}

func (c *Client) buildURL(path string, query url.Values) (string, error) {
	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("transport path must start with '/': got %q", path)
	}
	resolved := path
	if c.graphID != "" && !isRootPath(path) {
		resolved = "/graphs/" + url.PathEscape(c.graphID) + path
	}
	u, err := url.Parse(c.baseURL + resolved)
	if err != nil {
		return "", err
	}
	if len(query) > 0 {
		q := u.Query()
		for k, vals := range query {
			for _, v := range vals {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func isPublicPath(path string) bool {
	return path == "/healthz" ||
		path == "/readyz" ||
		path == "/.well-known/oauth-protected-resource" ||
		strings.HasPrefix(path, "/.well-known/oauth-protected-resource/")
}

func isRootPath(path string) bool {
	return isPublicPath(path) ||
		path == "/graphs" ||
		path == "/graphs/discovery" ||
		strings.HasPrefix(path, "/cluster/")
}

func canWrite(method, path string) bool {
	if method == http.MethodGet || method == http.MethodHead {
		return false
	}
	switch path {
	case "/query", "/export", "/changes/baseline", "/cluster/plan":
		return false
	default:
		return true
	}
}

func parseHTTPError(method, fullURL string, resp *http.Response, raw []byte, writeOp bool) *TransportError {
	var parsed struct {
		Error       string `json:"error"`
		Code        string `json:"code"`
		Disposition *struct {
			Effects string `json:"effects"`
			Action  string `json:"action"`
		} `json:"disposition"`
	}
	message := http.StatusText(resp.StatusCode)
	code := ""
	outcomeUnknown := writeOp
	if len(raw) > 0 && json.Unmarshal(raw, &parsed) == nil {
		if parsed.Error != "" {
			message = parsed.Error
		}
		code = parsed.Code
		if parsed.Disposition != nil && parsed.Disposition.Effects == "none" {
			outcomeUnknown = false
		}
	}

	return &TransportError{
		Kind:              KindHTTP,
		Status:            resp.StatusCode,
		Code:              code,
		Message:           message,
		RequestID:         resp.Header.Get("X-Request-Id"),
		Method:            method,
		URL:               fullURL,
		RequestDispatched: true,
		OutcomeUnknown:    outcomeUnknown,
		RetryAfter:        resp.Header.Get("Retry-After"),
		RawBody:           raw,
	}
}
