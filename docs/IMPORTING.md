# Importing Existing Webflow Sites into Pulumi

This guide explains how to import existing Webflow sites that were created manually through the Webflow UI or via other methods, so you can manage them through Pulumi Infrastructure as Code.

## Overview

The `pulumi import` command allows you to adopt existing Webflow sites into Pulumi state, converting manual management to infrastructure code. This is useful when:

- You have existing Webflow sites created before using Pulumi
- Your team wants to standardize on IaC for configuration management
- You need to include legacy sites in your Pulumi stack
- You're migrating from manual Webflow UI management to version-controlled infrastructure

## Quick Start

### Step 1: Find Your Site ID

1. Log into your Webflow account at https://webflow.com
2. Navigate to the site you want to import
3. Look at the URL in your browser: `https://webflow.com/dashboard/sites/{SITE_ID}/settings`
4. Copy the **24-character hexadecimal string** after `/sites/` - this is your Site ID
   - Example: `69307a0ff82ccd49b929ed6d`

### Step 2: Run the Import Command

Simply provide your Site ID - the provider automatically fetches the workspace ID from the Webflow API.

```bash
# Set your Webflow API token
export WEBFLOW_API_TOKEN="your-webflow-api-token"

# Import the site using just the Site ID
pulumi import webflow:index:Site my-site "SITE_ID"

# Example with real ID:
pulumi import webflow:index:Site my-site "69307a0ff82ccd49b929ed6d"
```

**Output:**
```
Previewing update (dev)

View Live: https://app.pulumi.com/...

Resources:
    + webflow:index:Site: (import)
        id: workspace-123/sites/64d7f7a60497dc89dd5e80ab
        workspaceId: "workspace-123"
        displayName: "My Existing Site"
        shortName: "my-existing-site"
        timeZone: "America/New_York"
        ...

Do you want to perform this import?  (yes/no):
```

Review the imported site properties, then type `yes` to confirm.

### Step 3: Define the Resource in Your Code

After import, you must add the resource definition to your Pulumi program so the imported state matches your code.

**Python:**
```python
import pulumi
import pulumi_webflow as webflow

# Match the imported site properties
my_site = webflow.Site("my-site",
    workspace_id="workspace-123",
    display_name="My Existing Site",
    short_name="my-existing-site",
    time_zone="America/New_York",
)

pulumi.export("site_id", my_site.id)
```

**TypeScript/JavaScript:**
```typescript
import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

// Match the imported site properties
const mySite = new webflow.Site("my-site", {
    workspaceId: "workspace-123",
    displayName: "My Existing Site",
    shortName: "my-existing-site",
    timeZone: "America/New_York",
});

export const siteId = mySite.id;
```

**Go:**
```go
package main

import (
    "github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        // Match the imported site properties
        site, err := webflow.NewSite(ctx, "my-site", &webflow.SiteArgs{
            WorkspaceId: pulumi.String("workspace-123"),
            DisplayName: pulumi.String("My Existing Site"),
            ShortName:   pulumi.String("my-existing-site"),
            TimeZone:    pulumi.String("America/New_York"),
        })
        if err != nil {
            return err
        }

        ctx.Export("siteId", site.ID())
        return nil
    })
}
```

### Step 4: Verify Import Success

```bash
# Preview to verify the import matches your code
pulumi preview

# Output should show no changes:
# Resources:
#     webflow:index:Site: (no changes)
```

If `pulumi preview` shows "no changes", the import was successful! Your code now matches the imported site state.

### Step 5: Start Managing Your Site

```bash
# Make changes to your site configuration in code
# For example, update the display name:
# displayName: "My Existing Site" → "My Updated Site"

# Preview the changes
pulumi preview

# Apply the changes (update in Webflow)
pulumi up
```

## Importing Multiple Sites

To import multiple sites, repeat the import command for each site:

```bash
# Import first site
pulumi import webflow:index:Site site-one SITE_ID_1

# Import second site  
pulumi import webflow:index:Site site-two SITE_ID_2

# Import third site
pulumi import webflow:index:Site site-three SITE_ID_3

# Add definitions for all three sites to your code
# Then verify with pulumi preview
```

**Bulk Import Script (Bash):**

```bash
#!/bin/bash

# Define your sites in this format: logical-name:site-id
SITES=(
  "site-one:64d7f7a60497dc89dd5e80ab"
  "site-two:64d7f7a60497dc89dd5e80ac"
  "site-three:64d7f7a60497dc89dd5e80ad"
)

for site in "${SITES[@]}"; do
  IFS=':' read -r name id <<< "$site"
  echo "Importing $name (site ID: $id)..."
  pulumi import webflow:index:Site "$name" "$id"
done

echo "Import complete! Add resource definitions to your code."
```

## Finding Your Workspace ID

After importing, Pulumi will display your workspace ID in the resource information. You'll need this for your code definitions.

**Method 1: From Import Output**
The import command shows it in the resource ID:
```
id: workspace-123/sites/64d7f7a60497dc89dd5e80ab
    ^^^^^^^^^^^^^^^ - this is your workspace ID
```

**Method 2: From Webflow Dashboard**
1. Log into https://webflow.com
2. Go to **Account Settings** → **Workspace**
3. Your workspace ID is shown in the workspace information

**Method 3: From Webflow API**
```bash
curl -H "Authorization: Bearer $WEBFLOW_API_TOKEN" \
  https://api.webflow.com/v2/workspaces
```

## Troubleshooting

### "Site not found" Error

**Error:**
```
Error: failed to import site: site not found (site ID: invalid-id-here)
```

**Solutions:**
1. Verify the site ID is correct (24-character hexadecimal string)
2. Confirm the site exists in your Webflow workspace
3. Check you're using the right Webflow account
4. Ensure your API token is valid and has access to this workspace

### "Forbidden" or "Insufficient Permissions" Error

**Error:**
```
Error: failed to import site: Forbidden - access denied to this resource
```

**Solutions:**
1. Verify your `WEBFLOW_API_TOKEN` is set correctly:
   ```bash
   echo $WEBFLOW_API_TOKEN
   ```
2. Check that your API token has the required scopes:
   - `site_config:read` - Read site configuration
   - `site_config:write` - Write site configuration (needed for future updates)
3. Ensure your token belongs to a workspace that has access to the site
4. Regenerate your API token if needed at https://webflow.com/dashboard/integrations/applications

### Import Shows Changes But Code Matches

**Problem:**
After importing and adding resource definitions, `pulumi preview` shows changes even though your code matches the imported site.

**Possible Causes:**
1. **Optional fields not provided in code:** Some site properties like `parentFolderId` are optional. If the imported site has these but your code doesn't, Pulumi will show a change.

   **Solution:** Add all populated fields to your code definition:
   ```typescript
   const mySite = new webflow.Site("my-site", {
       workspaceId: "workspace-123",
       displayName: "My Site",
       shortName: "my-site",
       timeZone: "America/New_York",
       parentFolderId: "folder-456", // Add this if your site has a parent folder
   });
   ```

2. **Whitespace or casing differences:** Field values might differ slightly in formatting.

   **Solution:** Ensure exact match between code and imported properties.

3. **Time zone formatting:** Timezone names must use exact IANA format.

   **Solution:** Use standard IANA timezone identifiers like `America/New_York`, not `EST` or `Eastern Time`.

### Cannot Find Site in Webflow UI

**Problem:**
You know the site ID but can't find the site in Webflow when searching.

**Possible Causes:**
1. Site was moved to a different workspace
2. Site was deleted but import command still works (404 error will occur)
3. You're logged into the wrong Webflow account

**Solution:**
Try importing anyway - you'll get a clear error message indicating the actual issue.

## Import and Version Control

After importing a site, commit your changes to version control:

```bash
# Add import to Git after verifying with pulumi preview
git add Pulumi.yaml my-site-definition.py  # or .ts, .go, etc.

git commit -m "feat: Import existing Webflow site

Imported site: My Existing Site (64d7f7a60497dc89dd5e80ab)
Workspace: workspace-123

This site was previously managed manually in Webflow.
Now managed through infrastructure code via Pulumi."

git push origin main
```

This creates an audit trail of when the site was imported and by whom.

## How Import Works Behind the Scenes

1. **User runs import command:**
   ```bash
   pulumi import webflow:index:Site my-site 64d7f7a60497dc89dd5e80ab
   ```

2. **Pulumi calls the provider's Read method** with the site ID

3. **Provider fetches site data** from Webflow API using `GetSite`

4. **Provider extracts workspace ID** from the API response

5. **Provider constructs full resource ID** in format: `{workspaceId}/sites/{siteId}`

6. **Pulumi saves this to state** under the logical name `my-site`

7. **User adds resource definition to code** with matching properties

8. **Next `pulumi preview`** compares code definition with state:
   - If code matches: "no changes"
   - If code differs: shows what changed
   - If code missing: may show updates

9. **Future `pulumi up`** operations apply infrastructure changes through Pulumi

## API Token Security

When importing sites, your Webflow API token is used to fetch site data. Protect your token:

1. **Never commit tokens to Git:**
   ```bash
   # ❌ WRONG
   export WEBFLOW_API_TOKEN="my-secret-token-123"
   git add script.sh
   
   # ✅ RIGHT
   export WEBFLOW_API_TOKEN="my-secret-token-123"  # Set in terminal only
   pulumi import ...
   ```

2. **Use environment variables or secret management:**
   ```bash
   # Set as environment variable
   export WEBFLOW_API_TOKEN=$(cat ~/.webflow-token)
   
   # Or use a secret manager
   export WEBFLOW_API_TOKEN=$(pass webflow/token)
   
   # Or load from secure location
   export WEBFLOW_API_TOKEN=$(aws secretsmanager get-secret-value --secret-id webflow-token --query SecretString --output text)
   ```

3. **In CI/CD, use secret variables:**
   ```yaml
   # GitHub Actions example
   - name: Import Webflow Site
     env:
       WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
     run: pulumi import webflow:index:Site my-site $SITE_ID
   ```

4. **Rotate tokens regularly:**
   - Regenerate API tokens periodically (e.g., quarterly)
   - Use token rotation policies in your CI/CD platform
   - Revoke old tokens immediately after rotation

## Next Steps

After successfully importing a site:

1. **Manage the site through Pulumi:**
   ```bash
   # Change configuration
   # displayName: "My Site" → "My Updated Site"
   
   pulumi preview   # See what will change
   pulumi up        # Apply changes
   ```

2. **Set up automation:**
   - Use Pulumi Deployments for CI/CD
   - Configure automatic previews on PRs
   - Set up approval gates before production changes

3. **Add more resources:**
   - Import other sites to the same stack
   - Import redirects and robots.txt rules
   - Organize resources into stack files

4. **Reference documentation:**
   - [Pulumi CLI Import Documentation](https://www.pulumi.com/docs/cli/commands/pulumi_import/)
   - [Webflow Pulumi Provider Reference](../README.md)
   - [Pulumi State Management](./state-management.md)

## FAQ

**Q: Can I import a site I don't own?**
A: No. You need API access to the site, which is restricted to workspace members with the appropriate permissions.

**Q: What happens to my site if I don't define it in code after importing?**
A: Pulumi tracks the import in state but doesn't manage it. The site continues to exist in Webflow unchanged. Once you define it in code, Pulumi will manage it going forward.

**Q: Can I uniport a site (stop managing it)?**
A: Yes, remove the resource from your code and run `pulumi destroy`. This removes it from Pulumi state but doesn't delete the site from Webflow.

**Q: What if I import the same site twice with different logical names?**
A: Pulumi will create duplicate entries in state with different logical names but same resource IDs. This is allowed but not recommended - manage one logical name per physical resource.

**Q: Can I import sites from multiple workspaces in the same stack?**
A: Yes, as long as your API token has access to all workspaces.

**Q: How do I track which sites are imported vs. created by Pulumi?**
A: Check your code and state. Imported sites will have manual resource definitions in code. Created sites will have been created by `pulumi up` after code definition. Consider using tags or naming conventions to identify imported resources.

## Support

If you encounter issues importing sites:

1. Check the [Troubleshooting](#troubleshooting) section above
2. Review your API token permissions
3. Verify the site ID format (24-character hexadecimal)
4. Check Webflow API status at https://webflow.status.io
5. Report issues or ask questions on the [GitHub repository](https://github.com/jdetmar/pulumi-webflow)
