package omnigraph

import "net/http"

type Option func(*clientConfig)

type clientConfig struct {
	token           string
	graphID         string
	httpClient      *http.Client
	contractVersion string
	skipDiscovery   bool
}

func WithToken(token string) Option {
	return func(c *clientConfig) {
		c.token = token
	}
}

func WithGraphID(graphID string) Option {
	return func(c *clientConfig) {
		c.graphID = graphID
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = httpClient
	}
}

func WithHTTPAPIContract(version string) Option {
	return func(c *clientConfig) {
		c.contractVersion = version
	}
}

func WithSkipDiscovery() Option {
	return func(c *clientConfig) {
		c.skipDiscovery = true
	}
}

type CallOptions struct {
	IfActor string
}

type ConditionalCallOptions struct {
	IfActor       string
	IfGraphCommit string
}
