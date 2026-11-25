#!/usr/bin/env python3
"""
Reproducible NFR Benchmark & CI Gate Aggregator for SOC Trainer (Spec 15 / Issue #15).

Executes multi-level concurrent read-only HTTP load benchmarks, optionally runs and parses
local CI/test/coverage & DevSecOps gates, and outputs a schema-valid JSON report suitable
for direct push (`POST /api/admin/nfr/runs`) or UI import (`Import Benchmark / CI JSON`).
"""

import argparse
import concurrent.futures
import json
import math
import os
import re
import subprocess
import sys
import time
import threading
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional, Tuple

import requests

READ_ONLY_ENDPOINTS = ["/api/health", "/api/exams/sample", "/api/learn/domains"]
_thread_local = threading.local()


def quantile_r7(sorted_vals: List[float], p: float) -> float:
    n = len(sorted_vals)
    if n == 0:
        return 0.0
    if n == 1:
        return sorted_vals[0]
    h = (n - 1) * p
    lo = math.floor(h)
    hi = math.ceil(h)
    frac = h - lo
    return sorted_vals[lo] + frac * (sorted_vals[hi] - sorted_vals[lo])


def load_env_defaults() -> Tuple[str, str]:
    email = os.environ.get("ADMIN_EMAIL", "admin@soc.local")
    password = os.environ.get("ADMIN_PASSWORD", "password")
    if os.path.exists(".env"):
        with open(".env", "r", encoding="utf-8") as f:
            for line in f:
                s = line.strip()
                if s.startswith("ADMIN_EMAIL=") and not os.environ.get("ADMIN_EMAIL"):
                    v = s.split("=", 1)[1].strip().strip('"').strip("'")
                    if v:
                        email = v
                elif s.startswith("ADMIN_PASSWORD=") and not os.environ.get("ADMIN_PASSWORD"):
                    v = s.split("=", 1)[1].strip().strip('"').strip("'")
                    if v:
                        password = v
    return email, password


def login_admin(api_url: str, email: str, password: str) -> str:
    resp = requests.post(
        f"{api_url.rstrip('/')}/api/login",
        json={"email": email, "password": password},
        timeout=15,
    )
    if resp.status_code != 200:
        raise RuntimeError(f"Admin login failed ({resp.status_code}): {resp.text}")
    tok = resp.json().get("token")
    if not tok:
        raise RuntimeError("Login response missing JWT token")
    return tok


def run_concurrency_wave(api_url: str, bearer_token: str, workers: int, total_requests: int) -> Dict[str, Any]:
    base = api_url.rstrip("/")
    headers = {"Authorization": f"Bearer {bearer_token}"}

    def get_session() -> requests.Session:
        sess = getattr(_thread_local, "session", None)
        if sess is None:
            sess = requests.Session()
            _thread_local.session = sess
        return sess

    def hit_one(idx: int) -> Tuple[int, float]:
        ep = READ_ONLY_ENDPOINTS[idx % len(READ_ONLY_ENDPOINTS)]
        sess = get_session()
        t0 = time.perf_counter()
        try:
            r = sess.get(f"{base}{ep}", headers=headers, timeout=15)
            elapsed_ms = (time.perf_counter() - t0) * 1000.0
            return r.status_code, elapsed_ms
        except requests.RequestException:
            elapsed_ms = (time.perf_counter() - t0) * 1000.0
            return 599, elapsed_ms

    t_start = time.perf_counter()
    durations: List[float] = []
    success_count = 0
    error_count = 0

    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        for code, elapsed_ms in pool.map(hit_one, range(total_requests)):
            durations.append(elapsed_ms)
            if 200 <= code < 400:
                success_count += 1
            else:
                error_count += 1

    total_ms = max(1.0, (time.perf_counter() - t_start) * 1000.0)
    sorted_d = sorted(durations)
    mean_ms = sum(sorted_d) / len(sorted_d) if sorted_d else 0.0
    err_pct = (error_count / total_requests) * 100.0 if total_requests > 0 else 0.0
    rps = total_requests / (total_ms / 1000.0)

    return {
        "concurrency": workers,
        "total_requests": total_requests,
        "success_count": success_count,
        "error_count": error_count,
        "error_rate_pct": round(err_pct, 2),
        "throughput_rps": round(rps, 2),
        "mean_ms": round(mean_ms, 2),
        "p50_ms": round(quantile_r7(sorted_d, 0.50), 2),
        "p95_ms": round(quantile_r7(sorted_d, 0.95), 2),
        "p99_ms": round(quantile_r7(sorted_d, 0.99), 2),
        "max_ms": round(sorted_d[-1] if sorted_d else 0.0, 2),
        "duration_ms": round(total_ms, 2),
    }


def run_cmd(cmd: List[str], cwd: Optional[str] = None) -> Tuple[bool, str]:
    try:
        proc = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, timeout=300)
        return proc.returncode == 0, f"{proc.stdout}\n{proc.stderr}"
    except Exception as exc:
        return False, str(exc)


def collect_ci_gates(include_security_scans: bool) -> Dict[str, Any]:
    report: Dict[str, Any] = {
        "collected_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
    }

    ok_git, git_out = run_cmd(["git", "rev-parse", "--short", "HEAD"])
    if ok_git:
        report["git_commit"] = git_out.strip().splitlines()[0]

    # 1. Go unit tests + coverage
    cover_file = os.path.join(".nfr_go_cover.tmp")
    go_ok, go_out = run_cmd(["go", "test", f"-coverprofile={cover_file}", "./..."])
    report["go_test_passed"] = go_ok
    cov_pct: Optional[float] = None
    if os.path.exists(cover_file):
        cov_ok, cov_out = run_cmd(["go", "tool", "cover", f"-func={cover_file}"])
        if cov_ok:
            m = re.search(r"total:\s+\(statements\)\s+([\d.]+)%", cov_out)
            if m:
                cov_pct = round(float(m.group(1)), 2)
        try:
            os.remove(cover_file)
        except OSError:
            pass
    if cov_pct is None:
        matches = [float(x) for x in re.findall(r"coverage:\s+([\d.]+)%\s+of statements", go_out)]
        if matches:
            cov_pct = round(sum(matches) / len(matches), 2)
    if cov_pct is not None:
        report["go_coverage_pct"] = cov_pct

    # 2. Frontend tests & TypeScript typecheck
    npm_cmd = "npm.cmd" if os.name == "nt" else "npm"
    npx_cmd = "npx.cmd" if os.name == "nt" else "npx"
    fe_test_ok, _ = run_cmd([npm_cmd, "test"], cwd="frontend")
    report["frontend_test_passed"] = fe_test_ok

    tsc_ok, _ = run_cmd([npx_cmd, "tsc", "-b"], cwd="frontend")
    report["frontend_tsc_passed"] = tsc_ok

    # 3. Evidence & bilingual theory gates
    ev_ok, _ = run_cmd(
        [sys.executable, "scripts/validate_evidence_mapping.py", "--mapping", "scripts/evidence_mapping.json", "--require-complete"]
    )
    report["evidence_gate_passed"] = ev_ok

    th_ok, _ = run_cmd(
        [sys.executable, "scripts/load_theory_to_db.py", "--validate-only", "--require-all-th"]
    )
    report["theory_gate_passed"] = th_ok

    if include_security_scans:
        gv_ok, gv_out = run_cmd(["govulncheck", "-json", "./..."])
        if gv_ok or "{" in gv_out:
            # Count distinct OSV vulnerability findings emitted
            vulns = len(re.findall(r'"osv"\s*:', gv_out))
            report["govulncheck_high_critical"] = vulns

        audit_ok, audit_out = run_cmd([npm_cmd, "audit", "--json"], cwd="frontend")
        if audit_out.strip():
            try:
                parsed = json.loads(audit_out)
                v = ((parsed.get("metadata") or {}).get("vulnerabilities") or {})
                report["npm_audit_high_critical"] = int(v.get("high", 0)) + int(v.get("critical", 0))
            except Exception:
                if audit_ok:
                    report["npm_audit_high_critical"] = 0

    return report


def main() -> int:
    parser = argparse.ArgumentParser(description="SOC Trainer NFR Concurrency & CI Benchmark Runner")
    parser.add_argument("--api-url", default="http://127.0.0.1:8080", help="Backend base URL")
    parser.add_argument("--email", default="", help="Admin email (defaults to ADMIN_EMAIL / .env)")
    parser.add_argument("--password", default="", help="Admin password (defaults to ADMIN_PASSWORD / .env)")
    parser.add_argument("--label", default="CLI NFR Benchmark", help="Run label for nfr_evaluation_runs")
    parser.add_argument("--environment", default="local-docker", help="Target environment name")
    parser.add_argument("--levels", default="10,25,50", help="Comma-separated concurrent worker counts")
    parser.add_argument("--requests-per-level", type=int, default=100, help="HTTP requests per concurrency wave")
    parser.add_argument("--include-ci", action="store_true", help="Run local Go/Frontend/Evidence/Theory CI gates")
    parser.add_argument("--include-security-scans", action="store_true", help="Opportunistically run govulncheck and npm audit")
    parser.add_argument("--push", action="store_true", help="Push result to POST /api/admin/nfr/runs")
    parser.add_argument("--out", default="", help="Write JSON payload to file for archiving or UI import")
    args = parser.parse_args()

    levels = [int(x.strip()) for x in args.levels.split(",") if x.strip()]
    if not levels or any(lv <= 0 for lv in levels):
        print("Error: --levels must be positive integers", file=sys.stderr)
        return 2

    def_email, def_pass = load_env_defaults()
    email = args.email or def_email
    password = args.password or def_pass

    token = login_admin(args.api_url, email, password)
    wave_results: List[Dict[str, Any]] = []
    for lv in levels:
        print(f"Running concurrency wave: {lv} workers x {args.requests_per_level} requests...")
        res = run_concurrency_wave(args.api_url, token, lv, args.requests_per_level)
        wave_results.append(res)
        print(
            f"  -> workers={res['concurrency']} rps={res['throughput_rps']:.1f} "
            f"p50={res['p50_ms']:.1f}ms p95={res['p95_ms']:.1f}ms p99={res['p99_ms']:.1f}ms err={res['error_rate_pct']:.2f}%"
        )

    payload: Dict[str, Any] = {
        "run_label": args.label,
        "environment": args.environment,
        "source": "cli_benchmark",
        "scalability": {
            "mode": "cli_benchmark",
            "executed_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
            "target_endpoints": READ_ONLY_ENDPOINTS,
            "levels": wave_results,
        },
    }

    if args.include_ci:
        print("Collecting local CI test, coverage, and curriculum validation gates...")
        payload["maintainability_report"] = collect_ci_gates(args.include_security_scans)
        print("  -> CI gates:", json.dumps(payload["maintainability_report"], indent=2))

    if args.out:
        out_dir = os.path.dirname(os.path.abspath(args.out))
        if out_dir:
            os.makedirs(out_dir, exist_ok=True)
        with open(args.out, "w", encoding="utf-8") as f:
            json.dump(payload, f, indent=2)
        print(f"Saved NFR benchmark report JSON to {args.out}")

    if args.push:
        resp = requests.post(
            f"{args.api_url.rstrip('/')}/api/admin/nfr/runs",
            headers={"Authorization": f"Bearer {token}"},
            json=payload,
            timeout=30,
        )
        if resp.status_code != 201:
            print(f"Error pushing NFR run ({resp.status_code}): {resp.text}", file=sys.stderr)
            return 1
        data = resp.json()
        run_info = data.get("run") or {}
        print(
            f"Pushed NFR run {run_info.get('run_id')} -> overall={run_info.get('overall_status')} "
            f"(pass={run_info.get('passed_count')} warn={run_info.get('warn_count')} "
            f"fail={run_info.get('fail_count')} pending={run_info.get('not_measured_count')})"
        )

    return 0


if __name__ == "__main__":
    sys.exit(main())
