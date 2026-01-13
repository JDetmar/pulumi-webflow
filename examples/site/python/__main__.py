import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values
workspace_id = config.require("workspaceId")
display_name = config.require("displayName")
short_name = config.require("shortName")
timezone = config.get("timezone") or "America/New_York"

"""
Site Example - Creating and Managing Webflow Sites

This example demonstrates how to create and manage Webflow sites using Pulumi.
"""

# Example 1: Basic Site Creation
basic_site = webflow.Site("basic-site",
    workspace_id=workspace_id,
    display_name=display_name,
    short_name=short_name,
    time_zone=timezone)

# Example 2: Multi-Environment Site Configuration
environments = ["development", "staging", "production"]
environment_sites = []
for env in environments:
    site = webflow.Site(f"site-{env}",
        workspace_id=workspace_id,
        display_name=f"{display_name}-{env}",
        short_name=f"{short_name}-{env}",
        time_zone=timezone)
    environment_sites.append(site)

# Example 3: Site with Configuration
configured_site = webflow.Site("configured-site",
    workspace_id=workspace_id,
    display_name=f"{display_name}-configured",
    short_name=f"{short_name}-configured",
    time_zone=timezone)

# Export values
pulumi.export("basic_site_id", basic_site.id)
pulumi.export("basic_site_name", basic_site.display_name)
pulumi.export("environment_site_ids", [s.id for s in environment_sites])
pulumi.export("configured_site_id", configured_site.id)

# Print success message
site_count = len(environment_sites) + 2
print(f"✅ Successfully created {site_count} sites")
