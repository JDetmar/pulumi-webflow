using System.Collections.Generic;
using Pulumi;
using Webflow = Pulumi.Webflow;

return await Deployment.RunAsync(() => 
{
    // Example: Configure robots.txt for a site
    var myRobotsTxt = new Webflow.RobotsTxt("myRobotsTxt", new()
    {
        SiteId = "your-site-id-here",
        Content = @"User-agent: *
Allow: /",
    });

    // Example: Create a redirect
    var myRedirect = new Webflow.Redirect("myRedirect", new()
    {
        SiteId = "your-site-id-here",
        SourcePath = "/old-page",
        DestinationPath = "/new-page",
        StatusCode = 301,
    });

    return new Dictionary<string, object?>
    {
        ["robotsTxtId"] = myRobotsTxt.Id,
        ["redirectId"] = myRedirect.Id,
    };
});
