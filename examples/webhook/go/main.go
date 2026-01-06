package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := pulumi.NewConfig(ctx, "")
		siteID := cfg.RequireSecret("siteId")
		environment := cfg.Get("environment")
		if environment == "" {
			environment = "development"
		}

		// Example 1: Form Submission Webhook
		// Receive notifications when users submit forms on your site
		formWebhook, err := webflow.NewWebhook(ctx, "form-submission-webhook", &webflow.WebhookArgs{
			SiteId:      siteID,
			TriggerType: pulumi.String("form_submission"),
			Url:         pulumi.String("https://your-api.example.com/webhooks/webflow/forms"),
		})
		if err != nil {
			return fmt.Errorf("failed to create form webhook: %w", err)
		}

		// Example 2: Site Publish Webhook
		// Get notified when your site is published
		publishWebhook, err := webflow.NewWebhook(ctx, "site-publish-webhook", &webflow.WebhookArgs{
			SiteId:      siteID,
			TriggerType: pulumi.String("site_publish"),
			Url:         pulumi.String("https://your-api.example.com/webhooks/webflow/publish"),
		})
		if err != nil {
			return fmt.Errorf("failed to create publish webhook: %w", err)
		}

		// Example 3: E-commerce Order Webhook
		// Track new orders in your Webflow e-commerce store
		ecommWebhook, err := webflow.NewWebhook(ctx, "ecomm-order-webhook", &webflow.WebhookArgs{
			SiteId:      siteID,
			TriggerType: pulumi.String("ecomm_new_order"),
			Url:         pulumi.String("https://your-api.example.com/webhooks/webflow/orders"),
		})
		if err != nil {
			return fmt.Errorf("failed to create ecomm webhook: %w", err)
		}

		// Example 4: Collection Item Webhook with Filter
		// Monitor changes to specific collection items
		// Note: Replace "your-collection-id-here" with an actual collection ID
		collectionWebhook, err := webflow.NewWebhook(ctx, "collection-item-webhook", &webflow.WebhookArgs{
			SiteId:      siteID,
			TriggerType: pulumi.String("collection_item_created"),
			Url:         pulumi.String("https://your-api.example.com/webhooks/webflow/collection"),
			Filter: pulumi.Map{
				"collectionIds": pulumi.Array{pulumi.String("your-collection-id-here")},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create collection webhook: %w", err)
		}

		// Example 5: Page Metadata Update Webhook
		// Track when page metadata changes (title, description, SEO settings)
		pageMetadataWebhook, err := webflow.NewWebhook(ctx, "page-metadata-webhook", &webflow.WebhookArgs{
			SiteId:      siteID,
			TriggerType: pulumi.String("page_metadata_updated"),
			Url:         pulumi.String("https://your-api.example.com/webhooks/webflow/pages"),
		})
		if err != nil {
			return fmt.Errorf("failed to create page metadata webhook: %w", err)
		}

		// Example 6: Membership User Account Webhook
		// Monitor user account creation in Webflow Memberships
		membershipWebhook, err := webflow.NewWebhook(ctx, "membership-webhook", &webflow.WebhookArgs{
			SiteId:      siteID,
			TriggerType: pulumi.String("memberships_user_account_added"),
			Url:         pulumi.String("https://your-api.example.com/webhooks/webflow/members"),
		})
		if err != nil {
			return fmt.Errorf("failed to create membership webhook: %w", err)
		}

		// Export webhook IDs and timestamps for reference
		ctx.Export("deployedSiteId", siteID)
		ctx.Export("formWebhookId", formWebhook.ID())
		ctx.Export("formWebhookCreated", formWebhook.CreatedOn)
		ctx.Export("publishWebhookId", publishWebhook.ID())
		ctx.Export("ecommWebhookId", ecommWebhook.ID())
		ctx.Export("collectionWebhookId", collectionWebhook.ID())
		ctx.Export("pageMetadataWebhookId", pageMetadataWebhook.ID())
		ctx.Export("membershipWebhookId", membershipWebhook.ID())

		webhookCount := 6
		ctx.Log.Info(
			fmt.Sprintf("✅ Successfully deployed %d webhooks in %s environment", webhookCount, environment),
			&pulumi.LogOptions{},
		)

		return nil
	})
}
