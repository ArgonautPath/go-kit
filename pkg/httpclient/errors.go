package httpclient

import (
	"fmt"
	"net/http"
)

// HTTPError represents an HTTP error response.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: %d %s", e.StatusCode, e.Status)
}

// RequestError represents an error that occurred while making an HTTP request.
type RequestError struct {
	Err error
}

// Error implements the error interface.
func (e *RequestError) Error() string {
	return fmt.Sprintf("request error: %v", e.Err)
}

// Unwrap returns the underlying error.
func (e *RequestError) Unwrap() error {
	return e.Err
}

// DecodeError represents an error that occurred while decoding a response body.
type DecodeError struct {
	Err error
}

// Error implements the error interface.
func (e *DecodeError) Error() string {
	return fmt.Sprintf("decode error: %v", e.Err)
}

// Unwrap returns the underlying error.
func (e *DecodeError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTPError from an HTTP response.
func NewHTTPError(resp *http.Response, body []byte) *HTTPError {
	return &HTTPError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       body,
	}
}

// IsHTTPError checks if an error is an HTTPError and returns it.
func IsHTTPError(err error) (*HTTPError, bool) {
	httpErr, ok := err.(*HTTPError)
	return httpErr, ok
}
