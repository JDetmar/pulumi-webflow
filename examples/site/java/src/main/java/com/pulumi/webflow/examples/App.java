package com.pulumi.webflow.examples;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.webflow.Site;
import com.pulumi.webflow.SiteArgs;

import java.util.List;
import java.util.ArrayList;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Get configuration values
            var config = ctx.config();
            var displayName = config.require("displayName");
            var shortName = config.require("shortName");
            var customDomain = config.get("customDomain");
            var timezone = config.get("timezone").orElse("America/New_York");

            // Example 1: Basic Site Creation
            // Create a simple site with required properties
            var basicSite = new Site("basic-site",
                SiteArgs.builder()
                    .displayName(displayName)
                    .shortName(shortName)
                    .timezone(timezone)
                    .build());

            // Example 2: Site with Custom Domain (conditional creation)
            // Only create if customDomain is provided in configuration
            Site siteWithDomain = null;
            if (customDomain.isPresent() && !customDomain.get().isEmpty()) {
                siteWithDomain = new Site("site-with-domain",
                    SiteArgs.builder()
                        .displayName(displayName + "-domain")
                        .shortName(shortName + "-domain")
                        .customDomain(customDomain.get())
                        .timezone(timezone)
                        .build());
            }

            // Example 3: Multi-Environment Site Configuration
            // Create sites for different environments using a loop
            var environments = List.of("development", "staging", "production");
            var environmentSites = new ArrayList<Site>();

            for (String env : environments) {
                var site = new Site("site-" + env,
                    SiteArgs.builder()
                        .displayName(displayName + "-" + env)
                        .shortName(shortName + "-" + env)
                        .timezone(timezone)
                        .build());
                environmentSites.add(site);
            }

            // Example 4: Site with Full Configuration
            // Demonstrates all available configuration options
            var configuredSite = new Site("configured-site",
                SiteArgs.builder()
                    .displayName(displayName + "-configured")
                    .shortName(shortName + "-configured")
                    .timezone(timezone)
                    .build());

            // Export values for reference
            ctx.export("basicSiteId", basicSite.id());
            ctx.export("basicSiteName", basicSite.displayName());
            ctx.export("customDomainSiteId",
                siteWithDomain != null ? siteWithDomain.id() : Output.of("not-created"));
            ctx.export("environmentSiteIds", Output.all(environmentSites.stream()
                .map(Site::id)
                .toList()));
            ctx.export("configuredSiteId", configuredSite.id());
        });
    }
}
