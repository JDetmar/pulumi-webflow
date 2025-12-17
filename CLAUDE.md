# CLAUDE.md - Webflow Pulumi Provider Developer Guide

This document provides comprehensive guidance for Claude instances working on the Webflow Pulumi Provider project.

## Quick Reference

### Essential Commands

**Build and Test:**
```bash
# Build the provider binary for local platform
go build -ldflags "-s -w -X main.Version=0.1.0" -o dist/pulumi-resource-webflow ./main.go

# Run all tests with coverage
go test -v -cover ./...

# Run specific test
go test ./provider -run TestRedirectCreate_ValidationErrors -v

# Run tests with output for a specific file
go test ./provider -run TestRedirect -v

# Run single test function
go test ./provider -run TestRedirectDiff_MultipleFieldsChange -v

# Install locally for testing
make install VERSION=0.1.0
```

**Build for specific platforms (release):**
```bash
make build-all       # Build for all platforms
make package         # Package all binaries
make checksums       # Generate SHA256 checksums
```

**Code Quality:**
```bash
# Format code
go fmt ./...

# Vet code for common errors
go vet ./...

# Check test coverage
go test -cover ./provider/...
```

**SDK Generation:**
```bash
# Generate provider schema from binary
make gen-schema

# Generate all language SDKs from schema
make gen-sdks

# Build and test all SDKs
make build-sdks

# Build individual SDKs
make build-sdk-nodejs    # TypeScript/JavaScript
make build-sdk-python    # Python
make build-sdk-go        # Go
make build-sdk-dotnet    # C#
make build-sdk-java      # Java (requires Maven)

# Clean SDK artifacts
make clean-sdks
```

## Project Structure

### Root Level
- `main.go` - Provider entry point, registers resources (RobotsTxt, Redirect) via `infer` package
- `go.mod` / `go.sum` - Dependency management
- `Makefile` - Build automation for multi-platform support
- `pulumi-plugin.json` - Provider plugin metadata
- `README.md` - User-facing project documentation
- `CONTRIBUTING.md` - Contribution guidelines

### /provider Directory
Core Pulumi provider implementation with resource definitions and API clients:

**Resource Implementations:**
- `redirect_resource.go` - CRUD methods for Redirect resource (Create, Read, Update, Delete, Diff)
- `redirect_resource_test.go` - Comprehensive tests for Redirect CRUD operations and drift detection
- `robotstxt_resource.go` - CRUD methods for RobotsTxt resource (reference pattern for Redirect)
- `robotstxt_test.go` - Extensive tests for RobotsTxt resource

**API Clients & Utilities:**
- `redirect.go` - Webflow API functions: GetRedirects, PostRedirect, PatchRedirect, DeleteRedirect + validation functions
- `redirect_test.go` - Tests for Redirect API functions and mock server patterns
- `robotstxt.go` - Webflow API functions for RobotsTxt (reference pattern)
- `robotstxt_test.go` - Tests for RobotsTxt API functions

**Infrastructure:**
- `auth.go` - HTTP client with exponential backoff, rate limiting (429 handling), context cancellation, User-Agent header
- `auth_test.go` - Tests for authentication and HTTP client behavior
- `config.go` - Provider configuration loading from Pulumi config

### /docs Directory
Development and project documentation:
- `epics.md` - Epic descriptions, requirements, and story breakdowns
- `prd.md` - Product Requirements Document with functional and non-functional requirements
- `state-management.md` - Detailed state management and drift detection explanation
- `UPGRADE.md` - Version upgrade and compatibility notes
- `sprint-artifacts/` - Story files tracking development progress
  - `sprint-status.yaml` - Master tracking file with epic and story status
  - `{epic}-{story}-{name}.md` - Individual story context files with acceptance criteria and dev notes

### /.bmad Directory
BMAD (Build Management and Development) infrastructure for workflow automation:
- `.bmad/bmm/workflows/` - Story creation, development, code review workflows
- `.bmad/bmm/agents/` - AI agent configurations for different roles (dev, PM, architect, etc.)
- `.bmad/bmm/config.yaml` - BMAD configuration

### /.github Directory
GitHub Actions CI/CD pipeline configuration

## Architecture & Code Patterns

### Pulumi Provider Architecture

The provider uses the **Pulumi Go Provider SDK** (`infer` package) which:
1. Auto-generates resource schemas from Go structs
2. Handles gRPC communication with Pulumi CLI
3. Calls resource methods (Create, Read, Update, Delete, Diff) as needed

**Resource Registration (main.go lines 31-42):**
```go
infer.NewProviderBuilder().
    WithConfig(infer.Config(&provider.Config{})).
    WithResources(
        infer.Resource(&provider.RobotsTxt{}),
        infer.Resource(&provider.Redirect{}),
    ).
    Build()
```

### Resource Implementation Pattern

Each resource implements the `CustomResource` interface with these methods:

**Redirect Resource Example:**

1. **Create** (redirect_resource.go:131-199)
   - Validates input using ValidateSourcePath, ValidateDestinationPath, ValidateStatusCode
   - Calls PostRedirect API to create in Webflow
   - Returns new resource with ID in format: `{siteId}/redirects/{redirectId}`
   - Supports DryRun mode (returns empty ID without creating)

2. **Read** (redirect_resource.go:205-249)
   - Parses resource ID to extract siteId and redirectId
   - Fetches current state from Webflow via GetRedirects API
   - Returns empty ID if resource was deleted (signals deletion to Pulumi)
   - Returns currentInputs (from API) for Pulumi to compare with code-defined inputs
   - **Critical for drift detection:** Pulumi compares returned inputs with code-defined inputs

3. **Update** (redirect_resource.go:267-305)
   - Validates inputs using validation functions
   - Calls PatchRedirect API with code-defined values
   - Used when drift detected or when user updates code
   - Returns updated state

4. **Delete** (redirect_resource.go:313-331)
   - Calls DeleteRedirect API
   - Handles 404 responses as idempotent (already deleted)
   - Supports DryRun mode

5. **Diff** (redirect_resource.go:88-131)
   - Compares old state (from Pulumi) with new inputs (from code)
   - Identifies what changed: Delete, Create, or Update
   - Returns DetailedDiff showing individual field changes
   - **Critical for preview:** Shows users what `pulumi preview` will do
   - **Bug Fix (Story 2.2):** Accumulates changes in single map (lines 111-128) to show all changed fields

### Drift Detection Flow

**How Pulumi detects and corrects drift:**

1. **Detection Phase** (`pulumi preview` or `pulumi refresh`):
   - Pulumi calls Read() for each managed resource
   - Read() fetches current state from Webflow API
   - Read() returns both currentInputs and currentState
   - Pulumi compares currentInputs with code-defined inputs
   - If different → drift detected

2. **Identification Phase**:
   - Pulumi calls Diff() with (old state, new inputs)
   - Diff() returns DetailedDiff showing which fields changed
   - Users see exact changes in `pulumi preview` output

3. **Correction Phase** (`pulumi up`):
   - If resource exists with drift: Pulumi calls Update() with code-defined inputs
   - If resource was deleted: Pulumi calls Create() with code-defined inputs
   - Update/Create use code values as source of truth, overwrites manual changes

**Example Drift Scenario:**
```
Code defines: destinationPath="/new", statusCode=302
Webflow has: destinationPath="/old", statusCode=301 (manual change in UI)

1. pulumi preview → Read() fetches API state → returns {"/old", 301}
2. Diff() compares {"/old", 301} vs {"/new", 302} → DetectedDiff = "destinationPath: /old→/new, statusCode: 301→302"
3. pulumi up → Update() sends {"/new", 302} to API → drift corrected
```

### API Client Pattern

All Webflow API functions follow this pattern (redirect.go):

**Structure:**
- Accept context for cancellation
- Use HTTP client from auth.go (with exponential backoff, rate limiting)
- Include error handling and rate limit handling (429 → exponential backoff)
- Return structured responses or errors

**Example: GetRedirects (redirect.go:129-213)**
```go
func GetRedirects(ctx context.Context, client *http.Client, siteId string) (RedirectResponse, error) {
    // 1. Build URL
    url := fmt.Sprintf("https://api.webflow.com/sites/%s/redirects", siteId)

    // 2. Create request with context
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

    // 3. Execute with rate limiting
    resp, err := client.Do(req)

    // 4. Parse response
    var result RedirectResponse
    json.NewDecoder(resp.Body).Decode(&result)

    // 5. Return structured data or error
    return result, nil
}
```

### Validation Pattern

Three-part error messages that tell users: what's wrong, expected format, how to fix.

**Example from ValidateSourcePath (redirect.go:44-68):**
```go
if !sourcePathRegex.MatchString(sourcePath) {
    return fmt.Errorf(
        "invalid source path format: path '%s' contains invalid characters\n"+
        "Expected format: /path-to-page (alphanumerics, hyphens, slashes, underscores only)\n"+
        "Fix: use only letters, numbers, hyphens (-), underscores (_), and forward slashes (/)",
        sourcePath,
    )
}
```

### Testing Pattern

Tests use multiple strategies:

1. **Table-Driven Tests** - Multiple test cases in single test function
2. **Mock HTTP Servers** - httptest.NewServer() to simulate Webflow API responses
3. **DryRun Mode Testing** - Verify operations don't have side effects with DryRun=true
4. **Validation Testing** - Test invalid inputs are rejected before API calls

**Example: Mock Server Pattern (redirect_test.go:313-352)**
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Simulate Webflow API behavior
    response := RedirectResponse{
        Redirects: []RedirectRule{
            {ID: "redirect1", SourcePath: "/old", DestinationPath: "/new", StatusCode: 301},
        },
    }
    json.NewEncoder(w).Encode(response)
}))
defer server.Close()

// Use server.URL instead of real Webflow API
```

## Story Progression & Status

### Epic 1: Provider Foundation (DONE)
All stories complete. Includes:
- Provider project setup and authentication
- RobotsTxt resource implementation
- Basic error handling and validation
- Provider distribution

### Epic 2: Redirect Management (IN PROGRESS)
- **Story 2.1:** Redirect Resource Schema (DONE)
  - Schema definition, validation functions
  - Bug fixed: Inline validation not using proper functions
- **Story 2.2:** Redirect CRUD Operations (DONE)
  - Create, Read, Update, Delete methods implemented
  - Bug fixed: Diff DetailedDiff overwriting instead of accumulating changes
  - Added test: TestRedirectDiff_MultipleFieldsChange
- **Story 2.3:** Drift Detection (READY-FOR-REVIEW)
  - Drift detection already implemented via Read/Diff/Update
  - Added 5 drift detection tests (TestDiff_WithDriftedDestinationPath, etc.)
  - No new code needed, only tests
- **Story 2.4:** State Refresh (BACKLOG)

### Epic 3-7: Future (BACKLOG)
Site lifecycle management, SDK distribution, CI/CD integration, documentation, compliance

## Development Workflow

### Starting Work on a Story

1. **Check Sprint Status:**
   ```bash
   cat docs/sprint-artifacts/sprint-status.yaml
   ```
   Find next backlog story (status: "backlog")

2. **Create Story Context:**
   ```bash
   /bmad:bmm:workflows:create-story
   ```
   Creates comprehensive story file with acceptance criteria and dev notes in `docs/sprint-artifacts/`

3. **Begin Implementation:**
   ```bash
   /bmad:bmm:workflows:dev-story
   ```
   Implements the story, running tests and updating status

4. **Code Review:**
   ```bash
   /bmad:bmm:workflows:code-review
   ```
   Performs adversarial code review, finds issues, fixes them

5. **Complete:**
   Story status updates from "in-progress" → "review" → "done"
   Sprint status file updates automatically

### Common Development Tasks

**Add a new validation function:**
1. Add regex pattern and validation logic to redirect.go
2. Add validation test cases to redirect_test.go
3. Call validation function from Create/Update in redirect_resource.go
4. Add integration tests to redirect_resource_test.go

**Add new test cases:**
1. Follow table-driven test pattern
2. Use mock HTTP servers for API tests
3. Test both success and failure cases
4. Test with DryRun mode where applicable
5. Name tests as: `Test{Operation}_{Scenario}` (e.g., TestRedirectCreate_ValidationErrors)

**Fix a bug:**
1. Write failing test that reproduces the bug
2. Fix the code
3. Verify test passes
4. Run full test suite: `go test -v -cover ./...`

## Key Technical Insights

### Story 2.1 - Redirect Schema
- Learned: Need dedicated validation functions with regex patterns for consistency
- Bug fixed: Create/Update methods had inline validation, not using proper functions
- Impact: Invalid characters would pass in code but fail at API level

### Story 2.2 - Redirect CRUD
- Learned: Diff method must accumulate changes into single map
- Bug fixed: Multiple field changes only showed last field in DetailedDiff
- Impact: Users only saw partial changes in `pulumi preview`
- Prevention: Added TestRedirectDiff_MultipleFieldsChange test

### Story 2.3 - Drift Detection
- Critical insight: Drift detection already implemented via Read/Diff/Update
- No new code needed, only comprehensive testing
- Read() returns API state, Pulumi compares with code, Diff() identifies changes, Update() fixes
- All infrastructure already in place from Story 2.2

### Performance Requirements (NFR3)
- Drift detection must complete within 10 seconds
- Current implementation: <2 seconds typical
- GetRedirects API call: ~200-500ms
- Find redirect in list: O(n) where n = number of redirects per site
- Building response: <1ms
- Status: Well within requirement

## Common Mistakes to Prevent

1. **Don't inline validation logic** - Always use dedicated validation functions
2. **Don't overwrite DetailedDiff maps** - Accumulate changes into single map
3. **Don't skip DryRun testing** - Verify operations don't have side effects
4. **Don't forget empty ID check on Read** - Critical for deletion detection
5. **Don't hardcode Webflow API URLs** - Use fmt.Sprintf with siteId
6. **Don't forget rate limit handling** - auth.go client handles 429, but be aware
7. **Don't test only happy path** - Include invalid input, missing resource, API errors
8. **Don't add unrelated refactoring** - Keep PRs focused on story requirements

## Configuration & Credentials

**Webflow API Token:**
- Set via `WEBFLOW_API_TOKEN` environment variable
- Used in Pulumi project files or CI/CD pipelines
- Example: `export WEBFLOW_API_TOKEN="your-token-here"`

**Pulumi Configuration:**
- Store in Pulumi.yaml (stack configuration)
- Example:
  ```yaml
  name: webflow-dev
  runtime: go
  config:
    webflow:token:
      secure: AAAxx...
  ```

**Local Testing:**
```bash
export WEBFLOW_API_TOKEN="test-token"
go test ./provider -v
```

## References & Resources

**Key Files by Purpose:**

| Purpose | File | Key Lines |
|---------|------|-----------|
| Resource CRUD | redirect_resource.go | Create (131-199), Read (205-249), Update (267-305), Delete (313-331), Diff (88-131) |
| API & Validation | redirect.go | GetRedirects (129-213), PostRedirect (221-317), PatchRedirect (325-420), Validate* (44-97) |
| Resource Tests | redirect_resource_test.go | All test cases with mock scenarios |
| API Tests | redirect_test.go | API function tests, mock server patterns (313-352) |
| Authentication | auth.go | HTTP client with exponential backoff, rate limiting |
| Provider Setup | main.go | Resource registration, provider initialization |
| Story Context | docs/sprint-artifacts/*.md | Story requirements, acceptance criteria, dev notes |
| Sprint Tracking | docs/sprint-artifacts/sprint-status.yaml | Epic and story status |

**Pulumi Provider SDK:**
- Import: `github.com/pulumi/pulumi-go-provider`
- Package: `infer` for automatic schema generation
- Docs: https://pkg.go.dev/github.com/pulumi/pulumi-go-provider

**Webflow API:**
- Endpoints used: GetRedirects, PostRedirect, PatchRedirect, DeleteRedirect
- Rate limiting: 429 handled with exponential backoff
- Authentication: Bearer token in Authorization header

**Go Version:**
- Required: Go 1.24.7 or higher
- Toolchain: 1.24.11

## Troubleshooting

**Tests failing with "connection refused":**
- Tests use mock HTTP servers
- Check that test is creating server properly with httptest.NewServer()
- Verify response format matches expected struct

**Provider not building:**
- Run `go mod tidy` to resolve dependencies
- Check Go version: `go version` (should be 1.24+)
- Verify all imports are present

**API calls failing with 429:**
- auth.go client handles rate limiting with exponential backoff
- Normal behavior - provider will retry automatically
- Check Webflow API rate limits if issue persists

**Drift not being detected:**
- Verify Read() is fetching from Webflow API correctly
- Confirm Read() returns empty ID if resource deleted
- Check that Diff() is comparing inputs correctly
- Run `pulumi preview -v` to see detailed drift detection logs

## Status & Next Steps

**Current Development Phase:** Epic 2 - Redirect Management

**Completed:**
- ✅ Epic 1 (Provider Foundation) - All 9 stories done
- ✅ Story 2.1 (Redirect Schema) - Done
- ✅ Story 2.2 (Redirect CRUD) - Done with bug fix
- ✅ Story 2.3 (Drift Detection) - Ready for review, comprehensive testing added

**Next:**
- Story 2.4 (State Refresh) - When 2.3 approved
- Epic 3 (Site Lifecycle) - Future work
- Epic 4-7 (SDKs, CI/CD, Documentation, Compliance) - Backlog

**How to Proceed:**
1. Review Story 2.3 context and testing if further work needed
2. When ready, run `create-story` for Story 2.4
3. Continue epic progression through to full SDK distribution

This document should be your primary reference for understanding the codebase. When in doubt, check the specific story files in `/docs/sprint-artifacts/` for detailed acceptance criteria and implementation notes.
