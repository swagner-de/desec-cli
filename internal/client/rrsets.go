package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) ListRRsets(domain, rrType, subname string) ([]RRset, error) {
	path := fmt.Sprintf("/domains/%s/rrsets/", domain)
	params := url.Values{}
	if rrType != "" { params.Set("type", rrType) }
	if subname != "" { params.Set("subname", subname) }
	if len(params) > 0 { path += "?" + params.Encode() }
	data, err := c.doRequestPaginated(path)
	if err != nil { return nil, err }
	var rrsets []RRset
	if err := json.Unmarshal(data, &rrsets); err != nil { return nil, fmt.Errorf("parsing rrsets: %w", err) }
	return rrsets, nil
}

func (c *Client) GetRRset(domain, subname, rrType string) (*RRset, error) {
	if subname == "" { subname = "@" }
	path := fmt.Sprintf("/domains/%s/rrsets/%s/%s/", domain, subname, rrType)
	data, err := c.doRequest("GET", path, nil)
	if err != nil { return nil, err }
	var rrset RRset
	if err := json.Unmarshal(data, &rrset); err != nil { return nil, fmt.Errorf("parsing rrset: %w", err) }
	return &rrset, nil
}

func (c *Client) CreateRRset(domain string, rrset *RRsetCreate) (*RRset, error) {
	path := fmt.Sprintf("/domains/%s/rrsets/", domain)
	data, err := c.doRequest("POST", path, rrset)
	if err != nil { return nil, err }
	var result RRset
	if err := json.Unmarshal(data, &result); err != nil { return nil, fmt.Errorf("parsing rrset: %w", err) }
	return &result, nil
}

func (c *Client) UpdateRRset(domain, subname, rrType string, update map[string]any) (*RRset, error) {
	if subname == "" { subname = "@" }
	path := fmt.Sprintf("/domains/%s/rrsets/%s/%s/", domain, subname, rrType)
	data, err := c.doRequest("PATCH", path, update)
	if err != nil { return nil, err }
	var result RRset
	if err := json.Unmarshal(data, &result); err != nil { return nil, fmt.Errorf("parsing rrset: %w", err) }
	return &result, nil
}

func (c *Client) DeleteRRset(domain, subname, rrType string) error {
	if subname == "" { subname = "@" }
	path := fmt.Sprintf("/domains/%s/rrsets/%s/%s/", domain, subname, rrType)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

func (c *Client) BulkRRsets(domain string, rrsets []RRsetCreate) ([]RRset, error) {
	path := fmt.Sprintf("/domains/%s/rrsets/", domain)
	data, err := c.doRequest("PATCH", path, rrsets)
	if err != nil { return nil, err }
	var result []RRset
	if err := json.Unmarshal(data, &result); err != nil { return nil, fmt.Errorf("parsing rrsets: %w", err) }
	return result, nil
}
