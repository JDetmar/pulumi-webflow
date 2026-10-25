# Get Asset

GET https://api.webflow.com/v2/assets/{asset_id}

Get details about an asset

Required scope | `assets:read`

Reference: https://developers.webflow.com/data/reference/assets/assets/get

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `asset_id` (string, required) — Unique identifier for an Asset on a site

### Query parameters

- `localeId` (string, optional) — Unique identifier for a specific Locale. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)

## Response

### 200

Request was successful

- `id` (string, required) — Unique identifier for this asset
- `contentType` (string, required) — File format type
- `size` (integer, required) — size in bytes
- `siteId` (string, required) — Unique identifier for the site that hosts this asset
- `hostedUrl` (string, required) — Link to the asset
- `originalFileName` (string, required) — Original file name at the time of upload
- `displayName` (string, required) — Display name of the asset
- `lastUpdated` (string, required) — Date the asset metadata was last updated
- `createdOn` (string, required) — Date the asset metadata was created
- `variants` (list of AssetsAssetIdGetResponsesContentApplicationJsonSchemaVariantsItems, required) — A list of [asset variants](https://help.webflow.com/hc/en-us/articles/33961378697107-Responsive-images) created by Webflow to serve your site responsively.
- `altText` (string, required, nullable) — The visual description of the asset
- `folderId` (string, optional, nullable) — The ID of the folder the asset belongs to, or `null` if the asset is at the site root. This field is present only in list responses (`GET /sites/{site_id}/assets`).

## Types

### AssetsAssetIdGetResponsesContentApplicationJsonSchemaVariantsItems

Asset variant details

- `hostedUrl` (string, required) — URL of where the asset variant is hosted
- `originalFileName` (string, required) — Original file name of the variant
- `displayName` (string, required) — Display name of the variant
- `format` (string, required) — format of the variant
- `width` (integer, required) — Width in pixels
- `height` (integer, required, nullable) — Height in pixels
- `quality` (integer, required) — Value between 0 and 100 representing the image quality
- `error` (string, optional, nullable) — Any associated validation errors

### AssetsAssetIdGetResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
