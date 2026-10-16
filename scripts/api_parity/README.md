# API parity snapshot

`snapshot.py` compares the Webflow Data API reference with the endpoints this provider calls.
It feeds the `api-parity-maintainer` agentic workflow (`.github/workflows/api-parity-maintainer.md`).

## Why the docs site and not the OpenAPI spec

The published spec at `webflow/openapi-spec` is incomplete.
In October 2026 it lacked whole sections, such as AEO Analytics, that the reference documents.
The official SDK (`webflow/js-webflow-api`) and the Webflow MCP server also cover only a subset.
The reference index at `https://developers.webflow.com/data/v2.0.0/llms.txt` lists every reference page, and each page is available as Markdown by appending `.md` to its URL.

## What it does

1. Reads the reference index and downloads every reference page.
2. Extracts the method and path from each page, and normalizes the page text.
   Normalizing drops the preamble, code samples, the error list, and the examples.
3. Hashes only the page's contract: the endpoint, content type, headings, each parameter or field with its type and required-ness, allowed values, and required scopes.
   Wording edits to descriptions change the stored text but not the hash, so they are not drift.
4. Scans `provider/*.go` (excluding tests and comment lines) for `/v2/...` URL templates.
   Each template is paired with the method of the `http.NewRequest` call in the same function.
   A documented endpoint is "wrapped" when its method and path, with parameters erased, match a provider call.
   A template in a function that sends no request matches any method on its path.
5. Compares the snapshot with the baseline:
   - **gap**: a documented endpoint the provider does not wrap.
   - **drift**: a wrapped endpoint whose contract hash changed.
   - **removed**: an endpoint in the baseline that the reference no longer lists.
   - **undocumented**: a method and path the provider calls that no reference page documents.
     This catches endpoints that left the reference before the first baseline, and methods the provider sends that the reference never listed, such as `PUT` where the reference documents `PATCH`.
     Calls are grouped by the provider files that make them.
6. Drops every finding whose marker is already published in an issue or an issue comment, in any state.
   Markers are inline code spans, such as `` `api-parity:gap GET /sites/{}/aeo/reports/bot_traffic` ``, because gh-aw's output sanitizer removes HTML comments.
   Closing an issue as "not planned" therefore stops its finding from coming back.
7. Writes `work.json` (the pending findings), the baseline to save, and `has_work` to `$GITHUB_OUTPUT`.
   Each gap section lists the provider files that already cover part of it, and sections with such files come first, so the agent reads the existing resource before recommending a change.
   The normalized text of every documented page, and an index of them, go in `reference/` next to `work.json`, and each finding names its page file (`page_file`) and links its reference page (`url`).
   The agent reads the reference from there, so it needs no network access to Webflow.

The agent runs only when `has_work` is true.

## The baseline

The baseline holds the contract hash of every endpoint and the normalized text of wrapped pages, about 350 KB.
It lives on the `memory/api-parity` branch, one commit per change, so the branch history doubles as a log of Webflow reference changes.

The `save_baseline` job commits it after every successful snapshot, whether or not the agent runs or succeeds.
A drift or removal stays pending until its marker is published: until then, the saved baseline keeps that endpoint's previous state, and every run reports it again.
A preview run, a failed agent, or an exhausted output budget therefore loses nothing.
Once an issue or comment carries the marker, the next snapshot advances the baseline for that endpoint.

gh-aw `repo-memory` is not used, because it lives inside the agent job and saves even when the agent fails.

## Run it locally

```bash
python3 scripts/api_parity/snapshot.py \
  --provider-dir provider \
  --baseline-dir /tmp/api-parity/baseline \
  --out-dir /tmp/api-parity/out
python3 -m unittest discover -s scripts/api_parity
```

Without `--issues` and `--comments`, nothing counts as already reported.
An empty or missing baseline directory means a first run: gaps are reported, and drift starts from the next run.

## Threat model

Assets: the repository's issue tracker, and the maintainer's attention.
The workflow cannot change code, releases, or settings.

| Threat | Entry point | Mitigation |
|---|---|---|
| Prompt injection in Webflow docs content tells the agent to do something else | Reference pages and diffs in `work.json` | The agent job has read-only permissions. Writes go only through gh-aw safe outputs, capped at a few issues and comments per run, and gh-aw's threat-detection job screens them. The prompt treats page content as data. The agent reads the reference from files the snapshot job wrote, and its network egress is limited to GitHub. |
| The docs site returns partial or bad content and the workflow reports false removals | Index or page fetch | Every fetch has a timeout and bounded retries, and any failure fails the job before the agent runs. The job also fails when the index or the pages document suspiciously few endpoints, when more than 10% of the baseline's endpoints (and more than 5) vanish in one run, when a page that documented an endpoint has no endpoint line, or when two pages document the same endpoint. |
| A transient GitHub error hides the baseline and a fresh one overwrites unpublished findings | Restoring the `memory/api-parity` branch | Only "branch does not exist" counts as a first run; any other `git` failure, or a branch without `manifest.json`, fails the job. |
| A hostile issue or comment forges markers to hide findings | Prior issues and comments read for dedupe | Only collaborators can trigger the workflow, but anyone can open an issue or comment on a public repo. A forged marker can only suppress a report, never cause an action, and it is visible in the issue. Accepted. |
| Runaway cost or issue spam | Agent loop | `timeout-minutes`, `max-turns`, per-output `max`, and the `has_work` gate bound each run. Stop it with `gh aw disable api-parity-maintainer`. |

Receipts: each issue and comment carries gh-aw's footer with the workflow run link, and the baseline branch keeps a commit per change.
Only the `save_baseline` job can write repository contents, it runs no AI, and it writes only the `memory/api-parity` branch.
