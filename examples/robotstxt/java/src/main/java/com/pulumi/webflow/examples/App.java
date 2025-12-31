package com.pulumi.webflow.examples;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.webflow.RobotsTxt;
import com.pulumi.webflow.RobotsTxtArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Get configuration values
            var config = ctx.config();
            var siteId = config.requireSecret("siteId");

            // Example 1: Allow All Crawlers (most common for public sites)
            var allowAllRobots = new RobotsTxt("allow-all-robots",
                RobotsTxtArgs.builder()
                    .siteId(siteId)
                    .content("User-agent: *\n" +
                            "Allow: /\n" +
                            "\n" +
                            "# Allow specific crawler access with no delays\n" +
                            "User-agent: Googlebot\n" +
                            "Allow: /\n" +
                            "Crawl-delay: 0\n" +
                            "\n" +
                            "User-agent: Bingbot\n" +
                            "Allow: /\n" +
                            "Crawl-delay: 1")
                    .build());

            // Example 2: Selective Blocking (for staging/development)
            var selectiveBlockRobots = new RobotsTxt("selective-block-robots",
                RobotsTxtArgs.builder()
                    .siteId(siteId)
                    .content("User-agent: *\n" +
                            "Allow: /\n" +
                            "\n" +
                            "# Disallow admin and private sections\n" +
                            "Disallow: /admin/\n" +
                            "Disallow: /private/\n" +
                            "Disallow: /staging/\n" +
                            "Disallow: /test/\n" +
                            "\n" +
                            "# Block specific crawlers\n" +
                            "User-agent: AhrefsBot\n" +
                            "Disallow: /\n" +
                            "\n" +
                            "User-agent: SemrushBot\n" +
                            "Disallow: /")
                    .build());

            // Example 3: Restrict Directories (protect API and backend)
            var restrictDirectoriesRobots = new RobotsTxt("restrict-directories-robots",
                RobotsTxtArgs.builder()
                    .siteId(siteId)
                    .content("User-agent: *\n" +
                            "Allow: /\n" +
                            "Disallow: /api/\n" +
                            "Disallow: /internal/\n" +
                            "Disallow: /*.json$\n" +
                            "Disallow: /*.xml$\n" +
                            "\n" +
                            "# Specify sitemap location\n" +
                            "Sitemap: https://example.com/sitemap.xml")
                    .build());

            // Export values for reference
            ctx.export("deployedSiteId", siteId);
            ctx.export("allowAllRobotsId", allowAllRobots.id());
            ctx.export("allowAllRobotsLastModified", allowAllRobots.lastModified());
            ctx.export("selectiveBlockRobotsId", selectiveBlockRobots.id());
            ctx.export("restrictDirectoriesRobotsId", restrictDirectoriesRobots.id());
        });
    }
}
