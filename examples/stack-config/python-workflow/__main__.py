# Copyright 2025, Pulumi Corporation.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import pulumi
import webflow_webflow as webflow

# Load configuration from current stack
config = pulumi.Config()

# Stack-specific configuration (defined in Pulumi.<stack>.yaml)
environment_name = config.require("environmentName")  # "dev", "staging", "prod"

# Validate environment configuration to prevent mistakes
valid_environments = ["dev", "staging", "prod"]
if environment_name not in valid_environments:
    raise ValueError(
        f"Invalid environment '{environment_name}'. "
        f"Must be one of: {', '.join(valid_environments)}"
    )

# Production safety check
is_prod = environment_name == "prod"
if is_prod:
    confirmation = config.get("prodDeploymentConfirmed")
    if confirmation != "yes":
        raise RuntimeError(
            "⚠️  Production deployment requires explicit confirmation.\n"
            "Run: pulumi config set prodDeploymentConfirmed yes\n"
            "This prevents accidental production deployments."
        )

# Load site configurations from stack config
# Each site is a distinct entity with its own purpose and settings
sites_config = config.require_object("sites")

pulumi.log.info(
    f"🚀 Deploying {len(sites_config)} sites to {environment_name} environment"
)

# Create each configured site
site_exports = {}

for site_key, site_config in sites_config.items():
    # Create the site with its specific configuration
    site = webflow.Site(
        site_key,
        display_name=site_config["displayName"],
        short_name=site_config["shortName"],
    )

    site_exports[f"{site_key}-id"] = site.id

    # Configure robots.txt based on site's indexing preference
    allow_indexing = site_config.get("allowIndexing", False)
    robots_content = (
        "User-agent: *\nAllow: /\n"
        if allow_indexing
        else f"User-agent: *\nDisallow: /\n\n# {environment_name.upper()} - Do not index"
    )

    webflow.RobotsTxt(
        f"{site_key}-robots",
        site_id=site.id,
        content=robots_content,
    )

    # Create any configured redirects for this site
    redirects = site_config.get("redirects", [])
    for redirect in redirects:
        redirect_name = f"{site_key}-redirect{redirect['sourcePath'].replace('/', '-')}"
        webflow.Redirect(
            redirect_name,
            site_id=site.id,
            source_path=redirect["sourcePath"],
            destination_path=redirect["destinationPath"],
            status_code=redirect["statusCode"],
        )

    pulumi.log.info(f"✅ Configured site: {site_key} ({site_config['displayName']})")

# Export site IDs for reference
for key, value in site_exports.items():
    pulumi.export(key, value)

# Export summary information
pulumi.export("environment", environment_name)
pulumi.export("siteCount", len(sites_config))
pulumi.export("siteNames", list(sites_config.keys()))
pulumi.export("stackName", pulumi.get_stack())
pulumi.export("projectName", pulumi.get_project())

pulumi.log.info(
    f"✅ Deployment complete: {len(sites_config)} sites configured for {environment_name}"
)
