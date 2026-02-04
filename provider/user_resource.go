// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// UserResource is a deprecated stub kept for one release cycle so that existing stacks
// containing webflow:index:User resources can gracefully migrate.  The Webflow User
// Management API has been deprecated by Webflow; this stub allows:
//   - Read  → returns empty state (resource will be removed from state on refresh)
//   - Delete → no-op (succeeds silently)
//   - Create/Update → return actionable deprecation errors
//
// Remove this file in the next breaking release.
type UserResource struct{}

// UserResourceArgs preserves the minimal input schema so the engine can deserialize
// existing state that still references this resource type.
type UserResourceArgs struct {
	SiteID       string   `pulumi:"siteId"`
	Email        string   `pulumi:"email"`
	AccessGroups []string `pulumi:"accessGroups,optional"`
	Name         string   `pulumi:"name,optional"`
}

// UserResourceState preserves the minimal output schema for the same reason.
type UserResourceState struct {
	UserResourceArgs
	UserID          string `pulumi:"userId,optional"`
	IsEmailVerified bool   `pulumi:"isEmailVerified,optional"`
	Status          string `pulumi:"status,optional"`
	CreatedOn       string `pulumi:"createdOn,optional"`
	InvitedOn       string `pulumi:"invitedOn,optional"`
	LastUpdated     string `pulumi:"lastUpdated,optional"`
	LastLogin       string `pulumi:"lastLogin,optional"`
}

const userDeprecationMsg = "The webflow:index:User resource has been removed because " +
	"the Webflow User Management API has been deprecated by Webflow. " +
	"Please remove this resource from your Pulumi program and run " +
	"\"pulumi state delete <URN>\" to clean up existing state."

// Annotate marks the User resource as deprecated with an actionable migration message.
func (r *UserResource) Annotate(a infer.Annotator) {
	a.SetToken("index", "User")
	a.Describe(r, "DEPRECATED: "+userDeprecationMsg)
	a.Deprecate(r, userDeprecationMsg)
}

// Annotate adds descriptions to the UserResourceArgs fields.
func (args *UserResourceArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID, "The Webflow site ID.")
	a.Describe(&args.Email, "The email address of the user.")
	a.Describe(&args.AccessGroups, "Access group slugs assigned to the user.")
	a.Describe(&args.Name, "Display name for the user.")
}

// Annotate adds descriptions to the UserResourceState fields.
func (state *UserResourceState) Annotate(a infer.Annotator) {
	a.Describe(&state.UserID, "The Webflow-assigned user ID.")
	a.Describe(&state.IsEmailVerified, "Whether the user has verified their email.")
	a.Describe(&state.Status, "The user's status.")
	a.Describe(&state.CreatedOn, "Timestamp when the user was created.")
	a.Describe(&state.InvitedOn, "Timestamp when the user was invited.")
	a.Describe(&state.LastUpdated, "Timestamp when the user was last updated.")
	a.Describe(&state.LastLogin, "Timestamp when the user last logged in.")
}

// Create returns an error directing users to remove this resource from their program.
func (r *UserResource) Create(
	ctx context.Context, req infer.CreateRequest[UserResourceArgs],
) (infer.CreateResponse[UserResourceState], error) {
	p.GetLogger(ctx).Warningf(userDeprecationMsg)
	return infer.CreateResponse[UserResourceState]{},
		fmt.Errorf("%s", userDeprecationMsg)
}

// Read returns an empty state so that `pulumi refresh` removes the resource from state.
func (r *UserResource) Read(
	ctx context.Context, req infer.ReadRequest[UserResourceArgs, UserResourceState],
) (infer.ReadResponse[UserResourceArgs, UserResourceState], error) {
	p.GetLogger(ctx).Warningf(userDeprecationMsg)
	return infer.ReadResponse[UserResourceArgs, UserResourceState]{}, nil
}

// Update returns an error directing users to remove this resource from their program.
func (r *UserResource) Update(
	ctx context.Context, req infer.UpdateRequest[UserResourceArgs, UserResourceState],
) (infer.UpdateResponse[UserResourceState], error) {
	p.GetLogger(ctx).Warningf(userDeprecationMsg)
	return infer.UpdateResponse[UserResourceState]{},
		fmt.Errorf("%s", userDeprecationMsg)
}

// Delete is a no-op so that `pulumi destroy` succeeds for stacks with existing User resources.
func (r *UserResource) Delete(
	ctx context.Context, req infer.DeleteRequest[UserResourceState],
) error {
	p.GetLogger(ctx).Warningf(userDeprecationMsg)
	return nil
}
