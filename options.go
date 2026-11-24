package omnigraph

import (
	"net/http"
	"os"
)

type Option func(*clientConfig)

type Config struct {
	BaseURL         string
	Token           string
	GraphID         string
	HTTPClient      *http.Client
	ContractVersion string
	SkipDiscovery   bool
}

type clientConfig struct {
	token           string
	graphID         string
	httpClient      *http.Client
	contractVersion string
	skipDiscovery   bool
}

func Connect(cfg Config) (*Client, error) {
	var opts []Option
	if cfg.Token != "" {
		opts = append(opts, WithToken(cfg.Token))
	}
	if cfg.GraphID != "" {
		opts = append(opts, WithGraphID(cfg.GraphID))
	}
	if cfg.HTTPClient != nil {
		opts = append(opts, WithHTTPClient(cfg.HTTPClient))
	}
	if cfg.ContractVersion != "" {
		opts = append(opts, WithHTTPAPIContract(cfg.ContractVersion))
	}
	if cfg.SkipDiscovery {
		opts = append(opts, WithSkipDiscovery())
	}
	return NewClient(cfg.BaseURL, opts...)
}

func FromEnv() (*Client, error) {
	return Connect(Config{
		BaseURL: os.Getenv("OMNIGRAPH_URL"),
		Token:   os.Getenv("OMNIGRAPH_TOKEN"),
		GraphID: os.Getenv("OMNIGRAPH_GRAPH"),
	})
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
