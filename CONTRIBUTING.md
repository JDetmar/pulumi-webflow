# Contributing to pulumi-webflow

Thank you for your interest in contributing to the Pulumi Webflow provider!

## Quick Start

1. **Fork and clone** the repository
2. **Install dependencies** - We use [mise](https://mise.jdx.dev/) for tool management:
   ```bash
   mise install
   ```
3. **Set up your Webflow API token** for testing:
   ```bash
   export WEBFLOW_API_TOKEN="your-token-here"
   ```
4. **Make your changes** following the guidelines below
5. **Test your changes**:
   ```bash
   make test_provider
   make test_examples
   ```
6. **Submit a pull request**

## Development Workflow

### Prerequisites

- Go 1.21+
- Node.js 20.x
- Python 3.11
- .NET 8.0 (for C# examples)
- Java 11 (for Java examples)

**Recommended:** Use [mise](https://mise.jdx.dev/) - it will install all required versions automatically.

### Making Changes to Provider Code

**CRITICAL:** After any change to Go code in `provider/`, you MUST run `make codegen`:

```bash
# 1. Edit provider code
vim provider/my_resource.go

# 2. Regenerate schema and SDKs (REQUIRED!)
make codegen

# 3. Commit everything together
git add .
git commit -m "feat: add new feature"
```

**Why?** CI validates that generated code is up-to-date. If you forget `make codegen`, CI will fail.

### Adding a New Resource

When adding a new resource, follow these steps:

1. **Implement the resource** in `provider/<resource>_resource.go`
   - Follow patterns from existing resources
   - Include proper schema definitions
   - Add CRUD operations

2. **Run codegen** to generate SDKs:
   ```bash
   make codegen
   ```

3. **Create examples** (see [EXAMPLES.md](EXAMPLES.md)):
   - **Minimum requirement:** TypeScript example + README + integration test
   - **Recommended for core resources:** Examples in all 5 languages

4. **Write tests**:
   - Unit tests in the provider code
   - Integration test in `examples/<resource>_test.go`

5. **Test everything**:
   ```bash
   make lint
   make test_provider
   make test_examples
   ```

6. **Commit and push**:
   ```bash
   git add .
   git commit -m "feat(resource): add NewResource"
   git push origin your-branch
   ```

### Example Requirements

**Every resource MUST have at least:**
- ✅ TypeScript example
- ✅ README explaining what it does
- ✅ Integration test

See [EXAMPLES.md](EXAMPLES.md) for complete guidelines, templates, and current coverage status.

## Project Structure

```
provider/           # Go provider implementation
  ├── provider.go   # Provider registration
  ├── *_resource.go # Resource implementations
  └── cmd/          # Provider binary + schema

sdk/                # Generated SDKs (DO NOT edit manually)
  ├── go/
  ├── nodejs/
  ├── python/
  ├── dotnet/
  └── java/

examples/           # Example programs + integration tests
  ├── <resource>/   # Per-resource examples
  │   ├── typescript/
  │   ├── python/
  │   ├── go/
  │   ├── csharp/
  │   ├── java/
  │   └── README.md
  └── *_test.go     # Integration tests
```

## Important Make Targets

| Command | Description |
|---------|-------------|
| `make codegen` | **Run after ANY provider code change** |
| `make build` | Build provider + all SDKs |
| `make test_provider` | Run Go unit tests |
| `make test_examples` | Run integration tests (needs WEBFLOW_API_TOKEN) |
| `make lint` | Run golangci-lint |

## Commit Message Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat(resource): add Webhook resource` - New feature
- `fix(redirect): handle null destination paths` - Bug fix
- `docs(examples): add Python example for Collection` - Documentation
- `chore: update dependencies` - Maintenance
- `test: add unit tests for Site resource` - Tests

Breaking changes should include `BREAKING CHANGE:` in the footer.

## Pull Request Process

1. **Ensure all tests pass** locally
2. **Update documentation** if needed
3. **Add examples** for new resources (see [EXAMPLES.md](EXAMPLES.md))
4. **Run `make codegen`** before committing provider changes
5. **Write a clear PR description**:
   - What does this change?
   - Why is it needed?
   - How was it tested?
6. **Link related issues** if applicable
7. **Ensure CI passes** - fix any failures before requesting review

## Code Review Standards

Reviewers will check for:
- ✅ All provider changes include `make codegen` output
- ✅ New resources have at least TypeScript examples
- ✅ Tests are passing
- ✅ Code follows Go best practices
- ✅ No breaking changes without version bump
- ✅ Documentation is clear and accurate
- ✅ Examples are runnable and tested

## Testing

### Unit Tests

Provider unit tests are in `provider/*_test.go`:

```bash
make test_provider
```

### Integration Tests

Integration tests run actual Pulumi programs against the Webflow API:

```bash
# Requires WEBFLOW_API_TOKEN
export WEBFLOW_API_TOKEN="your-token"
make test_examples
```

**Note:** Integration tests use real API calls and may create/modify/delete Webflow resources.

### Testing Specific Examples

```bash
cd examples/redirect/typescript
npm install
pulumi preview  # Preview changes
pulumi up       # Apply changes
pulumi destroy  # Clean up
```

## Guiding Principles

1. **Follow Pulumi provider boilerplate patterns**
   - This project is based on [pulumi/pulumi-provider-boilerplate](https://github.com/pulumi/pulumi-provider-boilerplate)
   - Check the boilerplate for guidance on structure and patterns

2. **Examples are not optional**
   - Every resource needs examples
   - Examples serve as documentation, tests, and validation
   - See [EXAMPLES.md](EXAMPLES.md) for requirements

3. **Generated code is part of the commit**
   - Always run `make codegen` after provider changes
   - Commit the generated SDK code with your changes
   - CI validates that codegen was run

4. **Test before submitting**
   - Run `make lint` to catch style issues
   - Run `make test_provider` for unit tests
   - Run `make test_examples` for integration tests (if you have API access)

## Resources

- **[CLAUDE.md](CLAUDE.md)** - Instructions for Claude Code (also useful for understanding the codebase)
- **[EXAMPLES.md](EXAMPLES.md)** - Detailed example guidelines and templates
- **[Pulumi Provider Boilerplate](https://github.com/pulumi/pulumi-provider-boilerplate)** - Upstream template
- **[Pulumi Provider Author Guide](https://www.pulumi.com/docs/guides/pulumi-packages/)** - Official Pulumi documentation

## Getting Help

- **Issues:** Open a [GitHub issue](https://github.com/jdetmar/pulumi-webflow/issues)
- **Discussions:** Start a [GitHub discussion](https://github.com/jdetmar/pulumi-webflow/discussions)
- **Documentation:** Check the [README.md](README.md)

## Code of Conduct

This project follows our [Code of Conduct](CODE-OF-CONDUCT.md). Please be respectful and constructive.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see [LICENSE](LICENSE)).

---

**Thank you for contributing to pulumi-webflow!** 🎉
