# Story 2.1: Redirect Resource Schema Definition

Status: done

## Story

As a Platform Engineer,
I want to define the Redirect resource schema,
So that I can specify redirect rules through infrastructure code.

## Acceptance Criteria

**AC1: Resource Schema Definition**

**Given** I'm writing a Pulumi program
**When** I define a Redirect resource
**Then** the resource accepts required properties: siteId, sourcePath, destinationPath, statusCode
**And** the schema validates statusCode is 301 or 302
**And** the schema validates paths are valid URL paths

**AC2: Validation Before API Calls**

**Given** invalid redirect configuration
**When** I run `pulumi preview`
**Then** validation errors are reported before making API calls (NFR33)
**And** error messages explain the validation failure clearly (NFR32)

## Tasks / Subtasks

- [x] Task 1: Define Redirect data structures (AC: #1)
  - [x] Create RedirectRule struct matching Webflow API v2 format
  - [x] Create RedirectResponse struct for API responses
  - [x] Create RedirectRequest struct for API requests
  - [x] Add JSON tags for proper serialization

- [x] Task 2: Implement path validation (AC: #1, #2)
  - [x] Create ValidateSourcePath function with actionable error messages
  - [x] Create ValidateDestinationPath function with actionable error messages
  - [x] Validate paths start with "/" (relative paths)
  - [x] Validate no query strings in source paths (if required by Webflow)

- [x] Task 3: Implement statusCode validation (AC: #1, #2)
  - [x] Create ValidateStatusCode function
  - [x] Accept only 301 (permanent) or 302 (temporary) redirects
  - [x] Provide clear error messages explaining redirect types

- [x] Task 4: Create Redirect resource ID utilities (AC: #1)
  - [x] Create GenerateRedirectResourceId function (format: {siteId}/redirects/{redirectId})
  - [x] Create ExtractIdsFromRedirectResourceId function
  - [x] Follow pattern established in robotstxt.go

- [x] Task 5: Write comprehensive schema validation tests (AC: #2)
  - [x] Test valid redirect configurations
  - [x] Test invalid statusCode values (400, 200, 500, etc.)
  - [x] Test invalid source paths (missing /, special chars, etc.)
  - [x] Test invalid destination paths
  - [x] Test empty/missing required fields
  - [x] Verify error messages are actionable (include how to fix)

- [x] Task 6: Create Redirect resource schema for Pulumi (AC: #1)
  - [x] Define RedirectArgs input type
  - [x] Define RedirectState output type
  - [x] Add documentation comments for IntelliSense (NFR22)
  - [x] Register resource with provider (main.go:37)
  - [x] Add CRUD stub methods (implemented fully in Story 2.2)
  - [x] Note: siteId validation performed during CRUD operations (Story 2.2)

## Dev Notes

### Architecture & Implementation Patterns

**CRITICAL: Follow Established Patterns from RobotsTxt**

The RobotsTxt resource implementation provides the proven patterns that MUST be followed:

**File Structure Pattern:**
```
provider/
├── redirect.go           # Redirect API logic (like robotstxt.go)
├── redirect_resource.go  # Redirect resource CRUD (like robotstxt_resource.go)
└── redirect_test.go      # Comprehensive tests (like robotstxt_test.go)
```

**Struct Definition Pattern (from robotstxt.go:17-35):**
```go
// RedirectRule represents a redirect configuration in Webflow.
type RedirectRule struct {
    ID              string `json:"id,omitempty"`       // Webflow-assigned redirect ID
    SourcePath      string `json:"sourcePath"`         // Path to redirect from (e.g., "/old-page")
    DestinationPath string `json:"destinationPath"`    // Path to redirect to (e.g., "/new-page")
    StatusCode      int    `json:"statusCode"`         // 301 (permanent) or 302 (temporary)
}
```

**Validation Pattern (from robotstxt.go:44-56):**
```go
// Actionable error messages that explain:
// 1. What is wrong
// 2. What format is expected
// 3. How to fix it
func ValidateSourcePath(path string) error {
    if path == "" {
        return fmt.Errorf("sourcePath is required but was not provided. " +
            "Please provide a valid URL path starting with '/' (e.g., '/old-page', '/blog/2023').")
    }
    if !strings.HasPrefix(path, "/") {
        return fmt.Errorf("sourcePath must start with '/': got '%s'. " +
            "Example valid paths: '/old-page', '/blog/2023', '/products/item-1'.", path)
    }
    return nil
}
```

**Resource ID Pattern (from robotstxt.go:60-77):**
- Format: `{siteId}/redirects/{redirectId}`
- Use `strings.Split` for extraction
- Validate format with clear error messages

### Webflow API Reference

**Webflow Redirects API (v2):**
- Base URL: `https://api.webflow.com/v2/sites/{site_id}/redirects`
- GET: List all redirects for a site
- POST: Create a new redirect
- PATCH: Update an existing redirect
- DELETE: Delete a redirect

**API Response Format:**
```json
{
  "redirects": [
    {
      "id": "redirect_id",
      "sourcePath": "/old-path",
      "destinationPath": "/new-path",
      "statusCode": 301
    }
  ]
}
```

### Project Structure Notes

**Existing Files (DO NOT MODIFY unless necessary):**
- [main.go](../../main.go) - Provider entrypoint
- [provider/config.go](../../provider/config.go) - Provider configuration
- [provider/auth.go](../../provider/auth.go) - Authentication logic
- [provider/robotstxt.go](../../provider/robotstxt.go) - Pattern reference for API logic
- [provider/robotstxt_resource.go](../../provider/robotstxt_resource.go) - Pattern reference for resource CRUD
- [provider/robotstxt_test.go](../../provider/robotstxt_test.go) - Pattern reference for comprehensive tests

**Files to Create:**
- `provider/redirect.go` - Redirect API logic, validation, data structures
- `provider/redirect_resource.go` - Redirect resource registration (Story 2.2)
- `provider/redirect_test.go` - Comprehensive test suite

### Testing Standards

**Follow robotstxt_test.go patterns (67,000+ lines of comprehensive tests):**

1. **Unit tests for validation functions:**
   - Test valid inputs
   - Test invalid inputs with specific error messages
   - Test edge cases (empty strings, special characters)

2. **Test naming convention:**
   - `TestValidateSourcePath_Valid`
   - `TestValidateSourcePath_Empty`
   - `TestValidateSourcePath_MissingSlash`

3. **Table-driven tests pattern:**
```go
tests := []struct {
    name    string
    input   string
    wantErr bool
    errMsg  string
}{
    {"valid path", "/old-page", false, ""},
    {"missing slash", "old-page", true, "must start with '/'"},
}
```

4. **Error message assertions:**
   - Verify error messages contain actionable guidance
   - Check for specific substrings like "Example valid paths:"

### Previous Story Intelligence

**From Epic 1 (Stories 1.1-1.9):**
- All 9 stories completed successfully
- 100+ tests in provider package all passing
- Established patterns in robotstxt.go are production-proven
- Error handling includes rate limiting, network failures, validation
- Code review process catches issues like missing file lists, incomplete task markers

**Key Learnings to Apply:**
1. Always mark tasks complete `[x]` as they're done
2. Populate File List section with all created/modified files
3. Follow established struct patterns exactly
4. Include comprehensive test coverage from the start
5. Use actionable error messages (what's wrong + how to fix)

### References

- [Epic 2: Redirect Management](../epics.md#epic-2-redirect-management) - Epic overview
- [Story 2.1: Redirect Resource Schema Definition](../epics.md#story-21-redirect-resource-schema-definition) - Original story
- [FR6: Create and manage redirects](../epics.md#functional-requirements) - Functional requirement
- [FR7: Update and delete redirects](../epics.md#functional-requirements) - Functional requirement
- [NFR32: Actionable error messages](../epics.md#non-functional-requirements) - Error message standard
- [NFR33: Validate before API calls](../epics.md#non-functional-requirements) - Validation standard
- [provider/robotstxt.go](../../provider/robotstxt.go) - Pattern reference for API logic
- [provider/robotstxt_resource.go](../../provider/robotstxt_resource.go) - Pattern reference for CRUD
- [provider/robotstxt_test.go](../../provider/robotstxt_test.go) - Pattern reference for tests
- [Webflow Redirects API](https://developers.webflow.com/data/reference/redirects/list) - Official API docs

## Senior Developer Review (AI)

**Reviewer:** Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)
**Review Date:** 2025-12-11
**Review Outcome:** ✅ **Approved with Fixes Applied**

### Review Summary

Story implementation was ~95% complete with excellent validation functions and comprehensive tests. Found 1 critical issue where Create/Update methods duplicated validation logic instead of using the properly implemented validation functions. All issues have been resolved.

### Issues Found and Resolved

**Total Issues:** 5 (1 Critical, 2 Medium, 2 Low)
**Status:** All fixed automatically during code review

#### Critical Issues (Fixed)

**Issue #1: Duplicate and Incomplete Path Validation**
- **Severity:** HIGH
- **Location:** [provider/redirect_resource.go](../../provider/redirect_resource.go) Create and Update methods
- **Problem:** Validation functions (ValidateSourcePath, ValidateDestinationPath, ValidateStatusCode) were properly implemented with regex pattern matching for invalid characters, but Create/Update methods didn't use them. Instead, they duplicated validation inline with incomplete checks (missing regex validation for invalid characters like spaces, @, #, ?).
- **Impact:** Paths with invalid characters would pass validation in Create/Update but fail at API level, violating AC#2 "validation before API calls"
- **Resolution:** ✅ Replaced all inline validation with calls to the proper validation functions
- **Files Modified:** [provider/redirect_resource.go](../../provider/redirect_resource.go)
- **Verification:** Added 6 new tests for invalid character handling in CRUD operations; all tests pass

#### Medium Issues (Fixed)

**Issue #2: ValidateStatusCode Not Used**
- **Severity:** MEDIUM
- **Problem:** DRY principle violation - statusCode validation duplicated instead of calling ValidateStatusCode()
- **Resolution:** ✅ Updated Create/Update to use ValidateStatusCode function
- **Files Modified:** [provider/redirect_resource.go](../../provider/redirect_resource.go)

**Issue #3: Missing Test Coverage for Invalid Characters in Resource CRUD**
- **Severity:** MEDIUM
- **Problem:** Validation functions had tests, but Create/Update methods lacked tests for invalid character rejection
- **Resolution:** ✅ Added 6 comprehensive tests for invalid character validation in Create/Update
- **Tests Added:**
  - TestRedirectCreate_ValidationErrors: sourcePath with space, query string, special char
  - TestRedirectCreate_ValidationErrors: destinationPath with space, hash
  - TestRedirectUpdate_ValidationErrors: destinationPath with invalid chars
- **Files Modified:** [provider/redirect_resource_test.go](../../provider/redirect_resource_test.go)
- **Verification:** All 46+ redirect tests pass

#### Low Issues (Fixed)

**Issue #4: Inconsistent Error Message Format**
- **Severity:** LOW
- **Problem:** Minor inconsistency between validation function errors and inline errors
- **Resolution:** ✅ Resolved by using validation functions consistently
- **Impact:** Error messages now uniform across codebase

**Issue #5: Missing Resource ID Validation in Create**
- **Severity:** LOW
- **Problem:** No check that Webflow API returns non-empty redirect ID
- **Resolution:** ✅ Added defensive check after PostRedirect call
- **Files Modified:** [provider/redirect_resource.go](../../provider/redirect_resource.go) line 177-179

### Testing After Fixes

- ✅ All 46+ redirect-specific tests passing
- ✅ All 100+ provider tests passing (no regressions)
- ✅ Invalid character validation working correctly in Create/Update
- ✅ Comprehensive test coverage for edge cases

### Acceptance Criteria Validation

**AC1: Resource Schema Definition** ✅ PASS
- Resource accepts all required properties (siteId, sourcePath, destinationPath, statusCode)
- StatusCode validation enforces 301 or 302
- Path validation enforces valid URL paths with regex pattern matching

**AC2: Validation Before API Calls** ✅ PASS
- All validation now occurs before API calls (after fixes)
- Error messages are actionable with 3-part guidance (what's wrong + expected format + how to fix)
- NFR33 satisfied: Validation before API calls
- NFR32 satisfied: Actionable error messages

### Code Quality Assessment

**Strengths:**
- Excellent validation function implementation with comprehensive regex and error messages
- 40+ tests covering validation functions thoroughly
- Proper resource registration and CRUD stub implementation
- IntelliSense documentation complete
- Follows established RobotsTxt patterns

**Improvements Made:**
- Eliminated duplicate validation logic (DRY principle)
- Added complete invalid character validation in CRUD methods
- Enhanced test coverage for resource-level validation
- Added defensive programming for API response validation

### Recommendation

**✅ APPROVED** - Story meets all acceptance criteria after review fixes. Implementation is production-ready with comprehensive validation, excellent test coverage, and proper error handling.

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

### Completion Notes List

- ✅ All 6 tasks completed successfully
- ✅ Implemented complete Redirect resource schema following RobotsTxt patterns
- ✅ Created comprehensive validation functions with actionable error messages (NFR32)
- ✅ Validation occurs before API calls (NFR33)
- ✅ 46+ unit tests (including 15 edge case tests) for validation functions all passing
- ✅ All 100+ provider tests passing - no regressions
- ✅ Full IntelliSense documentation added to all resource types and fields (NFR22)
- ✅ Resource schema follows Pulumi Go provider SDK patterns exactly
- ✅ Error messages explain what's wrong, expected format, and how to fix (3-part guidance)
- ✅ Redirect resource registered with provider in main.go
- ✅ CRUD stub methods satisfy infer.CustomResource interface requirements
- ✅ Code review completed: Found and fixed 5 issues (1 critical, 2 medium, 2 low)
- ✅ Validation functions now properly used in Create/Update methods (no duplicate logic)
- ✅ Comprehensive invalid character validation added to CRUD operations
- ✅ Added defensive check for empty redirect ID from API
- ✅ All acceptance criteria validated and passing after code review fixes

### File List

**Created:**

- [provider/redirect.go](../../provider/redirect.go) - Redirect API logic, validation functions, data structures (RedirectRule, RedirectResponse, RedirectRequest)
- [provider/redirect_resource.go](../../provider/redirect_resource.go) - Redirect resource schema (RedirectArgs, RedirectState, Annotate functions for IntelliSense, CRUD methods with proper validation)
- [provider/redirect_test.go](../../provider/redirect_test.go) - Comprehensive test suite (40+ tests)
- [provider/redirect_resource_test.go](../../provider/redirect_resource_test.go) - CRUD validation tests (additional 6+ tests for invalid character handling)

**Modified:**

- [main.go](../../main.go) - Registered Redirect resource with provider (line 37)
- [provider/redirect_resource.go](../../provider/redirect_resource.go) - Updated Create/Update to use validation functions (code review fix)
- [provider/redirect_resource_test.go](../../provider/redirect_resource_test.go) - Added 6 tests for invalid character validation
- [docs/sprint-artifacts/2-1-redirect-resource-schema-definition.md](2-1-redirect-resource-schema-definition.md) - Added Senior Developer Review section and updated completion notes

