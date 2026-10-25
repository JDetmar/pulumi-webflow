# Publish Site

POST https://api.webflow.com/v2/sites/{site_id}/publish
Content-Type: application/json

Publishes a site or an individual page to one or more domains.
If multiple individual pages are published to staging, publishing from staging to production publishes all staged changes.To publish to a specific custom domain, use the domain IDs from the [Get Custom Domains](/data/reference/sites/get-custom-domain) endpoint.You must include at least one of the `customDomains` or `publishToWebflowSubdomain` properties in the request body.To publish an individual page instead of the entire site, provide the ID of the page in the `pageId` parameter.This endpoint has a specific rate limit of one successful publish queue per minute.Required scope | `sites:write`

Reference: https://developers.webflow.com/data/reference/sites/publish

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `customDomains` (list of string, optional) — Array of Custom Domain IDs to publish
- `publishToWebflowSubdomain` (boolean, optional, default: false) — Choice of whether to publish to the default Webflow Subdomain
- `pageId` (string, optional) — The ID of the page to publish

## Response

### 202

Request accepted

- `customDomains` (list of SitesSiteIdPublishPostResponsesContentApplicationJsonSchemaCustomDomainsItems, optional) — Array of domains objects
- `publishToWebflowSubdomain` (boolean, optional, default: false) — Flag for publishing to webflow.io subdomain
- `publishScope` (enum, optional) — Whether the site or an individual page was published
  - Allowed values: `site`, `page`

## Types

### SitesSiteIdPublishPostResponsesContentApplicationJsonSchemaCustomDomainsItems

- `id` (string, required) — Unique identifier for the Domain
- `url` (string, optional) — The registered Domain name
- `lastPublished` (string, optional, nullable) — The date the custom domain was last published to

### SitesSiteIdPublishPostResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
