# Get Ecommerce Settings

GET https://api.webflow.com/v2/sites/{site_id}/ecommerce/settings

Retrieve ecommerce settings for a site.

Required scope | `ecommerce:read`

Reference: https://developers.webflow.com/data/reference/ecommerce/settings/get-settings

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

Request was successful

- `siteId` (string, optional) — The identifier of the Site
- `createdOn` (string, optional) — Date that the Site was created on
- `defaultCurrency` (string, optional) — The three-letter ISO currency code for the Site

## Types

### SitesSiteIdEcommerceSettingsGetResponsesContentApplicationJsonSchemaDetailsItems
