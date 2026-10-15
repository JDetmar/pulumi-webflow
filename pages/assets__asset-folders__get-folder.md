# Get Asset Folder

GET https://api.webflow.com/v2/asset_folders/{asset_folder_id}

Get details about a specific Asset Folder

Required scope | `assets:read`

Reference: https://developers.webflow.com/data/reference/assets/asset-folders/get-folder

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `asset_folder_id` (string, required) — Unique identifier for an Asset Folder

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

### AssetFoldersAssetFolderIdGetResponsesContentApplicationJsonSchemaDetailsItems
