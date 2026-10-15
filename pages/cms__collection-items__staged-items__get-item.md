# Get Item

GET https://api.webflow.com/v2/collections/{collection_id}/items/{item_id}

Rich Text field values can contain Webflow component instances as `<wf-component>` markup — see [Components in Rich Text](/data/docs/working-with-the-cms/components-in-rich-text) for the markup grammar, how to find component and property IDs, and the write constraints.

Get details of a selected Collection Item.

Required scope | `CMS:read`

Reference: https://developers.webflow.com/data/reference/cms/collection-items/staged-items/get-item

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `collection_id` (string, required) — Unique identifier for a Collection
- `item_id` (string, required) — Unique identifier for an Item

### Query parameters

- `cmsLocaleId` (string, optional) — Unique identifier for a CMS Locale. This UID is different from the Site locale identifier and is listed as `cmsLocaleId` in the Sites response. This endpoint returns a single item, so it accepts one locale. To retrieve an item in several locales, use [List Collection Items](/data/reference/cms/collection-items/staged-items/list-items) with `filter[id][eq]` and a comma separated `cmsLocaleId`.
- `translatable` (string, optional) — Unique identifier for the secondary Locale you're translating **into**. Returns only content that hasn't been excluded from translation for that locale. This is independent of `localeId`, which selects which version of the content is returned. To fetch the source text to translate, request the primary locale's content and set `translatable` to the locale you're translating into: `?localeId={primary locale id}&translatable={target locale id}` Only exclusion rules scoped to manual translation are respected — rules scoped only to automatic translation don't affect this parameter's response. Omitting `translatable` returns the same response as if this parameter didn't exist. The value must be the id of one of the site's secondary locales — the primary locale id, or any other value, returns a `400` error. Requires translation exclusions to be enabled for the site; if they aren't, the request returns a `403` error. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)

## Response

### 200

Request was successful

- `id` (string, required) — Unique identifier for the Item
- `lastPublished` (string, required) — The date the item was last published
- `lastUpdated` (string, required) — The date the item was last updated
- `createdOn` (string, required) — The date the item was created
- `fieldData` (CollectionsCollectionIdItemsItemIdGetResponsesContentApplicationJsonSchemaFieldData, required)
- `cmsLocaleId` (string, optional) — Identifier for the locale of the CMS item
- `isArchived` (boolean, optional, default: false) — Boolean determining if the Item is set to archived
- `isDraft` (boolean, optional, default: false) — Whether the item is in a draft state. Together with `lastPublished`, this determines the status shown in the Webflow UI. See [Publishing with the CMS API](/data/docs/working-with-the-cms/publishing) for the full mapping.

## Types

### CollectionsCollectionIdItemsItemIdGetResponsesContentApplicationJsonSchemaFieldData

- `name` (string, required) — Name of the Item
- `slug` (string, required) — URL structure of the Item in your site. Note: Updates to an item slug will break all links referencing the old slug.

### CollectionsCollectionIdItemsItemIdGetResponsesContentApplicationJsonSchemaDetailsItems
