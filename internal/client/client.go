// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// Default client identification for Jellyfin
	DefaultClientName    = "Terraform"
	DefaultDeviceName    = "Terraform Provider"
	DefaultDeviceID      = "terraform-provider-jellyfin"
	DefaultClientVersion = "1.0.0"
)

// Client is a Jellyfin API client.
type Client struct {
	endpoint    string
	accessToken string
	httpClient  *http.Client
}

// ClientConfig holds configuration for creating a new client.
type ClientConfig struct {
	Endpoint      string
	ClientName    string
	DeviceName    string
	DeviceID      string
	ClientVersion string
}

// AuthenticateRequest represents the request body for authentication.
type AuthenticateRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

// AuthenticateResponse represents the response from AuthenticateByName.
type AuthenticateResponse struct {
	AccessToken string `json:"AccessToken"`
	ServerId    string `json:"ServerId"`
	User        struct {
		Id   string `json:"Id"`
		Name string `json:"Name"`
	} `json:"User"`
	SessionInfo struct {
		Id string `json:"Id"`
	} `json:"SessionInfo"`
}

// APIKey represents a Jellyfin API key.
type APIKey struct {
	AccessToken string `json:"AccessToken"`
	AppName     string `json:"AppName"`
	DateCreated string `json:"DateCreated"`
}

// APIKeyQueryResult represents the response from GetKeys.
type APIKeyQueryResult struct {
	Items            []APIKey `json:"Items"`
	TotalRecordCount int      `json:"TotalRecordCount"`
	StartIndex       int      `json:"StartIndex"`
}

// NewClient creates a new Jellyfin API client with a pre-existing access token.
func NewClient(endpoint, accessToken string) *Client {
	return &Client{
		endpoint:    strings.TrimSuffix(endpoint, "/"),
		accessToken: accessToken,
		httpClient:  http.DefaultClient,
	}
}

// NewClientWithAuth creates a new Jellyfin API client by authenticating with username and password.
func NewClientWithAuth(ctx context.Context, endpoint, username, password string) (*Client, error) {
	return NewClientWithAuthAndConfig(ctx, endpoint, username, password, nil)
}

// NewClientWithAuthAndConfig creates a new Jellyfin API client with custom client configuration.
func NewClientWithAuthAndConfig(ctx context.Context, endpoint, username, password string, config *ClientConfig) (*Client, error) {
	endpoint = strings.TrimSuffix(endpoint, "/")

	// Use defaults if config not provided
	clientName := DefaultClientName
	deviceName := DefaultDeviceName
	deviceID := DefaultDeviceID
	clientVersion := DefaultClientVersion

	if config != nil {
		if config.ClientName != "" {
			clientName = config.ClientName
		}
		if config.DeviceName != "" {
			deviceName = config.DeviceName
		}
		if config.DeviceID != "" {
			deviceID = config.DeviceID
		}
		if config.ClientVersion != "" {
			clientVersion = config.ClientVersion
		}
	}

	// Create authentication request
	authReq := AuthenticateRequest{
		Username: username,
		Pw:       password,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/Users/AuthenticateByName", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	// Set headers for unauthenticated request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf(
		`MediaBrowser Client="%s", Device="%s", DeviceId="%s", Version="%s"`,
		clientName, deviceName, deviceID, clientVersion,
	))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var authResp AuthenticateResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return nil, fmt.Errorf("authentication succeeded but no access token returned")
	}

	return &Client{
		endpoint:    endpoint,
		accessToken: authResp.AccessToken,
		httpClient:  http.DefaultClient,
	}, nil
}

// doRequest makes an HTTP request to the Jellyfin API.
func (c *Client) doRequest(ctx context.Context, method, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use MediaBrowser authorization header format with token
	req.Header.Set("Authorization", fmt.Sprintf(`MediaBrowser Token="%s"`, c.accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// GetKeys retrieves all API keys.
func (c *Client) GetKeys(ctx context.Context) (*APIKeyQueryResult, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/Auth/Keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result APIKeyQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetKey retrieves a specific API key by its access token.
func (c *Client) GetKey(ctx context.Context, accessToken string) (*APIKey, error) {
	result, err := c.GetKeys(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range result.Items {
		if key.AccessToken == accessToken {
			return &key, nil
		}
	}

	return nil, nil // Not found
}

// CreateKey creates a new API key.
func (c *Client) CreateKey(ctx context.Context, appName string) error {
	path := fmt.Sprintf("/Auth/Keys?app=%s", url.QueryEscape(appName))

	resp, err := c.doRequest(ctx, http.MethodPost, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteKey deletes an API key by its access token.
func (c *Client) DeleteKey(ctx context.Context, accessToken string) error {
	path := fmt.Sprintf("/Auth/Keys/%s", url.PathEscape(accessToken))

	resp, err := c.doRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// FindKeyByAppName finds an API key by its application name.
// Since the Create API doesn't return the token, we need to find it by comparing before/after state.
func (c *Client) FindKeyByAppName(ctx context.Context, appName string) (*APIKey, error) {
	result, err := c.GetKeys(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range result.Items {
		if key.AppName == appName {
			return &key, nil
		}
	}

	return nil, nil // Not found
}
