# Update 301 redirect

PATCH https://api.webflow.com/v2/sites/{site_id}/redirects/{redirect_id}
Content-Type: application/json

Update a 301 redirection rule from a site.This endpoint requires an Enterprise workspace.Required scope: `sites:write`

Reference: https://developers.webflow.com/data/reference/enterprise/site-configuration/301-redirects/patch

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site
- `redirect_id` (string, required) — Unique identifier site redirect

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

### SitesSiteIdRedirectsRedirectIdPatchResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
