# Webflow Pulumi Provider Examples

Comprehensive examples for using the Webflow Pulumi Provider across all supported languages and use cases.

## 📚 Table of Contents

- [Quick Start](#quick-start)
- [Language-Specific Guides](#language-specific-guides)
- [Resource Examples](#resource-examples)
- [Complex Scenarios](#complex-scenarios)
- [Testing](#testing)
- [Best Practices](#best-practices)

## Quick Start

### Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) 3.0+
- [Webflow Account](https://webflow.com) with API access
- Language-specific runtime (Node.js, Python, Go, .NET, Java)

### Basic Workflow

```bash
# 1. Choose an example directory
cd examples/robotstxt/typescript

# 2. Install dependencies
npm install

# 3. Create a new Pulumi stack
pulumi stack init dev

# 4. Configure your Webflow site ID
pulumi config set webflow:siteId your-site-id --secret

# 5. Preview and deploy
pulumi preview
pulumi up

# 6. Clean up
pulumi destroy
```

## Language-Specific Guides

### TypeScript / Node.js

**Getting Started:**
```bash
npm install
npm run build
pulumi up
```

**Example Locations:**
- Quickstart: `examples/quickstart/typescript/`
- RobotsTxt: `examples/robotstxt/typescript/`
- Redirects: `examples/redirect/typescript/`
- Site Management: `examples/site/typescript/`

**Key Files:**
- `package.json` - Dependencies
- `tsconfig.json` - TypeScript configuration
- `index.ts` - Main Pulumi program
- `Pulumi.yaml` - Stack configuration

### Python

**Getting Started:**
```bash
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
pulumi up
```

**Example Locations:**
- Quickstart: `examples/quickstart/python/`
- RobotsTxt: `examples/robotstxt/python/`
- Redirects: `examples/redirect/python/`
- Site Management: `examples/site/python/`

**Key Files:**
- `requirements.txt` - Python dependencies
- `__main__.py` - Main Pulumi program
- `Pulumi.yaml` - Stack configuration

### Go

**Getting Started:**
```bash
go mod download
pulumi up
```

**Example Locations:**
- Quickstart: `examples/quickstart/go/`
- RobotsTxt: `examples/robotstxt/go/`
- Redirects: `examples/redirect/go/`
- Site Management: `examples/site/go/`

**Key Files:**
- `go.mod` - Go module definition
- `main.go` - Main Pulumi program
- `Pulumi.yaml` - Stack configuration

### C# / .NET

**Getting Started:**
```bash
dotnet restore
dotnet build
pulumi up
```

**Example Locations:**
- RobotsTxt: `examples/robotstxt/csharp/`
- Redirects: `examples/redirect/csharp/`
- Site Management: `examples/site/csharp/`

**Key Files:**
- `*.csproj` - Project file with dependencies
- `Program.cs` - Main Pulumi program
- `Pulumi.yaml` - Stack configuration

### Java

**Getting Started:**
```bash
mvn install
pulumi up
```

**Example Locations:**
- RobotsTxt: `examples/robotstxt/java/`
- Redirects: `examples/redirect/java/`
- Site Management: `examples/site/java/`

**Key Files:**
- `pom.xml` - Maven dependencies
- `App.java` - Main Pulumi program
- `Pulumi.yaml` - Stack configuration

## Resource Examples

### RobotsTxt Resource

The `robots.txt` file controls search engine crawler behavior on your site.

```
examples/robotstxt/
├── typescript/    - Complete TypeScript example
├── python/        - Complete Python example
├── go/            - Complete Go example
├── csharp/        - Complete C# example
└── java/          - Complete Java example
```

**What's Included:**
- Creating `robots.txt` files
- Allow all crawlers pattern
- Selective blocking patterns
- Directory restrictions
- Crawler-specific rules

**See:** [RobotsTxt README](robotstxt/typescript/README.md)

### Redirect Resource

Manage URL redirects (301 permanent, 302 temporary, external).

```
examples/redirect/
├── typescript/    - Complete TypeScript example
├── python/        - Complete Python example
├── go/            - Complete Go example
├── csharp/        - Complete C# example
└── java/          - Complete Java example
```

**What's Included:**
- Permanent redirects (301)
- Temporary redirects (302)
- External domain redirects
- Bulk redirect patterns
- Redirect management

**See:** [Redirect examples documentation](redirect/typescript/README.md)

### Site Resource

Manage Webflow sites including creation, configuration, and publishing.

```
examples/site/
├── typescript/    - Complete TypeScript example
├── python/        - Complete Python example
├── go/            - Complete Go example
├── csharp/        - Complete C# example
└── java/          - Complete Java example
```

**What's Included:**
- Site creation and configuration
- Custom domain setup
- Timezone configuration
- Site publishing
- Site state management
- Import existing sites

**See:** [Site management documentation](site/typescript/README.md)

## Complex Scenarios

### Multi-Site Management

Manage multiple Webflow sites in a single Pulumi program.

```
examples/multi-site/
├── basic-typescript/
├── basic-python/
├── basic-go/
├── config-driven-typescript/
├── template-python/
└── multi-env-go/
```

**Features:**
- Create and manage multiple sites
- Configure different options per site
- Manage resources across sites
- Environment-specific configurations

### Multi-Environment Stack Configuration

Deploy to different environments (dev, staging, production).

```
examples/stack-config/
├── typescript-complete/
├── python-workflow/
├── go-advanced/
```

**Features:**
- Environment-specific configuration
- Stack-based deployments
- Configuration inheritance
- Secret management

### CI/CD Integration

Integrate Pulumi deployments with CI/CD pipelines.

```
examples/ci-cd/
├── github-actions/
```

**Features:**
- Automated deployments
- Pull request previews
- Environment promotion
- Deployment automation

### Logging and Troubleshooting

Debug and troubleshoot Pulumi deployments.

```
examples/troubleshooting-logs/
├── typescript-troubleshooting/
├── python-cicd-logging/
└── go-log-analysis/
```

**Features:**
- Debug logging patterns
- Error analysis
- Performance monitoring
- Log aggregation

## Testing

### Running Example Tests

Tests validate that examples work correctly with the Webflow provider.

```bash
# Run all example tests
cd /path/to/pulumi-webflow
go test -v ./examples

# Run specific test
go test -v ./examples -run TestTypeScriptRobotsTxt

# Run with coverage
go test -v -cover ./examples
```

### Example Test Structure

```go
// examples/robotstxt_test.go
package examples

import (
  "path/filepath"
  "testing"
  "github.com/pulumi/providertest/pulumitest"
)

func TestTypeScriptRobotsTxtExample(t *testing.T) {
  test := pulumitest.NewPulumiTest(t,
    filepath.Join("robotstxt", "typescript"),
    opttest.YarnLink("pulumi-webflow"),
    opttest.AttachProviderServer("webflow", providerFactory),
  )

  test.Preview(t)
  test.Up(t)
  test.Destroy(t)
}
```

### Testing Coverage

- ✅ TypeScript examples tested with Yarn
- ✅ Python examples tested with pip
- ✅ Go examples tested with go mod
- ✅ C# examples tested with dotnet
- ✅ Java examples tested with Maven

## Best Practices

### 1. Configuration Management

**Secret Values:**
```bash
# Store secrets securely
pulumi config set webflow:siteId your-site-id --secret
```

**Configuration Files:**
```yaml
# Pulumi.yaml
name: my-project
runtime: nodejs

config:
  webflow:siteId:
    description: Webflow site ID
    secret: true
  environment:
    description: Deployment environment
    default: development
```

### 2. Naming Conventions

Different languages have different naming conventions:

| Language   | Convention  | Example              |
| ---------- | ----------- | -------------------- |
| TypeScript | camelCase   | `siteId`, `content`  |
| Python     | snake_case  | `site_id`, `content` |
| Go         | PascalCase  | `SiteId`, `Content`  |
| C#         | PascalCase  | `SiteId`, `Content`  |
| Java       | camelCase   | `siteId`, `content`  |

### 3. Error Handling

**TypeScript/JavaScript:**
```typescript
try {
  const robot = new webflow.RobotsTxt("example", {...});
} catch (error) {
  console.error("Failed to create robots.txt:", error);
}
```

**Python:**
```python
try:
  robot = webflow.RobotsTxt("example", ...)
except Exception as e:
  print(f"Failed to create robots.txt: {e}")
```

**Go:**
```go
robot, err := webflow.NewRobotsTxt(ctx, "example", &webflow.RobotsTxtArgs{...})
if err != nil {
  return fmt.Errorf("failed to create robots.txt: %w", err)
}
```

### 4. Resource Organization

**Do:**
- Group related resources logically
- Use meaningful resource names
- Document complex configurations
- Test your deployments

**Don't:**
- Hardcode sensitive values
- Create resources without naming them
- Deploy without previewing first
- Ignore error messages

### 5. Production Deployments

**Pre-Deployment Checklist:**
- [ ] Review `pulumi preview` output
- [ ] Verify configuration values
- [ ] Run tests locally
- [ ] Check resource dependencies
- [ ] Backup existing configurations

**Deployment:**
```bash
pulumi preview    # Review changes
pulumi up         # Deploy
pulumi stack output  # Verify results
```

## Troubleshooting

### "Site not found" Error

```
Error: webflow::RobotsTxt creation failed: site not found
```

**Solution:**
1. Verify your site ID: Settings → General in Webflow
2. Ensure correct format: `abc123def456`
3. Check API token has access to site

### Import Errors

```
Error: Cannot find module 'pulumi-webflow'
```

**Solution:**
```bash
# TypeScript/JavaScript
npm install
npm install --save pulumi-webflow

# Python
pip install -r requirements.txt
pip install pulumi-webflow

# Go
go get github.com/jdetmar/pulumi-webflow/sdk/go/webflow

# C#
dotnet add package Pulumi.Webflow

# Java
# Add to pom.xml or build.gradle
```

### Authentication Issues

```
Error: Authentication failed - invalid API token
```

**Solution:**
1. Check Pulumi config: `pulumi config list`
2. Update credentials: `pulumi config set webflow:apiToken ... --secret`
3. Verify token has necessary permissions in Webflow

## Additional Resources

- [Webflow API Documentation](https://developers.webflow.com/reference/webflow-rest-api)
- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Provider API Reference](../docs/api/)
- [Quickstart Guide](../docs/quickstart.md)
- [API Documentation](../docs/api-reference.md)

## Contributing

Have an example you'd like to share? We'd love to include it!

1. Create your example in the appropriate directory
2. Include README with setup instructions
3. Add tests using pulumitest
4. Submit a pull request

## Support

- [GitHub Issues](https://github.com/jdetmar/pulumi-webflow/issues)
- [Discussions](https://github.com/jdetmar/pulumi-webflow/discussions)
- [Webflow Slack](https://webflow.com/slack)
