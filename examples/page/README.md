# Page Data Source Examples

This directory contains examples demonstrating how to read page information from Webflow sites using Pulumi in all supported languages.

## What You'll Learn

- Retrieve all pages from a Webflow site
- Get metadata for a specific page by ID
- Access page properties (title, slug, creation date, etc.)
- Use page data in your infrastructure code

## Important Note

**Pages are READ-ONLY via the Webflow API.** Pages must be created in the Webflow Designer. This data source allows you to query existing pages and use their metadata in your infrastructure code.

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
pulumi config set siteId your-site-id --secret

# List all pages in the site
pulumi up

# Or get a specific page by ID
pulumi config set pageId your-page-id
pulumi up
```

### Python

```bash
cd python
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pulumi stack init dev
pulumi config set siteId your-site-id --secret

# List all pages in the site
pulumi up

# Or get a specific page by ID
pulumi config set pageId your-page-id
pulumi up
```

### Go

```bash
cd go
go mod download
pulumi stack init dev
pulumi config set siteId your-site-id --secret

# List all pages in the site
pulumi up

# Or get a specific page by ID
pulumi config set pageId your-page-id
pulumi up
```

## Use Cases

### 1. List All Pages

Query all pages in a site to understand your site structure, generate documentation, or reference in other resources.

```typescript
const allPages = new webflow.PageData("all-pages", {
  siteId: siteId,
  // pageId is omitted - retrieves all pages
});

export const pageList = allPages.pages;
```

### 2. Get Specific Page Details

Retrieve detailed metadata for a single page when you know its ID.

```typescript
const specificPage = new webflow.PageData("home-page", {
  siteId: siteId,
  pageId: "5f0c8c9e1c9d440000e8d8c4",
});

export const homePageTitle = specificPage.title;
export const homePageSlug = specificPage.slug;
```

### 3. Reference Pages in Other Resources

Use page information when configuring custom code, webhooks, or other resources.

```typescript
const homePage = new webflow.PageData("home-page", {
  siteId: siteId,
  pageId: homePageId,
});

// Use page data in custom code injection
const customCode = new webflow.PageCustomCode("analytics", {
  siteId: siteId,
  pageId: homePage.webflowPageId,
  headCode: `<!-- Analytics for ${homePage.title} -->`,
});
```

### 4. Filter Pages by Properties

Query all pages and filter based on their properties.

```typescript
export const draftPages = allPages.pages.apply(pages =>
  pages.filter(page => page.draft)
);

export const archivedPages = allPages.pages.apply(pages =>
  pages.filter(page => page.archived)
);
```

## Configuration

Each example requires the following configuration:

| Config Key  | Required | Description                                           |
|-------------|----------|-------------------------------------------------------|
| `siteId`    | Yes      | Your Webflow site ID (stored as secret)               |
| `pageId`    | No       | Specific page ID to retrieve (omit to list all pages) |

## Getting Your Site and Page IDs

1. **Site ID**:
   - Log in to Webflow
   - Go to Project Settings → General
   - Find your site ID (24-character hexadecimal string)

2. **Page ID**:
   - You can get page IDs by first running the example without `pageId` configured
   - The output will list all pages with their IDs
   - Or use the Webflow API to list pages

## Expected Output

### When Listing All Pages (pageId not set)

```
Outputs:
    pageCount   : 12
    pageIds     : ["5f0c...", "5f0d...", "5f0e...", ...]
    sitePages   : [
        {
            id       : "5f0c8c9e1c9d440000e8d8c4"
            title    : "Home"
            slug     : "home"
            draft    : false
            archived : false
        },
        {
            id       : "5f0c8c9e1c9d440000e8d8c5"
            title    : "About"
            slug     : "about"
            draft    : false
            archived : false
        },
        ...
    ]
```

### When Getting a Specific Page (pageId set)

```
Outputs:
    pageCollectionId : null
    pageCreatedOn    : "2024-01-15T10:30:00Z"
    pageIsArchived   : false
    pageIsDraft      : false
    pageLastUpdated  : "2024-03-20T14:22:00Z"
    pageParentId     : null
    pageSlug         : "home"
    pageTitle        : "Home"
    pageWebflowId    : "5f0c8c9e1c9d440000e8d8c4"
```

## Page Properties

Each page includes the following properties:

| Property       | Type    | Description                                        |
|----------------|---------|----------------------------------------------------|
| `pageId`       | string  | The Webflow page ID                                |
| `siteId`       | string  | The site ID this page belongs to                   |
| `title`        | string  | Page title (shown in browser tabs)                 |
| `slug`         | string  | URL slug (e.g., "about" for "/about")              |
| `parentId`     | string  | Parent page ID for nested pages (optional)         |
| `collectionId` | string  | CMS collection ID for collection pages (optional)  |
| `createdOn`    | string  | Creation timestamp (RFC3339 format)                |
| `lastUpdated`  | string  | Last update timestamp (RFC3339 format)             |
| `archived`     | boolean | Whether the page is archived                       |
| `draft`        | boolean | Whether the page is in draft mode                  |

## Cleanup

Data sources don't create resources in Webflow, so there's nothing to clean up:

```bash
pulumi destroy  # Removes from Pulumi state only
pulumi stack rm dev
```

## Troubleshooting

### "Site not found" Error

1. Verify your site ID in Webflow: Settings → General
2. Ensure correct format: 24-character lowercase hexadecimal
3. Check API token has access to the site

### "Page not found" Error

1. Verify the page exists in your Webflow site
2. Check the page ID is correct (24-character hexadecimal)
3. Try listing all pages first (omit pageId) to see available pages

### Empty Pages Array

If the pages array is empty:
1. Verify your site has published pages
2. Check that your API token has the correct scopes
3. Ensure you're using the correct site ID

## Related Resources

- [PageCustomCode Resource](../page-custom-code/)
- [Main Examples Index](../README.md)
- [Webflow Pages Documentation](https://university.webflow.com/lesson/intro-to-pages)
- [Webflow Pages API](https://developers.webflow.com/reference/pages)
