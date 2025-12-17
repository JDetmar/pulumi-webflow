import pulumi
import pulumi_webflow as webflow

# Example: Configure robots.txt for a site
my_robots_txt = webflow.RobotsTxt("myRobotsTxt",
    site_id="your-site-id-here",
    content="""User-agent: *
Allow: /""")

# Example: Create a redirect
my_redirect = webflow.Redirect("myRedirect",
    site_id="your-site-id-here",
    source_path="/old-page",
    destination_path="/new-page",
    status_code=301)

pulumi.export("robots_txt_id", my_robots_txt.id)
pulumi.export("redirect_id", my_redirect.id)
