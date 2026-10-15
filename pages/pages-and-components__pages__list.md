# List Pages

GET https://api.webflow.com/v2/sites/{site_id}/pages

List of all pages for a site.

Required scope | `pages:read`

Reference: https://developers.webflow.com/data/reference/pages-and-components/pages/list

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Query parameters

- `localeId` (string, optional) — Unique identifier for a specific Locale. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)
- `limit` (integer, optional) — Maximum number of records to be returned (max limit: 100)
- `offset` (integer, optional) — Offset used for pagination if the results have more than limit records

## Response

### 200

Request was successful

- `pages` (list of SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaPagesItems, optional)
- `pagination` (SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaPagination, optional) — Pagination object

## Types

### SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaPagesItems

The Page object

- `id` (string, required) — Unique identifier for the Page
- `siteId` (string, optional) — Unique identifier for the Site
- `title` (string, optional) — Title of the Page
- `slug` (string, optional) — slug of the Page (derived from title)
- `parentId` (string, optional) — Identifier of the parent folder
- `collectionId` (string, optional) — Unique identifier for a linked Collection, value will be null if the Page is not part of a Collection.
- `createdOn` (string, optional) — The date the Page was created
- `lastUpdated` (string, optional) — The date the Page was most recently updated
- `archived` (boolean, optional, default: false) — Whether the Page has been archived
- `draft` (boolean, optional, default: false) — Whether the Page is a draft
- `canBranch` (boolean, optional, default: false) — Indicates whether the Page supports [Page Branching](https://university.webflow.com/lesson/page-branching). Pages that are already branches cannot be branched again.
- `isBranch` (boolean, optional, default: false) — Indicates whether the Page is a Branch of another Page [Page Branching](https://university.webflow.com/lesson/page-branching)
- `branchId` (string, optional, nullable) — If the Page is a Branch of another Page, this is the ID of the Branch
- `seo` (SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaPagesItemsSeo, optional) — SEO-related fields for the Page
- `openGraph` (SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaPagesItemsOpenGraph, optional) — Open Graph fields for the Page
- `localeId` (string, optional, nullable) — Unique ID of the page locale
- `publishedPath` (string, optional) — Relative path of the published page URL

### SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaPagination

Pagination object

- `limit` (integer, required) — The limit used for pagination
- `offset` (integer, required) — The offset used for pagination
- `total` (integer, required) — The total number of records

### SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaDetailsItems

### SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaPagesItemsSeo

SEO-related fields for the Page

- `title` (string, optional) — The Page title shown in search engine results
- `description` (string, optional) — The Page description shown in search engine results

### SitesSiteIdPagesGetResponsesContentApplicationJsonSchemaPagesItemsOpenGraph

Open Graph fields for the Page

- `title` (string, optional) — The title supplied to Open Graph annotations
- `titleCopied` (boolean, optional, default: true) — Indicates the Open Graph title was copied from the SEO title
- `description` (string, optional) — The description supplied to Open Graph annotations
- `descriptionCopied` (boolean, optional, default: true) — Indicates the Open Graph description was copied from the SEO description
