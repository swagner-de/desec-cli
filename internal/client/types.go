package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Domain struct {
	Name       string    `json:"name"`
	Created    time.Time `json:"created"`
	Published  time.Time `json:"published"`
	MinimumTTL int       `json:"minimum_ttl"`
	Keys       []DNSKey  `json:"keys,omitempty"`
}

type DNSKey struct {
	DNSKey string   `json:"dnskey"`
	DS     []string `json:"ds"`
	Flags  int      `json:"flags"`
}

type RRset struct {
	Domain  string   `json:"domain,omitempty"`
	Subname string   `json:"subname"`
	Name    string   `json:"name,omitempty"`
	Type    string   `json:"type"`
	Records []string `json:"records"`
	TTL     int      `json:"ttl"`
	Created string   `json:"created,omitempty"`
	Touched string   `json:"touched,omitempty"`
}

type RRsetCreate struct {
	Subname string   `json:"subname"`
	Type    string   `json:"type"`
	Records []string `json:"records"`
	TTL     int      `json:"ttl"`
}

type Token struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Token            string   `json:"token,omitempty"`
	Created          string   `json:"created"`
	LastUsed         string   `json:"last_used"`
	PermManageTokens bool     `json:"perm_manage_tokens"`
	PermCreateDomain bool     `json:"perm_create_domain"`
	PermDeleteDomain bool     `json:"perm_delete_domain"`
	AllowedSubnets   []string `json:"allowed_subnets"`
	MaxAge           string   `json:"max_age"`
	MaxUnusedPeriod  string   `json:"max_unused_period"`
	IsValid          bool     `json:"is_valid"`
}

type TokenCreate struct {
	Name             string   `json:"name,omitempty"`
	PermManageTokens *bool    `json:"perm_manage_tokens,omitempty"`
	PermCreateDomain *bool    `json:"perm_create_domain,omitempty"`
	PermDeleteDomain *bool    `json:"perm_delete_domain,omitempty"`
	AllowedSubnets   []string `json:"allowed_subnets,omitempty"`
	MaxAge           string   `json:"max_age,omitempty"`
	MaxUnusedPeriod  string   `json:"max_unused_period,omitempty"`
}

type TokenPolicy struct {
	ID        string  `json:"id"`
	Domain    *string `json:"domain"`
	Subname   *string `json:"subname"`
	Type      *string `json:"type"`
	PermWrite bool    `json:"perm_write"`
}

type TokenPolicyCreate struct {
	Domain    *string `json:"domain"`
	Subname   *string `json:"subname"`
	Type      *string `json:"type"`
	PermWrite bool    `json:"perm_write"`
}

type APIError struct {
	StatusCode int
	Detail     string
	Body       string
}

func (e *APIError) UnmarshalJSON(data []byte) error {
	// Try {"detail": "..."} first
	var detail struct {
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(data, &detail); err == nil && detail.Detail != "" {
		e.Detail = detail.Detail
		return nil
	}

	// Try {"field": ["error", ...], ...} format
	var fields map[string][]string
	if err := json.Unmarshal(data, &fields); err == nil && len(fields) > 0 {
		var parts []string
		for field, msgs := range fields {
			parts = append(parts, fmt.Sprintf("%s: %s", field, strings.Join(msgs, "; ")))
		}
		e.Detail = strings.Join(parts, ", ")
		return nil
	}

	e.Body = string(data)
	return nil
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%d: %s", e.StatusCode, e.Detail)
	}
	if e.Body != "" {
		return fmt.Sprintf("%d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("%d: API error", e.StatusCode)
}
