# Webhook Resource Examples

This directory contains examples demonstrating how to create and manage webhooks for Webflow sites using Pulumi in multiple languages.

## What You'll Learn

- Set up webhooks to receive real-time notifications from Webflow
- Monitor form submissions, site publishes, and content changes
- Track e-commerce orders and inventory updates
- Receive alerts for collection item changes
- Monitor membership user account events
- Use filters to receive only specific webhook events

## What Are Webhooks?

Webhooks allow your application to receive real-time notifications when events occur in your Webflow site. Instead of constantly polling the Webflow API, webhooks push event data to your specified HTTPS endpoint when something happens.

**Common Use Cases:**
- Send email notifications when forms are submitted
- Trigger CI/CD pipelines when a site is published
- Update external databases when collection items change
- Process orders in your e-commerce backend
- Sync user accounts with external systems

## Available Languages

| Language   | Directory    | Entry Point    | Dependencies        |
|------------|--------------|----------------|---------------------|
| TypeScript | `typescript/`| `index.ts`     | `package.json`      |
| Python     | `python/`    | `__main__.py`  | `requirements.txt`  |
| Go         | `go/`        | `main.go`      | `go.mod`            |

## Quick Start

### TypeScript

```bash
cd typescript
npm install
pulumi stack init dev
pulumi config set siteId your-site-id-here
pulumi up
```

### Python

```bash
cd python
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pulumi stack init dev
pulumi config set siteId your-site-id-here
pulumi up
```

### Go

```bash
cd go
go mod download
pulumi stack init dev
pulumi config set siteId your-site-id-here
pulumi up
```

## Examples Included

### 1. Form Submission Webhook

Receive notifications when users submit forms on your site.

```typescript
triggerType: "form_submission"
url: "https://your-api.example.com/webhooks/webflow/forms"
```

**Use Case:** Send confirmation emails, store submissions in a database, trigger Slack notifications.

### 2. Site Publish Webhook

Get notified when your site is published (either full or partial publish).

```typescript
triggerType: "site_publish"
url: "https://your-api.example.com/webhooks/webflow/publish"
```

**Use Case:** Trigger cache invalidation, run automated tests, notify team members.

### 3. E-commerce Order Webhook

Track new orders in your Webflow e-commerce store.

```typescript
triggerType: "ecomm_new_order"
url: "https://your-api.example.com/webhooks/webflow/orders"
```

**Use Case:** Process payments, update inventory systems, send order confirmations, notify fulfillment.

### 4. Collection Item Webhook with Filter

Monitor changes to specific collection items. This example shows using a filter to receive events only for specific collections.

```typescript
triggerType: "collection_item_created"
url: "https://your-api.example.com/webhooks/webflow/collection"
filter: {
  collectionIds: ["your-collection-id-here"]
}
```

**Use Case:** Sync blog posts to external platforms, trigger social media posts, update search indexes.

### 5. Page Metadata Update Webhook

Track when page metadata changes (title, description, SEO settings).

```typescript
triggerType: "page_metadata_updated"
url: "https://your-api.example.com/webhooks/webflow/pages"
```

**Use Case:** Update sitemaps, invalidate CDN cache, trigger SEO audits.

### 6. Membership User Account Webhook

Monitor user account creation in Webflow Memberships.

```typescript
triggerType: "memberships_user_account_added"
url: "https://your-api.example.com/webhooks/webflow/members"
```

**Use Case:** Send welcome emails, create user profiles in external systems, trigger onboarding workflows.

## Available Trigger Types

The examples demonstrate these trigger types:

| Trigger Type | Description |
|--------------|-------------|
| `form_submission` | Form is submitted on your site |
| `site_publish` | Site is published (full or partial) |
| `page_created` | New page is created |
| `page_metadata_updated` | Page metadata (title, description, etc.) is updated |
| `page_deleted` | Page is deleted |
| `ecomm_new_order` | New order is placed |
| `ecomm_order_changed` | Order status or details change |
| `ecomm_inventory_changed` | Product inventory is updated |
| `memberships_user_account_added` | New user account is created |
| `memberships_user_account_updated` | User account is updated |
| `memberships_user_account_deleted` | User account is deleted |
| `collection_item_created` | Collection item is created |
| `collection_item_changed` | Collection item is updated |
| `collection_item_deleted` | Collection item is deleted |
| `collection_item_unpublished` | Collection item is unpublished |

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `siteId`          | Yes      | Your Webflow site ID                     |
| `environment`     | No       | Deployment environment (default: development) |

**Important:** Your webhook URL must:
- Use HTTPS (HTTP is not allowed by Webflow)
- Be publicly accessible
- Accept POST requests
- Return a 200-series status code

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    deployedSiteId           : "abc123..."
    formWebhookId            : "webhook_abc..."
    formWebhookCreated       : "2025-01-06T12:34:56Z"
    publishWebhookId         : "webhook_def..."
    ecommWebhookId           : "webhook_ghi..."
    collectionWebhookId      : "webhook_jkl..."
    pageMetadataWebhookId    : "webhook_mno..."
    membershipWebhookId      : "webhook_pqr..."
```

## Webhook Payload Example

When a webhook fires, Webflow sends a POST request to your URL with a JSON payload:

```json
{
  "triggerType": "form_submission",
  "site": "abc123...",
  "data": {
    "formName": "Contact Form",
    "email": "user@example.com",
    "name": "John Doe",
    ...
  },
  "_id": "webhook_event_xyz...",
  "triggeredAt": "2025-01-06T12:34:56Z"
}
```

The payload structure varies by trigger type. See [Webflow Webhooks Documentation](https://developers.webflow.com/data/docs/webhooks) for details.

## Important Notes

### Webhook Updates

**Webhooks cannot be updated in-place.** Any change to `triggerType`, `url`, or `filter` requires replacing the webhook (delete and recreate). This is a Webflow API limitation, not a provider limitation.

```
# Changing the URL will trigger replacement
pulumi up  # Shows: ~ replace webhook
```

### Testing Webhooks

To test your webhooks locally:

1. Use a service like [ngrok](https://ngrok.com/) to expose localhost:
   ```bash
   ngrok http 3000
   ```

2. Use the ngrok HTTPS URL in your webhook:
   ```typescript
   url: "https://your-subdomain.ngrok.io/webhooks/webflow"
   ```

3. Set up a simple HTTP server to receive webhooks:
   ```javascript
   // Node.js example
   const express = require('express');
   const app = express();
   app.post('/webhooks/webflow', (req, res) => {
     console.log('Webhook received:', req.body);
     res.sendStatus(200);
   });
   app.listen(3000);
   ```

### Security Best Practices

1. **Verify webhook signatures** - Webflow signs webhook requests. Verify the signature to ensure requests are authentic.
2. **Use HTTPS** - Required by Webflow.
3. **Implement idempotency** - Webhooks may be delivered multiple times. Handle duplicate events gracefully.
4. **Respond quickly** - Return 200 quickly, then process asynchronously. Webflow times out after 30 seconds.
5. **Log webhook events** - Keep audit logs of received webhooks for debugging.

## Cleanup

To remove all created webhooks:

```bash
pulumi destroy
pulumi stack rm dev
```

## Troubleshooting

### "Site not found" Error

1. Verify your site ID in Webflow: Settings → General
2. Ensure correct format: 24-character lowercase hexadecimal
3. Check API token has access to the site

### "Invalid URL" Error

- Ensure URL starts with `https://` (not `http://`)
- Verify the URL is publicly accessible
- Test the URL with `curl` to ensure it accepts POST requests

### Webhook Not Firing

1. Check that the event is actually occurring in Webflow
2. Verify your endpoint is publicly accessible
3. Check Webflow's webhook logs in the dashboard
4. Ensure your endpoint returns 200 status code within 30 seconds
5. Review webhook delivery logs in Webflow dashboard

### "Invalid trigger type" Error

- Ensure `triggerType` matches one of the supported values exactly
- Check for typos (e.g., `form_submissions` instead of `form_submission`)
- Refer to the Available Trigger Types table above

## Related Resources

- [Webhook API Reference](../../docs/api/webhook.md) (if available)
- [Main Examples Index](../README.md)
- [Webflow Webhooks Documentation](https://developers.webflow.com/data/docs/webhooks)
- [Webflow Webhook Trigger Types](https://developers.webflow.com/data/docs/webhook-triggers)

## Next Steps

After setting up webhooks, consider:
- Implementing webhook signature verification for security
- Setting up error handling and retry logic
- Creating monitoring/alerting for webhook failures
- Building a webhook event processing queue for reliability
