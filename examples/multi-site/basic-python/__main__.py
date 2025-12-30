import pulumi
import pulumi_webflow as webflow

# Example: Create multiple Webflow sites in a single Pulumi program
# This demonstrates the basic pattern for managing multiple sites using list comprehension

# Define site configurations
site_configs = [
    {
        "name": "marketing-site",
        "display_name": "Marketing Site",
        "short_name": "marketing-site",
        "timezone": "America/Los_Angeles",
    },
    {
        "name": "docs-site",
        "display_name": "Documentation Site",
        "short_name": "docs-site",
        "timezone": "America/New_York",
    },
    {
        "name": "blog-site",
        "display_name": "Blog Site",
        "short_name": "blog-site",
        "timezone": "America/Chicago",
    },
    {
        "name": "support-site",
        "display_name": "Support Portal",
        "short_name": "support-site",
        "timezone": "America/Denver",
    },
    {
        "name": "careers-site",
        "display_name": "Careers Page",
        "short_name": "careers-site",
        "timezone": "America/Los_Angeles",
    },
]

# Create all sites using list comprehension
sites = [
    webflow.Site(
        config["name"],
        display_name=config["display_name"],
        short_name=config["short_name"],
        time_zone=config["timezone"],
    )
    for config in site_configs
]

# Create robots.txt for each site
for i, config in enumerate(site_configs):
    webflow.RobotsTxt(
        f"{config['name']}-robots",
        site_id=sites[i].id,
        content="User-agent: *\nAllow: /",
    )

# Export site IDs individually
for i, config in enumerate(site_configs):
    pulumi.export(f"{config['name']}-id", sites[i].id)

# Export all site IDs as a list
pulumi.export("all-site-ids", [site.id for site in sites])
pulumi.export("site-count", len(sites))
