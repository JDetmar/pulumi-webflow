# Get Registered Scripts

GET https://api.webflow.com/v2/sites/{site_id}/registered_scripts

Get a list of scripts that have been registered to a site. A site can have a maximum of 800 registered scripts.To apply a script to a site or page, the script must first be registered to a site via the [Register Script](/data/reference/custom-code/custom-code/register-hosted) endpoints. Once registered, the script can be applied to a Site or Page using the appropriate endpoints. See the documentation on [working with Custom Code](/data/docs/custom-code) for more information.Access to this endpoint requires a bearer token obtained from an [OAuth Code Grant Flow](/data/reference/oauth-app).Required scope | `custom_code:read`

Reference: https://developers.webflow.com/data/reference/custom-code/custom-code/list

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

Request was successful

- `registeredScripts` (list of SitesSiteIdRegisteredScriptsGetResponsesContentApplicationJsonSchemaRegisteredScriptsItems, optional)
- `pagination` (SitesSiteIdRegisteredScriptsGetResponsesContentApplicationJsonSchemaPagination, optional) — Pagination object

## Types

### SitesSiteIdRegisteredScriptsGetResponsesContentApplicationJsonSchemaRegisteredScriptsItems

Registered custom code for application

- `id` (string, optional) — Human readable id, derived from the user-specified display name
- `canCopy` (boolean, optional, default: false) — Define whether the script can be copied on site duplication and transfer
- `displayName` (string, optional) — User-facing name for the script. Must be between 1 and 50 alphanumeric characters
- `hostedLocation` (string, optional) — URI for an externally hosted script location
- `integrityHash` (string, optional) — Sub-Resource Integrity Hash. Only required for externally hosted scripts (passed via hostedLocation)
- `createdOn` (string, optional) — Timestamp when the script version was created
- `lastUpdated` (string, optional) — Timestamp when the script version was last updated
- `version` (string, optional) — A Semantic Version (SemVer) string, denoting the version of the script

### SitesSiteIdRegisteredScriptsGetResponsesContentApplicationJsonSchemaPagination

Pagination object

- `limit` (integer, required) — The limit used for pagination
- `offset` (integer, required) — The offset used for pagination
- `total` (integer, required) — The total number of records

### SitesSiteIdRegisteredScriptsGetResponsesContentApplicationJsonSchemaDetailsItems
