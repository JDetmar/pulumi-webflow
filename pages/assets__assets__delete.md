# Delete Asset

DELETE https://api.webflow.com/v2/assets/{asset_id}

Delete an Asset

Required Scope: `assets: write`

Reference: https://developers.webflow.com/data/reference/assets/assets/delete

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `asset_id` (string, required) — Unique identifier for an Asset on a site

## Response

### 204

Request was successful. No Content is returned.

## Types

### AssetsAssetIdDeleteResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
