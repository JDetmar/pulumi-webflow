// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// UserResource is the resource controller for managing Webflow site users.
// It implements the infer.CustomResource interface for full CRUD operations.
type UserResource struct{}

// UserResourceArgs defines the input properties for the User resource.
type UserResourceArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
	// Email is the email address of the user to invite.
	// The user will receive an invitation email at this address.
	// NOTE: Once created, the email cannot be changed via the API.
	Email string `pulumi:"email"`
	// AccessGroups is an optional list of access group slugs to assign to the user.
	// Access groups are assigned as type 'admin' (assigned via API).
	// Example: ["premium-members", "beta-testers"]
	AccessGroups []string `pulumi:"accessGroups,optional"`
	// Name is the optional display name for the user.
	Name string `pulumi:"name,optional"`
}

// UserResourceState defines the output properties for the User resource.
// It embeds UserResourceArgs to include input properties in the output.
type UserResourceState struct {
	UserResourceArgs
	// UserID is the Webflow-assigned user ID (read-only).
	UserID string `pulumi:"userId,optional"`
	// IsEmailVerified indicates whether the user has verified their email (read-only).
	IsEmailVerified bool `pulumi:"isEmailVerified,optional"`
	// Status is the user's status: "invited", "verified", or "unverified" (read-only).
	Status string `pulumi:"status,optional"`
	// CreatedOn is the timestamp when the user was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
	// InvitedOn is the timestamp when the user was invited (read-only).
	InvitedOn string `pulumi:"invitedOn,optional"`
	// LastUpdated is the timestamp when the user was last updated (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
	// LastLogin is the timestamp when the user last logged in (read-only).
	LastLogin string `pulumi:"lastLogin,optional"`
}

// Annotate adds descriptions and constraints to the User resource.
func (r *UserResource) Annotate(a infer.Annotator) {
	a.SetToken("index", "User")
	a.Describe(r, "Manages users for a Webflow site. "+
		"This resource allows you to invite users to your site with specified access groups. "+
		"Users will receive an invitation email and must accept it to access paid content. "+
		"Note: The user's email cannot be changed after creation.")
}

// Annotate adds descriptions to the UserResourceArgs fields.
func (args *UserResourceArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find your site ID in the Webflow dashboard under Site Settings. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.Email,
		"The email address of the user to invite. "+
			"The user will receive an invitation email at this address. "+
			"IMPORTANT: The email cannot be changed after the user is created. "+
			"Changing the email will require replacing the resource (delete + recreate).")

	a.Describe(&args.AccessGroups,
		"Optional list of access group slugs to assign to the user. "+
			"Access groups control what content the user can access. "+
			"Groups are assigned as type 'admin' (assigned via API or designer). "+
			"Example: ['premium-members', 'beta-testers']. "+
			"Access group slugs can be found in the Webflow dashboard under Users > Access Groups.")

	a.Describe(&args.Name,
		"Optional display name for the user. "+
			"This will be shown in the Webflow dashboard and can be used in site personalization.")
}

// Annotate adds descriptions to the UserResourceState fields.
func (state *UserResourceState) Annotate(a infer.Annotator) {
	a.Describe(&state.UserID,
		"The Webflow-assigned unique identifier for this user. "+
			"This is automatically assigned when the user is created and is read-only.")

	a.Describe(&state.IsEmailVerified,
		"Indicates whether the user has verified their email address. "+
			"This is read-only and set by Webflow when the user verifies their email.")

	a.Describe(&state.Status,
		"The status of the user. Possible values: "+
			"'invited' (invitation sent but not accepted), "+
			"'verified' (email verified), "+
			"'unverified' (registered but email not verified). "+
			"This is read-only.")

	a.Describe(&state.CreatedOn,
		"The timestamp when the user was created (RFC3339 format). "+
			"This is automatically set and is read-only.")

	a.Describe(&state.InvitedOn,
		"The timestamp when the user was invited (RFC3339 format). "+
			"This is automatically set and is read-only.")

	a.Describe(&state.LastUpdated,
		"The timestamp when the user was last updated (RFC3339 format). "+
			"This is read-only.")

	a.Describe(&state.LastLogin,
		"The timestamp when the user last logged in (RFC3339 format). "+
			"This is read-only.")
}

// Diff determines what changes need to be made to the user resource.
// siteId and email changes trigger replacement (immutable fields).
// accessGroups and name changes trigger in-place update.
func (r *UserResource) Diff(
	ctx context.Context, req infer.DiffRequest[UserResourceArgs, UserResourceState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}

	// Check for siteId change (requires replacement)
	if req.State.SiteID != req.Inputs.SiteID {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"siteId": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// Check for email change (requires replacement - email is immutable via API)
	if req.State.Email != req.Inputs.Email {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"email": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// Check for accessGroups change (in-place update)
	if !stringSliceEqual(req.State.AccessGroups, req.Inputs.AccessGroups) {
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"accessGroups": {Kind: p.Update},
		}
	}

	// Check for name change (in-place update)
	if req.State.Name != req.Inputs.Name {
		diff.HasChanges = true
		if diff.DetailedDiff == nil {
			diff.DetailedDiff = make(map[string]p.PropertyDiff)
		}
		diff.DetailedDiff["name"] = p.PropertyDiff{Kind: p.Update}
	}

	return diff, nil
}

// Create creates a new user on the Webflow site by sending an invitation.
func (r *UserResource) Create(
	ctx context.Context, req infer.CreateRequest[UserResourceArgs],
) (infer.CreateResponse[UserResourceState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[UserResourceState]{}, fmt.Errorf("validation failed for User resource: %w", err)
	}
	if err := ValidateUserEmail(req.Inputs.Email); err != nil {
		return infer.CreateResponse[UserResourceState]{}, fmt.Errorf("validation failed for User resource: %w", err)
	}
	if len(req.Inputs.AccessGroups) > 0 {
		if err := ValidateUserAccessGroups(req.Inputs.AccessGroups); err != nil {
			return infer.CreateResponse[UserResourceState]{}, fmt.Errorf("validation failed for User resource: %w", err)
		}
	}

	state := UserResourceState{
		UserResourceArgs: req.Inputs,
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		state.UserID = fmt.Sprintf("preview-%d", time.Now().Unix())
		state.Status = "invited"
		state.CreatedOn = time.Now().Format(time.RFC3339)
		state.InvitedOn = time.Now().Format(time.RFC3339)
		return infer.CreateResponse[UserResourceState]{
			ID:     GenerateUserResourceID(req.Inputs.SiteID, state.UserID),
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[UserResourceState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to invite user
	user, err := InviteUser(ctx, client, req.Inputs.SiteID, req.Inputs.Email, req.Inputs.AccessGroups)
	if err != nil {
		return infer.CreateResponse[UserResourceState]{}, fmt.Errorf("failed to invite user: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid user ID
	if user.ID == "" {
		return infer.CreateResponse[UserResourceState]{}, errors.New(
			"webflow API returned empty user ID - " +
				"this is unexpected and may indicate an API issue")
	}

	// Update state with response data
	state.UserID = user.ID
	state.IsEmailVerified = user.IsEmailVerified
	state.Status = user.Status
	state.CreatedOn = user.CreatedOn
	state.InvitedOn = user.InvitedOn
	state.LastUpdated = user.LastUpdated
	state.LastLogin = user.LastLogin

	// Extract name from user data if present
	if user.Data != nil && user.Data.Name != "" {
		state.Name = user.Data.Name
	}

	// Extract access group slugs from response (including empty list to handle removals)
	state.AccessGroups = make([]string, len(user.AccessGroups))
	for i, ag := range user.AccessGroups {
		state.AccessGroups[i] = ag.Slug
	}

	resourceID := GenerateUserResourceID(req.Inputs.SiteID, user.ID)

	return infer.CreateResponse[UserResourceState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of a user from Webflow.
// Used for drift detection and import operations.
func (r *UserResource) Read(
	ctx context.Context, req infer.ReadRequest[UserResourceArgs, UserResourceState],
) (infer.ReadResponse[UserResourceArgs, UserResourceState], error) {
	// Extract siteID and userID from resource ID
	siteID, userID, err := ExtractIDsFromUserResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[UserResourceArgs, UserResourceState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[UserResourceArgs, UserResourceState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get user
	user, err := GetUser(ctx, client, siteID, userID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[UserResourceArgs, UserResourceState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[UserResourceArgs, UserResourceState]{}, fmt.Errorf("failed to read user: %w", err)
	}

	// Extract email and name from user data
	var email, name string
	if user.Data != nil {
		email = user.Data.Email
		name = user.Data.Name
	}

	// Extract access group slugs (including empty list to handle removals)
	accessGroups := make([]string, len(user.AccessGroups))
	for i, ag := range user.AccessGroups {
		accessGroups[i] = ag.Slug
	}

	// Build current state from API response
	currentInputs := UserResourceArgs{
		SiteID:       siteID,
		Email:        email,
		AccessGroups: accessGroups,
		Name:         name,
	}
	currentState := UserResourceState{
		UserResourceArgs: currentInputs,
		UserID:           user.ID,
		IsEmailVerified:  user.IsEmailVerified,
		Status:           user.Status,
		CreatedOn:        user.CreatedOn,
		InvitedOn:        user.InvitedOn,
		LastUpdated:      user.LastUpdated,
		LastLogin:        user.LastLogin,
	}

	return infer.ReadResponse[UserResourceArgs, UserResourceState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update modifies an existing user.
// Note: email and password cannot be updated via the API.
func (r *UserResource) Update(
	ctx context.Context, req infer.UpdateRequest[UserResourceArgs, UserResourceState],
) (infer.UpdateResponse[UserResourceState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.UpdateResponse[UserResourceState]{}, fmt.Errorf("validation failed for User resource: %w", err)
	}
	if len(req.Inputs.AccessGroups) > 0 {
		if err := ValidateUserAccessGroups(req.Inputs.AccessGroups); err != nil {
			return infer.UpdateResponse[UserResourceState]{}, fmt.Errorf("validation failed for User resource: %w", err)
		}
	}

	state := UserResourceState{
		UserResourceArgs: req.Inputs,
		UserID:           req.State.UserID,
		IsEmailVerified:  req.State.IsEmailVerified,
		Status:           req.State.Status,
		CreatedOn:        req.State.CreatedOn,
		InvitedOn:        req.State.InvitedOn,
		LastUpdated:      req.State.LastUpdated,
		LastLogin:        req.State.LastLogin,
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		return infer.UpdateResponse[UserResourceState]{
			Output: state,
		}, nil
	}

	// Extract the Webflow user ID from the Pulumi resource ID
	_, userID, err := ExtractIDsFromUserResourceID(req.ID)
	if err != nil {
		return infer.UpdateResponse[UserResourceState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[UserResourceState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Prepare user data for update
	var userData *UserData
	if req.Inputs.Name != "" {
		userData = &UserData{
			Name: req.Inputs.Name,
		}
	}

	// Call Webflow API
	user, err := UpdateUser(ctx, client, req.Inputs.SiteID, userID, req.Inputs.AccessGroups, userData)
	if err != nil {
		return infer.UpdateResponse[UserResourceState]{}, fmt.Errorf("failed to update user: %w", err)
	}

	// Update state with response data
	state.UserID = user.ID
	state.IsEmailVerified = user.IsEmailVerified
	state.Status = user.Status
	state.LastUpdated = user.LastUpdated
	state.LastLogin = user.LastLogin

	// Extract access group slugs from response (including empty list to handle removals)
	state.AccessGroups = make([]string, len(user.AccessGroups))
	for i, ag := range user.AccessGroups {
		state.AccessGroups[i] = ag.Slug
	}

	// Extract name from user data if present
	if user.Data != nil && user.Data.Name != "" {
		state.Name = user.Data.Name
	}

	return infer.UpdateResponse[UserResourceState]{
		Output: state,
	}, nil
}

// Delete removes a user from the Webflow site.
func (r *UserResource) Delete(
	ctx context.Context, req infer.DeleteRequest[UserResourceState],
) (infer.DeleteResponse, error) {
	// Extract siteID and userID from resource ID
	siteID, userID, err := ExtractIDsFromUserResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteUser(ctx, client, siteID, userID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete user: %w", err)
	}

	return infer.DeleteResponse{}, nil
}

// stringSliceEqual compares two string slices for equality.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
