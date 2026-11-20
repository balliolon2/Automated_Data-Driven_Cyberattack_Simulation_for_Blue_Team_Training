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
