# Get 301 redirects

GET https://api.webflow.com/v2/sites/{site_id}/redirects

Fetch a list of all 301 redirect rules configured for a specific site.Use this endpoint to review, audit, or manage the redirection rules that control how traffic is rerouted on your site.This endpoint requires an Enterprise workspace.Required scope: `sites:read`

Reference: https://developers.webflow.com/data/reference/enterprise/site-configuration/301-redirects/get

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

Request was successful

- `redirects` (list of SitesSiteIdRedirectsGetResponsesContentApplicationJsonSchemaRedirectsItems, optional) — List of redirects for a given site
- `pagination` (SitesSiteIdRedirectsGetResponsesContentApplicationJsonSchemaPagination, optional) — Pagination object

## Types

### SitesSiteIdRedirectsGetResponsesContentApplicationJsonSchemaRedirectsItems

A single redirection rule, specifying a source URL and a destination URL.

- `id` (string, optional) — The ID of the specific redirect rule
- `fromUrl` (string, optional) — The source URL path that will be redirected.
- `toUrl` (string, optional) — The target URL path where the user or client will be redirected.

### SitesSiteIdRedirectsGetResponsesContentApplicationJsonSchemaPagination

Pagination object

- `limit` (integer, required) — The limit used for pagination
- `offset` (integer, required) — The offset used for pagination
- `total` (integer, required) — The total number of records

### SitesSiteIdRedirectsGetResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
