# Get Custom Code

GET https://api.webflow.com/v2/sites/{site_id}/custom_code

Get all scripts applied to a site by the App.To apply a script to a site or page, the script must first be registered to a site via the [Register Script](/data/reference/custom-code/custom-code/register-hosted) endpoints. Once registered, the script can be applied to a Site or Page using the appropriate endpoints. See the documentation on [working with Custom Code](/data/docs/custom-code) for more information.Access to this endpoint requires a bearer token obtained from an [OAuth Code Grant Flow](/data/reference/oauth-app).Required scope | `custom_code:read`

Reference: https://developers.webflow.com/data/reference/custom-code/custom-code-sites/get-custom-code

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

Request was successful

- `scripts` (list of SitesSiteIdCustomCodeGetResponsesContentApplicationJsonSchemaScriptsItems, optional) — A list of scripts applied to a Site or a Page
- `lastUpdated` (string, optional) — Date when the Site's scripts were last updated
- `createdOn` (string, optional) — Date when the Site's scripts were created

## Types

### SitesSiteIdCustomCodeGetResponsesContentApplicationJsonSchemaScriptsItems

- `id` (string, required) — ID of the registered custom code script
- `location` (enum, required, default: header) — Location of the script, either in the header or footer of the published site
  - Allowed values: `header`, `footer`
- `version` (string, required) — Semantic Version String for the registered script *e.g. 0.0.1*
- `attributes` (SitesSiteIdCustomCodeGetResponsesContentApplicationJsonSchemaScriptsItemsAttributes, optional) — Developer-specified key/value pairs to be applied as attributes to the script

### SitesSiteIdCustomCodeGetResponsesContentApplicationJsonSchemaDetailsItems

### SitesSiteIdCustomCodeGetResponsesContentApplicationJsonSchemaScriptsItemsAttributes

Developer-specified key/value pairs to be applied as attributes to the script
