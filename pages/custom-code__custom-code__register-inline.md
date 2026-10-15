# Register Script - Inline

POST https://api.webflow.com/v2/sites/{site_id}/registered_scripts/inline
Content-Type: application/json

Register an inline script to a site. Inline scripts are limited to 2000 characters.To apply a script to a site or page, the script must first be registered to a site via the [Register Script](/data/reference/custom-code/custom-code/register-hosted) endpoints. Once registered, the script can be applied to a Site or Page using the appropriate endpoints. See the documentation on [working with Custom Code](/data/docs/custom-code) for more information.Access to this endpoint requires a bearer token obtained from an [OAuth Code Grant Flow](/data/reference/oauth-app).Required scope | `custom_code:write`

Reference: https://developers.webflow.com/data/reference/custom-code/custom-code/register-inline

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `sourceCode` (string, required) — The code to be added to the site (to be hosted by Webflow).
- `version` (string, required) — A Semantic Version (SemVer) string, denoting the version of the script
- `displayName` (string, required) — User-facing name for the script. Must be between 1 and 50 alphanumeric characters
- `integrityHash` (string, optional) — Sub-Resource Integrity Hash. Only required for externally hosted scripts (passed via hostedLocation)
- `canCopy` (boolean, optional, default: false) — Define whether the script can be copied on site duplication and transfer

## Response

### 201

Created

- `id` (string, optional) — Human readable id, derived from the user-specified display name
- `canCopy` (boolean, optional, default: false) — Define whether the script can be copied on site duplication and transfer
- `displayName` (string, optional) — User-facing name for the script. Must be between 1 and 50 alphanumeric characters
- `hostedLocation` (string, optional) — URI for an externally hosted script location
- `integrityHash` (string, optional) — Sub-Resource Integrity Hash. Only required for externally hosted scripts (passed via hostedLocation)
- `createdOn` (string, optional) — Timestamp when the script version was created
- `lastUpdated` (string, optional) — Timestamp when the script version was last updated
- `version` (string, optional) — A Semantic Version (SemVer) string, denoting the version of the script

## Types

### SitesSiteIdRegisteredScriptsInlinePostResponsesContentApplicationJsonSchemaDetailsItems
