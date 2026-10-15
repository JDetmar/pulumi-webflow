# Get Collection Details

GET https://api.webflow.com/v2/collections/{collection_id}

Get the full details of a collection from its ID.

Required scope | `cms:read`

Reference: https://developers.webflow.com/data/reference/cms/collections/get

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `collection_id` (string, required) — Unique identifier for a Collection

## Response

### 200

Request was successful

- `id` (string, required) — Unique identifier for a Collection
- `displayName` (string, required) — Name given to the Collection
- `singularName` (string, required) — The name of one Item in Collection (e.g. ”Blog Post” if the Collection is called “Blog Posts”)
- `fields` (list of CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaFieldsItems, required) — The list of fields in the Collection
- `slug` (string, optional) — Slug of Collection in Site URL structure
- `createdOn` (string, optional) — The date the collection was created
- `lastUpdated` (string, optional) — The date the collection was last updated
- `fieldGroups` (list of CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaFieldGroupsItems, optional) — The list of field groups in the Collection

## Types

### CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaFieldsItems

The details of a field in a collection

- `id` (string, required) — Unique identifier for a Field
- `isRequired` (boolean, required) — define whether a field is required in a collection
- `type` (enum, required) — Choose these appropriate field type for your collection data
  - Allowed values: `Color`, `DateTime`, `Email`, `ExtFileRef`, `File`, `Image`, `Link`, `MultiImage`, `MultiReference`, `Number`, `Option`, `Phone`, `PlainText`, `Reference`, `RichText`, `Switch`, `VideoLink`
- `displayName` (string, required) — The name of a field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `slug` (string, optional) — Slug of Field in Site URL structure. Slugs should be all lowercase with no spaces. Any spaces will be converted to "-."
- `helpText` (string, optional) — Additional text to help anyone filling out this field
- `validations` (CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaFieldsItemsValidations, optional, nullable) — The validations for the field

### CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaFieldGroupsItems

A group of fields within a collection

- `displayName` (string, required) — The display name of the field group
- `fieldIds` (list of string, required) — The IDs of the fields in this group
- `description` (string, optional) — A description of the field group

### CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaDetailsItems

### CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaFieldsItemsValidations

The validations for the field

- `additionalProperties` (CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaFieldsItemsValidationsAdditionalProperties, optional)

### CollectionsCollectionIdGetResponsesContentApplicationJsonSchemaFieldsItemsValidationsAdditionalProperties
