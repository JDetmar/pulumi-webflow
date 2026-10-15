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

{description}Required scope | `sites:read`

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

SITE_RESPONSE = "- `id` (string, required) — Unique identifier for the Site"


def page_text(title, method, path, description="Description.", response=SITE_RESPONSE):
    return PAGE_TEMPLATE.format(title=title, method=method, path=path, description=description, response=response)


def entry(slug, section="Sites", title="Get Site"):
    return snapshot.IndexEntry(section, title, f"https://developers.webflow.com/data/v2.0.0/reference/{slug}.md", slug)


def site_page(**kwargs):
    return snapshot.parse_page(entry("sites/get"), page_text("Get Site", "GET", "/sites/{site_id}", **kwargs))


def bot_traffic_page():
    return snapshot.parse_page(
        entry("aeo-analytics/get-bot-traffic", "AEO Analytics", "Get Bot Traffic"),
        page_text("Get Bot Traffic", "GET", "/sites/{site_id}/aeo/reports/bot_traffic"),
    )


def issue(number, state, body):
    return {"number": number, "state": state, "body": body}


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
        page = site_page()

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

    def test_contract_keeps_fields_types_and_scopes_but_not_prose(self):
        lines = snapshot.contract(site_page().normalized).splitlines()

        self.assertIn("GET https://api.webflow.com/v2/sites/{site_id}", lines)
        self.assertIn("- `site_id` (string, required)", lines)
        self.assertIn("scope sites:read", lines)
        self.assertNotIn("Description.", "\n".join(lines))
        self.assertNotIn("Unique identifier", "\n".join(lines))

    def test_wording_edits_do_not_change_the_digest(self):
        reworded = site_page(
            description="A clearer description.", response="- `id` (string, required) — The Site's unique identifier"
        )

        self.assertEqual(site_page().digest, reworded.digest)

    def test_field_changes_change_the_digest(self):
        optional = site_page(response="- `id` (string, optional) — Unique identifier for the Site")
        added = site_page(response=SITE_RESPONSE + "\n- `locale` (string, optional) — Default locale")

        self.assertNotEqual(site_page().digest, optional.digest)
        self.assertNotEqual(site_page().digest, added.digest)


class ProviderEndpointsTest(unittest.TestCase):
    def provider(self, files):
        tmp = tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        for name, text in files.items():
            (Path(tmp.name) / name).write_text(text)
        return snapshot.provider_endpoints(Path(tmp.name))

    def test_pairs_each_url_template_with_its_request_method(self):
        endpoints = self.provider(
            {
                "redirect.go": (
                    "func GetRedirects() {\n"
                    '\turl := fmt.Sprintf("%s/v2/sites/%s/redirects?offset=%d", base, siteID, offset)\n'
                    '\treq, _ := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)\n'
                    "}\n"
                    "func DeleteRedirect() {\n"
                    '\turl := fmt.Sprintf("%s/v2/sites/%s/redirects/%s", base, siteID, id)\n'
                    "\treq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)\n"
                    "}\n"
                ),
                "redirect_test.go": 'func TestX() { u := "/v2/sites/5f0c8c9e1c9d440000e8d8c3/webhooks" }\n',
            }
        )

        self.assertEqual(
            endpoints, {"DELETE /sites/{}/redirects/{}": ["redirect.go"], "GET /sites/{}/redirects": ["redirect.go"]}
        )

    def test_template_without_a_request_matches_any_method_and_comments_are_ignored(self):
        endpoints = self.provider(
            {
                "urls.go": (
                    '// Example: "https://api.webflow.com/v2/sites/{id}/pages"\n'
                    'func siteURL(id string) string { return fmt.Sprintf("%s/v2/sites/%s", base, id) }\n'
                )
            }
        )

        self.assertEqual(endpoints, {"* /sites/{}": ["urls.go"]})


class CompareTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.baseline_dir = self.root / "baseline"
        self.saves = 0
        self.called = {"GET /sites/{}": ["site.go"]}

    def tearDown(self):
        self.tmp.cleanup()

    def compare(self, pages, baseline=None, issues=(), comments=(), called=None):
        return snapshot.compare(
            pages, self.called if called is None else called, baseline, self.baseline_dir, list(issues), list(comments)
        )

    def save(self, pages, baseline=None, issues=(), comments=()):
        """Run compare and save its baseline, as the save_baseline job does after every snapshot."""
        work, manifest, baseline_pages = self.compare(pages, baseline, issues, comments)
        self.saves += 1
        out = self.root / f"out-{self.saves}"
        snapshot.write_outputs(out, work, manifest, baseline_pages)
        self.baseline_dir = out / "baseline"
        return work, snapshot.load_baseline(self.baseline_dir)

    def changed_site_page(self):
        return site_page(response=SITE_RESPONSE + "\n- `locale` (string, optional) — Default locale")

    def test_first_run_reports_gaps_grouped_by_section_with_code_span_markers(self):
        work, manifest, baseline_pages = self.compare([site_page(), bot_traffic_page()])

        self.assertEqual(work["baseline"], "missing")
        self.assertEqual([gap["title"] for gap in work["gaps"]], ["API gap: AEO Analytics"])
        self.assertEqual(work["gaps"][0]["markers"], "`api-parity:gap GET /sites/{}/aeo/reports/bot_traffic`")
        self.assertEqual(work["drift"], [])
        self.assertEqual(list(baseline_pages), ["sites__get.md"])
        self.assertEqual(
            {e["key"]: e["wrapped"] for e in manifest["endpoints"]},
            {"GET /sites/{}": True, "GET /sites/{}/aeo/reports/bot_traffic": False},
        )

    def test_other_method_on_a_wrapped_path_is_a_gap(self):
        delete_site = snapshot.parse_page(entry("sites/delete", title="Delete Site"), page_text("Delete Site", "DELETE", "/sites/{site_id}"))

        work, _, _ = self.compare([site_page(), delete_site])

        self.assertEqual([e["key"] for gap in work["gaps"] for e in gap["endpoints"]], ["DELETE /sites/{}"])

    def test_any_method_template_wraps_every_method_on_its_path(self):
        delete_site = snapshot.parse_page(entry("sites/delete", title="Delete Site"), page_text("Delete Site", "DELETE", "/sites/{site_id}"))

        work, _, _ = self.compare([site_page(), delete_site], called={"* /sites/{}": ["site.go"]})

        self.assertEqual(work["gaps"], [])

    def test_contract_change_on_wrapped_endpoint_is_drift_with_a_full_diff(self):
        _, baseline = self.save([site_page()])

        work, _, _ = self.compare([self.changed_site_page()], baseline)

        [drift] = work["drift"]
        self.assertEqual(drift["key"], "GET /sites/{}")
        self.assertEqual(drift["provider_files"], ["site.go"])
        self.assertIn("+- `locale` (string, optional) — Default locale", drift["diff"])
        self.assertRegex(drift["marker"], r"^`api-parity:drift GET /sites/\{\} [0-9a-f]{16}`$")
        self.assertIsNone(drift["open_issue"])

    def test_wording_change_is_not_drift_and_the_baseline_takes_the_new_text(self):
        _, baseline = self.save([site_page()])

        work, _, baseline_pages = self.compare([site_page(description="Reworded.")], baseline)

        self.assertFalse(snapshot.has_work(work))
        self.assertIn("Reworded.", baseline_pages["sites__get.md"])

    def test_unpublished_drift_keeps_the_old_baseline_and_is_reported_again(self):
        _, baseline = self.save([site_page()])

        first, kept = self.save([self.changed_site_page()], baseline)
        second, _, _ = self.compare([self.changed_site_page()], kept)

        self.assertEqual(kept["endpoints"], baseline["endpoints"])
        self.assertEqual(second["drift"][0]["marker"], first["drift"][0]["marker"])

    def test_published_drift_advances_the_baseline(self):
        _, baseline = self.save([site_page()])
        first, kept = self.save([self.changed_site_page()], baseline)
        published = [issue(12, "OPEN", "Details.\n\n" + first["drift"][0]["marker"])]

        work, advanced = self.save([self.changed_site_page()], kept, issues=published)

        self.assertFalse(snapshot.has_work(work))
        self.assertEqual(advanced["endpoints"][0]["hash"], self.changed_site_page().digest)

    def test_drift_published_in_a_comment_counts_and_routes_new_drift_to_that_open_issue(self):
        _, baseline = self.save([site_page()])
        first, kept = self.save([self.changed_site_page()], baseline)
        issues = [issue(12, "OPEN", "Original drift report.")]
        comments = [{"issue_number": 12, "body": "Another change.\n" + first["drift"][0]["marker"]}]

        settled, advanced = self.save([self.changed_site_page()], kept, issues, comments)
        newer = site_page(response=SITE_RESPONSE + "\n- `locale` (string, required) — Default locale")
        again, _, _ = self.compare([newer], advanced, issues, comments)

        self.assertFalse(snapshot.has_work(settled))
        self.assertEqual(again["drift"][0]["open_issue"], 12)

    def test_gap_published_in_a_closed_issue_is_not_reported_again(self):
        issues = [issue(3, "CLOSED", "Not planned.\n`api-parity:gap GET /sites/{}/aeo/reports/bot_traffic`")]

        work, _, _ = self.compare([site_page(), bot_traffic_page()], issues=issues)

        self.assertEqual(work["gaps"], [])

    def test_html_comment_markers_are_not_recognized(self):
        issues = [issue(3, "OPEN", "<!-- api-parity:gap GET /sites/{}/aeo/reports/bot_traffic -->")]

        work, _, _ = self.compare([site_page(), bot_traffic_page()], issues=issues)

        self.assertEqual(len(work["gaps"]), 1)

    def test_newly_wrapped_endpoint_starts_drift_tracking_after_the_next_save(self):
        tracked_gap = [issue(1, "OPEN", "`api-parity:gap GET /sites/{}`")]
        self.called = {}
        _, unwrapped = self.save([site_page()], issues=tracked_gap)
        self.called = {"GET /sites/{}": ["site.go"]}
        _, wrapped = self.save([site_page()], unwrapped, issues=tracked_gap)

        work, _, _ = self.compare([self.changed_site_page()], wrapped, issues=tracked_gap)

        self.assertEqual([d["key"] for d in work["drift"]], ["GET /sites/{}"])

    def test_unpublished_removal_stays_in_the_baseline_until_published(self):
        _, baseline = self.save([site_page(), bot_traffic_page()])

        first, kept = self.save([site_page()], baseline)
        second, dropped = self.save([site_page()], kept, issues=[issue(9, "OPEN", first["removed"][0]["marker"])])

        self.assertEqual(
            [(r["key"], r["was_wrapped"]) for r in first["removed"]], [("GET /sites/{}/aeo/reports/bot_traffic", False)]
        )
        self.assertIn("GET /sites/{}/aeo/reports/bot_traffic", {e["key"] for e in kept["endpoints"]})
        self.assertEqual(second["removed"], [])
        self.assertEqual([e["key"] for e in dropped["endpoints"]], ["GET /sites/{}"])


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
            (root / "provider" / "site.go").write_text(
                'func GetSite() {\n\tu := fmt.Sprintf("%s/v2/sites/%s", base, id)\n'
                '\treq, _ := http.NewRequestWithContext(ctx, "GET", u, http.NoBody)\n}\n'
            )
            comments = root / "comments.json"
            comments.write_text("[]")
            work = snapshot.run(
                root / "provider", root / "baseline", None, root / "out", fetch=fetch, min_index_entries=1, comments_path=comments
            )

            written = json.loads((root / "out" / "work.json").read_text())
            baseline_files = sorted(p.name for p in (root / "out" / "baseline" / "pages").iterdir())

        self.assertEqual(written, work)
        self.assertEqual(work["counts"]["documented_endpoints"], 4)
        self.assertEqual(work["counts"]["wrapped_endpoints"], 1)
        self.assertEqual(baseline_files, ["sites__get.md"])


if __name__ == "__main__":
    unittest.main()
