// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestGetTokenIntrospect_Valid tests retrieving token info successfully
func TestGetTokenIntrospect_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/token/introspect") {
			t.Errorf("Expected /token/introspect in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := TokenIntrospectResponse{
			Authorization: Authorization{
				ID:        "auth123",
				CreatedOn: "2024-01-01T00:00:00Z",
				LastUsed:  "2024-06-15T12:00:00Z",
				GrantType: "authorization_code",
				RateLimit: 60,
				Scope:     "sites:read sites:write",
				AuthorizedTo: AuthorizedTo{
					SiteIDs:      []string{"site1", "site2"},
					WorkspaceIDs: []string{"ws1"},
					UserIDs:      []string{"user1"},
				},
			},
			Application: Application{
				ID:          "app123",
				Description: "Test App",
				Homepage:    "https://example.com",
				DisplayName: "My Test App",
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL for this test
	oldURL := getTokenIntrospectBaseURL
	getTokenIntrospectBaseURL = server.URL
	defer func() { getTokenIntrospectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetTokenIntrospect(ctx, client)
	if err != nil {
		t.Fatalf("GetTokenIntrospect failed: %v", err)
	}

	if result.Authorization.ID != "auth123" {
		t.Errorf("Expected auth ID 'auth123', got '%s'", result.Authorization.ID)
	}
	if result.Authorization.RateLimit != 60 {
		t.Errorf("Expected rate limit 60, got %d", result.Authorization.RateLimit)
	}
	if result.Authorization.Scope != "sites:read sites:write" {
		t.Errorf("Expected scope 'sites:read sites:write', got '%s'", result.Authorization.Scope)
	}
	if len(result.Authorization.AuthorizedTo.SiteIDs) != 2 {
		t.Errorf("Expected 2 site IDs, got %d", len(result.Authorization.AuthorizedTo.SiteIDs))
	}
	if result.Application.DisplayName != "My Test App" {
		t.Errorf("Expected app name 'My Test App', got '%s'", result.Application.DisplayName)
	}
}

// TestGetTokenIntrospect_Unauthorized tests 401 handling
func TestGetTokenIntrospect_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid token"))
	}))
	defer server.Close()

	oldURL := getTokenIntrospectBaseURL
	getTokenIntrospectBaseURL = server.URL
	defer func() { getTokenIntrospectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetTokenIntrospect(ctx, client)
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("Expected 'unauthorized' in error, got: %v", err)
	}
}

// TestGetTokenIntrospect_Forbidden tests 403 handling
func TestGetTokenIntrospect_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("access denied"))
	}))
	defer server.Close()

	oldURL := getTokenIntrospectBaseURL
	getTokenIntrospectBaseURL = server.URL
	defer func() { getTokenIntrospectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetTokenIntrospect(ctx, client)
	if err == nil {
		t.Error("Expected error for 403, got nil")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("Expected 'forbidden' in error, got: %v", err)
	}
}

// TestGetTokenIntrospect_ServerError tests 500 handling
func TestGetTokenIntrospect_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	oldURL := getTokenIntrospectBaseURL
	getTokenIntrospectBaseURL = server.URL
	defer func() { getTokenIntrospectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetTokenIntrospect(ctx, client)
	if err == nil {
		t.Error("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' in error, got: %v", err)
	}
}

// TestGetTokenIntrospect_EmptyResponse tests handling of empty authorization
func TestGetTokenIntrospect_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := TokenIntrospectResponse{
			Authorization: Authorization{
				ID:           "auth123",
				AuthorizedTo: AuthorizedTo{}, // Empty authorized resources
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getTokenIntrospectBaseURL
	getTokenIntrospectBaseURL = server.URL
	defer func() { getTokenIntrospectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetTokenIntrospect(ctx, client)
	if err != nil {
		t.Fatalf("GetTokenIntrospect failed: %v", err)
	}

	if result.Authorization.ID != "auth123" {
		t.Errorf("Expected auth ID 'auth123', got '%s'", result.Authorization.ID)
	}
	// Empty slices should be nil in Go
	if result.Authorization.AuthorizedTo.SiteIDs != nil && len(result.Authorization.AuthorizedTo.SiteIDs) != 0 {
		t.Errorf("Expected nil or empty site IDs, got %v", result.Authorization.AuthorizedTo.SiteIDs)
	}
}

// TestGetAuthorizedBy_Valid tests retrieving authorized user info successfully
func TestGetAuthorizedBy_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/token/authorized_by") {
			t.Errorf("Expected /token/authorized_by in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := AuthorizedByResponse{
			ID:        "user123",
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getAuthorizedByBaseURL
	getAuthorizedByBaseURL = server.URL
	defer func() { getAuthorizedByBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetAuthorizedBy(ctx, client)
	if err != nil {
		t.Fatalf("GetAuthorizedBy failed: %v", err)
	}

	if result.ID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", result.ID)
	}
	if result.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", result.Email)
	}
	if result.FirstName != "John" {
		t.Errorf("Expected first name 'John', got '%s'", result.FirstName)
	}
	if result.LastName != "Doe" {
		t.Errorf("Expected last name 'Doe', got '%s'", result.LastName)
	}
}

// TestGetAuthorizedBy_Unauthorized tests 401 handling for authorized_by
func TestGetAuthorizedBy_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid token"))
	}))
	defer server.Close()

	oldURL := getAuthorizedByBaseURL
	getAuthorizedByBaseURL = server.URL
	defer func() { getAuthorizedByBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetAuthorizedBy(ctx, client)
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("Expected 'unauthorized' in error, got: %v", err)
	}
}

// TestGetAuthorizedBy_Forbidden tests 403 handling for authorized_by
func TestGetAuthorizedBy_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("missing authorized_user:read scope"))
	}))
	defer server.Close()

	oldURL := getAuthorizedByBaseURL
	getAuthorizedByBaseURL = server.URL
	defer func() { getAuthorizedByBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetAuthorizedBy(ctx, client)
	if err == nil {
		t.Error("Expected error for 403, got nil")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("Expected 'forbidden' in error, got: %v", err)
	}
}

// TestGetAuthorizedBy_MinimalResponse tests handling response with only required fields
func TestGetAuthorizedBy_MinimalResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Response with only ID and email (firstName/lastName are optional)
		response := AuthorizedByResponse{
			ID:    "user123",
			Email: "test@example.com",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getAuthorizedByBaseURL
	getAuthorizedByBaseURL = server.URL
	defer func() { getAuthorizedByBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetAuthorizedBy(ctx, client)
	if err != nil {
		t.Fatalf("GetAuthorizedBy failed: %v", err)
	}

	if result.ID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", result.ID)
	}
	if result.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", result.Email)
	}
	if result.FirstName != "" {
		t.Errorf("Expected empty first name, got '%s'", result.FirstName)
	}
	if result.LastName != "" {
		t.Errorf("Expected empty last name, got '%s'", result.LastName)
	}
}

// TestHandleTokenError_Messages tests that error messages are actionable
func TestHandleTokenError_Messages(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		contains   []string
	}{
		{
			name:       "401 unauthorized",
			statusCode: 401,
			body:       "invalid token",
			contains:   []string{"unauthorized", "invalid", "expired"},
		},
		{
			name:       "403 forbidden",
			statusCode: 403,
			body:       "access denied",
			contains:   []string{"forbidden", "permission", "scope"},
		},
		{
			name:       "500 server error",
			statusCode: 500,
			body:       "internal error",
			contains:   []string{"server error", "temporary", "Webflow"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleTokenError(tt.statusCode, []byte(tt.body))
			errMsg := err.Error()

			for _, expected := range tt.contains {
				if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(expected)) {
					t.Errorf("Error message missing '%s'. Got: %s", expected, errMsg)
				}
			}
		})
	}
}

// TestGetTokenIntrospect_ContextCancellation tests context cancellation
func TestGetTokenIntrospect_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &http.Client{Timeout: 30 * time.Second}

	_, err := GetTokenIntrospect(ctx, client)
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("Expected 'context cancelled' in error, got: %v", err)
	}
}

// TestGetAuthorizedBy_ContextCancellation tests context cancellation
func TestGetAuthorizedBy_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &http.Client{Timeout: 30 * time.Second}

	_, err := GetAuthorizedBy(ctx, client)
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("Expected 'context cancelled' in error, got: %v", err)
	}
}
