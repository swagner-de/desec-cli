package client

import (
	"encoding/json"
	"fmt"
)

func (c *Client) ListTokens() ([]Token, error) {
	data, err := c.doRequestPaginated("/auth/tokens/")
	if err != nil {
		return nil, err
	}
	var tokens []Token
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("parsing tokens: %w", err)
	}
	return tokens, nil
}

func (c *Client) GetToken(id string) (*Token, error) {
	data, err := c.doRequest("GET", fmt.Sprintf("/auth/tokens/%s/", id), nil)
	if err != nil {
		return nil, err
	}
	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	return &token, nil
}

func (c *Client) CreateToken(create *TokenCreate) (*Token, error) {
	data, err := c.doRequest("POST", "/auth/tokens/", create)
	if err != nil {
		return nil, err
	}
	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	return &token, nil
}

func (c *Client) UpdateToken(id string, update map[string]any) (*Token, error) {
	data, err := c.doRequest("PATCH", fmt.Sprintf("/auth/tokens/%s/", id), update)
	if err != nil {
		return nil, err
	}
	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	return &token, nil
}

func (c *Client) DeleteToken(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/auth/tokens/%s/", id), nil)
	return err
}
