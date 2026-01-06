using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Pulumi;
using Pulumi.Webflow;

class Program
{
    static Task<int> Main() => Deployment.RunAsync(() =>
    {
        // Create a Pulumi config object
        var config = new Config();

        // Get configuration values
        var siteId = config.RequireSecret("siteId");

        /**
         * Asset Example - Creating and Managing Webflow Assets
         *
         * This example demonstrates how to register assets with Webflow.
         * After registration, you'll receive uploadUrl and uploadDetails
         * to complete the file upload to S3.
         *
         * Two-step process:
         * 1. Register asset metadata (this provider handles this)
         * 2. Upload file to S3 using uploadUrl + uploadDetails (done separately)
         */

        // Example 1: Register a single asset
        // The fileHash is the MD5 hash of your file content
        // Generate with: md5sum logo.png (Linux) or md5 -q logo.png (macOS)
        var logoAsset = new Asset("company-logo", new AssetArgs
        {
            SiteId = siteId,
            FileName = "logo.png",
            FileHash = "d41d8cd98f00b204e9800998ecf8427e",
        });

        // Example 2: Asset with folder organization
        var heroAsset = new Asset("hero-image", new AssetArgs
        {
            SiteId = siteId,
            FileName = "hero-banner.jpg",
            FileHash = "a1b2c3d4e5f6789012345678abcdef12",
            // ParentFolder = "folder-id-here", // Uncomment to organize in a folder
        });

        // Example 3: Bulk asset registration
        var iconAssets = new List<Asset>();
        var icons = new[]
        {
            new { Name = "icon-home", FileName = "home.svg", FileHash = "11111111111111111111111111111111" },
            new { Name = "icon-settings", FileName = "settings.svg", FileHash = "22222222222222222222222222222222" },
            new { Name = "icon-user", FileName = "user.svg", FileHash = "33333333333333333333333333333333" },
        };

        foreach (var icon in icons)
        {
            var asset = new Asset(icon.Name, new AssetArgs
            {
                SiteId = siteId,
                FileName = icon.FileName,
                FileHash = icon.FileHash,
            });
            iconAssets.Add(asset);
        }

        // Export values for the logo asset
        // These are needed to complete the S3 upload
        var outputs = new Dictionary<string, object?>
        {
            ["logoAssetId"] = logoAsset.AssetId,
            ["logoUploadUrl"] = logoAsset.UploadUrl,
            ["logoUploadDetails"] = logoAsset.UploadDetails,
            ["logoAssetUrl"] = logoAsset.AssetUrl,
            ["logoHostedUrl"] = logoAsset.HostedUrl,

            // Export hero asset info
            ["heroAssetId"] = heroAsset.AssetId,
            ["heroHostedUrl"] = heroAsset.HostedUrl,

            // Export icon asset IDs
            ["iconAssetIds"] = Output.All(iconAssets.Select(a => a.AssetId).ToArray()),
        };

        // Print deployment message
        var assetCount = icons.Length + 2;
        siteId.Apply(s => Console.WriteLine($"Registered {assetCount} assets for site {s}. Use uploadUrl and uploadDetails to complete S3 uploads."));

        return outputs;
    });
}
