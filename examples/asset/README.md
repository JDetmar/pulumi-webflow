# Asset Resource Examples

This directory contains examples demonstrating how to create and manage assets (images, files, documents) for Webflow sites using Pulumi in all supported languages.

## What You'll Learn

- Register asset metadata with Webflow
- Obtain presigned S3 upload URLs
- Use upload details to complete file uploads
- Organize assets into folders

## Important: Two-Step Upload Process

The Webflow Asset API uses a two-step process:

1. **Register Asset Metadata** (handled by this provider)
   - Call `webflow.Asset()` with `fileName` and `fileHash`
   - Receive `uploadUrl` and `uploadDetails` for S3 upload
   - Asset ID is assigned immediately

2. **Upload File to S3** (done separately)
   - POST to `uploadUrl` with form fields from `uploadDetails`
   - Include the actual file binary
   - The `hostedUrl` becomes accessible after upload completes

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

### C#

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
mvn clean install
pulumi stack init dev
pulumi config set webflow:siteId your-site-id --secret
pulumi up
```

## Examples Included

### 1. Basic Asset Registration

Register a single asset and get the upload URL:

```typescript
const logoAsset = new webflow.Asset("company-logo", {
  siteId: siteId,
  fileName: "logo.png",
  fileHash: "d41d8cd98f00b204e9800998ecf8427e", // MD5 hash of file
});

// Use these to upload to S3:
export const uploadUrl = logoAsset.uploadUrl;
export const uploadDetails = logoAsset.uploadDetails;
```

### 2. Asset with Folder Organization

Place assets in specific folders:

```typescript
const heroImage = new webflow.Asset("hero-image", {
  siteId: siteId,
  fileName: "hero-banner.jpg",
  fileHash: "a1b2c3d4e5f6789012345678abcdef12",
  parentFolder: "folder-id-here", // Optional folder ID
});
```

### 3. Bulk Asset Registration

Register multiple assets efficiently:

```typescript
const assets = [
  { name: "icon-home", file: "home.svg", hash: "..." },
  { name: "icon-settings", file: "settings.svg", hash: "..." },
  { name: "icon-user", file: "user.svg", hash: "..." },
];

assets.forEach((asset) => {
  new webflow.Asset(asset.name, {
    siteId: siteId,
    fileName: asset.file,
    fileHash: asset.hash,
  });
});
```

## Completing the S3 Upload

After running `pulumi up`, use the output values to upload your file:

### Using curl

```bash
# Get the outputs
UPLOAD_URL=$(pulumi stack output uploadUrl)
# uploadDetails contains form fields - extract them

# Upload the file (example with common fields)
curl -X POST "$UPLOAD_URL" \
  -F "acl=public-read" \
  -F "key=<from uploadDetails>" \
  -F "Content-Type=image/png" \
  -F "X-Amz-Credential=<from uploadDetails>" \
  -F "X-Amz-Algorithm=AWS4-HMAC-SHA256" \
  -F "X-Amz-Date=<from uploadDetails>" \
  -F "Policy=<from uploadDetails>" \
  -F "X-Amz-Signature=<from uploadDetails>" \
  -F "file=@/path/to/your/logo.png"
```

### Using the Webflow JS SDK

```javascript
// The Webflow JS SDK has a helper that handles both steps:
await client.assets.utilities.createAndUpload(siteId, {
  fileName: "logo.png",
  file: fs.readFileSync("/path/to/logo.png"),
});
```

## Generating MD5 File Hash

The `fileHash` is required and must be the MD5 hash of your file:

```bash
# Linux
md5sum logo.png | awk '{print $1}'

# macOS
md5 -q logo.png

# Windows (PowerShell)
(Get-FileHash -Algorithm MD5 logo.png).Hash.ToLower()
```

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `webflow:siteId`  | Yes      | Your Webflow site ID (stored as secret)  |

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    assetId       : "5f0c8c9e1c9d440000e8d8c4"
    uploadUrl     : "https://s3.amazonaws.com/..."
    uploadDetails : { acl: "...", key: "...", ... }
    assetUrl      : "https://s3.amazonaws.com/..."
    hostedUrl     : "https://assets.website-files.com/..."
```

## Cleanup

To remove all created assets:

```bash
pulumi destroy
pulumi stack rm dev
```

## Troubleshooting

### "fileHash is required" Error

You must provide an MD5 hash of your file content. See "Generating MD5 File Hash" above.

### "hostedUrl not accessible"

The `hostedUrl` only becomes accessible after you complete the S3 upload using `uploadUrl` and `uploadDetails`.

### "Invalid fileHash format" Error

The MD5 hash must be a 32-character hexadecimal string (lowercase or uppercase).

## Related Resources

- [Asset API Reference](../../docs/api/asset.md)
- [AssetFolder Resource](../assetfolder/)
- [Main Examples Index](../README.md)
- [AWS S3 POST Documentation](https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectPOST.html)
