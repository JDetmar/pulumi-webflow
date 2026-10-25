# Add/Update Custom Code

PUT https://api.webflow.com/v2/pages/{page_id}/custom_code
Content-Type: application/json

Apply registered scripts to a page. If you have multiple scripts your App needs to apply or maintain on a page, ensure they are always included in the request body for this endpoint. To remove individual scripts, simply call this endpoint without the script in the request body.To apply a script to a page, the script must first be registered to a Site via the [Register Script](/data/reference/custom-code/custom-code/register-hosted) endpoints. Once registered, the script can be applied to a Site or Page using the appropriate endpoints. See the documentation on [working with Custom Code](/data/docs/custom-code) for more information.Access to this endpoint requires a bearer token obtained from an [OAuth Code Grant Flow](/data/reference/oauth-app).Required scope | `custom_code:write`

Reference: https://developers.webflow.com/data/reference/custom-code/custom-code-pages/upsert-custom-code

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `page_id` (string, required) — Unique identifier for a Page

### Body (application/json)

This endpoint expects an object.

- `scripts` (list of PagesPageIdCustomCodePutRequestBodyContentApplicationJsonSchemaScriptsItems, optional) — A list of scripts applied to a Site or a Page
- `lastUpdated` (string, optional) — Date when the Site's scripts were last updated
- `createdOn` (string, optional) — Date when the Site's scripts were created

## Response

### 200

Request was successful

- `scripts` (list of PagesPageIdCustomCodePutResponsesContentApplicationJsonSchemaScriptsItems, optional) — A list of scripts applied to a Site or a Page
- `lastUpdated` (string, optional) — Date when the Site's scripts were last updated
- `createdOn` (string, optional) — Date when the Site's scripts were created

## Types

### PagesPageIdCustomCodePutRequestBodyContentApplicationJsonSchemaScriptsItems

- `id` (string, required) — ID of the registered custom code script
- `location` (enum, required, default: header) — Location of the script, either in the header or footer of the published site
  - Allowed values: `header`, `footer`
- `version` (string, required) — Semantic Version String for the registered script *e.g. 0.0.1*
- `attributes` (PagesPageIdCustomCodePutRequestBodyContentApplicationJsonSchemaScriptsItemsAttributes, optional) — Developer-specified key/value pairs to be applied as attributes to the script

### PagesPageIdCustomCodePutResponsesContentApplicationJsonSchemaScriptsItems

- `id` (string, required) — ID of the registered custom code script
- `location` (enum, required, default: header) — Location of the script, either in the header or footer of the published site
  - Allowed values: `header`, `footer`
- `version` (string, required) — Semantic Version String for the registered script *e.g. 0.0.1*
- `attributes` (PagesPageIdCustomCodePutResponsesContentApplicationJsonSchemaScriptsItemsAttributes, optional) — Developer-specified key/value pairs to be applied as attributes to the script

### PagesPageIdCustomCodePutResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any

### PagesPageIdCustomCodePutRequestBodyContentApplicationJsonSchemaScriptsItemsAttributes

Developer-specified key/value pairs to be applied as attributes to the script

### PagesPageIdCustomCodePutResponsesContentApplicationJsonSchemaScriptsItemsAttributes

Developer-specified key/value pairs to be applied as attributes to the script
