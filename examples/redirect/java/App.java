package com.pulumi.webflow.examples;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.webflow.Redirect;
import com.pulumi.webflow.RedirectArgs;

import java.util.List;
import java.util.ArrayList;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Get configuration values
            var config = ctx.config();
            var siteId = config.requireSecret("siteId");
            var environment = config.get("environment").orElse("development");

            // Example 1: Permanent Redirect (301) - Best for content moves
            // Use 301 redirects when content has permanently moved to preserve SEO value
            var permanentRedirect = new Redirect("old-blog-to-new-blog",
                RedirectArgs.builder()
                    .siteId(siteId)
                    .sourcePath("/blog/old-article")
                    .destinationPath("/blog/articles/updated-article")
                    .statusCode(301)
                    .build());

            // Example 2: Temporary Redirect (302) - Use for temporary changes
            // Use 302 redirects for seasonal content or A/B testing
            var temporaryRedirect = new Redirect("temporary-landing-page",
                RedirectArgs.builder()
                    .siteId(siteId)
                    .sourcePath("/old-campaign")
                    .destinationPath("/new-campaign-2025")
                    .statusCode(302)
                    .build());

            // Example 3: External Redirect (301) - Redirect to another domain
            // Useful for partner links or moved subdomains
            var externalRedirect = new Redirect("external-partner-link",
                RedirectArgs.builder()
                    .siteId(siteId)
                    .sourcePath("/partner")
                    .destinationPath("https://partner-site.com")
                    .statusCode(301)
                    .build());

            // Example 4: Bulk Redirects using Loop
            // Efficient pattern for migrating multiple URLs at once
            var redirectMappings = List.of(
                new String[]{"/product-a", "/products/product-a"},
                new String[]{"/product-b", "/products/product-b"},
                new String[]{"/product-c", "/products/product-c"}
            );

            var bulkRedirects = new ArrayList<Redirect>();
            for (int i = 0; i < redirectMappings.size(); i++) {
                var mapping = redirectMappings.get(i);
                var redirect = new Redirect("bulk-redirect-" + i,
                    RedirectArgs.builder()
                        .siteId(siteId)
                        .sourcePath(mapping[0])
                        .destinationPath(mapping[1])
                        .statusCode(301)
                        .build());
                bulkRedirects.add(redirect);
            }

            // Export values for reference
            ctx.export("deployedSiteId", siteId);
            ctx.export("permanentRedirectId", permanentRedirect.id());
            ctx.export("temporaryRedirectId", temporaryRedirect.id());
            ctx.export("externalRedirectId", externalRedirect.id());
            ctx.export("bulkRedirectCount", Output.of(bulkRedirects.size()));
        });
    }
}
