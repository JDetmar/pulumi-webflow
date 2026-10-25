# Create field

POST https://api.webflow.com/v2/collections/{collection_id}/fields
Content-Type: application/json

Create a custom field in a collection.

Field validation is currently not available through the API.

Bulk creation of fields is not supported with this endpoint. To add multiple fields at once, include them when you [create the collection.](/data/v2.0.0/reference/cms/collections/create)

Required scope | `cms:write`

Reference: https://developers.webflow.com/data/reference/cms/collection-fields/create

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `collection_id` (string, required) — Unique identifier for a Collection

### Body (application/json)

This endpoint expects a collections_fields_create_Request.

- `collections_fields_create_Request`
  - Static Field: Static Field
  - Option Field: Option Field
  - Reference Field: Reference Field

## Response

### 200

Request was successful

- `collections_fields_create_Response_200`
  - Static Field: Static Field
  - Option Field: Option Field
  - Reference Field: Reference Field

## Types

### Static Field

- `type` (enum, required) — Choose these appropriate field type for your collection data
  - Allowed values: `Color`, `DateTime`, `Email`, `File`, `Image`, `Link`, `MultiImage`, `Number`, `Phone`, `PlainText`, `RichText`, `Switch`, `VideoLink`
- `displayName` (string, required) — The name of a field
- `id` (string, optional) — Unique identifier for a Field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `isRequired` (boolean, optional) — define whether a field is required in a collection
- `helpText` (string, optional) — Additional text to help anyone filling out this field

### Option Field

- `type` (enum, required) — The [Option field type](/data/reference/field-types-item-values#option)
  - Allowed values: `Option`
- `displayName` (string, required) — The name of a field
- `metadata` (CollectionsCollectionIdFieldsPostRequestBodyContentApplicationJsonSchemaOneOf1Metadata, required) — The metadata for the Option field.
- `id` (string, optional) — Unique identifier for a Field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `isRequired` (boolean, optional) — define whether a field is required in a collection
- `helpText` (string, optional) — Additional text to help anyone filling out this field

### Reference Field

- `type` (enum, required) — Choose these appropriate field type for your collection data
  - Allowed values: `MultiReference`, `Reference`
- `displayName` (string, required) — The name of a field
- `metadata` (CollectionsCollectionIdFieldsPostRequestBodyContentApplicationJsonSchemaOneOf2Metadata, required) — The collectionId for the referenced collection. Only applicable for Reference and MultiReference fields.
- `id` (string, optional) — Unique identifier for a Field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `isRequired` (boolean, optional) — define whether a field is required in a collection
- `helpText` (string, optional) — Additional text to help anyone filling out this field

### Static Field

- `type` (enum, required) — Choose these appropriate field type for your collection data
  - Allowed values: `Color`, `DateTime`, `Email`, `File`, `Image`, `Link`, `MultiImage`, `Number`, `Phone`, `PlainText`, `RichText`, `Switch`, `VideoLink`
- `displayName` (string, required) — The name of a field
- `id` (string, optional) — Unique identifier for a Field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `isRequired` (boolean, optional) — define whether a field is required in a collection
- `helpText` (string, optional) — Additional text to help anyone filling out this field

### Option Field

- `type` (enum, required) — The [Option field type](/data/reference/field-types-item-values#option)
  - Allowed values: `Option`
- `displayName` (string, required) — The name of a field
- `metadata` (CollectionsCollectionIdFieldsPostResponsesContentApplicationJsonSchemaOneOf1Metadata, required) — The metadata for the Option field.
- `id` (string, optional) — Unique identifier for a Field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `isRequired` (boolean, optional) — define whether a field is required in a collection
- `helpText` (string, optional) — Additional text to help anyone filling out this field

### Reference Field

- `type` (enum, required) — Choose these appropriate field type for your collection data
  - Allowed values: `MultiReference`, `Reference`
- `displayName` (string, required) — The name of a field
- `metadata` (CollectionsCollectionIdFieldsPostResponsesContentApplicationJsonSchemaOneOf2Metadata, required) — The collectionId for the referenced collection. Only applicable for Reference and MultiReference fields.
- `id` (string, optional) — Unique identifier for a Field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `isRequired` (boolean, optional) — define whether a field is required in a collection
- `helpText` (string, optional) — Additional text to help anyone filling out this field

### CollectionsCollectionIdFieldsPostResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any

### CollectionsCollectionIdFieldsPostRequestBodyContentApplicationJsonSchemaOneOf1Metadata

The metadata for the Option field.

- `options` (list of CollectionsCollectionIdFieldsPostRequestBodyContentApplicationJsonSchemaOneOf1MetadataOptionsItems, required) — The option values for the Option field.

### CollectionsCollectionIdFieldsPostRequestBodyContentApplicationJsonSchemaOneOf2Metadata

The collectionId for the referenced collection. Only applicable for Reference and MultiReference fields.

- `collectionId` (string, required) — The unique identifier of the collection

### CollectionsCollectionIdFieldsPostResponsesContentApplicationJsonSchemaOneOf1Metadata

The metadata for the Option field.

- `options` (list of CollectionsCollectionIdFieldsPostResponsesContentApplicationJsonSchemaOneOf1MetadataOptionsItems, required) — The option values for the Option field.

### CollectionsCollectionIdFieldsPostResponsesContentApplicationJsonSchemaOneOf2Metadata

The collectionId for the referenced collection. Only applicable for Reference and MultiReference fields.

- `collectionId` (string, required) — The unique identifier of the collection

### CollectionsCollectionIdFieldsPostRequestBodyContentApplicationJsonSchemaOneOf1MetadataOptionsItems

A single option value for the Option field.

- `name` (string, required) — The name of the option
- `id` (string, optional) — The unique identifier of the option

### CollectionsCollectionIdFieldsPostResponsesContentApplicationJsonSchemaOneOf1MetadataOptionsItems

A single option value for the Option field.

- `name` (string, required) — The name of the option
- `id` (string, optional) — The unique identifier of the option
