# List Asset Folders

GET https://api.webflow.com/v2/sites/{site_id}/asset_folders

List Asset Folders within a given site

Required scope | `assets:read`

Reference: https://developers.webflow.com/data/reference/assets/asset-folders/list-folders

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

Request was successful

- `assetFolders` (list of SitesSiteIdAssetFoldersGetResponsesContentApplicationJsonSchemaAssetFoldersItems, optional) — A list of Asset folders
- `pagination` (SitesSiteIdAssetFoldersGetResponsesContentApplicationJsonSchemaPagination, optional) — Pagination object

## Types

### SitesSiteIdAssetFoldersGetResponsesContentApplicationJsonSchemaAssetFoldersItems

Asset Folder details

- `id` (string, optional) — Unique identifier for the Asset Folder
- `displayName` (string, optional) — User visible name for the Asset Folder
- `parentFolder` (string, optional) — Pointer to parent Asset Folder (or null if root)
- `assets` (list of string, optional) — Array of Asset instances in the folder
- `siteId` (string, optional) — The unique ID of the site the Asset Folder belongs to
- `createdOn` (string, optional) — Date that the Asset Folder was created on
- `lastUpdated` (string, optional) — Date that the Asset Folder was last updated on

### SitesSiteIdAssetFoldersGetResponsesContentApplicationJsonSchemaPagination

Pagination object

- `limit` (integer, required) — The limit used for pagination
- `offset` (integer, required) — The offset used for pagination
- `total` (integer, required) — The total number of records

### SitesSiteIdAssetFoldersGetResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
