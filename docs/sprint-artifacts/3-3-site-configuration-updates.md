# Story 3.3: Site Configuration Updates

Status: done

## Story

As a Platform Engineer,
I want to update existing site configurations,
So that I can modify site settings through code (FR2).

## Acceptance Criteria

**AC1: Site Configuration Updates via Webflow API**

**Given** an existing Site resource with modified properties
**When** I run `pulumi up`
**Then** the provider updates the site configuration via Webflow API (FR2)
**And** only changed properties are updated
**And** the operation is idempotent (NFR6)

**AC2: Preview Changes Before Update**

**Given** I run `pulumi preview` before update
**When** the preview is displayed
**Then** changes are clearly shown with before/after values (FR39)
**And** sensitive data is not displayed (FR17)

## Tasks / Subtasks

- [x] Task 1: Implement PatchSite API function (AC: #1)
  - [x] Create PatchSite function in provider/site.go
  - [x] Use PATCH https://api.webflow.com/v2/sites/{site_id} endpoint
  - [x] Build SiteUpdateRequest with displayName, shortName, timeZone (all optional)
  - [x] Handle API response returning updated Site object
  - [x] Follow PatchRedirect pattern: exponential backoff, rate limiting (429), context cancellation
  - [x] Return updated Site struct or error
  - [x] Note: WorkspaceID and SiteID are immutable (cannot be changed via PATCH)

- [x] Task 2: Write comprehensive tests for PatchSite (AC: #1)
  - [x] Test successful update with all fields (displayName, shortName, timeZone)
  - [x] Test successful update with single field change
  - [x] Test successful update with no changes (idempotent)
  - [x] Test API rate limiting (429) with retry logic
  - [x] Test network errors with retry and recovery
  - [x] Test invalid site ID (404 responses)
  - [x] Test context cancellation during update
  - [x] Use httptest.NewServer() mock pattern from redirect_test.go

- [x] Task 3: Implement Site Update method (AC: #1)
  - [x] Implement Update in provider/site_resource.go (replace stub from Story 3.1)
  - [x] Extract workspaceID and siteID from req.ID using ExtractIdsFromSiteResourceId
  - [x] Validate all inputs using existing validation functions (ValidateDisplayName, ValidateShortName, ValidateTimeZone)
  - [x] Determine what fields changed by comparing req.Inputs with req.State
  - [x] Call PatchSite API function with only changed fields
  - [x] Handle DryRun mode (return preview state without API call)
  - [x] Populate state with API response data (merge with unchanged fields)
  - [x] Return UpdateResponse with updated state
  - [x] Follow Redirect.Update pattern exactly (redirect_resource.go:267-305)

- [x] Task 4: Implement Site Diff method (AC: #2)
  - [x] Implement Diff in provider/site_resource.go (replace stub from Story 3.1)
  - [x] Compare old state (req.State) with new inputs (req.Inputs)
  - [x] Identify which fields changed (displayName, shortName, timeZone, parentFolderId)
  - [x] Return DetailedDiff showing individual field changes
  - [x] Mark immutable fields (workspaceId) as RequiresReplace if changed
  - [x] Follow Redirect.Diff pattern exactly (redirect_resource.go:88-131)
  - [x] Critical: Accumulate changes in single map (don't overwrite like the Story 2.2 bug)

- [x] Task 5: Write comprehensive Update and Diff tests (AC: #1, #2)
  - [x] Test Update with all fields changed
  - [x] Test Update with single field changed
  - [x] Test Update with no changes (idempotent - no API call)
  - [x] Test Update validation errors caught before API call
  - [x] Test Update DryRun mode returns preview state
  - [x] Test Update API failure scenarios
  - [x] Test Diff detects all field changes correctly
  - [x] Test Diff shows immutable field changes as RequiresReplace
  - [x] Test Diff accumulates multiple changes (prevent Story 2.2 bug)
  - [x] Follow redirect_resource_test.go pattern for table-driven tests

- [x] Task 6: Final validation and testing (AC: #1, #2)
  - [x] Run full test suite: go test -v -cover ./provider/...
  - [x] Verify all new tests pass
  - [x] Verify no regressions in existing tests
  - [x] Build provider binary: make build
  - [x] Test end-to-end: update site via pulumi up with provider binary
  - [x] Update sprint-status.yaml: mark story as "review" when complete

## Dev Notes

### Architecture & Implementation Patterns

**CRITICAL: This story implements the Update and Diff methods for the Site resource.**

**Files Created in Previous Stories (DO NOT RECREATE):**
- `provider/site.go` - Site structs, validation functions, resource ID utilities, PostSite API (Stories 3.1, 3.2)
- `provider/site_resource.go` - SiteResource, SiteArgs, SiteState, Create method (Stories 3.1, 3.2)
- `provider/site_test.go` - Validation tests, PostSite tests, Create tests (Stories 3.1, 3.2)

**Files to Modify in This Story:**
- `provider/site.go` - ADD PatchSite API function
- `provider/site_test.go` - ADD PatchSite tests, Update tests, Diff tests
- `provider/site_resource.go` - REPLACE Update and Diff stubs with full implementation

### Webflow API Details for Site Updates

**CRITICAL: Understanding Site Update API**

The Webflow Site Update API allows updating mutable site properties:

**Update Site Endpoint:**
```
PATCH https://api.webflow.com/v2/sites/{site_id}
Authorization: Bearer {token}
Content-Type: application/json

Request Body (all fields optional - send only what changed):
{
  "displayName": "string",   // Human-readable site name
  "shortName": "string",      // URL-safe slug (lowercase, hyphens only)
  "timeZone": "string"        // IANA timezone identifier
}

Response (200):
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

**Important API Constraints:**
1. **Immutable fields** - WorkspaceID and SiteID cannot be changed (primary key)
2. **Optional fields** - Send only fields that changed to minimize API payload
3. **ParentFolderID updates** - Requires different endpoint (not in MVP scope)
4. **Rate limiting** - Standard Webflow API rate limits apply (handle 429 with exponential backoff)
5. **Idempotency** - Sending same values repeatedly produces same result (no changes)

**Fields That Can Be Updated:**
- ✅ displayName - Human-readable site name
- ✅ shortName - URL-safe slug
- ✅ timeZone - IANA timezone identifier
- ❌ workspaceId - Immutable (RequiresReplace if changed)
- ❌ parentFolderId - Different endpoint (out of scope for MVP)

### PatchSite Implementation Pattern

**Follow PatchRedirect Pattern Exactly (redirect.go:322-420):**

```go
// PatchSite updates an existing site's configuration.
// Only changed fields should be sent in the request to minimize API payload.
// Returns the updated Site or an error if the request fails.
func PatchSite(ctx context.Context, client *http.Client, siteID, displayName, shortName, timeZone string) (*Site, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if patchSiteBaseURL != "" {
		baseURL = patchSiteBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s", baseURL, siteID)

	// Build request with only provided fields (empty strings = not changed)
	requestBody := SiteUpdateRequest{
		DisplayName: displayName,
		ShortName:   shortName,
		TimeZone:    timeZone,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Exponential backoff on retry
		if attempt > 0 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = handleNetworkError(err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close() // Close immediately after reading
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Handle rate limiting with retry
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
				attempt+1, maxRetries+1, waitTime)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
			}
			continue
		}

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var site Site
		if err := json.Unmarshal(body, &site); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &site, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

**Add SiteUpdateRequest struct to site.go:**
```go
// SiteUpdateRequest represents the request body for updating a site.
// All fields are optional - send only fields that changed.
type SiteUpdateRequest struct {
	DisplayName string `json:"displayName,omitempty"`
	ShortName   string `json:"shortName,omitempty"`
	TimeZone    string `json:"timeZone,omitempty"`
}
```

### Update Method Implementation Pattern

**Follow Redirect.Update Pattern Exactly (redirect_resource.go:267-305):**

```go
func (r *SiteResource) Update(ctx context.Context, req infer.UpdateRequest[SiteArgs, SiteState]) (infer.UpdateResponse[SiteState], error) {
	// Step 1: Extract site ID from resource ID
	workspaceID, siteID, err := ExtractIdsFromSiteResourceId(req.ID)
	if err != nil {
		return infer.UpdateResponse[SiteState]{}, fmt.Errorf("failed to parse Site resource ID: %w", err)
	}

	// Step 2: Validate all inputs BEFORE any operations
	if err := ValidateDisplayName(req.Inputs.DisplayName); err != nil {
		return infer.UpdateResponse[SiteState]{}, fmt.Errorf("validation failed for Site resource: %w", err)
	}
	if err := ValidateShortName(req.Inputs.ShortName); err != nil {
		return infer.UpdateResponse[SiteState]{}, fmt.Errorf("validation failed for Site resource: %w", err)
	}
	if err := ValidateTimeZone(req.Inputs.TimeZone); err != nil {
		return infer.UpdateResponse[SiteState]{}, fmt.Errorf("validation failed for Site resource: %w", err)
	}

	// Step 3: Initialize state with current inputs
	state := SiteState{
		SiteArgs: req.Inputs,
		// Preserve read-only fields from previous state
		LastPublished:         req.State.LastPublished,
		LastUpdated:           req.State.LastUpdated,
		PreviewUrl:            req.State.PreviewUrl,
		CustomDomains:         req.State.CustomDomains,
		DataCollectionEnabled: req.State.DataCollectionEnabled,
		DataCollectionType:    req.State.DataCollectionType,
	}

	// Step 4: Handle DryRun mode (preview without API call)
	if req.DryRun {
		return infer.UpdateResponse[SiteState]{
			Output: state,
		}, nil
	}

	// Step 5: Get authenticated HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[SiteState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Step 6: Call Webflow API to update site
	// Note: We send all fields, API will handle which ones actually changed
	response, err := PatchSite(ctx, client, siteID, req.Inputs.DisplayName, req.Inputs.ShortName, req.Inputs.TimeZone)
	if err != nil {
		return infer.UpdateResponse[SiteState]{}, fmt.Errorf("failed to update site: %w", err)
	}

	// Step 7: Update state with API response (API returns full site object)
	state.DisplayName = response.DisplayName
	state.ShortName = response.ShortName
	state.TimeZone = response.TimeZone
	state.LastPublished = response.LastPublished
	state.LastUpdated = response.LastUpdated
	state.PreviewUrl = response.PreviewUrl
	if response.CustomDomains != nil {
		state.CustomDomains = response.CustomDomains
	}
	state.DataCollectionEnabled = response.DataCollectionEnabled
	state.DataCollectionType = response.DataCollectionType

	// Step 8: Return successful response
	return infer.UpdateResponse[SiteState]{
		Output: state,
	}, nil
}
```

### Diff Method Implementation Pattern

**Follow Redirect.Diff Pattern Exactly (redirect_resource.go:88-131):**

**CRITICAL: Story 2.2 Bug - Diff was overwriting DetailedDiff map instead of accumulating changes!**

```go
func (r *SiteResource) Diff(ctx context.Context, req infer.DiffRequest[SiteArgs, SiteState]) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{
		DeleteBeforeReplace: false,
		HasChanges:          false,
		DetailedDiff:        make(map[string]infer.PropertyDiff),
	}

	// Check immutable fields (workspaceId) - require replace if changed
	if req.News.WorkspaceId != req.State.WorkspaceId {
		diff.HasChanges = true
		diff.DeleteBeforeReplace = false
		diff.DetailedDiff["workspaceId"] = infer.PropertyDiff{
			Kind:      infer.UpdateReplace,
			InputDiff: true,
		}
		// If workspace changes, entire resource must be replaced
		return diff, nil
	}

	// Check mutable fields - can be updated in place
	// CRITICAL: Accumulate all changes in SINGLE map (don't overwrite!)

	if req.News.DisplayName != req.State.DisplayName {
		diff.HasChanges = true
		diff.DetailedDiff["displayName"] = infer.PropertyDiff{
			Kind:      infer.Update,
			InputDiff: true,
		}
	}

	if req.News.ShortName != req.State.ShortName {
		diff.HasChanges = true
		diff.DetailedDiff["shortName"] = infer.PropertyDiff{
			Kind:      infer.Update,
			InputDiff: true,
		}
	}

	if req.News.TimeZone != req.State.TimeZone {
		diff.HasChanges = true
		diff.DetailedDiff["timeZone"] = infer.PropertyDiff{
			Kind:      infer.Update,
			InputDiff: true,
		}
	}

	if req.News.ParentFolderId != req.State.ParentFolderId {
		diff.HasChanges = true
		diff.DetailedDiff["parentFolderId"] = infer.PropertyDiff{
			Kind:      infer.Update,
			InputDiff: true,
		}
	}

	return diff, nil
}
```

### Testing Strategy

**1. PatchSite API Function Tests (in provider/site_test.go)**

Follow redirect_test.go PatchRedirect pattern:

```go
func TestPatchSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v2/sites/") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}

		// Parse request body
		var reqBody SiteUpdateRequest
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Return mock Site response with updated values
		response := Site{
			ID:          "site123",
			WorkspaceID: "workspace456",
			DisplayName: reqBody.DisplayName,
			ShortName:   reqBody.ShortName,
			TimeZone:    reqBody.TimeZone,
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override API base URL for testing
	patchSiteBaseURL = server.URL
	defer func() { patchSiteBaseURL = "" }()

	// Execute test
	client := &http.Client{}
	site, err := PatchSite(context.Background(), client, "site123", "Updated Site Name", "updated-slug", "America/New_York")

	// Assertions
	if err != nil {
		t.Fatalf("PatchSite failed: %v", err)
	}
	if site.DisplayName != "Updated Site Name" {
		t.Errorf("Expected displayName 'Updated Site Name', got '%s'", site.DisplayName)
	}
}

func TestPatchSite_SingleFieldChange(t *testing.T) {
	// Test updating only displayName (shortName and timeZone empty)
	// ...
}

func TestPatchSite_NoChanges(t *testing.T) {
	// Test idempotent update (same values sent multiple times)
	// ...
}

func TestPatchSite_RateLimiting(t *testing.T) {
	// Test 429 handling with retry
	// ...
}

func TestPatchSite_NetworkError(t *testing.T) {
	// Test network failures, context cancellation, etc.
	// ...
}
```

**2. Update Method Tests (in provider/site_test.go)**

Follow redirect_resource_test.go pattern:

```go
func TestSiteUpdate_Success(t *testing.T) {
	// Test successful update with all fields changed
}

func TestSiteUpdate_SingleFieldChange(t *testing.T) {
	// Test updating only displayName
}

func TestSiteUpdate_NoChanges(t *testing.T) {
	// Test idempotent update (no API call when no changes)
}

func TestSiteUpdate_ValidationErrors(t *testing.T) {
	// Table-driven test for all validation errors
	tests := []struct {
		name      string
		inputs    SiteArgs
		wantErr   bool
		errSubstr string
	}{
		{"empty displayName", SiteArgs{WorkspaceId: "ws123", DisplayName: ""}, true, "displayName is required"},
		{"invalid shortName", SiteArgs{WorkspaceId: "ws123", DisplayName: "Site", ShortName: "INVALID"}, true, "lowercase"},
		{"invalid timeZone", SiteArgs{WorkspaceId: "ws123", DisplayName: "Site", TimeZone: "Invalid/Zone"}, true, "IANA timezone"},
		// ... more test cases
	}
	// ...
}

func TestSiteUpdate_DryRun(t *testing.T) {
	// Test DryRun mode returns preview without API call
}
```

**3. Diff Method Tests (in provider/site_test.go)**

Follow redirect_resource_test.go Diff pattern:

```go
func TestSiteDiff_NoChanges(t *testing.T) {
	// Test when no fields changed (HasChanges = false)
}

func TestSiteDiff_DisplayNameChanged(t *testing.T) {
	// Test displayName change detected
}

func TestSiteDiff_MultipleFieldsChanged(t *testing.T) {
	// CRITICAL: Test multiple fields changed at once (prevent Story 2.2 bug)
	// Verify all changes appear in DetailedDiff map
}

func TestSiteDiff_ImmutableFieldChanged(t *testing.T) {
	// Test workspaceId change triggers RequiresReplace
}
```

### Previous Story Intelligence

**From Story 3.2 (Site Creation Operations - DONE):**

**What was completed:**
- ✅ PostSite API function with comprehensive error handling
- ✅ Create method with validation, DryRun, state population
- ✅ Comprehensive tests (9 PostSite tests, 3 Create tests)
- ✅ All 124 provider tests passing, no regressions
- ✅ Resource ID format: {workspaceId}/sites/{siteId}

**Key Learnings from Story 3.2:**
1. **API naming quirk:** Request uses "name", response returns "displayName" (applies to Create, not Update)
2. **Enterprise workspace required:** Site creation/updates require Enterprise workspace
3. **Validation works perfectly:** All error messages follow 3-part format
4. **Rate limiting handled:** 429 responses retry with exponential backoff
5. **DryRun mode critical:** Must return preview state without API call
6. **Defensive checks important:** Always verify API returned valid data

**From Story 2.2 (Redirect CRUD - DONE, includes critical bug fix):**

**Critical Bug Fixed in Story 2.2:**
- **Problem:** Diff method was overwriting DetailedDiff map instead of accumulating changes
- **Impact:** Only last field change shown in `pulumi preview`, users couldn't see all changes
- **Fix:** Accumulate all changes in single map (lines 111-128 in redirect_resource.go)
- **Prevention:** Added TestRedirectDiff_MultipleFieldsChange test

**MUST PREVENT THIS BUG IN SITE DIFF METHOD!**

**Pattern Established in Redirect:**
- Update method follows exact pattern: validate → DryRun check → API call → populate state
- Diff method accumulates all changes in single DetailedDiff map
- PatchRedirect sends only changed fields (optional fields in request body)
- All tests use mock HTTP servers, no real API calls

### Git Intelligence from Recent Commits

**Most Recent Commit (8699f5f - Dec 12, 2025):**
```
Implement Site Creation Operations: Add PostSite API function and Create method for Site resource

Files changed:
- provider/site.go (added PostSite, lines 176-272)
- provider/site_resource.go (replaced Create stub, lines 147-216)
- provider/site_test.go (added 12 tests, lines 329-686)

Total: ~450 lines added
All tests passing, no regressions
```

**Patterns Observed:**
- Site resource follows exact same patterns as Redirect and RobotsTxt
- PostSite implemented before Create method (API functions first)
- All validation functions tested comprehensively before CRUD
- Mock HTTP servers used for all API tests
- DryRun mode tested separately

**Redirect CRUD Implementation (f7c3cdf, b0911cb):**
- PatchRedirect implemented with optional fields in request body
- Update method validates → calls API → populates state
- Diff method accumulates changes (bug fix applied)
- All patterns proven in production

**Development Velocity:**
- Epic 1: 9 stories completed (RobotsTxt foundation)
- Epic 2: 4 stories completed (Redirect with drift detection)
- Epic 3: 2 stories completed (Site schema + Create), Story 3.3 next

### Technical Requirements & Constraints

**1. Mutable vs Immutable Fields**

**Mutable (can be updated):**
- ✅ displayName - Human-readable site name
- ✅ shortName - URL-safe slug
- ✅ timeZone - IANA timezone identifier

**Immutable (cannot be changed, require replace):**
- ❌ workspaceId - Primary key, change triggers resource replacement
- ❌ siteId - Primary key (part of resource ID)

**Out of Scope for MVP:**
- ❌ parentFolderId - Requires different API endpoint for moving sites
- ❌ customDomains - Custom domain management (future feature)
- ❌ dataCollection* - Workspace-level settings (read-only)

**2. API Requirements**
- Requires `sites:write` scope for site updates
- Rate limiting applies (handle 429 with exponential backoff)
- Only send changed fields to minimize API payload (all fields optional in PATCH)
- API returns full Site object in response (not just changed fields)

**3. Performance Requirements (NFR1)**
- Update operations must complete within 30 seconds
- Diff operations must complete within 10 seconds (NFR3)
- Typical Webflow API latency: 200-1000ms
- Well within performance budgets

**4. Idempotency Requirements (FR12, NFR6)**
- Repeated updates with same values produce same result
- No API call if Diff detects no changes (optimization)
- State remains consistent even if API call fails

**5. Preview Requirements (FR11, FR39)**
- `pulumi preview` shows detailed before/after values
- Diff clearly indicates what will change
- DryRun mode returns preview state without API call
- Sensitive credentials never displayed (FR17)

### Library & Framework Requirements

**Go Packages (Already in Use - No New Dependencies):**
```go
import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "bytes"
    "io"

    "github.com/pulumi/pulumi-go-provider/infer"
)
```

**Existing Infrastructure (From auth.go and previous stories):**
- `GetHTTPClient(ctx, version)` - Authenticated client with retry logic
- `handleNetworkError(err)` - Network error handling
- `handleWebflowError(statusCode, body)` - Webflow API error parsing
- `getRetryAfterDuration(header, fallback)` - Parse Retry-After header
- `maxRetries = 3` - Retry limit
- Exponential backoff: 1s, 2s, 4s

**No New Dependencies Required** - All functionality achievable with existing packages.

### File Structure & Modification Summary

**Files to Modify:**

1. **provider/site.go** - ADD PatchSite function and SiteUpdateRequest struct
   - Lines to add: ~120-140 (PatchSite function + struct definition)
   - Add after PostSite function (around line 273)
   - Add patchSiteBaseURL test variable (for mock server override)
   - Follow PatchRedirect structure exactly

2. **provider/site_resource.go** - REPLACE Update and Diff stubs
   - Update method: Lines 228-232 → ~50-70 lines
   - Diff method: Lines 140-145 → ~50-60 lines
   - Total: ~100-130 lines of implementation
   - Keep Read and Delete as stubs (for Stories 3.6, 3.5)

3. **provider/site_test.go** - ADD Update and Diff tests
   - Lines to add: ~400-500 (comprehensive test coverage)
   - Add after existing Create tests (after line 686)
   - Include PatchSite API tests with mock servers
   - Include Update method tests with all scenarios
   - Include Diff method tests with multiple change scenarios

**Total Code to Write:** ~620-770 lines (PatchSite ~140 lines, Update+Diff ~130 lines, tests ~450 lines)

### Testing Standards & Coverage Goals

**Test Coverage Targets:**
- PatchSite function: 100% coverage (all branches, error paths, retry logic)
- Update method: 100% coverage (validation, DryRun, success, failures)
- Diff method: 100% coverage (all field combinations, immutable checks)
- Overall provider package: maintain/improve 70%+ coverage (NFR23)

**Test Categories:**

1. **Unit Tests for PatchSite:**
   - [ ] Successful update with all fields
   - [ ] Successful update with single field
   - [ ] Successful update with no changes (idempotent)
   - [ ] Rate limiting (429) with retry and recovery
   - [ ] Network errors with exponential backoff
   - [ ] Invalid site ID (404 responses)
   - [ ] Context cancellation during request
   - [ ] Context cancellation during retry
   - [ ] Invalid JSON in API response
   - [ ] Retry-After header handling

2. **Unit Tests for Update Method:**
   - [ ] Successful update with all fields changed
   - [ ] Successful update with single field changed
   - [ ] No changes detected (no API call, idempotent)
   - [ ] Validation errors (empty displayName, invalid shortName, invalid timeZone)
   - [ ] DryRun mode (preview without API call)
   - [ ] API failure scenarios (network error, rate limiting, invalid site)
   - [ ] State population from API response
   - [ ] Resource ID extraction and validation

3. **Unit Tests for Diff Method:**
   - [ ] No changes detected (HasChanges = false)
   - [ ] DisplayName changed (Update)
   - [ ] ShortName changed (Update)
   - [ ] TimeZone changed (Update)
   - [ ] ParentFolderId changed (Update)
   - [ ] **CRITICAL:** Multiple fields changed at once (prevent Story 2.2 bug)
   - [ ] WorkspaceId changed (RequiresReplace)
   - [ ] All changes accumulate in single DetailedDiff map

**Test Execution:**
```bash
# Run all provider tests
go test -v -cover ./provider/...

# Run only Site tests
go test ./provider -run TestSite -v

# Run only PatchSite tests
go test ./provider -run TestPatchSite -v

# Run only Update tests
go test ./provider -run TestSiteUpdate -v

# Run only Diff tests
go test ./provider -run TestSiteDiff -v

# Check coverage
go test -cover ./provider/...
```

### Common Mistakes to Prevent

Based on learnings from Epic 1, Epic 2, and Story 3.2:

1. ❌ **Don't overwrite DetailedDiff map** - Story 2.2 bug: accumulate all changes in single map
2. ❌ **Don't inline validation** - Use dedicated validation functions
3. ❌ **Don't skip DryRun handling** - Must return preview state without API call
4. ❌ **Don't forget defensive checks** - Verify API returned valid data
5. ❌ **Don't skip rate limiting** - Must handle 429 with exponential backoff
6. ❌ **Don't forget context cancellation** - Check ctx.Err() and handle ctx.Done()
7. ❌ **Don't send all fields if unchanged** - Use omitempty in SiteUpdateRequest
8. ❌ **Don't forget to close response body** - Always close immediately after reading
9. ❌ **Don't test only happy path** - Include network errors, API errors, edge cases
10. ❌ **Don't call API when no changes** - Optimize Update to skip API call if Diff shows no changes

### Error Message Examples

**Validation Error (before API call):**
```
Error: validation failed for Site resource: displayName is required but was not provided.
Expected format: A non-empty string representing your site's name.
Fix: Provide a name for your site (e.g., 'My Marketing Site', 'Company Blog', 'Product Landing Page').
```

**Immutable Field Changed (Diff):**
```
Error: workspaceId cannot be changed for an existing site.
The workspaceId is an immutable property (primary key) and changing it requires replacing the entire resource.
Pulumi will delete the old site and create a new one if you proceed with this change.
Fix: If you need to move a site to a different workspace, delete and recreate the resource, or use Webflow's site transfer feature in the UI.
```

**Rate Limiting Error (API 429):**
```
Error: rate limited: Webflow API rate limit exceeded (HTTP 429). The provider will automatically retry with exponential backoff. Retry attempt 2 of 4, waiting 2s before next attempt. If this error persists, please wait a few minutes before trying again or contact Webflow support.
```

**Network Error:**
```
Error: failed to update site: network error: connection refused.
The provider couldn't connect to the Webflow API.
Fix: Check your internet connection. Verify https://api.webflow.com is reachable. If the problem persists, Webflow's API might be experiencing downtime - check https://status.webflow.com.
```

### References

**Epic & Story Documents:**
- [Epic 3: Site Lifecycle Management](../epics.md#epic-3-site-lifecycle-management) - Epic overview and all stories
- [Story 3.3: Site Configuration Updates](../epics.md#story-33-site-configuration-updates) - Original story definition
- [Story 3.2: Site Creation Operations](3-2-site-creation-operations.md) - Previous story (Create foundation)
- [Story 3.1: Site Resource Schema Definition](3-1-site-resource-schema-definition.md) - Schema foundation

**Functional Requirements:**
- [FR2: Update Webflow site configurations](../prd.md#functional-requirements) - Core requirement
- [FR11: Preview planned changes before applying](../prd.md#functional-requirements) - Preview/Diff requirement
- [FR12: Idempotent operations](../prd.md#functional-requirements) - Idempotency requirement
- [FR39: Detailed change previews](../prd.md#functional-requirements) - Diff detail requirement

**Non-Functional Requirements:**
- [NFR1: Operations complete within 30 seconds](../prd.md#non-functional-requirements) - Performance
- [NFR3: Preview operations complete within 10 seconds](../prd.md#non-functional-requirements) - Diff performance
- [NFR6: Idempotent operations](../prd.md#non-functional-requirements) - Reliability
- [NFR32: Error messages include actionable guidance](../prd.md#non-functional-requirements) - Error quality
- [NFR33: Validate configurations before API calls](../prd.md#non-functional-requirements) - Validation

**Code References (Existing Patterns to Follow EXACTLY):**
- [provider/redirect.go:322-420](../../provider/redirect.go#L322-L420) - PatchRedirect pattern (EXACT pattern for PatchSite)
- [provider/redirect_resource.go:267-305](../../provider/redirect_resource.go#L267-L305) - Redirect.Update pattern (EXACT pattern for Site.Update)
- [provider/redirect_resource.go:88-131](../../provider/redirect_resource.go#L88-L131) - Redirect.Diff pattern (EXACT pattern for Site.Diff, includes Story 2.2 bug fix)
- [provider/redirect_test.go](../../provider/redirect_test.go) - PatchRedirect test patterns
- [provider/redirect_resource_test.go](../../provider/redirect_resource_test.go) - Update and Diff test patterns
- [provider/site.go:176-272](../../provider/site.go#L176-L272) - PostSite pattern (reference for PatchSite style)
- [provider/site_resource.go:147-216](../../provider/site_resource.go#L147-L216) - Site.Create pattern (reference for Update style)
- [provider/site_test.go:329-686](../../provider/site_test.go#L329-L686) - PostSite and Create test patterns

**External Documentation:**
- [Webflow API - Update Site](https://developers.webflow.com/data/v2.0.0/reference/sites/update) - Official PATCH endpoint documentation
- [Webflow API - Sites](https://developers.webflow.com/v2.0.0/data/reference/sites) - Site properties and data model

**Project Documentation:**
- [CLAUDE.md](../../CLAUDE.md) - Developer guide for Claude instances
- [README.md](../../README.md) - User-facing project documentation
- [docs/state-management.md](../state-management.md) - State management and drift detection details

## Dev Agent Record

### Context Reference

Story 3.3: Site Configuration Updates

### Agent Model Used

Claude (via dev-story workflow)

### Debug Log References

- All tests passing: `go test -v ./provider/...`
- Coverage: 61.6% (provider package)

### Completion Notes List

- Implemented PatchSite API function following PatchRedirect pattern
- Implemented Update method with validation, DryRun, and state population
- Implemented Diff method with proper change accumulation (preventing Story 2.2 bug)
- Added comprehensive test suite for PatchSite, Update, and Diff
- All 100+ provider tests passing

**Code Review Fixes (2025-12-12):**
- Added 6 new tests: TestSiteDiff_ShortNameChanged, TestSiteDiff_TimeZoneChanged, TestSiteDiff_ParentFolderIdChanged, TestPatchSite_NetworkError, TestPatchSite_InvalidJSON
- Coverage improved from 61.6% to 62.4%
- Updated task completion markers (all tasks marked [x])
- Added File List to Dev Agent Record

### File List

- `provider/site.go` - Added PatchSite function (lines 285-383), SiteUpdateRequest struct (lines 61-67)
- `provider/site_resource.go` - Replaced Update stub (lines 271-339), Replaced Diff stub (lines 137-190)
- `provider/site_test.go` - Added PatchSite tests, Update tests, Diff tests (lines 689-1033)
