"""
Webflow Pulumi Provider Quickstart - Python

Deploy a RobotsTxt resource to your Webflow site using Python.
This example creates a robots.txt file that allows all search engine crawlers.
"""

import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values (these should be set via `pulumi config set`)
site_id = config.require_secret("siteId")

# Deploy a RobotsTxt resource to your Webflow site
#
# This example creates a robots.txt file that:
# - Allows all search engine crawlers (User-agent: *)
# - Allows Google's bot (Googlebot) to crawl all pages
#
# You can customize the robots.txt content by modifying the content string below
robots_txt = webflow.RobotsTxt(
    "my-robots",
    site_id=site_id,
    content="""User-agent: *
Allow: /

User-agent: Googlebot
Allow: /""",
)

# Export values for reference
pulumi.export("deployed_site_id", site_id)
pulumi.export("robots_txt_id", robots_txt.id)

# Print a success message
print("✅ Successfully deployed robots.txt to your Webflow site")
