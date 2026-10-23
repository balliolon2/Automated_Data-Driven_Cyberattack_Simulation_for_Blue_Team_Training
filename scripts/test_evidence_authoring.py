import copy
import io
import json
import tempfile
import unittest
from contextlib import redirect_stderr
from pathlib import Path

import generate_scenarios
import validate_evidence_mapping as validator

SCRIPT_DIR = Path(__file__).resolve().parent
MAPPING_PATH = SCRIPT_DIR / "evidence_mapping.json"


def make_logs(count, severity="info"):
    return [
        {
            "timestamp": f"2026-07-13T08:00:{index:02d}Z",
            "source": "test_source",
            "severity": severity,
            "message": f"log message {index}",
        }
        for index in range(count)
    ]


def make_base_scenario(title="Scenario One", is_true_positive=True, log_count=10, findings=None):
    if findings is None:
        findings = [
            {"id": "f1", "points": 10},
            {"id": "f2", "points": 10},
        ]

    return {
        "title": title,
        "description": "test scenario",
        "domain_id": "domain1",
        "difficulty": 1,
        "is_true_positive": is_true_positive,
        "tp_fp_explanation": "test explanation",
        "initial_logs": make_logs(log_count),
        "expected_outcomes": {"key_findings": findings},
    }


def make_mapping(findings, title="Scenario One", schema_version=1):
    return {
        "schema_version": schema_version,
        "scenarios": [
            {
                "scenario_title": title,
                "findings": findings,
            }
        ],
    }


def make_finding(finding_id="f1", indices=None, rationale="Logs prove the finding.", support="malicious"):
    return {
        "finding_id": finding_id,
        "evidence_log_indices": [0, 1] if indices is None else indices,
        "rationale": rationale,
        "verdict_support": support,
    }


class ValidatorTestCase(unittest.TestCase):
    def assert_valid(self, errors):
        self.assertEqual([], errors)

    def assert_invalid(self, errors, *fragments):
        self.assertTrue(errors, "expected validation errors, got none")
        joined = "\n".join(errors)
        for fragment in fragments:
            self.assertIn(fragment, joined)

    def validate(self, mapping, base, **kwargs):
        return validator.validate_mapping(mapping, base, **kwargs)


class MappingSchemaTests(ValidatorTestCase):
    def test_valid_mapping_passes(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding(), make_finding("f2", [2, 3])])
        self.assert_valid(self.validate(mapping, base))

    def test_missing_schema_version_fails(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding()])
        del mapping["schema_version"]
        self.assert_invalid(self.validate(mapping, base), "schema_version")

    def test_unsupported_schema_version_fails(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding()], schema_version=2)
        self.assert_invalid(self.validate(mapping, base), "schema_version", "2")


class MalformedInputTests(ValidatorTestCase):
    def test_list_verdict_support_fails_cleanly(self):
        base = [make_base_scenario()]
        mapping = make_mapping(
            [make_finding("f1", support=["malicious"]), make_finding("f2", [4, 5])]
        )
        self.assert_invalid(self.validate(mapping, base), "f1", "verdict_support")

    def test_dict_schema_version_fails_cleanly(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding(), make_finding("f2", [4, 5])], schema_version={})
        self.assert_invalid(self.validate(mapping, base), "schema_version")

    def test_cli_malformed_mapping_exits_one_without_traceback(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "broken_mapping.json"
            path.write_text(
                json.dumps({"schema_version": {}, "scenarios": []}),
                encoding="utf-8",
            )
            stderr = io.StringIO()
            with redirect_stderr(stderr):
                return_code = validator.main(["--mapping", str(path)])

        self.assertEqual(1, return_code)
        self.assertNotIn("Traceback", stderr.getvalue())
        self.assertIn("schema_version", stderr.getvalue())


class ScenarioSetTests(ValidatorTestCase):
    def test_missing_scenario_fails_when_complete_coverage_required(self):
        base = [make_base_scenario(), make_base_scenario(title="Scenario Two")]
        mapping = make_mapping([make_finding(), make_finding("f2", [2, 3])])
        self.assert_invalid(
            self.validate(mapping, base, require_complete=True),
            "Scenario Two",
            "missing scenario",
        )

    def test_missing_scenario_is_allowed_in_expand_phase(self):
        base = [make_base_scenario(), make_base_scenario(title="Scenario Two")]
        mapping = make_mapping([make_finding(), make_finding("f2", [2, 3])])
        self.assert_valid(self.validate(mapping, base, require_complete=False))

    def test_extra_scenario_fails(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding()], title="Unknown Scenario")
        self.assert_invalid(self.validate(mapping, base), "Unknown Scenario", "not in base")

    def test_duplicate_scenario_fails(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding()])
        mapping["scenarios"].append(copy.deepcopy(mapping["scenarios"][0]))
        self.assert_invalid(self.validate(mapping, base), "duplicate scenario")


class FindingSetTests(ValidatorTestCase):
    def test_missing_finding_fails(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding("f1")])
        self.assert_invalid(self.validate(mapping, base), "Scenario One", "f2", "missing finding")

    def test_extra_finding_fails(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding("f1"), make_finding("f2", [2, 3]), make_finding("f9", [4])])
        self.assert_invalid(self.validate(mapping, base), "Scenario One", "f9", "not in base")

    def test_duplicate_finding_fails(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding("f1"), make_finding("f2", [2, 3]), make_finding("f2", [4, 5])])
        self.assert_invalid(self.validate(mapping, base), "f2", "duplicate finding")


class EvidenceIndexTests(ValidatorTestCase):
    def validate_single(self, indices):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding("f1", indices), make_finding("f2", [4, 5])])
        return self.validate(mapping, base)

    def test_empty_evidence_fails(self):
        self.assert_invalid(self.validate_single([]), "f1", "empty")

    def test_duplicate_index_fails(self):
        self.assert_invalid(self.validate_single([0, 0]), "f1", "duplicate")

    def test_out_of_range_index_fails(self):
        self.assert_invalid(self.validate_single([0, 10]), "f1", "out of range")

    def test_negative_index_fails(self):
        self.assert_invalid(self.validate_single([-1, 0]), "f1", "out of range")

    def test_non_integer_index_fails(self):
        self.assert_invalid(self.validate_single([0, "1"]), "f1", "integer")


class RationaleTests(ValidatorTestCase):
    def validate_rationale(self, rationale):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding("f1", rationale=rationale), make_finding("f2", [4, 5])])
        return self.validate(mapping, base)

    def test_empty_rationale_fails(self):
        self.assert_invalid(self.validate_rationale("   "), "f1", "rationale")

    def test_rationale_longer_than_300_chars_fails(self):
        self.assert_invalid(self.validate_rationale("x" * 301), "f1", "rationale", "300")

    def test_rationale_at_300_chars_passes(self):
        self.assert_valid(self.validate_rationale("x" * 300))


class VerdictSupportTests(ValidatorTestCase):
    def test_invalid_verdict_support_fails(self):
        base = [make_base_scenario()]
        mapping = make_mapping([make_finding("f1", support="unknown"), make_finding("f2", [4, 5])])
        self.assert_invalid(self.validate(mapping, base), "f1", "verdict_support")

    def test_missing_verdict_support_fails(self):
        base = [make_base_scenario()]
        finding = make_finding("f1")
        del finding["verdict_support"]
        mapping = make_mapping([finding, make_finding("f2", [4, 5])])
        self.assert_invalid(self.validate(mapping, base), "f1", "verdict_support")

    def test_true_positive_without_malicious_support_fails(self):
        base = [make_base_scenario(is_true_positive=True)]
        mapping = make_mapping([make_finding("f1", support="supporting"), make_finding("f2", [4, 5], support="benign")])
        self.assert_invalid(self.validate(mapping, base), "Scenario One", "malicious")

    def test_false_positive_without_benign_support_fails(self):
        base = [make_base_scenario(is_true_positive=False)]
        mapping = make_mapping([make_finding("f1", support="malicious"), make_finding("f2", [4, 5], support="supporting")])
        self.assert_invalid(self.validate(mapping, base), "Scenario One", "benign")


class CoverageTests(ValidatorTestCase):
    def test_union_at_fifty_percent_passes(self):
        base = [make_base_scenario(log_count=100)]
        mapping = make_mapping([make_finding("f1", list(range(50))), make_finding("f2", [49])])
        self.assert_valid(self.validate(mapping, base))

    def test_union_above_fifty_percent_fails(self):
        base = [make_base_scenario(log_count=100)]
        mapping = make_mapping([make_finding("f1", list(range(50))), make_finding("f2", [50])])
        self.assert_invalid(self.validate(mapping, base), "Scenario One", "union", "50%")

    def test_shared_evidence_is_counted_once_in_union(self):
        base = [make_base_scenario(log_count=10)]
        mapping = make_mapping([make_finding("f1", [0, 1, 2]), make_finding("f2", [2, 3, 4])])
        self.assert_valid(self.validate(mapping, base))

    def test_evidence_equal_to_all_logs_fails(self):
        base = [make_base_scenario(log_count=4)]
        mapping = make_mapping([make_finding("f1", [0, 1, 2, 3]), make_finding("f2", [0])])
        self.assert_invalid(self.validate(mapping, base), "Scenario One", "all logs")


class PointsTests(ValidatorTestCase):
    def validate_points(self, points):
        base = [make_base_scenario(findings=[{"id": "f1", "points": 10}, {"id": "f2", "points": 10}])]
        finding = make_finding("f1")
        finding["points"] = points
        mapping = make_mapping([finding, make_finding("f2", [4, 5])])
        return self.validate(mapping, base)

    def test_zero_points_fails(self):
        self.assert_invalid(self.validate_points(0), "f1", "points")

    def test_negative_points_fails(self):
        self.assert_invalid(self.validate_points(-5), "f1", "points")

    def test_fractional_points_fails(self):
        self.assert_invalid(self.validate_points(10.5), "f1", "points")

    def test_missing_points_defaults_to_positive(self):
        base = [make_base_scenario(findings=[{"id": "f1"}, {"id": "f2", "points": 10}])]
        mapping = make_mapping([make_finding("f1"), make_finding("f2", [4, 5])])
        self.assert_valid(self.validate(mapping, base))


class SeedValidationTests(ValidatorTestCase):
    def make_seed_scenario(self, is_true_positive=True, findings=None, title="Scenario One"):
        scenario = make_base_scenario(title=title, is_true_positive=is_true_positive)
        scenario["expected_outcomes"]["key_findings"] = findings or [
            {"id": "f1", "points": 10, "evidence_log_indices": [0, 1], "verdict_support": "malicious"},
            {"id": "f2", "points": 10, "evidence_log_indices": [2, 3], "verdict_support": "supporting"},
        ]
        return scenario

    def validate(self, seed, base=None, **kwargs):
        if base is None:
            base = [make_base_scenario()]
        return validator.validate_seed(seed, base, **kwargs)

    def test_valid_seed_passes(self):
        self.assert_valid(self.validate([self.make_seed_scenario()]))

    def test_seed_with_missing_evidence_array_fails(self):
        scenario = self.make_seed_scenario()
        del scenario["expected_outcomes"]["key_findings"][0]["evidence_log_indices"]
        self.assert_invalid(self.validate([scenario]), "Scenario One", "f1", "empty")

    def test_seed_with_duplicate_index_fails(self):
        scenario = self.make_seed_scenario()
        scenario["expected_outcomes"]["key_findings"][0]["evidence_log_indices"] = [0, 0]
        self.assert_invalid(self.validate([scenario]), "Scenario One", "f1", "duplicate")

    def test_seed_with_out_of_range_index_fails(self):
        scenario = self.make_seed_scenario()
        scenario["expected_outcomes"]["key_findings"][0]["evidence_log_indices"] = [0, 99]
        self.assert_invalid(self.validate([scenario]), "Scenario One", "f1", "out of range")

    def test_legacy_seed_without_verdict_support_is_tolerated_for_unmapped_fallback(self):
        scenario = self.make_seed_scenario()
        for finding in scenario["expected_outcomes"]["key_findings"]:
            del finding["verdict_support"]
        self.assert_valid(self.validate([scenario], mapped_titles=set()))

    def test_mapped_scenario_cannot_delete_all_verdict_support(self):
        scenario = self.make_seed_scenario()
        for finding in scenario["expected_outcomes"]["key_findings"]:
            del finding["verdict_support"]
        self.assert_invalid(
            self.validate([scenario], mapped_titles={"Scenario One"}),
            "Scenario One",
            "f1",
            "verdict_support",
        )

    def test_seed_with_partial_verdict_support_fails(self):
        scenario = self.make_seed_scenario()
        del scenario["expected_outcomes"]["key_findings"][1]["verdict_support"]
        self.assert_invalid(self.validate([scenario]), "Scenario One", "f2", "verdict_support")

    def test_seed_true_positive_without_malicious_fails(self):
        scenario = self.make_seed_scenario()
        scenario["expected_outcomes"]["key_findings"][0]["verdict_support"] = "benign"
        self.assert_invalid(self.validate([scenario]), "Scenario One", "malicious")

    def test_seed_false_positive_without_benign_fails(self):
        scenario = self.make_seed_scenario(is_true_positive=False)
        self.assert_invalid(self.validate([scenario]), "Scenario One", "benign")

    def test_seed_missing_scenario_fails(self):
        base = [make_base_scenario(), make_base_scenario(title="Scenario Two")]
        self.assert_invalid(
            self.validate([self.make_seed_scenario()], base),
            "Scenario Two",
            "missing scenario",
        )

    def test_seed_extra_scenario_fails(self):
        extra = self.make_seed_scenario(title="Scenario Two")
        self.assert_invalid(
            self.validate([self.make_seed_scenario(), extra], [make_base_scenario()]),
            "Scenario Two",
            "extra scenario",
        )

    def test_seed_duplicate_scenario_titles_fail(self):
        scenario = self.make_seed_scenario()
        self.assert_invalid(
            self.validate([scenario, copy.deepcopy(scenario)]),
            "Scenario One",
            "duplicate scenario",
        )

    def test_seed_duplicate_finding_ids_fail(self):
        scenario = self.make_seed_scenario()
        findings = scenario["expected_outcomes"]["key_findings"]
        findings.append(copy.deepcopy(findings[0]))
        self.assert_invalid(self.validate([scenario]), "Scenario One", "f1", "duplicate finding")

    def test_seed_empty_findings_fail(self):
        scenario = self.make_seed_scenario()
        scenario["expected_outcomes"]["key_findings"] = []
        self.assert_invalid(self.validate([scenario]), "Scenario One", "key_findings", "empty")

    def test_seed_missing_finding_fails(self):
        scenario = self.make_seed_scenario()
        scenario["expected_outcomes"]["key_findings"] = scenario["expected_outcomes"]["key_findings"][:1]
        self.assert_invalid(self.validate([scenario]), "Scenario One", "f2", "missing finding")

    def test_seed_extra_finding_fails(self):
        scenario = self.make_seed_scenario()
        scenario["expected_outcomes"]["key_findings"].append(
            {"id": "f9", "points": 10, "evidence_log_indices": [4], "verdict_support": "supporting"}
        )
        self.assert_invalid(self.validate([scenario]), "Scenario One", "f9", "extra finding")

    def test_seed_malformed_expected_outcomes_fails_cleanly(self):
        scenario = self.make_seed_scenario()
        scenario["expected_outcomes"] = ["not", "an", "object"]
        self.assert_invalid(self.validate([scenario]), "Scenario One", "expected_outcomes")


class SeedCliTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.mapping = generate_scenarios.load_evidence_mapping(str(MAPPING_PATH))
        seeded = generate_scenarios.assign_evidence_indices(generate_scenarios.create_scenarios())
        cls.full_seed = generate_scenarios.apply_evidence_mapping(seeded, cls.mapping)

    def write_seed(self, directory, seed):
        path = Path(directory) / "scenarios_seed.json"
        path.write_text(json.dumps(seed), encoding="utf-8")
        return path

    def test_cli_rejects_truncated_seed(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = self.write_seed(tmp, self.full_seed[:1])
            stderr = io.StringIO()
            with redirect_stderr(stderr):
                return_code = validator.main(["--seed", str(path)])

        self.assertEqual(1, return_code)
        self.assertIn("missing scenario", stderr.getvalue())
        self.assertNotIn("Traceback", stderr.getvalue())

    def test_cli_accepts_full_seed(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = self.write_seed(tmp, self.full_seed)
            stderr = io.StringIO()
            with redirect_stderr(stderr):
                return_code = validator.main(["--seed", str(path)])

        self.assertEqual(0, return_code, stderr.getvalue())


class WarningTests(ValidatorTestCase):
    def collect(self, mapping, base):
        warnings = []
        errors = self.validate(mapping, base, warnings=warnings)
        return errors, warnings

    def test_shared_log_emits_warning_without_failing(self):
        base = [make_base_scenario()]
        mapping = make_mapping(
            [make_finding("f1", [0, 1]), make_finding("f2", [1, 2])]
        )
        errors, warnings = self.collect(mapping, base)
        self.assert_valid(errors)
        self.assertTrue(any("shared" in warning for warning in warnings), warnings)

    def test_false_positive_all_high_severity_emits_warning(self):
        base = [make_base_scenario(is_true_positive=False)]
        base[0]["initial_logs"] = make_logs(10, severity="high")
        mapping = make_mapping(
            [
                make_finding("f1", [0], support="benign"),
                make_finding("f2", [1], support="supporting"),
            ]
        )
        errors, warnings = self.collect(mapping, base)
        self.assert_valid(errors)
        self.assertTrue(
            any("False Positive" in warning and "high/critical" in warning for warning in warnings),
            warnings,
        )

    def test_true_positive_all_info_severity_emits_warning(self):
        base = [make_base_scenario(is_true_positive=True)]
        base[0]["initial_logs"] = make_logs(10, severity="info")
        mapping = make_mapping(
            [
                make_finding("f1", [0], support="malicious"),
                make_finding("f2", [1], support="supporting"),
            ]
        )
        errors, warnings = self.collect(mapping, base)
        self.assert_valid(errors)
        self.assertTrue(
            any("True Positive" in warning and "info/low" in warning for warning in warnings),
            warnings,
        )

    def test_union_between_forty_and_fifty_percent_emits_warning(self):
        base = [make_base_scenario(log_count=10)]
        mapping = make_mapping(
            [
                make_finding("f1", [0, 1, 2], support="malicious"),
                make_finding("f2", [3], support="supporting"),
            ]
        )
        errors, warnings = self.collect(mapping, base)
        self.assert_valid(errors)
        self.assertTrue(
            any("union" in warning and "40%" in warning for warning in warnings), warnings
        )

    def test_seed_warnings_are_reported_without_failing(self):
        scenario = make_base_scenario(log_count=10)
        scenario["expected_outcomes"]["key_findings"] = [
            {"id": "f1", "points": 10, "evidence_log_indices": [0, 1], "verdict_support": "malicious"},
            {"id": "f2", "points": 10, "evidence_log_indices": [1, 2], "verdict_support": "supporting"},
        ]
        warnings = []
        errors = validator.validate_seed([scenario], [scenario], warnings=warnings)
        self.assert_valid(errors)
        self.assertTrue(any("shared" in warning for warning in warnings), warnings)
        self.assertTrue(any("finding 'f1'" in warning for warning in warnings), warnings)


class ApplyGuardTests(unittest.TestCase):
    def setUp(self):
        self.mapping = generate_scenarios.load_evidence_mapping(str(MAPPING_PATH))

    def seeded(self):
        return generate_scenarios.assign_evidence_indices(generate_scenarios.create_scenarios())

    def test_unknown_scenario_is_rejected(self):
        mapping = {"schema_version": 1, "scenarios": [{"scenario_title": "Unknown", "findings": []}]}
        with self.assertRaises(ValueError) as caught:
            generate_scenarios.apply_evidence_mapping(self.seeded(), mapping)
        self.assertIn("Unknown", str(caught.exception))

    def test_unknown_finding_id_is_rejected(self):
        entry = copy.deepcopy(self.mapping["scenarios"][0])
        entry["findings"].append(
            {
                "finding_id": "not_a_real_finding",
                "evidence_log_indices": [15],
                "rationale": "fabricated",
                "verdict_support": "malicious",
            }
        )
        mapping = {"schema_version": 1, "scenarios": [entry]}
        with self.assertRaises(ValueError) as caught:
            generate_scenarios.apply_evidence_mapping(self.seeded(), mapping)
        self.assertIn("not_a_real_finding", str(caught.exception))

    def test_partial_finding_coverage_is_rejected(self):
        entry = copy.deepcopy(self.mapping["scenarios"][0])
        entry["findings"] = entry["findings"][:1]
        mapping = {"schema_version": 1, "scenarios": [entry]}
        with self.assertRaises(ValueError) as caught:
            generate_scenarios.apply_evidence_mapping(self.seeded(), mapping)
        self.assertIn("missing finding", str(caught.exception))

    def test_malformed_mapping_root_is_rejected(self):
        with self.assertRaises(ValueError):
            generate_scenarios.apply_evidence_mapping(
                self.seeded(), {"schema_version": {}, "scenarios": []}
            )


class SeedApplicationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.mapping = generate_scenarios.load_evidence_mapping(str(MAPPING_PATH))

    def test_mapping_applies_when_every_scenario_is_authored(self):
        """The expand phase is over: the draft now covers all 15 Scenarios.

        T2 drafts the complete mapping but intentionally does not touch the
        seed; T3 approves and applies it. Applying the draft here must overlay
        every finding with its authored evidence.
        """
        base = generate_scenarios.create_scenarios()
        mapped_titles = {entry["scenario_title"] for entry in self.mapping["scenarios"]}
        self.assertEqual(15, len(mapped_titles), "T2 must draft the complete scenario set")

        seeded = generate_scenarios.assign_evidence_indices(generate_scenarios.create_scenarios())
        applied = generate_scenarios.apply_evidence_mapping(seeded, self.mapping)

        for scenario in applied:
            entry = next(e for e in self.mapping["scenarios"] if e["scenario_title"] == scenario["title"])
            by_id = {f["finding_id"]: f for f in entry["findings"]}
            for finding in scenario["expected_outcomes"]["key_findings"]:
                authored = by_id[finding["id"]]
                self.assertEqual(authored["evidence_log_indices"], finding["evidence_log_indices"])
                self.assertEqual(authored["verdict_support"], finding["verdict_support"])

        self.assertEqual(
            [],
            validator.validate_seed(applied, applied, mapped_titles=mapped_titles),
        )

    def test_real_mapping_passes_partial_validation(self):
        errors = validator.validate_mapping(
            self.mapping,
            generate_scenarios.create_scenarios(),
            require_complete=False,
        )
        self.assertEqual([], errors)

    def test_authored_mapping_chooses_semantic_logs(self):
        entry = self.mapping["scenarios"][0]
        self.assertEqual("Multi-Factor Authentication Bypass (Push Fatigue)", entry["scenario_title"])
        by_id = {f["finding_id"]: f for f in entry["findings"]}
        self.assertTrue({15, 16, 17}.issubset(set(by_id["d1_f1"]["evidence_log_indices"])))
        self.assertTrue(set(by_id["d1_f2"]["evidence_log_indices"]).intersection({18, 19, 20}))


if __name__ == "__main__":
    unittest.main(verbosity=2)
