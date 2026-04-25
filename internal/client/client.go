package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://desec.io/api/v1"

type Client struct {
	token     string
	baseURL   string
	dynDNSURL string
	http      *http.Client
}

func New(token string) *Client {
	return &Client{
		token:     token,
		baseURL:   defaultBaseURL,
		dynDNSURL: "https://update.dedyn.io",
		http:      &http.Client{},
	}
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

	url := c.baseURL + path
	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
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

		var page []json.RawMessage
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("parsing page: %w", err)
		}
		allResults = append(allResults, page...)

		url = getNextLink(resp.Header.Get("Link"))
	}

	return json.Marshal(allResults)
}

func getNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	for _, part := range strings.Split(linkHeader, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, `rel="next"`) {
			start := strings.Index(part, "<") + 1
			end := strings.Index(part, ">")
			if start > 0 && end > start {
				return part[start:end]
			}
		}
	}
	return ""
}
