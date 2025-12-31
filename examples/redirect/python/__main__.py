import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values
site_id = config.require_secret("siteId")
environment = config.get("environment") or "development"

"""
Redirect Example - Creating and Managing URL Redirects

This example demonstrates how to manage URL redirects for your Webflow sites.
"""

# Example 1: Permanent Redirect (301)
permanent_redirect = webflow.Redirect("old-blog-to-new-blog",
    site_id=site_id,
    source_path="/blog/old-article",
    destination_path="/blog/articles/updated-article",
    status_code=301)

# Example 2: Temporary Redirect (302)
temporary_redirect = webflow.Redirect("temporary-landing-page",
    site_id=site_id,
    source_path="/old-campaign",
    destination_path="/new-campaign-2025",
    status_code=302)

# Example 3: External Redirect
external_redirect = webflow.Redirect("external-partner-link",
    site_id=site_id,
    source_path="/partner",
    destination_path="https://partner-site.com",
    status_code=301)

# Example 4: Bulk Redirects
bulk_redirects = []
redirect_mappings = [
    {"old": "/product-a", "new": "/products/product-a"},
    {"old": "/product-b", "new": "/products/product-b"},
    {"old": "/product-c", "new": "/products/product-c"},
]

for i, mapping in enumerate(redirect_mappings):
    redirect = webflow.Redirect(f"bulk-redirect-{i}",
        site_id=site_id,
        source_path=mapping["old"],
        destination_path=mapping["new"],
        status_code=301)
    bulk_redirects.append(redirect)

# Export values
pulumi.export("deployed_site_id", site_id)
pulumi.export("permanent_redirect_id", permanent_redirect.id)
pulumi.export("temporary_redirect_id", temporary_redirect.id)
pulumi.export("external_redirect_id", external_redirect.id)
pulumi.export("bulk_redirect_ids", [r.id for r in bulk_redirects])

# Print success message
redirect_count = len(bulk_redirects) + 3
site_id.apply(lambda s: print(f"✅ Successfully deployed {redirect_count} redirects to site {s}"))
