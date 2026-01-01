package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RequestOption is a functional option for configuring HTTP requests.
type RequestOption func(*requestConfig)

// requestConfig holds the configuration for a single request.
type requestConfig struct {
	headers http.Header
	query   url.Values
	body    interface{}
	timeout time.Duration
	encoder func(interface{}) ([]byte, error)
}

// WithHeaders sets custom headers for the request.
func WithHeaders(headers map[string]string) RequestOption {
	return func(cfg *requestConfig) {
		if cfg.headers == nil {
			cfg.headers = make(http.Header)
		}
		for k, v := range headers {
			cfg.headers.Set(k, v)
		}
	}
}

// WithHeader sets a single header for the request.
func WithHeader(key, value string) RequestOption {
	return func(cfg *requestConfig) {
		if cfg.headers == nil {
			cfg.headers = make(http.Header)
		}
		cfg.headers.Set(key, value)
	}
}

// WithQuery sets query parameters for the request.
func WithQuery(params map[string]string) RequestOption {
	return func(cfg *requestConfig) {
		if cfg.query == nil {
			cfg.query = make(url.Values)
		}
		for k, v := range params {
			cfg.query.Set(k, v)
		}
	}
}

// WithQueryValue adds a single query parameter.
func WithQueryValue(key, value string) RequestOption {
	return func(cfg *requestConfig) {
		if cfg.query == nil {
			cfg.query = make(url.Values)
		}
		cfg.query.Set(key, value)
	}
}

// WithBody sets the request body with automatic JSON encoding.
func WithBody(body interface{}) RequestOption {
	return func(cfg *requestConfig) {
		cfg.body = body
		cfg.encoder = json.Marshal
	}
}

// WithBodyEncoder sets the request body with a custom encoder.
func WithBodyEncoder(body interface{}, encoder func(interface{}) ([]byte, error)) RequestOption {
	return func(cfg *requestConfig) {
		cfg.body = body
		cfg.encoder = encoder
	}
}

// WithTimeout sets a timeout for the request.
func WithTimeout(timeout time.Duration) RequestOption {
	return func(cfg *requestConfig) {
		cfg.timeout = timeout
	}
}

// buildRequest constructs an http.Request from the configuration.
func buildRequest(ctx context.Context, method, baseURL, path string, defaultHeaders http.Header, opts ...RequestOption) (*http.Request, error) {
	cfg := &requestConfig{
		headers: make(http.Header),
		query:   make(url.Values),
	}

	// Apply default headers
	for k, v := range defaultHeaders {
		cfg.headers[k] = v
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	// Build URL
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, &RequestError{Err: fmt.Errorf("invalid base URL: %w", err)}
	}

	reqURL.Path = path
	if len(cfg.query) > 0 {
		reqURL.RawQuery = cfg.query.Encode()
	}

	// Create request body if provided
	var body []byte
	if cfg.body != nil {
		if cfg.encoder == nil {
			cfg.encoder = json.Marshal
		}
		body, err = cfg.encoder(cfg.body)
		if err != nil {
			return nil, &RequestError{Err: fmt.Errorf("encode body: %w", err)}
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), nil)
	if err != nil {
		return nil, &RequestError{Err: fmt.Errorf("create request: %w", err)}
	}

	// Set headers
	for k, v := range cfg.headers {
		req.Header[k] = v
	}

	// Set body if provided
	if len(body) > 0 {
		req.Body = &bodyReader{data: body}
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	return req, nil
}

// bodyReader implements io.ReadCloser for request bodies.
type bodyReader struct {
	data []byte
	pos  int
}

func (r *bodyReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *bodyReader) Close() error {
	r.pos = 0
	return nil
}
