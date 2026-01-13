# CollectionField Resource Examples

This directory contains examples demonstrating how to create and manage fields for Webflow CMS collections using Pulumi.

## What You'll Learn

- Create various field types (PlainText, RichText, Number, DateTime, Switch, Email, Image, Phone, Color)
- Set up field validations (min/max for numbers, maxLength for text)
- Configure required vs. optional fields
- Add help text for content editors
- Use auto-generated slugs vs. custom slugs

## Available Languages

| Language   | Directory    | Entry Point    | Dependencies        |
|------------|--------------|----------------|---------------------|
| TypeScript | `typescript/`| `index.ts`     | `package.json`      |

## Quick Start

### TypeScript

```bash
cd typescript
npm install
pulumi stack init dev
pulumi config set webflow-collectionfield-example:collectionId your-collection-id --secret
pulumi up
```

## Prerequisites

Before running this example, you need:

1. **A Webflow Site**: Create a site in Webflow
2. **A Webflow Collection**: Create a collection in your site (Designer → CMS → Create Collection)
3. **Collection ID**: Get the collection ID from the Webflow API or dashboard (24-character hex string, e.g., `5f0c8c9e1c9d440000e8d8c3`)
4. **API Token**: Set `WEBFLOW_API_TOKEN` environment variable with your Webflow API token

### Finding Your Collection ID

Option 1 - Via Webflow API:
```bash
curl -H "Authorization: Bearer YOUR_API_TOKEN" \
     https://api.webflow.com/v2/sites/YOUR_SITE_ID/collections
```

Option 2 - From Collection URL in Webflow Designer:
The collection ID is in the URL when editing a collection: `https://webflow.com/design/your-site/cms/collections/COLLECTION_ID`

## Examples Included

### 1. Plain Text Field (Required)

A single-line text input with character limit validation. Best for titles, names, and short descriptions.

```typescript
type: "PlainText"
isRequired: true
validations: { maxLength: 100 }
```

### 2. Rich Text Field

Multi-line rich text editor for formatted content. Best for blog posts, articles, and long descriptions.

```typescript
type: "RichText"
isRequired: true
```

### 3. Number Field with Validations

Numeric input with min/max constraints. Best for prices, quantities, ratings, or read times.

```typescript
type: "Number"
validations: { min: 1, max: 120, decimalPlaces: 0 }
```

### 4. DateTime Field

Date and time picker. Best for publish dates, event dates, or deadlines.

```typescript
type: "DateTime"
isRequired: true
```

### 5. Switch Field (Boolean)

Toggle switch for true/false values. Best for feature flags or visibility controls.

```typescript
type: "Switch"
isRequired: false
```

### 6. Email Field

Email address input with built-in validation. Best for contact information.

```typescript
type: "Email"
```

### 7. Image Field

Single image reference. Best for cover images, thumbnails, or hero images.

```typescript
type: "Image"
```

### 8. Phone Field

Phone number input. Best for contact numbers.

```typescript
type: "Phone"
```

### 9. Color Field

Color picker. Best for theme colors or branding elements.

```typescript
type: "Color"
```

### 10. Auto-Generated Slug

When you don't specify a slug, Webflow automatically generates one from the displayName.

```typescript
displayName: "Short Description"
// slug will be auto-generated as "short-description"
```

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `collectionId`    | Yes      | Your Webflow collection ID (24-char hex) |
| `environment`     | No       | Deployment environment (default: development) |

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    deployedCollectionId       : [secret]
    titleFieldId               : "abc123..."
    contentFieldId             : "def456..."
    readTimeFieldId            : "ghi789..."
    publishDateFieldId         : "jkl012..."
    featuredFieldId            : "mno345..."
    authorEmailFieldId         : "pqr678..."
    coverImageFieldId          : "stu901..."
    shortDescriptionFieldId    : "vwx234..."
    phoneFieldId               : "yza567..."
    accentColorFieldId         : "bcd890..."
    summary                    : "✅ Successfully created 10 collection fields:
                                   1. Article Title
                                   2. Article Content
                                   ..."
```

## Field Type Reference

The CollectionField resource supports the following field types:

| Type            | Description                                  | Common Validations                |
|-----------------|----------------------------------------------|-----------------------------------|
| PlainText       | Single-line text input                       | maxLength                         |
| RichText        | Rich text editor with formatting             | -                                 |
| Number          | Numeric input                                | min, max, decimalPlaces           |
| DateTime        | Date and time picker                         | -                                 |
| Switch          | Boolean toggle (true/false)                  | -                                 |
| Email           | Email address with validation                | -                                 |
| Phone           | Phone number input                           | -                                 |
| Color           | Color picker                                 | -                                 |
| Image           | Single image reference                       | -                                 |
| MultiImage      | Multiple image references                    | -                                 |
| Video           | Video embed or upload                        | -                                 |
| Link            | URL/link input                               | -                                 |
| File            | File upload                                  | -                                 |
| Option          | Dropdown/select field                        | options (array of choices)        |
| Reference       | Reference to another collection item         | collectionId                      |
| MultiReference  | Multiple references to collection items      | collectionId                      |

## Important Notes

### Field Type Cannot Change

⚠️ **IMPORTANT**: The `type` field cannot be changed after creation. Changing it requires replacement (delete + recreate). This will result in data loss for existing collection items.

### Slug Generation

- If you don't provide a `slug`, Webflow auto-generates one from `displayName`
- Slugs are used in API requests and exports
- Once created, slugs can be updated

### Field Editability

Some system fields may not be editable (`isEditable: false`). This is determined by Webflow and returned as a read-only output property.

## Cleanup

To remove all created fields:

```bash
pulumi destroy
pulumi stack rm dev
```

⚠️ **Warning**: Deleting collection fields will also delete all data in those fields for existing collection items. This action cannot be undone.

## Troubleshooting

### "Collection not found" Error

1. Verify your collection ID is correct (24-character hex string)
2. Ensure the collection exists in your Webflow site
3. Check API token has access to the site

### "Field already exists" Error

A field with the same slug already exists. Either:
1. Import the existing field: `pulumi import webflow:index:CollectionField name collectionId/fieldId`
2. Use a different slug
3. Delete the existing field in Webflow first

### "Invalid field type" Error

Ensure you're using one of the supported field types listed in the Field Type Reference above. Field types are case-sensitive.

### "Validation failed" Error

Check that your validations match the field type:
- Number fields: Use `min`, `max`, `decimalPlaces`
- PlainText/RichText: Use `maxLength`
- Option fields: Use `options` array

## Related Resources

- [Collection Resource Examples](../collection/)
- [CollectionItem Resource Examples](../collectionitem/)
- [Main Examples Index](../README.md)
- [Webflow CMS Documentation](https://university.webflow.com/lesson/intro-to-the-cms)
- [Webflow Collection Fields API](https://developers.webflow.com/reference/collection-fields)

## Next Steps

After creating collection fields, you can:
1. Create collection items with the [CollectionItem resource](../collectionitem/)
2. Query and manage collections with the [Collection resource](../collection/)
3. Set up webhooks to track content changes with the [Webhook resource](../webhook/)
