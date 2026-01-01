package httpclient

import "net/http"

// Response represents an HTTP response with a typed body.
type Response[T any] struct {
	StatusCode int
	Headers    http.Header
	Body       T
	Raw        *http.Response
}

// NewResponse creates a new Response from an HTTP response and decoded body.
func NewResponse[T any](resp *http.Response, body T) *Response[T] {
	return &Response[T]{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		Raw:        resp,
	}
}

