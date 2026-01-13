package com.pulumi.webflow.examples;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.webflow.Site;
import com.pulumi.webflow.SiteArgs;

import java.util.List;
import java.util.ArrayList;
import java.util.stream.Collectors;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Get configuration values
            var config = ctx.config();
            var workspaceId = config.require("workspaceId");
            var displayName = config.require("displayName");
            var shortName = config.require("shortName");
            var timezone = config.get("timezone").orElse("America/New_York");

            // Example 1: Basic Site Creation
            // Create a simple site with required properties
            var basicSite = new Site("basic-site",
                SiteArgs.builder()
                    .workspaceId(workspaceId)
                    .displayName(displayName)
                    .shortName(shortName)
                    .timeZone(timezone)
                    .build());

            // Example 2: Multi-Environment Site Configuration
            // Create sites for different environments using a loop
            var environments = List.of("development", "staging", "production");
            var environmentSites = new ArrayList<Site>();

            for (String env : environments) {
                var site = new Site("site-" + env,
                    SiteArgs.builder()
                        .workspaceId(workspaceId)
                        .displayName(displayName + "-" + env)
                        .shortName(shortName + "-" + env)
                        .timeZone(timezone)
                        .build());
                environmentSites.add(site);
            }

            // Example 3: Site with Full Configuration
            // Demonstrates all available configuration options
            var configuredSite = new Site("configured-site",
                SiteArgs.builder()
                    .workspaceId(workspaceId)
                    .displayName(displayName + "-configured")
                    .shortName(shortName + "-configured")
                    .timeZone(timezone)
                    .build());

            // Export values for reference
            ctx.export("basicSiteId", basicSite.id());
            ctx.export("basicSiteName", basicSite.displayName());
            ctx.export("environmentSiteIds", Output.all(environmentSites.stream()
                .map(Site::id)
                .collect(Collectors.toList())));
            ctx.export("configuredSiteId", configuredSite.id());
        });
    }
}
