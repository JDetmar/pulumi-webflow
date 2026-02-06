# Changelog

All notable changes to the Pulumi Webflow provider will be documented in this file.

## [Unreleased]

## [v0.9.3] - 2026-02-06

### Bug Fixes

- fix(site): preserve TimeZone in Update preview to prevent forced replace
- fix: defer custom code validation to support unknown inputs during preview

## [v0.9.2] - 2026-02-04

### Features

- feat: add InlineScript resource for registering inline custom code scripts

### Bug Fixes

- fix(site): deprecate shortName input and fix PATCH field name
- fix: add parentFolderId support to PatchSite and lint fixes
- fix(site): make timezone a read-only output field

### Breaking Changes

- feat!: remove User Accounts resource (Webflow API deprecated)

## [v0.9.1] - 2026-01-14

### Bug Fixes

- fix(provider): add pluginDownloadURL for automatic provider installation
- fix(examples): correct package references for C#, Go, and Java

## [v0.9.0] - 2026-01-14

### Features

- feat: add rate limit handling, security policy, and performance docs
- feat(devcontainer): improve dev environment setup

### Bug Fixes

- fix(invoke): prevent crash in getTokenInfo/getAuthorizedUser functions
- fix(asset): parse variants as array instead of map
- fix(registeredscript): resolve version diff detection issue
- fix(registeredscript): all changes now trigger replacement instead of update
- fix: exclude unchanged slug from CollectionItem PATCH to prevent duplicate slug error
- fix: resolve drift detection issues and asset creation
- fix: add collectionId output and fix provider issues
- fix: release pipeline and npm publishing

[Unreleased]: https://github.com/JDetmar/pulumi-webflow/compare/v0.9.3...HEAD
[v0.9.3]: https://github.com/JDetmar/pulumi-webflow/compare/v0.9.2...v0.9.3
[v0.9.2]: https://github.com/JDetmar/pulumi-webflow/compare/v0.9.1...v0.9.2
[v0.9.1]: https://github.com/JDetmar/pulumi-webflow/compare/v0.9.0...v0.9.1
[v0.9.0]: https://github.com/JDetmar/pulumi-webflow/releases/tag/v0.9.0
