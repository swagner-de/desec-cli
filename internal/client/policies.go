package client

import (
	"encoding/json"
	"fmt"
)

func (c *Client) ListPolicies(tokenID string) ([]TokenPolicy, error) {
	data, err := c.doRequestPaginated(fmt.Sprintf("/auth/tokens/%s/policies/rrsets/", tokenID))
	if err != nil {
		return nil, err
	}
	var policies []TokenPolicy
	if err := json.Unmarshal(data, &policies); err != nil {
		return nil, fmt.Errorf("parsing policies: %w", err)
	}
	return policies, nil
}

func (c *Client) GetPolicy(tokenID, policyID string) (*TokenPolicy, error) {
	data, err := c.doRequest("GET", fmt.Sprintf("/auth/tokens/%s/policies/rrsets/%s/", tokenID, policyID), nil)
	if err != nil {
		return nil, err
	}
	var policy TokenPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("parsing policy: %w", err)
	}
	return &policy, nil
}

func (c *Client) CreatePolicy(tokenID string, create *TokenPolicyCreate) (*TokenPolicy, error) {
	data, err := c.doRequest("POST", fmt.Sprintf("/auth/tokens/%s/policies/rrsets/", tokenID), create)
	if err != nil {
		return nil, err
	}
	var policy TokenPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("parsing policy: %w", err)
	}
	return &policy, nil
}

func (c *Client) UpdatePolicy(tokenID, policyID string, update map[string]any) (*TokenPolicy, error) {
	data, err := c.doRequest("PATCH", fmt.Sprintf("/auth/tokens/%s/policies/rrsets/%s/", tokenID, policyID), update)
	if err != nil {
		return nil, err
	}
	var policy TokenPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("parsing policy: %w", err)
	}
	return &policy, nil
}

func (c *Client) DeletePolicy(tokenID, policyID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/auth/tokens/%s/policies/rrsets/%s/", tokenID, policyID), nil)
	return err
}
