# List Collections

GET https://api.webflow.com/v2/sites/{site_id}/collections

List of all Collections within a Site.

Required scope | `cms:read`

Reference: https://developers.webflow.com/data/reference/cms/collections/list

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

Request was successful

- `collections` (list of SitesSiteIdCollectionsGetResponsesContentApplicationJsonSchemaCollectionsItems, optional) — An array of Collections

## Types

### SitesSiteIdCollectionsGetResponsesContentApplicationJsonSchemaCollectionsItems

A collection object

- `id` (string, required) — Unique identifier for a Collection
- `displayName` (string, optional) — Name given to the Collection
- `singularName` (string, optional) — The name of one Item in Collection (e.g. ”Blog Post” if the Collection is called “Blog Posts”)
- `slug` (string, optional) — Slug of Collection in Site URL structure
- `createdOn` (string, optional) — The date the collection was created
- `lastUpdated` (string, optional) — The date the collection was last updated

### SitesSiteIdCollectionsGetResponsesContentApplicationJsonSchemaDetailsItems
