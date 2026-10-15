# Delete Collection

DELETE https://api.webflow.com/v2/collections/{collection_id}

Delete a collection using its ID.

Required scope | `cms:write`

Reference: https://developers.webflow.com/data/reference/cms/collections/delete

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `collection_id` (string, required) — Unique identifier for a Collection

## Response

### 204

Request was successful. No Content is returned.

## Types

### CollectionsCollectionIdDeleteResponsesContentApplicationJsonSchemaDetailsItems
