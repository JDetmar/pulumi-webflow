import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

/**
 * Token Data Sources Example
 *
 * This example demonstrates how to use the Token data sources to:
 * - Retrieve API token authorization information
 * - Get details about the user who authorized the token
 * - Use token info for validation and conditional logic
 */

// Example 1: Get Token Information
// This retrieves comprehensive details about the current API token
const tokenInfo = webflow.getTokenInfoOutput({});

// Example 2: Get Authorized User Information
// This retrieves information about the user who authorized the token
const authorizedUser = webflow.getAuthorizedUserOutput({});

// Export token authorization details
export const authorizationId = tokenInfo.authorization.id;
export const tokenCreatedOn = tokenInfo.authorization.createdOn;
export const tokenLastUsed = tokenInfo.authorization.lastUsed;
export const grantType = tokenInfo.authorization.grantType;
export const rateLimit = tokenInfo.authorization.rateLimit;
export const scopes = tokenInfo.authorization.scope;

// Export authorized resource IDs
export const authorizedSiteIds = tokenInfo.authorization.authorizedTo.siteIds;
export const authorizedWorkspaceIds =
  tokenInfo.authorization.authorizedTo.workspaceIds;
export const authorizedUserIds = tokenInfo.authorization.authorizedTo.userIds;

// Export application details
export const applicationId = tokenInfo.application.id;
export const applicationName = tokenInfo.application.displayName;
export const applicationHomepage = tokenInfo.application.homepage;

// Export authorized user details
export const authorizedUserId = authorizedUser.userId;
export const authorizedUserEmail = authorizedUser.email;
export const authorizedUserFirstName = authorizedUser.firstName;
export const authorizedUserLastName = authorizedUser.lastName;
export const authorizedUserFullName = pulumi.interpolate`${authorizedUser.firstName} ${authorizedUser.lastName}`;

// Example 3: Validation - Check if token has required scopes
// This demonstrates using token info for validation before creating resources
tokenInfo.authorization.scope.apply((scope) => {
  console.log(`Token scopes: ${scope}`);

  // Example validation checks
  const requiredScopes = ["sites:read"];
  const hasRequiredScopes = requiredScopes.every(
    (required) => scope && scope.includes(required)
  );

  if (hasRequiredScopes) {
    console.log("Token has all required scopes for this deployment");
  } else {
    console.warn(
      `Warning: Token may be missing required scopes: ${requiredScopes.join(", ")}`
    );
  }
});

// Example 4: Display summary
pulumi
  .all([
    authorizedUser.email,
    tokenInfo.authorization.rateLimit,
    tokenInfo.authorization.authorizedTo.siteIds,
  ])
  .apply(([email, limit, sites]) => {
    console.log("\n=== Token Summary ===");
    console.log(`Authorized by: ${email}`);
    console.log(`Rate limit: ${limit} requests/minute`);
    console.log(`Authorized sites: ${sites?.length || 0}`);
    console.log("=====================\n");
  });
