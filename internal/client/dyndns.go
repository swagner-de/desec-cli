package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) DynDNSUpdate(hostname, ipv4, ipv6 string) error {
	params := url.Values{}
	if hostname != "" { params.Set("hostname", hostname) }
	if ipv4 != "" { params.Set("myipv4", ipv4) }
	if ipv6 != "" { params.Set("myipv6", ipv6) }
	reqURL := c.dynDNSURL + "/?" + params.Encode()
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil { return fmt.Errorf("creating request: %w", err) }
	req.Header.Set("Authorization", "Token "+c.token)
	resp, err := c.http.Do(req)
	if err != nil { return fmt.Errorf("executing request: %w", err) }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dyndns update failed (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}
