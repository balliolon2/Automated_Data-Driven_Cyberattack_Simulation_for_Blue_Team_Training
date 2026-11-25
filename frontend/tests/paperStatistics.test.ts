import test from "node:test";
import assert from "node:assert/strict";
import {
  buildAcademicResultsParagraph,
  buildAllLatexTables,
  buildLatexTable1,
  buildLatexTable1Items,
  buildLatexTable2,
  buildLatexTable3,
  buildMarkdownTable1,
  buildMarkdownTable1Items,
  buildMarkdownTable2,
  buildMarkdownTable3,
  escapeLatex,
  type AdminPaperStatisticsResponse,
  type DescriptiveStat,
} from "../src/lib/paperStatistics.ts";

function makeDesc(key: string, n: number, mean: number | null, sd: number | null, band = "excellent"): DescriptiveStat {
  return {
    key,
    n,
    mean,
    sd,
    se: n >= 2 && sd !== null ? Number((sd / Math.sqrt(n)).toFixed(2)) : null,
    ci_low: n >= 2 && mean !== null && sd !== null ? Number((mean - 0.5).toFixed(2)) : null,
    ci_high: n >= 2 && mean !== null && sd !== null ? Number((mean + 0.5).toFixed(2)) : null,
    median: mean,
    q1: mean !== null ? Number((mean - 0.25).toFixed(2)) : null,
    q3: mean !== null ? Number((mean + 0.25).toFixed(2)) : null,
    iqr: n >= 1 ? 0.5 : null,
    min: mean !== null ? Number((mean - 1.0).toFixed(2)) : null,
    max: mean !== null ? Number((mean + 1.0).toFixed(2)) : null,
    band,
    status: n >= 2 ? "ok" : "insufficient_n",
  };
}

function makeSampleStats(n: number, pilot: boolean): AdminPaperStatisticsResponse {
  const ok2 = n >= 2;
  const ok3 = n >= 3;
  const ok5 = n >= 5;
  return {
    generated_at: "2026-11-25T12:00:00Z",
    protocol_version: "protocol-v1",
    csuq_schema_version: "csuq-v3-asc-7",
    nasa_tlx_schema_version: "nasa-tlx-rtlx-v1",
    is_pilot_sample: pilot,
    cohort: {
      total_learners: 40,
      completed_pre_test_learners: 36,
      completed_post_test_learners: 32,
      questionnaire_respondents: n,
      response_rate_pct: n > 0 ? Number(((n / 32) * 100).toFixed(2)) : 0,
    },
    csuq_subscales: [
      makeDesc("sysuse", n, n > 0 ? 6.2 : null, ok2 ? 0.45 : null, "excellent"),
      makeDesc("infoqual", n, n > 0 ? 6.05 : null, ok2 ? 0.52 : null, "excellent"),
      makeDesc("interqual", n, n > 0 ? 6.3 : null, ok2 ? 0.41 : null, "excellent"),
      makeDesc("overall", n, n > 0 ? 6.18 : null, ok2 ? 0.44 : null, "excellent"),
    ],
    csuq_items: Array.from({ length: 16 }, (_, idx) =>
      makeDesc(`q${idx + 1}`, n, n > 0 ? 6.1 : null, ok2 ? 0.5 : null, "excellent"),
    ),
    nasa_tlx: [
      makeDesc("mental_demand", n, n > 0 ? 55 : null, ok2 ? 10 : null, "high"),
      makeDesc("physical_demand", n, n > 0 ? 12 : null, ok2 ? 5 : null, "low"),
      makeDesc("temporal_demand", n, n > 0 ? 40 : null, ok2 ? 8 : null, "moderate"),
      makeDesc("performance_rating", n, n > 0 ? 85 : null, ok2 ? 7 : null, "neutral"),
      makeDesc("performance_workload", n, n > 0 ? 15 : null, ok2 ? 7 : null, "low"),
      makeDesc("effort", n, n > 0 ? 50 : null, ok2 ? 9 : null, "high"),
      makeDesc("frustration", n, n > 0 ? 22 : null, ok2 ? 6 : null, "low"),
      makeDesc("overall_workload", n, n > 0 ? 32.33 : null, ok2 ? 6.2 : null, "moderate"),
    ],
    reliability: [
      { key: "csuq_overall", item_count: 16, complete_cases_n: n, excluded_na_cases: 0, alpha: ok2 ? 0.912 : null, band: ok2 ? "excellent" : "insufficient_n", status: ok2 ? "ok" : "insufficient_n" },
      { key: "csuq_sysuse", item_count: 6, complete_cases_n: n, excluded_na_cases: 0, alpha: ok2 ? 0.87 : null, band: ok2 ? "good" : "insufficient_n", status: ok2 ? "ok" : "insufficient_n" },
      { key: "csuq_infoqual", item_count: 6, complete_cases_n: n, excluded_na_cases: 0, alpha: ok2 ? 0.84 : null, band: ok2 ? "good" : "insufficient_n", status: ok2 ? "ok" : "insufficient_n" },
      { key: "csuq_interqual", item_count: 3, complete_cases_n: n, excluded_na_cases: 0, alpha: ok2 ? 0.81 : null, band: ok2 ? "good" : "insufficient_n", status: ok2 ? "ok" : "insufficient_n" },
      { key: "nasa_tlx_overall", item_count: 6, complete_cases_n: n, excluded_na_cases: 0, alpha: ok2 ? 0.76 : null, band: ok2 ? "acceptable" : "insufficient_n", status: ok2 ? "ok" : "insufficient_n" },
    ],
    learning_gain: {
      pre: makeDesc("pre_score", n, n > 0 ? 56.5 : null, ok2 ? 8.1 : null, "neutral"),
      post: makeDesc("post_score", n, n > 0 ? 84.2 : null, ok2 ? 6.4 : null, "neutral"),
      difference: makeDesc("improvement_pct", n, n > 0 ? 27.7 : null, ok2 ? 7.2 : null, "neutral"),
      paired_t_test: {
        n,
        df: n > 0 ? n - 1 : 0,
        mean_diff: ok3 ? 27.7 : null,
        sd_diff: ok3 ? 7.2 : null,
        ci_low: ok3 ? 23.1 : null,
        ci_high: ok3 ? 32.3 : null,
        t_stat: ok3 ? 11.42 : null,
        p_value: ok3 ? 0.0001 : null,
        cohens_dz: ok3 ? 3.85 : null,
        effect_band: ok3 ? "large" : "insufficient_n",
        status: ok3 ? "ok" : "insufficient_n",
      },
      wilcoxon: {
        n,
        effective_n: n,
        zero_diff_count: 0,
        method: ok5 ? "exact" : "insufficient_n",
        w_stat: ok5 ? 0 : null,
        w_plus: ok5 ? 15 : null,
        w_minus: ok5 ? 0 : null,
        p_value: ok5 ? 0.002 : null,
        rank_biserial: ok5 ? 1.0 : null,
        status: ok5 ? "ok" : "insufficient_n",
      },
    },
    correlations: [
      { key: "learning_gain_vs_csuq_overall", n, rho: ok3 ? 0.54 : null, p_value: ok3 ? 0.014 : null, band: ok3 ? "strong" : "insufficient_n", status: ok3 ? "ok" : "insufficient_n" },
      { key: "learning_gain_vs_nasa_tlx_overall", n, rho: ok3 ? -0.32 : null, p_value: ok3 ? 0.12 : null, band: ok3 ? "moderate" : "insufficient_n", status: ok3 ? "ok" : "insufficient_n" },
      { key: "csuq_overall_vs_nasa_tlx_overall", n, rho: ok3 ? -0.45 : null, p_value: ok3 ? 0.04 : null, band: ok3 ? "moderate" : "insufficient_n", status: ok3 ? "ok" : "insufficient_n" },
    ],
  };
}

test("escapeLatex escapes special TeX characters cleanly", () => {
  assert.equal(escapeLatex("100% score_a & #1 $5"), "100\\% score\\_a \\& \\#1 \\$5");
});

test("LaTeX and Markdown generators render IEEEtran booktabs tables without NaN or undefined", () => {
  const full = makeSampleStats(32, false);
  const t1 = buildLatexTable1(full);
  const t1Items = buildLatexTable1Items(full);
  const t2 = buildLatexTable2(full);
  const t3 = buildLatexTable3(full);
  const bundle = buildAllLatexTables(full);

  for (const tex of [t1, t1Items, t2, t3, bundle]) {
    assert.ok(tex.includes("\\begin{table*}[t]"), "Expected IEEEtran wide table* float");
    assert.ok(tex.includes("\\toprule") && tex.includes("\\midrule") && tex.includes("\\bottomrule"), "Expected booktabs rules");
    assert.ok(!tex.includes("NaN") && !tex.includes("undefined"), "Must never emit NaN/undefined");
  }
  assert.ok(t1.includes("6.18"));
  assert.ok(t1.includes("0.912"));
  assert.ok(t1.includes("Min--Max") && t1.includes("Band") && t1.includes("Excellent"));
  assert.ok(t1Items.includes("Min--Max") && t1Items.includes("Band"));
  assert.ok(t2.includes("32.33") && t2.includes("Min--Max") && t2.includes("Band"));
  assert.ok(t3.includes("11.420") || t3.includes("11.42"));
  assert.ok(t3.includes("0.540") || t3.includes("0.54"));

  const md1 = buildMarkdownTable1(full);
  const mdItems = buildMarkdownTable1Items(full);
  const md2 = buildMarkdownTable2(full);
  const md3 = buildMarkdownTable3(full);
  for (const md of [md1, mdItems, md2, md3]) {
    assert.ok(md.includes("| --- |"), "Expected GFM markdown separator");
    assert.ok(!md.includes("NaN") && !md.includes("undefined"));
  }
  assert.ok(md1.includes("Min–Max") && md1.includes("Excellent"));
  assert.ok(mdItems.includes("Min–Max"));
  assert.ok(md2.includes("Min–Max"));
});

test("Small-n payloads render N/A gates cleanly in LaTeX, Markdown, and Academic Paragraphs", () => {
  for (const n of [0, 1, 2, 5]) {
    const stats = makeSampleStats(n, true);
    const bundle = buildAllLatexTables(stats);
    const md3 = buildMarkdownTable3(stats);
    const enPara = buildAcademicResultsParagraph(stats, "en");
    const thPara = buildAcademicResultsParagraph(stats, "th");

    for (const out of [bundle, md3, enPara, thPara]) {
      assert.ok(!out.includes("NaN") && !out.includes("undefined"), `Output for n=${n} leaked NaN/undefined`);
    }

    if (n <= 2) {
      assert.ok(bundle.includes("\\text{N/A"), `Expected \\text{N/A for n=${n}`);
      assert.ok(md3.includes("N/A (requires n >= 3)"), `Expected friendly gate text in Markdown Table 3 for n=${n}`);
      assert.ok(!md3.includes("insufficient_n"), "Markdown Table 3 must not leak raw insufficient_n enum");
    }
    if (n === 5) {
      assert.ok(enPara.toLowerCase().includes("pilot"), "Expected pilot sample caveat in EN paragraph for n=5");
      assert.ok(thPara.includes("นำร่อง"), "Expected pilot sample caveat in TH paragraph for n=5");
    }
  }

  const full32 = makeSampleStats(32, false);
  const enFull = buildAcademicResultsParagraph(full32, "en");
  const thFull = buildAcademicResultsParagraph(full32, "th");
  assert.ok(enFull.includes("N = 32") && enFull.includes("6.18") && enFull.includes("32.33"));
  assert.ok(enFull.includes("NASA-TLX workload ρ = -0.320") && enFull.includes("CSUQ overall vs. NASA-TLX workload ρ = -0.450"));
  assert.ok(thFull.includes("N = 32") && thFull.includes("6.18") && thFull.includes("32.33"));
  assert.ok(thFull.includes("พัฒนาการคะแนน↔NASA-TLX") && thFull.includes("CSUQ↔NASA-TLX"));

  // Negative learning gain uses appropriate Thai verb rather than "เพิ่มขึ้นจาก"
  const negStats = makeSampleStats(2, true);
  negStats.learning_gain.pre = makeDesc("pre_score", 2, 23.33, 14.14, "neutral");
  negStats.learning_gain.post = makeDesc("post_score", 2, 13.34, 9.43, "neutral");
  negStats.learning_gain.difference = makeDesc("improvement_pct", 2, -9.99, 4.71, "neutral");
  const thNeg = buildAcademicResultsParagraph(negStats, "th");
  assert.ok(thNeg.includes("ลดลงจาก") && !thNeg.includes("เพิ่มขึ้นจาก"), "Expected ลดลงจาก for negative delta");
});

test("Omitted undefined stat fields (n=2 with n=1 Q16 N/A) render cleanly without throwing", () => {
  const sparse = makeSampleStats(2, true);
  // Simulate Go JSON omission where q16 only has n=1 valid response and inferential tests are n<3 / n<5
  sparse.csuq_items[15] = {
    key: "q16",
    group: "OVERALL",
    n: 1,
    mean: 6.0,
    band: "excellent",
    status: "insufficient_n",
  };
  sparse.reliability[0] = {
    key: "csuq_overall",
    item_count: 16,
    complete_cases_n: 1,
    excluded_na_cases: 1,
    band: "insufficient_n",
    status: "insufficient_n",
  };
  sparse.learning_gain.paired_t_test = {
    n: 2,
    df: 1,
    effect_band: "insufficient_n",
    status: "insufficient_n",
  };
  sparse.learning_gain.wilcoxon = {
    n: 2,
    effective_n: 2,
    zero_diff_count: 0,
    method: "insufficient_n",
    status: "insufficient_n",
  };
  sparse.correlations = [
    { key: "learning_gain_vs_csuq_overall", n: 2, band: "insufficient_n", status: "insufficient_n" },
  ];

  const outputs = [
    buildLatexTable1(sparse),
    buildLatexTable1Items(sparse),
    buildLatexTable2(sparse),
    buildLatexTable3(sparse),
    buildAllLatexTables(sparse),
    buildMarkdownTable1(sparse),
    buildMarkdownTable1Items(sparse),
    buildMarkdownTable2(sparse),
    buildMarkdownTable3(sparse),
    buildAcademicResultsParagraph(sparse, "en"),
    buildAcademicResultsParagraph(sparse, "th"),
  ];

  for (const out of outputs) {
    assert.ok(!out.includes("NaN") && !out.includes("undefined"));
  }
});
