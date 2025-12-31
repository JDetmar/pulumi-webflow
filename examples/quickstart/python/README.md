# Webflow Pulumi Provider - Python Quickstart

Get started with the Webflow Pulumi Provider in Python. Deploy your first Webflow resource (RobotsTxt) in under 10 minutes.

## Prerequisites

- **Pulumi CLI** - [Install](https://www.pulumi.com/docs/install/)
- **Python** - [Install Python 3.8 or later](https://www.python.org/downloads/)
- **pip** - Comes with Python
- **Webflow account** - With API access enabled
- **Webflow API token** - [Create one](https://webflow.com) in Account Settings > API Tokens
- **Webflow Site ID** - Find in Webflow Designer > Project Settings > API & Webhooks

## Quick Start - 5 Minutes to Deployment

### Step 1: Install Dependencies (1 minute)

```bash
# Create a virtual environment (recommended)
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
```

This installs:
- `pulumi` - Pulumi SDK
- `pulumi-webflow` - Webflow provider

### Step 2: Configure Your Credentials (2 minutes)

```bash
# Set your Webflow API token (encrypted in Pulumi.dev.yaml)
pulumi config set webflow:apiToken --secret

# When prompted, paste your Webflow API token and press Enter
```

```bash
# Set your Webflow Site ID
pulumi config set siteId --secret

# When prompted, paste your 24-character site ID and press Enter
```

**Need your Site ID?**
1. Open Webflow Designer
2. Go to **Project Settings** → **API & Webhooks**
3. Copy your **Site ID** (24-character hex string like `5f0c8c9e1c9d440000e8d8c3`)

### Step 3: Initialize a Pulumi Stack (1 minute)

```bash
# Create a dev stack
pulumi stack init dev

# Or select existing stack
pulumi stack select dev
```

### Step 4: Deploy! (1 minute)

```bash
# Preview what will be created
pulumi preview

# Deploy to your Webflow site
pulumi up

# When prompted, select 'yes' to confirm
```

Expected output:
```
Previewing update (dev):

     Type                 Name            Plan
 +   webflow:RobotsTxt   my-robots       create

Resources:
    + 1 to create

Do you want to perform this update? yes

     Type                 Name            Plan      Status
 +   webflow:RobotsTxt   my-robots       create    created

Outputs:
    deployed_site_id: "5f0c8c9e1c9d440000e8d8c3"
    robots_txt_id: "xyz123"

Resources:
    + 1 created

Duration: 3s
```

### Step 5: Verify in Webflow (1 minute)

1. Open Webflow Designer
2. Go to **Project Settings** → **SEO** → **robots.txt**
3. You should see the robots.txt content you deployed!

## Code Overview

The main program is in `__main__.py`:

```python
import pulumi
import pulumi_webflow as webflow

config = pulumi.Config()
site_id = config.require_secret("siteId")

# Create a RobotsTxt resource
robots_txt = webflow.RobotsTxt(
    "my-robots",
    site_id=site_id,
    content="""User-agent: *
Allow: /""",
)

pulumi.export("deployed_site_id", site_id)
```

### Customization

**Change the robots.txt content:**

Edit the `content` parameter in `__main__.py`:

```python
robots_txt = webflow.RobotsTxt(
    "my-robots",
    site_id=site_id,
    content="""User-agent: *
Disallow: /admin/
Allow: /public/""",
)
```

Then deploy:
```bash
pulumi up
```

**Deploy to a different site:**

```bash
# Update the site ID
pulumi config set siteId --secret
# Paste your new site ID

# Deploy to the new site
pulumi up
```

## Cleanup

Remove the resource from Webflow:

```bash
pulumi destroy

# When prompted, select 'yes' to confirm
```

This removes the RobotsTxt resource from your Webflow site.

## Virtual Environment

It's recommended to use Python virtual environments to isolate dependencies:

```bash
# Create virtual environment
python3 -m venv venv

# Activate it
source venv/bin/activate  # Linux/macOS
# or
venv\Scripts\activate  # Windows

# Install dependencies
pip install -r requirements.txt

# When done, deactivate
deactivate
```

## Troubleshooting

### "Authentication failed" error

```
Error: Unauthorized - Invalid or expired Webflow API token
```

**Solution:**
1. Verify your token in Webflow Account Settings > API Tokens
2. Update your Pulumi config:
   ```bash
   pulumi config set webflow:apiToken --secret
   ```

### "Invalid site ID" error

```
Error: Invalid or malformed siteId
```

**Solution:**
1. Get the correct site ID from Webflow Designer > Project Settings > API & Webhooks
2. Update your Pulumi config:
   ```bash
   pulumi config set siteId --secret
   ```

### Plugin installation issues

```
Error: Plugin webflow not found
```

**Solution:**
```bash
pulumi plugin install resource webflow
```

### Python module not found errors

```
ModuleNotFoundError: No module named 'pulumi_webflow'
```

**Solution:**
```bash
# Make sure you've activated your virtual environment
source venv/bin/activate

# Reinstall dependencies
pip install -r requirements.txt
```

## Next Steps

- Explore other resource types (Redirects, Sites, etc.)
- Check the main [README](../../README.md) for comprehensive documentation
- View other examples in the [examples/](../) folder
- Learn Pulumi concepts at [pulumi.com/docs](https://www.pulumi.com/docs/)

## Files in This Example

- `__main__.py` - Main Pulumi program
- `Pulumi.yaml` - Project configuration
- `requirements.txt` - Python dependencies
- `.gitignore` - Files to exclude from Git
- `README.md` - This file

## Learn More

- [Webflow Pulumi Provider](../../README.md)
- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Pulumi Python SDK](https://www.pulumi.com/docs/reference/pkg/python/)
- [Webflow API Documentation](https://developers.webflow.com/)
