"""
Cross-OS Setup & Cloud TLS Acceptance Contract Tests (Spec 16 / Tickets #65-#68).
Tests externally observable behavior across shared config contracts, fail-closed
Compose overlays, line endings, PowerShell/Bash syntax, and Cloud safety guards.
"""

from __future__ import annotations

import os
import re
import shlex
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


def parse_env_lines(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, val = line.split("=", 1)
        values[key.strip()] = val.strip()
    return values


def render_caddyfile_from_template(template_text: str, domain: str, acme_email: str) -> str:
    email_block = f"{{\n    email {acme_email.strip()}\n}}\n\n" if acme_email.strip() else ""
    rendered = template_text.replace("__ACME_EMAIL_BLOCK__", email_block).replace("__DOMAIN__", domain.strip())
    return rendered


def test_gitattributes_and_line_endings() -> None:
    gitattributes = ROOT / ".gitattributes"
    assert gitattributes.exists(), ".gitattributes must exist for cross-OS LF/CRLF guarantees"
    content = gitattributes.read_text(encoding="utf-8")
    for required in ("*.sh", "eol=lf", "*.ps1", "*.bat", "eol=crlf", "Caddyfile"):
        assert required in content, f".gitattributes missing rule token: {required}"

    # Shell scripts and Dockerfiles must never contain CRLF bytes on disk
    lf_files = [
        ROOT / "Dockerfile",
        ROOT / "docker" / "seeder" / "Dockerfile",
        ROOT / "frontend" / "Dockerfile",
    ]
    for sh_path in [*ROOT.glob("*.sh"), *(ROOT / "scripts").glob("*.sh")]:
        lf_files.append(sh_path)
    for target in lf_files:
        if target.exists():
            raw = target.read_bytes()
            assert b"\r\n" not in raw, f"LF-enforced file contains CRLF bytes: {target}"


def test_env_example_fail_closed_and_complete() -> None:
    env_example = ROOT / ".env.example"
    assert env_example.exists(), ".env.example must exist"
    env_map = parse_env_lines(env_example)

    for blank_secret in ("POSTGRES_PASSWORD", "JWT_SECRET", "ADMIN_PASSWORD", "PGADMIN_DEFAULT_PASSWORD", "SMTP_PASSWORD"):
        assert blank_secret in env_map, f".env.example missing secret key: {blank_secret}"
        assert env_map[blank_secret] == "", f".env.example must leave {blank_secret} blank (fail-closed)"

    required_defaults = {
        "POSTGRES_USER": "admin",
        "POSTGRES_DB": "cyber_sim",
        "ADMIN_EMAIL": "admin@soc.local",
        "ADMIN_NICKNAME": "admin",
        "GIN_MODE": "release",
        "ENABLE_SWAGGER": "false",
        "BYPASS_POST_TEST_THRESHOLD": "false",
        "BYPASS_THEORY_GATE": "false",
        "APP_ENVIRONMENT": "local",
        "DOMAIN": "",
        "ACME_EMAIL": "",
        "COMPOSE_PATH_SEPARATOR": ":",
        "COMPOSE_FILE": "docker-compose.yml",
        "FRONTEND_URL": "http://localhost",
        "REQUIRE_EMAIL_VERIFICATION": "true",
        "APP_BASE_URL": "http://localhost",
        "SMTP_PORT": "587",
        "SMTP_TLS_MODE": "starttls",
    }
    for key, expected in required_defaults.items():
        assert key in env_map, f".env.example missing key: {key}"
        assert env_map[key] == expected, f".env.example {key}={env_map[key]!r}, expected {expected!r}"


def test_compose_fail_closed_and_cloud_caddy_overlay() -> None:
    cloud_compose = ROOT / "docker-compose.cloud.yml"
    caddy_tpl = ROOT / "Caddyfile.template"
    assert cloud_compose.exists(), "docker-compose.cloud.yml must exist"
    assert caddy_tpl.exists(), "Caddyfile.template must exist"

    # 1. Fail-closed when POSTGRES_PASSWORD and JWT_SECRET are empty
    with tempfile.NamedTemporaryFile("w", delete=False, suffix=".env", encoding="utf-8") as tmp_empty:
        tmp_empty.write("POSTGRES_USER=admin\nPOSTGRES_DB=cyber_sim\nPOSTGRES_PASSWORD=\nJWT_SECRET=\n")
        empty_env_path = Path(tmp_empty.name)
    try:
        res_fail = subprocess.run(
            ["docker", "compose", "--env-file", str(empty_env_path), "-f", str(ROOT / "docker-compose.yml"), "config"],
            capture_output=True,
            text=True,
            cwd=str(ROOT),
        )
        assert res_fail.returncode != 0, "Compose must fail closed when required secrets are blank"
    finally:
        empty_env_path.unlink(missing_ok=True)

    # 2. Local & Cloud Caddy overlay config resolution
    with tempfile.NamedTemporaryFile("w", delete=False, suffix=".env", encoding="utf-8") as tmp_ok:
        tmp_ok.write(
            "\n".join(
                [
                    "POSTGRES_USER=admin",
                    "POSTGRES_DB=cyber_sim",
                    "POSTGRES_PASSWORD=StrongTestDbPass12345",
                    "JWT_SECRET=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
                    "APP_ENVIRONMENT=cloud",
                    "DOMAIN=soc.example.org",
                    "ACME_EMAIL=ops@example.org",
                    "FRONTEND_URL=https://soc.example.org",
                    "COMPOSE_PATH_SEPARATOR=:",
                    "COMPOSE_FILE=docker-compose.yml:docker-compose.cloud.yml",
                ]
            )
            + "\n"
        )
        ok_env_path = Path(tmp_ok.name)
    try:
        res_ok = subprocess.run(
            ["docker", "compose", "--env-file", str(ok_env_path), "config"],
            capture_output=True,
            text=True,
            cwd=str(ROOT),
        )
        assert res_ok.returncode == 0, f"Cloud compose config failed:\n{res_ok.stderr}"
        cfg = res_ok.stdout
        assert "caddy:" in cfg, "Cloud config output missing caddy service"
        assert "cap_drop:" in cfg and "NET_BIND_SERVICE" in cfg, "Cloud caddy missing least-privilege capabilities"
        assert "published: \"443\"" in cfg or "published: 443" in cfg, "Cloud caddy missing port 443"
        # Verify frontend block in rendered YAML has no published host port 80
        m = re.search(r"(?m)^  frontend:\r?\n(.*?)(?=^  [a-zA-Z0-9_-]+:|^networks:)", cfg, re.DOTALL)
        assert m, "Could not locate frontend service block in rendered compose config"
        assert "published:" not in m.group(1), f"frontend must clear host ports via !reset [], got:\n{m.group(1)}"
        # Verify backend receives APP_ENVIRONMENT and BYPASS_THEORY_GATE from .env
        b_match = re.search(r"(?m)^  backend:\r?\n(.*?)(?=^  [a-zA-Z0-9_-]+:|^networks:)", cfg, re.DOTALL)
        assert b_match and "APP_ENVIRONMENT: cloud" in b_match.group(1) and "BYPASS_THEORY_GATE:" in b_match.group(1), (
            "backend service must forward APP_ENVIRONMENT and BYPASS_THEORY_GATE from .env"
        )
        assert "REQUIRE_EMAIL_VERIFICATION:" in b_match.group(1) and "APP_BASE_URL:" in b_match.group(1), (
            "backend service must forward REQUIRE_EMAIL_VERIFICATION and APP_BASE_URL"
        )

        # Verify opt-in dev-email profile enables mailpit cleanly without launching in default runs
        assert "mailpit:" not in cfg, "mailpit must stay opt-in behind dev-email profile"
        res_mailpit = subprocess.run(
            ["docker", "compose", "--env-file", str(ok_env_path), "-f", str(ROOT / "docker-compose.yml"), "--profile", "dev-email", "config"],
            capture_output=True,
            text=True,
            cwd=str(ROOT),
        )
        assert res_mailpit.returncode == 0 and "mailpit:" in res_mailpit.stdout, (
            f"dev-email profile must enable mailpit service cleanly:\n{res_mailpit.stderr}"
        )
    finally:
        ok_env_path.unlink(missing_ok=True)

    # 3. Caddyfile rendering with explicit email vs blank email + security headers
    tpl_text = caddy_tpl.read_text(encoding="utf-8")
    with_email = render_caddyfile_from_template(tpl_text, "soc.example.org", "ops@example.org")
    assert "email ops@example.org" in with_email
    assert "soc.example.org" in with_email
    assert "Strict-Transport-Security" in with_email
    assert "X-Content-Type-Options" in with_email
    assert "reverse_proxy frontend:80" in with_email

    without_email = render_caddyfile_from_template(tpl_text, "soc.example.org", "")
    assert "email" not in without_email, "Blank ACME_EMAIL must omit empty email directive cleanly"
    assert without_email.lstrip().startswith("soc.example.org")
    assert "reverse_proxy frontend:80" in without_email


def test_windows_installer_contracts_and_guards() -> None:
    install_ps1 = ROOT / "install.ps1"
    install_bat = ROOT / "install.bat"
    bootstrap_ps1 = ROOT / "scripts" / "bootstrap.ps1"
    for p in (install_ps1, install_bat, bootstrap_ps1):
        assert p.exists(), f"Missing Windows setup artifact: {p}"

    # 1. Verify Batch wrapper bypasses ExecutionPolicy and forwards CLI arguments
    bat_text = install_bat.read_text(encoding="utf-8")
    assert "ExecutionPolicy" in bat_text and "Bypass" in bat_text
    assert "install.ps1" in bat_text
    assert "%*" in bat_text

    # 2. Verify PowerShell AST parses cleanly with zero syntax errors
    for ps_file in (install_ps1, bootstrap_ps1):
        ast_cmd = (
            "$errs = $null; "
            f"[System.Management.Automation.Language.Parser]::ParseFile('{ps_file}', [ref]$null, [ref]$errs) | Out-Null; "
            "if ($errs -and $errs.Count -gt 0) { $errs | ForEach-Object { Write-Error $_.Message }; exit 1 }"
        )
        res_ast = subprocess.run(["powershell.exe", "-NoProfile", "-Command", ast_cmd], capture_output=True, text=True)
        assert res_ast.returncode == 0, f"PowerShell 5.1 AST parse failed for {ps_file}:\n{res_ast.stderr}"

    # 3. Help output exposes all 11 lifecycle subcommands via install.bat
    res_help = subprocess.run(
        ["cmd.exe", "/c", str(install_bat), "--help"],
        capture_output=True,
        text=True,
        cwd=str(ROOT),
    )
    assert res_help.returncode == 0, f"install.bat --help failed:\n{res_help.stderr}"
    for subcmd in ("install", "check", "status", "seed", "check-seed", "update", "backup", "restore", "logs", "uninstall", "reset-dev-data"):
        assert subcmd in res_help.stdout, f"install.ps1 --help missing subcommand: {subcmd}"

    # 4. Isolated temp-dir guards: Cloud refuses reset-dev-data and --verify-e2e
    with tempfile.TemporaryDirectory() as tmpdir:
        td = Path(tmpdir)
        (td / ".env.example").write_text((ROOT / ".env.example").read_text(encoding="utf-8"), encoding="utf-8")
        (td / ".env").write_text(
            "POSTGRES_USER=admin\nPOSTGRES_PASSWORD=TempSecret123\nPOSTGRES_DB=cyber_sim\nJWT_SECRET=abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789\nAPP_ENVIRONMENT=cloud\n",
            encoding="utf-8",
        )
        (td / "install.ps1").write_text(install_ps1.read_text(encoding="utf-8"), encoding="utf-8")

        res_cloud_reset = subprocess.run(
            ["powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", str(td / "install.ps1"), "reset-dev-data", "--non-interactive"],
            capture_output=True,
            text=True,
            cwd=str(td),
        )
        assert res_cloud_reset.returncode != 0, "reset-dev-data MUST be blocked when APP_ENVIRONMENT=cloud"

        res_cloud_e2e = subprocess.run(
            ["powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", str(td / "install.ps1"), "install", "--mode", "cloud", "--verify-e2e", "--non-interactive"],
            capture_output=True,
            text=True,
            cwd=str(td),
        )
        assert res_cloud_e2e.returncode != 0, "--verify-e2e MUST be blocked in Cloud mode"


def test_linux_installer_contracts_and_guards() -> None:
    install_sh = ROOT / "install.sh"
    bootstrap_sh = ROOT / "scripts" / "bootstrap.sh"
    for p in (install_sh, bootstrap_sh):
        assert p.exists(), f"Missing Linux setup artifact: {p}"
        raw = p.read_bytes()
        assert b"\r\n" not in raw, f"Linux script {p} contains CRLF bytes"

    # Legacy hardcoded weak fallback password & host-side seeding must not exist in install.sh
    sh_text = install_sh.read_text(encoding="utf-8")
    assert "POSTGRES_PASSWORD=password" not in sh_text, "install.sh still contains insecure hardcoded password fallback"
    for legacy_host_py in ("scripts/load_exams_to_db.py", "scripts/load_scenarios_to_db.py", "scripts/parse_exams.py"):
        assert legacy_host_py not in sh_text, f"install.sh must seed cleanly via Docker, not host Python: {legacy_host_py}"

    # Verify every table referenced by reset-dev-data TRUNCATE actually exists in init.sql
    init_tables = set(re.findall(r"(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([A-Za-z0-9_]+)", (ROOT / "init.sql").read_text(encoding="utf-8")))
    for script_path in (install_sh, ROOT / "install.ps1"):
        script_txt = script_path.read_text(encoding="utf-8")
        m_trunc = re.search(r"TRUNCATE\s+TABLE\s+([A-Za-z0-9_,\s]+)\s+CASCADE", script_txt)
        assert m_trunc, f"Could not find TRUNCATE TABLE statement in {script_path}"
        trunc_tables = [t.strip() for t in m_trunc.group(1).split(",") if t.strip()]
        for tbl in trunc_tables:
            assert tbl in init_tables, f"{script_path.name} reset-dev-data references nonexistent table: {tbl}"

    # Run genuine Linux Bash syntax + CLI + Cloud guard tests inside an ephemeral Alpine container
    script_check = (
        "apk add --no-cache bash openssl >/dev/null && "
        "bash -n /repo/install.sh && "
        "bash -n /repo/scripts/bootstrap.sh && "
        "tmpd=$(mktemp -d) && "
        "cp /repo/install.sh /repo/.env.example /repo/Caddyfile.template /repo/docker-compose.yml /repo/docker-compose.cloud.yml \"$tmpd/\" && "
        "chmod +x \"$tmpd/install.sh\" && "
        "help_out=$(\"$tmpd/install.sh\" --help) && "
        "for sub in install check status seed check-seed update backup restore logs uninstall reset-dev-data; do "
        "  echo \"$help_out\" | grep -q \"$sub\" || { echo \"Missing subcommand: $sub\"; exit 1; }; "
        "done && "
        "printf 'POSTGRES_USER=admin\\nPOSTGRES_PASSWORD=TempSecret123\\nPOSTGRES_DB=cyber_sim\\nJWT_SECRET=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\\nAPP_ENVIRONMENT=cloud\\n' > \"$tmpd/.env\" && "
        "if \"$tmpd/install.sh\" reset-dev-data --non-interactive >/dev/null 2>&1; then echo 'Cloud reset-dev-data should fail'; exit 1; fi && "
        "if \"$tmpd/install.sh\" install --mode cloud --verify-e2e --non-interactive >/dev/null 2>&1; then echo 'Cloud verify-e2e should fail'; exit 1; fi && "
        "echo LINUX_CONTRACT_OK"
    )
    res_linux = subprocess.run(
        ["docker", "run", "--rm", "-v", f"{ROOT}:/repo:ro", "alpine:latest", "sh", "-c", script_check],
        capture_output=True,
        text=True,
    )
    assert res_linux.returncode == 0, f"Linux Bash container verification failed:\nSTDOUT:\n{res_linux.stdout}\nSTDERR:\n{res_linux.stderr}"
    assert "LINUX_CONTRACT_OK" in res_linux.stdout


def test_docs_consolidated_and_legacy_superseded() -> None:
    readme = (ROOT / "README.md").read_text(encoding="utf-8")
    for token in ("install.sh", "install.bat", "install.ps1", "bootstrap.sh", "bootstrap.ps1", "--mode cloud", "--domain", "workflows/cross_platform_setup.md"):
        assert token in readme, f"README.md missing unified setup documentation reference: {token}"

    for legacy_name in ("INSTALL_SH_SPECIFICATION.md", "DEPLOYMENT_BOOTSTRAP_PLAN.md"):
        legacy_path = ROOT / "docs" / legacy_name
        assert legacy_path.exists(), f"Missing expected doc {legacy_path}"
        head = "\n".join(legacy_path.read_text(encoding="utf-8").splitlines()[:25]).lower()
        assert "superseded" in head, f"docs/{legacy_name} must clearly carry a SUPERSEDED notice at top"
        assert "cross_platform_setup.md" in head, f"docs/{legacy_name} must link to workflows/cross_platform_setup.md"


def test_secret_precedence_and_non_leakage_across_both_installers() -> None:
    # Verify both Windows (PS 5.1) and Linux (Bash) handle secret-file precedence,
    # preserve existing secrets cleanly on re-run, generate Caddyfile cleanly without email,
    # lock .env ACL/mode, and never echo DB/JWT/LLM secrets in non-interactive stdout.
    ps_secret_val = "SecretFromFileDb987654"
    jwt_secret_val = "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
    llm_secret_val = "sk-test-secret-do-not-leak-999"

    with tempfile.TemporaryDirectory() as tmpdir:
        td = Path(tmpdir)
        for fname in (".env.example", "Caddyfile.template", "docker-compose.yml", "docker-compose.cloud.yml", "install.ps1", "install.sh"):
            shutil_src = ROOT / fname
            (td / fname).write_bytes(shutil_src.read_bytes())

        pg_file = td / "pg.secret"
        pg_file.write_text(ps_secret_val + "\n", encoding="utf-8")
        jwt_file = td / "jwt.secret"
        jwt_file.write_text(jwt_secret_val + "\n", encoding="utf-8")
        llm_file = td / "llm.secret"
        llm_file.write_text(llm_secret_val + "\n", encoding="utf-8")

        # Test Windows Initialize-EnvConfig + idempotent re-run preservation + Cloud -> Local switch + native stderr handling
        ps_harness = (
            "$ErrorActionPreference = 'Stop'; "
            f"$src = [System.IO.File]::ReadAllText('{(td / 'install.ps1')}'); "
            "$marker = '# Dispatcher'; "
            "$lib = $src.Substring(0, $src.IndexOf($marker)); "
            "Invoke-Expression $lib; "
            "$stderrRc = Invoke-NativeQuiet 'cmd.exe' @('/c', 'echo errline 1>&2 & exit 7'); "
            "if ($stderrRc -ne 7) { throw ('Expected exit code 7 from stderr native cmd, got ' + $stderrRc) }; "
            "$Script:RootDir = '" + str(td) + "'; "
            "$Script:Mode = 'cloud'; "
            "$Script:Domain = 'blue.example.org'; "
            "$Script:NonInteractive = $true; "
            "$Script:SkipDnsCheck = $true; "
            "$env:POSTGRES_PASSWORD_FILE = '" + str(pg_file) + "'; "
            "$env:JWT_SECRET_FILE = '" + str(jwt_file) + "'; "
            "$env:LLM_API_KEY_FILE = '" + str(llm_file) + "'; "
            "$cfg1 = Initialize-EnvConfig; "
            "Write-Host ('FRONTEND=' + $cfg1.FrontendUrl); "
            # Verify .env ACL grants owner FullControl (F) (not R,W-only which breaks renames/OneDrive) and blocks Users/Everyone
            "$ic = (& icacls '.env' | Out-String); "
            "if ($ic -notmatch '\\(F\\)') { throw ('Locked .env must grant FullControl (F), got: ' + $ic) }; "
            "if ($ic -match 'BUILTIN\\\\Users|Everyone|Authenticated Users') { throw ('Locked .env leaked broad ACL: ' + $ic) }; "
            # Also confirm Test-EnvFileAcl runs cleanly without Microsoft.PowerShell.Security Get-Acl dependency
            "Test-EnvFileAcl '.env'; "
            # Simulate an older (R,W)-locked .env left in OneDrive/Windows and verify re-run rescues ACL cleanly
            "$id = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value; "
            "& icacls '.env' /inheritance:r /grant:r ('*' + $id + ':(R,W)') | Out-Null; "
            # Customize ENABLE_SWAGGER=true, drop *_FILE env vars, and verify re-run preserves secrets + custom flags
            "$c = ([System.IO.File]::ReadAllText('.env')).Replace('ENABLE_SWAGGER=false', 'ENABLE_SWAGGER=true'); "
            "[System.IO.File]::WriteAllText('.env', $c); "
            "Remove-Item Env:POSTGRES_PASSWORD_FILE, Env:JWT_SECRET_FILE, Env:LLM_API_KEY_FILE; "
            "$cfg2 = Initialize-EnvConfig; "
            "if ($cfg2.AdminPassword -ne $cfg1.AdminPassword) { throw 'AdminPassword changed on re-run!' }; "
            "$Script:Mode = 'local'; $Script:Domain = ''; "
            "$cfg3 = Initialize-EnvConfig; "
            "Write-Host ('SWITCH_LOCAL_URL=' + $cfg3.FrontendUrl)"
        )
        res_ps = subprocess.run(
            ["powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps_harness],
            capture_output=True,
            text=True,
            cwd=str(td),
        )
        assert res_ps.returncode == 0, f"Windows Initialize-EnvConfig harness failed:\nSTDOUT:\n{res_ps.stdout}\nSTDERR:\n{res_ps.stderr}"
        assert "FRONTEND=https://blue.example.org" in res_ps.stdout
        assert "SWITCH_LOCAL_URL=http://localhost" in res_ps.stdout
        combined_ps_out = res_ps.stdout + res_ps.stderr
        for secret in (ps_secret_val, jwt_secret_val, llm_secret_val):
            assert secret not in combined_ps_out, "Windows installer leaked secret to console output!"
        env_win = parse_env_lines(td / ".env")
        assert env_win["POSTGRES_PASSWORD"] == ps_secret_val
        assert env_win["JWT_SECRET"] == jwt_secret_val
        assert env_win["LLM_API_KEY"] == llm_secret_val
        assert env_win["ENABLE_SWAGGER"] == "true", "Custom ENABLE_SWAGGER must be preserved across re-runs"
        assert env_win["COMPOSE_FILE"] == "docker-compose.yml", "Cloud->Local switch must reset COMPOSE_FILE"
        assert env_win["DOMAIN"] == "", "Cloud->Local switch must clear DOMAIN"
        assert (td / "Caddyfile").exists()
        assert "email" not in (td / "Caddyfile").read_text(encoding="utf-8")

        # Test Linux setup_env_config + re-run preservation + Cloud -> Local switch inside Alpine bash
        (td / ".env").unlink()
        (td / "Caddyfile").unlink()
        sh_inner = (
            "set -e; "
            "tmpf=$(mktemp -d); cp -r /work/. \"$tmpf\"; "
            "cd \"$tmpf\"; "
            "sed 's/^main \"\\$@\"$//' install.sh > install_lib.sh; "
            "source ./install_lib.sh; "
            "MODE=cloud; DOMAIN=linux.example.org; NON_INTERACTIVE=true; SKIP_DNS_CHECK=true; "
            "export POSTGRES_PASSWORD_FILE=\"$tmpf/pg.secret\" JWT_SECRET_FILE=\"$tmpf/jwt.secret\" LLM_API_KEY_FILE=\"$tmpf/llm.secret\"; "
            "setup_env_config; "
            "perms=$(stat -c '%a' .env); echo \"ENV_PERMS=$perms\"; "
            "grep -q 'COMPOSE_FILE=docker-compose.yml:docker-compose.cloud.yml' .env; "
            "first_adm=\"$CFG_ADMIN_PASS\"; "
            "sed -i 's/^ENABLE_SWAGGER=false$/ENABLE_SWAGGER=true/' .env; "
            "unset POSTGRES_PASSWORD_FILE JWT_SECRET_FILE LLM_API_KEY_FILE; "
            "setup_env_config; "
            "[ \"$CFG_ADMIN_PASS\" = \"$first_adm\" ] || { echo 'AdminPassword mutated on Linux re-run'; exit 1; }; "
            "grep -q '^ENABLE_SWAGGER=true$' .env || { echo 'ENABLE_SWAGGER clobbered on Linux re-run'; exit 1; }; "
            "MODE=local; DOMAIN=''; setup_env_config; "
            "grep -q '^FRONTEND_URL=http://localhost$' .env || { echo 'Cloud->Local did not reset FRONTEND_URL'; exit 1; }; "
            "grep -q '^COMPOSE_FILE=docker-compose.yml$' .env || { echo 'Cloud->Local did not reset COMPOSE_FILE'; exit 1; }; "
            "echo LINUX_ENV_OK"
        )
        res_sh = subprocess.run(
            ["docker", "run", "--rm", "-v", f"{td}:/work:ro", "alpine:latest", "sh", "-c", f"apk add --no-cache bash openssl >/dev/null && bash -c {shlex.quote(sh_inner)}"],
            capture_output=True,
            text=True,
        )
        assert res_sh.returncode == 0, f"Linux setup_env_config harness failed:\nSTDOUT:\n{res_sh.stdout}\nSTDERR:\n{res_sh.stderr}"
        assert "ENV_PERMS=600" in res_sh.stdout
        assert "LINUX_ENV_OK" in res_sh.stdout
        for secret in (ps_secret_val, jwt_secret_val, llm_secret_val):
            assert secret not in (res_sh.stdout + res_sh.stderr), "Linux installer leaked secret to console output!"


def main() -> int:
    tests = [
        test_gitattributes_and_line_endings,
        test_env_example_fail_closed_and_complete,
        test_compose_fail_closed_and_cloud_caddy_overlay,
        test_windows_installer_contracts_and_guards,
        test_linux_installer_contracts_and_guards,
        test_docs_consolidated_and_legacy_superseded,
        test_secret_precedence_and_non_leakage_across_both_installers,
    ]
    for fn in tests:
        print(f"RUN {fn.__name__} ...")
        fn()
        print(f"OK  {fn.__name__}")
    print(f"ALL {len(tests)} CROSS-PLATFORM SETUP CONTRACT CHECKS PASSED.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
