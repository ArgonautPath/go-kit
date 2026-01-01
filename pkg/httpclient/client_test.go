package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestUser represents a test user type for generic testing.
type TestUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PostData represents a test post type for generic testing.
type PostData struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"user_id"`
}

func TestNewGeneric(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				BaseURL:        "https://api.example.com",
				DefaultTimeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			cfg: Config{
				BaseURL: "",
			},
			wantErr: true,
		},
		{
			name: "with default headers",
			cfg: Config{
				BaseURL: "https://api.example.com",
				DefaultHeaders: map[string]string{
					"Authorization": "Bearer token",
					"Content-Type":  "application/json",
				},
			},
			wantErr: false,
		},
		{
			name: "with custom HTTP client",
			cfg: Config{
				BaseURL: "https://api.example.com",
				HTTPClient: &http.Client{
					Timeout: 10 * time.Second,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeneric(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGeneric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewGeneric() returned nil client without error")
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantErr        bool
		expectedType   string
	}{
		{
			name:           "successful GET with JSON object",
			responseBody:   `{"id":1,"name":"John Doe","email":"john@example.com"}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedType:   "user",
		},
		{
			name:           "successful GET with JSON array",
			responseBody:   `[{"id":1,"name":"John"},{"id":2,"name":"Jane"}]`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedType:   "users",
		},
		{
			name:           "successful GET with string response",
			responseBody:   "plain text response",
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedType:   "string",
		},
		{
			name:           "HTTP error 404",
			responseBody:   `{"error":"not found"}`,
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name:           "HTTP error 500",
			responseBody:   `{"error":"internal server error"}`,
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := NewGeneric(Config{
				BaseURL:        server.URL,
				DefaultTimeout: 5 * time.Second,
			})
			if err != nil {
				t.Fatalf("NewGeneric() error = %v", err)
			}

			ctx := context.Background()

			switch tt.expectedType {
			case "user":
				resp, err := Get[TestUser](client, ctx, "/users/1")
				if (err != nil) != tt.wantErr {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					if resp.StatusCode != tt.responseStatus {
						t.Errorf("Get() status = %d, want %d", resp.StatusCode, tt.responseStatus)
					}
					if resp.Body.ID != 1 {
						t.Errorf("Get() body.ID = %d, want 1", resp.Body.ID)
					}
				} else {
					if resp != nil {
						t.Error("Get() should return nil response on error")
					}
					httpErr, ok := IsHTTPError(err)
					if !ok {
						t.Error("Get() should return HTTPError on HTTP error")
					} else if httpErr.StatusCode != tt.responseStatus {
						t.Errorf("HTTPError.StatusCode = %d, want %d", httpErr.StatusCode, tt.responseStatus)
					}
				}
			case "users":
				resp, err := Get[[]TestUser](client, ctx, "/users")
				if (err != nil) != tt.wantErr {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					if len(resp.Body) != 2 {
						t.Errorf("Get() body length = %d, want 2", len(resp.Body))
					}
				}
			case "string":
				resp, err := Get[string](client, ctx, "/text")
				if (err != nil) != tt.wantErr {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					if resp.Body != tt.responseBody {
						t.Errorf("Get() body = %q, want %q", resp.Body, tt.responseBody)
					}
				}
			default:
				// Test error case
				resp, err := Get[TestUser](client, ctx, "/users/1")
				if (err != nil) != tt.wantErr {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && resp == nil {
					t.Error("Get() returned nil response without error")
				}
			}
		})
	}
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var post PostData
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		post.ID = 1
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(post)
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx := context.Background()
	newPost := PostData{
		Title:  "Test Post",
		Body:   "This is a test post",
		UserID: 1,
	}

	resp, err := Post[PostData](client, ctx, "/posts", WithBody(newPost))
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Post() status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	if resp.Body.ID != 1 {
		t.Errorf("Post() body.ID = %d, want 1", resp.Body.ID)
	}

	if resp.Body.Title != newPost.Title {
		t.Errorf("Post() body.Title = %q, want %q", resp.Body.Title, newPost.Title)
	}
}

func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var user TestUser
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx := context.Background()
	updatedUser := TestUser{
		ID:    1,
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}

	resp, err := Put[TestUser](client, ctx, "/users/1", WithBody(updatedUser))
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Put() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if resp.Body.Name != updatedUser.Name {
		t.Errorf("Put() body.Name = %q, want %q", resp.Body.Name, updatedUser.Name)
	}
}

func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx := context.Background()
	resp, err := Delete[string](client, ctx, "/users/1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updates)
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx := context.Background()
	updates := map[string]interface{}{
		"name": "Updated Name",
	}

	resp, err := Patch[map[string]interface{}](client, ctx, "/users/1", WithBody(updates))
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Patch() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Test"}`))
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx := context.Background()
	resp, err := Get[TestUser](client, ctx, "/users/1", WithHeader("Authorization", "Bearer test-token"))
	if err != nil {
		t.Fatalf("Get() with header error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Get() with header status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestWithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("limit") != "10" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx := context.Background()
	resp, err := Get[[]TestUser](client, ctx, "/users", WithQuery(map[string]string{
		"page":  "1",
		"limit": "10",
	}))
	if err != nil {
		t.Fatalf("Get() with query error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Get() with query status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = Get[TestUser](client, ctx, "/users/1")
	if err == nil {
		t.Error("Get() should return error on context timeout")
	}
}

func TestDefaultHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "secret-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
		DefaultHeaders: map[string]string{
			"X-API-Key": "secret-key",
		},
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx := context.Background()
	resp, err := Get[TestUser](client, ctx, "/users/1")
	if err != nil {
		t.Fatalf("Get() with default headers error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Get() with default headers status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json{`))
	}))
	defer server.Close()

	client, err := NewGeneric(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGeneric() error = %v", err)
	}

	ctx := context.Background()
	_, err = Get[TestUser](client, ctx, "/users/1")
	if err == nil {
		t.Error("Get() should return error on invalid JSON")
	}

	if _, ok := err.(*DecodeError); !ok {
		t.Errorf("Get() error type = %T, want *DecodeError", err)
	}
}

func TestClientInterface(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Test"}`))
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:        server.URL,
		DefaultTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, "/users/1")
	if err != nil {
		t.Fatalf("Client.Get() error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Client.Get() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

