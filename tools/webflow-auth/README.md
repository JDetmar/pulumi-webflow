# Webflow OAuth Token Generator

One-time OAuth authorization tool to obtain an access token for the Webflow Pulumi Provider.

## Why This Tool?

Webflow's API requires OAuth tokens for multi-site management. This tool handles the interactive OAuth flow once, giving you a token you can use with the Pulumi provider.

**Key fact**: Webflow OAuth tokens currently do not expire, so you only need to run this once per workspace.

## Prerequisites

1. **Create a Webflow App** in your Workspace:
   - Go to Workspace Settings > Apps & Integrations
   - Click "Create App" (or use an existing one)
   - Note the **Client ID** and **Client Secret**
   - Set the Redirect URI to: `http://localhost:3000/callback`

2. **Configure OAuth Scopes** for your app:
   - `sites:read` - List and read site information
   - `sites:write` - Publish sites
   - `pages:read` - List pages
   - `pages:write` - Manage pages
   - `custom_code:read` - Read custom code
   - `custom_code:write` - Manage custom code

## Setup

```bash
cd tools/webflow-auth
npm install
cp .env.example .env
```

Edit `.env` with your Webflow App credentials:

```
WEBFLOW_CLIENT_ID=your_client_id_here
WEBFLOW_CLIENT_SECRET=your_client_secret_here
```

## Usage

```bash
npm start
```

This will:
1. Start a local server on port 3000
2. Open your browser to the Webflow authorization page
3. After you authorize, exchange the code for an access token
4. Print the token to your terminal

## Using the Token

Once you have the token, use it with the Pulumi provider:

### Option 1: Environment Variable

```bash
export WEBFLOW_API_TOKEN="your_token_here"
pulumi up
```

### Option 2: Pulumi Config

```bash
pulumi config set webflow:apiToken "your_token_here" --secret
```

### Option 3: Pulumi ESC (Recommended for teams)

```yaml
# environments/webflow.yaml
values:
  webflow:
    apiToken:
      fn::secret: "your_token_here"
  pulumiConfig:
    webflow:apiToken: ${webflow.apiToken}
```

## Modifying Scopes

If you need different scopes, edit the `SCOPES` array in `src/index.ts`:

```typescript
const SCOPES = [
  "sites:read",
  "sites:write",
  // Add or remove scopes as needed
].join(" ");
```

## Troubleshooting

**"Missing required environment variables"**
- Make sure `.env` file exists with `WEBFLOW_CLIENT_ID` and `WEBFLOW_CLIENT_SECRET`

**"Invalid redirect URI"**
- Check that your Webflow App has `http://localhost:3000/callback` as an allowed redirect URI

**"Authorization failed"**
- You may have denied the authorization request. Run the tool again and click "Authorize"

**Token not working with provider**
- Verify the token has the required scopes for the resources you're managing
- Check that the token was authorized for the correct workspace/sites
