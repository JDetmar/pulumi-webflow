import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values
display_name = config.require("displayName")
short_name = config.require("shortName")
custom_domain = config.get("customDomain")
timezone = config.get("timezone") or "America/New_York"
environment = config.get("environment") or "development"

"""
Site Example - Creating and Managing Webflow Sites

This example demonstrates how to create and manage Webflow sites using Pulumi.
"""

# Example 1: Basic Site Creation
basic_site = webflow.Site("basic-site",
    display_name=display_name,
    short_name=short_name,
    timezone=timezone)

# Example 2: Site with Custom Domain
site_with_domain = None
if custom_domain:
    site_with_domain = webflow.Site("site-with-domain",
        display_name=f"{display_name}-domain",
        short_name=f"{short_name}-domain",
        custom_domain=custom_domain,
        timezone=timezone)

# Example 3: Multi-Environment Site Configuration
environments = ["development", "staging", "production"]
environment_sites = []
for env in environments:
    site = webflow.Site(f"site-{env}",
        display_name=f"{display_name}-{env}",
        short_name=f"{short_name}-{env}",
        timezone=timezone)
    environment_sites.append(site)

# Example 4: Site with Configuration
configured_site = webflow.Site("configured-site",
    display_name=f"{display_name}-configured",
    short_name=f"{short_name}-configured",
    timezone=timezone)

# Export values
pulumi.export("basic_site_id", basic_site.id)
pulumi.export("basic_site_name", basic_site.display_name)
pulumi.export("custom_domain_site_id", site_with_domain.id if site_with_domain else "not-created")
pulumi.export("environment_site_ids", [s.id for s in environment_sites])
pulumi.export("configured_site_id", configured_site.id)

# Print success message
message = pulumi.interpolate(f"✅ Successfully created {len(environment_sites) + 1} sites")
message.apply(lambda m: print(m))
