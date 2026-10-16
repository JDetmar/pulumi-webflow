import http.server
import json
import tempfile
import threading
import unittest
from pathlib import Path
from unittest import mock

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
        gap = work["gaps"][0]["endpoints"][0]
        self.assertEqual(gap["url"], "https://developers.webflow.com/data/v2.0.0/reference/aeo-analytics/get-bot-traffic")
        self.assertEqual(work["gaps"][0]["provider_files"], [])
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
        self.assertEqual(work["gaps"][0]["provider_files"], ["site.go"])

    def test_provider_call_the_reference_does_not_document_is_reported_once(self):
        update_field = snapshot.parse_page(
            entry("cms/collection-fields/update", "CMS > Collection Fields", "Update field"),
            page_text("Update field", "PATCH", "/collections/{collection_id}/fields/{field_id}"),
        )
        called = {
            "PUT /collections/{}/fields/{}": ["collectionfield.go"],
            "* /collections/{}/legacy": ["urls.go"],
            "GET /sites/{}": ["site.go"],
        }

        work, _, _ = self.compare([site_page(), update_field], called=called)

        self.assertEqual(
            work["undocumented"],
            [
                {
                    "title": "API undocumented: collectionfield.go",
                    "provider_files": ["collectionfield.go"],
                    "endpoints": [{"key": "PUT /collections/{}/fields/{}", "method": "PUT", "path": "/collections/{}/fields/{}"}],
                    "markers": "`api-parity:undocumented PUT /collections/{}/fields/{}`",
                }
            ],
        )
        published = [issue(9, "CLOSED", work["undocumented"][0]["markers"])]
        work, _, _ = self.compare([site_page(), update_field], called=called, issues=published)
        self.assertEqual(work["undocumented"], [])

    def test_pending_removal_of_a_wrapped_endpoint_is_not_also_reported_as_undocumented(self):
        _, baseline = self.save([site_page()])

        work, _, _ = self.compare([bot_traffic_page()], baseline)

        self.assertEqual([item["key"] for item in work["removed"]], ["GET /sites/{}"])
        self.assertEqual(work["undocumented"], [])

        published = [issue(4, "OPEN", work["removed"][0]["marker"])]
        _, baseline = self.save([bot_traffic_page()], baseline, published)
        work, _, _ = self.compare([bot_traffic_page()], baseline, published)
        self.assertEqual((work["removed"], work["undocumented"]), ([], []))

    def test_gaps_touching_provider_code_come_first(self):
        def page(slug, section, method, path):
            return snapshot.parse_page(entry(slug, section, slug), page_text(slug, method, path))

        pages = [
            bot_traffic_page(),
            page("collections/get", "CMS > Collections", "GET", "/collections/{collection_id}"),
            page("collections/create-field", "CMS > Collections", "POST", "/collections/{collection_id}/fields"),
            site_page(),
            page("sites/delete", "Sites", "DELETE", "/sites/{site_id}"),
        ]

        work, _, _ = self.compare(pages, called={"GET /sites/{}": ["site.go"], "GET /collections/{}": ["collection.go"]})

        self.assertEqual([gap["section"] for gap in work["gaps"]], ["CMS > Collections", "Sites", "AEO Analytics"])

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
        self.assertEqual(drift["page_file"], "reference/sites__get.md")
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


class CheckPagesTest(unittest.TestCase):
    def test_known_endpoint_page_without_an_endpoint_line_fails_instead_of_reporting_removal(self):
        baseline = {"endpoints": [{"key": "GET /sites/{}", "slug": "sites/get"}]}
        error_page = snapshot.parse_page(entry("sites/get"), "# Something went wrong\n\nPlease try again.\n")

        with self.assertRaisesRegex(snapshot.SnapshotError, "sites/get"):
            snapshot.compare([error_page], {}, baseline, Path("unused"), [])

    def test_two_pages_for_the_same_endpoint_fail(self):
        legacy = snapshot.parse_page(entry("legacy/sites/get"), page_text("Get Site (legacy)", "GET", "/sites/{siteId}"))

        with self.assertRaisesRegex(snapshot.SnapshotError, r"GET /sites/\{\} \(sites/get, legacy/sites/get\)"):
            snapshot.compare([site_page(), legacy], {}, None, Path("unused"), [])

    def test_large_share_of_endpoints_vanishing_at_once_fails(self):
        baseline = {"endpoints": [{"key": f"GET /things/{n}", "slug": f"things/{n}"} for n in range(9)]}
        baseline["endpoints"].append({"key": "GET /sites/{}", "slug": "sites/get"})

        with self.assertRaisesRegex(snapshot.SnapshotError, "9 of 10 baseline endpoints are missing"):
            snapshot.compare([site_page()], {}, baseline, Path("unused"), [])

    def test_new_page_without_an_endpoint_line_is_allowed(self):
        event = snapshot.parse_page(entry("webhooks/events/site-publish"), "# Site publish\n\nPayload.\n")

        work, _, _ = snapshot.compare([site_page(), event], {"GET /sites/{}": ["site.go"]}, None, Path("unused"), [])

        self.assertEqual(work["counts"]["documented_endpoints"], 1)


class FetchTextTest(unittest.TestCase):
    """fetch_text against a local server that misbehaves the way a flaky CDN does."""

    def serve(self, responses):
        """Serve one scripted response per request: (status, body, declared length)."""
        requests = []

        class Handler(http.server.BaseHTTPRequestHandler):
            def do_GET(self):
                requests.append(self.path)
                status, body, length = responses[min(len(requests), len(responses)) - 1]
                self.send_response(status)
                self.send_header("Content-Length", str(length))
                self.end_headers()
                self.wfile.write(body)
                self.close_connection = True

            def log_message(self, *args):
                pass

        server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        threading.Thread(target=server.serve_forever, daemon=True).start()
        self.addCleanup(server.server_close)
        self.addCleanup(server.shutdown)
        sleep = mock.patch.object(snapshot.time, "sleep")
        sleep.start()
        self.addCleanup(sleep.stop)
        return f"http://127.0.0.1:{server.server_address[1]}/page.md", requests

    def test_retries_a_body_cut_short(self):
        body = b"# Get Site\n"
        url, requests = self.serve([(200, body[:4], len(body)), (200, body, len(body))])

        self.assertEqual(snapshot.fetch_text(url), "# Get Site\n")
        self.assertEqual(len(requests), 2)

    def test_retries_server_errors_then_gives_up(self):
        url, requests = self.serve([(503, b"busy", 4)])

        with self.assertRaisesRegex(RuntimeError, "after 3 attempt"):
            snapshot.fetch_text(url)
        self.assertEqual(len(requests), snapshot.FETCH_ATTEMPTS)

    def test_does_not_retry_a_missing_page(self):
        url, requests = self.serve([(404, b"gone", 4)])

        with self.assertRaisesRegex(RuntimeError, "after 1 attempt"):
            snapshot.fetch_text(url)
        self.assertEqual(len(requests), 1)


class RunTest(unittest.TestCase):
    def test_rejects_a_truncated_index_before_fetching_pages(self):
        fetched = []

        def fetch(url):
            fetched.append(url)
            return INDEX

        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaisesRegex(snapshot.SnapshotError, "listed only 5 reference pages"):
                snapshot.run(Path(tmp), Path(tmp) / "baseline", None, Path(tmp) / "out", fetch=fetch)

        self.assertEqual(fetched, [snapshot.DOCS_INDEX_URL])

    def test_rejects_pages_that_document_too_few_endpoints(self):
        def fetch(url):
            return INDEX if url == snapshot.DOCS_INDEX_URL else "# Maintenance\n\nBack soon.\n"

        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaisesRegex(snapshot.SnapshotError, "documented only 0 endpoints"):
                snapshot.run(Path(tmp), Path(tmp) / "baseline", None, Path(tmp) / "out", fetch=fetch, min_index_entries=1)

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
            gap_pages = [(root / "out" / e["page_file"]).read_text() for gap in work["gaps"] for e in gap["endpoints"]]
            index = (root / "out" / "reference" / "index.md").read_text()

        self.assertEqual(written, work)
        self.assertEqual(work["counts"]["documented_endpoints"], 4)
        self.assertEqual(work["counts"]["wrapped_endpoints"], 1)
        self.assertEqual(baseline_files, ["sites__get.md"])
        self.assertEqual(len(gap_pages), 3)
        self.assertTrue(all(text.startswith("# ") and "## Examples" not in text for text in gap_pages))
        self.assertIn(
            "- `DELETE /sites/{site_id}/access` | Enterprise > Site Access | Revoke Site Access (by email) | "
            "reference/enterprise__site-access__remove-member-by-email.md",
            index,
        )
        self.assertNotIn("site-publish", index)


if __name__ == "__main__":
    unittest.main()
