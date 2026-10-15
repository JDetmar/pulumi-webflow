# Create a 301 redirect

POST https://api.webflow.com/v2/sites/{site_id}/redirects
Content-Type: application/json

Add a new 301 redirection rule to a site.This endpoint allows you to define a source path (`fromUrl`) and its corresponding destination path (`toUrl`), which will dictate how traffic is rerouted on your site. This is useful for managing site changes, restructuring URLs, or handling outdated links.This endpoint requires an Enterprise workspace.Required scope: `sites:write`

Reference: https://developers.webflow.com/data/reference/enterprise/site-configuration/301-redirects/create

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `id` (string, optional) — The ID of the specific redirect rule
- `fromUrl` (string, optional) — The source URL path that will be redirected.
- `toUrl` (string, optional) — The target URL path where the user or client will be redirected.

## Response

### 200

Request was successful

- `id` (string, optional) — The ID of the specific redirect rule
- `fromUrl` (string, optional) — The source URL path that will be redirected.
- `toUrl` (string, optional) — The target URL path where the user or client will be redirected.

## Types

### SitesSiteIdRedirectsPostResponsesContentApplicationJsonSchemaDetailsItems
