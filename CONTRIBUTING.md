# Contributing to Webflow Pulumi Provider

Thank you for your interest in contributing to the Webflow Pulumi Provider! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

This project follows the Pulumi Community Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the [existing issues](https://github.com/jdetmar/pulumi-webflow/issues) to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs actual behavior
- **Environment details**: OS, Go version, Pulumi version
- **Code samples** or error messages if applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Clear use case** explaining why this enhancement would be useful
- **Detailed description** of the proposed functionality
- **Examples** of how the feature would be used

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following the code style guidelines below
3. **Add tests** for any new functionality
4. **Update documentation** as needed
5. **Ensure all tests pass** by running `go test ./...`
6. **Submit a pull request** with a clear description of changes

## Development Setup

### Prerequisites

- [Go](https://golang.org/dl/) 1.21 or later
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/)
- A Webflow account with API access

### Building from Source

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/pulumi-webflow
cd pulumi-webflow

# Install dependencies
go mod download

# Build the provider
go build -o pulumi-resource-webflow

# Run tests
go test ./...
```

### Testing Your Changes

> **Note:** Test coverage is actively being expanded as resources are implemented. Current test files include provider initialization and core functionality tests.

```bash
# Run unit tests
go test ./...

# Run integration tests (requires WEBFLOW_API_TOKEN - coming soon)
export WEBFLOW_API_TOKEN="your-token-here"
go test -tags=integration ./...

# Install locally for testing with Pulumi
cp pulumi-resource-webflow ~/.pulumi/plugins/resource-webflow-v0.1.0/
```

## Code Style Guidelines

### Go Code Standards

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format all Go code
- Run `go vet` and address all warnings
- Use meaningful variable and function names
- Document all exported functions, types, and constants

### Documentation

- Document all public APIs with clear, concise comments
- Include code examples in documentation where helpful
- Update README.md if adding new features or changing installation

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb in present tense (e.g., "Add", "Fix", "Update")
- Reference issue numbers when applicable (e.g., "Fixes #123")

Example:
```
Add support for custom domain configuration

- Implement CustomDomain resource
- Add CRUD operations for domain management
- Include integration tests

Fixes #45
```

## Provider Development Guidelines

### Adding New Resources

When adding a new Webflow resource:

1. **Define the schema** in `provider/schema.go`
2. **Implement CRUD operations** (Create, Read, Update, Delete)
3. **Add state management** for drift detection
4. **Write comprehensive tests** including:
   - Unit tests for validation logic
   - Integration tests with Webflow API
5. **Update documentation** with examples
6. **Add SDK examples** in TypeScript, Python, and Go

### Error Handling

- Return meaningful error messages
- Use `fmt.Errorf` with context for wrapped errors
- Never panic in production code
- Validate inputs early and provide clear feedback

### Testing Requirements

- All new code must include tests
- Maintain or improve code coverage (target: >70%)
- Integration tests should be tagged with `// +build integration`
- Mock external dependencies in unit tests

### Security Testing

When working with authentication or credential handling:

- **Never log tokens or secrets** - Use redaction functions like `RedactToken()`
- **Test token redaction** - Verify tokens never appear in error messages or logs
- **Test TLS configuration** - Ensure HTTPS is enforced (minimum TLS 1.2)
- **Test input validation** - Verify empty/invalid tokens are properly rejected
- **Test context cancellation** - Ensure graceful shutdown respects context
- **Use table-driven tests** - Test multiple validation scenarios systematically

Example security test pattern:
```go
func TestTokenRedaction(t *testing.T) {
    tests := []struct {
        name     string
        token    string
        expected string
    }{
        {"normal token", "secret123", "[REDACTED]"},
        {"empty token", "", "<empty>"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := RedactToken(tt.token)
            if result != tt.expected {
                t.Errorf("Expected '%s', got '%s'", tt.expected, result)
            }
        })
    }
}
```

## Project Structure

```
pulumi-webflow/
├── provider/          # Provider implementation
│   ├── provider.go    # Main provider logic
│   ├── schema.go      # Resource schemas
│   └── resources/     # Resource implementations
├── examples/          # Usage examples
│   ├── typescript/    # TypeScript examples
│   ├── python/        # Python examples
│   └── go/            # Go examples
├── sdk/              # Generated SDKs (auto-generated)
└── tests/            # Integration tests
```

## Release Process

Releases are handled by project maintainers. The process includes:

1. Version bump in relevant files
2. Update CHANGELOG.md
3. Create Git tag
4. Build and publish provider binary
5. Generate and publish multi-language SDKs

## Getting Help

- **Questions**: Open a [discussion](https://github.com/jdetmar/pulumi-webflow/discussions)
- **Bugs**: Create an [issue](https://github.com/jdetmar/pulumi-webflow/issues)
- **Community**: Join [Pulumi Community Slack](https://slack.pulumi.com/)

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to the Webflow Pulumi Provider!
