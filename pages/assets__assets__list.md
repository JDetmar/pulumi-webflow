# List Assets

GET https://api.webflow.com/v2/sites/{site_id}/assets

List of assets uploaded to a site

Required scope | `assets:read`

Reference: https://developers.webflow.com/data/reference/assets/assets/list

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Query parameters

- `localeId` (string, optional) — Unique identifier for a specific Locale. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)
- `offset` (integer, optional) — Offset used for pagination if the results have more than limit records
- `limit` (integer, optional) — Maximum number of records to be returned (max limit: 100)
- `folderId` (string, optional) — Filter assets to those in the specified folder and all descendant folders. Must be a 24-character hex ObjectId.

## Response

### 200

Request was successful

- `assets` (list of SitesSiteIdAssetsGetResponsesContentApplicationJsonSchemaAssetsItems, required)
- `pagination` (SitesSiteIdAssetsGetResponsesContentApplicationJsonSchemaPagination, required) — Pagination object

## Types

### SitesSiteIdAssetsGetResponsesContentApplicationJsonSchemaAssetsItems

Asset details

- `id` (string, required) — Unique identifier for this asset
- `contentType` (string, required) — File format type
- `size` (integer, required) — size in bytes
- `siteId` (string, required) — Unique identifier for the site that hosts this asset
- `hostedUrl` (string, required) — Link to the asset
- `originalFileName` (string, required) — Original file name at the time of upload
- `displayName` (string, required) — Display name of the asset
- `lastUpdated` (string, required) — Date the asset metadata was last updated
- `createdOn` (string, required) — Date the asset metadata was created
- `variants` (list of SitesSiteIdAssetsGetResponsesContentApplicationJsonSchemaAssetsItemsVariantsItems, required) — A list of [asset variants](https://help.webflow.com/hc/en-us/articles/33961378697107-Responsive-images) created by Webflow to serve your site responsively.
- `altText` (string, required, nullable) — The visual description of the asset
- `folderId` (string, optional, nullable) — The ID of the folder the asset belongs to, or `null` if the asset is at the site root. This field is present only in list responses (`GET /sites/{site_id}/assets`).

### SitesSiteIdAssetsGetResponsesContentApplicationJsonSchemaPagination

Pagination object

- `limit` (integer, required) — The limit used for pagination
- `offset` (integer, required) — The offset used for pagination
- `total` (integer, required) — The total number of records

### SitesSiteIdAssetsGetResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any

### SitesSiteIdAssetsGetResponsesContentApplicationJsonSchemaAssetsItemsVariantsItems

Asset variant details

- `hostedUrl` (string, required) — URL of where the asset variant is hosted
- `originalFileName` (string, required) — Original file name of the variant
- `displayName` (string, required) — Display name of the variant
- `format` (string, required) — format of the variant
- `width` (integer, required) — Width in pixels
- `height` (integer, required, nullable) — Height in pixels
- `quality` (integer, required) — Value between 0 and 100 representing the image quality
- `error` (string, optional, nullable) — Any associated validation errors
