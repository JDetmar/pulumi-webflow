# List Webhooks

GET https://api.webflow.com/v2/sites/{site_id}/webhooks

List all App-created Webhooks registered for a given site

Required scope | `sites:read`

Reference: https://developers.webflow.com/data/reference/webhooks/list

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.
- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

Request was successful

- `webhooks` (list of SitesSiteIdWebhooksGetResponsesContentApplicationJsonSchemaWebhooksItems, optional)
- `pagination` (SitesSiteIdWebhooksGetResponsesContentApplicationJsonSchemaPagination, optional) — Pagination object

## Types

### SitesSiteIdWebhooksGetResponsesContentApplicationJsonSchemaWebhooksItems

- `id` (string, optional) — Unique identifier for the Webhook registration
- `triggerType` (enum, optional) — The type of event that triggered the request. See the the documentation for details on [supported events](/data/reference/all-events).
  - Allowed values: `form_submission`, `site_publish`, `page_created`, `page_metadata_updated`, `page_deleted`, `ecomm_new_order`, `ecomm_order_changed`, `ecomm_inventory_changed`, `collection_item_created`, `collection_item_changed`, `collection_item_deleted`, `collection_item_published`, `collection_item_unpublished`, `comment_created`, `branch_created`, `branch_merged`, `branch_deleted`
- `url` (string, optional) — URL to send the Webhook payload to
- `workspaceId` (string, optional) — Unique identifier for the Workspace the Webhook is registered in
- `siteId` (string, optional) — Unique identifier for the Site the Webhook is registered in
- `filter` (SitesSiteIdWebhooksGetResponsesContentApplicationJsonSchemaWebhooksItemsFilter, optional) — Only supported for the `form_submission` trigger type. Filter for the form you want Webhooks to be sent for.
- `lastTriggered` (string, optional) — Date the Webhook instance was last triggered
- `createdOn` (string, optional) — Date the Webhook registration was created

### SitesSiteIdWebhooksGetResponsesContentApplicationJsonSchemaPagination

Pagination object

- `limit` (integer, required) — The limit used for pagination
- `offset` (integer, required) — The offset used for pagination
- `total` (integer, required) — The total number of records

### SitesSiteIdWebhooksGetResponsesContentApplicationJsonSchemaDetailsItems

### SitesSiteIdWebhooksGetResponsesContentApplicationJsonSchemaWebhooksItemsFilter

Only supported for the `form_submission` trigger type. Filter for the form you want Webhooks to be sent for.

- `name` (string, optional) — The name of the form you'd like to recieve notifications for.
