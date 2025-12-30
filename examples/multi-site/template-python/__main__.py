import pulumi
import pulumi_webflow as webflow
from site_templates import (
    create_campaign_site,
    create_product_site,
    create_event_site,
)

# Example: Template-based multi-site management
# This demonstrates reusable site factory functions for consistency across fleets

# Create campaign sites
campaigns = [
    ("q1-2025-promotion", "Q1 2025 Promotion"),
    ("summer-sale-2025", "Summer Sale 2025"),
    ("black-friday-2025", "Black Friday 2025"),
]

campaign_sites = [create_campaign_site(name, display) for name, display in campaigns]

# Create product landing pages
products = [
    ("product-alpha", "Product Alpha"),
    ("product-beta", "Product Beta"),
]

product_sites = [create_product_site(name, display) for name, display in products]

# Create event microsite
event_sites = [
    create_event_site("conference-2025", "Annual Conference 2025"),
]

# Combine all sites
all_sites = campaign_sites + product_sites + event_sites

# Export site counts by type
pulumi.export("campaign-sites-count", len(campaign_sites))
pulumi.export("product-sites-count", len(product_sites))
pulumi.export("event-sites-count", len(event_sites))
pulumi.export("total-sites", len(all_sites))

# Export all site IDs
pulumi.export("all-site-ids", [site.id for site in all_sites])
