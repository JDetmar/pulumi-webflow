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
import http.client
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
# The index listed ~160 reference pages and ~150 endpoints in October 2026.
# Far fewer means the docs site served something unexpected, and reporting
# every missing endpoint as removed would flood the issue tracker.
MIN_INDEX_ENTRIES = 100
# Webflow removing a whole product area (User Accounts, January 2026) dropped
# 6 of ~150 endpoints. Losing more than this share in one run is more likely a
# partial index than real removals. A few removals never trip it.
MAX_REMOVED_FRACTION = 0.1
MAX_REMOVED_ALWAYS_ALLOWED = 5
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
_MARKER = re.compile(r"`" + re.escape(MARKER_PREFIX) + r"(gap|drift|removed|undocumented) ([^`\n]+?)`")
_CONTRACT_FIELD = re.compile(r"^\s*- `[^`]+` \(")
_CONTRACT_SCOPE = re.compile(r"Required scope \| `([^`]+)`")


@dataclasses.dataclass(frozen=True)
class IndexEntry:
    section: str
    title: str
    url: str
    slug: str

    @property
    def page_url(self) -> str:
        """The reference page for people; `url` is its Markdown source."""
        return self.url.removesuffix(".md")


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


def reference_file(slug: str) -> str:
    """Where write_outputs puts a page's normalized text, relative to work.json."""
    return f"reference/{page_filename(slug)}"


def reference_files(pages: list[Page]) -> dict[str, str]:
    """Every documented page's normalized text, plus an index, for the agent to read offline."""
    endpoints = sorted((page for page in pages if page.key is not None), key=lambda p: (p.entry.section, p.key))
    files = {reference_file(page.entry.slug): page.normalized for page in endpoints}
    lines = [f"- `{page.method} {page.path}` | {page.entry.section} | {page.entry.title} | {reference_file(page.entry.slug)}" for page in endpoints]
    files["reference/index.md"] = "# Documented endpoints\n\n" + "\n".join(lines) + "\n"
    return files


def unified_diff(old: str, new: str, slug: str) -> tuple[str, bool]:
    lines = list(difflib.unified_diff(old.splitlines(), new.splitlines(), f"before/{slug}", f"after/{slug}", lineterm=""))
    truncated = len(lines) > MAX_DIFF_LINES
    return "\n".join(lines[:MAX_DIFF_LINES]), truncated


class SnapshotError(RuntimeError):
    """The reference served content that cannot be compared safely; fail before reporting anything."""


def check_pages(pages: list[Page], baseline: dict | None) -> None:
    """Refuse to compare when the fetched pages look wrong, instead of reporting false findings.

    A page the baseline knew as an endpoint that now has no endpoint line is
    most likely an error page or a format change, not a removal, and so is a
    large share of endpoints vanishing at once. Two pages with the same
    endpoint would make the baseline compare the wrong pages.
    """
    keys: dict[str, str] = {}
    duplicates = []
    for page in pages:
        if page.key is None:
            continue
        if page.key in keys:
            duplicates.append(f"{page.key} ({keys[page.key]}, {page.entry.slug})")
        keys[page.key] = page.entry.slug
    if duplicates:
        raise SnapshotError(f"endpoints documented on more than one page: {'; '.join(duplicates)}")

    known_endpoint_slugs = {item["slug"] for item in (baseline or {}).get("endpoints", [])}
    unparsed = sorted(page.entry.slug for page in pages if page.key is None and page.entry.slug in known_endpoint_slugs)
    if unparsed:
        raise SnapshotError(f"pages that documented an endpoint no longer have an endpoint line: {', '.join(unparsed)}")

    baseline_keys = {item["key"] for item in (baseline or {}).get("endpoints", [])}
    vanished = baseline_keys - set(keys)
    if len(vanished) > max(MAX_REMOVED_ALWAYS_ALLOWED, MAX_REMOVED_FRACTION * len(baseline_keys)):
        raise SnapshotError(
            f"{len(vanished)} of {len(baseline_keys)} baseline endpoints are missing from the reference; "
            "the index is probably incomplete"
        )


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
    Raises SnapshotError when the pages cannot be compared safely.
    """
    check_pages(pages, baseline)
    markers, open_drift = published_markers(issues, comments or [])
    endpoints = [page for page in pages if page.key is not None]

    previous = {item["key"]: item for item in (baseline or {}).get("endpoints", [])}
    current_keys = {page.key for page in endpoints}

    manifest_entries = []
    baseline_pages: dict[str, str] = {}
    gaps_by_section: dict[str, list[dict]] = {}
    section_files: dict[str, set[str]] = {}
    drift = []

    for page in sorted(endpoints, key=lambda p: (p.entry.section, p.key)):
        provider_files = wrapping_files(page, endpoints_called)
        summary = {
            "key": page.key,
            "method": page.method,
            "path": page.path,
            "section": page.entry.section,
            "title": page.entry.title,
            "url": page.entry.page_url,
        }
        current_entry = {**summary, "slug": page.entry.slug, "hash": page.digest, "wrapped": bool(provider_files)}
        section_files.setdefault(page.entry.section, set()).update(provider_files)

        if not provider_files:
            manifest_entries.append(current_entry)
            if ("gap", page.key) not in markers:
                gaps_by_section.setdefault(page.entry.section, []).append(
                    {**summary, "page_file": reference_file(page.entry.slug)}
                )
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
                "page_file": reference_file(page.entry.slug),
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
                "url": item["url"].removesuffix(".md"),
                "was_wrapped": item["wrapped"],
                "marker": marker("removed", key),
            }
        )

    gaps = [
        {
            "section": section,
            "title": f"API gap: {section}",
            "provider_files": sorted(section_files[section]),
            "endpoints": items,
            "markers": "\n".join(marker("gap", item["key"]) for item in items),
        }
        for section, items in gaps_by_section.items()
    ]
    # A gap next to a resource the provider already has comes first: it is
    # cheaper to close than a new area, and its issue can be specific.
    gaps.sort(key=lambda gap: (not gap["provider_files"], gap["section"]))

    # Calls the reference does not document were never in a baseline (or would
    # be pending removals), so the removal check cannot catch them.
    pending_removed = {item["key"] for item in removed}
    undocumented_by_files: dict[tuple[str, ...], list[dict]] = {}
    for key, files in endpoints_called.items():
        method, path = key.split(" ", 1)
        if method == ANY_METHOD or key in current_keys or key in pending_removed:
            continue
        if ("undocumented", key) in markers or ("removed", key) in markers:
            continue
        undocumented_by_files.setdefault(tuple(files), []).append({"key": key, "method": method, "path": path})
    undocumented = [
        {
            "title": f"API undocumented: {', '.join(files)}",
            "provider_files": list(files),
            "endpoints": items,
            "markers": "\n".join(marker("undocumented", item["key"]) for item in items),
        }
        for files, items in sorted(undocumented_by_files.items())
    ]

    work = {
        "baseline": "present" if baseline else "missing",
        "counts": {
            "documented_endpoints": len(endpoints),
            "wrapped_endpoints": sum(1 for page in endpoints if wrapping_files(page, endpoints_called)),
            "pending_removed": len(removed),
            "pending_drift": len(drift),
            "pending_undocumented": sum(len(item["endpoints"]) for item in undocumented),
            "pending_gap_sections": len(gaps),
            "pending_gap_endpoints": sum(len(gap["endpoints"]) for gap in gaps),
        },
        "removed": removed,
        "drift": drift,
        "undocumented": undocumented,
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
        # OSError covers URLError, timeouts, and resets; HTTPException covers a
        # body cut short (IncompleteRead), which the CDN does under load.
        except (OSError, http.client.HTTPException) as error:
            permanent = isinstance(error, urllib.error.HTTPError) and error.code < 500 and error.code != 429
            if permanent or attempt == FETCH_ATTEMPTS:
                raise RuntimeError(f"fetching {url} failed after {attempt} attempt(s): {error}") from error
            time.sleep(2**attempt)
    raise AssertionError("unreachable")


def fetch_pages(entries: list[IndexEntry], fetch) -> list[Page]:
    with concurrent.futures.ThreadPoolExecutor(max_workers=FETCH_WORKERS) as pool:
        texts = list(pool.map(lambda entry: fetch(entry.url), entries))
    return [parse_page(entry, text) for entry, text in zip(entries, texts)]


def write_outputs(
    out_dir: Path, work: dict, manifest: dict, baseline_pages: dict[str, str], reference: dict[str, str] | None = None
) -> None:
    baseline_out = out_dir / "baseline"
    (baseline_out / "pages").mkdir(parents=True, exist_ok=True)
    (out_dir / "work.json").write_text(json.dumps(work, indent=2) + "\n")
    (baseline_out / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    for filename, text in baseline_pages.items():
        (baseline_out / "pages" / filename).write_text(text)
    for relative, text in (reference or {}).items():
        (out_dir / relative).parent.mkdir(parents=True, exist_ok=True)
        (out_dir / relative).write_text(text)


def has_work(work: dict) -> bool:
    return bool(work["removed"] or work["drift"] or work["undocumented"] or work["gaps"])


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
        raise SnapshotError(f"docs index listed only {len(entries)} reference pages; expected at least {min_index_entries}")
    pages = fetch_pages(entries, fetch)
    endpoint_count = sum(1 for page in pages if page.key is not None)
    if endpoint_count < min_index_entries:
        raise SnapshotError(f"reference pages documented only {endpoint_count} endpoints; expected at least {min_index_entries}")
    issues = json.loads(issues_path.read_text()) if issues_path else []
    comments = json.loads(comments_path.read_text()) if comments_path else []
    work, manifest, baseline_pages = compare(
        pages, provider_endpoints(provider_dir), load_baseline(baseline_dir), baseline_dir, issues, comments
    )
    write_outputs(out_dir, work, manifest, baseline_pages, reference_files(pages))
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
