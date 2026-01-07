import express from "express";
import open from "open";
import * as dotenv from "dotenv";
import fetch from "node-fetch";
import crypto from "crypto";

dotenv.config();

// Configuration from environment
const CLIENT_ID = process.env.WEBFLOW_CLIENT_ID;
const CLIENT_SECRET = process.env.WEBFLOW_CLIENT_SECRET;
const REDIRECT_URI = process.env.REDIRECT_URI || "http://localhost:3000/callback";
const PORT = parseInt(process.env.PORT || "3000", 10);

// All available Webflow API scopes (site-level)
// https://developers.webflow.com/data/reference/scopes
const SCOPES = [
  "assets:read",
  "assets:write",
  "authorized_user:read",
  "cms:read",
  "cms:write",
  "comments:read",
  "comments:write",
  "components:read",
  "components:write",
  "custom_code:read",
  "custom_code:write",
  "ecommerce:read",
  "ecommerce:write",
  "forms:read",
  "forms:write",
  "pages:read",
  "pages:write",
  "sites:read",
  "sites:write",
  "site_activity:read",
  "site_config:read",
  "site_config:write",
  "users:read",
  "users:write",
  "workspace:read",
  "workspace:write",
].join(" ");

// Validate required credentials
if (!CLIENT_ID || !CLIENT_SECRET) {
  console.error("Error: Missing required environment variables");
  console.error("Please set WEBFLOW_CLIENT_ID and WEBFLOW_CLIENT_SECRET");
  console.error("See .env.example for reference");
  process.exit(1);
}

const app = express();

// Generate random state for CSRF protection
const state = crypto.randomBytes(32).toString('hex');

// OAuth callback handler
app.get("/callback", async (req, res) => {
  const { code, state: returnedState, error } = req.query;

  // Handle user denial or errors
  if (error) {
    console.error(`\nAuthorization failed: ${error}`);
    res.send("Authorization failed. You can close this window.");
    shutdown();
    return;
  }

  // Validate state parameter (CSRF protection)
  if (returnedState !== state) {
    console.error("\nState mismatch - possible CSRF attack");
    res.status(400).send("Invalid state parameter. You can close this window.");
    shutdown();
    return;
  }

  if (!code) {
    console.error("\nNo authorization code received");
    res.status(400).send("No authorization code received. You can close this window.");
    shutdown();
    return;
  }

  try {
    // Exchange authorization code for access token
    // Docs: https://developers.webflow.com/data/reference/oauth-app
    const tokenResponse = await fetch("https://api.webflow.com/oauth/access_token", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        client_id: CLIENT_ID,
        client_secret: CLIENT_SECRET,
        code: String(code),
        grant_type: "authorization_code",
        redirect_uri: REDIRECT_URI,
      }),
    });

    if (!tokenResponse.ok) {
      const errorText = await tokenResponse.text();
      throw new Error(`Token exchange failed: ${tokenResponse.status} - ${errorText}`);
    }

    const tokenData = await tokenResponse.json();
    const accessToken = tokenData.access_token;

    if (!accessToken) {
      throw new Error("No access token in response");
    }

    // Success! Show the token
    console.log("\n" + "=".repeat(60));
    console.log("SUCCESS! Your Webflow OAuth token:");
    console.log("=".repeat(60));
    console.log(`\n${accessToken}\n`);
    console.log("=".repeat(60));
    console.log("\nTo use with Pulumi, either:");
    console.log("\n1. Set environment variable:");
    console.log(`   export WEBFLOW_API_TOKEN="${accessToken}"`);
    console.log("\n2. Or set Pulumi config:");
    console.log(`   pulumi config set webflow:apiToken "${accessToken}" --secret`);
    console.log("\n" + "=".repeat(60));

    res.send(`
      <html>
        <body style="font-family: system-ui, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
          <h1 style="color: #10b981;">Authorization Successful!</h1>
          <p>Your access token has been printed to the terminal.</p>
          <p>You can close this window now.</p>
        </body>
      </html>
    `);
  } catch (err) {
    console.error("\nError exchanging code for token:", err);
    res.status(500).send("Failed to obtain access token. Check the terminal for details.");
  }

  shutdown();
});

// Graceful shutdown
let server: ReturnType<typeof app.listen>;

function shutdown() {
  console.log("\nShutting down...");
  setTimeout(() => {
    server.close();
    process.exit(0);
  }, 1000);
}

// Start server and open browser
server = app.listen(PORT, () => {
  const authUrl = new URL("https://webflow.com/oauth/authorize");
  authUrl.searchParams.set("client_id", CLIENT_ID!);
  authUrl.searchParams.set("response_type", "code");
  authUrl.searchParams.set("redirect_uri", REDIRECT_URI);
  authUrl.searchParams.set("scope", SCOPES);
  authUrl.searchParams.set("state", state);

  console.log("Webflow OAuth Authorization Tool");
  console.log("=".repeat(40));
  console.log(`\nStarting local server on port ${PORT}...`);
  console.log("Opening browser for Webflow authorization...\n");
  console.log("If browser doesn't open, visit:");
  console.log(authUrl.toString());
  console.log("\nWaiting for authorization...");

  open(authUrl.toString());
});

// Handle Ctrl+C
process.on("SIGINT", () => {
  console.log("\n\nCancelled by user.");
  process.exit(0);
});
