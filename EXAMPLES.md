# Example Guidelines for pulumi-webflow

This document defines the standards and requirements for examples in the pulumi-webflow provider.

## Overview

Every resource in this provider MUST have example code demonstrating its usage. Examples serve as:
- Documentation for users learning the provider
- Validation that resources work correctly
- Integration tests for CI/CD
- Reference implementations for common patterns

## Example Coverage Requirements

### Tier 1: Essential (REQUIRED)

**Every resource MUST have:**
- ✅ At least one working example in **TypeScript**
- ✅ A README.md explaining what the example does
- ✅ Integration test coverage (e.g., `examples/<resource>_test.go`)

**Minimum content:**
- Demonstrates all required properties
- Shows 2-3 common optional properties
- Includes helpful comments
- Exports meaningful outputs
- Uses placeholder values with clear naming (e.g., `your-site-id-here`)

### Tier 2: Multi-Language (RECOMMENDED)

**Core and frequently-used resources SHOULD have examples in:**
- ✅ TypeScript/JavaScript (REQUIRED)
- ✅ Python
- ✅ Go
- ✅ C# (if .NET SDK is supported)
- ✅ Java (if Java SDK is supported)

**Core resources include:**
- Site, Collection, CollectionItem, Page, Webhook, Redirect, RobotsTxt

### Tier 3: Integration Examples (NICE TO HAVE)

**Complex workflows SHOULD have integration examples showing:**
- Multiple resources working together
- Real-world use cases (e.g., "Setting up a complete Webflow site")
- Best practices (e.g., multi-site management, CI/CD patterns)
- Stack configuration patterns

### Tier 4: Advanced Patterns (OPTIONAL)

**Advanced scenarios MAY include:**
- Error handling and recovery
- Migration examples
- Performance optimization patterns
- Complex configuration scenarios

## Example Structure

### Directory Layout

Each resource example should follow this structure:

```
examples/
  <resource-name>/
    README.md                    # What this example does and how to run it
    typescript/
      index.ts
      package.json
      tsconfig.json
      Pulumi.yaml
    python/
      __main__.py
      requirements.txt
      Pulumi.yaml
    go/
      main.go
      go.mod
      Pulumi.yaml
    csharp/
      Program.cs
      <resource-name>.csproj
      Pulumi.yaml
    java/
      src/main/java/com/pulumi/webflow/examples/App.java
      pom.xml
      Pulumi.yaml
```

### README Template

Each example directory should have a README.md:

```markdown
# <Resource Name> Example

This example demonstrates how to use the `webflow.<ResourceName>` resource.

## What This Example Does

[Brief description of what the example creates/manages]

## Prerequisites

- Pulumi CLI installed
- Webflow API token set as `WEBFLOW_API_TOKEN` environment variable
- [Any other resource IDs or prerequisites]

## Running the Example

Choose your language:

### TypeScript
\`\`\`bash
cd typescript
npm install
pulumi up
\`\`\`

[Repeat for each language]

## Key Features Demonstrated

- [Feature 1]
- [Feature 2]
- [Feature 3]

## Outputs

- `<output-name>`: [Description]
```

### Code Quality Standards

All example code MUST:
- ✅ Be executable and tested
- ✅ Follow language-specific best practices
- ✅ Include helpful comments
- ✅ Use meaningful resource names
- ✅ Export useful outputs
- ✅ Handle configuration via Pulumi config when appropriate
- ✅ Use clear placeholder values (e.g., `your-site-id-here`, not `xxx`)

All example code SHOULD:
- ✅ Be as simple as possible while demonstrating the feature
- ✅ Avoid unnecessary complexity or dependencies
- ✅ Include error handling where appropriate
- ✅ Follow the patterns established in existing examples

## Testing Requirements

### Integration Tests

Each resource example MUST have a corresponding test in `examples/<resource>_test.go`:

```go
func TestAccRedirect(t *testing.T) {
    test := getBaseOptions(t).
        With(integration.ProgramTestOptions{
            Dir: filepath.Join(getCwd(t), "redirect", "typescript"),
        })
    integration.ProgramTest(t, &test)
}
```

Tests SHOULD cover:
- ✅ At least the TypeScript example
- ✅ Multiple languages for core resources
- ✅ Up and update operations
- ✅ Output validation

## Workflow: Adding a New Resource

When implementing a new resource, follow this checklist:

- [ ] Implement the resource in `provider/<resource>_resource.go`
- [ ] Run `make codegen` to generate SDKs
- [ ] Create `examples/<resource>/` directory
- [ ] Create TypeScript example (REQUIRED)
- [ ] Create README.md
- [ ] Create integration test in `examples/<resource>_test.go`
- [ ] Test the example: `cd examples/<resource>/typescript && pulumi up`
- [ ] If core resource: Add Python, Go, C#, Java examples
- [ ] Run `make test_examples` to verify tests pass
- [ ] Commit all generated SDK code along with examples

## Workflow: Updating an Existing Resource

When modifying a resource:

- [ ] Update affected examples to use new properties/behavior
- [ ] Run `make codegen`
- [ ] Test affected examples
- [ ] Update README if behavior changed
- [ ] Verify integration tests still pass

## Current Status

Track example coverage in this section:

### Complete Coverage (All 5 languages: TypeScript, Python, Go, C#, Java)
- ✅ Asset
- ✅ Collection
- ✅ CollectionItem
- ✅ Page
- ✅ Redirect
- ✅ RobotsTxt
- ✅ Site
- ✅ Webhook

### Multi-Language Coverage (TypeScript, Python, Go)
- ✅ Collection (TS, Python, Go, C#, Java)
- ✅ CollectionItem (TS, Python, Go, C#, Java)
- ✅ Page (TS, Python, Go)
- ✅ Webhook (TS, Python, Go)

### TypeScript-Only Coverage
- ✅ AssetFolder
- ✅ CollectionField
- ✅ PageContent
- ✅ PageCustomCode
- ✅ RegisteredScript
- ✅ SiteCustomCode

### TypeScript-Only Coverage (continued)
- ✅ User

### Missing Examples
- (None - all resources have examples!)

**Current Coverage: 100% (15/15 resources with at least TypeScript examples)**
**Multi-Language Coverage: 53% (8/15 resources with 3+ languages)**
**Complete Coverage: 53% (8/15 resources with all 5 languages)**

✅ **Target Met:** 100% of resources have at least TypeScript examples
✅ **Bonus:** All core resources have multi-language coverage

## Integration Examples

Current integration examples (these go beyond single resources):
- ✅ multi-site/ - Managing multiple Webflow sites
- ✅ stack-config/ - Configuration management patterns
- ✅ quickstart/ - Getting started guide
- ✅ troubleshooting-logs/ - Debugging patterns
- ✅ ci-cd/, git-workflows/ - Automation patterns

## Language-Specific Notes

### TypeScript
- Use modern async/await syntax
- Include proper type imports
- Use `@jdetmar/pulumi-webflow` package name

### Python
- Follow PEP 8 style guidelines
- Use `pulumi_webflow` package name
- Use snake_case for properties

### Go
- Follow Go conventions
- Use proper error handling
- Import from `github.com/jdetmar/pulumi-webflow/sdk/go/webflow`

### C#
- Follow .NET naming conventions (PascalCase)
- Use proper async/await patterns
- Include proper project file

### Java
- Follow Java conventions
- Use Maven for dependency management
- Include proper package structure

## Questions?

If you're unsure about example requirements:
1. Look at existing examples for the same resource type
2. Check `examples/redirect/`, `examples/robotstxt/`, or `examples/site/` as reference implementations
3. Ensure your example can be run with `pulumi up` after following the README
4. Ask: "Would a new user understand how to use this resource from this example?"
