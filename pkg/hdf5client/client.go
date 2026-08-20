// Package hdf5client is an HTTP client for the HDF5 Agent v1 API.
//
// Downstream data-platform services should depend on this package (or call the
// documented JSON API directly) instead of linking libhdf5.
package hdf5client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tacurran/hdf5-agent/internal/apierror"
	"github.com/tacurran/hdf5-agent/internal/hdf5store"
)

// Client talks to a running HDF5 Agent over HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// New constructs a Client for baseURL, for example http://hdf5-agent:8080.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Health calls GET /api/v1/health.
func (c *Client) Health(ctx context.Context) (*hdf5store.HealthResponse, error) {
	var out hdf5store.HealthResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/health", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Ready calls GET /api/v1/ready.
func (c *Client) Ready(ctx context.Context) (*hdf5store.ReadyResponse, error) {
	var out hdf5store.ReadyResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/ready", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListFiles calls GET /api/v1/files.
func (c *Client) ListFiles(ctx context.Context) ([]hdf5store.FileInfo, error) {
	var out hdf5store.FilesResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/files", nil, &out); err != nil {
		return nil, err
	}
	return out.Files, nil
}

// Structure calls GET /api/v1/files/{name}.
func (c *Client) Structure(ctx context.Context, name string) (*hdf5store.Node, error) {
	var out hdf5store.Node
	if err := c.do(ctx, http.MethodGet, "/api/v1/files/"+url.PathEscape(name), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReadDataset calls GET /api/v1/files/{name}/datasets?path=.
func (c *Client) ReadDataset(ctx context.Context, name, datasetPath string) (*hdf5store.Dataset, error) {
	q := url.Values{"path": {datasetPath}}
	var out hdf5store.Dataset
	path := "/api/v1/files/" + url.PathEscape(name) + "/datasets?" + q.Encode()
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateFile calls POST /api/v1/files.
func (c *Client) CreateFile(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodPost, "/api/v1/files", hdf5store.CreateFileRequest{Name: name}, nil)
}

// DeleteFile calls DELETE /api/v1/files/{name}.
func (c *Client) DeleteFile(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/files/"+url.PathEscape(name), nil, nil)
}

// UpdateDataset calls PUT /api/v1/files/{name}/datasets.
func (c *Client) UpdateDataset(ctx context.Context, name, datasetPath string, indices []int, values []any) error {
	body := hdf5store.UpdateRequest{Path: datasetPath, Indices: indices, Values: values}
	return c.do(ctx, http.MethodPut, "/api/v1/files/"+url.PathEscape(name)+"/datasets", body, nil)
}

// APIError is a structured error returned by the service.
type APIError struct {
	StatusCode int
	Body       apierror.Body
}

func (e *APIError) Error() string {
	return fmt.Sprintf("hdf5-agent: %s (%s, http %d)", e.Body.Error.Message, e.Body.Error.Code, e.StatusCode)
}

func (c *Client) do(ctx context.Context, method, path string, body, dest any) error {
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("X-Request-ID") == "" {
		if id, ok := ctx.Value(requestIDKey{}).(string); ok && id != "" {
			req.Header.Set("X-Request-ID", id)
		}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var env apierror.Body
		_ = json.Unmarshal(raw, &env)
		if env.Error.Message == "" {
			env.Error.Message = strings.TrimSpace(string(raw))
		}
		return &APIError{StatusCode: resp.StatusCode, Body: env}
	}
	if dest == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dest)
}

type requestIDKey struct{}

// WithRequestID stores an X-Request-ID value on ctx for outbound calls.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}
