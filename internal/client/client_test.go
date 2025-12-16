// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	endpoint := "http://localhost:8096"
	accessToken := "test-access-token"

	client := NewClient(endpoint, accessToken)

	if client.endpoint != endpoint {
		t.Errorf("Expected endpoint %s, got %s", endpoint, client.endpoint)
	}

	if client.accessToken != accessToken {
		t.Errorf("Expected accessToken %s, got %s", accessToken, client.accessToken)
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestNewClient_trailingSlash(t *testing.T) {
	endpoint := "http://localhost:8096/"
	client := NewClient(endpoint, "token")

	expected := "http://localhost:8096"
	if client.endpoint != expected {
		t.Errorf("Expected endpoint %s, got %s", expected, client.endpoint)
	}
}

func TestNewClientWithAuth_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/Users/AuthenticateByName" {
			t.Errorf("Expected path /Users/AuthenticateByName, got %s", r.URL.Path)
		}

		// Check authorization header format
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("Expected Authorization header to be set")
		}
		// Should contain MediaBrowser format
		if !contains(auth, "MediaBrowser") {
			t.Errorf("Expected Authorization header to contain MediaBrowser, got %s", auth)
		}
		if !contains(auth, "Client=") {
			t.Errorf("Expected Authorization header to contain Client=, got %s", auth)
		}

		// Decode and check request body
		var authReq AuthenticateRequest
		if err := json.NewDecoder(r.Body).Decode(&authReq); err != nil {
			t.Errorf("Failed to decode auth request: %v", err)
			return
		}

		if authReq.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got %s", authReq.Username)
		}
		if authReq.Pw != "testpass" {
			t.Errorf("Expected password 'testpass', got %s", authReq.Pw)
		}

		// Return mock response
		resp := AuthenticateResponse{
			AccessToken: "returned-access-token",
			ServerId:    "server-123",
		}
		resp.User.Id = "user-456"
		resp.User.Name = "testuser"

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClientWithAuth(context.Background(), server.URL, "testuser", "testpass")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be returned")
	}

	if client.accessToken != "returned-access-token" {
		t.Errorf("Expected accessToken 'returned-access-token', got %s", client.accessToken)
	}

	if client.endpoint != server.URL {
		t.Errorf("Expected endpoint %s, got %s", server.URL, client.endpoint)
	}
}

func TestNewClientWithAuth_trailingSlash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AuthenticateResponse{
			AccessToken: "test-token",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClientWithAuth(context.Background(), server.URL+"/", "user", "pass")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Endpoint should have trailing slash removed
	if client.endpoint != server.URL {
		t.Errorf("Expected endpoint %s, got %s", server.URL, client.endpoint)
	}
}

func TestNewClientWithAuth_invalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Invalid username or password"))
	}))
	defer server.Close()

	client, err := NewClientWithAuth(context.Background(), server.URL, "baduser", "badpass")

	if err == nil {
		t.Error("Expected error for invalid credentials")
	}

	if client != nil {
		t.Error("Expected nil client for invalid credentials")
	}
}

func TestNewClientWithAuth_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	client, err := NewClientWithAuth(context.Background(), server.URL, "user", "pass")

	if err == nil {
		t.Error("Expected error for server error")
	}

	if client != nil {
		t.Error("Expected nil client for server error")
	}
}

func TestNewClientWithAuth_emptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return response with empty access token
		resp := AuthenticateResponse{
			AccessToken: "",
			ServerId:    "server-123",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClientWithAuth(context.Background(), server.URL, "user", "pass")

	if err == nil {
		t.Error("Expected error for empty access token")
	}

	if client != nil {
		t.Error("Expected nil client for empty access token")
	}
}

func TestNewClientWithAuth_malformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client, err := NewClientWithAuth(context.Background(), server.URL, "user", "pass")

	if err == nil {
		t.Error("Expected error for malformed JSON response")
	}

	if client != nil {
		t.Error("Expected nil client for malformed JSON")
	}
}

func TestNewClientWithAuthAndConfig_customConfig(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")

		resp := AuthenticateResponse{
			AccessToken: "test-token",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &ClientConfig{
		ClientName:    "CustomClient",
		DeviceName:    "CustomDevice",
		DeviceID:      "custom-device-id",
		ClientVersion: "2.0.0",
	}

	client, err := NewClientWithAuthAndConfig(context.Background(), server.URL, "user", "pass", config)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be returned")
	}

	// Check that custom config values were used in the header
	if !contains(receivedAuth, "CustomClient") {
		t.Errorf("Expected auth header to contain CustomClient, got %s", receivedAuth)
	}
	if !contains(receivedAuth, "CustomDevice") {
		t.Errorf("Expected auth header to contain CustomDevice, got %s", receivedAuth)
	}
	if !contains(receivedAuth, "custom-device-id") {
		t.Errorf("Expected auth header to contain custom-device-id, got %s", receivedAuth)
	}
	if !contains(receivedAuth, "2.0.0") {
		t.Errorf("Expected auth header to contain 2.0.0, got %s", receivedAuth)
	}
}

func TestNewClientWithAuthAndConfig_nilConfig(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")

		resp := AuthenticateResponse{
			AccessToken: "test-token",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClientWithAuthAndConfig(context.Background(), server.URL, "user", "pass", nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be returned")
	}

	// Check that default values were used
	if !contains(receivedAuth, DefaultClientName) {
		t.Errorf("Expected auth header to contain default client name %s, got %s", DefaultClientName, receivedAuth)
	}
	if !contains(receivedAuth, DefaultDeviceName) {
		t.Errorf("Expected auth header to contain default device name %s, got %s", DefaultDeviceName, receivedAuth)
	}
}

func TestNewClientWithAuthAndConfig_partialConfig(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")

		resp := AuthenticateResponse{
			AccessToken: "test-token",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Only set some config values
	config := &ClientConfig{
		ClientName: "PartialClient",
		// Other fields empty - should use defaults
	}

	client, err := NewClientWithAuthAndConfig(context.Background(), server.URL, "user", "pass", config)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be returned")
	}

	// Check custom value
	if !contains(receivedAuth, "PartialClient") {
		t.Errorf("Expected auth header to contain PartialClient, got %s", receivedAuth)
	}
	// Check default values are used for unset fields
	if !contains(receivedAuth, DefaultDeviceName) {
		t.Errorf("Expected auth header to contain default device name %s, got %s", DefaultDeviceName, receivedAuth)
	}
}

// Helper function for string contains.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGetKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/Auth/Keys" {
			t.Errorf("Expected path /Auth/Keys, got %s", r.URL.Path)
		}

		// Check authorization header
		auth := r.Header.Get("Authorization")
		expected := `MediaBrowser Token="test-api-key"`
		if auth != expected {
			t.Errorf("Expected Authorization header %q, got %q", expected, auth)
		}

		// Return mock response
		result := APIKeyQueryResult{
			Items: []APIKey{
				{
					AccessToken: "token-1",
					AppName:     "App One",
					DateCreated: "2024-01-01T00:00:00.0000000Z",
				},
				{
					AccessToken: "token-2",
					AppName:     "App Two",
					DateCreated: "2024-01-02T00:00:00.0000000Z",
				},
			},
			TotalRecordCount: 2,
			StartIndex:       0,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	result, err := client.GetKeys(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}

	if result.Items[0].AccessToken != "token-1" {
		t.Errorf("Expected access token 'token-1', got %s", result.Items[0].AccessToken)
	}

	if result.Items[1].AppName != "App Two" {
		t.Errorf("Expected app name 'App Two', got %s", result.Items[1].AppName)
	}

	if result.TotalRecordCount != 2 {
		t.Errorf("Expected total record count 2, got %d", result.TotalRecordCount)
	}
}

func TestGetKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := APIKeyQueryResult{
			Items: []APIKey{
				{
					AccessToken: "token-1",
					AppName:     "App One",
					DateCreated: "2024-01-01T00:00:00.0000000Z",
				},
				{
					AccessToken: "token-2",
					AppName:     "App Two",
					DateCreated: "2024-01-02T00:00:00.0000000Z",
				},
			},
			TotalRecordCount: 2,
			StartIndex:       0,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	key, err := client.GetKey(context.Background(), "token-2")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if key == nil {
		t.Fatal("Expected key to be returned")
	}

	if key.AccessToken != "token-2" {
		t.Errorf("Expected access token 'token-2', got %s", key.AccessToken)
	}

	if key.AppName != "App Two" {
		t.Errorf("Expected app name 'App Two', got %s", key.AppName)
	}
}

func TestGetKey_notFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := APIKeyQueryResult{
			Items: []APIKey{
				{
					AccessToken: "token-1",
					AppName:     "App One",
					DateCreated: "2024-01-01T00:00:00.0000000Z",
				},
			},
			TotalRecordCount: 1,
			StartIndex:       0,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	key, err := client.GetKey(context.Background(), "nonexistent-token")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if key != nil {
		t.Error("Expected nil key for nonexistent token")
	}
}

func TestCreateKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/Auth/Keys" {
			t.Errorf("Expected path /Auth/Keys, got %s", r.URL.Path)
		}

		// Check query parameter
		appName := r.URL.Query().Get("app")
		if appName != "My New App" {
			t.Errorf("Expected app name 'My New App' in query, got %s", appName)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.CreateKey(context.Background(), "My New App")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestCreateKey_withSpecialCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameter is URL encoded
		appName := r.URL.Query().Get("app")
		if appName != "My App & Test" {
			t.Errorf("Expected app name 'My App & Test' in query, got %s", appName)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.CreateKey(context.Background(), "My App & Test")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDeleteKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/Auth/Keys/token-to-delete" {
			t.Errorf("Expected path /Auth/Keys/token-to-delete, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteKey(context.Background(), "token-to-delete")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDeleteKey_withSpecialCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check path encoding for special characters
		expectedPath := "/Auth/Keys/token%2Fwith%2Fslashes"
		if r.URL.RawPath != "" && r.URL.RawPath != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.RawPath)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteKey(context.Background(), "token/with/slashes")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestFindKeyByAppName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := APIKeyQueryResult{
			Items: []APIKey{
				{
					AccessToken: "token-1",
					AppName:     "App One",
					DateCreated: "2024-01-01T00:00:00.0000000Z",
				},
				{
					AccessToken: "token-2",
					AppName:     "My Target App",
					DateCreated: "2024-01-02T00:00:00.0000000Z",
				},
				{
					AccessToken: "token-3",
					AppName:     "App Three",
					DateCreated: "2024-01-03T00:00:00.0000000Z",
				},
			},
			TotalRecordCount: 3,
			StartIndex:       0,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	key, err := client.FindKeyByAppName(context.Background(), "My Target App")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if key == nil {
		t.Fatal("Expected key to be returned")
	}

	if key.AccessToken != "token-2" {
		t.Errorf("Expected access token 'token-2', got %s", key.AccessToken)
	}

	if key.AppName != "My Target App" {
		t.Errorf("Expected app name 'My Target App', got %s", key.AppName)
	}
}

func TestFindKeyByAppName_notFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := APIKeyQueryResult{
			Items: []APIKey{
				{
					AccessToken: "token-1",
					AppName:     "App One",
					DateCreated: "2024-01-01T00:00:00.0000000Z",
				},
			},
			TotalRecordCount: 1,
			StartIndex:       0,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	key, err := client.FindKeyByAppName(context.Background(), "Nonexistent App")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if key != nil {
		t.Error("Expected nil key for nonexistent app name")
	}
}

func TestFindKeyByAppName_emptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := APIKeyQueryResult{
			Items:            []APIKey{},
			TotalRecordCount: 0,
			StartIndex:       0,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	key, err := client.FindKeyByAppName(context.Background(), "Any App")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if key != nil {
		t.Error("Expected nil key when no keys exist")
	}
}

func TestClient_errorHandling_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	// Test GetKeys error
	_, err := client.GetKeys(context.Background())
	if err == nil {
		t.Error("Expected error for 500 response on GetKeys")
	}

	// Test CreateKey error
	err = client.CreateKey(context.Background(), "test")
	if err == nil {
		t.Error("Expected error for 500 response on CreateKey")
	}

	// Test DeleteKey error
	err = client.DeleteKey(context.Background(), "test")
	if err == nil {
		t.Error("Expected error for 500 response on DeleteKey")
	}
}

func TestClient_errorHandling_unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "invalid-api-key")

	_, err := client.GetKeys(context.Background())
	if err == nil {
		t.Error("Expected error for 401 response")
	}
}

func TestClient_errorHandling_forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Forbidden"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	_, err := client.GetKeys(context.Background())
	if err == nil {
		t.Error("Expected error for 403 response")
	}
}

func TestClient_errorHandling_malformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	_, err := client.GetKeys(context.Background())
	if err == nil {
		t.Error("Expected error for malformed JSON response")
	}
}

func TestClient_contextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should not be reached if context is cancelled
		result := APIKeyQueryResult{Items: []APIKey{}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetKeys(ctx)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}
