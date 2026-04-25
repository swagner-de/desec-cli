package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := New("test-token")
	if c.token != "test-token" {
		t.Fatalf("expected token 'test-token', got '%s'", c.token)
	}
	if c.baseURL != "https://desec.io/api/v1" {
		t.Fatalf("unexpected baseURL: %s", c.baseURL)
	}
}

func TestClientAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Token test-token" {
			t.Fatalf("expected 'Token test-token', got '%s'", auth)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL
	_, err := c.doRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"detail": "bad request"}`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL
	_, err := c.doRequest("GET", "/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", apiErr.StatusCode)
	}
}

func TestListDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/domains/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"name":"example.com","created":"2024-01-01T00:00:00Z","published":"2024-01-01T00:00:00Z","minimum_ttl":3600}]`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	domains, err := c.ListDomains()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(domains))
	}
	if domains[0].Name != "example.com" {
		t.Fatalf("expected 'example.com', got '%s'", domains[0].Name)
	}
}

func TestGetDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/domains/example.com/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"example.com","created":"2024-01-01T00:00:00Z","published":"2024-01-01T00:00:00Z","minimum_ttl":3600}`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	domain, err := c.GetDomain("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.Name != "example.com" {
		t.Fatalf("expected 'example.com', got '%s'", domain.Name)
	}
}

func TestCreateDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"name":"new.com","created":"2024-01-01T00:00:00Z","published":"2024-01-01T00:00:00Z","minimum_ttl":3600}`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	domain, err := c.CreateDomain("new.com", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.Name != "new.com" {
		t.Fatalf("expected 'new.com', got '%s'", domain.Name)
	}
}

func TestDeleteDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	err := c.DeleteDomain("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
