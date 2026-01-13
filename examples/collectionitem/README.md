# CollectionItem Resource Examples

This directory contains examples demonstrating how to create and manage CMS collection items for Webflow collections using Pulumi in all supported languages.

## What You'll Learn

- Create blog posts, products, or other CMS content items
- Manage draft and published states
- Set custom field data based on your collection schema
- Archive and unarchive collection items
- Work with localized content (optional)

## Available Languages

| Language   | Directory    | Entry Point    | Dependencies        |
|------------|--------------|----------------|---------------------|
| TypeScript | `typescript/`| `index.ts`     | `package.json`      |
| Python     | `python/`    | `__main__.py`  | `requirements.txt`  |
| Go         | `go/`        | `main.go`      | `go.mod`            |
| C#         | `csharp/`    | `Program.cs`   | `.csproj`           |
| Java       | `java/`      | `App.java`     | `pom.xml`           |

## Quick Start

### TypeScript

```bash
cd typescript
npm install
pulumi stack init dev
pulumi config set collectionId your-collection-id-here
pulumi up
```

### Python

```bash
cd python
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pulumi stack init dev
pulumi config set collectionId your-collection-id-here
pulumi up
```

### Go

```bash
cd go
go mod download
pulumi stack init dev
pulumi config set collectionId your-collection-id-here
pulumi up
```

### C# (.NET)

```bash
cd csharp
dotnet restore
pulumi stack init dev
pulumi config set collectionId your-collection-id-here
pulumi up
```

### Java

```bash
cd java
mvn install
pulumi stack init dev
pulumi config set collectionId your-collection-id-here
pulumi up
```

## Examples Included

### 1. Draft Blog Post

Create a blog post in draft mode (not published to the live site).

```
fieldData: {
  name: "Getting Started with Webflow CMS",
  slug: "getting-started-webflow-cms",
  content: "Learn how to use Webflow CMS..."
}
isDraft: true
```

### 2. Published Product

Create a published product with custom fields.

```
fieldData: {
  name: "Premium Widget",
  slug: "premium-widget",
  price: 99.99,
  description: "The best widget on the market"
}
isDraft: false
```

### 3. Archived Content

Create an item that is archived (hidden from the live site but retained in CMS).

```
isArchived: true
```

### 4. Bulk Content Creation

Efficiently create multiple collection items at once.

## Prerequisites

Before running these examples, you need:

1. **Pulumi CLI** installed ([Get started](https://www.pulumi.com/docs/get-started/))
2. **Webflow API Token** set as `WEBFLOW_API_TOKEN` environment variable
3. **Collection ID** from your Webflow site
   - Find this in the Webflow dashboard under your Collection settings
   - Format: 24-character hexadecimal string (e.g., `5f0c8c9e1c9d440000e8d8c3`)

### Finding Your Collection ID

1. Log into your Webflow dashboard
2. Navigate to your site's CMS Collections
3. Select the collection you want to use
4. The collection ID appears in the URL or collection settings

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                                   |
|-------------------|----------|-----------------------------------------------|
| `collectionId`    | Yes      | Your Webflow collection ID                    |
| `environment`     | No       | Deployment environment (default: development) |

## Important Notes

### Field Data Schema

The `fieldData` property must match your collection's schema. Common fields include:

- **name** (required): Display name of the item
- **slug** (required): URL-friendly identifier
- Custom fields you've added to your collection (text, rich text, number, date, etc.)

Example:
```typescript
{
  name: "My Blog Post",
  slug: "my-blog-post",
  "post-body": "Content here...",  // Custom rich text field
  "author": "John Doe",            // Custom text field
  "publish-date": "2025-01-06"     // Custom date field
}
```

### Draft vs Published

- `isDraft: true` - Item exists in CMS but is NOT visible on the live site
- `isDraft: false` - Item is published and visible on the live site
- Default: `true` (draft mode)

### Archived Items

- `isArchived: true` - Item is hidden from both CMS and live site but retained for records
- `isArchived: false` - Item is active in CMS
- Default: `false` (not archived)

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    collectionId          : "abc123def456..."
    draftPostId           : "item_abc123..."
    draftPostItemId       : "670d3e4a..."
    publishedProductId    : "item_def456..."
    publishedProductItemId: "670d3e4b..."
    archivedItemId        : "item_ghi789..."
    bulkItemIds           : ["item_jkl...", "item_mno...", "item_pqr..."]
```

## Cleanup

To remove all created collection items:

```bash
pulumi destroy
pulumi stack rm dev
```

## Troubleshooting

### "Collection not found" Error

1. Verify your collection ID in Webflow CMS settings
2. Ensure correct format: 24-character hexadecimal string
3. Check API token has access to the site and collection

### "Invalid field data" Error

The field names in `fieldData` must match your collection schema exactly:
1. Check field slugs in Webflow CMS Collection settings
2. Ensure all required fields are provided (usually "name" and "slug")
3. Verify field types match (e.g., numbers for number fields, strings for text)

### "Item already exists with this slug" Error

Slugs must be unique within a collection. Either:
1. Use a different slug value
2. Import the existing item: `pulumi import webflow:index:CollectionItem name collectionId/itemId`
3. Delete the existing item in Webflow first

## Related Resources

- [Collection Resource Example](../collection/)
- [Main Examples Index](../README.md)
- [Webflow CMS Documentation](https://university.webflow.com/lesson/intro-to-the-cms)
- [Webflow Collection Items API](https://developers.webflow.com/reference/collection-items)
