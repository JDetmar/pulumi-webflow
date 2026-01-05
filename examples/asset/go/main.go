package main

import (
	"fmt"

	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := pulumi.NewConfig(ctx, "")
		siteID := cfg.RequireSecret("siteId")

		// Example 1: Register a single asset
		// The fileHash is the MD5 hash of your file content
		// Generate with: md5sum logo.png (Linux) or md5 -q logo.png (macOS)
		logoAsset, err := webflow.NewAsset(ctx, "company-logo", &webflow.AssetArgs{
			SiteId:   siteID,
			FileName: pulumi.String("logo.png"),
			FileHash: pulumi.String("d41d8cd98f00b204e9800998ecf8427e"),
		})
		if err != nil {
			return fmt.Errorf("failed to create logo asset: %w", err)
		}

		// Example 2: Asset with folder organization
		heroAsset, err := webflow.NewAsset(ctx, "hero-image", &webflow.AssetArgs{
			SiteId:   siteID,
			FileName: pulumi.String("hero-banner.jpg"),
			FileHash: pulumi.String("a1b2c3d4e5f6789012345678abcdef12"),
			// ParentFolder: pulumi.String("folder-id-here"), // Uncomment to organize
		})
		if err != nil {
			return fmt.Errorf("failed to create hero asset: %w", err)
		}

		// Example 3: Bulk asset registration
		iconAssets := []struct {
			name     string
			fileName string
			fileHash string
		}{
			{"icon-home", "home.svg", "11111111111111111111111111111111"},
			{"icon-settings", "settings.svg", "22222222222222222222222222222222"},
			{"icon-user", "user.svg", "33333333333333333333333333333333"},
		}

		iconAssetIDs := []pulumi.StringOutput{}
		for _, icon := range iconAssets {
			asset, err := webflow.NewAsset(ctx, icon.name, &webflow.AssetArgs{
				SiteId:   siteID,
				FileName: pulumi.String(icon.fileName),
				FileHash: pulumi.String(icon.fileHash),
			})
			if err != nil {
				return fmt.Errorf("failed to create icon asset %s: %w", icon.name, err)
			}
			iconAssetIDs = append(iconAssetIDs, asset.ID().ToStringOutput())
		}

		// Export values for the logo asset
		// These are needed to complete the S3 upload
		ctx.Export("logoAssetId", logoAsset.AssetId)
		ctx.Export("logoUploadUrl", logoAsset.UploadUrl)
		ctx.Export("logoUploadDetails", logoAsset.UploadDetails)
		ctx.Export("logoAssetUrl", logoAsset.AssetUrl)
		ctx.Export("logoHostedUrl", logoAsset.HostedUrl)

		// Export hero asset info
		ctx.Export("heroAssetId", heroAsset.AssetId)
		ctx.Export("heroHostedUrl", heroAsset.HostedUrl)

		// Export icon asset IDs
		ctx.Export("iconAssetIds", pulumi.StringArray(iconAssetIDs))

		ctx.Log.Info(
			fmt.Sprintf("Registered %d assets. Use uploadUrl and uploadDetails to complete S3 uploads.", len(iconAssets)+2),
			&pulumi.LogOptions{},
		)

		return nil
	})
}
