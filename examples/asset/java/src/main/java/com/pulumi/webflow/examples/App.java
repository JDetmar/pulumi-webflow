package com.pulumi.webflow.examples;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.webflow.Asset;
import com.pulumi.webflow.AssetArgs;

import java.util.List;
import java.util.ArrayList;
import java.util.stream.Collectors;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Get configuration values
            var config = ctx.config();
            var siteId = config.requireSecret("siteId");

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
            var logoAsset = new Asset("company-logo",
                AssetArgs.builder()
                    .siteId(siteId)
                    .fileName("logo.png")
                    .fileHash("d41d8cd98f00b204e9800998ecf8427e")
                    .build());

            // Example 2: Asset with folder organization
            var heroAsset = new Asset("hero-image",
                AssetArgs.builder()
                    .siteId(siteId)
                    .fileName("hero-banner.jpg")
                    .fileHash("a1b2c3d4e5f6789012345678abcdef12")
                    // .parentFolder("folder-id-here") // Uncomment to organize in a folder
                    .build());

            // Example 3: Bulk asset registration
            var icons = List.of(
                new String[]{"icon-home", "home.svg", "11111111111111111111111111111111"},
                new String[]{"icon-settings", "settings.svg", "22222222222222222222222222222222"},
                new String[]{"icon-user", "user.svg", "33333333333333333333333333333333"}
            );

            var iconAssets = new ArrayList<Asset>();
            for (var icon : icons) {
                var asset = new Asset(icon[0],
                    AssetArgs.builder()
                        .siteId(siteId)
                        .fileName(icon[1])
                        .fileHash(icon[2])
                        .build());
                iconAssets.add(asset);
            }

            // Export values for the logo asset
            // These are needed to complete the S3 upload
            ctx.export("logoAssetId", logoAsset.assetId());
            ctx.export("logoUploadUrl", logoAsset.uploadUrl());
            ctx.export("logoUploadDetails", logoAsset.uploadDetails());
            ctx.export("logoAssetUrl", logoAsset.assetUrl());
            ctx.export("logoHostedUrl", logoAsset.hostedUrl());

            // Export hero asset info
            ctx.export("heroAssetId", heroAsset.assetId());
            ctx.export("heroHostedUrl", heroAsset.hostedUrl());

            // Export icon asset IDs
            ctx.export("iconAssetIds", Output.all(iconAssets.stream()
                .map(Asset::assetId)
                .collect(Collectors.toList())));

            // Print deployment message
            int assetCount = icons.size() + 2;
            siteId.applyValue(s -> {
                System.out.println(String.format("Registered %d assets for site %s. Use uploadUrl and uploadDetails to complete S3 uploads.", assetCount, s));
                return null;
            });
        });
    }
}
