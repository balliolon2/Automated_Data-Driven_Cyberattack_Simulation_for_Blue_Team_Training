"""Draft authored evidence mappings for the static Scenario set.

The normal path asks an OpenAI-compatible chat-completions model for one mapping
entry per Scenario. Every response is validated with
``validate_evidence_mapping``; invalid responses are retried with the exact
validator errors. A Scenario that still fails is recorded as failed, written to
the report, and makes the command exit non-zero. There is deliberately no
positional-chunking fallback.

``--offline`` is a separate, explicit authoring path. It matches each finding's
keyword/entity rules against the actual log message text and records the
matched log indices. It never reads a fixed index range and never calls
``generate_scenarios.assign_evidence_indices``.
"""

from __future__ import annotations

import argparse
import concurrent.futures
import json
import os
import re
import sys
import urllib.error
import urllib.request
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
if str(SCRIPT_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPT_DIR))

import validate_evidence_mapping as validator  # noqa: E402

DEFAULT_BASE_URL = "https://api.openai.com/v1"
DEFAULT_MODEL = "gpt-4o-mini"
PROMPT_VERSION = "evidence-mapping-v1"
SCHEMA_VERSION = 1
DEFAULT_TIMEOUT_SECONDS = 90
MAX_REPORT_SNIPPET = 140
MAX_OFFLINE_EVIDENCE = 4

SYSTEM_PROMPT = (
    "You are a CompTIA Security+ 701 blue-team scenario editor. "
    "Assign telemetry logs to key findings. Return JSON only."
)

AUTHORING_RULES = """Authoring rules:
1. Pick the minimum decisive set: the smallest set of logs that would convince a SOC analyst.
2. Never select a contiguous range merely because it sits between noise blocks.
3. True Positive evidence must directly support the malicious conclusion.
4. False Positive evidence must establish the benign/authorized explanation (approved change, HR travel, scheduled maintenance, authorized scan). Do not select only the high-severity alert that opened the case.
5. Do not select a log just because it is high severity.
6. Do not select generic noise or baseline traffic.
7. Overlap between findings is allowed only when a log genuinely proves both.
8. The union of all findings' evidence indices must be at most 50% of the logs.
9. A True Positive needs at least one finding with verdict_support "malicious"; a False Positive needs at least one with "benign".
10. Every finding needs evidence and a one-sentence rationale of at most 300 characters.
11. Use only the provided log indices. Never invent an index or a finding id.
12. For a False Positive, the verdict-supporting finding must point to the log that proves the benign or authorized explanation."""

OUTPUT_SCHEMA = """Return exactly one JSON object with this shape and no Markdown:
{
  "scenario_title": "<the provided title>",
  "findings": [
    {
      "finding_id": "<one of the provided finding ids>",
      "evidence_log_indices": [<provided integer indices>],
      "rationale": "<one sentence, <= 300 characters>",
      "verdict_support": "malicious" | "benign" | "supporting"
    }
  ]
}
Return one entry for every provided finding and no others."""

BENIGN_CONTEXT_SOURCES = {
    "change_management",
    "change_mgmt",
    "hr_portal",
    "hr_system",
}
BENIGN_CONTEXT_TOKENS = (
    "change request",
    "approved change",
    "out of office",
    "travel notification",
    "scheduled",
    "authorized",
    "compliance",
)


class DraftError(Exception):
    """A response could not be parsed into a scenario mapping entry."""


class LLMError(Exception):
    """The model endpoint was unavailable or returned an unusable response."""


@dataclass
class ScenarioResult:
    scenario: dict
    entry: dict | None
    errors: list = field(default_factory=list)
    attempts: int = 0
    drafted_by: str = "llm"
    notes: list = field(default_factory=list)

    @property
    def failed(self):
        return self.entry is None


def _utc_now():
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def _truncate(text, limit):
    text = " ".join(str(text).split())
    if len(text) <= limit:
        return text
    return text[: max(0, limit - 1)].rstrip() + "..."


def load_env_file(path=None):
    """Load a simple KEY=VALUE .env file without printing any values."""
    env_path = Path(path) if path else REPO_ROOT / ".env"
    values = {}
    if not env_path.is_file():
        return values
    try:
        lines = env_path.read_text(encoding="utf-8").splitlines()
    except OSError:
        return values
    for line in lines:
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        if line.startswith("export "):
            line = line[len("export ") :].lstrip()
        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in {"'", '"'}:
            value = value[1:-1]
        if key:
            values[key] = value
    return values


def resolve_llm_config(environ=None, dotenv=None):
    """Resolve model settings; real environment variables win over .env."""
    environ = os.environ if environ is None else environ
    dotenv = load_env_file() if dotenv is None else dotenv

    def lookup(name, default=None):
        value = environ.get(name)
        if value is None or not str(value).strip():
            value = dotenv.get(name)
        if value is None or not str(value).strip():
            return default
        return str(value).strip()

    return {
        "api_key": lookup("LLM_API_KEY"),
        "base_url": lookup("LLM_BASE_URL", DEFAULT_BASE_URL),
        "model": lookup("LLM_MODEL", DEFAULT_MODEL),
    }


def _log_search_text(log):
    if not isinstance(log, dict):
        return ""
    keys = (
        "message",
        "source",
        "src_ip",
        "dst_ip",
        "user",
        "host",
        "action",
        "url",
        "query",
        "domain",
    )
    return " ".join(str(log.get(key, "")) for key in keys)


def scenario_prompt_payload(scenario):
    """Build the indexed Scenario payload shown to the model or matcher."""
    logs = scenario.get("initial_logs", [])
    logs = logs if isinstance(logs, list) else []
    findings = scenario.get("expected_outcomes", {}).get("key_findings", [])
    findings = findings if isinstance(findings, list) else []
    return {
        "title": scenario.get("title"),
        "description": scenario.get("description"),
        "is_true_positive": scenario.get("is_true_positive"),
        "tp_fp_explanation": scenario.get("tp_fp_explanation"),
        "key_findings": [
            {
                "id": finding.get("id"),
                "description": finding.get("description"),
                "domain_id": finding.get("domain_id"),
                "points": finding.get("points", 10),
            }
            for finding in findings
            if isinstance(finding, dict)
        ],
        "initial_logs": [
            dict({"index": index}, **log) if isinstance(log, dict) else {"index": index}
            for index, log in enumerate(logs)
        ],
    }


def build_user_prompt(payload):
    return (
        "Scenario input:\n"
        + json.dumps(payload, indent=2, ensure_ascii=False)
        + "\n\n"
        + AUTHORING_RULES
        + "\n\n"
        + OUTPUT_SCHEMA
        + "\n"
    )


def build_retry_prompt(payload, errors):
    return (
        build_user_prompt(payload)
        + "\nThe previous response was invalid. Fix exactly these validation errors:\n"
        + "\n".join(f"- {error}" for error in errors)
        + "\nReturn the corrected JSON object only. Do not change any other choice unless the error requires it.\n"
    )


def _strip_markdown_fences(text):
    stripped = text.strip()
    if not stripped.startswith("```"):
        return stripped
    stripped = re.sub(r"^```[a-zA-Z0-9_-]*\s*", "", stripped)
    stripped = re.sub(r"\s*```$", "", stripped)
    return stripped.strip()


def parse_llm_json(text):
    if not isinstance(text, str) or not text.strip():
        raise DraftError("response was empty")
    cleaned = _strip_markdown_fences(text)
    try:
        return json.loads(cleaned)
    except json.JSONDecodeError:
        start = cleaned.find("{")
        end = cleaned.rfind("}")
        if start != -1 and end > start:
            try:
                return json.loads(cleaned[start : end + 1])
            except json.JSONDecodeError:
                pass
        raise DraftError(f"response was not valid JSON: {cleaned[:160]}")


def normalize_entry(draft, scenario):
    """Normalize a model response into one mapping entry for this Scenario."""
    if not isinstance(draft, dict):
        raise DraftError("response must be a JSON object")
    if "scenarios" in draft:
        entries = draft.get("scenarios")
        if not isinstance(entries, list) or len(entries) != 1:
            raise DraftError("response.scenarios must contain exactly one scenario entry")
        draft = entries[0]
        if not isinstance(draft, dict):
            raise DraftError("response.scenarios[0] must be a JSON object")

    findings = draft.get("findings")
    if not isinstance(findings, list):
        raise DraftError("response.findings must be a list")

    normalized_findings = []
    for finding in findings:
        if not isinstance(finding, dict):
            normalized_findings.append(finding)
            continue
        normalized_findings.append(
            {
                "finding_id": finding.get("finding_id"),
                "evidence_log_indices": finding.get("evidence_log_indices"),
                "rationale": finding.get("rationale"),
                "verdict_support": finding.get("verdict_support"),
            }
        )

    return {
        "scenario_title": scenario.get("title"),
        "is_true_positive": scenario.get("is_true_positive"),
        "findings": normalized_findings,
    }


def validate_scenario_entry(entry, scenario, warnings=None):
    mapping = {"schema_version": SCHEMA_VERSION, "scenarios": [entry]}
    return validator.validate_mapping(
        mapping, [scenario], require_complete=True, warnings=warnings
    )


def is_benign_context_log(log):
    if not isinstance(log, dict):
        return False
    source = str(log.get("source", "")).strip().lower()
    if source in BENIGN_CONTEXT_SOURCES:
        return True
    message = str(log.get("message", "")).lower()
    return any(token in message for token in BENIGN_CONTEXT_TOKENS)


def _is_integer(value):
    return isinstance(value, int) and not isinstance(value, bool)


def _offline_rules():
    """Explicit content patterns per finding, matched against real log text.

    The tuple key is (Scenario title, finding id). Each pattern is a literal
    substring; ``max`` limits how many matching rows that pattern contributes.
    Indices are computed from the matches, never from a position in the array.
    """
    return {
        (
            "Multi-Factor Authentication Bypass (Push Fatigue)",
            "d1_f1",
        ): [
            {"pattern": "MFA push sent to user 'ajones' - rejected by user", "max": 3},
            {"pattern": "MFA push sent to user 'ajones' - approved", "max": 1},
        ],
        (
            "Multi-Factor Authentication Bypass (Push Fatigue)",
            "d1_f2",
        ): [
            {"pattern": "hijacked session token", "max": 1},
            {"pattern": "MFA push sent to user 'ajones' - approved", "max": 1},
        ],
        (
            "Atypical Travel VPN Connection Alert",
            "d1_f3",
        ): [{"pattern": "Out of Office / Travel notification", "max": 1}],
        (
            "Atypical Travel VPN Connection Alert",
            "d1_f4",
        ): [
            {"pattern": "MFA prompt completed successfully", "max": 1},
            {"pattern": "Device posture check passed", "max": 1},
        ],
        (
            "Unauthorized Privileged Access / DACL Abuse",
            "d1_f5",
        ): [{"pattern": "DACL changed on CN=Payroll_Confidential", "max": 1}],
        (
            "Unauthorized Privileged Access / DACL Abuse",
            "d1_f6",
        ): [
            {"pattern": "Access granted: read folder 'Payroll_Confidential'", "max": 1},
            {"pattern": "File read: 'salary_matrix_2026.xlsx'", "max": 1},
        ],
        (
            "SQL Injection & Database Exfiltration",
            "d2_f1",
        ): [{"pattern": "UNION SELECT", "max": 2}],
        (
            "SQL Injection & Database Exfiltration",
            "d2_f2",
        ): [
            {"pattern": "Database output payload size: 1.2MB", "max": 1},
            {"pattern": "Response size 1205300 bytes", "max": 1},
        ],
        (
            "Internal Port Scan and SYN Flood Alarm",
            "d2_f3",
        ): [{"pattern": "CR-4099 Approved", "max": 1}],
        (
            "Internal Port Scan and SYN Flood Alarm",
            "d2_f4",
        ): [
            {"pattern": "SYN Flood Scan from 10.0.99.12", "max": 1},
            {
                "pattern": "TCP Port scan detected from 10.0.99.12 to 10.0.1.0 on Port 80",
                "max": 1,
            },
        ],
        (
            "Ransomware Execution & Volume Shadow Copy Deletion",
            "d2_f5",
        ): [
            {"pattern": "vssadmin.exe delete shadows", "max": 1},
            {"pattern": "wmic shadowcopy delete", "max": 1},
        ],
        (
            "Ransomware Execution & Volume Shadow Copy Deletion",
            "d2_f6",
        ): [
            {"pattern": "renamed to extension '.locked'", "max": 1},
            {"pattern": "ransom_decrypt_instruction.txt", "max": 1},
        ],
        (
            "VLAN Segmentation Failure / Guest Network Bypass",
            "d3_f1",
        ): [{"pattern": "192.168.10.145:49810 -> 10.0.50.12:5432", "max": 1}],
        (
            "VLAN Segmentation Failure / Guest Network Bypass",
            "d3_f2",
        ): [{"pattern": "Database login failed", "max": 2}],
        (
            "Exorbitant Bandwidth Consumption on Guest Network",
            "d3_f3",
        ): [
            {"pattern": "Excessive outbound bandwidth", "max": 1},
            {"pattern": "CONNECT netflix.com", "max": 1},
            {"pattern": "CONNECT youtube.com", "max": 1},
        ],
        (
            "AWS S3 Cloud Storage Misconfiguration",
            "d3_f4",
        ): [{"pattern": "PutBucketAcl", "max": 1}],
        (
            "AWS S3 Cloud Storage Misconfiguration",
            "d3_f5",
        ): [
            {
                "pattern": "GetObject: S3Bucket=customer-contracts-prod Key=contract_",
                "max": 2,
            }
        ],
        (
            "DNS Tunneling / Data Exfiltration via DNS TXT Queries",
            "d4_f1",
        ): [{"pattern": "Type=TXT", "max": 2}],
        (
            "DNS Tunneling / Data Exfiltration via DNS TXT Queries",
            "d4_f2",
        ): [
            {"pattern": "c2Vj", "max": 1},
            {"pattern": "cGFz", "max": 1},
        ],
        (
            "Anomalous Volume of NTP Network Requests",
            "d4_f3",
        ): [{"pattern": "CR-3212", "max": 1}],
        (
            "Spear Phishing leading to Credential Harvesting",
            "d4_f4",
        ): [{"pattern": "secure-office365-portal.com", "max": 2}],
        ("Spear Phishing leading to Credential Harvesting", "d4_f5"): [
            {"pattern": "adm_backdoor", "max": 1}
        ],
        ("Insider Threat: Customer Database Exfiltration", "d5_f1"): [
            {"pattern": "Bulk report exported", "max": 1}
        ],
        ("Insider Threat: Customer Database Exfiltration", "d5_f2"): [
            {"pattern": "personal-gmail.com", "max": 1},
            {"pattern": "File deleted: Customer_Billing_Details_All.csv", "max": 1},
        ],
        ("Anomalous Cloud Data Archiving Alert", "d5_f3"): [
            {"pattern": "CR-9022", "max": 1},
            {"pattern": "Glacier transfer complete", "max": 1},
        ],
        ("Rogue Wireless Access Point Detection", "d5_f4"): [
            {"pattern": "Port Security violation", "max": 1},
            {"pattern": "8 unauthenticated devices requested IPs", "max": 1},
        ],
        ("Rogue Wireless Access Point Detection", "d5_f5"): [
            {"pattern": "Unauthorized SSIDs detected", "max": 1}
        ],
    }


OFFLINE_STOPWORDS = {
    "a",
    "an",
    "and",
    "are",
    "as",
    "at",
    "be",
    "by",
    "for",
    "from",
    "in",
    "is",
    "it",
    "of",
    "on",
    "or",
    "that",
    "the",
    "this",
    "to",
    "was",
    "were",
    "with",
    "identified",
    "found",
    "discovered",
    "correlated",
    "verified",
}


def _derive_offline_indices(finding, logs):
    """Content fallback for a finding without an explicit offline rule."""
    words = {
        word.lower()
        for word in re.findall(r"[A-Za-z0-9_.-]{4,}", " ".join(
            str(finding.get(key, ""))
            for key in ("description", "explanation")
        ))
        if word.lower() not in OFFLINE_STOPWORDS
    }
    if not words:
        return []
    scored = []
    for index, log in enumerate(logs):
        text = _log_search_text(log).lower()
        score = sum(1 for word in words if word in text)
        if score:
            scored.append((score, index))
    scored.sort(key=lambda item: (-item[0], item[1]))
    return sorted(index for _, index in scored[:2])


def _match_offline_indices(finding, logs):
    title = finding.get("_scenario_title")
    finding_id = finding.get("id")
    rules = _offline_rules().get((title, finding_id))
    matched = set()

    if rules:
        for rule in rules:
            pattern = rule["pattern"].lower()
            limit = rule.get("max", 1)
            count = 0
            for index, log in enumerate(logs):
                if pattern in _log_search_text(log).lower():
                    matched.add(index)
                    count += 1
                    if count >= limit:
                        break
    else:
        matched.update(_derive_offline_indices(finding, logs))

    return sorted(matched)[:MAX_OFFLINE_EVIDENCE]


def _offline_support(scenario, finding, indices):
    if scenario.get("is_true_positive") is True:
        return "malicious"

    logs = scenario.get("initial_logs", [])
    if any(
        _is_integer(index)
        and 0 <= index < len(logs)
        and is_benign_context_log(logs[index])
        for index in indices
    ):
        return "benign"

    findings = scenario.get("expected_outcomes", {}).get("key_findings", [])
    if isinstance(findings, list) and len(findings) == 1:
        return "benign"
    return "supporting"


def offline_entry(scenario):
    """Build a deterministic mapping entry from log content, never positions."""
    logs = scenario.get("initial_logs", [])
    logs = logs if isinstance(logs, list) else []
    findings = scenario.get("expected_outcomes", {}).get("key_findings", [])
    findings = findings if isinstance(findings, list) else []

    entry_findings = []
    for finding in findings:
        if not isinstance(finding, dict):
            continue
        tagged = dict(finding)
        tagged["_scenario_title"] = scenario.get("title")
        indices = _match_offline_indices(tagged, logs)
        if not indices:
            raise DraftError(
                "offline matcher found no log for finding "
                f"'{finding.get('id')}' in {scenario.get('title')!r}"
            )
        snippets = ", ".join(
            f"#{index} '{_truncate(logs[index].get('message', ''), 55)}'"
            for index in indices[:2]
        )
        rationale = (
            f"Keyword/entity match selects {snippets} because the text directly "
            f"supports '{finding.get('description')}'."
        )
        entry_findings.append(
            {
                "finding_id": finding.get("id"),
                "evidence_log_indices": indices,
                "rationale": _truncate(rationale, validator.MAX_RATIONALE_LENGTH),
                "verdict_support": _offline_support(scenario, finding, indices),
            }
        )

    return {
        "scenario_title": scenario.get("title"),
        "is_true_positive": scenario.get("is_true_positive"),
        "findings": entry_findings,
    }


class LLMClient:
    """Minimal OpenAI-compatible chat-completions client."""

    def __init__(self, api_key, base_url=DEFAULT_BASE_URL, model=DEFAULT_MODEL,
                 timeout=DEFAULT_TIMEOUT_SECONDS):
        if not api_key:
            raise LLMError("LLM_API_KEY is not set")
        self.api_key = api_key
        self.base_url = base_url.rstrip("/")
        self.model = model
        self.timeout = timeout

    def complete(self, system_prompt, user_prompt):
        payload = {
            "model": self.model,
            "temperature": 0,
            "response_format": {"type": "json_object"},
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt},
            ],
        }
        request = urllib.request.Request(
            self.base_url + "/chat/completions",
            data=json.dumps(payload).encode("utf-8"),
            headers={
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json",
            },
            method="POST",
        )
        try:
            with urllib.request.urlopen(request, timeout=self.timeout) as response:
                body = response.read().decode("utf-8", errors="replace")
        except urllib.error.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace") if exc.fp else ""
            raise LLMError(
                f"model endpoint returned HTTP {exc.code}: {_truncate(detail, 400)}"
            ) from exc
        except (urllib.error.URLError, TimeoutError, OSError) as exc:
            raise LLMError(f"model endpoint request failed: {exc}") from exc

        try:
            parsed = json.loads(body)
            content = parsed["choices"][0]["message"]["content"]
        except (json.JSONDecodeError, KeyError, IndexError, TypeError) as exc:
            raise LLMError(
                f"model endpoint returned an unusable response: {_truncate(body, 300)}"
            ) from exc
        if not isinstance(content, str) or not content.strip():
            raise LLMError("model endpoint returned empty message content")
        return content


def draft_scenario(scenario, client, max_retries=2, prompt_version=PROMPT_VERSION):
    """Draft and validate one Scenario, retrying with exact validator errors."""
    payload = scenario_prompt_payload(scenario)
    user_prompt = build_user_prompt(payload)
    attempts = 0
    last_errors = []

    for attempt in range(max_retries + 1):
        attempts = attempt + 1
        content = client.complete(SYSTEM_PROMPT, user_prompt)
        errors = []
        try:
            draft = parse_llm_json(content)
            entry = normalize_entry(draft, scenario)
        except DraftError as exc:
            errors = [f"JSON response error: {exc}"]
            entry = None

        if entry is not None:
            warnings = []
            errors = validate_scenario_entry(entry, scenario, warnings=warnings)
            if not errors:
                return ScenarioResult(
                    scenario=scenario,
                    entry=entry,
                    errors=[],
                    attempts=attempts,
                    drafted_by="llm",
                )

        last_errors = errors
        if attempt < max_retries:
            user_prompt = build_retry_prompt(payload, errors)

    return ScenarioResult(
        scenario=scenario,
        entry=None,
        errors=last_errors,
        attempts=attempts,
        drafted_by="llm",
    )


def draft_offline_scenario(scenario):
    try:
        entry = offline_entry(scenario)
    except DraftError as exc:
        return ScenarioResult(
            scenario=scenario, entry=None, errors=[str(exc)], attempts=0,
            drafted_by="offline",
        )

    warnings = []
    errors = validate_scenario_entry(entry, scenario, warnings=warnings)
    if errors:
        return ScenarioResult(
            scenario=scenario, entry=None, errors=errors, attempts=1,
            drafted_by="offline",
        )
    return ScenarioResult(
        scenario=scenario, entry=entry, errors=[], attempts=1, drafted_by="offline"
    )


def draft_scenarios(scenarios, client=None, max_retries=2, concurrency=3,
                    offline=False):
    """Draft every Scenario, preserving the input order."""
    if offline:
        return [draft_offline_scenario(scenario) for scenario in scenarios]

    if client is None:
        raise ValueError("an LLM client is required unless offline=True")

    if concurrency <= 1 or len(scenarios) <= 1:
        return [
            draft_scenario(scenario, client, max_retries=max_retries)
            for scenario in scenarios
        ]

    results = [None] * len(scenarios)
    with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as pool:
        futures = {
            pool.submit(
                draft_scenario, scenario, client, max_retries=max_retries
            ): index
            for index, scenario in enumerate(scenarios)
        }
        for future in concurrent.futures.as_completed(futures):
            results[futures[future]] = future.result()
    return results


def build_mapping(results, model, drafted_by, prompt_version=PROMPT_VERSION,
                  generated_at=None, notes=None):
    entries = [result.entry for result in results if result.entry is not None]
    failures = [
        {
            "scenario_title": result.scenario.get("title"),
            "errors": list(result.errors),
        }
        for result in results
        if result.failed
    ]
    mapping = {
        "schema_version": SCHEMA_VERSION,
        "generated_at": generated_at or _utc_now(),
        "model": model,
        "prompt_version": prompt_version,
        "drafted_by": drafted_by,
        "approval": {
            "approved": False,
            "approved_by": None,
            "approved_at": None,
            "mapping_sha256": None,
        },
        "scenarios": entries,
    }
    if failures:
        mapping["failed_scenarios"] = failures
    if notes:
        mapping["notes"] = list(notes)
    return mapping


def _union(entry):
    union = set()
    for finding in entry.get("findings", []):
        if not isinstance(finding, dict):
            continue
        for index in finding.get("evidence_log_indices", []) or []:
            if _is_integer(index):
                union.add(index)
    return sorted(union)


def _scenario_warnings(entry, scenario):
    warnings = []
    validate_scenario_entry(entry, scenario, warnings=warnings)
    return warnings


def _fp_missing_benign_context(entry, scenario):
    if scenario.get("is_true_positive") is not False:
        return None
    logs = scenario.get("initial_logs", [])
    logs = logs if isinstance(logs, list) else []
    for finding in entry.get("findings", []):
        if not isinstance(finding, dict) or finding.get("verdict_support") != "benign":
            continue
        for index in finding.get("evidence_log_indices", []) or []:
            if _is_integer(index) and 0 <= index < len(logs) and is_benign_context_log(logs[index]):
                return None
    return (
        f"Scenario '{scenario.get('title')}': False Positive has no "
        "change/HR/scheduled/approved log in its evidence; the benign context is "
        "the observed authorized activity itself."
    )


def _md_cell(text):
    return str(text).replace("\\", "\\\\").replace("|", "\\|").replace("\n", " ")


def _evidence_cell(indices, logs, support):
    rows = []
    for index in indices:
        if not _is_integer(index) or not (0 <= index < len(logs)):
            rows.append(f"#{index} (missing log)")
            continue
        log = logs[index]
        marker = ""
        if support == "benign" and is_benign_context_log(log):
            marker = " **[BENIGN CONTEXT]**"
        rows.append(
            f"#{index} {log.get('timestamp', '')} "
            f"{log.get('source', '')}/{log.get('severity', '')} — "
            f"\"{_truncate(log.get('message', ''), MAX_REPORT_SNIPPET)}\"{marker}"
        )
    return "<br>".join(rows) if rows else "(none)"


def render_report(scenarios, results, model, prompt_version, drafted_by,
                  generated_at=None, notes=None):
    """Render the compact human review brief required by the workflow."""
    notes = notes or {}
    results_by_title = {result.scenario.get("title"): result for result in results}
    failed = [result for result in results if result.failed]
    all_warnings = []
    warnings_by_title = {}

    for scenario in scenarios:
        result = results_by_title.get(scenario.get("title"))
        if result is None or result.entry is None:
            continue
        warnings = _scenario_warnings(result.entry, scenario)
        missing = _fp_missing_benign_context(result.entry, scenario)
        if missing:
            warnings.append(missing)
        warnings_by_title[scenario.get("title")] = warnings
        all_warnings.extend(warnings)

    status = "PENDING APPROVAL"
    if failed:
        status = f"FAILED — {len(failed)} scenario(s) need manual mapping"

    lines = [
        "# Evidence Mapping Review",
        "",
        f"Batch: {len(scenarios)} scenarios · Model: {model} · Prompt: {prompt_version}",
        f"Drafted by: {drafted_by} · Generated: {generated_at or _utc_now()}",
        "Mapping SHA-256: not computed (unapproved draft)",
        f"Status: {status}",
        "",
        "## Coverage summary",
        "",
        "| Scenario | Verdict | Findings | Evidence indices |",
        "|---|---|---|---|",
    ]

    for index, scenario in enumerate(scenarios, 1):
        result = results_by_title.get(scenario.get("title"))
        verdict = "TRUE POSITIVE" if scenario.get("is_true_positive") else "FALSE POSITIVE"
        if result is None:
            lines.append(
                f"| {index}. {_md_cell(scenario.get('title'))} | {verdict} | — | not drafted |"
            )
            continue
        if result.entry is None:
            lines.append(
                f"| {index}. {_md_cell(scenario.get('title'))} | {verdict} | FAILED | — |"
            )
            continue
        entry = result.entry
        finding_count = len(entry.get("findings", []))
        indices = ", ".join(f"#{value}" for value in _union(entry)) or "—"
        lines.append(
            f"| {index}. {_md_cell(scenario.get('title'))} | {verdict} | "
            f"{finding_count} | {indices} |"
        )

    lines.extend(["", "## Warnings", ""])
    if all_warnings:
        for warning in all_warnings:
            lines.append(f"- {_md_cell(warning)}")
    else:
        lines.append("No warnings.")
    lines.append("")

    for index, scenario in enumerate(scenarios, 1):
        title = scenario.get("title")
        result = results_by_title.get(title)
        verdict = "TRUE POSITIVE" if scenario.get("is_true_positive") else "FALSE POSITIVE"
        lines.append(f"## {index}. {title} — {verdict}")
        lines.append("")

        if result is None:
            lines.append("Status: NOT DRAFTED — no mapping entry was provided.")
            lines.append("")
            continue
        if result.entry is None:
            lines.append("Status: FAILED — no valid mapping was produced.")
            lines.append("")
            for error in result.errors:
                lines.append(f"- {_md_cell(error)}")
            for note in result.notes + list(notes.get(title, [])):
                lines.append(f"- Note: {_md_cell(note)}")
            lines.append("")
            continue

        entry = result.entry
        logs = scenario.get("initial_logs", [])
        logs = logs if isinstance(logs, list) else []
        union = _union(entry)
        percentage = (100.0 * len(union) / len(logs)) if logs else 0.0
        lines.extend(
            [
                f"Status: DRAFTED ({result.drafted_by}, {result.attempts} attempt(s))",
                f"TP/FP Decision: {verdict}",
                f"Evidence union: {len(union)} / {len(logs)} logs ({percentage:.0f}%)",
                "",
                "| Scenario | Finding | Evidence logs | Rationale | Support |",
                "|---|---|---|---|---|",
            ]
        )

        for finding in entry.get("findings", []):
            if not isinstance(finding, dict):
                continue
            finding_id = finding.get("finding_id")
            finding_meta = next(
                (
                    candidate
                    for candidate in scenario.get("expected_outcomes", {}).get(
                        "key_findings", []
                    )
                    if isinstance(candidate, dict) and candidate.get("id") == finding_id
                ),
                {},
            )
            description = finding_meta.get("description", "")
            evidence = _evidence_cell(
                finding.get("evidence_log_indices", []) or [],
                logs,
                finding.get("verdict_support"),
            )
            lines.append(
                f"| {index}. {_md_cell(title)} ({'TP' if scenario.get('is_true_positive') else 'FP'}) "
                f"| {_md_cell(finding_id)} — {_md_cell(description)} "
                f"| {evidence} "
                f"| {_md_cell(finding.get('rationale', ''))} "
                f"| {_md_cell(finding.get('verdict_support', ''))} |"
            )

        lines.append("")
        flags = list(warnings_by_title.get(title, []))
        flags.extend(result.notes)
        flags.extend(notes.get(title, []))
        if flags:
            lines.append("Flags:")
            for flag in flags:
                lines.append(f"- {_md_cell(flag)}")
        else:
            lines.append("Flags: none")
        lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def write_artifacts(mapping, report, output_path, report_path):
    output_path = Path(output_path)
    report_path = Path(report_path)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(
        json.dumps(mapping, indent=2, ensure_ascii=False) + "\n", encoding="utf-8"
    )
    report_path.write_text(report, encoding="utf-8")


def _load_existing_entries(path):
    try:
        payload = json.loads(Path(path).read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}
    entries = {}
    for entry in payload.get("scenarios", []) if isinstance(payload, dict) else []:
        if isinstance(entry, dict) and isinstance(entry.get("scenario_title"), str):
            entries[entry["scenario_title"]] = entry
    return entries


def _merge_existing(results, scenarios, existing_entries):
    results_by_title = {
        result.scenario.get("title"): result for result in results
    }
    merged = []
    for scenario in scenarios:
        title = scenario.get("title")
        result = results_by_title.get(title)
        if result is not None:
            merged.append(result)
            continue
        entry = existing_entries.get(title)
        if entry is None:
            continue
        merged.append(
            ScenarioResult(
                scenario=scenario,
                entry=entry,
                errors=[],
                attempts=0,
                drafted_by="existing",
            )
        )
    return merged


def main(argv=None, client=None, scenarios=None):
    parser = argparse.ArgumentParser(
        description=(
            "Draft evidence mappings for the static Scenario set. The LLM path "
            "requires LLM_API_KEY; --offline is the explicit deterministic path."
        )
    )
    parser.add_argument(
        "--output", default=str(SCRIPT_DIR / "evidence_mapping.json"),
        help="Mapping JSON output path (default: scripts/evidence_mapping.json)",
    )
    parser.add_argument(
        "--report", default=str(SCRIPT_DIR / "evidence_mapping_report.md"),
        help="Human review brief output path (default: scripts/evidence_mapping_report.md)",
    )
    parser.add_argument("--concurrency", type=int, default=3)
    parser.add_argument("--max-retries", type=int, default=2)
    parser.add_argument(
        "--scenario",
        help="Redraft one Scenario by exact title, preserving other existing entries",
    )
    parser.add_argument(
        "--offline",
        action="store_true",
        help=(
            "Use the explicit deterministic keyword/entity drafter instead of an "
            "LLM. This path never chunks indices positionally."
        ),
    )
    args = parser.parse_args(argv)

    all_scenarios = scenarios if scenarios is not None else _load_base_scenarios()
    if not all_scenarios:
        print("No base scenarios were loaded.", file=sys.stderr)
        return 2

    selected = all_scenarios
    if args.scenario:
        selected = [
            scenario for scenario in all_scenarios
            if scenario.get("title") == args.scenario
        ]
        if not selected:
            print(f"Unknown scenario title: {args.scenario}", file=sys.stderr)
            return 2

    existing_entries = {}
    if args.scenario:
        existing_entries = _load_existing_entries(args.output)

    if args.offline:
        model = "offline-deterministic-keyword-match"
        drafted_by = "offline"
        results = draft_scenarios(
            selected, client=None, max_retries=args.max_retries,
            concurrency=1, offline=True,
        )
    else:
        config = resolve_llm_config()
        if client is None:
            if not config["api_key"]:
                print(
                    "LLM_API_KEY is not set in the environment or .env. "
                    "Run again with --offline for the explicit deterministic drafter.",
                    file=sys.stderr,
                )
                return 2
            client = LLMClient(
                api_key=config["api_key"],
                base_url=config["base_url"],
                model=config["model"],
            )
        model = config["model"]
        drafted_by = "llm"
        try:
            results = draft_scenarios(
                selected, client=client, max_retries=args.max_retries,
                concurrency=args.concurrency, offline=False,
            )
        except LLMError as exc:
            print(f"LLM drafting failed: {exc}", file=sys.stderr)
            print(
                "No offline fallback was applied. Re-run with --offline to use "
                "the explicit deterministic drafter.",
                file=sys.stderr,
            )
            return 2

    merge_note = None
    if args.scenario:
        results = _merge_existing(results, all_scenarios, existing_entries)
        merge_note = (
            f"--scenario redraft of {args.scenario!r}; "
            f"{len(existing_entries)} existing entr(y/ies) were available to preserve."
        )

    generated_at = _utc_now()
    notes = [merge_note] if merge_note else None
    mapping = build_mapping(
        results,
        model=model,
        drafted_by=drafted_by,
        generated_at=generated_at,
        notes=notes,
    )
    report = render_report(
        all_scenarios if args.scenario else selected,
        results,
        model=model,
        prompt_version=PROMPT_VERSION,
        drafted_by=drafted_by,
        generated_at=generated_at,
        notes={},
    )
    write_artifacts(mapping, report, args.output, args.report)

    failed = [result for result in results if result.failed]
    for result in failed:
        print(
            f"FAILED: {result.scenario.get('title')}: "
            + "; ".join(result.errors),
            file=sys.stderr,
        )
    print(
        f"Wrote {args.output} and {args.report} with "
        f"{len(mapping['scenarios'])} scenario mapping(s), {len(failed)} failed."
    )
    return 1 if failed else 0


def _load_base_scenarios():
    from generate_scenarios import create_scenarios

    return create_scenarios()


if __name__ == "__main__":
    sys.exit(main())
