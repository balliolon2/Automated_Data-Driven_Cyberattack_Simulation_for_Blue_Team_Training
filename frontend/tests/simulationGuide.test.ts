import test from "node:test";
import assert from "node:assert/strict";
import {
  BEGINNER_GLOSSARY_TERM_IDS,
  getBeginnerGlossaryTerms,
  getSimulationGuideDismissStorageKey,
  shouldAutoExpandBeginnerGuide,
} from "../src/lib/simulationGuide.ts";

test("shouldAutoExpandBeginnerGuide expands only on Scenario 1 (completedScenarios === 0) unless dismissed", () => {
  assert.equal(shouldAutoExpandBeginnerGuide(0, false), true);
  assert.equal(shouldAutoExpandBeginnerGuide(0, true), false);

  // Scenario 2+ always starts collapsed
  assert.equal(shouldAutoExpandBeginnerGuide(1, false), false);
  assert.equal(shouldAutoExpandBeginnerGuide(3, false), false);

  // Fail-open when status has not loaded or failed
  assert.equal(shouldAutoExpandBeginnerGuide(null, false), false);
  assert.equal(shouldAutoExpandBeginnerGuide(undefined, false), false);
});

test("getSimulationGuideDismissStorageKey namespaces per learner userId cleanly", () => {
  assert.equal(getSimulationGuideDismissStorageKey("user-123"), "cybersim_guide_collapsed:user-123");
  assert.equal(getSimulationGuideDismissStorageKey("  user-abc "), "cybersim_guide_collapsed:user-abc");
  assert.equal(getSimulationGuideDismissStorageKey(""), "cybersim_guide_collapsed:anon");
  assert.equal(getSimulationGuideDismissStorageKey(null), "cybersim_guide_collapsed:anon");
});

test("getBeginnerGlossaryTerms exposes exact 10 beginner SOC terms cleanly in EN and TH", () => {
  assert.equal(BEGINNER_GLOSSARY_TERM_IDS.length, 10);

  const expectedIds = [
    "siem",
    "kql",
    "ioc",
    "true_positive",
    "false_positive",
    "containment",
    "eradication",
    "recovery",
    "brute_force",
    "phishing",
  ];
  assert.deepEqual([...BEGINNER_GLOSSARY_TERM_IDS], expectedIds);

  const enTerms = getBeginnerGlossaryTerms("en");
  const thTerms = getBeginnerGlossaryTerms("th");
  assert.equal(enTerms.length, 10);
  assert.equal(thTerms.length, 10);

  for (const list of [enTerms, thTerms]) {
    for (const item of list) {
      assert.ok(item.term.trim().length > 0);
      assert.ok(item.definition.trim().length > 10);
      assert.ok(item.example.trim().length > 0);
    }
  }
});
