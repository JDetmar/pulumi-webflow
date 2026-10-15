# Get Page Metadata

GET https://api.webflow.com/v2/pages/{page_id}

Get metadata information for a single page.

Required scope | `pages:read`

Reference: https://developers.webflow.com/data/reference/pages-and-components/pages/get-metadata

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `page_id` (string, required) — Unique identifier for a Page

### Query parameters

- `localeId` (string, optional) — Unique identifier for a specific Locale. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)
- `translatable` (string, optional) — Unique identifier for the secondary Locale you're translating **into**. Returns only content that hasn't been excluded from translation for that locale. This is independent of `localeId`, which selects which version of the content is returned. To fetch the source text to translate, request the primary locale's content and set `translatable` to the locale you're translating into: `?localeId={primary locale id}&translatable={target locale id}` Only exclusion rules scoped to manual translation are respected — rules scoped only to automatic translation don't affect this parameter's response. Omitting `translatable` returns the same response as if this parameter didn't exist. The value must be the id of one of the site's secondary locales — the primary locale id, or any other value, returns a `400` error. Requires translation exclusions to be enabled for the site; if they aren't, the request returns a `403` error. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)

## Response

### 200

Request was successful

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
- `seo` (PagesPageIdGetResponsesContentApplicationJsonSchemaSeo, optional) — SEO-related fields for the Page
- `openGraph` (PagesPageIdGetResponsesContentApplicationJsonSchemaOpenGraph, optional) — Open Graph fields for the Page
- `localeId` (string, optional, nullable) — Unique ID of the page locale
- `publishedPath` (string, optional) — Relative path of the published page URL

## Types

### PagesPageIdGetResponsesContentApplicationJsonSchemaSeo

SEO-related fields for the Page

- `title` (string, optional) — The Page title shown in search engine results
- `description` (string, optional) — The Page description shown in search engine results

### PagesPageIdGetResponsesContentApplicationJsonSchemaOpenGraph

Open Graph fields for the Page

- `title` (string, optional) — The title supplied to Open Graph annotations
- `titleCopied` (boolean, optional, default: true) — Indicates the Open Graph title was copied from the SEO title
- `description` (string, optional) — The description supplied to Open Graph annotations
- `descriptionCopied` (boolean, optional, default: true) — Indicates the Open Graph description was copied from the SEO description

### PagesPageIdGetResponsesContentApplicationJsonSchemaDetailsItems
