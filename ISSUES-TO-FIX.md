# Issues to Fix

Issues discovered during testing session on 2026-01-09.

## 1. RegisteredScript Update Returns 404

**Status:** ✅ FIXED in PR #51

**File:** `provider/registeredscript_resource.go` (Update method)

**Problem:** When Pulumi detects a diff in RegisteredScript (e.g., version change), it calls the Update method which uses `PutRegisteredScript`. The Webflow API returns 404.

**Root Cause:** Webflow API v2 does not support PATCH/PUT for registered scripts. Only GET (list), POST (create), and DELETE are available.

**Solution:** Changed all field changes to trigger replacement (delete + recreate) instead of in-place update. All changes now use `p.UpdateReplace` in the Diff method with `DeleteBeforeReplace = true`.

---

## 2. SiteCustomCode Script ID Format

**Status:** ✅ RESOLVED (was blocked by Issue 1 & 4)

**File:** `provider/sitecustomcode_resource.go`

**Problem:** When creating SiteCustomCode with a registered script, the API returns "invalid id and version".

**Root Cause:** This issue was caused by Issues 1 and 4. When RegisteredScript had problems (update 404 errors, false version diffs triggering unnecessary replacements), the script ID would become unknown during replacement, causing SiteCustomCode validation to fail.

**Resolution:** After fixing Issues 1 and 4, SiteCustomCode works correctly. The script ID format (human-readable string derived from displayName) is correct as documented.

---

## 3. getTokenInfo / getAuthorizedUser Invoke Functions Crash

**Status:** ❌ NOT YET INVESTIGATED

**Files:** `provider/token.go`, `provider/authorized_user.go`

**Problem:** When using these invoke functions in Pulumi YAML, the provider crashes with gRPC errors.

**Error:**
```
error: rpc error: code = Unknown desc = invocation of webflow:index:getTokenInfo returned an error: grpc: the client connection is closing
```

**Investigation needed:**
- Check if invoke functions are implemented correctly
- May be a context/lifecycle issue
- Test invoke functions in isolation

---

## 4. RegisteredScript Version Diff on Every Run

**Status:** ✅ FIXED in PR #51 + additional fix

**File:** `provider/registeredscript_resource.go` (Diff method)

**Problem:** Even after refresh, Pulumi keeps detecting a version diff on RegisteredScript, triggering unnecessary updates.

**Root Causes:**

1. All changes should trigger replacement, not update (fixed in PR #51)
2. Pulumi's struct embedding doesn't properly deserialize the version field into the embedded `RegisteredScriptResourceArgs` struct, causing `req.State.Version` to be empty when compared

**Solution:**

- Changed version field back to optional in struct tag (for backwards compatibility with existing state)
- Updated Diff method to only compare version if both state and inputs have non-empty values
- Create method still validates that version is provided for new resources

---

## 5. Asset Variants Parsing Error

**Status:** ✅ FIXED

**File:** `provider/asset.go`

**Problem:** The Asset resource fails to read with a JSON parsing error.

**Error:**
```
error: Preview failed: failed to read asset: failed to parse response: json: cannot unmarshal array into Go struct field AssetResponse.variants of type map[string]provider.AssetVariant
```

**Root Cause:** The Webflow API returns `variants` as an array, but the Go struct expected a map.

**Solution:**

- Changed `Variants` field from `map[string]AssetVariant` to `[]AssetVariant`
- Updated `AssetVariant` struct fields to match API response: `hostedUrl`, `originalFileName`, `displayName`, `format`, `width`, `height`, `quality`, `error`

---

## 6. CollectionItem Slug Uniqueness Error

**Status:** ✅ ALREADY FIXED (verified by api-verifier audit)

**File:** `provider/collectionitem_resource.go`

**Problem:** When updating a CollectionItem, the API rejects the request with a slug uniqueness error even when the slug hasn't changed.

**Error:**

```text
error: failed to update collection item: bad request: {"message":"Validation Error","code":"validation_error","details":[{"param":"slug","description":"Unique value is already in database: 'test-blog-post'"}]}
```

**Root Cause:** The update request includes the unchanged slug, and Webflow rejects it because the slug already exists (for the same item).

**Solution:** Code at lines 351-368 in `collectionitem_resource.go` already excludes unchanged slug from PATCH requests. The error may have occurred during a transient state before the fix was applied.

---

## Test Stack Location

Test stack is at `/private/tmp/test-webflow` with these working resources:

- Site
- Redirect
- RobotsTxt
- Collection (with new `collectionId` output)
- CollectionField
- CollectionItem ✅ (slug exclusion working)
- Webhook
- AssetFolder
- Asset ✅ (variants parsing fixed)
- RegisteredScript ✅ (fully working)
- SiteCustomCode ✅ (fully working)
