# List Items

GET https://api.webflow.com/v2/collections/{collection_id}/items

Rich Text field values can contain Webflow component instances as `<wf-component>` markup — see [Components in Rich Text](/data/docs/working-with-the-cms/components-in-rich-text) for the markup grammar, how to find component and property IDs, and the write constraints.

List of all Items within a Collection.

This endpoint supports:

* Custom `filter[...]` queries support up to 10 filter terms and 2 text-search terms per request.
* Custom `sort[...]` queries support up to 3 sort fields per request.

Required scope | `CMS:read`

Reference: https://developers.webflow.com/data/reference/cms/collection-items/staged-items/list-items

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `collection_id` (string, required) — Unique identifier for a Collection

### Query parameters

- `cmsLocaleId` (string, optional) — Unique identifier for a CMS Locale. This UID is different from the Site locale identifier and is listed as `cmsLocaleId` in the Sites response. To query multiple locales, input a comma separated string.
- `offset` (integer, optional) — Offset used for pagination if the results have more than limit records
- `limit` (integer, optional) — Maximum number of records to be returned (max limit: 100)
- `name` (string, optional) — Filter by the exact name of the item(s)
- `slug` (string, optional) — Filter by the exact slug of the item
- `createdOn` (CollectionsCollectionIdItemsGetParametersCreatedOn, optional) — Filter by the creation date of the item(s)
- `lastPublished` (CollectionsCollectionIdItemsGetParametersLastPublished, optional) — Filter by the last published date of the item(s)
- `lastUpdated` (CollectionsCollectionIdItemsGetParametersLastUpdated, optional) — Filter by the last updated date of the item(s)
- `filter` (map from string to CollectionsCollectionIdItemsGetParametersFilterSchema, optional) — Filter collection items by custom field values. Use bracket notation: `filter[<fieldSlug>][<operator>]=<value>`. Example: `filter[price][gte]=10&filter[price][lte]=100&filter[name][contains]=shirt`. Filters are combined with AND. You can combine custom field filters with top-level filters such as `name`, `slug`, `createdOn`, `lastPublished`, and `lastUpdated`. OR logic and nested filter groups are not supported on GET requests. More filter terms can increase request latency. Supported operators by field type: | Field type | Supported operators | | ---------------------------------------------------------------- | ----------------------------------------------------------- | | `id` | `eq`, `ne`, `in`, `nin` | | `PlainText` | `eq`, `ne`, `in`, `nin`, `contains`, `ncontains`, `exists` | | `Number` | `eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `in`, `nin`, `exists` | | `Switch` | `eq`, `ne`, `in`, `nin`, `exists` | | `DateTime` | `eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `in`, `nin`, `exists` | | `Email`, `Phone`, `Link` | `eq`, `ne`, `in`, `nin`, `contains`, `ncontains`, `exists` | | `Color` | `eq`, `ne`, `in`, `nin`, `exists` | | `Reference` | `eq`, `ne`, `in`, `nin`, `exists` | | `Option` | `eq`, `ne`, `in`, `nin` | | `RichText`, `Image`, `MultiImage`, `VideoLink`, `MultiReference` | `exists` | `contains` and `ncontains` are case-insensitive. `ncontains` also matches items where the field is empty or not set. `exists=true` matches items where the field has a value. `exists=false` matches items where the field is missing or null. For `Switch` fields, `false` is still a set value. Value formats: | Field type | Value format | | ----------------- | -------------------------------------- | | `Number` | A valid number, such as `10` or `12.5` | | `Switch` | `true` or `false` | | `DateTime` | ISO 8601 date-time string | | `id`, `Reference` | 24-character item ID | | `Option` | Option ID | | `in`, `nin` | Comma-separated list, up to 100 values | Invalid fields, invalid values, and operators that do not apply to a field type return a `400 BadArgument` response.
- `sortBy` (enum, optional) — Sort results by the provided value
  - Allowed values: `createdOn`, `lastPublished`, `lastUpdated`, `name`, `slug`
- `sortOrder` (enum, optional) — Sorts the results by asc or desc
  - Allowed values: `asc`, `desc`
- `sort` (map from string to enum, optional) — Sort collection items by custom fields using bracket notation: `sort[<fieldSlug>]=<asc|desc>`. * Example: `sort[price]=desc` * Multiple sort fields are applied in query-string order. When `sort[...]` is provided, it takes precedence over `sortBy` and `sortOrder`. * Sortable field types: `PlainText`, `Email`, `Phone`, `Number`, `DateTime`, and `Switch`. * Unknown fields, invalid sort directions, and non-sortable field types return a `400 BadArgument` response.
  - Allowed values: `asc`, `desc`
- `translatable` (string, optional) — Unique identifier for the secondary Locale you're translating **into**. Returns only content that hasn't been excluded from translation for that locale. This is independent of `localeId`, which selects which version of the content is returned. To fetch the source text to translate, request the primary locale's content and set `translatable` to the locale you're translating into: `?localeId={primary locale id}&translatable={target locale id}` Only exclusion rules scoped to manual translation are respected — rules scoped only to automatic translation don't affect this parameter's response. Omitting `translatable` returns the same response as if this parameter didn't exist. The value must be the id of one of the site's secondary locales — the primary locale id, or any other value, returns a `400` error. Requires translation exclusions to be enabled for the site; if they aren't, the request returns a `403` error. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)

## Response

### 200

Request was successful

- `items` (list of CollectionsCollectionIdItemsGetResponsesContentApplicationJsonSchemaItemsItems, required) — List of Items within the collection
- `pagination` (CollectionsCollectionIdItemsGetResponsesContentApplicationJsonSchemaPagination, required)

## Types

### CollectionsCollectionIdItemsGetParametersCreatedOn

- `lte` (string, optional) — Filter items created before this date
- `gte` (string, optional) — Filter items created after this date

### CollectionsCollectionIdItemsGetParametersLastPublished

- `lte` (string, optional) — Filter items last published before this date
- `gte` (string, optional) — Filter items last published after this date

### CollectionsCollectionIdItemsGetParametersLastUpdated

- `lte` (string, optional) — Filter items last updated before this date
- `gte` (string, optional) — Filter items last updated after this date

### CollectionsCollectionIdItemsGetParametersFilterSchema

- `eq` (string, optional)
- `ne` (string, optional)
- `gt` (string, optional)
- `gte` (string, optional)
- `lt` (string, optional)
- `lte` (string, optional)
- `in` (CollectionsCollectionIdItemsGetParametersFilterSchemaIn, optional)
- `nin` (CollectionsCollectionIdItemsGetParametersFilterSchemaNin, optional)
- `contains` (string, optional)
- `ncontains` (string, optional)
- `exists` (enum, optional)
  - Allowed values: `true`, `false`

### CollectionsCollectionIdItemsGetResponsesContentApplicationJsonSchemaItemsItems

A Collection Item represents a single entry in your collection. Each item includes: * **System metadata** - Automatically managed fields like IDs and timestamp&#x20; * **Status flags** - Controls for managing content state: `isDraft`, `isArchived ` * **Content fields** - Stored in `fieldData`. Each item needs a `name` and `slug`, and may include additional fields matching your collection's schema definition.

- `id` (string, required) — Unique identifier for the Item
- `lastPublished` (string, required) — The date the item was last published
- `lastUpdated` (string, required) — The date the item was last updated
- `createdOn` (string, required) — The date the item was created
- `fieldData` (CollectionsCollectionIdItemsGetResponsesContentApplicationJsonSchemaItemsItemsFieldData, required)
- `cmsLocaleId` (string, optional) — Identifier for the locale of the CMS item
- `isArchived` (boolean, optional, default: false) — Boolean determining if the Item is set to archived
- `isDraft` (boolean, optional, default: false) — Whether the item is in a draft state. Together with `lastPublished`, this determines the status shown in the Webflow UI. See [Publishing with the CMS API](/data/docs/working-with-the-cms/publishing) for the full mapping.

### CollectionsCollectionIdItemsGetResponsesContentApplicationJsonSchemaPagination

- `limit` (integer, optional, default: 100) — The limit specified in the request
- `offset` (integer, optional, default: 0) — The offset specified for pagination
- `total` (integer, optional) — Total number of items in the collection

### CollectionsCollectionIdItemsGetResponsesContentApplicationJsonSchemaDetailsItems

### CollectionsCollectionIdItemsGetParametersFilterSchemaIn

### CollectionsCollectionIdItemsGetParametersFilterSchemaNin

### CollectionsCollectionIdItemsGetResponsesContentApplicationJsonSchemaItemsItemsFieldData

- `name` (string, required) — Name of the Item
- `slug` (string, required) — URL structure of the Item in your site. Note: Updates to an item slug will break all links referencing the old slug.
