import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * User Example - Creating and Managing Webflow Site Users
 *
 * This example demonstrates how to invite and manage users for Webflow Memberships and gated content.
 * Users can be assigned to access groups to control what content they can view.
 *
 * Use Cases:
 * - Invite users to access premium/paid content
 * - Manage beta testers or preview access
 * - Set up user access for client sites
 * - Automate user provisioning for membership sites
 */

// Example 1: Basic User Invitation
// Invite a user with just an email address
const basicUser = new webflow.User("basic-user", {
  siteId: siteId,
  email: "user@example.com",
});

// Example 2: User with Display Name
// Include a display name that appears in the Webflow dashboard
const namedUser = new webflow.User("named-user", {
  siteId: siteId,
  email: "john.doe@example.com",
  name: "John Doe",
});

// Example 3: User with Access Groups
// Assign user to specific access groups for content gating
// Access groups control what content the user can access
const premiumUser = new webflow.User("premium-user", {
  siteId: siteId,
  email: "premium@example.com",
  name: "Premium User",
  accessGroups: ["premium-members", "beta-access"],
});

// Example 4: Beta Tester
// Create a user for testing features before general release
const betaTester = new webflow.User("beta-tester", {
  siteId: siteId,
  email: "beta.tester@example.com",
  name: "Beta Tester",
  accessGroups: ["beta-testers"],
});

// Example 5: User with Multiple Access Levels
// Users can be in multiple access groups for granular permissions
const powerUser = new webflow.User("power-user", {
  siteId: siteId,
  email: "power.user@example.com",
  name: "Power User",
  accessGroups: ["premium-members", "early-access", "community-moderators"],
});

// Export user information for reference
export const deployedSiteId = siteId;

// Basic user outputs
export const basicUserId = basicUser.userId;
export const basicUserEmail = basicUser.email;
export const basicUserStatus = basicUser.status;
export const basicUserVerified = basicUser.isEmailVerified;

// Named user outputs
export const namedUserId = namedUser.userId;
export const namedUserName = namedUser.name;

// Premium user outputs
export const premiumUserId = premiumUser.userId;
export const premiumUserGroups = premiumUser.accessGroups;
export const premiumUserCreated = premiumUser.createdOn;
export const premiumUserInvited = premiumUser.invitedOn;

// Beta tester outputs
export const betaTesterUserId = betaTester.userId;
export const betaTesterStatus = betaTester.status;

// Power user outputs
export const powerUserId = powerUser.userId;
export const powerUserGroups = powerUser.accessGroups;

// Print deployment success message
const userCount = 5;
const message = pulumi.interpolate`✅ Successfully invited ${userCount} users to site ${siteId}`;
message.apply((m) => console.log(m));
