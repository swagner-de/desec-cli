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
		_, _ = w.Write([]byte("[]"))
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
		_, _ = w.Write([]byte(`{"detail": "bad request"}`))
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
		_, _ = w.Write([]byte(`[{"name":"example.com","created":"2024-01-01T00:00:00Z","published":"2024-01-01T00:00:00Z","minimum_ttl":3600}]`))
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
		_, _ = w.Write([]byte(`{"name":"example.com","created":"2024-01-01T00:00:00Z","published":"2024-01-01T00:00:00Z","minimum_ttl":3600}`))
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
		_, _ = w.Write([]byte(`{"name":"new.com","created":"2024-01-01T00:00:00Z","published":"2024-01-01T00:00:00Z","minimum_ttl":3600}`))
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

func TestListRRsets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/domains/example.com/rrsets/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"subname":"www","type":"A","records":["1.2.3.4"],"ttl":3600}]`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	rrsets, err := c.ListRRsets("example.com", "", "")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(rrsets) != 1 { t.Fatalf("expected 1 rrset, got %d", len(rrsets)) }
	if rrsets[0].Type != "A" { t.Fatalf("expected type 'A', got '%s'", rrsets[0].Type) }
}

func TestListRRsetsWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "A" { t.Fatalf("expected type filter 'A'") }
		if r.URL.Query().Get("subname") != "www" { t.Fatalf("expected subname filter 'www'") }
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"subname":"www","type":"A","records":["1.2.3.4"],"ttl":3600}]`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	_, err := c.ListRRsets("example.com", "A", "www")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestCreateRRset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" { t.Fatalf("expected POST, got %s", r.Method) }
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"subname":"www","type":"A","records":["1.2.3.4"],"ttl":3600}`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	rrset, err := c.CreateRRset("example.com", &RRsetCreate{Subname: "www", Type: "A", Records: []string{"1.2.3.4"}, TTL: 3600})
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rrset.Subname != "www" { t.Fatalf("expected subname 'www', got '%s'", rrset.Subname) }
}

func TestDeleteRRset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" { t.Fatalf("expected DELETE, got %s", r.Method) }
		if r.URL.Path != "/api/v1/domains/example.com/rrsets/www/A/" { t.Fatalf("unexpected path: %s", r.URL.Path) }
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	err := c.DeleteRRset("example.com", "www", "A")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestListTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/tokens/" { t.Fatalf("unexpected path: %s", r.URL.Path) }
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"abc123","name":"test","perm_manage_tokens":true,"is_valid":true}]`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	tokens, err := c.ListTokens()
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(tokens) != 1 { t.Fatalf("expected 1, got %d", len(tokens)) }
	if tokens[0].ID != "abc123" { t.Fatalf("expected 'abc123', got '%s'", tokens[0].ID) }
}

func TestCreateToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" { t.Fatalf("expected POST, got %s", r.Method) }
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"new123","name":"new-token","token":"secret-value","is_valid":true}`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	token, err := c.CreateToken(&TokenCreate{Name: "new-token"})
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if token.Token != "secret-value" { t.Fatalf("expected 'secret-value', got '%s'", token.Token) }
}

func TestDeleteToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" { t.Fatalf("expected DELETE") }
		if r.URL.Path != "/api/v1/auth/tokens/abc123/" { t.Fatalf("unexpected path: %s", r.URL.Path) }
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	err := c.DeleteToken("abc123")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestListPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/tokens/tok1/policies/rrsets/" { t.Fatalf("unexpected path: %s", r.URL.Path) }
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"pol1","domain":"example.com","subname":"www","type":"A","perm_write":true}]`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	policies, err := c.ListPolicies("tok1")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(policies) != 1 { t.Fatalf("expected 1, got %d", len(policies)) }
	if policies[0].ID != "pol1" { t.Fatalf("expected 'pol1', got '%s'", policies[0].ID) }
}

func TestCreatePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" { t.Fatalf("expected POST, got %s", r.Method) }
		if r.URL.Path != "/api/v1/auth/tokens/tok1/policies/rrsets/" { t.Fatalf("unexpected path: %s", r.URL.Path) }
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"pol2","domain":"example.com","subname":null,"type":null,"perm_write":false}`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	domain := "example.com"
	policy, err := c.CreatePolicy("tok1", &TokenPolicyCreate{Domain: &domain})
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if policy.ID != "pol2" { t.Fatalf("expected 'pol2', got '%s'", policy.ID) }
}

func TestDeletePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" { t.Fatalf("expected DELETE, got %s", r.Method) }
		if r.URL.Path != "/api/v1/auth/tokens/tok1/policies/rrsets/pol1/" { t.Fatalf("unexpected path: %s", r.URL.Path) }
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	err := c.DeletePolicy("tok1", "pol1")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestDynDNSUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" { t.Fatalf("expected GET, got %s", r.Method) }
		if r.URL.Query().Get("hostname") != "example.com" { t.Fatalf("expected hostname 'example.com'") }
		if r.URL.Query().Get("myipv4") != "1.2.3.4" { t.Fatalf("expected myipv4 '1.2.3.4'") }
		auth := r.Header.Get("Authorization")
		if auth != "Token test-token" { t.Fatalf("expected 'Token test-token', got '%s'", auth) }
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("good"))
	}))
	defer server.Close()
	c := New("test-token")
	c.dynDNSURL = server.URL
	err := c.DynDNSUpdate("example.com", "1.2.3.4", "")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}
