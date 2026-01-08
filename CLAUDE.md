# Claude Code Instructions

This file provides guidance for Claude Code when working on this repository.

## Project Overview

This is a Pulumi native provider for Webflow, allowing users to manage Webflow resources (sites, redirects, robots.txt) as infrastructure-as-code.

## Guiding Principle

**Follow the Pulumi provider boilerplate patterns wherever possible.**

This project is based on the [Pulumi provider boilerplate](https://github.com/pulumi/pulumi-provider-boilerplate). When making changes to workflows, Makefile targets, project structure, or build processes, check how the boilerplate handles it first and maintain consistency. This ensures:

- Easier upgrades when the boilerplate improves
- Familiarity for contributors who know Pulumi providers
- Proven patterns that work with Pulumi's tooling

## Development Workflow

### Provider Code Changes

**IMPORTANT:** After modifying any Go code in `provider/`, you MUST run `make codegen` before committing.

```bash
# 1. Make changes to provider Go code
#    Edit files in provider/*.go

# 2. Regenerate schema and SDKs
make codegen

# 3. Commit everything together
git add .
git commit -m "your message"
```

**Why?** CI runs `make codegen` and checks if the working tree is clean ("Check worktree clean" step). If you forget to regenerate, CI will fail because the regenerated files differ from what you committed.

### What `make codegen` does

1. Builds the provider binary (`bin/pulumi-resource-webflow`)
2. Extracts `schema.json` from the provider binary
3. Generates SDK source files for all languages:
   - `sdk/go/` - Go SDK
   - `sdk/nodejs/` - TypeScript/JavaScript SDK
   - `sdk/python/` - Python SDK
   - `sdk/dotnet/` - .NET SDK
   - `sdk/java/` - Java SDK

### Key Make Targets

| Command | Description |
|---------|-------------|
| `make codegen` | Regenerate schema + all SDK source files (run after provider changes) |
| `make build` | Build provider + compile all SDKs |
| `make provider` | Build only the provider binary |
| `make test_provider` | Run provider unit tests |
| `make lint` | Run golangci-lint on provider code |

### Adding New Resources

**IMPORTANT:** When adding a new resource, you MUST also create examples. See [EXAMPLES.md](EXAMPLES.md) for complete requirements.

Required steps:
1. Implement resource in `provider/<resource>_resource.go`
2. Run `make codegen` to generate SDKs
3. Create at minimum: TypeScript example in `examples/<resource>/typescript/`
4. Create `examples/<resource>/README.md`
5. For core resources: Add Python, Go, C#, Java examples

**Minimum example coverage:** Every resource must have at least a TypeScript example with README.

**See:** [EXAMPLES.md](EXAMPLES.md) for detailed guidelines, templates, and current coverage status.

### Java SDK Build Process

The Java SDK requires post-processing after generation because `pulumi-java-gen` doesn't support all Maven Central requirements.

**What `pulumi-java-gen` supports via schema.json:**
- `basePackage` - Java package prefix (set to `io.github.jdetmar`)
- `buildFiles` - Generates Gradle build files

**What requires post-processing** (in `scripts/patch-java-build-gradle.py`):
- `groupId` / `artifactId` - Maven coordinates
- POM metadata: name, license, developers, SCM URLs
- GPG signing with 3-parameter `useInMemoryPgpKeys(keyId, key, password)`

The Makefile automatically runs the post-processing script after Java SDK generation.

**Maven Central coordinates:** `io.github.jdetmar:pulumi-webflow`

### Testing

```bash
# Run provider tests (uses mocked HTTP, no API token needed)
make test_provider
```

## Project Structure

```
provider/           # Go provider implementation
  ├── provider.go   # Main provider setup
  ├── *_resource.go # Resource implementations (redirect, robotstxt, site)
  └── cmd/          # Provider binary entry point + schema.json

sdk/                # Generated SDK code (DO NOT edit manually)
  ├── go/
  ├── nodejs/
  ├── python/
  ├── dotnet/
  └── java/

scripts/            # Build scripts
  └── patch-java-build-gradle.py  # Post-processes Java SDK for Maven Central

examples/           # Example Pulumi programs for each language
```

## CI/CD

- **build.yml**: Runs on pushes to main - builds provider, generates SDKs, runs tests
- **run-acceptance-tests.yml**: Runs on PRs - full test suite
- **release.yml**: Runs on tags - publishes to npm, PyPI, NuGet, Maven, GitHub Releases

## Environment

Tools are managed via [mise](https://mise.jdx.dev/). Key versions in `.config/mise.toml`:
- Go (latest)
- Node.js 20.x
- Python 3.11
- .NET 8.0
- Java 11 (Corretto)
- Gradle 7.6.6
