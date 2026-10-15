# Remove Webhook

DELETE https://api.webflow.com/v2/webhooks/{webhook_id}

Remove a Webhook

Required scope: `sites:read`

Reference: https://developers.webflow.com/data/reference/webhooks/delete

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.
- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `webhook_id` (string, required) — Unique identifier for a Webhook

## Response

### 204

Request was successful. No Content is returned.

## Types

### WebhooksWebhookIdDeleteResponsesContentApplicationJsonSchemaDetailsItems
