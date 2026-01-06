import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values
site_id = config.require_secret("siteId")
page_id = config.get("pageId")  # Optional: set to get a specific page

"""
Page Data Source Example - Reading Page Information

This example demonstrates how to read page information from a Webflow site.
Pages cannot be created via the API - they must be created in the Webflow designer.
"""

# Example 1: Get all pages for a site
all_pages = webflow.PageData("all-pages",
    site_id=site_id)

# Example 2: Get a specific page by ID (conditional on config)
specific_page = None
if page_id:
    specific_page = webflow.PageData("specific-page",
        site_id=site_id,
        page_id=page_id)

# Export outputs for all pages scenario
def transform_pages(pages):
    """Transform pages array into readable format"""
    return [
        {
            "id": page.page_id,
            "title": page.title,
            "slug": page.slug,
            "draft": page.draft,
            "archived": page.archived,
        }
        for page in pages
    ]

pulumi.export("site_pages", all_pages.pages.apply(transform_pages))
pulumi.export("page_count", all_pages.pages.apply(lambda pages: len(pages)))
pulumi.export("page_ids", all_pages.pages.apply(lambda pages: [p.page_id for p in pages]))

# Export outputs for specific page scenario (if configured)
if specific_page:
    pulumi.export("page_title", specific_page.title)
    pulumi.export("page_slug", specific_page.slug)
    pulumi.export("page_webflow_id", specific_page.webflow_page_id)
    pulumi.export("page_created_on", specific_page.created_on)
    pulumi.export("page_last_updated", specific_page.last_updated)
    pulumi.export("page_is_draft", specific_page.draft)
    pulumi.export("page_is_archived", specific_page.archived)
    pulumi.export("page_parent_id", specific_page.parent_id)
    pulumi.export("page_collection_id", specific_page.collection_id)

# Print helpful information
def print_pages_info(pages):
    print(f"\n📄 Found {len(pages)} pages in the site")

    # Show a sample of pages
    sample_size = min(5, len(pages))
    if sample_size > 0:
        print(f"\nFirst {sample_size} pages:")
        for idx, page in enumerate(pages[:sample_size]):
            print(f"  {idx + 1}. \"{page.title}\" (/{page.slug})")

        if len(pages) > sample_size:
            print(f"  ... and {len(pages) - sample_size} more")

all_pages.pages.apply(print_pages_info)

if specific_page:
    specific_page.title.apply(lambda title: print(f"\n✅ Retrieved page: \"{title}\""))
