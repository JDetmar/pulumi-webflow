# User Resource Examples

This directory contains examples demonstrating how to create and manage users for Webflow sites using Pulumi in multiple languages.

## What You'll Learn

- Invite users to access your Webflow site content
- Assign users to access groups for content gating
- Manage user permissions for premium/paid content
- Automate user provisioning for membership sites
- Set up beta testers and preview access
- Monitor user invitation and verification status

## What Are Webflow Users?

Webflow Users allow you to manage who can access gated content on your site. Users are typically used for:
- **Memberships**: Give paying customers access to premium content
- **Beta Testing**: Invite testers to preview new features
- **Client Sites**: Provide clients with access to their content
- **Communities**: Manage member access to community resources

Users can be assigned to **Access Groups** which control what content they can view on your site.

## Available Languages

| Language   | Directory    | Entry Point    | Dependencies        |
|------------|--------------|----------------|---------------------|
| TypeScript | `typescript/`| `index.ts`     | `package.json`      |

## Quick Start

### TypeScript

```bash
cd typescript
npm install
pulumi stack init dev
pulumi config set siteId your-site-id-here
pulumi up
```

## Examples Included

### 1. Basic User Invitation

Invite a user with just an email address. The user will receive an invitation email.

```typescript
new webflow.User("basic-user", {
  siteId: siteId,
  email: "user@example.com",
});
```

**Use Case:** Simple user invitation without additional metadata.

### 2. User with Display Name

Include a display name that appears in the Webflow dashboard and can be used for personalization.

```typescript
new webflow.User("named-user", {
  siteId: siteId,
  email: "john.doe@example.com",
  name: "John Doe",
});
```

**Use Case:** Professional user management with readable names in the dashboard.

### 3. User with Access Groups

Assign users to specific access groups to control what content they can access.

```typescript
new webflow.User("premium-user", {
  siteId: siteId,
  email: "premium@example.com",
  name: "Premium User",
  accessGroups: ["premium-members", "beta-access"],
});
```

**Use Case:** Grant premium subscribers access to paid content, or give beta testers early access to new features.

### 4. Beta Tester

Create a user specifically for testing features before general release.

```typescript
new webflow.User("beta-tester", {
  siteId: siteId,
  email: "beta.tester@example.com",
  name: "Beta Tester",
  accessGroups: ["beta-testers"],
});
```

**Use Case:** Invite trusted users to test new features or content before public launch.

### 5. User with Multiple Access Levels

Users can be assigned to multiple access groups for granular permission control.

```typescript
new webflow.User("power-user", {
  siteId: siteId,
  email: "power.user@example.com",
  name: "Power User",
  accessGroups: ["premium-members", "early-access", "community-moderators"],
});
```

**Use Case:** Power users who need access to multiple content tiers or have special privileges.

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `siteId`          | Yes      | Your Webflow site ID                     |
| `environment`     | No       | Deployment environment (default: development) |

**Finding Your Site ID:**
1. Log in to Webflow
2. Go to Site Settings → General
3. Copy the Site ID (24-character lowercase hexadecimal string)

**Finding Access Group Slugs:**
1. Go to your Webflow site dashboard
2. Navigate to Users → Access Groups
3. The slug is the URL-friendly version of the group name

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    deployedSiteId           : "abc123..."
    basicUserId              : "user_abc..."
    basicUserEmail           : "user@example.com"
    basicUserStatus          : "invited"
    basicUserVerified        : false
    premiumUserId            : "user_def..."
    premiumUserGroups        : ["premium-members", "beta-access"]
    premiumUserCreated       : "2025-01-06T12:34:56Z"
```

## User Status Values

Users will have one of these status values:

| Status      | Description |
|-------------|-------------|
| `invited`   | Invitation sent, user hasn't accepted yet |
| `verified`  | User accepted invitation and verified email |
| `unverified`| User registered but hasn't verified email |

## Important Notes

### Email is Immutable

**The user's email address cannot be changed after creation.** If you need to change a user's email:

1. Delete the old user resource
2. Create a new user resource with the new email
3. The user will receive a new invitation

```typescript
// This will trigger replacement (delete + recreate)
email: "newemail@example.com"  // Changed from "oldemail@example.com"
```

### User Invitation Flow

1. You create the user via Pulumi
2. Webflow sends an invitation email to the user
3. User clicks the link in the email
4. User accepts the invitation and sets a password
5. User status changes to "verified"

### Access Groups Must Exist

Before assigning users to access groups, ensure the groups exist in Webflow:

1. Go to your Webflow site dashboard
2. Navigate to Users → Access Groups
3. Create any access groups you plan to use
4. Note the group slugs for use in your code

### User Updates

You can update these fields without replacing the user:
- ✅ `name` - Update user's display name
- ✅ `accessGroups` - Change access group assignments

These fields require replacement:
- ❌ `siteId` - User must be deleted and recreated
- ❌ `email` - User must be deleted and recreated

## Testing

To test user creation without sending real emails:

1. Use test email addresses that you control
2. Check the Webflow dashboard to verify users appear
3. Monitor user status changes after invitation acceptance

## Cleanup

To remove all created users:

```bash
pulumi destroy
pulumi stack rm dev
```

**Note:** Deleting a user resource removes their access to the site but does not delete their account data from Webflow immediately. Users may still appear in the Webflow dashboard until they're permanently deleted.

## Troubleshooting

### "Site not found" Error

1. Verify your site ID in Webflow: Settings → General
2. Ensure correct format: 24-character lowercase hexadecimal
3. Check API token has access to the site

### "Invalid email" Error

- Ensure email is in valid format: `user@domain.com`
- Check for typos in the email address
- Verify email doesn't have leading/trailing whitespace

### "Access group not found" Error

1. Verify access groups exist in Webflow dashboard
2. Check that you're using the group **slug**, not the display name
3. Group slugs are case-sensitive and URL-friendly (lowercase, hyphens)

### User Shows "invited" Status

This is normal! Users remain in "invited" status until they:
1. Click the invitation link in their email
2. Accept the invitation
3. Verify their email address

Check the user's email for the invitation.

### Cannot Update User Email

This is expected behavior. Email addresses are immutable in the Webflow API. To change a user's email:
1. Remove the user resource from your code
2. Run `pulumi up` to delete the user
3. Create a new user with the new email
4. Run `pulumi up` again

## Related Resources

- [Main Examples Index](../README.md)
- [Webflow Users Documentation](https://university.webflow.com/lesson/memberships)
- [Access Groups Guide](https://university.webflow.com/lesson/access-groups)
- [Webflow Users API](https://developers.webflow.com/reference/users)

## Next Steps

After creating users, consider:
- Setting up webhooks to monitor user account events
- Creating automated user provisioning workflows
- Integrating with your payment system for membership access
- Building user management dashboards
- Implementing user lifecycle automation (welcome emails, access expiration, etc.)
