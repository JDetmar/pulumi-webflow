# Copilot Instructions for Pulumi Webflow Provider Development

This guide provides comprehensive instructions for developing and maintaining the Pulumi Webflow provider. It serves as the canonical reference for contributors, maintainers, and AI assistants working on this open-source community project.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Guiding Principles](#guiding-principles)
3. [Architecture & Design](#architecture--design)
4. [Development Workflow](#development-workflow)
5. [Code Standards & Best Practices](#code-standards--best-practices)
6. [Testing Guidelines](#testing-guidelines)
7. [Documentation Requirements](#documentation-requirements)
8. [Community & Contribution Guidelines](#community--contribution-guidelines)
9. [CI/CD & Release Process](#cicd--release-process)
10. [Troubleshooting & Support](#troubleshooting--support)

---

## Project Overview

**Pulumi Webflow Provider** is an unofficial, community-maintained Pulumi native provider that enables infrastructure-as-code management of Webflow resources (sites, redirects, robots.txt, collections, webhooks, and more).

### Key Information

- **Repository**: https://github.com/jdetmar/pulumi-webflow
- **Provider Type**: Pulumi Native Provider (using pulumi-go-provider SDK)
- **Language**: Go (provider), with generated SDKs for TypeScript, Python, Go, .NET, Java
- **License**: MIT
- **Status**: Community-maintained, not officially affiliated with Pulumi Corporation or Webflow, Inc.

### Architecture Foundation

This provider is built on the [Pulumi provider boilerplate](https://github.com/pulumi/pulumi-provider-boilerplate), which provides proven patterns for provider development, CI/CD, and multi-language SDK generation.

---

## Guiding Principles

### 1. Follow the Pulumi Provider Boilerplate

**Primary Rule**: When making changes to workflows, Makefile targets, project structure, or build processes, always check how the [Pulumi provider boilerplate](https://github.com/pulumi/pulumi-provider-boilerplate) handles it first and maintain consistency.

**Why?**
- Easier upgrades when the boilerplate improves
- Familiarity for contributors who know Pulumi providers
- Proven patterns that work with Pulumi's tooling ecosystem
- Consistent developer experience across Pulumi providers

### 2. Reference Official Pulumi Documentation

Always refer to official Pulumi documentation when implementing provider features:

- [Provider Architecture](https://www.pulumi.com/docs/iac/guides/building-extending/providers/provider-architecture/) - Understanding provider design patterns
- [Build a Provider](https://www.pulumi.com/docs/iac/guides/building-extending/providers/build-a-provider/) - Step-by-step provider development guide
- [Pulumi Go Provider SDK](https://www.pulumi.com/docs/iac/guides/building-extending/providers/sdks/pulumi-go-provider-sdk/) - SDK-specific implementation details

### 3. Minimal, Surgical Changes

- Make the smallest possible changes to achieve your goal
- Avoid refactoring unrelated code unless absolutely necessary
- Keep pull requests focused on a single concern
- Document the rationale for design decisions

### 4. Community-First Mindset

- Prioritize transparency and clear communication
- Write documentation that helps others succeed
- Be welcoming and supportive in code reviews
- Follow the [Code of Conduct](./CODE-OF-CONDUCT.md) in all interactions

---

## Architecture & Design

### Provider Structure

```
provider/           # Go provider implementation
  ├── provider.go   # Main provider setup and registration
  ├── config.go     # Provider configuration (API token, etc.)
  ├── auth.go       # Authentication and HTTP client setup
  ├── *_resource.go # Resource implementations (one per resource type)
  ├── *_test.go     # Unit tests for each resource
  └── cmd/          # Provider binary entry point + schema.json

sdk/                # Generated SDK code (DO NOT edit manually)
  ├── go/           # Go SDK (auto-generated)
  ├── nodejs/       # TypeScript/JavaScript SDK (auto-generated)
  ├── python/       # Python SDK (auto-generated)
  ├── dotnet/       # .NET SDK (auto-generated)
  └── java/         # Java SDK (auto-generated)

examples/           # Example Pulumi programs demonstrating provider usage
  ├── quickstart/   # Getting started examples for each language
  ├── redirect/     # Redirect resource examples
  ├── robotstxt/    # RobotsTxt resource examples
  ├── site/         # Site resource examples
  └── ...           # Additional pattern examples

docs/               # Documentation and guides
tests/              # Integration and acceptance tests
scripts/            # Build and utility scripts
```

### Key Architecture Patterns

#### Resource Implementation Pattern

Each resource follows this structure (example: `redirect_resource.go`):

```go
// 1. Input/Output types define the resource schema
type RedirectArgs struct {
    SiteId      string `pulumi:"siteId"`
    SourcePath  string `pulumi:"sourcePath"`
    DestPath    string `pulumi:"destPath"`
    StatusCode  int    `pulumi:"statusCode,optional"`
}

type RedirectState struct {
    RedirectArgs
    RedirectId string `pulumi:"redirectId"`
}

// 2. Resource struct implements Create/Read/Update/Delete
type Redirect struct{}

func (r *Redirect) Create(ctx p.Context, name string, input RedirectArgs, 
    preview bool) (string, RedirectState, error) {
    // Implementation
}

func (r *Redirect) Read(ctx p.Context, id string, inputs RedirectArgs, 
    state RedirectState) (canonicalID string, normalizedInputs RedirectArgs, 
    normalizedState RedirectState, err error) {
    // Implementation
}

// ... Update, Delete methods
```

#### Authentication Pattern

All API requests use a shared HTTP client configured in `auth.go`:

```go
// GetAuthenticatedClient returns an HTTP client with proper headers
func GetAuthenticatedClient(ctx context.Context, apiToken string) *http.Client {
    // Sets Authorization header, User-Agent, etc.
}
```

#### Schema Generation

The provider schema is automatically extracted from Go types using the `infer` package:

```go
// In provider.go
prov, err := infer.NewProviderBuilder().
    WithConfig(infer.Config(&Config{})).
    WithResources(
        infer.Resource(&Redirect{}),
        infer.Resource(&RobotsTxt{}),
        // ...
    ).
    Build()
```

---

## Development Workflow

### Critical Rule: Always Run `make codegen` After Provider Changes

**IMPORTANT**: After modifying any Go code in `provider/`, you **MUST** run `make codegen` before committing.

```bash
# 1. Make changes to provider Go code
#    Edit files in provider/*.go

# 2. Regenerate schema and SDKs
make codegen

# 3. Verify changes compile
make build

# 4. Run tests
make test_provider

# 5. Commit everything together
git add .
git commit -m "feat: add new resource"
git push
```

**Why?** CI runs `make codegen` and checks if the working tree is clean ("Check worktree clean" step). If you forget to regenerate, CI will fail because the regenerated files differ from what you committed.

### What `make codegen` Does

1. Builds the provider binary (`bin/pulumi-resource-webflow`)
2. Runs the provider with `--get-schema` flag to extract `schema.json`
3. Uses Pulumi's schema tools to generate SDK source files:
   - `sdk/go/webflow/` - Go SDK
   - `sdk/nodejs/` - TypeScript/JavaScript SDK  
   - `sdk/python/pulumi_webflow/` - Python SDK
   - `sdk/dotnet/Pulumi.Webflow/` - .NET SDK
   - `sdk/java/` - Java SDK

### Key Make Targets

| Command | Description | When to Use |
|---------|-------------|-------------|
| `make codegen` | Regenerate schema + all SDK source files | **After every provider code change** |
| `make build` | Build provider + compile all SDKs | Before testing, to ensure everything compiles |
| `make provider` | Build only the provider binary | Quick iteration on provider code |
| `make test_provider` | Run provider unit tests | After implementing features |
| `make test_examples` | Run example integration tests | Before merging, requires WEBFLOW_API_TOKEN |
| `make lint` | Run golangci-lint on provider code | Before committing |
| `make clean` | Remove build artifacts | When you want a fresh start |

### Development Environment Setup

This project uses [mise](https://mise.jdx.dev/) to manage tool versions. Key tools defined in `.mise.toml`:

```toml
[tools]
golangci-lint = "1.64.8"
```

Additional required tools (install via mise or manually):

- **Go**: Latest version (check `go.mod` for minimum version)
- **Node.js**: 20.x (for TypeScript SDK)
- **Python**: 3.11+ (for Python SDK)
- **.NET**: 8.0+ (for .NET SDK)
- **Java**: 11+ (Corretto) (for Java SDK)
- **Gradle**: 7.6+ (for Java SDK builds)
- **Pulumi CLI**: 3.0+ (for testing examples)

Install mise and tools:

```bash
# Install mise (macOS/Linux)
curl https://mise.run | sh

# Install tools defined in .mise.toml
mise install

# Or install tools manually via your package manager
```

### Adding a New Resource

1. **Define the resource types** in a new file `provider/myresource.go`:
   ```go
   package provider
   
   type MyResourceArgs struct {
       SiteId string `pulumi:"siteId"`
       Name   string `pulumi:"name"`
   }
   
   type MyResourceState struct {
       MyResourceArgs
       ResourceId string `pulumi:"resourceId"`
   }
   
   type MyResource struct{}
   ```

2. **Implement CRUD methods**:
   ```go
   func (r *MyResource) Create(ctx p.Context, name string, input MyResourceArgs, 
       preview bool) (string, MyResourceState, error) {
       // Call Webflow API to create resource
       // Return ID and state
   }
   
   // Implement Read, Update, Delete
   ```

3. **Register the resource** in `provider/provider.go`:
   ```go
   WithResources(
       // ...existing resources...
       infer.Resource(&MyResource{}),
   )
   ```

4. **Add unit tests** in `provider/myresource_test.go`

5. **Run codegen**:
   ```bash
   make codegen
   ```

6. **Create examples** in `examples/myresource/`

7. **Document the resource** in `docs/`

### Making Changes to Existing Resources

1. **Modify the resource** in `provider/myresource_resource.go`

2. **Update unit tests** if behavior changed

3. **Run codegen** to regenerate SDKs:
   ```bash
   make codegen
   ```

4. **Update examples** if the API changed

5. **Test your changes**:
   ```bash
   make test_provider
   make test_examples  # Requires WEBFLOW_API_TOKEN
   ```

### Working with the Webflow API

- Use the shared HTTP client from `auth.go`
- Follow existing patterns in `redirect_resource.go`, `robotstxt_resource.go`, etc.
- Webflow API documentation: https://developers.webflow.com/
- Always handle API errors gracefully and provide helpful error messages
- Respect rate limits (implement retries with exponential backoff if needed)

---

## Code Standards & Best Practices

### Go Code Style

Follow standard Go conventions and the project's golangci-lint configuration (`.golangci.yml`):

```bash
# Run linter before committing
make lint

# Linter checks include:
# - errcheck: Check for unchecked errors
# - gosec: Security issues
# - govet: Suspicious constructs
# - gofumpt: Stricter formatting than gofmt
# - ineffassign: Ineffectual assignments
# - staticcheck: Static analysis
```

### Code Organization

- **One resource per file**: `redirect_resource.go`, `robotstxt_resource.go`
- **Separate test files**: `redirect_test.go`, `robotstxt_test.go`
- **Shared utilities in separate files**: `auth.go`, `config.go`
- **Keep files focused**: Each file should have a single responsibility

### Naming Conventions

- **Resources**: PascalCase, singular (e.g., `Redirect`, `RobotsTxt`, not `Redirects`)
- **Functions**: camelCase for private, PascalCase for public
- **Variables**: camelCase
- **Constants**: PascalCase or ALL_CAPS depending on context
- **Pulumi struct tags**: Use camelCase (e.g., `pulumi:"siteId"`)

### Error Handling

```go
// ✅ Good: Descriptive error messages with context
if resp.StatusCode == 404 {
    return "", state, fmt.Errorf("redirect with ID %s not found in site %s", 
        state.RedirectId, state.SiteId)
}

// ✅ Good: Wrap errors with context
if err := json.Unmarshal(body, &response); err != nil {
    return "", state, fmt.Errorf("failed to parse Webflow API response: %w", err)
}

// ❌ Bad: Generic error messages
if err != nil {
    return "", state, err
}
```

### Documentation Comments

Follow Go documentation conventions:

```go
// Redirect manages URL redirects in a Webflow site.
//
// This resource allows you to create, update, and delete 301/302 redirects,
// which are useful for SEO, site migrations, and URL restructuring.
//
// Example:
//
//   redirect := &webflow.Redirect{
//       SiteId: "abc123...",
//       SourcePath: "/old-page",
//       DestPath: "/new-page",
//       StatusCode: 301,
//   }
type Redirect struct{}
```

### Security Best Practices

- **Never log sensitive data**: API tokens, credentials, or user data
- **Validate all inputs**: Check for malformed site IDs, invalid paths, etc.
- **Use HTTPS exclusively**: Never allow HTTP API calls
- **Sanitize user inputs**: Prevent injection attacks in URL paths
- **Handle secrets properly**: Use Pulumi's secret handling for API tokens

```go
// ✅ Good: Validate inputs
func (r *Redirect) Create(ctx p.Context, name string, input RedirectArgs, 
    preview bool) (string, RedirectState, error) {
    
    if err := validateSiteId(input.SiteId); err != nil {
        return "", RedirectState{}, err
    }
    
    // ...
}

// ❌ Bad: No input validation
func (r *Redirect) Create(ctx p.Context, name string, input RedirectArgs, 
    preview bool) (string, RedirectState, error) {
    // Directly use input.SiteId without validation
}
```

---

## Testing Guidelines

### Unit Tests

Every resource should have comprehensive unit tests:

```bash
# Run all provider unit tests
make test_provider

# Run specific test
go test ./provider -run TestRedirectCreate
```

**Test Structure** (example: `redirect_test.go`):

```go
func TestRedirectCreate(t *testing.T) {
    // Arrange
    input := RedirectArgs{
        SiteId:     "test-site-id",
        SourcePath: "/old",
        DestPath:   "/new",
        StatusCode: 301,
    }
    
    // Act
    id, state, err := redirect.Create(ctx, "test-redirect", input, false)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, id)
    assert.Equal(t, input.SourcePath, state.SourcePath)
}
```

### Integration Tests

Integration tests require a real Webflow API token:

```bash
# Set your API token
export WEBFLOW_API_TOKEN="your-token-here"

# Run integration tests (slower, hits real API)
make test_examples
```

### Test Coverage Goals

- **Unit tests**: 80%+ coverage for provider code
- **Integration tests**: At least one example per resource
- **Error cases**: Test error handling paths
- **Edge cases**: Test boundary conditions (empty strings, max lengths, etc.)

### Testing Best Practices

- Use table-driven tests for multiple scenarios
- Mock external dependencies when possible
- Test both success and failure cases
- Use descriptive test names: `TestRedirectCreate_InvalidSiteId_ReturnsError`
- Clean up resources in tests (if hitting real API)

---

## Documentation Requirements

### Code Documentation

- **Public functions**: Must have doc comments
- **Resources**: Document purpose, usage, and examples
- **Complex logic**: Inline comments explaining the "why"
- **Examples**: Include example usage in doc comments

### User-Facing Documentation

When adding or modifying resources:

1. **README.md**: Update if user-facing changes
2. **docs/**: Create detailed resource documentation
3. **examples/**: Provide working examples in multiple languages
4. **Changelog**: Document changes (automated via release process)

### Example Structure

Each resource should have examples in `examples/`:

```
examples/
  myresource/
    typescript/
      index.ts
      package.json
      Pulumi.yaml
      README.md
    python/
      __main__.py
      requirements.txt
      Pulumi.yaml
      README.md
    go/
      main.go
      go.mod
      Pulumi.yaml
      README.md
```

### Writing Good Examples

```typescript
// ✅ Good: Self-contained, well-documented example
import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Get configuration
const config = new pulumi.Config();
const siteId = config.requireSecret("siteId");

// Create a redirect from old blog posts to new location
const blogRedirect = new webflow.Redirect("blog-redirect", {
    siteId: siteId,
    sourcePath: "/blog/*",        // Wildcard redirect
    destPath: "/posts/$1",         // Capture and reuse path segment
    statusCode: 301,               // Permanent redirect
});

// Export the redirect ID for reference
export const redirectId = blogRedirect.redirectId;
```

---

## Community & Contribution Guidelines

### Open Source Etiquette

This is a community project. We value:

- **Transparency**: Communicate openly about decisions and changes
- **Respect**: Treat all contributors with kindness and professionalism
- **Collaboration**: Work together to find the best solutions
- **Inclusivity**: Welcome contributors of all skill levels and backgrounds

### Code of Conduct

All contributors must follow the [Code of Conduct](./CODE-OF-CONDUCT.md). Key principles:

- Use welcoming and inclusive language
- Be respectful of differing viewpoints
- Accept constructive criticism gracefully
- Focus on what's best for the community
- Show empathy towards others

### Contribution Workflow

1. **Find or create an issue**: Check [GitHub Issues](https://github.com/jdetmar/pulumi-webflow/issues)

2. **Fork and create a branch**:
   ```bash
   git checkout -b feat/my-new-feature
   ```

3. **Make your changes**: Follow guidelines in this document

4. **Run checks**:
   ```bash
   make lint
   make test_provider
   make codegen  # If you changed provider code
   ```

5. **Commit with clear messages**:
   ```bash
   git commit -m "feat(redirect): add wildcard redirect support
   
   - Implement pattern matching for source paths
   - Add tests for wildcard scenarios
   - Update documentation and examples
   
   Closes #123"
   ```

6. **Push and create a Pull Request**:
   ```bash
   git push origin feat/my-new-feature
   # Then create PR on GitHub
   ```

### Commit Message Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

Examples:
```bash
feat(redirect): add support for wildcard redirects
fix(robotstxt): handle empty content correctly
docs: update quick start guide
test(site): add integration tests for site creation
```

### Code Review Process

- **Be constructive**: Provide specific, actionable feedback
- **Be responsive**: Address review comments promptly
- **Be open-minded**: Consider alternative approaches
- **Be thorough**: Test changes locally before approving

Reviewers should check:
- Code follows project conventions
- Tests are included and passing
- Documentation is updated
- Commit messages are clear
- No security issues introduced

### Getting Help

- **Documentation**: Check [docs/](./docs/) first
- **Examples**: Look at [examples/](./examples/)
- **Issues**: Search [existing issues](https://github.com/jdetmar/pulumi-webflow/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/jdetmar/pulumi-webflow/discussions)
- **Slack**: Join [Pulumi Community Slack](https://pulumi-community.slack.com/) (unofficial)

---

## CI/CD & Release Process

### Continuous Integration Workflows

The project uses GitHub Actions for CI/CD (`.github/workflows/`):

#### 1. Build Workflow (`build.yml`)

Runs on every push to `main`:

```yaml
# Triggered by: Push to main branch
# Steps:
- Checkout code
- Setup Go, Node.js, Python, .NET, Java
- Install dependencies
- Run linter (golangci-lint)
- Build provider binary
- Run `make codegen`
- Check worktree clean (ensures codegen was run)
- Build all SDKs
- Run provider unit tests
```

#### 2. Pull Request Workflow (`run-acceptance-tests.yml`)

Runs on every PR:

```yaml
# Triggered by: Pull request opened/updated
# Steps:
- All steps from build.yml
- Run integration tests (requires secrets)
- Run example tests
- Report test results
```

#### 3. Release Workflow (`release.yml`)

Runs when a version tag is pushed (e.g., `v1.2.3`):

```yaml
# Triggered by: Tag push (v*)
# Steps:
- Build provider for all platforms (Linux, macOS, Windows)
- Build and publish SDKs to:
  - npm (TypeScript/JavaScript)
  - PyPI (Python)
  - NuGet (.NET)
  - Maven Central (Java)
- Create GitHub Release with binaries
- Update provider in Pulumi Registry (if applicable)
```

### Release Process

1. **Update version** in relevant files (typically automated)

2. **Create and push tag**:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

3. **Wait for CI**: Release workflow automatically publishes

4. **Verify release**: Check GitHub Releases and package registries

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **Major** (v1.0.0 → v2.0.0): Breaking changes
- **Minor** (v1.0.0 → v1.1.0): New features, backward-compatible
- **Patch** (v1.0.0 → v1.0.1): Bug fixes, backward-compatible

Development versions: `v1.0.0-alpha.0+dev`

### CI Best Practices

- **Keep workflows fast**: Use caching, parallelize when possible
- **Fail fast**: Lint and test early in the pipeline
- **Provide clear errors**: Make it easy to diagnose failures
- **Secure secrets**: Use GitHub Secrets for tokens, never commit credentials
- **Test before merging**: Require CI to pass before merge

---

## Troubleshooting & Support

### Common Development Issues

#### Issue: CI fails with "worktree not clean"

**Cause**: You modified provider code but didn't run `make codegen` before committing.

**Solution**:
```bash
make codegen
git add .
git commit -m "chore: regenerate SDKs"
git push
```

#### Issue: golangci-lint fails

**Cause**: Code doesn't meet linting standards defined in `.golangci.yml`.

**Solution**:
```bash
# Run locally to see errors
make lint

# Fix issues manually or use auto-fix
golangci-lint run --fix

# Commit fixes
git add .
git commit -m "chore: fix linter issues"
```

#### Issue: Tests fail with "unauthorized"

**Cause**: Missing or invalid Webflow API token.

**Solution**:
```bash
# Set your token
export WEBFLOW_API_TOKEN="your-actual-token"

# Verify it works
curl -H "Authorization: Bearer $WEBFLOW_API_TOKEN" \
  https://api.webflow.com/v2/sites

# Re-run tests
make test_provider
```

#### Issue: Build fails after updating dependencies

**Cause**: Go module dependencies are out of sync.

**Solution**:
```bash
# Update dependencies
go mod tidy
go mod download

# Rebuild
make clean
make build
```

### Getting Support

#### For Users

- **Documentation**: https://github.com/jdetmar/pulumi-webflow/tree/main/docs
- **Examples**: https://github.com/jdetmar/pulumi-webflow/tree/main/examples
- **Issues**: https://github.com/jdetmar/pulumi-webflow/issues
- **Discussions**: https://github.com/jdetmar/pulumi-webflow/discussions

#### For Contributors

- **Contributing Guide**: [CONTRIBUTING.md](./CONTRIBUTING.md) (if it exists)
- **Code of Conduct**: [CODE-OF-CONDUCT.md](./CODE-OF-CONDUCT.md)
- **Developer Guide**: This document (copilot-instructions.md)
- **Pulumi Docs**: https://www.pulumi.com/docs/

### Debugging Tips

#### Enable verbose logging

```bash
# Provider-level logging
PULUMI_DEBUG_COMMANDS=true pulumi up -v=9

# HTTP request/response logging (add to provider code)
import "github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
logging.V(9).Infof("API Request: %s %s", method, url)
```

#### Test individual resources

```bash
# Run specific test
go test ./provider -run TestRedirectCreate -v

# Run with coverage
go test ./provider -cover -v
```

#### Inspect schema

```bash
# Build provider
make provider

# Extract schema
./bin/pulumi-resource-webflow --get-schema | jq . > schema.json

# Inspect schema
cat schema.json | jq '.resources["webflow:index/redirect:Redirect"]'
```

---

## Quick Reference

### Essential Commands

```bash
# After modifying provider code
make codegen

# Full build
make build

# Run tests
make test_provider
make test_examples

# Lint code
make lint

# Clean build artifacts
make clean
```

### Essential Files

- `provider/provider.go` - Provider setup and resource registration
- `provider/*_resource.go` - Individual resource implementations
- `Makefile` - Build targets and automation
- `.golangci.yml` - Linting configuration
- `.github/workflows/` - CI/CD pipelines
- `schema.json` - Auto-generated provider schema
- `CLAUDE.md` - Additional Claude-specific instructions

### Essential Links

- **Pulumi Provider Boilerplate**: https://github.com/pulumi/pulumi-provider-boilerplate
- **Provider Architecture**: https://www.pulumi.com/docs/iac/guides/building-extending/providers/provider-architecture/
- **Build a Provider**: https://www.pulumi.com/docs/iac/guides/building-extending/providers/build-a-provider/
- **Go Provider SDK**: https://www.pulumi.com/docs/iac/guides/building-extending/providers/sdks/pulumi-go-provider-sdk/
- **Webflow API**: https://developers.webflow.com/
- **Project Repository**: https://github.com/jdetmar/pulumi-webflow

---

## Contributing to This Guide

This copilot-instructions.md file itself is a living document. If you notice:

- Missing information
- Outdated instructions
- Unclear explanations
- Opportunities for improvement

Please open a PR to update this guide! Clear, accurate documentation benefits everyone in the community.

---

**Happy coding! 🚀**

Remember: We're building this together as a community. Your contributions, questions, and feedback make this provider better for everyone.
