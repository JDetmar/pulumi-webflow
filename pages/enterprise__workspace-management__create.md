# Create Site

POST https://api.webflow.com/v2/workspaces/{workspace_id}/sites
Content-Type: application/json

Create a site.This endpoint requires an Enterprise workspace.Required scope | `workspace:write`

Reference: https://developers.webflow.com/data/reference/enterprise/workspace-management/create

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `workspace_id` (string, required) — Unique identifier for a Workspace. Find it with [List Workspaces](/data/reference/enterprise/workspace-management/list) (workspace API tokens and Data Client apps), or with `authorization.authorizedTo.workspaceIds` from [Get Authorization Info](/data/reference/token/introspect). With a site token, read `workspaceId` from [Get Site](/data/reference/sites/get).

### Body (application/json)

This endpoint expects an object.

- `name` (string, required) — The name of the site
- `templateName` (string, optional) — The workspace or marketplace template to use
- `parentFolderId` (string, optional, nullable) — MegaDodo Publications - Potential Book Ideas

## Response

### 201

Request was successful

- `id` (string, required) — Unique identifier for the Site
- `workspaceId` (string, required) — Unique identifier for the Workspace
- `createdOn` (string, required) — Date the Site was created
- `displayName` (string, required) — Name given to Site
- `shortName` (string, required) — Slugified version of name
- `lastPublished` (string, required) — Date the Site was last published
- `lastUpdated` (string, required) — Date the Site was last updated
- `previewUrl` (string, required) — URL of a generated image for the given Site
- `timeZone` (string, required) — Site timezone set under Site Settings
- `dataCollectionEnabled` (boolean, required) — Indicates if data collection is enabled for the site.
- `dataCollectionType` (enum, required) — The type of data collection enabled for the site.
  - Allowed values: `always`, `optOut`, `disabled`
- `parentFolderId` (string, optional, nullable) — The ID of the parent folder the Site exists in
- `customDomains` (list of WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaCustomDomainsItems, optional)
- `locales` (WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaLocales, optional)

## Types

### WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaCustomDomainsItems

- `id` (string, required) — Unique identifier for the Domain
- `url` (string, optional) — The registered Domain name
- `lastPublished` (string, optional, nullable) — The date the custom domain was last published to

### WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaLocales

- `primary` (WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaLocalesPrimary, optional) — The primary locale for the site or application.
- `secondary` (list of WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaLocalesSecondaryItems, optional) — A list of secondary locales available for the site or application.

### WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any

### WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaLocalesPrimary

The primary locale for the site or application.

- `id` (string, optional) — The unique identifier for the locale.
- `cmsLocaleId` (string, optional) — A CMS-specific identifier for the locale.
- `enabled` (boolean, optional) — Indicates if the locale is enabled.
- `displayName` (string, optional) — The display name of the locale, typically in English.
- `displayImageId` (string, optional, nullable) — An optional ID for an image associated with the locale, nullable.
- `redirect` (boolean, optional) — Determines if requests should redirect to the locale's subdirectory.
- `subdirectory` (string, optional) — The subdirectory path for the locale, used in URLs.
- `tag` (string, optional) — A tag or code representing the locale, often following a standard format like 'en-US'.

### WorkspacesWorkspaceIdSitesPostResponsesContentApplicationJsonSchemaLocalesSecondaryItems

- `id` (string, optional) — The unique identifier for the locale.
- `cmsLocaleId` (string, optional) — A CMS-specific identifier for the locale.
- `enabled` (boolean, optional) — Indicates if the locale is enabled.
- `displayName` (string, optional) — The display name of the locale, typically in English.
- `displayImageId` (string, optional, nullable) — An optional ID for an image associated with the locale, nullable.
- `redirect` (boolean, optional) — Determines if requests should redirect to the locale's subdirectory.
- `subdirectory` (string, optional) — The subdirectory path for the locale, used in URLs.
- `tag` (string, optional) — A tag or code representing the locale, often following a standard format like 'en-US'.
