# Delete field

DELETE https://api.webflow.com/v2/collections/{collection_id}/fields/{field_id}

Delete a custom field in a collection. This endpoint does not currently support bulk deletion.

Required scope | `cms:write`

Reference: https://developers.webflow.com/data/reference/cms/collection-fields/delete

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `collection_id` (string, required) — Unique identifier for a Collection
- `field_id` (string, required) — Unique identifier for a Field in a collection

## Response

### 204

Request was successful. No Content is returned.

## Types

### CollectionsCollectionIdFieldsFieldIdDeleteResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any
