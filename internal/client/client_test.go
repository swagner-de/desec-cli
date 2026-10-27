package client

import (
	"encoding/json"
	"fmt"
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rrsets) != 1 {
		t.Fatalf("expected 1 rrset, got %d", len(rrsets))
	}
	if rrsets[0].Type != "A" {
		t.Fatalf("expected type 'A', got '%s'", rrsets[0].Type)
	}
}

func TestListRRsetsWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "A" {
			t.Fatalf("expected type filter 'A'")
		}
		if r.URL.Query().Get("subname") != "www" {
			t.Fatalf("expected subname filter 'www'")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"subname":"www","type":"A","records":["1.2.3.4"],"ttl":3600}]`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	_, err := c.ListRRsets("example.com", "A", "www")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateRRset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"subname":"www","type":"A","records":["1.2.3.4"],"ttl":3600}`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	rrset, err := c.CreateRRset("example.com", &RRsetCreate{Subname: "www", Type: "A", Records: []string{"1.2.3.4"}, TTL: 3600})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rrset.Subname != "www" {
		t.Fatalf("expected subname 'www', got '%s'", rrset.Subname)
	}
}

func TestDeleteRRset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/domains/example.com/rrsets/www/A/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	err := c.DeleteRRset("example.com", "www", "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/tokens/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"abc123","name":"test","perm_manage_tokens":true,"is_valid":true}]`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	tokens, err := c.ListTokens()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1, got %d", len(tokens))
	}
	if tokens[0].ID != "abc123" {
		t.Fatalf("expected 'abc123', got '%s'", tokens[0].ID)
	}
}

func TestCreateToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"new123","name":"new-token","token":"secret-value","is_valid":true}`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	token, err := c.CreateToken(&TokenCreate{Name: "new-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.Token != "secret-value" {
		t.Fatalf("expected 'secret-value', got '%s'", token.Token)
	}
}

func TestDeleteToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("expected DELETE")
		}
		if r.URL.Path != "/api/v1/auth/tokens/abc123/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	err := c.DeleteToken("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/tokens/tok1/policies/rrsets/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"pol1","domain":"example.com","subname":"www","type":"A","perm_write":true}]`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	policies, err := c.ListPolicies("tok1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 1 {
		t.Fatalf("expected 1, got %d", len(policies))
	}
	if policies[0].ID != "pol1" {
		t.Fatalf("expected 'pol1', got '%s'", policies[0].ID)
	}
}

func TestCreatePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/auth/tokens/tok1/policies/rrsets/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"pol2","domain":"example.com","subname":null,"type":null,"perm_write":false}`))
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	domain := "example.com"
	policy, err := c.CreatePolicy("tok1", &TokenPolicyCreate{Domain: &domain})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != "pol2" {
		t.Fatalf("expected 'pol2', got '%s'", policy.ID)
	}
}

func TestDeletePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/auth/tokens/tok1/policies/rrsets/pol1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	err := c.DeletePolicy("tok1", "pol1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDynDNSUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("hostname") != "example.com" {
			t.Fatalf("expected hostname 'example.com'")
		}
		if r.URL.Query().Get("myipv4") != "1.2.3.4" {
			t.Fatalf("expected myipv4 '1.2.3.4'")
		}
		auth := r.Header.Get("Authorization")
		if auth != "Token test-token" {
			t.Fatalf("expected 'Token test-token', got '%s'", auth)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("good"))
	}))
	defer server.Close()
	c := New("test-token")
	c.dynDNSURL = server.URL
	c.dynDNSHTTP = server.Client()
	err := c.DynDNSUpdate("example.com", "1.2.3.4", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaginationRejectsCrossOriginLink(t *testing.T) {
	var foreignRequests int
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		foreignRequests++
		_, _ = w.Write([]byte(`[]`))
	}))
	defer foreign.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", fmt.Sprintf("<%s/api/v1/domains/?cursor=next>; rel=\"next\"", foreign.URL))
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	if _, err := c.ListDomains(); err == nil {
		t.Fatal("expected cross-origin pagination link to be rejected")
	}
	if foreignRequests != 0 {
		t.Fatalf("foreign server received %d requests", foreignRequests)
	}
}

func TestPaginationDetectsLoop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", fmt.Sprintf("<http://%s%s?cursor=next>; rel=\"next\"", r.Host, r.URL.Path))
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	if _, err := c.ListDomains(); err == nil {
		t.Fatal("expected pagination loop to be rejected")
	}
}

func TestPaginationFollowsSameOriginLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("cursor") {
		case "":
			w.Header().Set("Link", fmt.Sprintf("<http://%s%s?cursor=next>; rel=\"next\"", r.Host, r.URL.Path))
			_, _ = w.Write([]byte(`[{"name":"first.example","created":"2024-01-01T00:00:00Z","published":"2024-01-01T00:00:00Z","minimum_ttl":3600}]`))
		case "next":
			_, _ = w.Write([]byte(`[{"name":"second.example","created":"2024-01-01T00:00:00Z","published":"2024-01-01T00:00:00Z","minimum_ttl":3600}]`))
		default:
			http.Error(w, "unexpected cursor", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	domains, err := c.ListDomains()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}
	if domains[0].Name != "first.example" || domains[1].Name != "second.example" {
		t.Fatalf("unexpected domains: %#v", domains)
	}
}

func TestRedirectRejectsCrossOrigin(t *testing.T) {
	var foreignRequests int
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		foreignRequests++
		_, _ = w.Write([]byte(`{}`))
	}))
	defer foreign.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, foreign.URL, http.StatusFound)
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	if _, err := c.GetDomain("example.com"); err == nil {
		t.Fatal("expected cross-origin redirect to be rejected")
	}
	if foreignRequests != 0 {
		t.Fatalf("foreign server received %d requests", foreignRequests)
	}
}

func TestGetNextLink(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "next",
			header: `<https://desec.io/api/v1/domains/?cursor=abc>; rel="next"`,
			want:   "https://desec.io/api/v1/domains/?cursor=abc",
		},
		{
			name:   "multiple links",
			header: `<https://desec.io/api/v1/domains/?cursor=first>; rel="first", <https://desec.io/api/v1/domains/?cursor=abc>; rel="next"`,
			want:   "https://desec.io/api/v1/domains/?cursor=abc",
		},
		{
			name:   "unquoted rel",
			header: `<https://desec.io/api/v1/domains/?cursor=abc>; rel=next`,
			want:   "https://desec.io/api/v1/domains/?cursor=abc",
		},
		{
			name:   "comma inside URI",
			header: `<https://desec.io/api/v1/domains/?cursor=a,b>; rel="next"`,
			want:   "https://desec.io/api/v1/domains/?cursor=a,b",
		},
		{
			name:   "no next",
			header: `<https://desec.io/api/v1/domains/?cursor=abc>; rel="prev"`,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNextLink(tt.header); got != tt.want {
				t.Fatalf("getNextLink() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClientHasRequestTimeout(t *testing.T) {
	c := New("test-token")
	if c.http.Timeout != defaultHTTPTimeout {
		t.Fatalf("expected timeout %s, got %s", defaultHTTPTimeout, c.http.Timeout)
	}
}

func TestClientHasDedicatedDynDNSHTTPClient(t *testing.T) {
	c := New("test-token")
	if c.dynDNSHTTP == c.http {
		t.Fatal("expected dynDNS to use a dedicated HTTP client")
	}
	if c.dynDNSHTTP.Timeout != defaultDynDNSTimeout {
		t.Fatalf("expected dynDNS timeout %s, got %s", defaultDynDNSTimeout, c.dynDNSHTTP.Timeout)
	}
	if c.dynDNSHTTP.CheckRedirect != nil {
		t.Fatal("expected dynDNS client to use default redirect behavior")
	}
	if c.http.CheckRedirect == nil {
		t.Fatal("expected API client to keep the redirect policy")
	}
}

func TestPathComponentsAreEscaped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/domains/example.com%2Fextra/rrsets/www%2Fsub/A%3F/"
		if got := r.URL.EscapedPath(); got != want {
			t.Errorf("escaped path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"subname":"www/sub","type":"A?","records":[],"ttl":3600}`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	_, err := c.GetRRset("example.com/extra", "www/sub", "A?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaginationStartsWithEmptyCursor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, ok := r.URL.Query()["cursor"]
		if !ok {
			t.Error("expected cursor query parameter")
		} else if len(values) != 1 || values[0] != "" {
			t.Errorf("cursor = %#v, want one empty value", values)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	if _, err := c.ListDomains(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateRRsetNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	rrset, err := c.UpdateRRset("example.com", "www", "A", map[string]any{"records": []string{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rrset != nil {
		t.Fatalf("expected nil rrset after 204, got %#v", rrset)
	}
}

func TestBulkRRsetsPreservesPartialPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		var body []map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if len(body) != 1 {
			t.Fatalf("expected one RRset, got %d", len(body))
		}
		if _, ok := body[0]["ttl"]; ok {
			t.Error("expected ttl to be omitted")
		}
		records, ok := body[0]["records"].([]any)
		if !ok || len(records) != 0 {
			t.Errorf("expected empty records array, got %#v", body[0]["records"])
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	records := []string{}
	if _, err := c.BulkRRsets("example.com", []RRsetUpdate{{Subname: "www", Type: "A", Records: &records}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListRRsetsApexFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, ok := r.URL.Query()["subname"]
		if !ok {
			t.Error("expected subname query parameter for apex filter")
		} else if len(values) != 1 || values[0] != "" {
			t.Errorf("subname = %#v, want one empty value", values)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	if _, err := c.ListRRsets("example.com", "", "@"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateRRsetUnexpectedEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := New("test-token")
	c.baseURL = server.URL + "/api/v1"
	_, err := c.UpdateRRset("example.com", "www", "A", map[string]any{"ttl": 600})
	if err == nil {
		t.Fatal("expected unexpected empty response to be rejected")
	}
}
