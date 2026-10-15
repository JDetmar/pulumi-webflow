# Create Collection

POST https://api.webflow.com/v2/sites/{site_id}/collections
Content-Type: application/json

Create a Collection for a site with collection fields.

Each collection includes the required _name_ and _slug_ fields, which are generated automatically. You can update the `displayName` of these fields, but the slug for them cannot be changed. Fields slugs are automatically converted to lowercase. Spaces in slugs are replaced with hyphens.

Required scope | `cms:write`

Reference: https://developers.webflow.com/data/reference/cms/collections/create

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

### Body (application/json)

This endpoint expects an object.

- `displayName` (string, required) — Name of the collection. Each collection name must be distinct.
- `singularName` (string, required) — Singular name of each item.
- `slug` (string, optional) — Part of a URL that identifier
- `fields` (list of SitesSiteIdCollectionsPostRequestBodyContentApplicationJsonSchemaFieldsItems, optional) — An array of custom fields to add to the collection

## Response

### 200

Request was successful

- `id` (string, required) — Unique identifier for a Collection
- `displayName` (string, required) — Name given to the Collection
- `singularName` (string, required) — The name of one Item in Collection (e.g. ”Blog Post” if the Collection is called “Blog Posts”)
- `fields` (list of SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaFieldsItems, required) — The list of fields in the Collection
- `slug` (string, optional) — Slug of Collection in Site URL structure
- `createdOn` (string, optional) — The date the collection was created
- `lastUpdated` (string, optional) — The date the collection was last updated
- `fieldGroups` (list of SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaFieldGroupsItems, optional) — The list of field groups in the Collection

## Types

### SitesSiteIdCollectionsPostRequestBodyContentApplicationJsonSchemaFieldsItems

Details about the field of a collection

### SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaFieldsItems

The details of a field in a collection

- `id` (string, required) — Unique identifier for a Field
- `isRequired` (boolean, required) — define whether a field is required in a collection
- `type` (enum, required) — Choose these appropriate field type for your collection data
  - Allowed values: `Color`, `DateTime`, `Email`, `ExtFileRef`, `File`, `Image`, `Link`, `MultiImage`, `MultiReference`, `Number`, `Option`, `Phone`, `PlainText`, `Reference`, `RichText`, `Switch`, `VideoLink`
- `displayName` (string, required) — The name of a field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `slug` (string, optional) — Slug of Field in Site URL structure. Slugs should be all lowercase with no spaces. Any spaces will be converted to "-."
- `helpText` (string, optional) — Additional text to help anyone filling out this field
- `validations` (SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaFieldsItemsValidations, optional, nullable) — The validations for the field

### SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaFieldGroupsItems

A group of fields within a collection

- `displayName` (string, required) — The display name of the field group
- `fieldIds` (list of string, required) — The IDs of the fields in this group
- `description` (string, optional) — A description of the field group

### SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaDetailsItems

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
- `metadata` (SitesSiteIdCollectionsPostRequestBodyContentApplicationJsonSchemaFieldsItemsOneOf1Metadata, required) — The metadata for the Option field.
- `id` (string, optional) — Unique identifier for a Field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `isRequired` (boolean, optional) — define whether a field is required in a collection
- `helpText` (string, optional) — Additional text to help anyone filling out this field

### Reference Field

- `type` (enum, required) — Choose these appropriate field type for your collection data
  - Allowed values: `MultiReference`, `Reference`
- `displayName` (string, required) — The name of a field
- `metadata` (SitesSiteIdCollectionsPostRequestBodyContentApplicationJsonSchemaFieldsItemsOneOf2Metadata, required) — The collectionId for the referenced collection. Only applicable for Reference and MultiReference fields.
- `id` (string, optional) — Unique identifier for a Field
- `isEditable` (boolean, optional) — Define whether the field is editable
- `isRequired` (boolean, optional) — define whether a field is required in a collection
- `helpText` (string, optional) — Additional text to help anyone filling out this field

### SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaFieldsItemsValidations

The validations for the field

- `additionalProperties` (SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaFieldsItemsValidationsAdditionalProperties, optional)

### SitesSiteIdCollectionsPostRequestBodyContentApplicationJsonSchemaFieldsItemsOneOf1Metadata

The metadata for the Option field.

- `options` (list of SitesSiteIdCollectionsPostRequestBodyContentApplicationJsonSchemaFieldsItemsOneOf1MetadataOptionsItems, required) — The option values for the Option field.

### SitesSiteIdCollectionsPostRequestBodyContentApplicationJsonSchemaFieldsItemsOneOf2Metadata

The collectionId for the referenced collection. Only applicable for Reference and MultiReference fields.

- `collectionId` (string, required) — The unique identifier of the collection

### SitesSiteIdCollectionsPostResponsesContentApplicationJsonSchemaFieldsItemsValidationsAdditionalProperties

### SitesSiteIdCollectionsPostRequestBodyContentApplicationJsonSchemaFieldsItemsOneOf1MetadataOptionsItems

A single option value for the Option field.

- `name` (string, required) — The name of the option
- `id` (string, optional) — The unique identifier of the option
