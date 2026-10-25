# Delete robots.txt

DELETE https://api.webflow.com/v2/sites/{site_id}/robots_txt
Content-Type: application/json

Remove specific rules for a user-agent in your `robots.txt` file. To delete all rules for a user-agent, provide an empty rule set. This will remove the user-agent's entry entirely, leaving it subject to your site's default crawling behavior.**Note:** Deleting a user-agent with no rules will make the user-agent's access unrestricted unless other directives apply.This endpoint requires an Enterprise workspace.Required scope: `site_config:write`

Reference: https://developers.webflow.com/data/reference/enterprise/site-configuration/robots-txt/delete

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `rules` (list of SitesSiteIdRobotsTxtDeleteRequestBodyContentApplicationJsonSchemaRulesItems, optional) — List of rules for user agents.
- `sitemap` (string, optional) — URL to the sitemap.

## Response

### 200

Request was successful

- `rules` (list of SitesSiteIdRobotsTxtDeleteResponsesContentApplicationJsonSchemaRulesItems, optional) — List of rules for user agents.
- `sitemap` (string, optional) — URL to the sitemap.

## Types

### SitesSiteIdRobotsTxtDeleteRequestBodyContentApplicationJsonSchemaRulesItems

- `userAgent` (string, required) — The user agent the rules apply to.
- `allows` (list of string, optional) — List of paths allowed for this user agent.
- `disallows` (list of string, optional) — List of paths disallowed for this user agent.

### SitesSiteIdRobotsTxtDeleteResponsesContentApplicationJsonSchemaRulesItems

- `userAgent` (string, required) — The user agent the rules apply to.
- `allows` (list of string, optional) — List of paths allowed for this user agent.
- `disallows` (list of string, optional) — List of paths disallowed for this user agent.

### SitesSiteIdRobotsTxtDeleteResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
