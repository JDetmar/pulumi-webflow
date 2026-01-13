# Collection Resource Examples

This directory contains examples demonstrating how to create and manage CMS collections for Webflow sites using Pulumi in all supported languages.

## What You'll Learn

- Create CMS collections for structured content (blog posts, products, team members, etc.)
- Use required properties: `siteId`, `displayName`, `singularName`
- Control URL slugs with the optional `slug` property
- Understand collection lifecycle (collections require replacement for any changes)
- Access read-only timestamps: `createdOn`, `lastUpdated`

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
pulumi config set webflow:siteId your-site-id --secret
pulumi up
```

### Python

```bash
cd python
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pulumi stack init dev
pulumi config set webflow:siteId your-site-id --secret
pulumi up
```

### Go

```bash
cd go
go mod download
pulumi stack init dev
pulumi config set webflow:siteId your-site-id --secret
pulumi up
```

### C# (.NET)

```bash
cd csharp
dotnet restore
pulumi stack init dev
pulumi config set webflow:siteId your-site-id --secret
pulumi up
```

### Java

```bash
cd java
mvn install
pulumi stack init dev
pulumi config set webflow:siteId your-site-id --secret
pulumi up
```

## Examples Included

### 1. Blog Posts Collection

Standard blog content collection with explicit slug.

```typescript
displayName: "Blog Posts"
singularName: "Blog Post"
slug: "blog-posts"
```

### 2. Products Collection (Auto-Generated Slug)

Demonstrates omitting the slug to let Webflow auto-generate it from displayName.

```typescript
displayName: "Products"
singularName: "Product"
// slug auto-generated as "products"
```

### 3. Team Members Collection

Shows custom slug different from display name.

```typescript
displayName: "Team Members"
singularName: "Team Member"
slug: "team"
```

### 4. Portfolio Items Collection

Another common use case for showcasing creative work.

```typescript
displayName: "Portfolio Items"
singularName: "Portfolio Item"
slug: "portfolio"
```

### 5. Dynamic Environment-Based Collection

Creates collections with environment-specific naming for multi-stage deployments.

```typescript
displayName: "Test Collection (development)"
singularName: "Test Item"
slug: "test-development"
```

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `webflow:siteId`  | Yes      | Your Webflow site ID (stored as secret)  |
| `environment`     | No       | Deployment environment (default: development) |

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    deployedSiteId           : [secret]
    blogCollectionId         : "6123abc..."
    blogCollectionName       : "Blog Posts"
    blogCollectionSlug       : "blog-posts"
    blogCollectionCreatedOn  : "2025-01-06T12:00:00Z"
    productsCollectionId     : "6124def..."
    teamCollectionId         : "6125ghi..."
    portfolioCollectionId    : "6126jkl..."
    testCollectionId         : "6127mno..."
    allCollections           : "Blog Posts, Products, Team Members, Portfolio Items, Test Collection (development)"
```

## Important Notes

### Collections Do Not Support Updates

**CRITICAL**: Webflow's API does not provide an update endpoint for collections. Any changes to collection properties (`displayName`, `singularName`, `slug`, or `siteId`) will trigger a replacement (delete + recreate).

This means:
- Changing a collection's name will delete and recreate it
- All collection items (blog posts, products, etc.) will be lost
- Plan carefully before creating collections in production

### Replacement Behavior

When you modify any collection property:

```bash
~ webflow:index:Collection: (replace)
    [urn=urn:pulumi:dev::example::webflow:index/collection:Collection::blog-posts-collection]
    [id=site_abc123:collection_def456]
  - displayName: "Blog Posts"
  + displayName: "Articles"  # This change triggers replacement
```

### Collection Items

This example creates empty collections. To add content items (blog posts, products, etc.), use the `webflow.CollectionItem` resource (see [CollectionItem examples](../collectionitem/)).

## Cleanup

To remove all created collections:

```bash
pulumi destroy
pulumi stack rm dev
```

**WARNING**: Destroying collections will also delete all collection items (content) within them. This operation cannot be undone.

## Troubleshooting

### "Site not found" Error

1. Verify your site ID in Webflow: Settings → General
2. Ensure correct format: 24-character lowercase hexadecimal (e.g., `5f0c8c9e1c9d440000e8d8c3`)
3. Check API token has access to the site

### "Collection already exists" Error

If Webflow already has a collection with the same slug:
1. Import the existing collection: `pulumi import webflow:index:Collection name siteId:collectionId`
2. Use a different slug for your new collection
3. Delete the existing collection in Webflow first (only if safe to do so)

### "Validation failed" Errors

Common validation issues:
- **Invalid siteId**: Must be 24-character lowercase hexadecimal
- **DisplayName too long**: Maximum 255 characters
- **SingularName too long**: Maximum 255 characters
- **Invalid slug format**: Must be URL-friendly (lowercase, hyphens, no spaces)

## Next Steps

After creating collections, you'll typically want to:

1. **Add Collection Fields**: Use `webflow.CollectionField` to define custom fields (title, body, images, etc.)
2. **Add Collection Items**: Use `webflow.CollectionItem` to populate collections with content
3. **Configure Collection Settings**: Adjust SEO, templates, and publishing options in Webflow Designer

## Related Resources

- [CollectionField Examples](../collectionfield/) - Define custom fields for your collections
- [CollectionItem Examples](../collectionitem/) - Add content items to collections
- [Main Examples Index](../README.md)
- [Webflow CMS Documentation](https://university.webflow.com/lesson/intro-to-the-cms)
- [Webflow Collections API](https://developers.webflow.com/reference/collections)

## Understanding Collections vs Collection Items

**Collections** are like database tables or content types:
- Define the structure and schema
- Created with `webflow.Collection`
- Cannot be updated once created

**Collection Items** are like database rows or content entries:
- Individual pieces of content (a blog post, a product)
- Created with `webflow.CollectionItem`
- Can be updated after creation

This example shows how to create the collection "containers" - see the CollectionItem examples for adding actual content.
