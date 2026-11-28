import test from "node:test";
import assert from "node:assert/strict";
import {
  NASA_TLX_WORKLOAD_AXIS_KEYS,
  buildCSUQDistributionData,
  buildCohortFunnelData,
  buildPersonalCSUQData,
  buildPersonalProgressData,
  buildPrePostPairedData,
  buildSpearmanScatterSeries,
  buildTLXRadarData,
  isCorrelationChartUnlocked,
} from "../src/lib/evaluationCharts.ts";
import { normalizeClonedSvgForExport, serializeSvgForExport } from "../src/lib/svgExport.ts";

test("buildTLXRadarData uses inverted performance_workload across 6 canonical axes clamped to 0..100", () => {
  const labels = {
    mental_demand: "Mental",
    physical_demand: "Physical",
    temporal_demand: "Temporal",
    performance_workload: "Perf Load",
    effort: "Effort",
    frustration: "Frustration",
  };

  const radar = buildTLXRadarData(
    {
      mental_demand: 75,
      physical_demand: 15,
      temporal_demand: 60,
      performance_rating: 90,
      performance_workload: 10,
      effort: 55,
      frustration: 25,
    },
    labels,
  );

  assert.deepEqual(
    radar.map((d) => d.key),
    [...NASA_TLX_WORKLOAD_AXIS_KEYS],
  );
  assert.equal(radar[0].axis, "Mental");
  assert.equal(radar[0].workload, 75);
  assert.equal(radar[0].benchmark, 50);

  const perfAxis = radar.find((d) => d.key === "performance_workload");
  assert.equal(perfAxis?.workload, 10);

  // Fallback inversion when performance_workload is absent but performance_rating exists
  const fallbackRadar = buildTLXRadarData(
    {
      mental_demand: 120,
      physical_demand: -10,
      temporal_demand: Number.NaN,
      performance_rating: 85,
      effort: 40,
      frustration: 30,
    },
    labels,
  );
  assert.equal(fallbackRadar[0].workload, 100);
  assert.equal(fallbackRadar[1].workload, 0);
  assert.equal(fallbackRadar[2].workload, 0);
  assert.equal(fallbackRadar.find((d) => d.key === "performance_workload")?.workload, 15);
});

test("buildPersonalProgressData and buildPersonalCSUQData build bounded personal learner charts", () => {
  const bars = buildPersonalProgressData(46.67, 83.33, { pre: "Pre-Test", post: "Post-Test" });
  assert.deepEqual(bars, [
    { key: "pre", stage: "Pre-Test", score: 46.67 },
    { key: "post", stage: "Post-Test", score: 83.33 },
  ]);

  const clamped = buildPersonalProgressData(-5, 112, { pre: "Pre", post: "Post" });
  assert.deepEqual(clamped, [
    { key: "pre", stage: "Pre", score: 0 },
    { key: "post", stage: "Post", score: 100 },
  ]);

  const csuq = buildPersonalCSUQData(
    { sysuse: 6.5, infoqual: null, interqual: 5.25, overall: 6.1 },
    { sysuse: "SYSUSE", infoqual: "INFOQUAL", interqual: "INTERQUAL", overall: "OVERALL" },
  );
  assert.deepEqual(csuq, [
    { key: "sysuse", subscale: "SYSUSE", score: 6.5 },
    { key: "infoqual", subscale: "INFOQUAL", score: null },
    { key: "interqual", subscale: "INTERQUAL", score: 5.25 },
    { key: "overall", subscale: "OVERALL", score: 6.1 },
  ]);
});

test("serializeSvgForExport produces self-contained standalone SVG XML with backdrop and dimensions", () => {
  const raw = `<svg class="recharts-surface" width="480" height="260"><g><rect width="40" height="100" fill="#10b981"/></g></svg>`;
  const out = serializeSvgForExport(raw, {
    width: 480,
    height: 260,
    title: "My Learning Progress",
    backgroundColor: "#18181b",
  });

  assert.ok(out.startsWith('<?xml version="1.0" encoding="UTF-8"?>'));
  assert.ok(out.includes('xmlns="http://www.w3.org/2000/svg"'));
  assert.ok(out.includes('viewBox="0 0 480 260"'));
  assert.ok(out.includes('fill="#18181b"'));
  assert.ok(out.includes("IBM Plex Sans Thai"));
  assert.ok(out.includes("<title>My Learning Progress</title>"));
});

test("normalizeClonedSvgForExport strips transient tooltip cursor elements and fills missing text color", () => {
  const removedSelectors: string[] = [];
  const textNode = {
    attrs: new Map<string, string>(),
    getAttribute(name: string) {
      return this.attrs.get(name) ?? null;
    },
    setAttribute(name: string, val: string) {
      this.attrs.set(name, val);
    },
  };
  const fakeClone = {
    querySelectorAll(selector: string) {
      if (selector.includes("tooltip")) {
        return [{ remove: () => removedSelectors.push("tooltip") }];
      }
      if (selector.includes("text")) {
        return [textNode];
      }
      return [];
    },
  } as unknown as SVGSVGElement;

  normalizeClonedSvgForExport(fakeClone);
  assert.deepEqual(removedSelectors, ["tooltip"]);
  assert.equal(textNode.getAttribute("fill"), "#e4e4e7");
});

test("buildCohortFunnelData computes 4-stage funnel counts cleanly", () => {
  const stages = buildCohortFunnelData(
    {
      total_learners: 40,
      completed_pre_test_learners: 36,
      completed_post_test_learners: 28,
      questionnaire_respondents: 20,
    },
    {
      totalLearners: "Registered",
      preCompleted: "Pre-Test",
      postCompleted: "Post-Test",
      respondents: "Survey",
    },
  );
  assert.equal(stages.length, 4);
  assert.deepEqual(stages[0], { key: "total_learners", stage: "Registered", count: 40 });
  assert.deepEqual(stages[3], { key: "questionnaire_respondents", stage: "Survey", count: 20 });

  const zeroTotal = buildCohortFunnelData(
    {
      total_learners: -2,
      completed_pre_test_learners: 0,
      completed_post_test_learners: 0,
      questionnaire_respondents: 0,
    },
    {
      totalLearners: "Registered",
      preCompleted: "Pre-Test",
      postCompleted: "Post-Test",
      respondents: "Survey",
    },
  );
  assert.equal(zeroTotal[0].count, 0);
});

test("buildCSUQDistributionData computes 100% stacked Likert + N/A distributions for subscales and 16 items", () => {
  assert.deepEqual(buildCSUQDistributionData([], "subscales", {}), []);

  const mkItems = (overrides: Record<string, number | null>) => {
    const base: Record<string, number | null> = {};
    for (let i = 1; i <= 16; i++) base[`q${i}`] = 6;
    return { ...base, ...overrides };
  };

  const rows = [
    { csuq_scores: { items: mkItems({ q1: 2, q7: null, q13: 5 }) } },
    { csuq_scores: { items: mkItems({ q1: 4, q7: 7, q13: 3 }) } },
    { csuq_scores: { items: mkItems({ q1: 7, q7: 5, q13: 6 }) } },
  ];

  const subscales = buildCSUQDistributionData(rows, "subscales", {
    sysuse: "SYSUSE",
    infoqual: "INFOQUAL",
    interqual: "INTERQUAL",
    overall: "OVERALL",
  });
  assert.equal(subscales.length, 4);
  for (const row of subscales) {
    const sumPct = Number((row.lowPct + row.neutralPct + row.positivePct + row.strongPct + row.naPct).toFixed(2));
    assert.equal(sumPct, 100, `Expected 100% sum on subscale ${row.key}, got ${sumPct}`);
  }
  const infoqual = subscales.find((s) => s.key === "infoqual")!;
  assert.equal(infoqual.naCount, 1);
  assert.equal(infoqual.n, 17);

  const itemRows = buildCSUQDistributionData(rows, "items", { q1: "Q1", q7: "Q7" });
  assert.equal(itemRows.length, 16);
  const q1 = itemRows.find((r) => r.key === "q1")!;
  assert.equal(q1.lowCount, 1);
  assert.equal(q1.neutralCount, 1);
  assert.equal(q1.strongCount, 1);
  assert.equal(q1.mean, 4.33);
  const q7 = itemRows.find((r) => r.key === "q7")!;
  assert.equal(q7.naCount, 1);
  assert.equal(q7.n, 2);
});

test("buildPrePostPairedData sorts eval_date ASC when timestamps exist, preserves server order without eval_date, and honors server means", () => {
  const dated = buildPrePostPairedData([
    { participant_code: "P-AAA00002", pre_score: 60, post_score: 90, eval_date: "2026-11-28T12:00:00Z" },
    { participant_code: "P-ZZZ00001", pre_score: 40, post_score: 70, eval_date: "2026-11-28T10:00:00Z" },
  ]);
  assert.deepEqual(
    dated.rows.map((r) => r.label),
    ["P-ZZZ00001", "P-AAA00002"],
  );
  assert.equal(dated.preMean, 50);
  assert.equal(dated.postMean, 80);

  // Backend /statistics participant_points omits eval_date and is already ordered chronologically (eval_date ASC)
  const serverOrdered = buildPrePostPairedData(
    [
      { participant_code: "P-ZZZ11111", pre_score: 30, post_score: 60 },
      { participant_code: "P-AAA22222", pre_score: 50, post_score: 90 },
    ],
    { preMean: 41.25, postMean: 76.5 },
  );
  assert.deepEqual(
    serverOrdered.rows.map((r) => r.label),
    ["P-ZZZ11111", "P-AAA22222"],
  );
  assert.equal(serverOrdered.preMean, 41.25);
  assert.equal(serverOrdered.postMean, 76.5);
});

test("buildSpearmanScatterSeries filters nulls and enforces n >= 3 scatter guard threshold", () => {
  assert.equal(isCorrelationChartUnlocked(0), false);
  assert.equal(isCorrelationChartUnlocked(2), false);
  assert.equal(isCorrelationChartUnlocked(3), true);

  const series = buildSpearmanScatterSeries([
    {
      participant_code: "P-001",
      pre_score: 40,
      post_score: 75,
      improvement_pct: 35,
      csuq_overall: 6.4,
      nasa_tlx_overall: 32.5,
    },
    {
      participant_code: "P-002",
      pre_score: 50,
      post_score: 70,
      improvement_pct: 20,
      csuq_overall: null,
      nasa_tlx_overall: 45,
    },
    {
      participant_code: "P-003",
      pre_score: 55,
      post_score: 80,
      improvement_pct: 25,
      csuq_overall: 5.9,
      nasa_tlx_overall: null,
    },
  ]);

  // 2 complete pairs remain < 3 threshold
  assert.equal(series.gainVsCsuq.length, 2);
  assert.equal(isCorrelationChartUnlocked(series.gainVsCsuq.length), false);
  assert.deepEqual(
    series.gainVsCsuq.map((p) => p.participantCode),
    ["P-001", "P-003"],
  );
  assert.deepEqual(series.gainVsCsuq[0], {
    participantCode: "P-001",
    x: 35,
    y: 6.4,
  });

  assert.equal(series.gainVsTlx.length, 2);
  assert.equal(series.csuqVsTlx.length, 1);
});
