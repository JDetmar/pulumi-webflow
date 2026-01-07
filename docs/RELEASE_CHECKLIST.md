# Release Checklist

Quick reference for releasing new versions. See [RELEASE_GUIDE.md](RELEASE_GUIDE.md) for detailed instructions.

## One-Time Setup Status

Check off when configured:

- [ ] **npm**: Trusted Publisher configured on npmjs.com (no secret needed!)
- [ ] **PyPI**: Trusted Publisher configured on pypi.org (no secret needed!)
- [ ] **NuGet**: Trusted Publisher configured + `NUGET_USERNAME` secret added
- [ ] **Maven Central**: All Java secrets configured
  - [ ] `OSSRH_USERNAME`
  - [ ] `OSSRH_PASSWORD`
  - [ ] `JAVA_SIGNING_KEY_ID`
  - [ ] `JAVA_SIGNING_KEY`
  - [ ] `JAVA_SIGNING_PASSWORD`
- [ ] **GitHub Actions**: Write permissions enabled

---

## Release Checklist

### Before Release

```bash
# 1. Ensure on main and up to date
git checkout main && git pull origin main

# 2. Verify codegen is current
make codegen && git status  # Should show no changes

# 3. Run tests
make test_provider

# 4. Run lint
make lint
```

- [ ] CI green on main branch
- [ ] Version number decided: `v____.____.____`
- [ ] Commits use conventional prefixes (`feat:`, `fix:`, `docs:`) for auto-changelog

### Create Release

```bash
# Create and push tag
git tag -a vX.Y.Z -m "Release vX.Y.Z: Brief description"
git push origin vX.Y.Z
```

### Monitor & Verify

- [ ] Watch workflow: https://github.com/JDetmar/pulumi-webflow/actions
- [ ] GitHub Release created with:
  - [ ] Binaries for all platforms
  - [ ] SBOM files (`.sbom.json`)
  - [ ] Auto-generated changelog
- [ ] npm package visible (with provenance badge)
- [ ] PyPI package visible (with Sigstore attestation)
- [ ] NuGet package visible
- [ ] Maven package visible (may take 1-2 hours)

### Test Installation

```bash
npm install @jdetmar/pulumi-webflow@X.Y.Z
pip install pulumi-webflow==X.Y.Z
dotnet add package Pulumi.Webflow --version X.Y.Z
```

---

## Emergency: Redo Release

If release failed or needs correction:

```bash
# Delete tag locally and remotely
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z

# Delete GitHub Release manually if created
# Then re-run the release process
```

---

## Version Guidelines

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| Bug fix | PATCH | 0.1.0 → 0.1.1 |
| New feature | MINOR | 0.1.1 → 0.2.0 |
| Breaking change | MAJOR | 0.2.0 → 1.0.0 |
| Pre-release | suffix | 0.1.0-alpha.1 |
