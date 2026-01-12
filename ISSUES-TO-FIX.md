# Issues to Fix

Issues discovered during testing session on 2026-01-09.

## 1. RegisteredScript Update Returns 404

**File:** `provider/registeredscript_resource.go` (Update method)

**Problem:** When Pulumi detects a diff in RegisteredScript (e.g., version change), it calls the Update method which uses `PutRegisteredScript`. The Webflow API returns 404.

**Error:**
```
error: failed to update registered script: not found: the Webflow site or robots.txt configuration does not exist.
```

**Investigation needed:**
- Check if Webflow's API supports updating registered scripts at all
- Verify the PUT endpoint URL is correct
- May need to implement as delete+recreate instead of update

---

## 2. SiteCustomCode Script ID Format

**File:** `provider/sitecustomcode_resource.go`

**Problem:** When creating SiteCustomCode with a registered script, the API returns "invalid id and version".

**Error:**
```
error: failed to create site custom code: bad request: {"message":"Bad Request: At least one script contained an invalid id and version","code":"bad_request"}
```

**Investigation needed:**
- Verify what format the `id` field expects (displayName vs script ID)
- Check if version must match exactly what's registered
- Test with hardcoded values to isolate the issue
- May be blocked until RegisteredScript issues are resolved

---

## 3. getTokenInfo / getAuthorizedUser Invoke Functions Crash

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

**File:** `provider/registeredscript_resource.go` (Diff method)

**Problem:** Even after refresh, Pulumi keeps detecting a version diff on RegisteredScript, triggering unnecessary updates.

**Root cause:** Webflow's list scripts API doesn't return the `version` field. The Read method preserves version from state, but something causes Diff to still see a change.

**Workaround applied:** Made `version` optional in schema (line 43), but this may have side effects.

**Investigation needed:**
- Debug what values Diff receives for `req.State.Version` vs `req.Inputs.Version`
- May need to adjust how Read populates the response
- Consider if version should trigger replacement instead of update

---

## Test Stack Location

Test stack is at `/private/tmp/test-webflow` with these working resources:
- Site
- Redirect
- RobotsTxt
- Collection (with new `collectionId` output)
- CollectionField
- CollectionItem
- Webhook
- AssetFolder
- RegisteredScript (Read works, Update broken)

SiteCustomCode is defined but not yet created due to above issues.
