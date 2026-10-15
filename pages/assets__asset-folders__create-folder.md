# Create Asset Folder

POST https://api.webflow.com/v2/sites/{site_id}/asset_folders
Content-Type: application/json

Create an Asset Folder within a given site

Required scope | `assets:write`

Reference: https://developers.webflow.com/data/reference/assets/asset-folders/create-folder

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `displayName` (string, required) — A human readable name for the Asset Folder
- `parentFolder` (string, optional) — An (optional) pointer to a parent Asset Folder (or null for root)

## Response

### 200

Request was successful

- `id` (string, optional) — Unique identifier for the Asset Folder
- `displayName` (string, optional) — User visible name for the Asset Folder
- `parentFolder` (string, optional) — Pointer to parent Asset Folder (or null if root)
- `assets` (list of string, optional) — Array of Asset instances in the folder
- `siteId` (string, optional) — The unique ID of the site the Asset Folder belongs to
- `createdOn` (string, optional) — Date that the Asset Folder was created on
- `lastUpdated` (string, optional) — Date that the Asset Folder was last updated on

## Types

### SitesSiteIdAssetFoldersPostResponsesContentApplicationJsonSchemaDetailsItems
