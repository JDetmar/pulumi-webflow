# Delete Custom Code

DELETE https://api.webflow.com/v2/pages/{page_id}/custom_code

Remove all scripts from a page applied by the App. This endpoint will not remove scripts from the site's registered scripts.To remove individual scripts applied by the App, use the [Add/Update Custom Code](/data/reference/custom-code/custom-code-pages/upsert-custom-code) endpoint.Access to this endpoint requires a bearer token obtained from an [OAuth Code Grant Flow](/data/reference/oauth-app).Required scope | `custom_code:write`

Reference: https://developers.webflow.com/data/reference/custom-code/custom-code-pages/delete-custom-code

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `page_id` (string, required) — Unique identifier for a Page

## Response

### 204

Request was successful. No Content is returned.

## Types

### PagesPageIdCustomCodeDeleteResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
