#!/usr/bin/env python3
"""Compare the Webflow Data API reference with the endpoints the provider calls.

Writes work.json (findings not yet published in an issue or comment) and a new
baseline directory for the api-parity-maintainer agentic workflow. See README.md.
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
# Matches any method: the URL template is built in a function that sends no request.
ANY_METHOD = "*"

_INDEX_LINE = re.compile(
    r"^- API Reference > (?P<section>.+?) \[(?P<title>.+)\]"
    r"\((?P<url>https://developers\.webflow\.com/data/v2\.0\.0/reference/(?P<slug>[^)\s]+)\.md)\)\s*$"
)
_ENDPOINT_LINE = re.compile(r"^(GET|POST|PUT|PATCH|DELETE) https://api\.webflow\.com/v2(/\S*)\s*$")
_PATH_PARAM = re.compile(r"\{[^}]+\}")
_GO_FUNC = re.compile(r"^func ", re.MULTILINE)
_GO_LINE_COMMENT = re.compile(r"^\s*//.*$", re.MULTILINE)
_GO_V2_LITERAL = re.compile(r'"[^"\n]*?/v2(/[^"\n]*)"')
_GO_VERB = re.compile(r"%[sdv]")
_GO_REQUEST_METHOD = re.compile(
    r'NewRequest(?:WithContext)?\([^,]*,\s*(?:"(GET|POST|PUT|PATCH|DELETE)"|http\.Method(Get|Post|Put|Patch|Delete)\b)'
)
# Markers are inline code spans: gh-aw's output sanitizer removes HTML comments.
_MARKER = re.compile(r"`" + re.escape(MARKER_PREFIX) + r"(gap|drift|removed) ([^`\n]+?)`")
_CONTRACT_FIELD = re.compile(r"^\s*- `[^`]+` \(")
_CONTRACT_SCOPE = re.compile(r"Required scope \| `([^`]+)`")


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
        """Hash of the documented contract only, so wording edits are not drift."""
        return hashlib.sha256(contract(self.normalized).encode()).hexdigest()[:16]


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
    """Keep the documented contract with its prose: title, endpoint, description, request, response, types.

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


def contract(normalized: str) -> str:
    """Reduce a normalized page to the lines that define the API contract.

    Keeps the endpoint, content type, section and status headings, each
    parameter or field with its type and required-ness, allowed values, and
    required scopes. Drops descriptions and other prose.
    """
    lines = []
    for line in normalized.splitlines():
        if _ENDPOINT_LINE.match(line) or line.startswith("Content-Type:") or line.startswith(("## ", "### ")):
            lines.append(line)
        elif _CONTRACT_FIELD.match(line):
            lines.append(line.split(" — ", 1)[0])
        elif line.lstrip().startswith("- Allowed values:"):
            lines.append(line)
        lines.extend(f"scope {scope}" for scope in _CONTRACT_SCOPE.findall(line))
    return "\n".join(lines) + "\n"


def parse_page(entry: IndexEntry, text: str) -> Page:
    method = path = None
    for line in text.splitlines():
        match = _ENDPOINT_LINE.match(line)
        if match:
            method, path = match[1], match[2]
            break
    return Page(entry, method, path, normalize_page(text))


def provider_endpoints(provider_dir: Path) -> dict[str, list[str]]:
    """Map each endpoint the provider calls, as "METHOD /normalized/path", to the files that call it.

    The method comes from the http.NewRequest call in the same function as the
    URL template. A template in a function without a request matches any method.
    """
    found: dict[str, set[str]] = {}
    for go_file in sorted(provider_dir.glob("*.go")):
        if go_file.name.endswith("_test.go"):
            continue
        source = _GO_LINE_COMMENT.sub("", go_file.read_text())
        for chunk in _GO_FUNC.split(source):
            paths = {normalize_path(_GO_VERB.sub("{}", literal)) for literal in _GO_V2_LITERAL.findall(chunk)}
            methods = {(literal or constant).upper() for literal, constant in _GO_REQUEST_METHOD.findall(chunk)}
            for path in paths:
                for method in methods or {ANY_METHOD}:
                    found.setdefault(f"{method} {path}", set()).add(go_file.name)
    return {key: sorted(files) for key, files in sorted(found.items())}


def wrapping_files(page: Page, endpoints: dict[str, list[str]]) -> list[str]:
    path = normalize_path(page.path)
    return sorted(set(endpoints.get(f"{page.method} {path}", []) + endpoints.get(f"{ANY_METHOD} {path}", [])))


def published_markers(issues: list[dict], comments: list[dict]) -> tuple[set[tuple[str, str]], dict[str, int]]:
    """Collect markers from prior issues and their comments, in any state.

    Returns every (kind, value) marker, and for drift the open issue number per
    endpoint key so a new change can be added there as a comment.
    """
    state_by_number = {issue["number"]: issue.get("state", "").upper() for issue in issues}
    sources = [(issue["number"], issue.get("body")) for issue in issues]
    sources += [(comment["issue_number"], comment.get("body")) for comment in comments]
    markers: set[tuple[str, str]] = set()
    open_drift: dict[str, int] = {}
    for number, body in sources:
        for kind, value in _MARKER.findall(body or ""):
            markers.add((kind, value))
            if kind == "drift" and state_by_number.get(number) == "OPEN":
                open_drift.setdefault(value.rsplit(" ", 1)[0], number)
    return markers, open_drift


def marker(kind: str, value: str) -> str:
    return f"`{MARKER_PREFIX}{kind} {value}`"


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
    endpoints_called: dict[str, list[str]],
    baseline: dict | None,
    baseline_dir: Path,
    issues: list[dict],
    comments: list[dict] | None = None,
) -> tuple[dict, dict, dict[str, str]]:
    """Compute pending findings and the baseline to save.

    A drift or removal stays pending, and the saved baseline keeps its previous
    state, until an issue or comment publishes its marker. A run that publishes
    nothing (a preview, a failure, or an exhausted output budget) therefore loses
    nothing, and the next run reports it again.

    Returns (work, manifest to save, page texts to save by filename).
    """
    markers, open_drift = published_markers(issues, comments or [])
    endpoints = [page for page in pages if page.key is not None]

    previous = {item["key"]: item for item in (baseline or {}).get("endpoints", [])}
    current_keys = {page.key for page in endpoints}

    manifest_entries = []
    baseline_pages: dict[str, str] = {}
    gaps_by_section: dict[str, list[dict]] = {}
    drift = []

    for page in sorted(endpoints, key=lambda p: (p.entry.section, p.key)):
        provider_files = wrapping_files(page, endpoints_called)
        summary = {
            "key": page.key,
            "method": page.method,
            "path": page.path,
            "section": page.entry.section,
            "title": page.entry.title,
            "url": page.entry.url,
        }
        current_entry = {**summary, "slug": page.entry.slug, "hash": page.digest, "wrapped": bool(provider_files)}

        if not provider_files:
            manifest_entries.append(current_entry)
            if ("gap", page.key) not in markers:
                gaps_by_section.setdefault(page.entry.section, []).append(summary)
            continue

        before = previous.get(page.key)
        old_page = baseline_dir / "pages" / page_filename(before["slug"]) if before else None
        changed = bool(before and before.get("wrapped") and before["hash"] != page.digest and old_page.is_file())
        if not changed or ("drift", f"{page.key} {page.digest}") in markers:
            manifest_entries.append(current_entry)
            baseline_pages[page_filename(page.entry.slug)] = page.normalized
            continue

        old_text = old_page.read_text()
        manifest_entries.append(before)
        baseline_pages[page_filename(before["slug"])] = old_text
        diff, truncated = unified_diff(old_text, page.normalized, page.entry.slug)
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

    removed = []
    for key, item in sorted(previous.items()):
        if key in current_keys or ("removed", key) in markers:
            continue
        manifest_entries.append(item)
        old_page = baseline_dir / "pages" / page_filename(item["slug"])
        if item.get("wrapped") and old_page.is_file():
            baseline_pages[page_filename(item["slug"])] = old_page.read_text()
        removed.append(
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
        )

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
            "wrapped_endpoints": sum(1 for page in endpoints if wrapping_files(page, endpoints_called)),
            "pending_removed": len(removed),
            "pending_drift": len(drift),
            "pending_gap_sections": len(gaps),
            "pending_gap_endpoints": sum(len(gap["endpoints"]) for gap in gaps),
        },
        "removed": removed,
        "drift": drift,
        "gaps": gaps,
    }
    manifest = {
        "version": BASELINE_VERSION,
        "index": DOCS_INDEX_URL,
        "endpoints": sorted(manifest_entries, key=lambda item: (item["section"], item["key"])),
    }
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
    comments_path: Path | None = None,
) -> dict:
    entries = parse_index(fetch(DOCS_INDEX_URL))
    if len(entries) < min_index_entries:
        raise RuntimeError(f"docs index listed only {len(entries)} reference pages; expected at least {min_index_entries}")
    pages = fetch_pages(entries, fetch)
    issues = json.loads(issues_path.read_text()) if issues_path else []
    comments = json.loads(comments_path.read_text()) if comments_path else []
    work, manifest, baseline_pages = compare(
        pages, provider_endpoints(provider_dir), load_baseline(baseline_dir), baseline_dir, issues, comments
    )
    write_outputs(out_dir, work, manifest, baseline_pages)
    return work


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--provider-dir", type=Path, required=True)
    parser.add_argument("--baseline-dir", type=Path, required=True, help="previous baseline; missing or empty on the first run")
    parser.add_argument("--issues", type=Path, help="JSON array of prior issues with number, state and body")
    parser.add_argument("--comments", type=Path, help="JSON array of issue comments with issue_number and body")
    parser.add_argument("--out-dir", type=Path, required=True)
    args = parser.parse_args()

    work = run(args.provider_dir, args.baseline_dir, args.issues, args.out_dir, comments_path=args.comments)
    print(json.dumps(work["counts"]))
    pending = has_work(work)
    if github_output := os.environ.get("GITHUB_OUTPUT"):
        with open(github_output, "a") as output:
            output.write(f"has_work={'true' if pending else 'false'}\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
