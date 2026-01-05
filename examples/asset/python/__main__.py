import pulumi
import pulumi_webflow as webflow

# Create a Pulumi config object
config = pulumi.Config()

# Get configuration values
site_id = config.require_secret("siteId")

"""
Asset Example - Creating and Managing Webflow Assets

This example demonstrates how to register assets with Webflow.
After registration, you'll receive upload_url and upload_details
to complete the file upload to S3.

Two-step process:
1. Register asset metadata (this provider handles this)
2. Upload file to S3 using upload_url + upload_details (done separately)
"""

# Example 1: Register a single asset
# The file_hash is the MD5 hash of your file content
# Generate with: md5sum logo.png (Linux) or md5 -q logo.png (macOS)
logo_asset = webflow.Asset("company-logo",
    site_id=site_id,
    file_name="logo.png",
    file_hash="d41d8cd98f00b204e9800998ecf8427e")

# Example 2: Asset with folder organization
hero_asset = webflow.Asset("hero-image",
    site_id=site_id,
    file_name="hero-banner.jpg",
    file_hash="a1b2c3d4e5f6789012345678abcdef12")
    # parent_folder="folder-id-here"  # Uncomment to organize in a folder

# Example 3: Bulk asset registration
icon_assets = []
icons = [
    {"name": "icon-home", "file_name": "home.svg", "file_hash": "11111111111111111111111111111111"},
    {"name": "icon-settings", "file_name": "settings.svg", "file_hash": "22222222222222222222222222222222"},
    {"name": "icon-user", "file_name": "user.svg", "file_hash": "33333333333333333333333333333333"},
]

for icon in icons:
    asset = webflow.Asset(icon["name"],
        site_id=site_id,
        file_name=icon["file_name"],
        file_hash=icon["file_hash"])
    icon_assets.append(asset)

# Export values for the logo asset
# These are needed to complete the S3 upload
pulumi.export("logo_asset_id", logo_asset.asset_id)
pulumi.export("logo_upload_url", logo_asset.upload_url)
pulumi.export("logo_upload_details", logo_asset.upload_details)
pulumi.export("logo_asset_url", logo_asset.asset_url)
pulumi.export("logo_hosted_url", logo_asset.hosted_url)

# Export hero asset info
pulumi.export("hero_asset_id", hero_asset.asset_id)
pulumi.export("hero_hosted_url", hero_asset.hosted_url)

# Export icon asset IDs
pulumi.export("icon_asset_ids", [a.asset_id for a in icon_assets])

# Print success message
asset_count = len(icons) + 2
site_id.apply(lambda s: print(f"Registered {asset_count} assets for site {s}. Use upload_url and upload_details to complete S3 uploads."))
