import test from "node:test";
import assert from "node:assert/strict";
import {
  CSUQ_ITEMS,
  NASA_TLX_DIMENSIONS,
  MAX_QUALITATIVE_CHARS,
  canAdvanceCSUQStep,
  canAdvanceTLXStep,
  computeCSUQScoresPreview,
  computeNASATLXPreview,
  createEmptyQuestionnaireDraft,
  getQuestionnaireDraftStorageKey,
  parseQuestionnaireDraft,
  serializeQuestionnaireDraft,
  countQualitativeChars,
  isQualitativeStepValid,
  type CSUQAnswerValue,
} from "../src/lib/questionnaire.ts";

test("CSUQ and NASA-TLX catalogs have complete bilingual metadata and canonical counts", () => {
  assert.equal(CSUQ_ITEMS.length, 16);
  const ids = CSUQ_ITEMS.map((item) => item.key);
  assert.deepEqual(ids, [
    "q1", "q2", "q3", "q4", "q5", "q6",
    "q7", "q8", "q9", "q10", "q11", "q12",
    "q13", "q14", "q15", "q16",
  ]);
  for (const item of CSUQ_ITEMS) {
    assert.ok(item.en.trim().length > 10, `CSUQ ${item.key} EN missing`);
    assert.ok(item.th.trim().length > 10 && /[\u0E00-\u0E7F]/.test(item.th), `CSUQ ${item.key} TH missing`);
  }

  assert.equal(NASA_TLX_DIMENSIONS.length, 6);
  for (const dim of NASA_TLX_DIMENSIONS) {
    assert.ok(dim.en.label.trim().length > 0 && dim.th.label.trim().length > 0);
    assert.ok(dim.en.prompt.trim().length > 10 && dim.th.prompt.trim().length > 10);
    assert.ok(dim.en.lowAnchor.trim().length > 0 && dim.en.highAnchor.trim().length > 0);
    assert.ok(dim.th.lowAnchor.trim().length > 0 && dim.th.highAnchor.trim().length > 0);
  }
  const perfDim = NASA_TLX_DIMENSIONS.find((d) => d.key === "performance_rating");
  assert.equal(perfDim?.invertForWorkload, true);
});

test("CSUQ preview calculation, N/A exclusion, and step validation enforce spec rules", () => {
  const answers: Record<string, CSUQAnswerValue | undefined> = {
    q1: 6, q2: 7, q3: 6, q4: 6, q5: 7, q6: 6,
    q7: null, q8: 6, q9: 6, q10: 6, q11: 7, q12: 6,
    q13: 7, q14: 6, q15: 6, q16: 7,
  };

  assert.equal(canAdvanceCSUQStep(answers), true);
  const preview = computeCSUQScoresPreview(answers);
  assert.equal(preview.sysuse, 6.33);
  assert.equal(preview.infoqual, 6.2);
  assert.equal(preview.interqual, 6.33);
  assert.equal(preview.overall, 6.33);
  assert.equal(preview.validOverallCount, 15);

  // Missing q16 -> cannot advance
  const incomplete = { ...answers, q16: undefined };
  assert.equal(canAdvanceCSUQStep(incomplete), false);

  // All 16 N/A -> cannot advance
  const allNA: Record<string, CSUQAnswerValue | undefined> = {};
  for (const item of CSUQ_ITEMS) {
    allNA[item.key] = null;
  }
  assert.equal(canAdvanceCSUQStep(allNA), false);
});

test("NASA-TLX preview inverts performance rating and step validation checks confirmation", () => {
  const tlx = {
    mental_demand: 60,
    physical_demand: 10,
    temporal_demand: 45,
    performance_rating: 85,
    effort: 55,
    frustration: 20,
  };
  const preview = computeNASATLXPreview(tlx);
  assert.equal(preview.performanceWorkload, 15);
  assert.equal(preview.overallWorkload, 34.17);

  const partialConfirmed = {
    mental_demand: true,
    physical_demand: true,
    temporal_demand: true,
    performance_rating: true,
    effort: true,
    frustration: false,
  };
  assert.equal(canAdvanceTLXStep(partialConfirmed), false);
  assert.equal(canAdvanceTLXStep({ ...partialConfirmed, frustration: true }), true);
});

test("Draft serialization preserves answers safely across refreshes and rejects corrupt storage", () => {
  assert.equal(getQuestionnaireDraftStorageKey("user-123"), "cybersim_questionnaire_draft:user-123");
  assert.equal(getQuestionnaireDraftStorageKey(""), "cybersim_questionnaire_draft:anon");

  const draft = createEmptyQuestionnaireDraft();
  draft.step = 2;
  draft.csuq.q1 = 7;
  draft.csuq.q7 = null;
  draft.tlx.mental_demand = 75;
  draft.tlxConfirmed.mental_demand = true;
  draft.qualitative.most_impressive = "Clear telemetry walkthrough";

  const encoded = serializeQuestionnaireDraft(draft);
  const decoded = parseQuestionnaireDraft(encoded);
  assert.equal(decoded.step, 2);
  assert.equal(decoded.csuq.q1, 7);
  assert.equal(decoded.csuq.q7, null);
  assert.equal(decoded.tlx.mental_demand, 75);
  assert.equal(decoded.tlxConfirmed.mental_demand, true);
  assert.equal(decoded.qualitative.most_impressive, "Clear telemetry walkthrough");

  const corrupted = parseQuestionnaireDraft("{not-json");
  assert.equal(corrupted.step, 1);
  assert.equal(Object.keys(corrupted.csuq).length, 0);
});

test("Qualitative feedback validates Unicode character limit accurately", () => {
  const okText = "ก".repeat(MAX_QUALITATIVE_CHARS);
  assert.equal(countQualitativeChars(okText), MAX_QUALITATIVE_CHARS);
  assert.equal(isQualitativeStepValid({
    most_impressive: okText,
    areas_to_improve: "",
    additional_comments: "",
  }), true);

  const overText = "ก".repeat(MAX_QUALITATIVE_CHARS + 1);
  assert.equal(isQualitativeStepValid({
    most_impressive: overText,
    areas_to_improve: "",
    additional_comments: "",
  }), false);
});
