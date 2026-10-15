#!/usr/bin/env python3
"""Compare the Webflow Data API reference with the endpoints the provider calls.

Writes work.json (findings not yet recorded in an issue) and a new baseline
directory for the api-parity-maintainer agentic workflow. See README.md.
"""

import argparse
import concurrent.futures
import dataclasses
import difflib
import hashlib
import json
import os
import re
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

DOCS_INDEX_URL = "https://developers.webflow.com/data/v2.0.0/llms.txt"
FETCH_TIMEOUT_SECONDS = 30
FETCH_ATTEMPTS = 3
FETCH_WORKERS = 8
# The index listed ~160 reference pages in October 2026. Far fewer means the
# docs site served something unexpected, and reporting every missing endpoint
# as removed would flood the issue tracker.
MIN_INDEX_ENTRIES = 100
MAX_DIFF_LINES = 300
BASELINE_VERSION = 1
MARKER_PREFIX = "api-parity:"

_INDEX_LINE = re.compile(
    r"^- API Reference > (?P<section>.+?) \[(?P<title>.+)\]"
    r"\((?P<url>https://developers\.webflow\.com/data/v2\.0\.0/reference/(?P<slug>[^)\s]+)\.md)\)\s*$"
)
_ENDPOINT_LINE = re.compile(r"^(GET|POST|PUT|PATCH|DELETE) https://api\.webflow\.com/v2(/\S*)\s*$")
_PATH_PARAM = re.compile(r"\{[^}]+\}")
_GO_V2_LITERAL = re.compile(r'"[^"\n]*?/v2(/[^"\n]*)"')
_GO_VERB = re.compile(r"%[sdv]")
_MARKER = re.compile(r"<!--\s*" + re.escape(MARKER_PREFIX) + r"(gap|drift|removed) (.+?)\s*-->")


@dataclasses.dataclass(frozen=True)
class IndexEntry:
    section: str
    title: str
    url: str
    slug: str


@dataclasses.dataclass(frozen=True)
class Page:
    entry: IndexEntry
    method: str | None
    path: str | None
    normalized: str

    @property
    def key(self) -> str | None:
        """Stable identity of an endpoint: method plus path with parameter names erased."""
        if self.method is None or self.path is None:
            return None
        return f"{self.method} {normalize_path(self.path)}"

    @property
    def digest(self) -> str:
        return hashlib.sha256(self.normalized.encode()).hexdigest()


def normalize_path(path: str) -> str:
    """Erase parameter names so /sites/{site_id} and /sites/%s compare equal."""
    path = path.split("?", 1)[0]
    path = _PATH_PARAM.sub("{}", path)
    return path.rstrip("/") or "/"


def parse_index(text: str) -> list[IndexEntry]:
    entries = []
    for line in text.splitlines():
        match = _INDEX_LINE.match(line)
        if match:
            entries.append(IndexEntry(match["section"].strip(), match["title"].strip(), match["url"], match["slug"]))
    return entries


def normalize_page(text: str) -> str:
    """Keep only the documented contract: title, endpoint, description, request, response, types.

    Drops the site-wide preamble, fenced code samples, the error list and the
    examples, which change for reasons unrelated to the API contract.
    """
    kept = []
    in_code = False
    skipping = False
    seen_title = False
    for line in text.splitlines():
        if not seen_title:
            if not line.startswith("# "):
                continue
            seen_title = True
        if line.startswith("```"):
            in_code = not in_code
            continue
        if in_code:
            continue
        if line.startswith("## "):
            heading = line[3:].strip()
            if heading == "Examples":
                break
            skipping = heading == "Errors"
            if skipping:
                continue
        if skipping:
            continue
        kept.append(line.rstrip())
    return re.sub(r"\n{3,}", "\n\n", "\n".join(kept)).strip() + "\n"


def parse_page(entry: IndexEntry, text: str) -> Page:
    method = path = None
    for line in text.splitlines():
        match = _ENDPOINT_LINE.match(line)
        if match:
            method, path = match[1], match[2]
            break
    return Page(entry, method, path, normalize_page(text))


def provider_paths(provider_dir: Path) -> dict[str, list[str]]:
    """Map each normalized /v2 path the provider builds to the files that build it."""
    found: dict[str, set[str]] = {}
    for go_file in sorted(provider_dir.glob("*.go")):
        if go_file.name.endswith("_test.go"):
            continue
        for literal in _GO_V2_LITERAL.findall(go_file.read_text()):
            path = normalize_path(_GO_VERB.sub("{}", literal))
            found.setdefault(path, set()).add(go_file.name)
    return {path: sorted(files) for path, files in sorted(found.items())}


def reported_markers(issues: list[dict]) -> tuple[set[tuple[str, str]], dict[str, int]]:
    """Collect markers from prior issues, in any state.

    Returns every (kind, value) marker, and for drift the open issue number per
    endpoint key so a new change can be added there as a comment.
    """
    markers: set[tuple[str, str]] = set()
    open_drift: dict[str, int] = {}
    for issue in issues:
        for kind, value in _MARKER.findall(issue.get("body") or ""):
            markers.add((kind, value))
            if kind == "drift" and issue.get("state", "").upper() == "OPEN":
                endpoint_key = value.rsplit(" ", 1)[0]
                open_drift.setdefault(endpoint_key, issue["number"])
    return markers, open_drift


def marker(kind: str, value: str) -> str:
    return f"<!-- {MARKER_PREFIX}{kind} {value} -->"


def load_baseline(baseline_dir: Path) -> dict | None:
    manifest_path = baseline_dir / "manifest.json"
    if not manifest_path.is_file():
        return None
    manifest = json.loads(manifest_path.read_text())
    if manifest.get("version") != BASELINE_VERSION:
        raise SystemExit(f"unsupported baseline version {manifest.get('version')!r} in {manifest_path}")
    return manifest


def page_filename(slug: str) -> str:
    return slug.replace("/", "__") + ".md"


def unified_diff(old: str, new: str, slug: str) -> tuple[str, bool]:
    lines = list(difflib.unified_diff(old.splitlines(), new.splitlines(), f"before/{slug}", f"after/{slug}", lineterm=""))
    truncated = len(lines) > MAX_DIFF_LINES
    return "\n".join(lines[:MAX_DIFF_LINES]), truncated


def compare(
    pages: list[Page],
    wrapped_paths: dict[str, list[str]],
    baseline: dict | None,
    baseline_dir: Path,
    issues: list[dict],
) -> tuple[dict, dict, dict[str, str]]:
    """Compute pending findings. Returns (work, new manifest, wrapped page texts by filename)."""
    markers, open_drift = reported_markers(issues)
    endpoints = [page for page in pages if page.key is not None]

    previous = {item["key"]: item for item in (baseline or {}).get("endpoints", [])}
    current_keys = {page.key for page in endpoints}

    manifest_entries = []
    baseline_pages: dict[str, str] = {}
    gaps_by_section: dict[str, list[dict]] = {}
    drift = []

    for page in sorted(endpoints, key=lambda p: (p.entry.section, p.key)):
        provider_files = wrapped_paths.get(normalize_path(page.path))
        summary = {
            "key": page.key,
            "method": page.method,
            "path": page.path,
            "section": page.entry.section,
            "title": page.entry.title,
            "url": page.entry.url,
        }
        manifest_entries.append({**summary, "slug": page.entry.slug, "hash": page.digest, "wrapped": bool(provider_files)})

        if not provider_files:
            if ("gap", page.key) not in markers:
                gaps_by_section.setdefault(page.entry.section, []).append(summary)
            continue

        filename = page_filename(page.entry.slug)
        baseline_pages[filename] = page.normalized
        before = previous.get(page.key)
        old_page = baseline_dir / "pages" / page_filename(before["slug"]) if before else None
        if not before or not before.get("wrapped") or before["hash"] == page.digest or not old_page.is_file():
            continue
        if ("drift", f"{page.key} {page.digest}") in markers:
            continue
        diff, truncated = unified_diff(old_page.read_text(), page.normalized, page.entry.slug)
        drift.append(
            {
                **summary,
                "provider_files": provider_files,
                "open_issue": open_drift.get(page.key),
                "diff": diff,
                "diff_truncated": truncated,
                "marker": marker("drift", f"{page.key} {page.digest}"),
            }
        )

    removed = [
        {
            "key": key,
            "method": item["method"],
            "path": item["path"],
            "section": item["section"],
            "title": item["title"],
            "url": item["url"],
            "was_wrapped": item["wrapped"],
            "marker": marker("removed", key),
        }
        for key, item in sorted(previous.items())
        if key not in current_keys and ("removed", key) not in markers
    ]

    gaps = [
        {
            "section": section,
            "title": f"API gap: {section}",
            "endpoints": items,
            "markers": "\n".join(marker("gap", item["key"]) for item in items),
        }
        for section, items in sorted(gaps_by_section.items())
    ]

    work = {
        "baseline": "present" if baseline else "missing",
        "counts": {
            "documented_endpoints": len(endpoints),
            "wrapped_endpoints": len(baseline_pages),
            "pending_removed": len(removed),
            "pending_drift": len(drift),
            "pending_gap_sections": len(gaps),
            "pending_gap_endpoints": sum(len(gap["endpoints"]) for gap in gaps),
        },
        "removed": removed,
        "drift": drift,
        "gaps": gaps,
    }
    manifest = {"version": BASELINE_VERSION, "index": DOCS_INDEX_URL, "endpoints": manifest_entries}
    return work, manifest, baseline_pages


def fetch_text(url: str) -> str:
    """GET a URL, retrying transient failures. Raises after the last attempt."""
    request = urllib.request.Request(url, headers={"User-Agent": "pulumi-webflow-api-parity"})
    for attempt in range(1, FETCH_ATTEMPTS + 1):
        try:
            with urllib.request.urlopen(request, timeout=FETCH_TIMEOUT_SECONDS) as response:
                return response.read().decode("utf-8")
        except (urllib.error.URLError, TimeoutError) as error:
            permanent = isinstance(error, urllib.error.HTTPError) and error.code < 500 and error.code != 429
            if permanent or attempt == FETCH_ATTEMPTS:
                raise RuntimeError(f"fetching {url} failed after {attempt} attempt(s): {error}") from error
            time.sleep(2**attempt)
    raise AssertionError("unreachable")


def fetch_pages(entries: list[IndexEntry], fetch) -> list[Page]:
    with concurrent.futures.ThreadPoolExecutor(max_workers=FETCH_WORKERS) as pool:
        texts = list(pool.map(lambda entry: fetch(entry.url), entries))
    return [parse_page(entry, text) for entry, text in zip(entries, texts)]


def write_outputs(out_dir: Path, work: dict, manifest: dict, baseline_pages: dict[str, str]) -> None:
    baseline_out = out_dir / "baseline"
    (baseline_out / "pages").mkdir(parents=True, exist_ok=True)
    (out_dir / "work.json").write_text(json.dumps(work, indent=2) + "\n")
    (baseline_out / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    for filename, text in baseline_pages.items():
        (baseline_out / "pages" / filename).write_text(text)


def has_work(work: dict) -> bool:
    return bool(work["removed"] or work["drift"] or work["gaps"])


def run(
    provider_dir: Path,
    baseline_dir: Path,
    issues_path: Path | None,
    out_dir: Path,
    fetch=fetch_text,
    min_index_entries: int = MIN_INDEX_ENTRIES,
) -> dict:
    entries = parse_index(fetch(DOCS_INDEX_URL))
    if len(entries) < min_index_entries:
        raise RuntimeError(f"docs index listed only {len(entries)} reference pages; expected at least {min_index_entries}")
    pages = fetch_pages(entries, fetch)
    issues = json.loads(issues_path.read_text()) if issues_path else []
    work, manifest, baseline_pages = compare(pages, provider_paths(provider_dir), load_baseline(baseline_dir), baseline_dir, issues)
    write_outputs(out_dir, work, manifest, baseline_pages)
    return work


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--provider-dir", type=Path, required=True)
    parser.add_argument("--baseline-dir", type=Path, required=True, help="previous baseline; missing or empty on the first run")
    parser.add_argument("--issues", type=Path, help="JSON array of prior issues with number, state and body")
    parser.add_argument("--out-dir", type=Path, required=True)
    args = parser.parse_args()

    work = run(args.provider_dir, args.baseline_dir, args.issues, args.out_dir)
    print(json.dumps(work["counts"]))
    pending = has_work(work)
    if github_output := os.environ.get("GITHUB_OUTPUT"):
        with open(github_output, "a") as output:
            output.write(f"has_work={'true' if pending else 'false'}\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
