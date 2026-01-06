import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values
site_id = config.require_secret("siteId")

"""
Collection Example - Creating and Managing CMS Collections

This example demonstrates how to create CMS collections for your Webflow sites.
Collections are containers for structured content items (blog posts, products, etc.).

Important: Webflow collections do not support updates via API.
Any changes to collection properties will require replacement (delete + recreate).
"""

# Example 1: Blog Posts Collection
blog_collection = webflow.Collection("blog-posts-collection",
    site_id=site_id,
    display_name="Blog Posts",
    singular_name="Blog Post",
    slug="blog-posts")

# Example 2: Products Collection with Auto-Generated Slug
products_collection = webflow.Collection("products-collection",
    site_id=site_id,
    display_name="Products",
    singular_name="Product")
    # slug is optional - Webflow will generate automatically

# Example 3: Team Members Collection
team_collection = webflow.Collection("team-members-collection",
    site_id=site_id,
    display_name="Team Members",
    singular_name="Team Member",
    slug="team")

# Example 4: Portfolio Items Collection
portfolio_collection = webflow.Collection("portfolio-collection",
    site_id=site_id,
    display_name="Portfolio Items",
    singular_name="Portfolio Item",
    slug="portfolio")

# Example 5: Dynamic Collections Based on Config
environment = config.get("environment") or "development"
test_collection = webflow.Collection(f"test-collection-{environment}",
    site_id=site_id,
    display_name=f"Test Collection ({environment})",
    singular_name="Test Item",
    slug=f"test-{environment}")

# Export collection details for reference
pulumi.export("deployed_site_id", site_id)
pulumi.export("blog_collection_id", blog_collection.id)
pulumi.export("blog_collection_name", blog_collection.display_name)
pulumi.export("blog_collection_slug", blog_collection.slug)
pulumi.export("blog_collection_created_on", blog_collection.created_on)

pulumi.export("products_collection_id", products_collection.id)
pulumi.export("team_collection_id", team_collection.id)
pulumi.export("portfolio_collection_id", portfolio_collection.id)
pulumi.export("test_collection_id", test_collection.id)

# Export a summary of all collections
pulumi.export("all_collections", pulumi.Output.all(
    blog_collection.display_name,
    products_collection.display_name,
    team_collection.display_name,
    portfolio_collection.display_name,
    test_collection.display_name
).apply(lambda names: ", ".join(names)))

# Print success message
site_id.apply(lambda s: print(f"✅ Successfully deployed 5 collections to site {s}"))
