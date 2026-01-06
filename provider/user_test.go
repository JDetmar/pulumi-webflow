// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestValidateUserEmail_Valid tests valid email addresses
func TestValidateUserEmail_Valid(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"simple email", "user@example.com"},
		{"email with subdomain", "user@mail.example.com"},
		{"email with plus", "user+tag@example.com"},
		{"email with dots", "first.last@example.com"},
		{"email with numbers", "user123@example.com"},
		{"email with hyphen domain", "user@my-domain.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserEmail(tt.email)
			if err != nil {
				t.Errorf("ValidateUserEmail(%q) = %v, want nil", tt.email, err)
			}
		})
	}
}

// TestValidateUserEmail_Empty tests empty email
func TestValidateUserEmail_Empty(t *testing.T) {
	err := ValidateUserEmail("")
	if err == nil {
		t.Error("ValidateUserEmail(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidateUserEmail_Invalid tests invalid email formats
func TestValidateUserEmail_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"no at sign", "userexample.com"},
		{"no domain", "user@"},
		{"no local part", "@example.com"},
		{"spaces", "user @example.com"},
		{"double at", "user@@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserEmail(tt.email)
			if err == nil {
				t.Errorf("ValidateUserEmail(%q) = nil, want error", tt.email)
			}
			if !strings.Contains(err.Error(), "invalid") {
				t.Errorf("Expected error to mention 'invalid', got: %v", err)
			}
		})
	}
}

// TestValidateUserAccessGroups_Valid tests valid access group slugs
func TestValidateUserAccessGroups_Valid(t *testing.T) {
	tests := []struct {
		name   string
		groups []string
	}{
		{"single group", []string{"premium"}},
		{"multiple groups", []string{"premium", "beta-testers"}},
		{"group with hyphen", []string{"premium-members"}},
		{"group with underscore", []string{"beta_testers"}},
		{"group with numbers", []string{"tier1", "tier2"}},
		{"empty slice", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserAccessGroups(tt.groups)
			if err != nil {
				t.Errorf("ValidateUserAccessGroups(%v) = %v, want nil", tt.groups, err)
			}
		})
	}
}

// TestValidateUserAccessGroups_Invalid tests invalid access group slugs
func TestValidateUserAccessGroups_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		groups []string
	}{
		{"empty string in slice", []string{"premium", ""}},
		{"only empty string", []string{""}},
		{"group with spaces", []string{"premium members"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserAccessGroups(tt.groups)
			if err == nil {
				t.Errorf("ValidateUserAccessGroups(%v) = nil, want error", tt.groups)
			}
		})
	}
}

// TestGenerateUserResourceID tests resource ID generation
func TestGenerateUserResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	userID := "6287ec36a841b25637c663df"

	resourceID := GenerateUserResourceID(siteID, userID)
	expected := "5f0c8c9e1c9d440000e8d8c3/users/6287ec36a841b25637c663df"

	if resourceID != expected {
		t.Errorf("GenerateUserResourceID() = %q, want %q", resourceID, expected)
	}
}

// TestExtractIDsFromUserResourceID_Valid tests extracting IDs from valid resource ID
func TestExtractIDsFromUserResourceID_Valid(t *testing.T) {
	resourceID := "5f0c8c9e1c9d440000e8d8c3/users/6287ec36a841b25637c663df"

	siteID, userID, err := ExtractIDsFromUserResourceID(resourceID)
	if err != nil {
		t.Errorf("ExtractIDsFromUserResourceID() error = %v, want nil", err)
	}
	if siteID != "5f0c8c9e1c9d440000e8d8c3" {
		t.Errorf("ExtractIDsFromUserResourceID() siteID = %q, want %q", siteID, "5f0c8c9e1c9d440000e8d8c3")
	}
	if userID != "6287ec36a841b25637c663df" {
		t.Errorf("ExtractIDsFromUserResourceID() userID = %q, want %q", userID, "6287ec36a841b25637c663df")
	}
}

// TestExtractIDsFromUserResourceID_Empty tests empty resource ID
func TestExtractIDsFromUserResourceID_Empty(t *testing.T) {
	_, _, err := ExtractIDsFromUserResourceID("")
	if err == nil {
		t.Error("ExtractIDsFromUserResourceID(\"\") error = nil, want error")
	}
}

// TestExtractIDsFromUserResourceID_InvalidFormat tests invalid format
func TestExtractIDsFromUserResourceID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{"missing users part", "5f0c8c9e1c9d440000e8d8c3/6287ec36a841b25637c663df"},
		{"wrong middle part", "5f0c8c9e1c9d440000e8d8c3/redirects/6287ec36a841b25637c663df"},
		{"too few parts", "5f0c8c9e1c9d440000e8d8c3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ExtractIDsFromUserResourceID(tt.resourceID)
			if err == nil {
				t.Errorf("ExtractIDsFromUserResourceID(%q) error = nil, want error", tt.resourceID)
			}
		})
	}
}

// TestErrorMessagesAreActionable_User verifies error messages contain guidance
func TestErrorMessagesAreActionable_User(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		contains []string
	}{
		{
			"ValidateUserEmail empty",
			func() error { return ValidateUserEmail("") },
			[]string{"required", "email"},
		},
		{
			"ValidateUserEmail invalid",
			func() error { return ValidateUserEmail("invalid") },
			[]string{"invalid", "format"},
		},
		{
			"ValidateUserAccessGroups empty slug",
			func() error { return ValidateUserAccessGroups([]string{"valid", ""}) },
			[]string{"empty", "non-empty"},
		},
		{
			"ValidateUserAccessGroups spaces",
			func() error { return ValidateUserAccessGroups([]string{"premium members"}) },
			[]string{"spaces", "hyphens"},
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

// TestGetUser_Valid tests retrieving a user successfully
func TestGetUser_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/users/") {
			t.Errorf("Expected /users/ in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := User{
			ID:              "6287ec36a841b25637c663df",
			IsEmailVerified: true,
			Status:          "verified",
			CreatedOn:       "2022-05-20T13:46:12.093Z",
			AccessGroups: []UserAccessGroup{
				{Slug: "premium", Type: "admin"},
			},
			Data: &UserData{
				Name:  "Arthur Dent",
				Email: "arthur@example.com",
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL for this test
	oldURL := getUserBaseURL
	getUserBaseURL = server.URL
	defer func() { getUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "6287ec36a841b25637c663df")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if result.ID != "6287ec36a841b25637c663df" {
		t.Errorf("Expected ID 6287ec36a841b25637c663df, got %s", result.ID)
	}
	if result.Status != "verified" {
		t.Errorf("Expected status verified, got %s", result.Status)
	}
	if result.Data == nil || result.Data.Name != "Arthur Dent" {
		t.Errorf("Expected name Arthur Dent, got %v", result.Data)
	}
}

// TestGetUser_NotFound tests 404 handling
func TestGetUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"user_not_found","message":"User not found"}`))
	}))
	defer server.Close()

	oldURL := getUserBaseURL
	getUserBaseURL = server.URL
	defer func() { getUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestInviteUser_Valid tests inviting a user successfully
func TestInviteUser_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/users/invite") {
			t.Errorf("Expected /users/invite in path, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req InviteUserRequest
		_ = json.Unmarshal(body, &req)

		if req.Email != "arthur@example.com" {
			t.Errorf("Expected email arthur@example.com, got %s", req.Email)
		}
		if len(req.AccessGroups) != 1 || req.AccessGroups[0] != "premium" {
			t.Errorf("Expected accessGroups [premium], got %v", req.AccessGroups)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := User{
			ID:              "6287ec36a841b25637c663df",
			IsEmailVerified: false,
			Status:          "invited",
			InvitedOn:       "2022-05-20T13:46:12.093Z",
			CreatedOn:       "2022-05-20T13:46:12.093Z",
			AccessGroups: []UserAccessGroup{
				{Slug: "premium", Type: "admin"},
			},
			Data: &UserData{
				Email: "arthur@example.com",
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := inviteUserBaseURL
	inviteUserBaseURL = server.URL
	defer func() { inviteUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := InviteUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "arthur@example.com", []string{"premium"})
	if err != nil {
		t.Fatalf("InviteUser failed: %v", err)
	}

	if result.ID != "6287ec36a841b25637c663df" {
		t.Errorf("Expected ID 6287ec36a841b25637c663df, got %s", result.ID)
	}
	if result.Status != "invited" {
		t.Errorf("Expected status invited, got %s", result.Status)
	}
}

// TestInviteUser_DuplicateEmail tests 409 handling for duplicate email
func TestInviteUser_DuplicateEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"code":"duplicate_user_email","message":"A user with this email already exists"}`))
	}))
	defer server.Close()

	oldURL := inviteUserBaseURL
	inviteUserBaseURL = server.URL
	defer func() { inviteUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := InviteUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "existing@example.com", nil)
	if err == nil {
		t.Error("Expected error for 409, got nil")
	}
	// The error handler returns "unexpected error" for 409 status code
	if !strings.Contains(err.Error(), "duplicate_user_email") {
		t.Errorf("Expected 'duplicate_user_email' in error, got: %v", err)
	}
}

// TestInviteUser_ValidationError tests 400 handling
func TestInviteUser_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"validation_error","message":"Invalid email address"}`))
	}))
	defer server.Close()

	oldURL := inviteUserBaseURL
	inviteUserBaseURL = server.URL
	defer func() { inviteUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := InviteUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "invalid", nil)
	if err == nil {
		t.Error("Expected error for 400, got nil")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("Expected 'bad request' in error, got: %v", err)
	}
}

// TestUpdateUser_Valid tests updating a user successfully
func TestUpdateUser_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req UpdateUserRequest
		_ = json.Unmarshal(body, &req)

		if len(req.AccessGroups) != 2 {
			t.Errorf("Expected 2 accessGroups, got %d", len(req.AccessGroups))
		}
		if req.Data == nil || req.Data.Name != "Arthur Dent" {
			t.Errorf("Expected name Arthur Dent, got %v", req.Data)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := User{
			ID:              "6287ec36a841b25637c663df",
			IsEmailVerified: true,
			Status:          "verified",
			LastUpdated:     "2022-05-21T10:00:00.000Z",
			AccessGroups: []UserAccessGroup{
				{Slug: "premium", Type: "admin"},
				{Slug: "beta", Type: "admin"},
			},
			Data: &UserData{
				Name:  "Arthur Dent",
				Email: "arthur@example.com",
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := updateUserBaseURL
	updateUserBaseURL = server.URL
	defer func() { updateUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	userData := &UserData{Name: "Arthur Dent"}
	result, err := UpdateUser(
		ctx, client, "5f0c8c9e1c9d440000e8d8c3", "6287ec36a841b25637c663df",
		[]string{"premium", "beta"}, userData,
	)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	if len(result.AccessGroups) != 2 {
		t.Errorf("Expected 2 access groups, got %d", len(result.AccessGroups))
	}
	if result.Data == nil || result.Data.Name != "Arthur Dent" {
		t.Errorf("Expected name Arthur Dent, got %v", result.Data)
	}
}

// TestUpdateUser_NotFound tests 404 handling for update
func TestUpdateUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"user_not_found","message":"User not found"}`))
	}))
	defer server.Close()

	oldURL := updateUserBaseURL
	updateUserBaseURL = server.URL
	defer func() { updateUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := UpdateUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "nonexistent", nil, nil)
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestDeleteUser_Valid tests deleting a user successfully
func TestDeleteUser_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	oldURL := deleteUserBaseURL
	deleteUserBaseURL = server.URL
	defer func() { deleteUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "6287ec36a841b25637c663df")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}

// TestDeleteUser_NotFound_Idempotent tests that 404 on delete is treated as success (idempotent)
func TestDeleteUser_NotFound_Idempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"user_not_found","message":"User not found"}`))
	}))
	defer server.Close()

	oldURL := deleteUserBaseURL
	deleteUserBaseURL = server.URL
	defer func() { deleteUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "nonexistent")
	if err != nil {
		t.Errorf("DeleteUser should handle 404 as success (idempotent), got error: %v", err)
	}
}

// TestDeleteUser_ServerError tests error handling
func TestDeleteUser_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":"internal_error","message":"Internal server error"}`))
	}))
	defer server.Close()

	oldURL := deleteUserBaseURL
	deleteUserBaseURL = server.URL
	defer func() { deleteUserBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteUser(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "6287ec36a841b25637c663df")
	if err == nil {
		t.Error("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' in error, got: %v", err)
	}
}

// TestStringSliceEqual tests the helper function
func TestStringSliceEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{"both nil", nil, nil, true},
		{"both empty", []string{}, []string{}, true},
		{"equal single", []string{"a"}, []string{"a"}, true},
		{"equal multiple", []string{"a", "b", "c"}, []string{"a", "b", "c"}, true},
		{"different length", []string{"a"}, []string{"a", "b"}, false},
		{"different values", []string{"a", "b"}, []string{"a", "c"}, false},
		{"different order", []string{"a", "b"}, []string{"b", "a"}, false},
		{"nil vs empty", nil, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSliceEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("stringSliceEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
