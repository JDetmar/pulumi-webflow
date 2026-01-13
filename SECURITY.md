# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Security Advisories (Preferred)**: Use the [Security Advisories](https://github.com/JDetmar/pulumi-webflow/security/advisories/new) feature to privately report the vulnerability.

2. **Email**: Contact the maintainer directly at the email address listed in the repository.

### What to Include

When reporting a vulnerability, please include:

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact of the vulnerability
- Any suggested fixes (if applicable)

### Response Timeline

- **Initial Response**: Within 48 hours of receiving your report
- **Status Update**: Within 7 days with an assessment and remediation plan
- **Resolution**: Depending on complexity, typically within 30 days

### What to Expect

1. **Acknowledgment**: We will acknowledge receipt of your report within 48 hours.
2. **Assessment**: We will investigate and assess the severity of the issue.
3. **Communication**: We will keep you informed of our progress.
4. **Resolution**: Once fixed, we will release a patch and credit you (unless you prefer to remain anonymous).
5. **Disclosure**: We will coordinate with you on public disclosure timing.

## Security Best Practices for Users

### API Token Security

- **Never commit API tokens** to version control
- Use `pulumi config set webflow:apiToken <token> --secret` to securely store tokens
- Alternatively, use the `WEBFLOW_API_TOKEN` environment variable
- Rotate tokens regularly
- Use tokens with minimal required permissions

### Infrastructure Security

- Review Pulumi state files for sensitive data before sharing
- Use Pulumi's built-in encryption for secrets
- Consider using Pulumi Cloud or a secure backend for state storage

## Security Features

This provider implements several security measures:

- **TLS 1.2+**: All API communications enforce TLS 1.2 or higher
- **Token Redaction**: API tokens are never logged in plain text
- **Input Validation**: All inputs are validated before API calls
- **Rate Limiting**: Automatic retry with backoff for rate-limited requests
- **SBOM Generation**: Software Bill of Materials included with releases
- **SLSA Provenance**: Build provenance attestations for Go binaries (verifiable with `gh attestation verify`)
- **Signed Package Releases**: npm and PyPI packages published with Sigstore attestations

## Acknowledgments

We appreciate the security research community's efforts in helping keep this project secure. Contributors who report valid security issues will be acknowledged here (with permission).
