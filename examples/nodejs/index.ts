import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

// Example: Configure robots.txt for a site
const myRobotsTxt = new webflow.RobotsTxt("myRobotsTxt", {
    siteId: "your-site-id-here",
    content: `User-agent: *
Allow: /`,
});

// Example: Create a redirect
const myRedirect = new webflow.Redirect("myRedirect", {
    siteId: "your-site-id-here",
    sourcePath: "/old-page",
    destinationPath: "/new-page",
    statusCode: 301,
});

export const robotsTxtId = myRobotsTxt.id;
export const redirectId = myRedirect.id;
