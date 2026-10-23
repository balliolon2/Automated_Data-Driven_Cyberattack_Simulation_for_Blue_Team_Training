"""Pure validation for authored evidence mappings and generated scenario seeds.

The module is deliberately free of database and HTTP dependencies. The CLI at the
bottom is a thin wrapper around ``validate_mapping`` and ``validate_seed``.

Expand phase (ticket #17): mapping-mode validation accepts a partial mapping so
one authored Scenario can be proven without rewriting the other fourteen. Pass
``--require-complete`` (or ``require_complete=True``) to enforce that every base
Scenario is mapped. The later contract phase removes the partial mode.

Seed mode (also ticket #17) always compares the seed against the base Scenario
set from ``generate_scenarios.create_scenarios()``. While the sibling mapping is
an unapproved draft, seed mode relies on the seed's own ``verdict_support``
markers; once the mapping is approved, every mapped Scenario must carry complete
verdict markers in the seed.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

SCHEMA_VERSION = 1
SUPPORTED_SCHEMA_VERSIONS = {SCHEMA_VERSION}
VALID_VERDICT_SUPPORT = {"malicious", "benign", "supporting"}
MAX_RATIONALE_LENGTH = 300
DEFAULT_POINTS = 10
UNION_CAP_RATIO = 0.5
UNION_WARNING_MIN_RATIO = 0.4
HIGH_SEVERITIES = {"high", "critical"}
LOW_SEVERITIES = {"info", "low"}


def _is_integer(value):
    return isinstance(value, int) and not isinstance(value, bool)


def _scenario_prefix(title):
    return f"Scenario '{title}'"


def _finding_prefix(title, finding_id):
    return f"Scenario '{title}' / finding '{finding_id}'"


def _base_findings(base_scenario):
    if not isinstance(base_scenario, dict):
        return []
    expected = base_scenario.get("expected_outcomes", {})
    if not isinstance(expected, dict):
        return []
    findings = expected.get("key_findings", [])
    return findings if isinstance(findings, list) else []


def _base_findings_by_id(base_scenario):
    return {
        finding.get("id"): finding
        for finding in _base_findings(base_scenario)
        if isinstance(finding, dict)
    }


def _append_warning(warnings, message):
    if warnings is not None and message not in warnings:
        warnings.append(message)


def _severity_bucket(log):
    """Map a log's severity to ``high``, ``medium``, ``low``, or ``None``."""
    if not isinstance(log, dict):
        return None

    severity = log.get("severity")
    if isinstance(severity, str):
        normalized = severity.strip().lower()
        if normalized in HIGH_SEVERITIES:
            return "high"
        if normalized in LOW_SEVERITIES:
            return "low"
        if normalized == "medium":
            return "medium"

    level = log.get("severity_level")
    if _is_integer(level):
        if level >= 4:
            return "high"
        if level <= 2:
            return "low"
        return "medium"
    return None


def _evidence_severity_buckets(indices, logs):
    buckets = []
    if not isinstance(indices, list) or not isinstance(logs, list):
        return buckets
    for index in indices:
        if _is_integer(index) and 0 <= index < len(logs):
            bucket = _severity_bucket(logs[index])
            if bucket is not None:
                buckets.append(bucket)
    return buckets


def _validate_indices(indices, log_count, title, finding_id):
    prefix = _finding_prefix(title, finding_id)

    if indices is None:
        return [f"{prefix}: evidence_log_indices is empty (missing)"]
    if not isinstance(indices, list):
        return [f"{prefix}: evidence_log_indices must be a list of integers"]
    if not indices:
        return [f"{prefix}: evidence_log_indices is empty"]

    errors = []
    seen = set()
    for index in indices:
        if not _is_integer(index):
            errors.append(
                f"{prefix}: evidence_log_indices contains a non-integer value {index!r}"
            )
            continue
        if index < 0 or index >= log_count:
            errors.append(
                f"{prefix}: evidence_log_indices value {index} is out of range for {log_count} logs"
            )
        if index in seen:
            errors.append(f"{prefix}: duplicate evidence_log_index {index}")
        seen.add(index)

    return errors


def _validate_rationale(rationale, title, finding_id):
    prefix = _finding_prefix(title, finding_id)
    if not isinstance(rationale, str) or not rationale.strip():
        return [f"{prefix}: rationale is empty"]
    if len(rationale.strip()) > MAX_RATIONALE_LENGTH:
        return [
            f"{prefix}: rationale is {len(rationale.strip())} characters; "
            f"maximum is {MAX_RATIONALE_LENGTH}"
        ]
    return []


def _validate_verdict(verdict_support, title, finding_id):
    prefix = _finding_prefix(title, finding_id)
    if not isinstance(verdict_support, str) or verdict_support not in VALID_VERDICT_SUPPORT:
        return [
            f"{prefix}: verdict_support {verdict_support!r} is invalid; "
            f"expected one of {sorted(VALID_VERDICT_SUPPORT)}"
        ]
    return []


def _resolve_points(mapping_finding, base_finding):
    if isinstance(mapping_finding, dict) and "points" in mapping_finding:
        return mapping_finding["points"]
    if isinstance(base_finding, dict) and "points" in base_finding:
        return base_finding["points"]
    return DEFAULT_POINTS


def _validate_points(points, title, finding_id):
    if not _is_integer(points) or points <= 0:
        return [
            f"{_finding_prefix(title, finding_id)}: points must be a positive integer, got {points!r}"
        ]
    return []


def _validate_all_logs_rule(indices, log_count, title, finding_id):
    if log_count <= 0:
        return []
    integer_indices = {index for index in indices if _is_integer(index)}
    if integer_indices == set(range(log_count)):
        return [
            f"{_finding_prefix(title, finding_id)}: evidence cannot equal all logs in the scenario"
        ]
    return []


def _validate_union(union, log_count, title):
    cap = int(log_count * UNION_CAP_RATIO)
    if len(union) > cap:
        return [
            f"{_scenario_prefix(title)}: evidence union {len(union)}/{log_count} "
            f"exceeds the 50% cap ({cap} logs)"
        ]
    return []


def _warn_shared_evidence(warnings, index_to_findings, title):
    for index in sorted(index_to_findings):
        finding_ids = index_to_findings[index]
        if len(finding_ids) > 1:
            _append_warning(
                warnings,
                f"{_scenario_prefix(title)}: log index {index} is shared evidence for "
                f"findings {', '.join(sorted(finding_ids))}",
            )


def _warn_union_near_cap(warnings, union, log_count, title):
    if log_count <= 0 or not union:
        return
    ratio = len(union) / log_count
    if UNION_WARNING_MIN_RATIO <= ratio <= UNION_CAP_RATIO:
        _append_warning(
            warnings,
            f"{_scenario_prefix(title)}: evidence union {len(union)}/{log_count} "
            f"({ratio:.0%}) is between 40% and 50% of logs",
        )


def _warn_verdict_severity(warnings, title, is_true_positive, findings, logs):
    if is_true_positive is None:
        return

    for finding in findings:
        if not isinstance(finding, dict):
            continue
        finding_id = finding.get("finding_id") or finding.get("id")
        buckets = _evidence_severity_buckets(finding.get("evidence_log_indices"), logs)
        if not buckets:
            continue
        if is_true_positive is False and all(bucket == "high" for bucket in buckets):
            _append_warning(
                warnings,
                f"{_finding_prefix(title, finding_id)}: False Positive finding evidence "
                "is entirely high/critical severity (possible alert-only mapping)",
            )
        if is_true_positive is True and all(bucket == "low" for bucket in buckets):
            _append_warning(
                warnings,
                f"{_finding_prefix(title, finding_id)}: True Positive finding evidence "
                "is entirely info/low severity (unlikely to prove malice)",
            )


def _validate_finding_evidence(finding, log_count, title):
    finding_id = finding.get("finding_id")
    errors = []
    indices = finding.get("evidence_log_indices")
    errors.extend(_validate_indices(indices, log_count, title, finding_id))
    if isinstance(indices, list):
        errors.extend(_validate_all_logs_rule(indices, log_count, title, finding_id))
    return errors


def validate_mapping(mapping, base_scenarios, *, require_complete=True, warnings=None):
    """Validate a mapping file against the base scenarios.

    Returns a list of human-readable errors. An empty list means the mapping is
    structurally valid. ``warnings`` may be a list that receives non-fatal
    authoring warnings. Approval is intentionally not checked here; ticket #19
    adds the approval hash gate.
    """
    errors = []

    if not isinstance(mapping, dict):
        return ["Mapping root must be a JSON object"]

    schema_version = mapping.get("schema_version")
    if not _is_integer(schema_version) or schema_version not in SUPPORTED_SCHEMA_VERSIONS:
        errors.append(
            "Mapping schema_version is missing or unsupported; "
            f"expected {SCHEMA_VERSION}, got {schema_version!r}"
        )

    scenario_entries = mapping.get("scenarios")
    if not isinstance(scenario_entries, list):
        errors.append("Mapping 'scenarios' must be a list")
        return errors

    base_by_title = {}
    for base_scenario in base_scenarios:
        if isinstance(base_scenario, dict):
            base_by_title[base_scenario.get("title")] = base_scenario

    seen_titles = set()
    for entry in scenario_entries:
        if not isinstance(entry, dict):
            errors.append("Each mapping scenario entry must be a JSON object")
            continue

        title = entry.get("scenario_title")
        if not isinstance(title, str) or not title:
            errors.append("A mapping scenario entry is missing scenario_title")
            continue

        if title in seen_titles:
            errors.append(f"{_scenario_prefix(title)}: duplicate scenario title in mapping")
            continue
        seen_titles.add(title)

        base_scenario = base_by_title.get(title)
        if base_scenario is None:
            errors.append(f"{_scenario_prefix(title)}: scenario is not in base; extra scenario")
            continue

        findings = entry.get("findings")
        if not isinstance(findings, list):
            errors.append(f"{_scenario_prefix(title)}: findings must be a list")
            continue

        base_by_id = _base_findings_by_id(base_scenario)
        mapped_by_id = {}
        union = set()
        index_to_findings = {}
        logs = base_scenario.get("initial_logs", [])
        log_count = len(logs) if isinstance(logs, list) else 0

        for finding in findings:
            if not isinstance(finding, dict):
                errors.append(f"{_scenario_prefix(title)}: each finding entry must be a JSON object")
                continue

            finding_id = finding.get("finding_id")
            if not isinstance(finding_id, str) or not finding_id:
                errors.append(f"{_scenario_prefix(title)}: a finding entry is missing finding_id")
                continue

            if finding_id in mapped_by_id:
                errors.append(
                    f"{_finding_prefix(title, finding_id)}: duplicate finding in mapping"
                )
                continue
            mapped_by_id[finding_id] = finding

            base_finding = base_by_id.get(finding_id)
            if base_finding is None:
                errors.append(
                    f"{_finding_prefix(title, finding_id)}: finding is not in base; extra finding"
                )
                continue

            errors.extend(_validate_finding_evidence(finding, log_count, title))

            indices = finding.get("evidence_log_indices")
            if isinstance(indices, list):
                for index in indices:
                    if _is_integer(index):
                        union.add(index)
                        index_to_findings.setdefault(index, set()).add(finding_id)

            errors.extend(_validate_rationale(finding.get("rationale"), title, finding_id))
            errors.extend(_validate_verdict(finding.get("verdict_support"), title, finding_id))
            errors.extend(
                _validate_points(_resolve_points(finding, base_finding), title, finding_id)
            )

        for base_id in base_by_id:
            if base_id not in mapped_by_id:
                errors.append(
                    f"{_finding_prefix(title, base_id)}: missing finding in mapping"
                )

        errors.extend(_validate_union(union, log_count, title))

        is_true_positive = base_scenario.get("is_true_positive")
        supports = [
            finding.get("verdict_support")
            for finding in findings
            if isinstance(finding, dict)
        ]
        if is_true_positive is True and "malicious" not in supports:
            errors.append(
                f"{_scenario_prefix(title)}: True Positive requires at least one malicious finding"
            )
        if is_true_positive is False and "benign" not in supports:
            errors.append(
                f"{_scenario_prefix(title)}: False Positive requires at least one benign finding"
            )

        _warn_shared_evidence(warnings, index_to_findings, title)
        _warn_union_near_cap(warnings, union, log_count, title)
        _warn_verdict_severity(
            warnings, title, is_true_positive, findings, logs
        )

    if require_complete:
        for title in base_by_title:
            if title not in seen_titles:
                errors.append(f"{_scenario_prefix(title)}: missing scenario in mapping")

    return errors


def validate_seed(seed_scenarios, base_scenarios, *, mapped_titles=None, warnings=None):
    """Validate a seed against the base Scenario set and finding IDs.

    ``base_scenarios`` is the authoritative set from
    ``generate_scenarios.create_scenarios()``. Missing, extra, and duplicate
    scenarios or findings are errors, so truncating the seed cannot bypass the
    gate. ``mapped_titles`` names the scenarios that have an authored mapping;
    those must carry complete ``verdict_support`` markers. Unmapped scenarios
    remain tolerated while the expand-phase chunked fallback is in use.
    """
    errors = []

    if not isinstance(seed_scenarios, list):
        return ["Seed root must be a JSON array of scenarios"]

    base_by_title = {}
    for base_scenario in base_scenarios if isinstance(base_scenarios, list) else []:
        if isinstance(base_scenario, dict):
            base_by_title[base_scenario.get("title")] = base_scenario

    mapped_titles = set(mapped_titles or ())

    seen_titles = set()
    for scenario in seed_scenarios:
        if not isinstance(scenario, dict):
            errors.append("Each seed scenario must be a JSON object")
            continue

        title = scenario.get("title")
        if not isinstance(title, str) or not title:
            errors.append("A seed scenario is missing its title")
            continue

        if title in seen_titles:
            errors.append(f"{_scenario_prefix(title)}: duplicate scenario title in seed")
            continue
        seen_titles.add(title)

        base_scenario = base_by_title.get(title)
        if base_scenario is None:
            errors.append(f"{_scenario_prefix(title)}: scenario is not in base; extra scenario")
            continue

        expected = scenario.get("expected_outcomes")
        if not isinstance(expected, dict):
            errors.append(
                f"{_scenario_prefix(title)}: expected_outcomes must be a JSON object"
            )
            continue

        findings = expected.get("key_findings")
        if not isinstance(findings, list):
            errors.append(
                f"{_scenario_prefix(title)}: expected_outcomes.key_findings must be a list"
            )
            continue
        if not findings:
            errors.append(f"{_scenario_prefix(title)}: key_findings is empty")
            continue

        base_by_id = _base_findings_by_id(base_scenario)
        seen_ids = set()
        logs = scenario.get("initial_logs", [])
        logs = logs if isinstance(logs, list) else []
        log_count = len(logs)
        is_mapped = title in mapped_titles
        has_any_verdict_support = any(
            isinstance(finding, dict) and "verdict_support" in finding for finding in findings
        )
        require_verdicts = is_mapped or has_any_verdict_support

        union = set()
        index_to_findings = {}
        supports = []

        for finding in findings:
            if not isinstance(finding, dict):
                errors.append(f"{_scenario_prefix(title)}: each key finding must be a JSON object")
                continue

            finding_id = finding.get("id")
            if not isinstance(finding_id, str) or not finding_id:
                errors.append(f"{_scenario_prefix(title)}: a key finding is missing its id")
                continue

            if finding_id in seen_ids:
                errors.append(
                    f"{_finding_prefix(title, finding_id)}: duplicate finding in seed"
                )
                continue
            seen_ids.add(finding_id)

            if finding_id not in base_by_id:
                errors.append(
                    f"{_finding_prefix(title, finding_id)}: finding is not in base; extra finding"
                )
                continue

            indices = finding.get("evidence_log_indices")
            errors.extend(_validate_indices(indices, log_count, title, finding_id))
            if isinstance(indices, list):
                for index in indices:
                    if _is_integer(index):
                        union.add(index)
                        index_to_findings.setdefault(index, set()).add(finding_id)
                errors.extend(_validate_all_logs_rule(indices, log_count, title, finding_id))

            errors.extend(
                _validate_points(finding.get("points", DEFAULT_POINTS), title, finding_id)
            )

            if require_verdicts:
                errors.extend(
                    _validate_verdict(finding.get("verdict_support"), title, finding_id)
                )
            supports.append(finding.get("verdict_support"))

        for base_id in base_by_id:
            if base_id not in seen_ids:
                errors.append(
                    f"{_finding_prefix(title, base_id)}: missing finding in seed"
                )

        errors.extend(_validate_union(union, log_count, title))

        if require_verdicts:
            if scenario.get("is_true_positive") is True and "malicious" not in supports:
                errors.append(
                    f"{_scenario_prefix(title)}: True Positive requires at least one malicious finding"
                )
            if scenario.get("is_true_positive") is False and "benign" not in supports:
                errors.append(
                    f"{_scenario_prefix(title)}: False Positive requires at least one benign finding"
                )

        _warn_shared_evidence(warnings, index_to_findings, title)
        _warn_union_near_cap(warnings, union, log_count, title)
        _warn_verdict_severity(
            warnings, title, scenario.get("is_true_positive"), findings, logs
        )

    for title in base_by_title:
        if title not in seen_titles:
            errors.append(f"{_scenario_prefix(title)}: missing scenario in seed")

    return errors


def _load_json(path):
    with open(path, "r", encoding="utf-8") as handle:
        return json.load(handle)


def _load_base_scenarios():
    scripts_dir = str(Path(__file__).resolve().parent)
    if scripts_dir not in sys.path:
        sys.path.insert(0, scripts_dir)
    from generate_scenarios import create_scenarios

    return create_scenarios()


def _load_mapped_titles(path=None, seed=None):
    """Load mapped scenario titles from the sibling mapping file, if present.

    Seed mode is phase-aware. While the mapping is still an unapproved draft
    (the expand phase), the committed seed may not have the draft applied yet,
    so seed mode relies on the seed's own verdict markers rather than demanding
    verdicts for every drafted Scenario. Once the mapping is approved, all
    mapped Scenarios must carry verdict markers in the seed.
    """
    mapping_path = Path(path) if path else Path(__file__).resolve().parent / "evidence_mapping.json"
    if not mapping_path.is_file():
        return set()
    try:
        mapping = _load_json(mapping_path)
    except (OSError, json.JSONDecodeError):
        return set()
    if not isinstance(mapping, dict):
        return set()
    entries = mapping.get("scenarios")
    if not isinstance(entries, list):
        return set()
    titles = {
        entry["scenario_title"]
        for entry in entries
        if isinstance(entry, dict) and isinstance(entry.get("scenario_title"), str)
    }
    if seed is not None:
        approval = mapping.get("approval")
        approved = isinstance(approval, dict) and approval.get("approved") is True
        if not approved:
            return set()
    return titles


def _print_warnings(warnings):
    for warning in warnings:
        print(f"Warning: {warning}", file=sys.stderr)


def main(argv=None):
    parser = argparse.ArgumentParser(description="Validate an evidence mapping or a generated scenario seed.")
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--mapping", help="Path to evidence_mapping.json")
    mode.add_argument("--seed", help="Path to scenarios_seed.json")
    parser.add_argument(
        "--require-complete",
        action="store_true",
        help="Mapping mode only: require every base scenario to be mapped (contract phase).",
    )
    args = parser.parse_args(argv)

    label = args.mapping or args.seed
    warnings = []

    try:
        if args.mapping:
            try:
                mapping = _load_json(args.mapping)
            except (OSError, json.JSONDecodeError) as exc:
                print(f"Could not read mapping {args.mapping}: {exc}", file=sys.stderr)
                return 1
            base_scenarios = _load_base_scenarios()
            errors = validate_mapping(
                mapping,
                base_scenarios,
                require_complete=args.require_complete,
                warnings=warnings,
            )
            if not args.require_complete:
                mapped = {
                    entry.get("scenario_title")
                    for entry in mapping.get("scenarios", [])
                    if isinstance(entry, dict)
                } if isinstance(mapping, dict) else set()
                unmapped = [
                    scenario.get("title")
                    for scenario in base_scenarios
                    if scenario.get("title") not in mapped
                ]
                if unmapped:
                    print(
                        f"Expand phase: {len(unmapped)} scenario(s) use the existing fallback "
                        f"evidence behavior: {', '.join(unmapped)}",
                        file=sys.stderr,
                    )
        else:
            try:
                seed = _load_json(args.seed)
            except (OSError, json.JSONDecodeError) as exc:
                print(f"Could not read seed {args.seed}: {exc}", file=sys.stderr)
                return 1
            base_scenarios = _load_base_scenarios()
            mapped_titles = _load_mapped_titles(seed=seed)
            errors = validate_seed(
                seed,
                base_scenarios,
                mapped_titles=mapped_titles,
                warnings=warnings,
            )
    except (TypeError, ValueError, AttributeError, KeyError, IndexError) as exc:
        print(
            f"Validation failed for {label}: malformed input: {exc}",
            file=sys.stderr,
        )
        return 1

    _print_warnings(warnings)

    if errors:
        print(f"Validation failed for {label} with {len(errors)} error(s):", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print(f"Validation passed: {label}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
