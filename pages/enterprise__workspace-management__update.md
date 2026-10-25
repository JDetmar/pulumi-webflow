# Update Site

PATCH https://api.webflow.com/v2/sites/{site_id}
Content-Type: application/json

Update a site.This endpoint requires an Enterprise workspace.Required scope | `sites:write`

Reference: https://developers.webflow.com/data/reference/enterprise/workspace-management/update

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `name` (string, optional) — The name of the site
- `parentFolderId` (string, optional, nullable) — The parent folder ID of the site

## Response

### 200

Request was successful

- `id` (string, required) — Unique identifier for the Site
- `workspaceId` (string, optional) — Unique identifier for the Workspace
- `createdOn` (string, optional) — Date the Site was created
- `displayName` (string, optional) — Name given to Site
- `shortName` (string, optional) — Slugified version of name
- `lastPublished` (string, optional) — Date the Site was last published
- `lastUpdated` (string, optional) — Date the Site was last updated
- `previewUrl` (string, optional) — URL of a generated image for the given Site
- `timeZone` (string, optional) — Site timezone set under Site Settings
- `parentFolderId` (string, optional, nullable) — The ID of the parent folder the Site exists in
- `customDomains` (list of SitesSiteIdPatchResponsesContentApplicationJsonSchemaCustomDomainsItems, optional)
- `locales` (SitesSiteIdPatchResponsesContentApplicationJsonSchemaLocales, optional)
- `dataCollectionEnabled` (boolean, optional) — Indicates if data collection is enabled for the site.
- `dataCollectionType` (enum, optional) — The type of data collection enabled for the site.
  - Allowed values: `always`, `optOut`, `disabled`

## Types

### SitesSiteIdPatchResponsesContentApplicationJsonSchemaCustomDomainsItems

- `id` (string, required) — Unique identifier for the Domain
- `url` (string, optional) — The registered Domain name
- `lastPublished` (string, optional, nullable) — The date the custom domain was last published to

### SitesSiteIdPatchResponsesContentApplicationJsonSchemaLocales

- `primary` (SitesSiteIdPatchResponsesContentApplicationJsonSchemaLocalesPrimary, optional) — The primary locale for the site or application.
- `secondary` (list of SitesSiteIdPatchResponsesContentApplicationJsonSchemaLocalesSecondaryItems, optional) — A list of secondary locales available for the site or application.

### SitesSiteIdPatchResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any

### SitesSiteIdPatchResponsesContentApplicationJsonSchemaLocalesPrimary

The primary locale for the site or application.

- `id` (string, optional) — The unique identifier for the locale.
- `cmsLocaleId` (string, optional) — A CMS-specific identifier for the locale.
- `enabled` (boolean, optional) — Indicates if the locale is enabled.
- `displayName` (string, optional) — The display name of the locale, typically in English.
- `displayImageId` (string, optional, nullable) — An optional ID for an image associated with the locale, nullable.
- `redirect` (boolean, optional) — Determines if requests should redirect to the locale's subdirectory.
- `subdirectory` (string, optional) — The subdirectory path for the locale, used in URLs.
- `tag` (string, optional) — A tag or code representing the locale, often following a standard format like 'en-US'.

### SitesSiteIdPatchResponsesContentApplicationJsonSchemaLocalesSecondaryItems

- `id` (string, optional) — The unique identifier for the locale.
- `cmsLocaleId` (string, optional) — A CMS-specific identifier for the locale.
- `enabled` (boolean, optional) — Indicates if the locale is enabled.
- `displayName` (string, optional) — The display name of the locale, typically in English.
- `displayImageId` (string, optional, nullable) — An optional ID for an image associated with the locale, nullable.
- `redirect` (boolean, optional) — Determines if requests should redirect to the locale's subdirectory.
- `subdirectory` (string, optional) — The subdirectory path for the locale, used in URLs.
- `tag` (string, optional) — A tag or code representing the locale, often following a standard format like 'en-US'.
