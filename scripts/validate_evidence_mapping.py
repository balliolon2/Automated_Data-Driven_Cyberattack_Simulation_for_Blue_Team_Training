"""Pure validation for authored evidence mappings and generated scenario seeds.

The module is deliberately free of database and HTTP dependencies. The CLI at the
bottom is a thin wrapper around ``validate_mapping`` and ``validate_seed``.

The contract is complete: every base Scenario and finding must be mapped, and
mapping mode requires a human approval whose ``mapping_sha256`` matches the
canonical ``scenarios`` array. Any later edit to the mapping invalidates that
hash and blocks seeding.

Seed mode always compares the seed against the base Scenario set from
``generate_scenarios.create_scenarios()`` and requires authored evidence plus
valid verdict support for every finding. There is no legacy tolerance.
"""

from __future__ import annotations

import argparse
import hashlib
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


def canonical_scenarios(mapping):
    """Return the canonical ``scenarios`` array used for approval hashing.

    Scenario entries are ordered by title, findings by ID, and evidence indices
    by value. Key order is normalized by ``sort_keys=True`` in
    :func:`mapping_sha256`. Only the ``scenarios`` array participates; approval
    metadata and provenance fields never affect the hash.
    """
    scenarios = mapping.get("scenarios") if isinstance(mapping, dict) else None
    if not isinstance(scenarios, list):
        return []

    normalized = []
    for entry in scenarios:
        if not isinstance(entry, dict):
            normalized.append(entry)
            continue

        entry_copy = dict(entry)
        findings = entry_copy.get("findings")
        if isinstance(findings, list):
            finding_copies = []
            for finding in findings:
                if not isinstance(finding, dict):
                    finding_copies.append(finding)
                    continue
                finding_copy = dict(finding)
                indices = finding_copy.get("evidence_log_indices")
                if isinstance(indices, list) and all(_is_integer(value) for value in indices):
                    finding_copy["evidence_log_indices"] = sorted(indices)
                finding_copies.append(finding_copy)
            finding_copies.sort(
                key=lambda item: item.get("finding_id")
                if isinstance(item, dict) and isinstance(item.get("finding_id"), str)
                else ""
            )
            entry_copy["findings"] = finding_copies
        normalized.append(entry_copy)

    normalized.sort(
        key=lambda item: item.get("scenario_title")
        if isinstance(item, dict) and isinstance(item.get("scenario_title"), str)
        else ""
    )
    return normalized


def mapping_sha256(mapping):
    """Return the SHA-256 of the canonicalized ``scenarios`` array."""
    if not isinstance(mapping, dict):
        raise ValueError("Mapping root must be a JSON object")
    canonical = json.dumps(
        canonical_scenarios(mapping),
        sort_keys=True,
        separators=(",", ":"),
        ensure_ascii=False,
    )
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


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


def _validate_approval(mapping):
    """Require a complete human approval whose hash matches the scenarios."""
    approval = mapping.get("approval")
    if not isinstance(approval, dict):
        return ["Mapping approval block is missing"]

    errors = []
    if approval.get("approved") is not True:
        errors.append("Mapping is not approved; approval.approved must be true")
    approved_by = approval.get("approved_by")
    if not isinstance(approved_by, str) or not approved_by.strip():
        errors.append("Mapping approval is missing approved_by")
    approved_at = approval.get("approved_at")
    if not isinstance(approved_at, str) or not approved_at.strip():
        errors.append("Mapping approval is missing approved_at")

    expected = mapping_sha256(mapping)
    recorded = approval.get("mapping_sha256")
    if not isinstance(recorded, str) or not recorded.strip():
        errors.append("Mapping approval is missing mapping_sha256")
    elif recorded.strip().lower() != expected:
        errors.append(
            "Mapping approval hash mismatch: mapping_sha256 does not match the "
            "current scenarios content"
        )
    return errors


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


def validate_mapping(
    mapping,
    base_scenarios,
    *,
    require_complete=True,
    require_approval=False,
    warnings=None,
):
    """Validate a mapping file against the base scenarios.

    Returns a list of human-readable errors. An empty list means the mapping is
    structurally valid. ``warnings`` may be a list that receives non-fatal
    authoring warnings. ``require_approval`` turns on the human approval hash
    gate; mapping mode and the seed pipeline always enable it, while drafting
    validates individual unapproved entries.
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

    if require_approval:
        errors.extend(_validate_approval(mapping))

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


def validate_seed(seed_scenarios, base_scenarios, *, warnings=None):
    """Validate a seed against the base Scenario set and finding IDs.

    ``base_scenarios`` is the authoritative set from
    ``generate_scenarios.create_scenarios()``. Missing, extra, and duplicate
    scenarios or findings are errors, so truncating or hand-editing the seed
    cannot bypass the gate. Every finding must carry non-empty, in-range, unique
    evidence, a positive point value, and valid verdict support; every Scenario's
    evidence union must stay at or below 50% of its logs.
    """
    errors = []

    if not isinstance(seed_scenarios, list):
        return ["Seed root must be a JSON array of scenarios"]

    base_by_title = {}
    for base_scenario in base_scenarios if isinstance(base_scenarios, list) else []:
        if isinstance(base_scenario, dict):
            base_by_title[base_scenario.get("title")] = base_scenario

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
        is_true_positive = base_scenario.get("is_true_positive")

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
        _warn_verdict_severity(warnings, title, is_true_positive, findings, logs)

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
        help="Accepted for compatibility; mapping mode always requires complete coverage.",
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
                require_complete=True,
                require_approval=True,
                warnings=warnings,
            )
        else:
            try:
                seed = _load_json(args.seed)
            except (OSError, json.JSONDecodeError) as exc:
                print(f"Could not read seed {args.seed}: {exc}", file=sys.stderr)
                return 1
            base_scenarios = _load_base_scenarios()
            errors = validate_seed(
                seed,
                base_scenarios,
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
