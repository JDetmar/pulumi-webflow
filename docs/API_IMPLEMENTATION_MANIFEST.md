# Webflow API Implementation Manifest

## Official API Specification

**OpenAPI Spec:** https://raw.githubusercontent.com/webflow/openapi-spec/refs/heads/main/openapi/v2.yml

**Base URL:** `https://api.webflow.com/v2`

**Rate Limit:** 60 req/min (respect `X-RateLimit-Remaining` header)

---

## Implementation Status

### ✅ Completed
| Resource | Files |
|----------|-------|
| Site | `site_resource.go`, `site.go` |
| Redirect | `redirect_resource.go`, `redirect.go` |
| RobotsTxt | `robotstxt_resource.go`, `robotstxt.go` |

---

## 🚀 APIs to Implement

### Sites
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| SitePublish | POST | `/sites/{site_id}/publish` | sites:write |
| CustomDomain | GET | `/sites/{site_id}/custom_domains` | sites:read |

### Pages
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| Page | GET, GET | `/sites/{site_id}/pages`, `/pages/{page_id}` | pages:read |
| PageContent | GET, PUT | `/pages/{page_id}/dom` | pages:read/write |

### CMS Collections
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| Collection | GET, GET, POST, DELETE | `/sites/{site_id}/collections`, `/collections/{id}` | cms:read/write |
| CollectionField | POST, PUT, DELETE | `/collections/{id}/fields` | cms:write |

### CMS Items
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| CollectionItem | GET, GET, POST, PATCH, DELETE | `/collections/{id}/items` | cms:read/write |
| CollectionItemLive | GET | `/collections/{id}/items/{item_id}/live` | cms:read |
| CollectionItemPublish | POST | `/collections/{id}/items/publish` | cms:write |

### Assets
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| Asset | GET, GET, POST, DELETE | `/sites/{site_id}/assets`, `/assets/{id}` | assets:read/write |
| AssetFolder | GET, GET, POST | `/sites/{site_id}/asset_folders` | assets:read/write |

### Custom Code
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| SiteCustomCode | GET, PUT, DELETE | `/sites/{site_id}/custom_code` | custom_code:read/write |
| PageCustomCode | GET, PUT, DELETE | `/pages/{page_id}/custom_code` | custom_code:read/write |
| RegisteredScript | GET, POST (inline), POST (hosted) | `/sites/{site_id}/registered_scripts` | custom_code:read/write |

### Forms
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| Form | GET, GET | `/sites/{site_id}/forms`, `/forms/{id}` | forms:read |
| FormSubmission | GET, GET, PATCH | `/forms/{id}/submissions` | forms:read/write |

### Users & Access
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| User | GET, GET, POST, PATCH, DELETE | `/sites/{site_id}/users` | users:read/write |
| AccessGroup | GET | `/sites/{site_id}/accessgroups` | users:read |

### Webhooks
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| Webhook | GET, GET, POST, DELETE | `/sites/{site_id}/webhooks`, `/webhooks/{id}` | sites:read/write |

### E-commerce
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| Product | GET, GET, POST, PATCH | `/sites/{site_id}/products` | ecommerce:read/write |
| SKU | POST, PATCH | `/sites/{site_id}/products/{id}/skus` | ecommerce:write |
| Order | GET, GET, PATCH | `/sites/{site_id}/orders` | ecommerce:read/write |
| OrderFulfill | POST | `/sites/{site_id}/orders/{id}/fulfill` | ecommerce:write |
| OrderRefund | POST | `/sites/{site_id}/orders/{id}/refund` | ecommerce:write |
| Inventory | GET, PATCH | `/collections/{id}/items/{item_id}/inventory` | ecommerce:read/write |
| EcommerceSettings | GET | `/sites/{site_id}/ecommerce/settings` | ecommerce:read |

### Meta/Token
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| AuthorizedUser | GET | `/token/authorized_by` | authorized_user:read |
| TokenInfo | GET | `/token/introspect` | - |

### Enterprise
| Resource | Methods | Endpoints | Scope |
|----------|---------|-----------|-------|
| WorkspaceAuditLog | GET | `/workspaces/{id}/audit_logs` | workspace_activity:read |
| SiteActivityLog | GET | `/sites/{id}/activity_logs` | site_activity:read |

---

## Recommended Implementation Batches

### Batch 1: Content Management (4 parallel agents)
- **Agent A:** Collection + CollectionField
- **Agent B:** CollectionItem + CollectionItemLive
- **Agent C:** Page + PageContent
- **Agent D:** Asset + AssetFolder

### Batch 2: Site Configuration (3 parallel agents)
- **Agent E:** Webhook
- **Agent F:** SiteCustomCode + PageCustomCode
- **Agent G:** RegisteredScript

### Batch 3: Forms & Users (2 parallel agents)
- **Agent H:** Form + FormSubmission
- **Agent I:** User + AccessGroup

### Batch 4: E-commerce (2 parallel agents)
- **Agent J:** Product + SKU + Inventory
- **Agent K:** Order + OrderFulfill + OrderRefund

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
