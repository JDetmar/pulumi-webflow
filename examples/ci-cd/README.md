# CI/CD Pipeline Integration Examples

This directory contains example configurations for integrating the Webflow Pulumi Provider into your CI/CD pipelines.

## Overview

The Webflow Pulumi Provider supports non-interactive automation in CI/CD pipelines, enabling you to:

- **Automate deployments** - Run `pulumi up --yes` without user confirmation
- **Preview changes** - Use `pulumi preview` to validate changes before applying
- **Multi-environment management** - Deploy to dev, staging, and production stacks
- **Secure credential handling** - Use CI/CD secrets for API tokens
- **Proper exit codes** - Integrate with CI pipeline notifications

## Prerequisites

1. **Pulumi Account** - Create a free account at [pulumi.com](https://pulumi.com)
2. **Webflow API Token** - Generate from [Webflow Dashboard](https://webflow.com/dashboard/settings/integrations)
3. **Pulumi Access Token** - Generate from [Pulumi Console](https://app.pulumi.com/account/tokens)
4. **Infrastructure Code** - Your Pulumi program in `infrastructure/` directory

## Setup Guides

### GitHub Actions

1. **Store Secrets in GitHub:**
   - Go to your repository Settings → Secrets and variables → Actions
   - Add `WEBFLOW_API_TOKEN`: Your Webflow API token
   - Add `PULUMI_ACCESS_TOKEN`: Your Pulumi access token

2. **Create Workflow File:**
   ```bash
   mkdir -p .github/workflows
   cp examples/ci-cd/github-actions.yaml .github/workflows/deploy.yml
   ```

3. **Configure Project Structure:**
   - Place your Pulumi program in `infrastructure/` directory
   - Ensure `infrastructure/Pulumi.yaml` exists with your stack configurations

4. **Trigger Deployment:**
   - Push to `main` branch to automatically preview and deploy
   - Or manually trigger with "Run workflow" button

### GitLab CI

1. **Store Variables in GitLab:**
   - Go to Settings → CI/CD → Variables
   - Add `WEBFLOW_API_TOKEN`: Your Webflow API token
   - Add `WEBFLOW_API_TOKEN_STAGING`: Token for staging environment
   - Add `WEBFLOW_API_TOKEN_PRODUCTION`: Token for production environment
   - Mark secrets with "Protect variable" checkbox

2. **Create CI Configuration:**
   ```bash
   cp examples/ci-cd/gitlab-ci.yaml .gitlab-ci.yml
   ```

3. **Configure Project Structure:**
   - Place your Pulumi program in `infrastructure/` directory
   - Ensure `infrastructure/Pulumi.yaml` exists

4. **Trigger Pipeline:**
   - Push to `develop` to deploy to staging
   - Push to `main` to create production deployment option
   - Manually approve production deployment when ready

## CI/CD Patterns

### 1. Non-Interactive Deployment

```bash
# This runs without prompting for confirmation
pulumi up --yes --stack prod
```

**Key Flags:**
- `--yes`: Skip confirmation prompts (required for CI/CD)
- `--stack STACKNAME`: Select target stack
- `--refresh`: Refresh state before deployment (optional)
- `--parallel N`: Run operations in parallel (improves speed)

### 2. Preview-First Approach

```bash
# Preview changes without applying
pulumi preview --stack dev

# If preview looks good, deploy
pulumi up --yes --stack dev
```

**Benefits:**
- See what will change before applying
- Catch configuration errors early
- Better change tracking for compliance

### 3. Multi-Environment Management

```yaml
# Example with dev/staging/prod stacks
stages:
  - preview    # All branches preview against dev
  - deploy     # develop → staging, main → prod

# Environment-specific configuration
variables:
  WEBFLOW_API_TOKEN_DEV: ${{ secrets.WEBFLOW_API_TOKEN_DEV }}
  WEBFLOW_API_TOKEN_STAGING: ${{ secrets.WEBFLOW_API_TOKEN_STAGING }}
  WEBFLOW_API_TOKEN_PROD: ${{ secrets.WEBFLOW_API_TOKEN_PROD }}
```

### 4. Credential Management

**Environment Variables:**
```bash
export WEBFLOW_API_TOKEN=your_api_token
pulumi up --yes
```

**Pulumi Configuration:**
```bash
pulumi config set webflow:apiToken $WEBFLOW_API_TOKEN --secret
```

**Best Practices:**
✅ Store tokens in CI/CD secrets management
✅ Never commit tokens to git
✅ Use environment variables or config files
✅ Verify tokens never appear in logs

### 5. Error Handling and Exit Codes

Pulumi automatically returns proper exit codes:

- **0**: Operation successful
- **1**: Operation failed or blocked
- **255**: Error occurred (resource errors, API failures)

**Example Error Handling:**
```yaml
- name: Deploy
  run: pulumi up --yes --stack prod
  env:
    WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
  continue-on-error: true

- name: Notify on Failure
  if: failure()
  run: echo "Deployment failed - check logs"
```

## Acceptance Criteria Validation

### AC1: Non-Interactive Execution
✅ `pulumi up --yes` runs without prompts
✅ Exit codes properly indicate success/failure
✅ Output formatted for CI/CD log parsing

**Testing:**
```bash
# Test non-interactive mode locally
pulumi up --yes --stack test-env
echo "Exit code: $?"
```

### AC2: Secure Credential Management
✅ Credentials retrieved from environment variables
✅ Credentials never logged to output
✅ Follows CI/CD secrets best practices

**Verification:**
```bash
# Confirm no tokens in logs
pulumi up --yes 2>&1 | grep -i "token" || echo "No tokens exposed"
```

## Troubleshooting

### Provider Authentication Fails
**Problem:** "Invalid API token" error
```
Error: authentication failed - invalid token
```

**Solution:**
1. Verify token is correctly set: `echo $WEBFLOW_API_TOKEN`
2. Check token permissions in Webflow settings
3. Ensure token hasn't expired
4. Try regenerating token in Webflow dashboard

### Non-Interactive Mode Still Prompts
**Problem:** Pulumi asks for confirmation despite `--yes` flag
```
Please confirm that you want to proceed: (yes/no)
```

**Solution:**
- Add `PULUMI_SKIP_CONFIRMATIONS: true` to environment
- Verify using `pulumi --version` to confirm CLI version
- Check for custom resource providers that override behavior

### Deployment Timeout
**Problem:** Pipeline times out during deployment
```
Error: operation timed out after 10 minutes
```

**Solution:**
1. Increase timeout in CI configuration:
   ```yaml
   timeout-minutes: 30
   ```
2. Check Webflow API status
3. Use `--parallel N` to speed up operations
4. Consider splitting into smaller stacks

### Credential Leakage
**Problem:** Token appears in logs
```
2024-01-15 10:23:45 DEBUG: Using token abc123xyz...
```

**Solution:**
1. Verify masking in CI platform settings
2. Check application code for logging tokens
3. Review logs at `~/.pulumi/logs`
4. Use `--suppress-outputs` to hide sensitive values

## Real-World Example

### Multi-Site Deployment with Approvals

```yaml
# GitLab CI example with approval gates
deploy_all_sites:
  stage: deploy
  script:
    - cd infrastructure
    - npm install

    # Deploy main sites
    - pulumi up --yes --stack main-sites

    # Deploy marketing sites (with approval required)
    - pulumi up --yes --stack marketing-sites

    # Deploy client sites (requires manual approval)
    - pulumi up --yes --stack client-sites
  only:
    - main
  when: manual
  environment:
    name: production
```

## Security Best Practices

1. **Rotate Tokens Regularly**
   - Regenerate API tokens quarterly
   - Update CI/CD secrets immediately

2. **Use Environment-Specific Tokens**
   - Separate tokens for dev/staging/prod
   - Limit token permissions to necessary operations

3. **Audit Trail**
   - Keep git history of infrastructure changes
   - Monitor Pulumi activity logs

4. **Deployment Approvals**
   - Require manual approval for production
   - Use branch protection rules
   - Implement code review process

5. **Monitoring**
   - Alert on deployment failures
   - Track deployment frequency and success rate
   - Monitor API rate limits

## Additional Resources

- [Pulumi Documentation - Automation API](https://www.pulumi.com/docs/concepts/automation-api/)
- [GitHub Actions Guide](https://docs.github.com/en/actions)
- [GitLab CI/CD Guide](https://docs.gitlab.com/ee/ci/)
- [Webflow API Documentation](https://webflow.com/developers)
- [Pulumi Automation Best Practices](https://www.pulumi.com/docs/guides/continuous-delivery/github-actions/)

## Support

For issues with:
- **Webflow Provider**: [GitHub Issues](https://github.com/jdetmar/pulumi-webflow/issues)
- **Pulumi Platform**: [Pulumi Community Slack](https://pulumi-community.slack.com/)
- **CI/CD Integration**: Consult your platform documentation

## Example Workflows Summary

| Platform | File | Features | Setup Time |
|----------|------|----------|-----------|
| GitHub Actions | `github-actions.yaml` | Preview, multi-env, manual approval | 5 min |
| GitLab CI | `gitlab-ci.yaml` | Staging/prod, rollback, approval | 5 min |
| Jenkins | Custom | Flexible, requires groovy | 15 min |
| CircleCI | Custom | Orbs-based, modern config | 10 min |

Start with the provided examples and customize for your environment!
