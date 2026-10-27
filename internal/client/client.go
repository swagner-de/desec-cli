package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL       = "https://desec.io/api/v1"
	maxPaginationPages   = 1000
	maxHTTPRedirectCount = 10
	defaultHTTPTimeout   = 30 * time.Second
	defaultDynDNSTimeout = 60 * time.Second
)

type Client struct {
	token      string
	baseURL    string
	dynDNSURL  string
	http       *http.Client
	dynDNSHTTP *http.Client
}

func New(token string) *Client {
	return &Client{
		token:     token,
		baseURL:   defaultBaseURL,
		dynDNSURL: "https://update.dedyn.io",
		http: &http.Client{
			Timeout:       defaultHTTPTimeout,
			CheckRedirect: sameOriginRedirectPolicy,
		},
		dynDNSHTTP: &http.Client{
			Timeout: defaultDynDNSTimeout,
		},
	}
}

// sameOriginRedirectPolicy prevents credentials from being forwarded to a
// different origin via an HTTP redirect.
func sameOriginRedirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	if len(via) >= maxHTTPRedirectCount {
		return fmt.Errorf("stopped after %d redirects", maxHTTPRedirectCount)
	}
	origin := via[0].URL
	if req.URL.Scheme != origin.Scheme || !strings.EqualFold(req.URL.Host, origin.Host) {
		return fmt.Errorf("refusing redirect from %s to %s", origin, req.URL)
	}
	return nil
}

func (c *Client) doRequest(method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(respBody, apiErr); err != nil {
			apiErr.Detail = string(respBody)
		}
		return nil, apiErr
	}

	return respBody, nil
}

func (c *Client) doRequestPaginated(path string) ([]byte, error) {
	var allResults []json.RawMessage

	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parsing pagination URL: %w", err)
	}
	query := u.Query()
	query.Set("cursor", "")
	u.RawQuery = query.Encode()
	nextURL := u.String()
	seen := make(map[string]struct{})

	for page := 0; nextURL != ""; page++ {
		if page >= maxPaginationPages {
			return nil, fmt.Errorf("pagination exceeded %d pages", maxPaginationPages)
		}
		if _, ok := seen[nextURL]; ok {
			return nil, fmt.Errorf("pagination loop detected at %s", nextURL)
		}
		seen[nextURL] = struct{}{}

		pageURL, err := c.validatePaginationURL(nextURL)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("GET", pageURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Token "+c.token)

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		if resp.StatusCode >= 400 {
			apiErr := &APIError{StatusCode: resp.StatusCode}
			if err := json.Unmarshal(body, apiErr); err != nil {
				apiErr.Detail = string(body)
			}
			return nil, apiErr
		}

		var pageResults []json.RawMessage
		if err := json.Unmarshal(body, &pageResults); err != nil {
			return nil, fmt.Errorf("parsing page: %w", err)
		}
		allResults = append(allResults, pageResults...)

		nextURL = getNextLink(strings.Join(resp.Header.Values("Link"), ","))
	}

	return json.Marshal(allResults)
}

// validatePaginationURL ensures that a Link header cannot redirect the
// authenticated request to a different origin or outside the API base path.
func (c *Client) validatePaginationURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parsing pagination URL: %w", err)
	}
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	if u.User != nil || u.Scheme != base.Scheme || !strings.EqualFold(u.Host, base.Host) {
		return nil, fmt.Errorf("refusing pagination link to untrusted URL %q", raw)
	}

	basePath := strings.TrimSuffix(base.EscapedPath(), "/")
	if basePath != "" {
		path := u.EscapedPath()
		if path != basePath && !strings.HasPrefix(path, basePath+"/") {
			return nil, fmt.Errorf("refusing pagination link outside API base path: %q", raw)
		}
	}

	return u, nil
}

func getNextLink(linkHeader string) string {
	for _, part := range splitLinkHeader(linkHeader) {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(part, "<") {
			continue
		}
		end := strings.Index(part, ">")
		if end < 0 {
			continue
		}
		target := part[1:end]
		for _, param := range strings.Split(part[end+1:], ";") {
			param = strings.TrimSpace(param)
			if !strings.HasPrefix(strings.ToLower(param), "rel=") {
				continue
			}
			rel := strings.Trim(strings.TrimSpace(param[4:]), `"`)
			for _, value := range strings.Fields(rel) {
				if strings.EqualFold(value, "next") {
					return target
				}
			}
		}
	}
	return ""
}

// splitLinkHeader splits a Link header on commas that are outside URI
// references and quoted parameter values.
func splitLinkHeader(header string) []string {
	var parts []string
	start := 0
	inURI := false
	inQuote := false

	for i := 0; i < len(header); i++ {
		switch header[i] {
		case '<':
			if !inQuote {
				inURI = true
			}
		case '>':
			if !inQuote {
				inURI = false
			}
		case '"':
			if !inURI {
				inQuote = !inQuote
			}
		case ',':
			if !inURI && !inQuote {
				parts = append(parts, header[start:i])
				start = i + 1
			}
		}
	}

	return append(parts, header[start:])
}
