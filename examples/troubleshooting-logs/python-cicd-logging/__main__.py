#!/usr/bin/env python3
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

import os
import pulumi
import pulumi_webflow as webflow

# Detect CI/CD environment
is_ci = os.getenv("CI") == "true"
environment = os.getenv("PULUMI_STACK", "unknown")

# Configure logging based on environment
if is_ci:
    pulumi.log.info(f"🤖 Running in CI/CD environment: {environment}")
    pulumi.log.debug("Verbose logging enabled for CI/CD troubleshooting")
else:
    pulumi.log.info(f"💻 Running in local environment: {environment}")

# Always log credential source (without exposing values)
token_source = "environment" if os.getenv("WEBFLOW_API_TOKEN") else "pulumi_config"
pulumi.log.info(f"🔐 Using API token from: {token_source}")
pulumi.log.debug("API credentials are redacted from all log output")

# Production: minimal logging
# Development: verbose logging
log_level = "info" if environment == "prod" else "debug"
pulumi.log.info(f"📊 Log level: {log_level}")

# Create a Webflow site for CI/CD environment
pulumi.log.info("🚀 Creating Webflow site for CI/CD pipeline")

site = webflow.Site(
    "cicd-logging-site",
    display_name="CI/CD Logging Example Site",
    short_name="cicd-logging-demo",
)

pulumi.log.debug(f"Site creation request submitted (details redacted)")

# Configure robots.txt with environment-specific settings
site_id = site.id.apply(lambda id: id)

def configure_robots(site_id):
    pulumi.log.info(f"🤖 Configuring robots.txt for site: {site_id}")

    # Environment-specific robots.txt content
    if environment == "prod":
        robots_content = "User-agent: *\nAllow: /\n"
        pulumi.log.debug("Using production robots.txt (allow all)")
    else:
        robots_content = f"User-agent: *\nDisallow: /\n\n# {environment.upper()} ENVIRONMENT - NOT FOR INDEXING"
        pulumi.log.debug(f"Using {environment} robots.txt (disallow all)")

    robots = webflow.RobotsTxt(
        f"{environment}-robots",
        site_id=site_id,
        content=robots_content,
    )

    pulumi.log.info(f"✅ Robots.txt configured for {environment} environment")
    return robots

robots = site_id.apply(configure_robots)

# Export results
pulumi.export("site_id", site.id)
pulumi.export("environment", environment)
pulumi.export("is_ci", is_ci)
pulumi.log.info("📤 Exported stack outputs for CI/CD pipeline")
