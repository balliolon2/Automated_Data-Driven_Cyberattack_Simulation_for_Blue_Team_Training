import io
import json
import tempfile
import unittest
from contextlib import redirect_stderr
from pathlib import Path

import draft_evidence_mapping as drafting
import generate_scenarios
import validate_evidence_mapping as validator


def make_scenario():
    logs = [
        {
            "timestamp": "2026-07-13T08:00:00Z",
            "source": "ad_controller",
            "severity": "info",
            "message": "User login successful",
        },
        {
            "timestamp": "2026-07-13T08:00:10Z",
            "source": "ad_controller",
            "severity": "info",
            "message": "User login successful",
        },
        {
            "timestamp": "2026-07-13T08:00:20Z",
            "source": "ad_controller",
            "severity": "info",
            "message": "User login successful",
        },
        {
            "timestamp": "2026-07-13T08:03:00Z",
            "source": "edr",
            "severity": "high",
            "message": "Malicious command executed on workstation",
        },
        {
            "timestamp": "2026-07-13T08:03:10Z",
            "source": "firewall",
            "severity": "warning",
            "message": "Outbound connection to known command-and-control host",
        },
        {
            "timestamp": "2026-07-13T08:03:20Z",
            "source": "proxy",
            "severity": "info",
            "message": "Routine proxy traffic",
        },
    ]
    return {
        "title": "Draft Test Scenario",
        "description": "A test scenario for the drafting seam.",
        "domain_id": "domain1",
        "difficulty": 1,
        "is_true_positive": True,
        "tp_fp_explanation": "The malicious command proves the incident.",
        "initial_logs": logs,
        "expected_outcomes": {
            "key_findings": [
                {"id": "f1", "description": "Found the malicious command", "points": 10},
                {"id": "f2", "description": "Found the exfiltration channel", "points": 10},
            ]
        },
    }


def make_entry(f1_indices=None, f2_indices=None):
    return {
        "scenario_title": "Draft Test Scenario",
        "findings": [
            {
                "finding_id": "f1",
                "evidence_log_indices": [3] if f1_indices is None else f1_indices,
                "rationale": "Log 3 shows the malicious command execution.",
                "verdict_support": "malicious",
            },
            {
                "finding_id": "f2",
                "evidence_log_indices": [4] if f2_indices is None else f2_indices,
                "rationale": "Log 4 shows the command-and-control connection.",
                "verdict_support": "malicious",
            },
        ],
    }


class FakeClient:
    def __init__(self, responses):
        self.responses = list(responses)
        self.calls = []

    def complete(self, system_prompt, user_prompt):
        self.calls.append((system_prompt, user_prompt))
        if not self.responses:
            raise AssertionError("fake client received an unexpected request")
        response = self.responses.pop(0)
        if isinstance(response, Exception):
            raise response
        return response


class Fixture(object):
    def __init__(self, scenario=None, **kwargs):
        self.scenario = scenario or make_scenario()
        self.kwargs = kwargs

    def results(self, responses, max_retries=2):
        client = FakeClient(responses)
        results = drafting.draft_scenarios(
            [self.scenario],
            client=client,
            max_retries=max_retries,
            concurrency=1,
            **self.kwargs
        )
        return client, results


class ValidLlmOutputTests(unittest.TestCase):
    def test_valid_output_produces_mapping_and_report(self):
        client = FakeClient([json.dumps(make_entry())])
        results = drafting.draft_scenarios(
            [make_scenario()], client=client, max_retries=2, concurrency=1
        )

        self.assertEqual(1, len(results))
        self.assertFalse(results[0].failed)
        self.assertEqual(1, len(client.calls))

        mapping = drafting.build_mapping(
            results,
            model="fake-model",
            drafted_by="llm",
            generated_at="2026-10-22T00:00:00Z",
        )
        self.assertFalse(mapping["approval"]["approved"])
        self.assertIsNone(mapping["approval"]["mapping_sha256"])
        self.assertEqual("fake-model", mapping["model"])
        self.assertEqual("evidence-mapping-v1", mapping["prompt_version"])
        self.assertEqual("llm", mapping["drafted_by"])
        self.assertEqual("2026-10-22T00:00:00Z", mapping["generated_at"])
        self.assertEqual(2, len(mapping["scenarios"][0]["findings"]))

        report = drafting.render_report(
            [make_scenario()],
            results,
            model="fake-model",
            prompt_version="evidence-mapping-v1",
            drafted_by="llm",
            generated_at="2026-10-22T00:00:00Z",
        )
        self.assertIn("Draft Test Scenario", report)
        self.assertIn("PENDING APPROVAL", report)
        self.assertIn("malicious", report)


class RetryTests(unittest.TestCase):
    def test_invalid_output_retries_with_exact_errors(self):
        invalid = make_entry(f1_indices=[99])
        client = FakeClient([json.dumps(invalid), json.dumps(make_entry())])
        results = drafting.draft_scenarios(
            [make_scenario()], client=client, max_retries=2, concurrency=1
        )

        self.assertEqual(1, len(results))
        self.assertFalse(results[0].failed)
        self.assertEqual(2, results[0].attempts)
        self.assertEqual(2, len(client.calls))
        retry_prompt = client.calls[1][1]
        self.assertIn("out of range", retry_prompt)
        self.assertIn("99", retry_prompt)
        self.assertIn("f1", retry_prompt)

    def test_retry_exhaustion_marks_scenario_failed_and_exits_non_zero(self):
        invalid = json.dumps(make_entry(f1_indices=[99]))
        client = FakeClient([invalid, invalid, invalid])

        with tempfile.TemporaryDirectory() as tmp:
            output = Path(tmp) / "mapping.json"
            report = Path(tmp) / "report.md"
            stderr = io.StringIO()
            with redirect_stderr(stderr):
                return_code = drafting.main(
                    [
                        "--output",
                        str(output),
                        "--report",
                        str(report),
                        "--concurrency",
                        "1",
                        "--max-retries",
                        "2",
                    ],
                    client=client,
                    scenarios=[make_scenario()],
                )

            self.assertNotEqual(0, return_code)
            self.assertEqual(3, len(client.calls))
            self.assertIn("out of range", stderr.getvalue())

            mapping = json.loads(output.read_text(encoding="utf-8"))
            self.assertEqual(1, len(mapping["failed_scenarios"]))
            self.assertEqual(
                "Draft Test Scenario", mapping["failed_scenarios"][0]["scenario_title"]
            )
            self.assertIn("FAILED", report.read_text(encoding="utf-8"))


class ReportTests(unittest.TestCase):
    def test_report_contains_required_columns_and_snippets(self):
        client = FakeClient([json.dumps(make_entry())])
        results = drafting.draft_scenarios(
            [make_scenario()], client=client, max_retries=2, concurrency=1
        )
        report = drafting.render_report(
            [make_scenario()],
            results,
            model="fake-model",
            prompt_version="evidence-mapping-v1",
            drafted_by="llm",
            generated_at="2026-10-22T00:00:00Z",
        )

        self.assertIn("| Scenario | Finding | Evidence logs | Rationale | Support |", report)
        self.assertIn("Status: PENDING APPROVAL", report)
        self.assertIn("TRUE POSITIVE", report)
        self.assertIn("Evidence union: 2 / 6 logs (33%)", report)
        self.assertIn("#3", report)
        self.assertIn("2026-07-13T08:03:00Z", report)
        self.assertIn("edr", report)
        self.assertIn("high", report)
        self.assertIn("Malicious command executed on workstation", report)
        self.assertIn("Log 3 shows the malicious command execution.", report)
        self.assertIn("malicious", report)
        self.assertIn("## Warnings", report)


class OfflineTests(unittest.TestCase):
    def test_drafting_module_never_chunks_positionally(self):
        source = Path(drafting.__file__).read_text(encoding="utf-8")
        self.assertNotIn("assign_evidence_indices(", source)
        self.assertNotIn("evidence_start", source)

    def test_offline_mode_produces_complete_mapping_without_chunking(self):
        scenarios = generate_scenarios.create_scenarios()
        results = drafting.draft_scenarios(
            scenarios, client=None, max_retries=0, concurrency=1, offline=True
        )

        self.assertEqual(15, len(results))
        self.assertFalse([result for result in results if result.failed])
        self.assertTrue(all(result.drafted_by == "offline" for result in results))

        mapping = drafting.build_mapping(results, model="offline", drafted_by="offline")
        errors = validator.validate_mapping(mapping, scenarios, require_complete=True)
        self.assertEqual([], errors)

        entries = {entry["scenario_title"]: entry for entry in mapping["scenarios"]}
        self.assertEqual(15, len(entries))
        for scenario in scenarios:
            entry = entries[scenario["title"]]
            self.assertEqual(
                {finding["id"] for finding in scenario["expected_outcomes"]["key_findings"]},
                {finding["finding_id"] for finding in entry["findings"]},
            )
            for finding in entry["findings"]:
                self.assertTrue(finding["evidence_log_indices"])
                self.assertTrue(finding["rationale"])
                self.assertIn(
                    finding["verdict_support"], {"malicious", "benign", "supporting"}
                )

        travel = entries["Atypical Travel VPN Connection Alert"]
        by_id = {finding["finding_id"]: finding for finding in travel["findings"]}
        self.assertEqual([18], by_id["d1_f3"]["evidence_log_indices"])
        self.assertEqual([16, 17], by_id["d1_f4"]["evidence_log_indices"])

        chunked = generate_scenarios.assign_evidence_indices(
            generate_scenarios.create_scenarios()
        )
        chunked_travel = next(
            scenario
            for scenario in chunked
            if scenario["title"] == "Atypical Travel VPN Connection Alert"
        )
        chunked_by_id = {
            finding["id"]: finding
            for finding in chunked_travel["expected_outcomes"]["key_findings"]
        }
        self.assertNotEqual(
            chunked_by_id["d1_f3"]["evidence_log_indices"],
            by_id["d1_f3"]["evidence_log_indices"],
        )


if __name__ == "__main__":
    unittest.main(verbosity=2)
