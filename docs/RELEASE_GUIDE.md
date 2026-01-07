# Release Guide for Pulumi Webflow Provider

This guide covers the complete CD (Continuous Deployment) setup, including one-time configuration and the release process.

## Overview

Your provider is configured to publish to **5 package managers** automatically when you push a git tag:

| Package Manager | Package Name | Language | Security |
|-----------------|--------------|----------|----------|
| **GitHub Releases** | `pulumi-resource-webflow` | Provider binary | SBOM included |
| **npm** | `@jdetmar/pulumi-webflow` | TypeScript/JavaScript | Trusted Publishing + Provenance |
| **PyPI** | `pulumi-webflow` | Python | Trusted Publishing + Sigstore |
| **NuGet** | `Pulumi.Webflow` | .NET/C# | Trusted Publishing |
| **Maven Central** | `com.pulumi:webflow` | Java | GPG signed |

The Go SDK is published to a separate GitHub repository branch.

### Security Features

This provider uses modern supply chain security practices:

- **npm Trusted Publishing**: No long-lived API tokens. Uses GitHub OIDC to authenticate directly with npm. Automatic provenance attestations.
- **PyPI Trusted Publishing**: No long-lived API tokens. Uses GitHub OIDC to authenticate directly with PyPI. Automatic Sigstore attestations.
- **NuGet Trusted Publishing**: No long-lived API keys. Uses GitHub OIDC to obtain short-lived, single-use API keys.
- **SBOM Generation**: Each GitHub Release includes Software Bill of Materials (`.sbom.json`) files for vulnerability scanning.
- **Auto-generated Changelog**: Release notes are automatically generated from commit messages.

---

## Part 1: One-Time Setup Guide

### Step 1: Create Package Manager Accounts

You need accounts on each package registry:

1. **npm** (https://www.npmjs.com)
   - Create account or login
   - Verify email address
   - Enable 2FA (recommended)

2. **PyPI** (https://pypi.org)
   - Create account
   - Verify email address
   - Enable 2FA (recommended)

3. **NuGet** (https://www.nuget.org)
   - Sign in with Microsoft account
   - Verify email address

4. **Maven Central / Sonatype OSSRH** (https://central.sonatype.org)
   - Create Sonatype JIRA account: https://issues.sonatype.org/secure/Signup!default.jspa
   - Open a ticket to claim your group ID (e.g., `com.pulumi` or your own)
   - Wait for approval (can take 1-2 business days)

### Step 2: Configure Trusted Publishers

#### npm Trusted Publisher (No Token Needed!)

npm uses Trusted Publishing via GitHub OIDC - no API token required.

**For a new package (first release):**

You need to publish a placeholder first. Use `npx setup-npm-trusted-publish @jdetmar/pulumi-webflow` or publish manually once with a token, then configure Trusted Publishing.

**For an existing package:**
1. Go to https://www.npmjs.com/package/@jdetmar/pulumi-webflow/access
2. Scroll to "Trusted Publishers" section
3. Click "Add Trusted Publisher"
4. Select "GitHub Actions"
5. Fill in:
   - Owner: `JDetmar`
   - Repository: `pulumi-webflow`
   - Workflow: `release.yml`
   - Environment: (leave blank)
6. Click "Add"

See: https://docs.npmjs.com/trusted-publishers/

#### PyPI Trusted Publisher (No Token Needed!)

PyPI uses Trusted Publishing via GitHub OIDC - no API token required.

**For a new package (first release):**
1. Go to https://pypi.org/manage/account/publishing/
2. Click "Add a new pending publisher"
3. Fill in:
   - PyPI Project Name: `pulumi-webflow`
   - Owner: `JDetmar`
   - Repository name: `pulumi-webflow`
   - Workflow name: `release.yml`
   - Environment name: (leave blank)
4. Click "Add"

**For an existing package:**
1. Go to https://pypi.org/manage/project/pulumi-webflow/settings/publishing/
2. Click "Add a new publisher"
3. Fill in the same details as above

See: https://docs.pypi.org/trusted-publishers/creating-a-project-through-oidc/

#### NuGet Trusted Publisher (No API Key Needed!)

NuGet uses Trusted Publishing via GitHub OIDC - no long-lived API key required.

1. Go to https://www.nuget.org/ and sign in
2. Click your username → "Trusted Publishing"
3. Click "Add new trusted publishing policy"
4. Fill in:
   - Package name: `Pulumi.Webflow`
   - Owner: `JDetmar`
   - Repository: `pulumi-webflow`
   - Workflow: `release.yml`
   - Environment: (leave blank)
5. Click "Add"

**Note:** You still need to add `NUGET_USERNAME` as a GitHub secret (your nuget.org username).

See: https://learn.microsoft.com/en-us/nuget/nuget-org/trusted-publishing

#### Java/Maven Central GPG Key

Maven Central requires signed artifacts. You need a GPG key:

```bash
# Generate GPG key
gpg --full-generate-key
# Choose: RSA and RSA, 4096 bits, no expiration
# Enter your name and email

# List keys to get the key ID
gpg --list-secret-keys --keyid-format=long
# Output looks like: sec   rsa4096/ABCD1234EFGH5678 2024-01-01 [SC]
# The key ID is: ABCD1234EFGH5678

# Export the private key (base64 encoded for GitHub secrets)
gpg --armor --export-secret-keys ABCD1234EFGH5678 | base64

# Publish public key to keyserver (required for Maven Central verification)
gpg --keyserver keyserver.ubuntu.com --send-keys ABCD1234EFGH5678
```

#### Sonatype OSSRH Credentials

1. Your JIRA username from Step 1
2. Generate a user token at https://oss.sonatype.org/#profile;User%20Token

### Step 3: Configure GitHub Repository Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions → New repository secret

Add these secrets:

| Secret Name | Value | Required |
|-------------|-------|----------|
| `NUGET_USERNAME` | Your nuget.org username | Yes |
| `OSSRH_USERNAME` | Sonatype JIRA username | Yes |
| `OSSRH_PASSWORD` | Sonatype user token | Yes |
| `JAVA_SIGNING_KEY_ID` | GPG key ID (e.g., `ABCD1234EFGH5678`) | Yes |
| `JAVA_SIGNING_KEY` | Base64-encoded GPG private key | Yes |
| `JAVA_SIGNING_PASSWORD` | GPG key passphrase | Yes |

**Note:** npm, PyPI, and NuGet all use Trusted Publishing via OIDC - no API tokens/keys needed!

**Optional secrets (for Windows binary signing):**

| Secret Name | Value | Required |
|-------------|-------|----------|
| `AZURE_SIGNING_CLIENT_ID` | Azure AD app client ID | No |
| `AZURE_SIGNING_CLIENT_SECRET` | Azure AD app secret | No |
| `AZURE_SIGNING_TENANT_ID` | Azure AD tenant ID | No |
| `AZURE_SIGNING_KEY_VAULT_URI` | Azure Key Vault URI | No |
| `SKIP_SIGNING` | Set to `true` to skip Windows signing | No |

### Step 4: Verify Workflow Permissions

Go to Settings → Actions → General:

1. **Workflow permissions**: Select "Read and write permissions"
2. **Allow GitHub Actions to create and approve pull requests**: Check this box

### Step 5: Test Your Setup (Dry Run)

Before your first real release, you can verify the build process:

```bash
# Build everything locally
make build

# Verify provider works
./bin/pulumi-resource-webflow --version

# Run tests
make test_provider
```

---

## Part 2: Release Checklist

Use this checklist every time you release a new version.

### Pre-Release Checklist

- [ ] **All code changes committed and pushed to main**
  ```bash
  git status  # Should be clean
  ```

- [ ] **Codegen is up to date**
  ```bash
  make codegen
  git status  # No changes should appear
  ```

- [ ] **Tests pass**
  ```bash
  make test_provider
  ```

- [ ] **Lint passes**
  ```bash
  make lint
  ```

- [ ] **CI is green on main branch**
  - Check: https://github.com/JDetmar/pulumi-webflow/actions

- [ ] **Examples work** (optional but recommended)
  ```bash
  cd examples/redirect/typescript
  npm install && pulumi preview
  ```

- [ ] **Version number decided**
  - Follow [Semantic Versioning](https://semver.org/):
    - `MAJOR.MINOR.PATCH` (e.g., `0.1.0`, `1.0.0`, `1.2.3`)
    - MAJOR: Breaking changes
    - MINOR: New features (backwards compatible)
    - PATCH: Bug fixes

- [ ] **Changelog will be auto-generated**
  - GoReleaser generates changelog from commit messages
  - Use conventional commit prefixes for best results: `feat:`, `fix:`, `docs:`, `chore:`

### Release Process

1. **Create and push the tag**
   ```bash
   # Ensure you're on main and up to date
   git checkout main
   git pull origin main

   # Create annotated tag
   git tag -a v0.1.0 -m "Release v0.1.0: Initial release with Site, Redirect, and RobotsTxt resources"

   # Push the tag
   git push origin v0.1.0
   ```

2. **Monitor the release workflow**
   - Go to: https://github.com/JDetmar/pulumi-webflow/actions/workflows/release.yml
   - Watch for the workflow triggered by your tag
   - Expected duration: 10-20 minutes

3. **Verify publications** (after workflow completes)

   | Registry | Verification URL |
   |----------|------------------|
   | GitHub Releases | https://github.com/JDetmar/pulumi-webflow/releases |
   | npm | https://www.npmjs.com/package/@jdetmar/pulumi-webflow |
   | PyPI | https://pypi.org/project/pulumi-webflow/ |
   | NuGet | https://www.nuget.org/packages/Pulumi.Webflow |
   | Maven | https://central.sonatype.com/artifact/com.pulumi/webflow |

### Post-Release Checklist

- [ ] **GitHub Release exists** with binaries for all platforms
- [ ] **npm package published** and version matches
- [ ] **PyPI package published** and version matches
- [ ] **NuGet package published** and version matches
- [ ] **Maven package published** (may take up to 2 hours to sync)
- [ ] **Test installation works**
  ```bash
  # TypeScript/JavaScript
  npm install @jdetmar/pulumi-webflow@0.1.0

  # Python
  pip install pulumi-webflow==0.1.0

  # .NET
  dotnet add package Pulumi.Webflow --version 0.1.0
  ```

---

## Part 3: Versioning Strategy

### Recommended Approach

For a new provider, start with `0.x.x` versions:

| Version | Meaning |
|---------|---------|
| `0.1.0` | First release - core functionality |
| `0.1.1` | Bug fixes |
| `0.2.0` | New features added |
| `1.0.0` | Stable API - production ready |

### Pre-release Versions

For testing before official release:

```bash
git tag v0.1.0-alpha.1
git tag v0.1.0-beta.1
git tag v0.1.0-rc.1
```

GoReleaser automatically detects these as pre-releases.

---

## Part 4: Troubleshooting

### Common Issues

#### npm publish fails with 403
- Check NPM_TOKEN is valid and not expired
- Ensure 2FA is configured correctly
- Verify package name isn't taken by someone else

#### PyPI publish fails
- Verify Trusted Publisher is configured on PyPI (see setup guide above)
- Check that repository name, owner, and workflow name match exactly
- For new packages, use "pending publisher" before first release
- Check if package name is available/not taken

#### NuGet publish fails
- API keys expire after 365 days - regenerate if needed
- Verify the glob pattern matches your package name

#### Maven Central publish fails
- GPG key must be published to a keyserver
- Sonatype credentials must be from user token, not password
- Group ID must be approved by Sonatype

#### GoReleaser fails
- Check GITHUB_TOKEN has write permissions
- Ensure tag format is correct (`v*.*.*`)

### Viewing Logs

1. Go to Actions → Select the failed workflow run
2. Click on the failed job
3. Expand the failed step to see detailed logs

### Manual Republish

If only some packages failed, you can:

1. Fix the issue (usually a secret)
2. Delete the tag locally and remotely:
   ```bash
   git tag -d v0.1.0
   git push origin :refs/tags/v0.1.0
   ```
3. Delete the GitHub Release (if created)
4. Re-create and push the tag

---

## Part 5: Automation Enhancements (Optional)

### Already Automated

- ✅ Multi-platform binary builds (6 platforms)
- ✅ SDK generation for 5 languages
- ✅ Package publishing to all registries
- ✅ Version extraction from git tags
- ✅ Pre-release detection
- ✅ **Changelog generation** from commit messages
- ✅ **SBOM generation** for supply chain security
- ✅ **npm Trusted Publishing** with automatic provenance
- ✅ **PyPI Trusted Publishing** with Sigstore attestations
- ✅ **NuGet Trusted Publishing** with short-lived API keys

### Could Be Added

1. **Release Approval Gates**
   - Add environment protection rules in GitHub settings

2. **Automated Version Bumping**
   - Use tools like `standard-version` or `release-please`

3. **Slack/Discord Notifications**
   - Add notification step to release workflow

---

## Quick Reference

### Release Commands

```bash
# Full build and test
make build && make test_provider && make lint

# Create release
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# Delete tag if needed
git tag -d v0.1.0 && git push origin :refs/tags/v0.1.0
```

### Required Secrets Summary

```
NUGET_USERNAME         # NuGet Trusted Publishing (username only)
OSSRH_USERNAME         # Maven Central
OSSRH_PASSWORD         # Maven Central
JAVA_SIGNING_KEY_ID    # Maven GPG signing
JAVA_SIGNING_KEY       # Maven GPG signing
JAVA_SIGNING_PASSWORD  # Maven GPG signing
```

**Note:** npm, PyPI, and NuGet all use Trusted Publishing - no API keys/tokens needed!

### Package URLs

- GitHub: https://github.com/JDetmar/pulumi-webflow/releases
- npm: https://www.npmjs.com/package/@jdetmar/pulumi-webflow
- PyPI: https://pypi.org/project/pulumi-webflow/
- NuGet: https://www.nuget.org/packages/Pulumi.Webflow
- Maven: https://central.sonatype.com/artifact/com.pulumi/webflow
