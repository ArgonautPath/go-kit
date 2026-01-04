package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ArgonautPath/go-kit/pkg/logger"
)

func TestChain(t *testing.T) {
	callOrder := []string{}

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder = append(callOrder, "m1-before")
			next.ServeHTTP(w, r)
			callOrder = append(callOrder, "m1-after")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder = append(callOrder, "m2-before")
			next.ServeHTTP(w, r)
			callOrder = append(callOrder, "m2-after")
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callOrder = append(callOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})

	chain := Chain(m1, m2)
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	callOrder = []string{}
	wrapped.ServeHTTP(w, req)

	expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if len(callOrder) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(callOrder))
	}

	for i, exp := range expected {
		if i < len(callOrder) && callOrder[i] != exp {
			t.Errorf("Expected call[%d] = %q, got %q", i, exp, callOrder[i])
		}
	}
}

func TestApply(t *testing.T) {
	called := false
	m := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := Apply(m, handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if !called {
		t.Error("Middleware was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestApplyFunc(t *testing.T) {
	called := false
	m := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	wrapped := ApplyFunc(m, handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if !called {
		t.Error("Middleware was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequestID(t *testing.T) {
	tests := []struct {
		name           string
		headerValue    string
		expectGenerated bool
	}{
		{
			name:            "no header - generate",
			headerValue:     "",
			expectGenerated: true,
		},
		{
			name:            "with header - use existing",
			headerValue:     "existing-id",
			expectGenerated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestID string
			handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestID = GetRequestID(r.Context())
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerValue != "" {
				req.Header.Set(RequestIDHeader, tt.headerValue)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if requestID == "" {
				t.Error("Request ID was not set")
			}

			if tt.expectGenerated {
				if requestID == tt.headerValue {
					t.Error("Expected generated ID, got existing")
				}
			} else {
				if requestID != tt.headerValue {
					t.Errorf("Expected %q, got %q", tt.headerValue, requestID)
				}
			}

			// Check response header
			if w.Header().Get(RequestIDHeader) != requestID {
				t.Errorf("Response header mismatch: expected %q, got %q", requestID, w.Header().Get(RequestIDHeader))
			}
		})
	}
}

func TestRequestID_CustomHeader(t *testing.T) {
	handler := RequestID(WithRequestIDHeader("X-Custom-ID"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Custom-ID", "custom-id")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Custom-ID") != "custom-id" {
		t.Errorf("Expected custom header %q, got %q", "custom-id", w.Header().Get("X-Custom-ID"))
	}
}

func TestRequestID_NoResponse(t *testing.T) {
	handler := RequestID(WithRequestIDResponse(false))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get(RequestIDHeader) != "" {
		t.Error("Expected no response header when disabled")
	}
}

func TestLogging(t *testing.T) {
	log, _ := logger.New(logger.Config{
		Level:  logger.InfoLevel,
		Output: logger.NewStdoutWriter(logger.JSONFormat),
		Format: logger.JSONFormat,
	})

	handler := Logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLogging_SkipPath(t *testing.T) {
	log, _ := logger.New(logger.Config{
		Level:  logger.InfoLevel,
		Output: logger.NewStdoutWriter(logger.JSONFormat),
		Format: logger.JSONFormat,
	})

	handler := Logging(log, WithSkipPaths("/health", "/metrics"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		path     string
		shouldLog bool
	}{
		{"/health", false},
		{"/metrics", false},
		{"/api/users", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			// Just verify it doesn't crash - actual logging verification would require a mock logger
		})
	}
}

func TestRecovery(t *testing.T) {
	handler := Recovery()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Error("Expected error message in response")
	}
}

func TestRecovery_CustomHandler(t *testing.T) {
	called := false
	handler := Recovery(WithRecoveryHandler(func(w http.ResponseWriter, r *http.Request, err interface{}) {
		called = true
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Custom error"))
	}))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("Custom handler was not called")
	}

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	handler := Recovery()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expectHeader   bool
	}{
		{
			name:           "allowed origin",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com"},
			expectHeader:   true,
		},
		{
			name:           "disallowed origin",
			origin:         "https://evil.com",
			allowedOrigins: []string{"https://example.com"},
			expectHeader:   false,
		},
		{
			name:           "wildcard origin",
			origin:         "https://any.com",
			allowedOrigins: []string{"*"},
			expectHeader:   true,
		},
		{
			name:           "no origin header",
			origin:         "",
			allowedOrigins: []string{"https://example.com"},
			expectHeader:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CORS(CORSConfig{
				AllowedOrigins: tt.allowedOrigins,
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			header := w.Header().Get("Access-Control-Allow-Origin")
			if tt.expectHeader && header == "" {
				t.Error("Expected CORS header, got none")
			}
			if !tt.expectHeader && header != "" {
				t.Errorf("Expected no CORS header, got %q", header)
			}
		})
	}
}

func TestCORS_Preflight(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("Expected CORS origin header")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected CORS methods header")
	}
}

func TestTimeout(t *testing.T) {
	handler := Timeout(100 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Exceed timeout
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	handler.ServeHTTP(w, req)
	duration := time.Since(start)

	if duration > 150*time.Millisecond {
		t.Error("Handler should have timed out quickly")
	}

	if w.Code != http.StatusRequestTimeout {
		t.Errorf("Expected status %d, got %d", http.StatusRequestTimeout, w.Code)
	}
}

func TestTimeout_NoTimeout(t *testing.T) {
	handler := Timeout(1 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Within timeout
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetRequestID(t *testing.T) {
	ctx := context.Background()
	if GetRequestID(ctx) != "" {
		t.Error("Expected empty request ID for empty context")
	}

	ctx = context.WithValue(ctx, RequestIDContextKey, "test-id")
	if GetRequestID(ctx) != "test-id" {
		t.Errorf("Expected %q, got %q", "test-id", GetRequestID(ctx))
	}
}

