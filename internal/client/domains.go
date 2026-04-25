package client

import (
	"encoding/json"
	"fmt"
)

func (c *Client) ListDomains() ([]Domain, error) {
	data, err := c.doRequestPaginated("/domains/")
	if err != nil {
		return nil, err
	}
	var domains []Domain
	if err := json.Unmarshal(data, &domains); err != nil {
		return nil, fmt.Errorf("parsing domains: %w", err)
	}
	return domains, nil
}

func (c *Client) GetDomain(name string) (*Domain, error) {
	data, err := c.doRequest("GET", fmt.Sprintf("/domains/%s/", name), nil)
	if err != nil {
		return nil, err
	}
	var domain Domain
	if err := json.Unmarshal(data, &domain); err != nil {
		return nil, fmt.Errorf("parsing domain: %w", err)
	}
	return &domain, nil
}

func (c *Client) CreateDomain(name string, zonefile string) (*Domain, error) {
	body := map[string]string{"name": name}
	if zonefile != "" {
		body["zonefile"] = zonefile
	}
	data, err := c.doRequest("POST", "/domains/", body)
	if err != nil {
		return nil, err
	}
	var domain Domain
	if err := json.Unmarshal(data, &domain); err != nil {
		return nil, fmt.Errorf("parsing domain: %w", err)
	}
	return &domain, nil
}

func (c *Client) DeleteDomain(name string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/domains/%s/", name), nil)
	return err
}

func (c *Client) ExportDomain(name string) (string, error) {
	data, err := c.doRequest("GET", fmt.Sprintf("/domains/%s/zonefile/", name), nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
