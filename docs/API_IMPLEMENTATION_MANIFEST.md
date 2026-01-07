# Webflow API Implementation Manifest

## Official API Specification

**OpenAPI Spec:** https://raw.githubusercontent.com/webflow/openapi-spec/refs/heads/main/openapi/v2.yml

**Base URL:** `https://api.webflow.com/v2`

**Rate Limit:** 60 req/min (respect `X-RateLimit-Remaining` header)

---

## Implementation Status

### ✅ Completed (15 Resources)

| Resource | Files | Description |
|----------|-------|-------------|
| Site | `site_resource.go`, `site.go` | Site management with publish support |
| Redirect | `redirect_resource.go`, `redirect.go` | URL redirect rules |
| RobotsTxt | `robotstxt_resource.go`, `robotstxt.go` | SEO robots.txt configuration |
| Collection | `collection_resource.go`, `collection.go` | CMS collection schema |
| CollectionField | `collectionfield_resource.go`, `collectionfield.go` | CMS collection field definitions |
| CollectionItem | `collectionitem_resource.go`, `collectionitem.go` | CMS content items (draft/publish via isDraft) |
| Page | `page_resource.go`, `page.go` | Static page management |
| PageContent | `pagecontent_resource.go`, `pagecontent.go` | Page DOM content |
| Asset | `asset_resource.go`, `asset.go` | File/image uploads |
| AssetFolder | `assetfolder_resource.go`, `assetfolder.go` | Asset organization |
| Webhook | `webhook_resource.go`, `webhook.go` | Event webhooks |
| SiteCustomCode | `sitecustomcode_resource.go`, `sitecustomcode.go` | Site-level custom scripts |
| PageCustomCode | `pagecustomcode_resource.go`, `pagecustomcode.go` | Page-level custom scripts |
| RegisteredScript | `registeredscript_resource.go`, `registeredscript.go` | Hosted/inline script registration |
| User | `user_resource.go`, `user.go` | Site membership users |

---

## 🚀 Future Candidates

These APIs were evaluated as good IaC candidates but not yet implemented:

### E-commerce Configuration
| Resource | Methods | Endpoints | Scope | Notes |
|----------|---------|-----------|-------|-------|
| EcommerceSettings | GET | `/sites/{site_id}/ecommerce/settings` | ecommerce:read | Site-wide e-commerce configuration (payment, tax, shipping settings) |

### Provider Utilities (Data Sources)
| Resource | Methods | Endpoints | Scope | Notes |
|----------|---------|-----------|-------|-------|
| TokenInfo | GET | `/token/introspect` | - | Token permission introspection for validation |

---

## ❌ Not Suitable for IaC

The following Webflow APIs were evaluated and determined **not suitable** for Infrastructure-as-Code:

### Read-Only APIs (No CRUD Support)
| API | Reason |
|-----|--------|
| CustomDomain | Read-only; domains managed via Webflow UI and DNS registrars |
| Form | Read-only; forms created in Webflow designer, not via API |
| AccessGroup | Read-only; groups defined in Webflow UI for membership tiers |
| AuthorizedUser | Audit metadata about token creator; not infrastructure state |

### Already Covered by Existing Resources
| API | Covered By |
|-----|------------|
| SitePublish | `Site` resource with `publish: true` property |
| CollectionItemPublish | `CollectionItem` resource with `isDraft: false` property |
| CollectionItemLive | `CollectionItem` resource exposes `lastPublished` timestamp |

### Application/Business Data (Not Infrastructure)
| API | Reason |
|-----|--------|
| FormSubmission | User-generated form data; operational, not infrastructure |
| Product | Business catalog data with high churn; owned by product teams |
| SKU | Product variant data; same concerns as Product |
| Order | Customer transaction data; temporal workflow, not idempotent |
| OrderFulfill | Transaction action; not declarative infrastructure |
| OrderRefund | Transaction action; not declarative infrastructure |
| Inventory | Real-time stock levels; changes independently of IaC state |

---

## Implementation Pattern

### File Structure
```
provider/
├── {resource}.go           # API client
├── {resource}_resource.go  # Pulumi resource
└── {resource}_test.go      # Tests
```

### Requirements
1. **Fetch schemas from OpenAPI spec** - exact request/response types
2. **Validate inputs** before API calls
3. **Rate limit handling** - exponential backoff on 429
4. **Idempotent delete** - 404 = success
5. **DryRun support** - return early during preview
6. **Resource ID format** - `{siteId}/{type}/{resourceId}`

### Reference Implementation
Use `provider/redirect_resource.go` as the template.

---

## Fetching OpenAPI Schemas

```bash
# Download full spec for reference
curl -o webflow-openapi.yml \
  https://raw.githubusercontent.com/webflow/openapi-spec/refs/heads/main/openapi/v2.yml

# Parse with yq for specific endpoint
yq '.paths["/sites/{site_id}/webhooks"]' webflow-openapi.yml
```
