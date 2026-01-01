package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client defines the interface for making HTTP requests (non-generic version).
type Client interface {
	Get(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error)
	Post(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error)
	Put(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error)
	Delete(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error)
	Patch(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error)
}

// GenericClient provides type-safe HTTP client methods using generics.
// This is the recommended way to use the client for type-safe operations.
// Methods on GenericClient support type parameters for type-safe request/response handling.
type GenericClient struct {
	client *client
}

// Config holds configuration for the HTTP client.
type Config struct {
	BaseURL        string
	DefaultTimeout time.Duration
	DefaultHeaders map[string]string
	HTTPClient     *http.Client
}

// client is the concrete implementation of Client.
type client struct {
	baseURL        string
	defaultTimeout time.Duration
	defaultHeaders http.Header
	httpClient     *http.Client
}

// New creates a new HTTP client with the given configuration.
func New(cfg Config) (Client, error) {
	c, err := NewGeneric(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewGeneric creates a new generic HTTP client with the given configuration.
func NewGeneric(cfg Config) (*GenericClient, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.DefaultTimeout,
		}
	}

	defaultHeaders := make(http.Header)
	for k, v := range cfg.DefaultHeaders {
		defaultHeaders.Set(k, v)
	}

	baseClient := &client{
		baseURL:        cfg.BaseURL,
		defaultTimeout: cfg.DefaultTimeout,
		defaultHeaders: defaultHeaders,
		httpClient:     httpClient,
	}

	return &GenericClient{client: baseClient}, nil
}

// Get performs a GET request.
func (c *client) Get(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c, ctx, http.MethodGet, path, opts...)
}

// Post performs a POST request.
func (c *client) Post(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c, ctx, http.MethodPost, path, opts...)
}

// Put performs a PUT request.
func (c *client) Put(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c, ctx, http.MethodPut, path, opts...)
}

// Delete performs a DELETE request.
func (c *client) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c, ctx, http.MethodDelete, path, opts...)
}

// Patch performs a PATCH request.
func (c *client) Patch(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c, ctx, http.MethodPatch, path, opts...)
}

// Get implements Client interface for GenericClient.
func (c *GenericClient) Get(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c.client, ctx, http.MethodGet, path, opts...)
}

// Post implements Client interface for GenericClient.
func (c *GenericClient) Post(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c.client, ctx, http.MethodPost, path, opts...)
}

// Put implements Client interface for GenericClient.
func (c *GenericClient) Put(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c.client, ctx, http.MethodPut, path, opts...)
}

// Delete implements Client interface for GenericClient.
func (c *GenericClient) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c.client, ctx, http.MethodDelete, path, opts...)
}

// Patch implements Client interface for GenericClient.
func (c *GenericClient) Patch(ctx context.Context, path string, opts ...RequestOption) (*Response[any], error) {
	return do[any](c.client, ctx, http.MethodPatch, path, opts...)
}

// Get performs a type-safe GET request.
func Get[T any](c *GenericClient, ctx context.Context, path string, opts ...RequestOption) (*Response[T], error) {
	return do[T](c.client, ctx, http.MethodGet, path, opts...)
}

// Post performs a type-safe POST request.
func Post[T any](c *GenericClient, ctx context.Context, path string, opts ...RequestOption) (*Response[T], error) {
	return do[T](c.client, ctx, http.MethodPost, path, opts...)
}

// Put performs a type-safe PUT request.
func Put[T any](c *GenericClient, ctx context.Context, path string, opts ...RequestOption) (*Response[T], error) {
	return do[T](c.client, ctx, http.MethodPut, path, opts...)
}

// Delete performs a type-safe DELETE request.
func Delete[T any](c *GenericClient, ctx context.Context, path string, opts ...RequestOption) (*Response[T], error) {
	return do[T](c.client, ctx, http.MethodDelete, path, opts...)
}

// Patch performs a type-safe PATCH request.
func Patch[T any](c *GenericClient, ctx context.Context, path string, opts ...RequestOption) (*Response[T], error) {
	return do[T](c.client, ctx, http.MethodPatch, path, opts...)
}

// do performs the HTTP request and decodes the response.
func do[T any](c *client, ctx context.Context, method, path string, opts ...RequestOption) (*Response[T], error) {
	// Build request
	req, err := buildRequest(ctx, method, c.baseURL, path, c.defaultHeaders, opts...)
	if err != nil {
		return nil, err
	}

	// Apply request timeout if specified in options
	cfg := &requestConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &RequestError{Err: fmt.Errorf("execute request: %w", err)}
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &RequestError{Err: fmt.Errorf("read response body: %w", err)}
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, NewHTTPError(resp, bodyBytes)
	}

	// Decode response body
	var body T
	if len(bodyBytes) > 0 {
		var zero T
		switch any(zero).(type) {
		case string:
			body = any(string(bodyBytes)).(T)
		case []byte:
			body = any(bodyBytes).(T)
		default:
			// Decode as JSON
			if err := json.Unmarshal(bodyBytes, &body); err != nil {
				return nil, &DecodeError{Err: fmt.Errorf("decode response: %w", err)}
			}
		}
	}

	return NewResponse(resp, body), nil
}
