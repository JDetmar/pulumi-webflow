# Story 3.1: Site Resource Schema Definition

Status: done

## Story

As a Platform Engineer,
I want to define the Site resource schema,
So that I can specify complete site configurations through infrastructure code.

## Acceptance Criteria

**AC1: Resource Schema Definition**

**Given** I'm writing a Pulumi program
**When** I define a Site resource
**Then** the resource accepts properties: displayName, shortName, customDomain (optional), timezone
**And** the schema validates displayName is a non-empty string
**And** the schema validates shortName meets Webflow's constraints
**And** all exported types include clear documentation comments (NFR22)

**AC2: Validation Before API Calls**

**Given** invalid site configuration
**When** I run `pulumi preview`
**Then** validation errors are shown with actionable guidance (NFR32, NFR33)

## Tasks / Subtasks

- [x] Task 1: Define Site data structures (AC: #1)
  - [x] Create Site struct matching Webflow API v2 format
  - [x] Create SiteResponse struct for API responses (list and get operations)
  - [x] Create SiteCreateRequest struct for API requests
  - [x] Add JSON tags for proper serialization (displayName, shortName, timeZone, etc.)
  - [x] Include all Site properties: id, workspaceId, displayName, shortName, timeZone, lastPublished, lastUpdated, previewUrl, parentFolderId, customDomains, dataCollectionEnabled, dataCollectionType

- [x] Task 2: Implement displayName validation (AC: #1, #2)
  - [x] Create ValidateDisplayName function with actionable error messages
  - [x] Validate displayName is non-empty string
  - [x] Validate reasonable length constraints (Webflow likely has max length)
  - [x] Provide clear error messages explaining requirements

- [x] Task 3: Implement shortName validation (AC: #1, #2)
  - [x] Create ValidateShortName function with actionable error messages
  - [x] Validate shortName is lowercase alphanumeric with hyphens only
  - [x] Validate no leading/trailing hyphens
  - [x] Validate reasonable length constraints
  - [x] Provide clear error messages explaining Webflow's slugified constraints

- [x] Task 4: Implement timezone validation (AC: #1, #2)
  - [x] Create ValidateTimeZone function with actionable error messages
  - [x] Validate timezone is a valid IANA timezone identifier (or accept Webflow's supported list)
  - [x] Provide examples of valid timezones in error messages

- [x] Task 5: Create Site resource ID utilities (AC: #1)
  - [x] Create GenerateSiteResourceId function (format: {workspaceId}/sites/{siteId} or just {siteId})
  - [x] Create ExtractSiteIdFromResourceId function
  - [x] Follow pattern established in robotstxt.go and redirect.go

- [x] Task 6: Write comprehensive schema validation tests (AC: #2)
  - [x] Test valid site configurations
  - [x] Test invalid displayName values (empty, too long, etc.)
  - [x] Test invalid shortName values (uppercase, spaces, special chars, leading/trailing hyphens)
  - [x] Test invalid timezone values
  - [x] Test empty/missing required fields
  - [x] Verify error messages are actionable (include how to fix)

- [x] Task 7: Create Site resource schema for Pulumi (AC: #1)
  - [x] Define SiteArgs input type with displayName, shortName (optional), timeZone (optional), workspaceId
  - [x] Define SiteState output type including computed properties (id, lastPublished, lastUpdated, previewUrl, etc.)
  - [x] Add Annotate functions for IntelliSense documentation (NFR22)
  - [x] Register resource with provider (main.go)
  - [x] Add CRUD stub methods (implemented fully in Story 3.2)
  - [x] Note: Full CRUD implementation deferred to Story 3.2 - Site Creation Operations

## Dev Notes

### Architecture & Implementation Patterns

**CRITICAL: Follow Established Patterns from RobotsTxt and Redirect**

The RobotsTxt and Redirect resource implementations provide proven patterns that MUST be followed:

**File Structure Pattern:**
```
provider/
├── site.go              # Site API logic (like robotstxt.go, redirect.go)
├── site_resource.go     # Site resource CRUD (like robotstxt_resource.go, redirect_resource.go)
└── site_test.go         # Comprehensive tests (like robotstxt_test.go, redirect_test.go)
```

**Struct Definition Pattern (from redirect.go and robotstxt.go):**
```go
// Site represents a Webflow site configuration.
type Site struct {
    ID                     string         `json:"id,omitempty"`
    WorkspaceId            string         `json:"workspaceId,omitempty"`
    DisplayName            string         `json:"displayName"`
    ShortName              string         `json:"shortName,omitempty"`
    TimeZone               string         `json:"timeZone,omitempty"`
    LastPublished          string         `json:"lastPublished,omitempty"`
    LastUpdated            string         `json:"lastUpdated,omitempty"`
    PreviewUrl             string         `json:"previewUrl,omitempty"`
    ParentFolderId         string         `json:"parentFolderId,omitempty"`
    CustomDomains          []CustomDomain `json:"customDomains,omitempty"`
    DataCollectionEnabled  bool           `json:"dataCollectionEnabled,omitempty"`
    DataCollectionType     string         `json:"dataCollectionType,omitempty"`
}

type CustomDomain struct {
    ID   string `json:"id"`
    Url  string `json:"url"`
}
```

**Validation Pattern (from redirect.go:44-97):**
```go
// Actionable error messages that explain:
// 1. What is wrong
// 2. What format is expected
// 3. How to fix it
func ValidateDisplayName(displayName string) error {
    if displayName == "" {
        return fmt.Errorf("displayName is required but was not provided.\n" +
            "Expected format: A non-empty string representing your site's name.\n" +
            "Fix: Provide a name for your site (e.g., 'My Marketing Site', 'Company Blog').")
    }
    // Add length validation if needed
    return nil
}

func ValidateShortName(shortName string) error {
    if shortName == "" {
        return nil // shortName is optional - Webflow will generate one from displayName
    }
    // Webflow shortName must be lowercase alphanumeric with hyphens
    shortNameRegex := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
    if !shortNameRegex.MatchString(shortName) {
        return fmt.Errorf("invalid shortName format: '%s' contains invalid characters.\n" +
            "Expected format: lowercase letters, numbers, and hyphens only (e.g., 'my-site', 'company-blog-2024').\n" +
            "Fix: Use only lowercase letters (a-z), numbers (0-9), and hyphens (-). No leading/trailing hyphens.", shortName)
    }
    return nil
}
```

**Resource Schema Pattern (from redirect_resource.go:21-45):**
```go
// SiteArgs defines the input properties for the Site resource.
type SiteArgs struct {
    // WorkspaceId is the Webflow workspace ID where the site will be created.
    // Required for site creation operations (Enterprise workspace required).
    WorkspaceId string `pulumi:"workspaceId"`
    // DisplayName is the name of the site shown in the Webflow dashboard.
    // Required - must be a non-empty string.
    DisplayName string `pulumi:"displayName"`
    // ShortName is the slugified version of the site name used in URLs.
    // Optional - if not provided, Webflow will generate one from displayName.
    ShortName string `pulumi:"shortName,optional"`
    // TimeZone is the IANA timezone identifier for the site.
    // Optional - defaults to Webflow's default timezone if not specified.
    TimeZone string `pulumi:"timeZone,optional"`
    // ParentFolderId is the folder ID where the site will be organized.
    // Optional - site will be placed at workspace root if not specified.
    ParentFolderId string `pulumi:"parentFolderId,optional"`
}

// SiteState defines the output properties for the Site resource.
type SiteState struct {
    SiteArgs
    // LastPublished is the timestamp of the last site publish (read-only).
    LastPublished string `pulumi:"lastPublished,optional"`
    // LastUpdated is the timestamp of the last site update (read-only).
    LastUpdated string `pulumi:"lastUpdated,optional"`
    // PreviewUrl is the URL to a preview image of the site (read-only).
    PreviewUrl string `pulumi:"previewUrl,optional"`
    // CustomDomains is the list of custom domains attached to the site (read-only for now).
    CustomDomains []string `pulumi:"customDomains,optional"`
}
```

### Webflow API Reference

**IMPORTANT: Site Creation Requires Enterprise Workspace**

Per the Webflow API documentation, creating sites programmatically requires an Enterprise workspace. The API endpoint is:

**Create Site:**
- URL: `POST https://api.webflow.com/v2/workspaces/{workspace_id}/sites`
- Scope: `workspace:write`
- Request body: `{ "name": "string", "templateName": "string" (optional), "parentFolderId": "string" (optional) }`
- Note: The API uses "name" in the request but returns "displayName" in the response

**List Sites:**
- URL: `GET https://api.webflow.com/v2/sites`
- Scope: `sites:read`
- Returns: Array of Site objects

**Get Site:**
- URL: `GET https://api.webflow.com/v2/sites/{site_id}`
- Scope: `sites:read`
- Returns: Single Site object

**Publish Site:**
- URL: `POST https://api.webflow.com/v2/sites/{site_id}/publish`
- Scope: `sites:write`
- Request body: `{ "customDomains": ["domain_id"], "publishToWebflowSubdomain": boolean }`
- Rate limit: One successful publish per minute

**Site Object Properties (from API response):**
```json
{
  "id": "string",
  "workspaceId": "string",
  "createdOn": "datetime",
  "displayName": "string",
  "shortName": "string",
  "lastPublished": "datetime",
  "lastUpdated": "datetime",
  "previewUrl": "string",
  "timeZone": "string",
  "parentFolderId": "string",
  "customDomains": [{ "id": "string", "url": "string" }],
  "dataCollectionEnabled": "boolean",
  "dataCollectionType": "enum: always|optOut|disabled"
}
```

### Project Structure Notes

**Existing Files (Reference patterns - DO NOT MODIFY unless registering resource):**
- [main.go](../../main.go) - Provider entrypoint (ADD Site resource registration here)
- [provider/config.go](../../provider/config.go) - Provider configuration
- [provider/auth.go](../../provider/auth.go) - Authentication logic with exponential backoff
- [provider/robotstxt.go](../../provider/robotstxt.go) - Pattern reference for API logic
- [provider/robotstxt_resource.go](../../provider/robotstxt_resource.go) - Pattern reference for resource CRUD
- [provider/redirect.go](../../provider/redirect.go) - Pattern reference for validation functions
- [provider/redirect_resource.go](../../provider/redirect_resource.go) - Pattern reference for resource schema

**Files to Create:**
- `provider/site.go` - Site API logic, validation functions, data structures
- `provider/site_resource.go` - Site resource schema, Annotate functions, CRUD stubs
- `provider/site_test.go` - Comprehensive test suite for validation and API functions

### Testing Standards

**Follow redirect_test.go and robotstxt_test.go patterns:**

1. **Unit tests for validation functions:**
   - Test valid inputs
   - Test invalid inputs with specific error messages
   - Test edge cases (empty strings, special characters, boundary lengths)

2. **Test naming convention:**
   - `TestValidateDisplayName_Valid`
   - `TestValidateDisplayName_Empty`
   - `TestValidateShortName_Uppercase`
   - `TestValidateShortName_SpecialChars`
   - `TestValidateTimeZone_Valid`
   - `TestValidateTimeZone_Invalid`

3. **Table-driven tests pattern:**
```go
tests := []struct {
    name    string
    input   string
    wantErr bool
    errMsg  string
}{
    {"valid name", "My Marketing Site", false, ""},
    {"empty name", "", true, "is required"},
    {"valid shortName", "my-site", false, ""},
    {"uppercase shortName", "My-Site", true, "lowercase"},
    {"shortName with spaces", "my site", true, "invalid characters"},
}
```

4. **Error message assertions:**
   - Verify error messages contain actionable guidance
   - Check for specific substrings like "Fix:" or "Expected format:"

### Previous Story Intelligence

**From Epic 1 & 2 (Stories 1.1-2.4):**
- All stories completed successfully with comprehensive testing
- 100+ tests in provider package all passing
- Established patterns in robotstxt.go and redirect.go are production-proven
- Code review process catches issues like duplicate validation logic
- Diff method must accumulate changes (not overwrite) for multiple field changes

**Key Learnings to Apply:**
1. Always mark tasks complete `[x]` as they're done
2. Populate File List section with all created/modified files
3. Follow established struct patterns exactly
4. Use dedicated validation functions (don't inline validation logic)
5. Use actionable error messages (what's wrong + expected format + how to fix)
6. Include comprehensive test coverage from the start
7. Register resource in main.go

### Git Intelligence

**Recent commits showing development patterns:**
- `f7c3cdf` - Fix compilation errors and incorrect expectations in tests
- `b0911cb` - Support sourcePath in redirects and enhance RobotsTxt handling
- `3393d09` - Implement state refresh for Redirect resources
- `e0a19d7` - Implement drift detection with comprehensive tests

**Patterns observed:**
- Tests are written alongside implementation
- Bug fixes include test updates
- Comprehensive test coverage for edge cases
- Clear commit messages describing what changed

### Technical Constraints & Considerations

1. **Enterprise Workspace Requirement:** Site creation via API requires Enterprise workspace
2. **API Name Mapping:** Request uses "name" but response returns "displayName"
3. **Optional ShortName:** If not provided, Webflow generates from displayName
4. **Rate Limiting:** Site publish is limited to 1/minute - handle gracefully
5. **Custom Domains:** Read-only in this story - domain management is out of MVP scope

### Library & Framework Requirements

**Go Packages (already in use):**
- `github.com/pulumi/pulumi-go-provider` - Pulumi provider SDK
- `github.com/pulumi/pulumi-go-provider/infer` - Resource inference
- `net/http` - HTTP client
- `encoding/json` - JSON serialization
- `regexp` - Validation patterns
- `strings`, `fmt` - String utilities

**No new dependencies required** - all functionality achievable with existing packages.

### References

- [Epic 3: Site Lifecycle Management](../epics.md#epic-3-site-lifecycle-management) - Epic overview
- [Story 3.1: Site Resource Schema Definition](../epics.md#story-31-site-resource-schema-definition) - Original story
- [FR1: Create Webflow sites programmatically](../epics.md#functional-requirements) - Functional requirement
- [FR2: Update Webflow site configurations](../epics.md#functional-requirements) - Functional requirement
- [NFR22: Clear documentation comments](../epics.md#non-functional-requirements) - Documentation standard
- [NFR32: Actionable error messages](../epics.md#non-functional-requirements) - Error message standard
- [NFR33: Validate before API calls](../epics.md#non-functional-requirements) - Validation standard
- [provider/robotstxt.go](../../provider/robotstxt.go) - Pattern reference for API logic
- [provider/redirect.go](../../provider/redirect.go) - Pattern reference for validation functions
- [provider/redirect_resource.go](../../provider/redirect_resource.go) - Pattern reference for resource schema
- [Webflow Sites API - Create Site](https://developers.webflow.com/data/v2.0.0/reference/enterprise/workspace-management/create) - Official API docs
- [Webflow Sites API - Get Site](https://developers.webflow.com/v2.0.0/data/reference/sites/get) - Site properties reference
- [Webflow Sites API - Publish Site](https://developers.webflow.com/data/reference/sites/publish) - Publish endpoint

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Haiku 4.5 (claude-haiku-4-5-20251001)

### Debug Log References

### Completion Notes List

- ✅ All 7 tasks completed successfully
- ✅ Implemented complete Site resource schema following established patterns from RobotsTxt and Redirect
- ✅ Created comprehensive validation functions with actionable error messages (ValidateDisplayName, ValidateShortName, ValidateTimeZone, ValidateWorkspaceID)
- ✅ 40+ unit tests for validation functions all passing (covering valid inputs, invalid inputs, edge cases)
- ✅ All 100+ provider tests passing - no regressions introduced
- ✅ Full IntelliSense documentation added to all Site resource types and fields (NFR22)
- ✅ Site resource schema follows Pulumi Go provider SDK patterns exactly
- ✅ Error messages explain what's wrong, expected format, and how to fix (3-part guidance)
- ✅ Site resource registered with provider in main.go (line 38)
- ✅ CRUD stub methods satisfy infer.CustomResource interface requirements
- ✅ Provider successfully compiled to dist/pulumi-resource-webflow binary
- ✅ All acceptance criteria validated and passing

### File List

**Created:**

- [provider/site.go](../../provider/site.go) - Site API data structures (Site, SiteResponse, SiteCreateRequest) and validation functions (ValidateDisplayName, ValidateShortName, ValidateTimeZone, ValidateWorkspaceID) with actionable error messages
- [provider/site_resource.go](../../provider/site_resource.go) - Site resource schema (SiteArgs, SiteState, SiteResource) with Annotate functions for IntelliSense documentation and CRUD method stubs for future implementation
- [provider/site_test.go](../../provider/site_test.go) - Comprehensive test suite with 40+ tests covering all validation functions, edge cases, and resource ID utilities

**Modified:**

- [main.go](../../main.go) - Registered Site resource with provider (line 38: `infer.Resource(&provider.SiteResource{})`)
- [docs/sprint-artifacts/3-1-site-resource-schema-definition.md](3-1-site-resource-schema-definition.md) - Added completion notes and file list

### Change Log

**Code Review Fixes Applied (2025-12-12):**

- Issue #1 FIXED: Marked all 7 tasks as `[x]` complete (were incorrectly showing `[ ]`)
- Issue #2 FIXED: Updated ValidateTimeZone regex to require capital `E` in `Etc/GMT` variants
- Issue #3 FIXED: Removed misleading `UTC+5:30` from error message examples (not actually supported)
- Issue #4 FIXED: Replaced custom `containsSubstring` helper with standard `strings.Contains`
- Issue #5 FIXED: Updated assertion in TestValidateShortName_LeadingTrailingHyphens to check for `"leading/trailing"`
- Issue #6 FIXED: Removed duplicate helper functions (`containsSubstring`, `findSubstring`)

**Final Test Results:**

- All provider tests passing (68.5s runtime)
- Coverage: 59.9% of statements
- Story status: **done**

