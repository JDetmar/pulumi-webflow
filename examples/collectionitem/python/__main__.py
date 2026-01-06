import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values
collection_id = config.require("collectionId")
environment = config.get("environment") or "development"

"""
CollectionItem Example - Creating and Managing CMS Content

This example demonstrates how to manage collection items (blog posts, products, etc.)
in your Webflow CMS. Collection items are the individual content entries within a
CMS collection.
"""

# Example 1: Draft Blog Post
# Create a blog post that exists in the CMS but is NOT published to the live site
draft_blog_post = webflow.CollectionItem("draft-blog-post",
    collection_id=collection_id,
    field_data={
        "name": "Getting Started with Webflow CMS",
        "slug": "getting-started-webflow-cms",
        # Add your custom fields here based on your collection schema
        # Example custom fields (uncomment and modify based on your schema):
        # "post-body": "Learn how to use Webflow CMS to manage your content...",
        # "author": "John Doe",
        # "publish-date": "2025-01-06",
        # "featured-image": "https://example.com/image.jpg",
    },
    is_draft=True,  # Not published to live site
    is_archived=False)

# Example 2: Published Product
# Create a product that is immediately visible on the live site
published_product = webflow.CollectionItem("published-product",
    collection_id=collection_id,
    field_data={
        "name": "Premium Widget",
        "slug": "premium-widget",
        # Add your custom fields here based on your collection schema
        # Example custom fields (uncomment and modify based on your schema):
        # "price": 99.99,
        # "description": "The best widget on the market",
        # "category": "Electronics",
        # "in-stock": True,
    },
    is_draft=False,  # Published to live site
    is_archived=False)

# Example 3: Archived Content
# Create an item that is archived (hidden but retained for records)
archived_item = webflow.CollectionItem("archived-item",
    collection_id=collection_id,
    field_data={
        "name": "Discontinued Product",
        "slug": "discontinued-product-archive",
    },
    is_draft=True,
    is_archived=True)  # Hidden from both CMS and live site

# Example 4: Bulk Content Creation
# Create multiple items efficiently using a loop
bulk_items = []
content_data = [
    {
        "name": "Introduction to TypeScript",
        "slug": "intro-typescript",
        "category": "Tutorial",
    },
    {
        "name": "Advanced Pulumi Patterns",
        "slug": "advanced-pulumi-patterns",
        "category": "Tutorial",
    },
    {
        "name": "Webflow API Best Practices",
        "slug": "webflow-api-best-practices",
        "category": "Guide",
    },
]

for i, data in enumerate(content_data):
    item = webflow.CollectionItem(f"bulk-item-{i}",
        collection_id=collection_id,
        field_data={
            "name": data["name"],
            "slug": data["slug"],
            # Add your custom fields here
            # "category": data["category"],
        },
        is_draft=True)  # Start as drafts
    bulk_items.append(item)

# Example 5: Localized Content (optional - only if your site uses localization)
# Uncomment if your Webflow site has localization enabled
# localized_item = webflow.CollectionItem("localized-item",
#     collection_id=collection_id,
#     field_data={
#         "name": "Bienvenue",
#         "slug": "bienvenue",
#     },
#     cms_locale_id="fr-FR",  # French locale
#     is_draft=False)

# Export the resource IDs for reference
pulumi.export("deployed_collection_id", collection_id)
pulumi.export("deployed_environment", environment)

# Draft blog post exports
pulumi.export("draft_post_id", draft_blog_post.id)
pulumi.export("draft_post_item_id", draft_blog_post.item_id)
pulumi.export("draft_post_created_on", draft_blog_post.created_on)

# Published product exports
pulumi.export("published_product_id", published_product.id)
pulumi.export("published_product_item_id", published_product.item_id)
pulumi.export("published_product_last_updated", published_product.last_updated)

# Archived item exports
pulumi.export("archived_item_id", archived_item.id)
pulumi.export("archived_item_item_id", archived_item.item_id)

# Bulk items exports
pulumi.export("bulk_item_ids", [item.id for item in bulk_items])
pulumi.export("bulk_item_item_ids", [item.item_id for item in bulk_items])

# Print deployment success message
total_items = 3 + len(bulk_items)
print(f"✅ Successfully deployed {total_items} collection items to collection {collection_id}")
print(f"   Environment: {environment}")
print(f"   Draft items: {1 + len(bulk_items)}")
print(f"   Published items: 1")
print(f"   Archived items: 1")
