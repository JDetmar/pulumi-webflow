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

// TestValidateWebhookID_Valid tests valid webhook IDs
func TestValidateWebhookID_Valid(t *testing.T) {
	tests := []struct {
		name      string
		webhookID string
	}{
		{"valid lowercase hex", "5f0c8c9e1c9d440000e8d8c3"},
		{"another valid ID", "507f1f77bcf86cd799439011"},
		{"all zeros", "000000000000000000000000"},
		{"all fs", "ffffffffffffffffffffffff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookID(tt.webhookID)
			if err != nil {
				t.Errorf("ValidateWebhookID(%q) = %v, want nil", tt.webhookID, err)
			}
		})
	}
}

// TestValidateWebhookID_Empty tests empty webhook ID
func TestValidateWebhookID_Empty(t *testing.T) {
	err := ValidateWebhookID("")
	if err == nil {
		t.Error("ValidateWebhookID(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidateWebhookID_InvalidFormat tests invalid webhook ID formats
func TestValidateWebhookID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name      string
		webhookID string
	}{
		{"too short", "5f0c8c9e1c9d"},
		{"too long", "5f0c8c9e1c9d440000e8d8c3extra"},
		{"uppercase", "5F0C8C9E1C9D440000E8D8C3"},
		{"invalid chars", "5f0c8c9e1c9d440000e8d8cg"},
		{"with spaces", "5f0c8c9e 1c9d440000e8d8c3"},
		{"with hyphens", "5f0c8c9e-1c9d-4400-00e8-d8c3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookID(tt.webhookID)
			if err == nil {
				t.Errorf("ValidateWebhookID(%q) = nil, want error", tt.webhookID)
			}
			if !strings.Contains(err.Error(), "invalid format") {
				t.Errorf("Expected error to mention 'invalid format', got: %v", err)
			}
		})
	}
}

// TestValidateWebhookURL_Valid tests valid webhook URLs
func TestValidateWebhookURL_Valid(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"simple https", "https://example.com/webhook"},
		{"with path", "https://api.example.com/webhooks/webflow"},
		{"with port", "https://example.com:8443/webhook"},
		{"with query", "https://example.com/webhook?source=webflow"},
		{"subdomain", "https://webhooks.example.com/webflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookURL(tt.url)
			if err != nil {
				t.Errorf("ValidateWebhookURL(%q) = %v, want nil", tt.url, err)
			}
		})
	}
}

// TestValidateWebhookURL_Empty tests empty webhook URL
func TestValidateWebhookURL_Empty(t *testing.T) {
	err := ValidateWebhookURL("")
	if err == nil {
		t.Error("ValidateWebhookURL(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidateWebhookURL_NotHTTPS tests URLs that don't use HTTPS
func TestValidateWebhookURL_NotHTTPS(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"http", "http://example.com/webhook"},
		{"no protocol", "example.com/webhook"},
		{"ftp", "ftp://example.com/webhook"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookURL(tt.url)
			if err == nil {
				t.Errorf("ValidateWebhookURL(%q) = nil, want error", tt.url)
			}
			if !strings.Contains(err.Error(), "HTTPS") {
				t.Errorf("Expected error to mention 'HTTPS', got: %v", err)
			}
		})
	}
}

// TestValidateWebhookURL_Invalid tests invalid URL formats
func TestValidateWebhookURL_Invalid(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"no domain", "https://"},
		{"no tld", "https://example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookURL(tt.url)
			if err == nil {
				t.Errorf("ValidateWebhookURL(%q) = nil, want error", tt.url)
			}
		})
	}
}

// TestValidateTriggerType_Valid tests valid trigger types
func TestValidateTriggerType_Valid(t *testing.T) {
	tests := []struct {
		name        string
		triggerType string
	}{
		{"form_submission", "form_submission"},
		{"site_publish", "site_publish"},
		{"page_created", "page_created"},
		{"page_metadata_updated", "page_metadata_updated"},
		{"page_deleted", "page_deleted"},
		{"ecomm_new_order", "ecomm_new_order"},
		{"ecomm_order_changed", "ecomm_order_changed"},
		{"ecomm_inventory_changed", "ecomm_inventory_changed"},
		{"memberships_user_account_added", "memberships_user_account_added"},
		{"memberships_user_account_updated", "memberships_user_account_updated"},
		{"memberships_user_account_deleted", "memberships_user_account_deleted"},
		{"collection_item_created", "collection_item_created"},
		{"collection_item_changed", "collection_item_changed"},
		{"collection_item_deleted", "collection_item_deleted"},
		{"collection_item_unpublished", "collection_item_unpublished"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTriggerType(tt.triggerType)
			if err != nil {
				t.Errorf("ValidateTriggerType(%q) = %v, want nil", tt.triggerType, err)
			}
		})
	}
}

// TestValidateTriggerType_Empty tests empty trigger type
func TestValidateTriggerType_Empty(t *testing.T) {
	err := ValidateTriggerType("")
	if err == nil {
		t.Error("ValidateTriggerType(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidateTriggerType_Invalid tests invalid trigger types
func TestValidateTriggerType_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		triggerType string
	}{
		{"invalid type", "invalid_trigger"},
		{"typo", "form_submision"},
		{"uppercase", "FORM_SUBMISSION"},
		{"spaces", "form submission"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTriggerType(tt.triggerType)
			if err == nil {
				t.Errorf("ValidateTriggerType(%q) = nil, want error", tt.triggerType)
			}
			if !strings.Contains(err.Error(), "not a valid") {
				t.Errorf("Expected error to mention 'not a valid', got: %v", err)
			}
		})
	}
}

// TestGenerateWebhookResourceID tests resource ID generation
func TestGenerateWebhookResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	webhookID := "507f1f77bcf86cd799439011"

	resourceID := GenerateWebhookResourceID(siteID, webhookID)
	expected := "5f0c8c9e1c9d440000e8d8c3/webhooks/507f1f77bcf86cd799439011"

	if resourceID != expected {
		t.Errorf("GenerateWebhookResourceID() = %q, want %q", resourceID, expected)
	}
}

// TestExtractIDsFromWebhookResourceID_Valid tests extracting IDs from valid resource ID
func TestExtractIDsFromWebhookResourceID_Valid(t *testing.T) {
	resourceID := "5f0c8c9e1c9d440000e8d8c3/webhooks/507f1f77bcf86cd799439011"

	siteID, webhookID, err := ExtractIDsFromWebhookResourceID(resourceID)
	if err != nil {
		t.Errorf("ExtractIDsFromWebhookResourceID() error = %v, want nil", err)
	}
	if siteID != "5f0c8c9e1c9d440000e8d8c3" {
		t.Errorf("ExtractIDsFromWebhookResourceID() siteID = %q, want %q", siteID, "5f0c8c9e1c9d440000e8d8c3")
	}
	if webhookID != "507f1f77bcf86cd799439011" {
		t.Errorf("ExtractIDsFromWebhookResourceID() webhookID = %q, want %q", webhookID, "507f1f77bcf86cd799439011")
	}
}

// TestExtractIDsFromWebhookResourceID_Empty tests empty resource ID
func TestExtractIDsFromWebhookResourceID_Empty(t *testing.T) {
	_, _, err := ExtractIDsFromWebhookResourceID("")
	if err == nil {
		t.Error("ExtractIDsFromWebhookResourceID(\"\") error = nil, want error")
	}
}

// TestExtractIDsFromWebhookResourceID_InvalidFormat tests invalid format
func TestExtractIDsFromWebhookResourceID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{"missing webhooks part", "5f0c8c9e1c9d440000e8d8c3/507f1f77bcf86cd799439011"},
		{"wrong middle part", "5f0c8c9e1c9d440000e8d8c3/redirects/507f1f77bcf86cd799439011"},
		{"too few parts", "5f0c8c9e1c9d440000e8d8c3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ExtractIDsFromWebhookResourceID(tt.resourceID)
			if err == nil {
				t.Errorf("ExtractIDsFromWebhookResourceID(%q) error = nil, want error", tt.resourceID)
			}
		})
	}
}

// TestGetWebhooks_Valid tests retrieving webhooks successfully
func TestGetWebhooks_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/webhooks") {
			t.Errorf("Expected /webhooks in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := WebhooksListResponse{
			Webhooks: []WebhookResponse{
				{
					ID:          "webhook1",
					TriggerType: "form_submission",
					URL:         "https://example.com/webhook",
					SiteID:      "5f0c8c9e1c9d440000e8d8c3",
					CreatedOn:   "2024-01-01T00:00:00Z",
				},
				{
					ID:          "webhook2",
					TriggerType: "site_publish",
					URL:         "https://example.com/publish",
					SiteID:      "5f0c8c9e1c9d440000e8d8c3",
					CreatedOn:   "2024-01-02T00:00:00Z",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL for this test
	oldURL := getWebhooksBaseURL
	getWebhooksBaseURL = server.URL
	defer func() { getWebhooksBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetWebhooks(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("GetWebhooks failed: %v", err)
	}

	if len(result.Webhooks) != 2 {
		t.Errorf("Expected 2 webhooks, got %d", len(result.Webhooks))
	}
	if result.Webhooks[0].ID != "webhook1" {
		t.Errorf("Expected webhook1, got %s", result.Webhooks[0].ID)
	}
	if result.Webhooks[0].TriggerType != "form_submission" {
		t.Errorf("Expected form_submission, got %s", result.Webhooks[0].TriggerType)
	}
}

// TestGetWebhooks_NotFound tests 404 handling
func TestGetWebhooks_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("site not found"))
	}))
	defer server.Close()

	oldURL := getWebhooksBaseURL
	getWebhooksBaseURL = server.URL
	defer func() { getWebhooksBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetWebhooks(ctx, client, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestGetWebhook_Valid tests retrieving a single webhook successfully
func TestGetWebhook_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := WebhookResponse{
			ID:          "webhook1",
			TriggerType: "form_submission",
			URL:         "https://example.com/webhook",
			SiteID:      "5f0c8c9e1c9d440000e8d8c3",
			CreatedOn:   "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getWebhookBaseURL
	getWebhookBaseURL = server.URL
	defer func() { getWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetWebhook(ctx, client, "webhook1")
	if err != nil {
		t.Fatalf("GetWebhook failed: %v", err)
	}

	if result.ID != "webhook1" {
		t.Errorf("Expected ID webhook1, got %s", result.ID)
	}
	if result.TriggerType != "form_submission" {
		t.Errorf("Expected triggerType form_submission, got %s", result.TriggerType)
	}
}

// TestPostWebhook_Valid tests creating a webhook successfully
func TestPostWebhook_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req WebhookRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.TriggerType != "form_submission" {
			t.Errorf("Expected triggerType form_submission, got %s", req.TriggerType)
		}
		if req.URL != "https://example.com/webhook" {
			t.Errorf("Expected url https://example.com/webhook, got %s", req.URL)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := WebhookResponse{
			ID:          "new-webhook-1",
			TriggerType: "form_submission",
			URL:         "https://example.com/webhook",
			SiteID:      "5f0c8c9e1c9d440000e8d8c3",
			CreatedOn:   "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postWebhookBaseURL
	postWebhookBaseURL = server.URL
	defer func() { postWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := PostWebhook(
		ctx, client, "5f0c8c9e1c9d440000e8d8c3", "form_submission",
		"https://example.com/webhook", nil,
	)
	if err != nil {
		t.Fatalf("PostWebhook failed: %v", err)
	}

	if result.ID != "new-webhook-1" {
		t.Errorf("Expected ID new-webhook-1, got %s", result.ID)
	}
	if result.TriggerType != "form_submission" {
		t.Errorf("Expected triggerType form_submission, got %s", result.TriggerType)
	}
}

// TestPostWebhook_WithFilter tests creating a webhook with filter
func TestPostWebhook_WithFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req WebhookRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Filter == nil {
			t.Error("Expected filter to be set")
		}
		if req.Filter["collectionId"] != "test-collection" {
			t.Errorf("Expected filter.collectionId test-collection, got %v", req.Filter["collectionId"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := WebhookResponse{
			ID:          "new-webhook-1",
			TriggerType: "collection_item_created",
			URL:         "https://example.com/webhook",
			SiteID:      "5f0c8c9e1c9d440000e8d8c3",
			CreatedOn:   "2024-01-01T00:00:00Z",
			Filter:      req.Filter,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postWebhookBaseURL
	postWebhookBaseURL = server.URL
	defer func() { postWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	filter := map[string]interface{}{
		"collectionId": "test-collection",
	}

	result, err := PostWebhook(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "collection_item_created", "https://example.com/webhook", filter)
	if err != nil {
		t.Fatalf("PostWebhook failed: %v", err)
	}

	if result.Filter["collectionId"] != "test-collection" {
		t.Errorf("Expected filter.collectionId test-collection, got %v", result.Filter["collectionId"])
	}
}

// TestPostWebhook_ValidationError tests 400 handling
func TestPostWebhook_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid webhook configuration"))
	}))
	defer server.Close()

	oldURL := postWebhookBaseURL
	postWebhookBaseURL = server.URL
	defer func() { postWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := PostWebhook(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "invalid", "https://example.com/webhook", nil)
	if err == nil {
		t.Error("Expected error for 400, got nil")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("Expected 'bad request' in error, got: %v", err)
	}
}

// TestDeleteWebhook_Valid tests deleting a webhook successfully
func TestDeleteWebhook_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	oldURL := deleteWebhookBaseURL
	deleteWebhookBaseURL = server.URL
	defer func() { deleteWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteWebhook(ctx, client, "webhook1")
	if err != nil {
		t.Fatalf("DeleteWebhook failed: %v", err)
	}
}

// TestDeleteWebhook_NotFound_Idempotent tests that 404 on delete is treated as success (idempotent)
func TestDeleteWebhook_NotFound_Idempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("webhook not found"))
	}))
	defer server.Close()

	oldURL := deleteWebhookBaseURL
	deleteWebhookBaseURL = server.URL
	defer func() { deleteWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteWebhook(ctx, client, "nonexistent")
	if err != nil {
		t.Errorf("DeleteWebhook should handle 404 as success (idempotent), got error: %v", err)
	}
}

// TestDeleteWebhook_ServerError tests error handling
func TestDeleteWebhook_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	oldURL := deleteWebhookBaseURL
	deleteWebhookBaseURL = server.URL
	defer func() { deleteWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteWebhook(ctx, client, "webhook1")
	if err == nil {
		t.Error("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' in error, got: %v", err)
	}
}

// TestErrorMessagesAreActionable verifies error messages contain guidance
func TestWebhookErrorMessagesAreActionable(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		contains []string
	}{
		{
			"ValidateWebhookID empty",
			func() error { return ValidateWebhookID("") },
			[]string{"required", "24-character"},
		},
		{
			"ValidateWebhookURL empty",
			func() error { return ValidateWebhookURL("") },
			[]string{"required", "HTTPS"},
		},
		{
			"ValidateWebhookURL not HTTPS",
			func() error { return ValidateWebhookURL("http://example.com") },
			[]string{"HTTPS", "security"},
		},
		{
			"ValidateTriggerType empty",
			func() error { return ValidateTriggerType("") },
			[]string{"required", "Valid trigger types"},
		},
		{
			"ValidateTriggerType invalid",
			func() error { return ValidateTriggerType("invalid") },
			[]string{"not a valid", "form_submission"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()
			if err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
				return
			}

			errMsg := err.Error()
			for _, expectedStr := range tt.contains {
				if !strings.Contains(errMsg, expectedStr) {
					t.Errorf("%s: error message missing %q. Got: %s", tt.name, expectedStr, errMsg)
				}
			}
		})
	}
}

// TestGetWebhooks_RateLimited tests 429 rate limiting with retry
func TestGetWebhooks_RateLimited(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.Header().Set("Retry-After", "0") // Use 0 seconds for fast test
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		// Succeed on second attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := WebhooksListResponse{
			Webhooks: []WebhookResponse{
				{
					ID:          "webhook1",
					TriggerType: "form_submission",
					URL:         "https://example.com/webhook",
					SiteID:      "5f0c8c9e1c9d440000e8d8c3",
					CreatedOn:   "2024-01-01T00:00:00Z",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getWebhooksBaseURL
	getWebhooksBaseURL = server.URL
	defer func() { getWebhooksBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetWebhooks(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("GetWebhooks should succeed after rate limit retry, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts (1 rate limited + 1 success), got %d", attempts)
	}
	if len(result.Webhooks) != 1 {
		t.Errorf("Expected 1 webhook, got %d", len(result.Webhooks))
	}
}

// TestPostWebhook_RateLimited tests 429 rate limiting with retry
func TestPostWebhook_RateLimited(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		// Succeed on second attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := WebhookResponse{
			ID:          "newwebhook",
			TriggerType: "form_submission",
			URL:         "https://example.com/webhook",
			SiteID:      "5f0c8c9e1c9d440000e8d8c3",
			CreatedOn:   "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postWebhookBaseURL
	postWebhookBaseURL = server.URL
	defer func() { postWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := PostWebhook(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "form_submission", "https://example.com/webhook", nil)
	if err != nil {
		t.Fatalf("PostWebhook should succeed after rate limit retry, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts (1 rate limited + 1 success), got %d", attempts)
	}
	if result.ID != "newwebhook" {
		t.Errorf("Expected ID newwebhook, got %s", result.ID)
	}
}

// TestDeleteWebhook_RateLimited tests 429 rate limiting with retry
func TestDeleteWebhook_RateLimited(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		// Succeed on second attempt
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	oldURL := deleteWebhookBaseURL
	deleteWebhookBaseURL = server.URL
	defer func() { deleteWebhookBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteWebhook(ctx, client, "webhook1")
	if err != nil {
		t.Fatalf("DeleteWebhook should succeed after rate limit retry, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts (1 rate limited + 1 success), got %d", attempts)
	}
}

// TestMapsEqual tests the map comparison utility
func TestMapsEqual(t *testing.T) {
	tests := []struct {
		name  string
		a     map[string]interface{}
		b     map[string]interface{}
		equal bool
	}{
		{"both nil", nil, nil, true},
		{"one nil", nil, map[string]interface{}{"key": "value"}, false},
		{"empty maps", map[string]interface{}{}, map[string]interface{}{}, true},
		{"same content", map[string]interface{}{"key": "value"}, map[string]interface{}{"key": "value"}, true},
		{"different values", map[string]interface{}{"key": "value1"}, map[string]interface{}{"key": "value2"}, false},
		{"different keys", map[string]interface{}{"key1": "value"}, map[string]interface{}{"key2": "value"}, false},
		{"different length", map[string]interface{}{"key": "value"}, map[string]interface{}{"key": "value", "key2": "value2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapsEqual(tt.a, tt.b)
			if result != tt.equal {
				t.Errorf("mapsEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.equal)
			}
		})
	}
}
