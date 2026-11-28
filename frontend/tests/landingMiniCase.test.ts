import test from "node:test";
import assert from "node:assert/strict";
import {
  evaluateLandingMiniCase,
  getLandingMiniCases,
} from "../src/lib/landingMiniCase.ts";

test("getLandingMiniCases returns 2 rotatable beginner cases (1 TP and 1 FP) with 3 English SIEM logs each", () => {
  const thCases = getLandingMiniCases("th");
  const enCases = getLandingMiniCases("en");

  assert.equal(thCases.length, 2);
  assert.equal(enCases.length, 2);

  for (const item of [...thCases, ...enCases]) {
    assert.ok(item.id.length > 0);
    assert.ok(item.title.trim().length > 0);
    assert.ok(item.briefing.trim().length > 0);
    assert.equal(item.logs.length, 3, `Case ${item.id} must have 3 clean SIEM log lines`);
    for (const log of item.logs) {
      assert.ok(log.timestamp.length > 0);
      assert.ok(log.source.length > 0);
      assert.ok(log.severity.length > 0);
      assert.ok(log.message.length > 0);
      // Keep SIEM logs authentic English without pre-revealing TP/FP verdict text
      assert.equal(/true positive|false positive/i.test(log.message), false);
    }
    assert.equal("expectedVerdict" in item, false, "Public mini-case view must not leak verdict before selection");
  }
});

test("evaluateLandingMiniCase grades TP vs FP decisions accurately with localized log-referenced feedback", () => {
  const cases = getLandingMiniCases("th");
  const firstResult = evaluateLandingMiniCase(cases[0].id, "tp", "th");
  const secondResult = evaluateLandingMiniCase(cases[1].id, "fp", "th");

  assert.ok(firstResult !== null);
  assert.equal(firstResult.expectedVerdict, "tp");
  assert.equal(firstResult.isCorrect, true);
  assert.ok(firstResult.rationale.trim().length > 20);

  assert.ok(secondResult !== null);
  assert.equal(secondResult.expectedVerdict, "fp");
  assert.equal(secondResult.isCorrect, true);
  assert.ok(secondResult.rationale.trim().length > 20);

  const wrongPick = evaluateLandingMiniCase(cases[0].id, "fp", "en");
  assert.ok(wrongPick !== null);
  assert.equal(wrongPick.isCorrect, false);
  assert.equal(wrongPick.expectedVerdict, "tp");
  assert.ok(wrongPick.rationale.trim().length > 20);

  assert.equal(evaluateLandingMiniCase("unknown-case", "tp", "en"), null);
});
