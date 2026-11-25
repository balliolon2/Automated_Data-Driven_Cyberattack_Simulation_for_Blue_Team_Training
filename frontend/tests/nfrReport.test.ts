import test from "node:test";
import assert from "node:assert/strict";
import {
  NFR_METHODOLOGY_PILLARS,
  buildNfrEvaluationParagraph,
  buildNfrLatexTable,
  buildNfrMarkdownTable,
  computeNfrBaselineDelta,
  type NFRReportResponse,
} from "../src/lib/nfrReport.ts";

function makeSampleNfrReport(overrides: Partial<NFRReportResponse> = {}): NFRReportResponse {
  return {
    run_id: "run-123",
    run_label: "Local Docker Baseline_v1",
    environment: "local-docker",
    source: "cli_benchmark",
    rubric_version: "nfr-rubric-v1",
    evaluated_at: "2026-11-25T09:00:00Z",
    summary: {
      overall_status: "partial_pending",
      total_dimensions: 10,
      passed_count: 8,
      warn_count: 0,
      fail_count: 0,
      not_measured_count: 2,
      measured_pass_rate: 100,
    },
    runtime: {
      process_started_at: "2026-11-25T08:00:00Z",
      window_started_at: "2026-11-25T08:00:00Z",
      uptime_seconds: 3600,
      gin_mode: "release",
      total_requests: 300,
      non_llm_requests: 295,
      llm_requests: 5,
      status_2xx: 298,
      status_3xx: 0,
      status_4xx: 2,
      status_5xx: 0,
      availability_pct: 100,
      error_rate_5xx_pct: 0,
      in_flight_current: 0,
      peak_concurrent_requests: 25,
      db_status: "up",
      db_latency_ms: 2.8,
      non_llm_latency: { count: 295, min: 1.1, mean: 14.5, p50: 11.2, p95: 42.6, p99: 78.4, max: 95.0 },
      llm_latency: { count: 5, min: 2100, mean: 4200, p50: 4100, p95: 6200, p99: 6500, max: 6600 },
      llm_calls: {
        total_calls: 5,
        success_count: 4,
        fallback_count: 1,
        error_count: 0,
        total_retries: 1,
        providers: ["https://api.openai.com"],
        models: ["glm-5.3-flash"],
        outcome_counts: { ai_success: 4, fallback: 1 },
        latency: { count: 5, min: 2100, mean: 4200, p50: 4100, p95: 6200, p99: 6500, max: 6600 },
      },
      endpoints: [
        { route: "GET /api/health", category: "non_llm_api", count: 100, error_count: 0, mean_ms: 2.1, p50_ms: 1.9, p95_ms: 4.5, p99_ms: 8.2, max_ms: 11.0 },
      ],
    },
    dimensions: [
      {
        id: "performance",
        iso_characteristic: "Performance Efficiency (Time Behaviour)",
        status: "pass",
        summary_metric: "p95 42.6 ms | DB 2.8 ms",
        target_expression: "API p95 <= 300 ms, DB <= 50 ms",
        evidence_source: "HTTP Ring & DB Ping",
        checks: [],
      },
      {
        id: "availability",
        iso_characteristic: "Reliability (Availability)",
        status: "pass",
        summary_metric: "100.00% Non-5xx | DB up",
        target_expression: "DB up & Non-5xx >= 99.0%",
        evidence_source: "SELECT 1 + HTTP Counters",
        checks: [],
      },
      {
        id: "scalability",
        iso_characteristic: "Performance Efficiency (Capacity)",
        status: "pass",
        summary_metric: "25 workers | 412.5 RPS | p95 68.4 ms",
        target_expression: ">= 10 workers, p95 <= 500 ms, Error <= 1.0%",
        evidence_source: "CLI Benchmark",
        checks: [],
      },
      {
        id: "reliability",
        iso_characteristic: "Reliability (Fault Tolerance & Recoverability)",
        status: "pass",
        summary_metric: "5xx 0.00% | SysFail 0/8 | Snap 8/8",
        target_expression: "5xx <= 1%, SysFail = 0%",
        evidence_source: "HTTP + DB",
        checks: [],
      },
      {
        id: "security",
        iso_characteristic: "Security",
        status: "pass",
        summary_metric: "6/6 Probes | 0 non-bcrypt",
        target_expression: "100% Auth/RBAC probes & posture",
        evidence_source: "httptest + Posture",
        checks: [],
      },
      {
        id: "data_integrity",
        iso_characteristic: "Functional Suitability & Integrity",
        status: "pass",
        summary_metric: "Overlap 0 | Orphans 0 | Q=540",
        target_expression: "0 overlaps & orphans",
        evidence_source: "PostgreSQL Audit",
        checks: [],
      },
      {
        id: "reproducibility",
        iso_characteristic: "Auditability & Replayability",
        status: "pass",
        summary_metric: "Link 100.0% (8/8) | Meta 100.0%",
        target_expression: "100% snapshot coverage",
        evidence_source: "scenario_snapshots",
        checks: [],
      },
      {
        id: "usability",
        iso_characteristic: "Interaction Capability (Usability)",
        status: "insufficient_data",
        summary_metric: "Insufficient Data (0 questionnaires)",
        target_expression: "CSUQ >= 5.0 & TLX <= 55",
        evidence_source: "evaluation_metrics",
        checks: [],
      },
      {
        id: "maintainability",
        iso_characteristic: "Maintainability",
        status: "not_measured",
        summary_metric: "Not Measured (awaiting CI report)",
        target_expression: "100% CI gates & >= 60% Go cov",
        evidence_source: "CI Gate Report",
        checks: [],
      },
      {
        id: "safety",
        iso_characteristic: "Safety / Freedom from Risk",
        status: "pass",
        summary_metric: "0 safety boundary violations",
        target_expression: "0 Thai/secret/boilerplate violations",
        evidence_source: "DB Safety Scan",
        checks: [],
      },
    ],
    scalability: {
      mode: "cli_benchmark",
      executed_at: "2026-11-25T08:59:00Z",
      target_endpoints: ["/api/health", "/api/exams/sample"],
      levels: [
        {
          concurrency: 10,
          total_requests: 100,
          success_count: 100,
          error_count: 0,
          error_rate_pct: 0,
          throughput_rps: 310.4,
          mean_ms: 28.2,
          p50_ms: 25.0,
          p95_ms: 52.1,
          p99_ms: 71.0,
          max_ms: 82.0,
          duration_ms: 322.1,
        },
        {
          concurrency: 25,
          total_requests: 100,
          success_count: 100,
          error_count: 0,
          error_rate_pct: 0,
          throughput_rps: 412.5,
          mean_ms: 35.6,
          p50_ms: 31.2,
          p95_ms: 68.4,
          p99_ms: 94.2,
          max_ms: 112.0,
          duration_ms: 242.4,
        },
      ],
    },
    reliability_breakdown: {
      total_sessions: 8,
      ai_generated_count: 5,
      static_fallback_count: 3,
      system_failure_count: 0,
      valid_snapshot_count: 8,
      total_snapshot_count: 8,
      fallback_reasons: { "LLM_API_KEY not configured": 3 },
    },
    thresholds: {
      rubric_version: "nfr-rubric-v1",
      non_llm_p95_target_ms: 300,
      non_llm_p95_warn_ms: 500,
      non_llm_p99_target_ms: 800,
      non_llm_p99_warn_ms: 1200,
      llm_p95_target_ms: 20000,
      llm_p95_warn_ms: 45000,
      db_ping_target_ms: 50,
      db_ping_warn_ms: 150,
      availability_target_pct: 99,
      availability_warn_pct: 95,
      scalability_p95_target_ms: 500,
      scalability_p95_warn_ms: 1000,
      scalability_error_target_pct: 1,
      scalability_error_warn_pct: 5,
      coverage_target_pct: 60,
      coverage_warn_pct: 40,
      csuq_target_mean: 5,
      csuq_warn_mean: 4,
      tlx_target_max: 55,
      tlx_warn_max: 70,
    },
    ...overrides,
  };
}

test("buildNfrLatexTable emits escaped IEEEtran booktabs matrix and scalability sub-table", () => {
  const report = makeSampleNfrReport();
  const tex = buildNfrLatexTable(report);

  assert.match(tex, /\\begin\{table\*\}\[t\]/);
  assert.match(tex, /\\toprule/);
  assert.match(tex, /\\bottomrule/);
  assert.match(tex, /\\label\{tab:nfr_evaluation\}/);
  assert.match(tex, /\\label\{tab:nfr_scalability\}/);
  assert.ok(tex.includes("Evidence Source"), "LaTeX table must include Evidence Source column");
  // Special characters must be properly escaped
  assert.ok(tex.includes("Local Docker Baseline\\_v1"), "underscore in run label must be escaped");
  assert.ok(tex.includes("100.00\\%"), "percent sign must be escaped");
  assert.ok(!tex.includes("NaN") && !tex.includes("undefined"));
});

test("buildNfrMarkdownTable and buildNfrEvaluationParagraph stay honest and support EN/TH", () => {
  const report = makeSampleNfrReport();
  const md = buildNfrMarkdownTable(report);
  assert.match(md, /\| Dimension \| ISO\/IEC 25010 Characteristic \| Target \| Measured Result \| Evidence Source \| Verdict \|/);
  assert.ok(md.includes("p95 42.6 ms \\| DB 2.8 ms"), "pipe inside cell must be escaped for GFM");
  assert.match(md, /### Scalability Concurrency Curve/);
  assert.ok(!md.includes("NaN") && !md.includes("undefined"));

  const enPara = buildNfrEvaluationParagraph(report, "en");
  assert.match(enPara, /ISO\/IEC 25010/);
  assert.match(enPara, /2026-11-25/);
  assert.match(enPara, /Security/);
  assert.match(enPara, /Safety/);
  assert.match(enPara, /42\.6 ms/);
  assert.match(enPara, /412\.5 RPS/);
  assert.match(enPara, /awaiting data|pending/i);
  assert.ok(!enPara.includes("NaN") && !enPara.includes("undefined"));

  const thPara = buildNfrEvaluationParagraph(report, "th");
  assert.ok(/[\u0E00-\u0E7F]/.test(thPara), "Expected Thai prose in TH paragraph");
  assert.match(thPara, /2026-11-25/);
  assert.match(thPara, /Security/);
  assert.match(thPara, /Safety/);
  assert.match(thPara, /ISO\/IEC 25010/);
  assert.match(thPara, /42\.6 ms/);
  assert.ok(!thPara.includes("NaN") && !thPara.includes("undefined"));
});

test("computeNfrBaselineDelta computes signed improvements accurately", () => {
  const baseline = makeSampleNfrReport();
  const current = makeSampleNfrReport({
    summary: {
      overall_status: "all_passed",
      total_dimensions: 10,
      passed_count: 10,
      warn_count: 0,
      fail_count: 0,
      not_measured_count: 0,
      measured_pass_rate: 100,
    },
    runtime: {
      ...baseline.runtime,
      availability_pct: 99.5,
      non_llm_latency: { ...baseline.runtime.non_llm_latency, p95: 30.6 },
    },
    scalability: {
      ...baseline.scalability!,
      levels: [
        { ...baseline.scalability!.levels[0], throughput_rps: 450.0 },
      ],
    },
  });

  const delta = computeNfrBaselineDelta(current, baseline);
  assert.equal(delta.nonLlmP95DeltaMs, -12.0);
  assert.equal(delta.availabilityDeltaPct, -0.5);
  assert.equal(delta.passedCountDelta, 2);
  assert.equal(delta.peakRpsDelta, 37.5);
  assert.ok(NFR_METHODOLOGY_PILLARS.length >= 5);
});
