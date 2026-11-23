import test from "node:test";
import assert from "node:assert/strict";
import { buildInvestigationShortcuts } from "../src/lib/siemShortcuts.ts";

test("buildInvestigationShortcuts produces identical deterministic shortcuts from canonical EN fields", () => {
  const canonicalScenario = {
    title: "MFA Push Fatigue Bypass on Okta SSO",
    description: "Alert triggered for ajones on host okta-gw-01 from external IP 198.51.100.42.",
    initial_logs: [
      {
        timestamp: "2026-11-20T09:00:00Z",
        source: "okta_sso",
        severity: "high",
        host: "okta-gw-01",
        src_ip: "198.51.100.42",
        user: "ajones",
        message: "Multiple MFA push notifications denied followed by approved login",
      },
      {
        timestamp: "2026-11-20T09:01:00Z",
        source: "firewall",
        severity: "medium",
        host: "fw-edge-01",
        src_ip: "10.0.0.5",
        message: "Routine outbound connection",
      },
    ],
  };

  const runEnContext = buildInvestigationShortcuts({
    title: canonicalScenario.title,
    description: canonicalScenario.description,
    initial_logs: canonicalScenario.initial_logs,
  });

  const runThUiContext = buildInvestigationShortcuts({
    title: canonicalScenario.title,
    description: canonicalScenario.description,
    initial_logs: canonicalScenario.initial_logs,
  });

  assert.deepEqual(runThUiContext, runEnContext);
  assert.ok(runEnContext.length > 0);
  assert.equal(runEnContext[0].labelPrefix, "Host:");
  assert.equal(runEnContext[0].label, "okta-gw-01");
  assert.ok(
    runEnContext.some((s) => s.query === "198.51.100.42" && s.labelPrefix === "IP:"),
    "expected external attacker IP shortcut",
  );
  assert.ok(
    runEnContext.every((s) => !/[\u0E00-\u0E7F]/.test(s.query)),
    "KQL queries must contain zero Thai characters",
  );
});

test("buildInvestigationShortcuts preserves primary Host/IP/User extraction on 3-paragraph 4-beat EN narratives", () => {
  const storyScenario = {
    title: "Multi-Factor Authentication Bypass (Push Fatigue)",
    description:
      "You are an on-duty SOC Analyst monitoring identity and single sign-on security for a corporate enterprise workforce. While clearing the morning SIEM alert queue as employees begin logging in for the workday, an urgent multi-factor authentication ticket catches your attention.\n\n" +
      "Security logs show repeated sign-in approval notifications being sent to employee ajones on DC-01 that were rejected multiple times within a single minute, followed shortly by a successful login coming from unfamiliar external IP address 198.51.100.42.\n\n" +
      "If someone is bombarding ajones with unwanted phone prompts until one gets tapped by mistake, an attacker may already be inside company apps; however, if ajones was retrying a flaky mobile connection, locking the account could block urgent daily work—verify the timeline first.",
    initial_logs: [
      {
        timestamp: "2026-11-23T08:00:00Z",
        source: "ad_controller",
        severity: "info",
        host: "FIN-WS-01",
        src_ip: "10.0.1.25",
        user: "kpark",
        message: "Routine workstation Kerberos renewal",
      },
      {
        timestamp: "2026-11-23T08:01:00Z",
        source: "okta",
        severity: "high",
        host: "DC-01",
        src_ip: "198.51.100.42",
        user: "ajones",
        message: "Repeated MFA push rejected then approved from 198.51.100.42",
      },
    ],
  };

  const shortcuts = buildInvestigationShortcuts(storyScenario);
  assert.ok(shortcuts.some((s) => s.labelPrefix === "Host:" && s.label === "DC-01"));
  assert.ok(
    shortcuts.some(
      (s) => (s.labelPrefix === "IP:" && s.label === "198.51.100.42") || (s.labelPrefix === "User:" && s.label === "ajones"),
    ),
  );
  assert.ok(shortcuts.some((s) => s.labelPrefix === "Source:" && s.query === 'source == "okta"'));
});
