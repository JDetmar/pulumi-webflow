# Delete 301 redirects

DELETE https://api.webflow.com/v2/sites/{site_id}/redirects/{redirect_id}

Remove a 301 redirection rule from a site.This is useful for cleaning up outdated or unnecessary redirects, ensuring that your site's routing behavior remains efficient and up-to-date.This endpoint requires an Enterprise workspace.Required scope: `sites:write`

Reference: https://developers.webflow.com/data/reference/enterprise/site-configuration/301-redirects/delete

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site
- `redirect_id` (string, required) — Unique identifier site redirect

## Response

### 200

Request was successful

- `redirects` (list of SitesSiteIdRedirectsRedirectIdDeleteResponsesContentApplicationJsonSchemaRedirectsItems, optional) — List of redirects for a given site
- `pagination` (SitesSiteIdRedirectsRedirectIdDeleteResponsesContentApplicationJsonSchemaPagination, optional) — Pagination object

## Types

### SitesSiteIdRedirectsRedirectIdDeleteResponsesContentApplicationJsonSchemaRedirectsItems

A single redirection rule, specifying a source URL and a destination URL.

- `id` (string, optional) — The ID of the specific redirect rule
- `fromUrl` (string, optional) — The source URL path that will be redirected.
- `toUrl` (string, optional) — The target URL path where the user or client will be redirected.

### SitesSiteIdRedirectsRedirectIdDeleteResponsesContentApplicationJsonSchemaPagination

Pagination object

- `limit` (integer, required) — The limit used for pagination
- `offset` (integer, required) — The offset used for pagination
- `total` (integer, required) — The total number of records

### SitesSiteIdRedirectsRedirectIdDeleteResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
