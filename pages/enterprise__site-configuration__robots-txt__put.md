# Replace robots.txt

PUT https://api.webflow.com/v2/sites/{site_id}/robots_txt
Content-Type: application/json

Replace the `robots.txt` configuration for various user agents.This endpoint requires an Enterprise workspace.Required scope | `site_config:write`

Reference: https://developers.webflow.com/data/reference/enterprise/site-configuration/robots-txt/put

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `rules` (list of SitesSiteIdRobotsTxtPutRequestBodyContentApplicationJsonSchemaRulesItems, optional) — List of rules for user agents.
- `sitemap` (string, optional) — URL to the sitemap.

## Response

### 200

Request was successful

- `rules` (list of SitesSiteIdRobotsTxtPutResponsesContentApplicationJsonSchemaRulesItems, optional) — List of rules for user agents.
- `sitemap` (string, optional) — URL to the sitemap.

## Types

### SitesSiteIdRobotsTxtPutRequestBodyContentApplicationJsonSchemaRulesItems

- `userAgent` (string, required) — The user agent the rules apply to.
- `allows` (list of string, optional) — List of paths allowed for this user agent.
- `disallows` (list of string, optional) — List of paths disallowed for this user agent.

### SitesSiteIdRobotsTxtPutResponsesContentApplicationJsonSchemaRulesItems

- `userAgent` (string, required) — The user agent the rules apply to.
- `allows` (list of string, optional) — List of paths allowed for this user agent.
- `disallows` (list of string, optional) — List of paths disallowed for this user agent.

### SitesSiteIdRobotsTxtPutResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
