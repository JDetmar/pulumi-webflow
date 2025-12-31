# Story 3.7: Import Existing Sites

Status: done

## Story

As a Platform Engineer,
I want to import existing Webflow sites into Pulumi state,
So that I can manage legacy sites through infrastructure code (FR14).

## Acceptance Criteria

**AC1: Import Single Site**

**Given** an existing Webflow site not managed by Pulumi
**When** I run `pulumi import webflow:index:Site mysite <siteId>`
**Then** the provider imports the site into Pulumi state (FR14)
**And** all current site configuration is captured
**And** subsequent `pulumi up` operations manage the imported site

**AC2: Import Multiple Sites Sequentially**

**Given** multiple existing sites need import
**When** I import them sequentially
**Then** each import operation succeeds independently
**And** import operations are properly logged for audit (FR37)

## Tasks / Subtasks

- [x] Task 1: Understand Pulumi Import Mechanism (AC: #1)
  - [x] Research Pulumi's `pulumi import` command and how it works with providers
  - [x] Understand the provider's role in import operations
  - [x] Review RobotsTxt and Redirect import patterns if they exist
  - [x] Identify what methods the provider must implement for import support
  - [x] Document the import flow: user command → Pulumi CLI → provider methods called

- [x] Task 2: Analyze Site Resource Import Requirements (AC: #1)
  - [x] Determine resource ID format for import command: `<workspaceId>/sites/<siteId>` or just `<siteId>`
  - [x] Identify which existing methods support import (Read method critical)
  - [x] Determine if Create method needs modification for import scenario
  - [x] Analyze state mapping: API response → SiteState struct
  - [x] Document what configuration properties are required vs auto-detected on import

- [x] Task 3: Implement or Verify Import Support (AC: #1)
  - [x] Verify Read method properly handles import scenario
  - [x] Read method should accept resource ID, call GetSite, return full state
  - [x] Ensure resource ID parsing works for import format
  - [x] If needed: implement CustomCheck method for import validation
  - [x] Test that imported resources match expected state format

- [x] Task 4: Write Comprehensive Import Tests (AC: #1, #2)
  - [x] Test successful import with valid siteId
  - [x] Test import with workspaceId/sites/siteId format
  - [x] Test import with invalid siteId (404 error)
  - [x] Test import with permission errors (403)
  - [x] Test import captures all site properties correctly
  - [x] Test subsequent pulumi up after import manages site correctly
  - [x] Test multiple sequential imports don't interfere with each other

- [x] Task 5: Create Import Documentation and Examples (AC: #1, #2)
  - [x] Document import command syntax and usage
  - [x] Create example workflow: identify site → import → verify → manage
  - [x] Document how to find siteId from Webflow UI
  - [x] Provide examples for single site and bulk import patterns
  - [x] Document common import errors and troubleshooting
  - [x] Add audit trail examples (Git commit after import)

- [x] Task 6: End-to-End Import Validation (AC: #1, #2)
  - [x] Run full test suite: go test -v -cover ./provider/...
  - [ ] Manually test import with real Webflow site
  - [ ] Verify imported site shows correct state in pulumi preview
  - [ ] Verify pulumi up after import makes no changes (idempotent)
  - [ ] Test drift detection works on imported site
  - [x] Update sprint-status.yaml: mark story as "review" when complete

## Dev Notes

### Architecture & Implementation Patterns

**CRITICAL: Pulumi Import is NOT Create - It's Read + State Adoption**

Pulumi's `pulumi import` command does NOT create new resources. Instead:
1. User specifies existing resource ID (e.g., siteId from Webflow)
2. Pulumi calls provider's **Read** method to fetch current state
3. Pulumi saves that state to state file
4. Future `pulumi up` commands manage the resource normally

**Key Insight: If Read method works, import likely works automatically!**

Story 3.6 already implemented the Read method for Site resources. Import support may already be functional - this story is about **validation, testing, and documentation**.

### Pulumi Import Architecture

**How `pulumi import` Works:**

```bash
# User runs import command
pulumi import webflow:index:Site mysite 64d7f7a60497dc89dd5e80ab

# Pulumi CLI does:
# 1. Parses resource type: webflow:index:Site
# 2. Parses logical name: mysite
# 3. Parses resource ID: 64d7f7a60497dc89dd5e80ab
# 4. Calls provider.Read(ctx, ReadRequest{
#      Id: "workspace123/sites/64d7f7a60497dc89dd5e80ab", # Provider constructs full ID
#      State: empty SiteState
#    })
# 5. Provider returns currentInputs and currentState from Webflow API
# 6. Pulumi saves state with logical name "mysite"
# 7. User's code must define resource "mysite" to match imported state
```

**Provider's Role in Import:**

The provider's **Read** method (already implemented in Story 3.6) handles import automatically:
- Read method accepts resource ID
- Read method calls GetSite API
- Read method returns currentInputs and currentState
- Pulumi saves that state

**What This Story Adds:**
- Validation that import actually works end-to-end
- Comprehensive testing of import scenarios
- Documentation for users on how to import sites
- Error handling specific to import use cases

### Resource ID Format for Import

**Critical Question: What ID format does user provide?**

Option 1: Full ID format (what provider uses internally):
```bash
pulumi import webflow:index:Site mysite workspace123/sites/64d7f7a60497dc89dd5e80ab
```

Option 2: Just siteId (requires provider to construct full ID):
```bash
pulumi import webflow:index:Site mysite 64d7f7a60497dc89dd5e80ab
```

**RECOMMENDED: Option 2 (siteId only)**

Reasoning:
- Users can easily find siteId from Webflow UI or API
- WorkspaceId can be auto-detected by calling GetSite API (response includes workspaceId)
- Simpler user experience
- Matches how users think about resources

**Implementation Approach for Option 2:**

If import provides just siteId, the Read method needs to:
1. Detect if ID is simple siteId (no "/" characters) vs full ID
2. If simple siteId: Call GetSite to fetch site data, which includes workspaceId
3. Construct full resource ID: `{workspaceId}/sites/{siteId}`
4. Return state with full ID

**Modify Read Method (site_resource.go):**

```go
func (r *SiteResource) Read(ctx context.Context, req infer.ReadRequest[SiteArgs, SiteState]) (
	infer.ReadResponse[SiteArgs, SiteState], error) {

	client, err := GetHTTPClient(ctx, "0.1.0")
	if err != nil {
		return infer.ReadResponse[SiteArgs, SiteState]{}, fmt.Errorf("failed to get HTTP client: %w", err)
	}

	id := req.State.Id
	var workspaceId, siteId string

	// Check if this is a simple siteId (import scenario) or full ID (normal read)
	if !strings.Contains(id, "/") {
		// Simple siteId - this is an import scenario
		siteId = id

		// Call GetSite to fetch site data which includes workspaceId
		siteData, err := GetSite(ctx, client, siteId)
		if err != nil {
			return infer.ReadResponse[SiteArgs, SiteState]{}, fmt.Errorf("failed to read site during import (site ID: %s): %w", siteId, err)
		}

		if siteData == nil {
			// Site not found (404)
			return infer.ReadResponse[SiteArgs, SiteState]{
				Inputs: SiteArgs{},
				State: SiteState{
					SiteArgs: SiteArgs{},
					Id:       "", // Empty ID signals deletion
				},
			}, nil
		}

		// Extract workspaceId from API response
		workspaceId = siteData.WorkspaceId

		// Construct full resource ID for future operations
		id = FormatSiteId(workspaceId, siteId)

	} else {
		// Full ID format - normal read scenario
		workspaceId, siteId, err = ParseSiteId(id)
		if err != nil {
			return infer.ReadResponse[SiteArgs, SiteState]{}, fmt.Errorf("invalid resource ID format: %w", err)
		}

		// Continue with existing Read logic...
	}

	// Rest of Read method continues normally...
}
```

### Import Workflow for Users

**Step-by-Step Import Process:**

1. **Identify Site to Import:**
   ```
   User logs into Webflow UI → finds site → copies siteId from URL or API
   Example siteId: 64d7f7a60497dc89dd5e80ab
   ```

2. **Run Import Command:**
   ```bash
   pulumi import webflow:index:Site my-existing-site 64d7f7a60497dc89dd5e80ab
   ```

3. **Pulumi Calls Provider:**
   ```
   - Pulumi calls Read(id = "64d7f7a60497dc89dd5e80ab")
   - Provider detects simple siteId format
   - Provider calls GetSite API
   - API returns full site data including workspaceId
   - Provider constructs full ID: "workspace123/sites/64d7f7a60497dc89dd5e80ab"
   - Provider returns currentInputs and currentState
   - Pulumi saves state
   ```

4. **User Defines Resource in Code:**
   ```typescript
   // User must add this to their Pulumi program
   const mySite = new webflow.Site("my-existing-site", {
       workspaceId: "workspace123",  // From import output
       displayName: "My Existing Site",
       shortName: "my-existing-site",
       timezone: "America/New_York",
       // ... other properties from imported state
   });
   ```

5. **Verify Import:**
   ```bash
   pulumi preview  # Should show "no changes" if code matches imported state
   ```

6. **Manage Site Going Forward:**
   ```bash
   # Modify code, then:
   pulumi up  # Provider calls Update() to apply changes
   ```

### Testing Strategy

**1. Unit Tests for Import ID Handling (provider/site_resource_test.go)**

```go
func TestSiteRead_ImportWithSimpleSiteId(t *testing.T) {
	// Test that Read method handles simple siteId (import scenario)
	// Mock GetSite to return full site data with workspaceId
	// Verify Read constructs full resource ID
	// Verify state is populated correctly
}

func TestSiteRead_ImportSiteNotFound(t *testing.T) {
	// Test import with invalid siteId (404)
	// Verify appropriate error message
	// Error should indicate site doesn't exist
}

func TestSiteRead_ImportPermissionDenied(t *testing.T) {
	// Test import with insufficient permissions (403)
	// Verify clear error message
}

func TestSiteRead_NormalReadStillWorks(t *testing.T) {
	// Verify normal Read (full ID format) still works after import changes
	// Ensure backward compatibility
}
```

**2. Integration Tests for Import Flow (manual/example-based)**

Create example program demonstrating import:

```typescript
// examples/import-existing-site/index.ts

import * as webflow from "@pulumi/webflow";

// After running: pulumi import webflow:index:Site imported-site 64d7f7a60497dc89dd5e80ab

const importedSite = new webflow.Site("imported-site", {
    workspaceId: "workspace123",
    displayName: "Imported Site",
    shortName: "imported-site",
    timezone: "America/New_York",
});

export const siteId = importedSite.id;
```

**3. End-to-End Import Validation**

Manual testing checklist:
- [ ] Import real Webflow site using simple siteId
- [ ] Verify `pulumi preview` shows site state correctly
- [ ] Add resource definition to code matching imported state
- [ ] Run `pulumi preview` → should show "no changes"
- [ ] Modify resource in code → run `pulumi up` → verify update works
- [ ] Verify drift detection works on imported site
- [ ] Import multiple sites sequentially → verify no interference

### Previous Story Intelligence

**From Story 3.6 (Site State Reading Operations - DONE):**

**What was completed:**
- ✅ GetSite API function retrieves current site state from Webflow
- ✅ Read method implemented with full drift detection support
- ✅ Read handles 404 (site deleted) by returning empty ID
- ✅ Read returns currentInputs and currentState for Pulumi
- ✅ 7 comprehensive GetSite tests, all passing
- ✅ All 128 provider tests passing with 64.4% coverage

**Key Learnings from Story 3.6:**

1. **Read Method is Foundation:** The Read method is the critical piece for both drift detection AND import
2. **ID Format Parsing:** ParseSiteId() already exists and handles `workspaceId/sites/siteId` format
3. **GetSite API Pattern:** GetSite returns full site data including workspaceId
4. **Error Handling:** 404 returns nil (not error), 403 handled with clear messaging
5. **Test Coverage:** Mock HTTP servers with httptest.NewServer() pattern proven reliable

**Critical Insight from Story 3.6:**

The Read method is ALREADY implemented and functional. Import support likely already works! This story is about:
- Validating import actually works
- Handling simple siteId format (not just full ID)
- Writing comprehensive import tests
- Creating user documentation

**From Story 2.3 (Drift Detection for Redirects - DONE):**

**Drift Detection Pattern:**
- Read method enables drift detection
- Pulumi compares code inputs vs Read() currentInputs
- Import is just "adopt existing resource state"
- No separate import logic needed if Read works correctly

**From Story 1.5 (RobotsTxt CRUD Operations - DONE):**

**First Resource Implementation:**
- RobotsTxt was first resource implemented
- Check if RobotsTxt supports import (reference pattern)
- If RobotsTxt import works, Site import should follow same pattern

### Git Intelligence from Recent Commits

**Most Recent Site Work (Dec 13, 2025):**

1. **317848a - Site State Reading Operations (Story 3.6):**
   - GetSite API function implemented
   - Read method fully functional with drift detection
   - 7 comprehensive tests for GetSite
   - Read integration tested through full suite
   - **CRITICAL FOR IMPORT:** Read method is the foundation for import

2. **61471d8 - Site Deletion Operations (Story 3.5):**
   - DeleteSite API function with idempotent 404 handling
   - Delete method with comprehensive error handling
   - 8 tests for DeleteSite
   - Pattern: API function first, then resource method

3. **0b23099 - Site Publishing Operations (Story 3.4):**
   - PublishSite API function for async operations
   - Optional publishing after Create/Update
   - 11 tests for PublishSite
   - Pattern: Optional operations integrated cleanly

**Development Velocity & Patterns:**

Epic 3 Progress:
- ✅ Story 3.1: Site Resource Schema Definition
- ✅ Story 3.2: Site Creation Operations
- ✅ Story 3.3: Site Configuration Updates
- ✅ Story 3.4: Site Publishing Operations
- ✅ Story 3.5: Site Deletion Operations
- ✅ Story 3.6: Site State Reading Operations
- ⏳ Story 3.7: Import Existing Sites (THIS STORY)

**Proven Patterns to Reuse:**
- Read method already implemented → import foundation exists
- Test with httptest.NewServer() for API mocking
- Three-part error messages (what's wrong, expected, how to fix)
- Comprehensive test coverage (7-11 tests per feature)
- No new files needed → modify existing site_resource.go and site_resource_test.go

### Technical Requirements & Constraints

**1. Pulumi Import Requirements**

**Provider Interface for Import:**
- Provider must implement Read method (DONE in Story 3.6)
- Read method called when user runs `pulumi import`
- Read must return currentInputs and currentState
- Pulumi saves returned state to state file

**Import ID Format:**
- User provides: `pulumi import webflow:index:Site mysite <ID>`
- `<ID>` can be simple siteId OR full resource ID
- Provider must handle both formats gracefully
- RECOMMENDED: Accept simple siteId, auto-construct full ID

**2. Webflow API Requirements**

**GetSite API (Already Implemented):**
- GET /v2/sites/{siteId}
- Returns full site data including workspaceId
- 200 OK: Site found, return data
- 404 Not Found: Site doesn't exist (import fails)
- 403 Forbidden: Insufficient permissions (import fails)

**No New API Calls Required:**
- Import uses existing GetSite function
- All API infrastructure already in place from Story 3.6

**3. State Management Requirements**

**Import State Capture:**
- All site properties must be captured on import
- WorkspaceId, DisplayName, ShortName, Timezone
- Optional: ParentFolderId (if site in folder)
- Read-only fields: PreviewUrl, CreatedOn, LastUpdated

**Resource ID Construction:**
- Import with siteId only → construct `{workspaceId}/sites/{siteId}`
- GetSite API response provides workspaceId
- Use existing FormatSiteId() function

**4. User Experience Requirements**

**Import Command Simplicity:**
- User should only need siteId (easiest to find in Webflow UI)
- Clear error messages if import fails
- Guidance on how to find siteId

**Post-Import Workflow:**
- After import, `pulumi preview` should show current state
- User writes code to match imported state
- `pulumi preview` again → "no changes" if code matches
- Future `pulumi up` manages imported site normally

**5. Audit Trail Requirements (FR37)**

**Import Logging:**
- Import command execution logged to terminal
- Git commit after import captures state adoption
- Example commit: "feat: Import existing Webflow site my-existing-site (siteId: 64d7f7a60497dc89dd5e80ab)"

**6. Multi-Site Import Requirements (AC2)**

**Sequential Import Support:**
- Each import independent
- No shared state between imports
- Import 10 sites → 10 separate `pulumi import` commands
- Example bulk import script documented

### Library & Framework Requirements

**No New Dependencies Required!**

All functionality achievable with existing infrastructure:

```go
import (
    "context"
    "fmt"
    "strings"

    "github.com/pulumi/pulumi-go-provider/infer"
)
```

**Existing Functions to Reuse:**
- `GetSite(ctx, client, siteId)` - Fetch site data (Story 3.6)
- `ParseSiteId(id)` - Parse full resource ID (Story 3.1)
- `FormatSiteId(workspaceId, siteId)` - Construct resource ID (Story 3.1)
- `GetHTTPClient(ctx, version)` - Authenticated HTTP client (Story 1.2)

**Existing Structs:**
- `SiteData` - API response format (Story 3.1)
- `SiteArgs` - Resource inputs (Story 3.1)
- `SiteState` - Resource state (Story 3.1)

### File Structure & Modification Summary

**Files to Modify:**

1. **provider/site_resource.go** - MODIFY Read method to handle simple siteId
   - Lines to modify: ~10-15 lines
   - Add ID format detection (simple siteId vs full ID)
   - Call GetSite with simple siteId
   - Extract workspaceId from API response
   - Construct full resource ID
   - Location: Read method (around line 293-355)

2. **provider/site_resource_test.go** - ADD import-specific tests
   - Lines to add: ~150-200 lines
   - TestSiteRead_ImportWithSimpleSiteId
   - TestSiteRead_ImportSiteNotFound
   - TestSiteRead_ImportPermissionDenied
   - TestSiteRead_NormalReadStillWorks
   - Add after existing Read tests

3. **examples/import-existing-site/** - CREATE example program
   - New directory with example code
   - index.ts (TypeScript example)
   - README.md (step-by-step import guide)
   - Shows import command and resource definition

4. **docs/** - CREATE or UPDATE import documentation
   - Add import section to README.md or create separate IMPORTING.md
   - Document import command syntax
   - Document how to find siteId
   - Document post-import workflow
   - Document bulk import patterns

**Total Code to Write:** ~165-230 lines
- Read method modification: ~15 lines
- Import tests: ~150-200 lines
- Example code: ~15 lines
- Documentation: varies

### Testing Standards & Coverage Goals

**Test Coverage Targets:**
- Import ID handling: 100% coverage (simple siteId and full ID paths)
- Error scenarios: 100% coverage (404, 403, network errors)
- Overall provider package: maintain/improve 64.4% coverage (NFR23)

**Test Categories:**

1. **Unit Tests for Import ID Handling:**
   - [ ] Import with simple siteId (most common case)
   - [ ] Import with full resource ID (backward compatibility)
   - [ ] Import with site not found (404)
   - [ ] Import with permission denied (403)
   - [ ] Import with network error
   - [ ] Verify full resource ID constructed correctly
   - [ ] Verify workspaceId extracted from API response

2. **Integration Tests for Import Flow:**
   - [ ] Normal Read still works after import changes
   - [ ] Imported resource managed correctly on subsequent pulumi up
   - [ ] Drift detection works on imported resource
   - [ ] Multiple imports don't interfere with each other

3. **Manual/Example-Based Tests:**
   - [ ] Real Webflow site import end-to-end
   - [ ] Verify pulumi preview after import
   - [ ] Verify pulumi up manages imported site
   - [ ] Bulk import script example

**Test Execution:**
```bash
# Run all provider tests
go test -v -cover ./provider/...

# Run only Site Read tests (including import)
go test ./provider -run TestSiteRead -v

# Run specific import test
go test ./provider -run TestSiteRead_ImportWithSimpleSiteId -v

# Check coverage
go test -cover ./provider/...
```

### Common Mistakes to Prevent

Based on learnings from Epic 1, Epic 2, and Epic 3 Stories 3.1-3.6:

1. ❌ **Don't create separate import method** - Read method handles import automatically
2. ❌ **Don't assume full ID format** - Users will provide simple siteId, handle both
3. ❌ **Don't fail on simple siteId** - Call GetSite to fetch workspaceId, construct full ID
4. ❌ **Don't break normal Read** - Ensure existing Read workflow still works
5. ❌ **Don't forget error messages** - Clear guidance for 404 (site not found) and 403 (no permission)
6. ❌ **Don't skip documentation** - Users need examples to understand import workflow
7. ❌ **Don't test only happy path** - Test 404, 403, network errors, invalid IDs
8. ❌ **Don't forget audit trail** - Document Git commit after import for compliance
9. ❌ **Don't assume users know Pulumi import** - Provide step-by-step guide
10. ❌ **Don't forget backward compatibility** - Full ID format must still work

### Import Documentation Examples

**1. Finding SiteId in Webflow UI:**

```markdown
### How to Find Your Site ID

1. Log into Webflow (https://webflow.com)
2. Navigate to your site
3. Look at the URL: https://webflow.com/dashboard/sites/{siteId}/settings
4. Copy the siteId (24-character hexadecimal string)
5. Example: `64d7f7a60497dc89dd5e80ab`
```

**2. Import Command Examples:**

```bash
# Import single site
pulumi import webflow:index:Site my-production-site 64d7f7a60497dc89dd5e80ab

# Bulk import script example
#!/bin/bash
sites=(
  "prod-site-1:64d7f7a60497dc89dd5e80ab"
  "prod-site-2:64d7f7a60497dc89dd5e80ac"
  "staging-site:64d7f7a60497dc89dd5e80ad"
)

for site in "${sites[@]}"; do
  IFS=':' read -r name id <<< "$site"
  pulumi import webflow:index:Site "$name" "$id"
done
```

**3. Post-Import Code Definition:**

```typescript
// After import, add resource definition to match imported state

import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@pulumi/webflow";

const myProductionSite = new webflow.Site("my-production-site", {
    workspaceId: "workspace123",  // From pulumi stack output after import
    displayName: "My Production Site",
    shortName: "my-production-site",
    timezone: "America/New_York",
    // Add other properties to match imported state
});

export const siteId = myProductionSite.id;
```

**4. Verification Workflow:**

```bash
# Step 1: Import site
pulumi import webflow:index:Site my-site 64d7f7a60497dc89dd5e80ab

# Step 2: Review imported state
pulumi stack output

# Step 3: Add resource definition to code (see above)

# Step 4: Verify code matches imported state
pulumi preview
# Should show: "no changes" if code matches imported state

# Step 5: Manage site going forward
# Modify code, then:
pulumi up
```

### Error Message Examples

**Site Not Found (404) During Import:**
```
Error: failed to import site: site not found (site ID: 64d7f7a60497dc89dd5e80ab).

The site ID you provided doesn't exist in Webflow. This could mean:
  - The site ID is incorrect or mistyped
  - The site was deleted
  - You don't have access to this site

Fix:
1. Verify the site ID by logging into Webflow and checking the URL
2. Ensure you have access to the site in your workspace
3. Try again with the correct site ID

Example: pulumi import webflow:index:Site mysite <correct-site-id>
```

**Permission Denied (403) During Import:**
```
Error: failed to import site: insufficient permissions (site ID: 64d7f7a60497dc89dd5e80ab).

Your Webflow API token doesn't have permission to read this site.

Possible reasons:
  - API token doesn't have 'sites:read' scope
  - User doesn't have access to this workspace
  - Site is in a different workspace

Fix:
1. Verify your Webflow API token has the necessary scopes
2. Ensure you have read permissions in the workspace
3. Check that the site ID belongs to a workspace you can access

Check your token permissions at: https://webflow.com/dashboard/integrations/applications
```

**Invalid Site ID Format:**
```
Error: invalid site ID format: "not-a-valid-id"

Site IDs must be 24-character hexadecimal strings (MongoDB ObjectId format).

Example valid site ID: 64d7f7a60497dc89dd5e80ab

Find your site ID:
1. Log into Webflow: https://webflow.com
2. Navigate to your site
3. Check the URL: https://webflow.com/dashboard/sites/{siteId}/settings
4. Copy the 24-character ID

Then retry: pulumi import webflow:index:Site mysite <valid-site-id>
```

### Performance Considerations

**Import Performance:**
- Single import: ~200-500ms (one GetSite API call)
- 10 sites sequential: ~2-5 seconds total
- 100 sites sequential: ~20-50 seconds total
- Well within acceptable limits for one-time operation

**No Performance Optimization Needed:**
- Import is one-time operation per resource
- API call latency is acceptable for manual workflow
- Bulk import handled via script, not provider optimization

### Compliance & Audit Trail (FR37)

**Import Audit Requirements:**

After each import, document in Git:

```bash
# After import, commit the state change
git add .pulumi/
git commit -m "feat: Import existing Webflow site my-production-site

Imported site ID: 64d7f7a60497dc89dd5e80ab
Workspace: workspace123
Display Name: My Production Site
Short Name: my-production-site
Timezone: America/New_York

This site was previously managed manually in Webflow UI.
Now managed through infrastructure as code via Pulumi."

# Push to remote for team visibility
git push origin main
```

**Audit Trail Benefits:**
- Git history shows when site was imported (author, timestamp)
- Commit message documents what was imported
- Future changes tracked in subsequent commits
- Full audit trail from import forward

### References

**Epic & Story Documents:**
- [Epic 3: Site Lifecycle Management](docs/epics.md#epic-3-site-lifecycle-management) - Epic overview and all stories
- [Story 3.7: Import Existing Sites](docs/epics.md#story-37-import-existing-sites) - Original story definition with acceptance criteria
- [Story 3.6: Site State Reading Operations](docs/sprint-artifacts/3-6-site-state-reading-operations.md) - Read method implementation (foundation for import)
- [Story 3.1: Site Resource Schema Definition](docs/sprint-artifacts/3-1-site-resource-schema-definition.md) - Resource ID format and parsing

**Functional Requirements:**
- [FR14: Import existing Webflow sites into managed state](docs/prd.md#functional-requirements) - Core requirement for this story
- [FR37: Track configuration changes through version control](docs/prd.md#functional-requirements) - Audit trail requirement

**Code References:**
- [provider/site_resource.go:293-355](provider/site_resource.go#L293-L355) - Read method (import foundation)
- [provider/site.go:599-692](provider/site.go#L599-L692) - GetSite API function
- [provider/site.go:78-91](provider/site.go#L78-L91) - ParseSiteId function
- [provider/site.go:93-96](provider/site.go#L93-L96) - FormatSiteId function

**External Documentation:**
- [Pulumi Import Command](https://www.pulumi.com/docs/cli/commands/pulumi_import/) - Official Pulumi import documentation
- [Pulumi Provider SDK - Import Support](https://www.pulumi.com/docs/guides/pulumi-packages/how-to-author/#importing-existing-resources) - Provider implementation guide

**Project Documentation:**
- [CLAUDE.md](CLAUDE.md) - Developer guide for Claude instances
- [README.md](README.md) - User-facing project documentation
- [docs/state-management.md](docs/state-management.md) - State management details

## Dev Agent Record

### Context Reference

Story 3.7: Import Existing Sites - Comprehensive developer implementation guide created via create-story workflow

### Agent Model Used

Claude Sonnet 4.5 (via create-story workflow, model ID: claude-sonnet-4-5-20250929)

### Debug Log References

(To be filled by dev agent during implementation)

### Completion Notes List

(To be filled by dev agent during implementation)

### File List

**Files to Create:**
- examples/import-existing-site/index.ts
- examples/import-existing-site/README.md
- docs/IMPORTING.md (or add section to README.md)

**Files to Modify:**
- provider/site_resource.go - MODIFY Read method for simple siteId handling
- provider/site_resource_test.go - ADD import-specific tests
- docs/sprint-artifacts/sprint-status.yaml - UPDATE story status
