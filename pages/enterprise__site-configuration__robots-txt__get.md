# Get robots.txt

GET https://api.webflow.com/v2/sites/{site_id}/robots_txt

Retrieve the robots.txt configuration for various user agents.This endpoint requires an Enterprise workspace.Required scope: `site_config:read`

Reference: https://developers.webflow.com/data/reference/enterprise/site-configuration/robots-txt/get

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

Request was successful

- `rules` (list of SitesSiteIdRobotsTxtGetResponsesContentApplicationJsonSchemaRulesItems, optional) — List of rules for user agents.
- `sitemap` (string, optional) — URL to the sitemap.

## Types

### SitesSiteIdRobotsTxtGetResponsesContentApplicationJsonSchemaRulesItems

- `userAgent` (string, required) — The user agent the rules apply to.
- `allows` (list of string, optional) — List of paths allowed for this user agent.
- `disallows` (list of string, optional) — List of paths disallowed for this user agent.

### SitesSiteIdRobotsTxtGetResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
