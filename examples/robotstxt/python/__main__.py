import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values
site_id = config.require_secret("siteId")

"""
RobotsTxt Example - Creating and Managing robots.txt Files

This example demonstrates how to manage robots.txt files for your Webflow sites.
The robots.txt file controls how search engine crawlers interact with your site.
"""

# Example 1: Allow All Crawlers (most common for public sites)
allow_all_robots = webflow.RobotsTxt("allow-all-robots",
    site_id=site_id,
    content="""User-agent: *
Allow: /

# Allow specific crawler access with no delays
User-agent: Googlebot
Allow: /
Crawl-delay: 0

User-agent: Bingbot
Allow: /
Crawl-delay: 1""")

# Example 2: Selective Blocking (for staging/development)
selective_block_robots = webflow.RobotsTxt("selective-block-robots",
    site_id=site_id,
    content="""User-agent: *
Allow: /

# Disallow admin and private sections
Disallow: /admin/
Disallow: /private/
Disallow: /staging/
Disallow: /test/

# Block specific crawlers
User-agent: AhrefsBot
Disallow: /

User-agent: SemrushBot
Disallow: /""")

# Example 3: Restrict Directories (protect API and backend)
restrict_directories_robots = webflow.RobotsTxt("restrict-directories-robots",
    site_id=site_id,
    content="""User-agent: *
Allow: /
Disallow: /api/
Disallow: /internal/
Disallow: /*.json$
Disallow: /*.xml$

# Specify sitemap location
Sitemap: https://example.com/sitemap.xml""")

# Export the robot resources for reference
pulumi.export("deployed_site_id", site_id)
pulumi.export("allow_all_robots_id", allow_all_robots.id)
pulumi.export("allow_all_robots_last_modified", allow_all_robots.last_modified)
pulumi.export("selective_block_robots_id", selective_block_robots.id)
pulumi.export("restrict_directories_robots_id", restrict_directories_robots.id)

# Print deployment success message
site_id.apply(print)
