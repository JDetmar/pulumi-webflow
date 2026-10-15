# Delete Custom Code

DELETE https://api.webflow.com/v2/sites/{site_id}/custom_code

Remove all scripts from a site applied by the App. This endpoint will not remove scripts from the site's registered scripts.To remove individual scripts applied by the App, use the [Add/Update Custom Code](/data/reference/custom-code/custom-code-sites/upsert-custom-code) endpoint.Access to this endpoint requires a bearer token obtained from an [OAuth Code Grant Flow](/data/reference/oauth-app).Required scope | `custom_code:write`

Reference: https://developers.webflow.com/data/reference/custom-code/custom-code-sites/delete-custom-code

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 204

Request was successful. No Content is returned.

## Types

### SitesSiteIdCustomCodeDeleteResponsesContentApplicationJsonSchemaDetailsItems
