# Upload Asset

POST https://api.webflow.com/v2/sites/{site_id}/assets
Content-Type: application/json

The first step in uploading an asset to a site.

This endpoint generates a response with the following information: `uploadUrl` and `uploadDetails`.

Use these properties in the header of a [POST request to Amazson s3](https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectPOST.html) to complete the upload.

To learn more about how to upload assets to Webflow, see our [assets guide](/data/docs/working-with-assets).

 Required scope | `assets:write`

Reference: https://developers.webflow.com/data/reference/assets/assets/create

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `fileName` (string, required) — File name including file extension. File names must be less than 100 characters.
- `fileHash` (string, required) — MD5 hash of the file
- `parentFolder` (string, optional) — ID of the Asset folder (optional)

## Response

### 200

Request was successful

- `uploadDetails` (SitesSiteIdAssetsPostResponsesContentApplicationJsonSchemaUploadDetails, optional) — Metadata for uploading the asset binary
- `contentType` (string, optional)
- `id` (string, optional)
- `parentFolder` (string, optional) — Parent folder for the asset
- `uploadUrl` (string, optional)
- `assetUrl` (string, optional) — S3 link to the asset
- `hostedUrl` (string, optional) — Represents the link to the asset
- `originalFileName` (string, optional) — Original file name when uploaded. If not specified at time of upload, it may be extracted from the raw file name
- `createdOn` (string, optional) — Date the asset metadata was created
- `lastUpdated` (string, optional) — Date the asset metadata was last updated

## Types

### SitesSiteIdAssetsPostResponsesContentApplicationJsonSchemaUploadDetails

Metadata for uploading the asset binary

- `acl` (string, optional)
- `bucket` (string, optional)
- `X-Amz-Algorithm` (string, optional)
- `X-Amz-Credential` (string, optional)
- `X-Amz-Date` (string, optional)
- `key` (string, optional)
- `Policy` (string, optional)
- `X-Amz-Signature` (string, optional)
- `success_action_status` (string, optional)
- `content-type` (string, optional)
- `Cache-Control` (string, optional)

### SitesSiteIdAssetsPostResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
