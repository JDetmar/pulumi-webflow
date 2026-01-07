# Security & Best Practices Analysis

This document analyzes the current release workflow against 2025 best practices for each package manager.

## Executive Summary

| Package Manager | Current Approach | Best Practice | Gap |
|-----------------|------------------|---------------|-----|
| **npm** | Long-lived token | Trusted Publishing + Provenance | ⚠️ Major |
| **PyPI** | Long-lived token | Trusted Publishing + Attestations | ⚠️ Major |
| **NuGet** | API key only | Trusted Publishing OR Code Signing | ⚠️ Moderate |
| **Maven Central** | GPG signing ✅ | GPG signing | ✅ Good |
| **GitHub Releases** | Basic GoReleaser | SBOM + Provenance + Signing | ⚠️ Major |

---

## 1. npm (Node.js SDK)

### Current Implementation
```yaml
- name: Publish Node.js SDK
  run: npm publish --access public
  env:
    NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

### Best Practice (2025)
npm now supports **Trusted Publishing** via OIDC (as of July 2025), which eliminates long-lived tokens entirely.

**Recommended approach:**
```yaml
- name: Publish Node.js SDK
  run: npm publish --access public --provenance
  permissions:
    id-token: write  # Required for OIDC
    contents: read
```

### Benefits of Upgrading
- **No more NPM_TOKEN secret** - Uses GitHub's OIDC identity
- **Automatic provenance** - Cryptographic proof of build origin
- **Sigstore integration** - Logged in public transparency ledger
- **Short-lived tokens** - Cannot be stolen/reused

### How to Upgrade
1. Go to npmjs.com → Package Settings → Publishing Access
2. Configure Trusted Publisher for your GitHub repo
3. Update workflow to use `--provenance` flag
4. Remove `NPM_TOKEN` secret

**Reference:** [npm Trusted Publishing Docs](https://docs.npmjs.com/trusted-publishers/)

---

## 2. PyPI (Python SDK)

### Current Implementation
```yaml
- name: Publish Python SDK
  run: twine upload dist/*
  env:
    TWINE_USERNAME: __token__
    TWINE_PASSWORD: ${{ secrets.PYPI_API_TOKEN }}
```

### Best Practice (2025)
PyPI has supported **Trusted Publishing** since 2023. This is now the recommended approach.

**Recommended approach:**
```yaml
- name: Publish Python SDK
  uses: pypa/gh-action-pypi-publish@release/v1
  permissions:
    id-token: write  # Required for OIDC
```

### Benefits of Upgrading
- **No more PYPI_API_TOKEN secret** - Uses GitHub's OIDC identity
- **Automatic attestations** - Digital signatures via Sigstore
- **Short-lived tokens** - Expire after use
- **Simpler workflow** - Single action, no twine setup

### How to Upgrade
1. Go to pypi.org → Your Project → Settings → Publishing
2. Add new "Trusted Publisher" for GitHub Actions
3. Replace twine upload with `pypa/gh-action-pypi-publish@release/v1`
4. Remove `PYPI_API_TOKEN` secret

**Reference:** [PyPI Trusted Publishers Docs](https://docs.pypi.org/trusted-publishers/)

---

## 3. NuGet (.NET SDK)

### Current Implementation
```yaml
- name: Publish .NET SDK
  run: dotnet nuget push *.nupkg --api-key ${{ secrets.NUGET_PUBLISH_KEY }} --source https://api.nuget.org/v3/index.json
```

### Best Practice (2025)
NuGet now supports **Trusted Publishing** (similar to npm/PyPI).

**Option A: Trusted Publishing (Recommended)**
```yaml
# Configure trust policy on nuget.org first
- name: Publish .NET SDK
  run: |
    # Exchange OIDC token for short-lived NuGet API key
    dotnet nuget push *.nupkg --source https://api.nuget.org/v3/index.json
  permissions:
    id-token: write
```

**Option B: Code Signing (Traditional)**
Requires X.509 certificate from a trusted CA. More complex but provides author signatures on packages.

### Current Gap
- No code signing on packages
- Using long-lived API key
- No provenance attestations

### How to Upgrade
1. Configure Trusted Publishing on nuget.org
2. Update workflow to use OIDC
3. Optionally add `dotnet nuget sign` for author signatures

**Reference:** [NuGet Trusted Publishing](https://andrewlock.net/easily-publishing-nuget-packages-from-github-actions-with-trusted-publishing/)

---

## 4. Maven Central (Java SDK)

### Current Implementation
```yaml
- name: Publish Java SDK
  run: gradle publishToSonatype closeAndReleaseSonatypeStagingRepository
  env:
    SIGNING_KEY_ID: ${{ secrets.JAVA_SIGNING_KEY_ID }}
    SIGNING_KEY: ${{ secrets.JAVA_SIGNING_KEY }}
    SIGNING_PASSWORD: ${{ secrets.JAVA_SIGNING_PASSWORD }}
```

### Assessment: ✅ Good
This follows best practices for Maven Central:
- **GPG signing** - Required by Maven Central, properly configured
- **Sonatype staging** - Proper release flow with staging → release
- **Key management** - Keys stored as secrets

### Minor Improvements Possible
- Consider using `cosign` for keyless signing (emerging practice)
- Add SBOM generation for Java artifacts

---

## 5. GitHub Releases (Provider Binaries)

### Current Implementation
```yaml
# .goreleaser.yml
builds:
  - binary: pulumi-resource-pulumi-webflow
    # No signing, no SBOM, no provenance

changelog:
  disable: true
```

### Best Practice (2025)
GoReleaser supports SBOM generation, Sigstore signing, and SLSA provenance.

**Recommended .goreleaser.yml additions:**
```yaml
# Add SBOM generation
sboms:
  - artifacts: archive
    documents:
      - "${artifact}.sbom.json"

# Add Sigstore signing (keyless)
signs:
  - cmd: cosign
    env:
      - COSIGN_EXPERIMENTAL=1
    signature: "${artifact}.sig"
    certificate: "${artifact}.pem"
    args:
      - sign-blob
      - "--output-signature=${signature}"
      - "--output-certificate=${certificate}"
      - "${artifact}"
    artifacts: checksum

# Enable changelog
changelog:
  disable: false
  use: github
```

**Add to workflow:**
```yaml
permissions:
  id-token: write      # For Sigstore
  contents: write
  attestations: write  # For GitHub attestations

- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v5
  # ...

- name: Generate SLSA Provenance
  uses: actions/attest-build-provenance@v2
  with:
    subject-checksums: ./dist/checksums.txt
```

### Benefits of Upgrading
- **SBOM** - Users can audit dependencies for vulnerabilities
- **Sigstore signing** - Cryptographic verification of binaries
- **SLSA provenance** - Verifiable build origin (SLSA Level 3)
- **Changelog** - Automatic release notes from commits

**Reference:** [GoReleaser Supply Chain Example](https://github.com/goreleaser/example-supply-chain)

---

## 6. Go SDK

### Current Implementation
```yaml
- name: Publish Go SDK
  uses: pulumi/publish-go-sdk-action@v1
```

### Assessment: ✅ Acceptable
The Pulumi action handles Go module publishing correctly. Go modules don't have a central signing mechanism like other registries.

---

## Recommended Upgrade Priority

### Phase 1: Quick Wins (Low effort, high impact)
1. **PyPI → Trusted Publishing** - Easiest to implement, removes a secret
2. **npm → Trusted Publishing + Provenance** - Similar ease, major security boost
3. **Enable GoReleaser changelog** - One line change

### Phase 2: Medium Effort
4. **NuGet → Trusted Publishing** - Newer feature, may need testing
5. **GoReleaser → SBOM generation** - Add syft integration
6. **GoReleaser → Sigstore signing** - Keyless signing for binaries

### Phase 3: Advanced (Optional)
7. **SLSA Provenance for binaries** - Requires slsa-github-generator
8. **NuGet code signing** - Requires purchasing X.509 certificate
9. **Full supply chain verification** - End-to-end attestation chain

---

## Example: Upgraded Workflow Snippet

Here's what an upgraded `publish_sdks` job could look like:

```yaml
publish_sdks:
  runs-on: ubuntu-latest
  needs: publish
  permissions:
    id-token: write    # Required for OIDC/Trusted Publishing
    contents: read
  steps:
    # ... download artifacts ...

    # npm with Trusted Publishing + Provenance
    - name: Publish Node.js SDK
      run: |
        cd sdk/nodejs
        npm publish --access public --provenance
      # No NPM_TOKEN needed!

    # PyPI with Trusted Publishing + Attestations
    - name: Publish Python SDK
      uses: pypa/gh-action-pypi-publish@release/v1
      with:
        packages-dir: sdk/python/bin/dist/
      # No PYPI_API_TOKEN needed!

    # NuGet (keep API key for now, or upgrade to Trusted Publishing)
    - name: Publish .NET SDK
      run: |
        cd sdk/dotnet
        dotnet nuget push *.nupkg --api-key ${{ secrets.NUGET_PUBLISH_KEY }} --source https://api.nuget.org/v3/index.json
```

---

## Sources

- [npm Trusted Publishing](https://docs.npmjs.com/trusted-publishers/)
- [npm Provenance](https://docs.npmjs.com/generating-provenance-statements/)
- [PyPI Trusted Publishers](https://docs.pypi.org/trusted-publishers/)
- [pypa/gh-action-pypi-publish](https://github.com/pypa/gh-action-pypi-publish)
- [NuGet Trusted Publishing](https://andrewlock.net/easily-publishing-nuget-packages-from-github-actions-with-trusted-publishing/)
- [NuGet Provenance Attestations](https://andrewlock.net/creating-provenance-attestations-for-nuget-packages-in-github-actions/)
- [GoReleaser Attestations](https://goreleaser.com/customization/attestations/)
- [GoReleaser Supply Chain Example](https://github.com/goreleaser/example-supply-chain)
- [GoReleaser SLSA Provenance](https://goreleaser.com/blog/slsa-generation-for-your-artifacts/)
