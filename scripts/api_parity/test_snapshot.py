import json
import tempfile
import unittest
from pathlib import Path

import snapshot

INDEX = """\
# Webflow Data API

- [Working with AEO Analytics](https://developers.webflow.com/data/v2.0.0/docs/working-with-aeo-analytics.md): Guide.
- API Reference > Sites [Get Site](https://developers.webflow.com/data/v2.0.0/reference/sites/get.md)
- API Reference > Enterprise > Site Configuration > 301 Redirects [List URL redirects](https://developers.webflow.com/data/v2.0.0/reference/enterprise/site-configuration/301-redirects/get.md)
- API Reference > AEO Analytics [Get Bot Traffic](https://developers.webflow.com/data/v2.0.0/reference/aeo-analytics/get-bot-traffic.md)
- API Reference > Enterprise > Site Access [Revoke Site Access (by email)](https://developers.webflow.com/data/v2.0.0/reference/enterprise/site-access/remove-member-by-email.md)
- API Reference > Webhooks > Events [Site publish](https://developers.webflow.com/data/v2.0.0/reference/webhooks/events/site-publish.md)
"""

PAGE_TEMPLATE = """\
> This page is for Data API, version v2 (default).
> For a complete documentation index, see https://developers.webflow.com/llms.txt.

# {title}

{method} https://api.webflow.com/v2{path}

{description}

## Request

### Path parameters

- `site_id` (string, required) — Unique identifier for a Site

## Response

### 200

{response}

## Errors

### 401 Unauthorized Error

```json
{{"message": "Request not authorized"}}
```

## Examples

```shell
curl https://api.webflow.com/v2{path}
```
"""


def page_text(title, method, path, description="Description.", response="- `id` (string, required)"):
    return PAGE_TEMPLATE.format(title=title, method=method, path=path, description=description, response=response)


def entry(slug, section="Sites", title="Get Site"):
    return snapshot.IndexEntry(section, title, f"https://developers.webflow.com/data/v2.0.0/reference/{slug}.md", slug)


class ParseIndexTest(unittest.TestCase):
    def test_reads_reference_entries_and_ignores_guides(self):
        entries = snapshot.parse_index(INDEX)

        self.assertEqual(
            [(e.section, e.title, e.slug) for e in entries],
            [
                ("Sites", "Get Site", "sites/get"),
                ("Enterprise > Site Configuration > 301 Redirects", "List URL redirects", "enterprise/site-configuration/301-redirects/get"),
                ("AEO Analytics", "Get Bot Traffic", "aeo-analytics/get-bot-traffic"),
                ("Enterprise > Site Access", "Revoke Site Access (by email)", "enterprise/site-access/remove-member-by-email"),
                ("Webhooks > Events", "Site publish", "webhooks/events/site-publish"),
            ],
        )


class ParsePageTest(unittest.TestCase):
    def test_extracts_method_and_path(self):
        page = snapshot.parse_page(entry("sites/get"), page_text("Get Site", "GET", "/sites/{site_id}"))

        self.assertEqual((page.method, page.path, page.key), ("GET", "/sites/{site_id}", "GET /sites/{}"))

    def test_event_page_without_endpoint_has_no_key(self):
        page = snapshot.parse_page(entry("webhooks/events/site-publish"), "# Site publish\n\nPayload sent when a site publishes.\n")

        self.assertIsNone(page.key)

    def test_normalization_drops_preamble_errors_examples_and_code(self):
        normalized = snapshot.normalize_page(page_text("Get Site", "GET", "/sites/{site_id}"))

        self.assertTrue(normalized.startswith("# Get Site\n"))
        self.assertIn("## Response", normalized)
        for dropped in ("This page is for Data API", "## Errors", "Unauthorized", "## Examples", "curl"):
            self.assertNotIn(dropped, normalized)

    def test_hash_ignores_changes_outside_the_contract(self):
        before = snapshot.parse_page(entry("sites/get"), page_text("Get Site", "GET", "/sites/{site_id}"))
        after_text = page_text("Get Site", "GET", "/sites/{site_id}").replace("curl https", "wget https").replace(
            "Request not authorized", "Not authorized"
        )
        after = snapshot.parse_page(entry("sites/get"), after_text)

        self.assertEqual(before.digest, after.digest)


class ProviderPathsTest(unittest.TestCase):
    def test_collects_v2_templates_and_skips_tests(self):
        with tempfile.TemporaryDirectory() as tmp:
            provider = Path(tmp)
            (provider / "redirect.go").write_text(
                'url := fmt.Sprintf("%s/v2/sites/%s/redirects/%s", base, siteID, id)\n'
                'list := fmt.Sprintf("%s/v2/sites/%s/redirects?offset=%d", base, siteID, offset)\n'
            )
            (provider / "redirect_test.go").write_text('url := "/v2/sites/5f0c8c9e1c9d440000e8d8c3/webhooks"\n')

            paths = snapshot.provider_paths(provider)

        self.assertEqual(paths, {"/sites/{}/redirects": ["redirect.go"], "/sites/{}/redirects/{}": ["redirect.go"]})


class CompareTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.baseline_dir = self.root / "baseline"
        self.wrapped = {"/sites/{}": ["site.go"]}

    def tearDown(self):
        self.tmp.cleanup()

    def pages(self, site_response="- `id` (string, required)", include_bot_traffic=True):
        pages = [snapshot.parse_page(entry("sites/get"), page_text("Get Site", "GET", "/sites/{site_id}", response=site_response))]
        if include_bot_traffic:
            pages.append(
                snapshot.parse_page(
                    entry("aeo-analytics/get-bot-traffic", "AEO Analytics", "Get Bot Traffic"),
                    page_text("Get Bot Traffic", "GET", "/sites/{site_id}/aeo/reports/bot_traffic"),
                )
            )
        return pages

    def save_baseline(self, pages):
        _, manifest, baseline_pages = snapshot.compare(pages, self.wrapped, None, self.baseline_dir, [])
        snapshot.write_outputs(self.root, {"removed": [], "drift": [], "gaps": []}, manifest, baseline_pages)
        return snapshot.load_baseline(self.baseline_dir)

    def test_first_run_reports_gaps_grouped_by_section_and_no_drift(self):
        work, manifest, baseline_pages = snapshot.compare(self.pages(), self.wrapped, None, self.baseline_dir, [])

        self.assertEqual(work["baseline"], "missing")
        self.assertEqual([gap["title"] for gap in work["gaps"]], ["API gap: AEO Analytics"])
        self.assertEqual(
            work["gaps"][0]["markers"], "<!-- api-parity:gap GET /sites/{}/aeo/reports/bot_traffic -->"
        )
        self.assertEqual(work["drift"], [])
        self.assertEqual(list(baseline_pages), ["sites__get.md"])
        self.assertEqual(
            {e["key"]: e["wrapped"] for e in manifest["endpoints"]},
            {"GET /sites/{}": True, "GET /sites/{}/aeo/reports/bot_traffic": False},
        )

    def test_changed_wrapped_page_is_drift_with_a_diff(self):
        baseline = self.save_baseline(self.pages())

        work, _, _ = snapshot.compare(
            self.pages(site_response="- `id` (string, required)\n- `locale` (string, optional)"),
            self.wrapped,
            baseline,
            self.baseline_dir,
            [],
        )

        [drift] = work["drift"]
        self.assertEqual(drift["key"], "GET /sites/{}")
        self.assertEqual(drift["provider_files"], ["site.go"])
        self.assertIn("+- `locale` (string, optional)", drift["diff"])
        self.assertIsNone(drift["open_issue"])

    def test_unchanged_pages_produce_no_work(self):
        baseline = self.save_baseline(self.pages())
        issues = [{"number": 7, "state": "OPEN", "body": "<!-- api-parity:gap GET /sites/{}/aeo/reports/bot_traffic -->"}]

        work, _, _ = snapshot.compare(self.pages(), self.wrapped, baseline, self.baseline_dir, issues)

        self.assertFalse(snapshot.has_work(work))

    def test_gap_already_in_a_closed_issue_is_not_reported_again(self):
        issues = [{"number": 3, "state": "CLOSED", "body": "Not planned.\n<!-- api-parity:gap GET /sites/{}/aeo/reports/bot_traffic -->"}]

        work, _, _ = snapshot.compare(self.pages(), self.wrapped, None, self.baseline_dir, issues)

        self.assertEqual(work["gaps"], [])

    def test_new_drift_on_endpoint_with_open_drift_issue_points_at_that_issue(self):
        baseline = self.save_baseline(self.pages())
        issues = [{"number": 12, "state": "OPEN", "body": "<!-- api-parity:drift GET /sites/{} oldhash -->"}]

        work, _, _ = snapshot.compare(self.pages(site_response="- `name` (string)"), self.wrapped, baseline, self.baseline_dir, issues)

        self.assertEqual(work["drift"][0]["open_issue"], 12)

    def test_drift_already_reported_for_the_same_content_is_skipped(self):
        baseline = self.save_baseline(self.pages())
        changed = self.pages(site_response="- `name` (string)")
        issues = [{"number": 12, "state": "CLOSED", "body": f"<!-- api-parity:drift GET /sites/{{}} {changed[0].digest} -->"}]

        work, _, _ = snapshot.compare(changed, self.wrapped, baseline, self.baseline_dir, issues)

        self.assertEqual(work["drift"], [])

    def test_endpoint_missing_from_index_is_reported_as_removed(self):
        baseline = self.save_baseline(self.pages())

        work, _, _ = snapshot.compare(self.pages(include_bot_traffic=False), self.wrapped, baseline, self.baseline_dir, [])

        self.assertEqual(
            [(r["key"], r["was_wrapped"]) for r in work["removed"]], [("GET /sites/{}/aeo/reports/bot_traffic", False)]
        )


class RunTest(unittest.TestCase):
    def test_rejects_a_truncated_index_before_fetching_pages(self):
        fetched = []

        def fetch(url):
            fetched.append(url)
            return INDEX

        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaisesRegex(RuntimeError, "listed only 5 reference pages"):
                snapshot.run(Path(tmp), Path(tmp) / "baseline", None, Path(tmp) / "out", fetch=fetch)

        self.assertEqual(fetched, [snapshot.DOCS_INDEX_URL])

    def test_writes_work_and_baseline_from_fetched_pages(self):
        pages = {
            "sites/get": page_text("Get Site", "GET", "/sites/{site_id}"),
            "enterprise/site-configuration/301-redirects/get": page_text("List URL redirects", "GET", "/sites/{site_id}/redirects"),
            "aeo-analytics/get-bot-traffic": page_text("Get Bot Traffic", "GET", "/sites/{site_id}/aeo/reports/bot_traffic"),
            "enterprise/site-access/remove-member-by-email": page_text("Revoke Site Access", "DELETE", "/sites/{site_id}/access"),
            "webhooks/events/site-publish": "# Site publish\n\nPayload.\n",
        }

        def fetch(url):
            if url == snapshot.DOCS_INDEX_URL:
                return INDEX
            return pages[url.split("/reference/", 1)[1].removesuffix(".md")]

        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "provider").mkdir()
            (root / "provider" / "site.go").write_text('u := fmt.Sprintf("%s/v2/sites/%s", base, id)\n')
            work = snapshot.run(root / "provider", root / "baseline", None, root / "out", fetch=fetch, min_index_entries=1)

            written = json.loads((root / "out" / "work.json").read_text())
            baseline_files = sorted(p.name for p in (root / "out" / "baseline" / "pages").iterdir())

        self.assertEqual(written, work)
        self.assertEqual(work["counts"]["documented_endpoints"], 4)
        self.assertEqual(work["counts"]["wrapped_endpoints"], 1)
        self.assertEqual(baseline_files, ["sites__get.md"])


if __name__ == "__main__":
    unittest.main()
