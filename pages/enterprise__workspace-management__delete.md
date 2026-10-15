# Delete Site

DELETE https://api.webflow.com/v2/sites/{site_id}

Delete a site.This endpoint requires an Enterprise workspace.Required scope | `sites:write`

Reference: https://developers.webflow.com/data/reference/enterprise/workspace-management/delete

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 204

Request was successful. No Content is returned.

## Types

### SitesSiteIdDeleteResponsesContentApplicationJsonSchemaDetailsItems
