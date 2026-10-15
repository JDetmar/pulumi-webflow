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
   Normalizing drops the preamble, code samples, the error list, and the examples, so the hash changes only when the documented contract changes.
3. Scans `provider/*.go` (excluding tests) for `/v2/...` URL templates.
   A documented endpoint is "wrapped" when its path, with parameters erased, matches a provider template.
   Matching is by path only, so an unimplemented method on a wrapped path counts as wrapped.
4. Compares the snapshot with the baseline from the previous run:
   - **gap**: a documented endpoint the provider does not wrap.
   - **drift**: a wrapped endpoint whose normalized page changed.
   - **removed**: an endpoint in the baseline that the reference no longer lists.
5. Drops every finding already recorded in an issue.
   Issues carry hidden markers such as `<!-- api-parity:gap GET /sites/{}/aeo/reports/bot_traffic -->`.
   A marker counts in any issue state, so closing an issue as "not planned" stops it from coming back.
6. Writes `work.json` (the pending findings), a new `baseline/` directory, and `has_work` to `$GITHUB_OUTPUT`.

The agent runs only when `has_work` is true.
The baseline holds hashes for every page and the normalized text for wrapped pages only, about 350 KB.
It lives on the `memory/api-parity` branch, one commit per change, so the branch history doubles as a log of Webflow reference changes.

The workflow's `save_baseline` job commits the new baseline only after both the agent and the safe-output jobs succeed.
If either fails, the old baseline stays and the next run sees the same drift again.
gh-aw `repo-memory` is not used, because it saves memory even when the agent fails.

To re-report drift after a run that filed nothing useful, revert the last commit on `memory/api-parity` and run the workflow again.

## Run it locally

```bash
python3 scripts/api_parity/snapshot.py \
  --provider-dir provider \
  --baseline-dir /tmp/api-parity/baseline \
  --out-dir /tmp/api-parity/out
python3 -m unittest discover -s scripts/api_parity
```

Without `--issues`, nothing counts as already reported.
An empty or missing baseline directory means a first run: gaps are reported, and drift starts from the next run.

## Threat model

Assets: the repository's issue tracker, and the maintainer's attention.
The workflow cannot change code, releases, or settings.

| Threat | Entry point | Mitigation |
|---|---|---|
| Prompt injection in Webflow docs content tells the agent to do something else | Reference pages and diffs in `work.json` | The agent job has read-only permissions. Writes go only through gh-aw safe outputs, capped at a few issues and comments per run, and gh-aw's threat-detection job screens them. The prompt treats page content as data. Network egress is limited to GitHub and `developers.webflow.com`. |
| The docs site returns partial or bad content and the workflow reports false removals | Index or page fetch | Every fetch has a timeout and bounded retries. Any failure fails the job before the agent runs. An index with suspiciously few endpoints is rejected. |
| A hostile issue body forges markers to hide findings | Prior issues read for dedupe | Only collaborators can trigger the workflow, but anyone can open an issue on a public repo. A forged marker can only suppress a report, never cause an action, and suppression is visible in the issue list. Accepted. |
| Runaway cost or issue spam | Agent loop | `timeout-minutes`, `max-turns`, per-output `max`, and the `has_work` gate bound each run. Stop it with `gh aw disable api-parity-maintainer`. |

Receipts: each issue and comment carries gh-aw's footer with the workflow run link, and the baseline branch keeps a commit per change.
Only the `save_baseline` job can write repository contents, it runs no AI, and it writes only the `memory/api-parity` branch.
