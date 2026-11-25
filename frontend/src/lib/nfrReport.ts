import { escapeLatex, isFiniteStat } from "./paperStatistics.ts";

export interface NFRThresholds {
  rubric_version: string;
  non_llm_p95_target_ms: number;
  non_llm_p95_warn_ms: number;
  non_llm_p99_target_ms: number;
  non_llm_p99_warn_ms: number;
  llm_p95_target_ms: number;
  llm_p95_warn_ms: number;
  db_ping_target_ms: number;
  db_ping_warn_ms: number;
  availability_target_pct: number;
  availability_warn_pct: number;
  scalability_p95_target_ms: number;
  scalability_p95_warn_ms: number;
  scalability_error_target_pct: number;
  scalability_error_warn_pct: number;
  coverage_target_pct: number;
  coverage_warn_pct: number;
  csuq_target_mean: number;
  csuq_warn_mean: number;
  tlx_target_max: number;
  tlx_warn_max: number;
}

export interface NFRPercentiles {
  count: number;
  min?: number | null;
  mean?: number | null;
  p50?: number | null;
  p95?: number | null;
  p99?: number | null;
  max?: number | null;
}

export interface NFREndpointLatency {
  route: string;
  category: string;
  count: number;
  error_count: number;
  mean_ms?: number | null;
  p50_ms?: number | null;
  p95_ms?: number | null;
  p99_ms?: number | null;
  max_ms?: number | null;
}

export interface NFRLLMTelemetry {
  total_calls: number;
  success_count: number;
  fallback_count: number;
  error_count: number;
  total_retries: number;
  providers: string[];
  models: string[];
  outcome_counts: Record<string, number>;
  latency: NFRPercentiles;
}

export interface NFRRuntimeSnapshot {
  process_started_at: string;
  window_started_at: string;
  uptime_seconds: number;
  gin_mode: string;
  total_requests: number;
  non_llm_requests: number;
  llm_requests: number;
  status_2xx: number;
  status_3xx: number;
  status_4xx: number;
  status_5xx: number;
  availability_pct?: number | null;
  error_rate_5xx_pct?: number | null;
  in_flight_current: number;
  peak_concurrent_requests: number;
  db_status: string;
  db_latency_ms: number;
  non_llm_latency: NFRPercentiles;
  llm_latency: NFRPercentiles;
  llm_calls: NFRLLMTelemetry;
  endpoints: NFREndpointLatency[];
}

export interface NFRCheckItem {
  key: string;
  measured: string;
  unit: string;
  target: string;
  status: "pass" | "warn" | "fail" | "not_measured" | "insufficient_data" | string;
}

export interface NFRDimensionResult {
  id: string;
  iso_characteristic: string;
  status: "pass" | "warn" | "fail" | "not_measured" | "insufficient_data" | string;
  summary_metric: string;
  target_expression: string;
  evidence_source: string;
  checks: NFRCheckItem[];
}

export interface NFRSummary {
  overall_status: "all_passed" | "pass_with_warnings" | "partial_pending" | "fail" | string;
  total_dimensions: number;
  passed_count: number;
  warn_count: number;
  fail_count: number;
  not_measured_count: number;
  measured_pass_rate: number;
}

export interface NFRReliabilityBreakdown {
  total_sessions: number;
  ai_generated_count: number;
  static_fallback_count: number;
  system_failure_count: number;
  valid_snapshot_count: number;
  total_snapshot_count: number;
  fallback_reasons: Record<string, number>;
}

export interface NFRScalabilityLevel {
  concurrency: number;
  total_requests: number;
  success_count: number;
  error_count: number;
  error_rate_pct: number;
  throughput_rps: number;
  mean_ms: number;
  p50_ms: number;
  p95_ms: number;
  p99_ms: number;
  max_ms: number;
  duration_ms: number;
}

export interface NFRScalabilityBenchmark {
  mode: string;
  preset?: string;
  executed_at: string;
  target_endpoints: string[];
  levels: NFRScalabilityLevel[];
}

export interface NFRMaintainabilityReport {
  collected_at: string;
  git_commit?: string;
  go_test_passed?: boolean | null;
  go_coverage_pct?: number | null;
  frontend_test_passed?: boolean | null;
  frontend_tsc_passed?: boolean | null;
  evidence_gate_passed?: boolean | null;
  theory_gate_passed?: boolean | null;
  integration_harness_passed?: boolean | null;
  govulncheck_high_critical?: number | null;
  npm_audit_high_critical?: number | null;
  gitleaks_findings?: number | null;
}

export interface NFRReportResponse {
  run_id?: string | null;
  run_label: string;
  environment: string;
  source: string;
  rubric_version: string;
  evaluated_at: string;
  notes?: string;
  summary: NFRSummary;
  runtime: NFRRuntimeSnapshot;
  dimensions: NFRDimensionResult[];
  scalability?: NFRScalabilityBenchmark | null;
  reliability_breakdown: NFRReliabilityBreakdown;
  maintainability_report?: NFRMaintainabilityReport | null;
  thresholds: NFRThresholds;
}

export interface NFRRunListItem {
  run_id: string;
  run_label: string;
  environment: string;
  source: string;
  rubric_version: string;
  overall_status: string;
  passed_count: number;
  warn_count: number;
  fail_count: number;
  not_measured_count: number;
  measured_pass_rate: number;
  non_llm_p95_ms?: number | null;
  availability_pct?: number | null;
  notes?: string;
  created_at: string;
}

export interface NFRRunDetailResponse {
  run: NFRRunListItem;
  report: NFRReportResponse;
}

export interface ScalabilityProbeResponse {
  benchmark: NFRScalabilityBenchmark;
  saved_run?: NFRRunListItem | null;
}

export interface NFRBaselineDelta {
  nonLlmP95DeltaMs: number | null;
  availabilityDeltaPct: number | null;
  passedCountDelta: number;
  measuredPassRateDelta: number;
  peakRpsDelta: number | null;
}

export interface NFRMethodologyPillar {
  id: string;
  citation?: string;
  en: { title: string; formula: string; body: string };
  th: { title: string; formula: string; body: string };
}

export const NFR_METHODOLOGY_PILLARS: readonly NFRMethodologyPillar[] = [
  {
    id: "hybrid_architecture",
    citation: "[R10] ISO/IEC 25010:2023",
    en: {
      title: "1. 3-Layer Hybrid Real Measurement Architecture",
      formula: "Layer 1 (In-Memory Ring) + Layer 2 (Live DB & Security Probes) + Layer 3 (Benchmarks)",
      body: "All 10 ISO/IEC 25010 product quality characteristics are measured directly from live platform state without mocked or hardcoded scores: passive HTTP/LLM telemetry in memory (zero per-request DB write overhead), live PostgreSQL integrity/safety/reproducibility/usability audits, and reproducible concurrency/CI benchmark snapshots persisted in nfr_evaluation_runs.",
    },
    th: {
      title: "1. สถาปัตยกรรมการวัดผลจริง 3 ชั้น (3-Layer Hybrid Measurement)",
      formula: "ชั้นที่ 1 (Ring Buffer) + ชั้นที่ 2 (ตรวจสอบ DB/Security สด) + ชั้นที่ 3 (Benchmark)",
      body: "ประเมินคุณลักษณะคุณภาพซอฟต์แวร์ทั้ง 10 มิติตามมาตรฐาน ISO/IEC 25010 จากข้อมูลจริงทั้งหมดโดยไม่ Mock ค่า: เก็บ Latency ในหน่วยความจำ (ไม่เพิ่มภาระเขียน DB ระหว่างใช้งานจริง), ตรวจสอบความถูกต้อง/ความปลอดภัยของฐานข้อมูลสด และบันทึก Snapshot ประวัติการทดสอบลงตาราง nfr_evaluation_runs",
    },
  },
  {
    id: "r7_percentiles",
    en: {
      title: "2. Route-Normalized R-7 Percentiles & LLM Isolation",
      formula: "h = p × (n − 1);  Q(p) = x_⌊h⌋ + (h − ⌊h⌋) · (x_⌈h⌉ − x_⌊h⌋)",
      body: "Requests are normalized by matched Gin route template inside a bounded 10,000-sample ring buffer. Standard API calls (non_llm_api) are strictly separated from POST /api/simulation/start (llm_simulation_start) so slow external LLM roundtrips never distort CRUD API p50/p95/p99 benchmarks. Percentiles use exact R-7 linear interpolation and return N/A when n = 0.",
    },
    th: {
      title: "2. การคำนวณ R-7 Percentile แบบแยกทราฟฟิก LLM",
      formula: "h = p × (n − 1);  Q(p) = x_⌊h⌋ + (h − ⌊h⌋) · (x_⌈h⌉ − x_⌊h⌋)",
      body: "จัดกลุ่มคำขอตาม Route Template ภายใน Ring Buffer ขนาด 10,000 รายการ โดยแยก API ทั่วไปออกจาก POST /api/simulation/start เพื่อไม่ให้เวลาการเรียก LLM มาบิดเบือนค่า p50/p95/p99 ของระบบหลัก และใช้สูตร R-7 Linear Interpolation ตามมาตรฐานสถิติ (แสดง N/A เมื่อยังไม่มีคำขอ)",
    },
  },
  {
    id: "rubric_rollup",
    en: {
      title: "3. Numeric Thresholds (nfr-rubric-v1) & Honest Pending Rollup",
      formula: "Pass Rate = Passed / (Passed + Warn + Fail) × 100%",
      body: "Pre-declared cutoffs govern every dimension (e.g., Non-LLM p95 <= 300 ms, Availability >= 99.0%, RQ4 system_failure = 0.0%, Pre/Post overlap = 0, Snapshot linkage = 100%). Unmeasured benchmarks (not_measured) and zero-sample constructs (insufficient_data) are tracked separately from failures so overall_status never hides unverified areas.",
    },
    th: {
      title: "3. เกณฑ์ตัวเลข nfr-rubric-v1 และการแยกสถานะรอข้อมูลอย่างโปร่งใส",
      formula: "อัตราผ่านที่วัดแล้ว = Passed / (Passed + Warn + Fail) × 100%",
      body: "กำหนดเกณฑ์ตัวเลขล่วงหน้าครบทุกมิติ (เช่น API p95 <= 300 ms, Availability >= 99.0%, RQ4 system_failure = 0.0%, Pre/Post ซ้ำ = 0) พร้อมแยกมิติที่ยังไม่ได้รัน Benchmark (not_measured) หรือยังไม่ครบตัวอย่าง (insufficient_data) ออกจาก Fail เพื่อไม่ให้สรุปผ่านโดยไม่มีหลักฐาน",
    },
  },
  {
    id: "probe_isolation",
    en: {
      title: "4. Probe Telemetry Isolation, Concurrency Caps & Zero-Secret Rule",
      formula: "Workers <= 20 | Requests <= 200 | Timeout 15s | Mutex 409 | Secret Redaction = 100%",
      body: "In-process read-only Security probes (6 Auth/RBAC/SQLi checks) and Safe In-App Concurrency Probes inject an unspoofable per-process secret header so synthetic audit traffic never contaminates live user latency or error rates. Security posture checks evaluate boolean strength/flags only—raw secrets, DSNs, and keys are never serialized.",
    },
    th: {
      title: "4. การแยกทราฟฟิก Probe เพดานโหลดปลอดภัย และนโยบาย Zero-Secret",
      formula: "Workers <= 20 | Requests <= 200 | Timeout 15s | Mutex 409 | ปกปิด Secret 100%",
      body: "การทดสอบ Security และ Concurrency Probe ภายในระบบจะแนบ Header ลับระดับโปรเซสเพื่อแยกออกจาก Telemetry ของผู้ใช้จริง มีเพดานป้องกันไม่ให้ระบบล่ม พร้อมทั้งตรวจสอบค่าความปลอดภัยแบบ Boolean เท่านั้นโดยไม่เปิดเผยรหัสผ่าน JWT_SECRET, DSN หรือ API Key ออกมาทาง API",
    },
  },
  {
    id: "usability_rq4",
    en: {
      title: "5. RQ4 Fallback Reliability & Issue #13 Usability Bridge",
      formula: "Reliability = HTTP 5xx <= 1% & RQ4 system_failure = 0% & Valid Snapshots = 100%",
      body: "Static fallback activations are tracked transparently with fallback_reasons without penalizing Reliability—fallback guarantees 100% Scenario delivery when LLM providers time out or lack keys. Usability bridges live CSUQ v3 (target >= 5.0/7.0) and Raw NASA-TLX (target <= 55/100) cohort means alongside objective Exam and Simulation completion rates.",
    },
    th: {
      title: "5. เสถียรภาพ RQ4 Fallback และการเชื่อมผล Usability จาก Issue #13",
      formula: "Reliability = HTTP 5xx <= 1% และ RQ4 system_failure = 0% และ Valid Snapshots = 100%",
      body: "ระบบแยกบันทึกการสลับไปใช้ Static Fallback เป็นข้อมูลอธิบายเหตุการณ์โดยไม่หักคะแนนความน่าเชื่อถือ เพราะ Fallback คือกลไกป้องกันระบบล่มเมื่อ LLM ขัดข้อง และเชื่อมค่าเฉลี่ย CSUQ v3 / Raw NASA-TLX จริงจาก Issue #13 ร่วมกับอัตราการทำข้อสอบและ Simulation สำเร็จ",
    },
  },
];

const VERDICT_DISPLAY: Record<string, string> = {
  pass: "Pass",
  warn: "Warn",
  fail: "Fail",
  not_measured: "Not Measured",
  insufficient_data: "Insufficient Data",
};

const DIM_SHORT_NAMES: Record<string, string> = {
  performance: "Performance",
  availability: "Availability",
  scalability: "Scalability",
  reliability: "Reliability (RQ4)",
  security: "Security",
  data_integrity: "Data Integrity",
  reproducibility: "Reproducibility",
  usability: "Usability",
  maintainability: "Maintainability",
  safety: "Safety",
};

function roundDelta2(val: number): number {
  return Math.round(val * 100) / 100;
}

function peakRpsForReport(report: NFRReportResponse): number | null {
  const levels = report.scalability?.levels ?? [];
  if (levels.length === 0) return null;
  let peak = 0;
  for (const lv of levels) {
    if (isFiniteStat(lv.throughput_rps) && lv.throughput_rps > peak) {
      peak = lv.throughput_rps;
    }
  }
  return peak;
}

/** Computes signed deltas comparing current report against a selected baseline snapshot. */
export function computeNfrBaselineDelta(
  current: NFRReportResponse,
  baseline: NFRReportResponse,
): NFRBaselineDelta {
  const curP95 = current.runtime.non_llm_latency.p95;
  const baseP95 = baseline.runtime.non_llm_latency.p95;
  const nonLlmP95DeltaMs =
    isFiniteStat(curP95) && isFiniteStat(baseP95) ? roundDelta2(curP95 - baseP95) : null;

  const curAvail = current.runtime.availability_pct;
  const baseAvail = baseline.runtime.availability_pct;
  const availabilityDeltaPct =
    isFiniteStat(curAvail) && isFiniteStat(baseAvail) ? roundDelta2(curAvail - baseAvail) : null;

  const curPeakRps = peakRpsForReport(current);
  const basePeakRps = peakRpsForReport(baseline);
  const peakRpsDelta =
    isFiniteStat(curPeakRps) && isFiniteStat(basePeakRps) ? roundDelta2(curPeakRps - basePeakRps) : null;

  return {
    nonLlmP95DeltaMs,
    availabilityDeltaPct,
    passedCountDelta: current.summary.passed_count - baseline.summary.passed_count,
    measuredPassRateDelta: roundDelta2(current.summary.measured_pass_rate - baseline.summary.measured_pass_rate),
    peakRpsDelta,
  };
}

function formatLatexExpr(text: string): string {
  return escapeLatex(text)
    .replace(/<=/g, "$\\le$")
    .replace(/>=/g, "$\\ge$");
}

function escapeMarkdownCell(text: string): string {
  return text.replace(/\|/g, "\\|").replace(/\r?\n/g, " ").trim();
}

/** Builds an IEEEtran booktabs LaTeX table bundle for the 10-dimension NFR matrix and scalability curve. */
export function buildNfrLatexTable(report: NFRReportResponse): string {
  const rows = report.dimensions.map((d) => {
    const dimName = escapeLatex(DIM_SHORT_NAMES[d.id] ?? d.id);
    const iso = formatLatexExpr(d.iso_characteristic);
    const target = formatLatexExpr(d.target_expression);
    const measured = formatLatexExpr(d.summary_metric);
    const evidence = escapeLatex(d.evidence_source);
    const verdict = escapeLatex(VERDICT_DISPLAY[d.status] ?? d.status);
    return `    ${dimName} & ${iso} & ${target} & ${measured} & ${evidence} & ${verdict} \\\\`;
  });

  const blocks: string[] = [
    [
      "\\begin{table*}[t]",
      "  \\centering",
      `  \\caption{ISO/IEC 25010 Non-Functional Requirement Evaluation Matrix (${escapeLatex(report.run_label)}, ${escapeLatex(report.rubric_version)})}`,
      "  \\label{tab:nfr_evaluation}",
      "  \\begin{tabular}{llllll}",
      "    \\toprule",
      "    Dimension & ISO/IEC 25010 Characteristic & Target & Measured Result & Evidence Source & Verdict \\\\",
      "    \\midrule",
      ...rows,
      "    \\bottomrule",
      "  \\end{tabular}",
      "\\end{table*}",
    ].join("\n"),
  ];

  const levels = report.scalability?.levels ?? [];
  if (levels.length > 0) {
    const scaleRows = levels.map(
      (lv) =>
        `    ${lv.concurrency} & ${lv.total_requests} & ${lv.throughput_rps.toFixed(1)} & ${lv.mean_ms.toFixed(2)} & ${lv.p50_ms.toFixed(2)} & ${lv.p95_ms.toFixed(2)} & ${lv.p99_ms.toFixed(2)} & ${lv.error_rate_pct.toFixed(2)}\\% \\\\`,
    );
    blocks.push(
      [
        "\\begin{table*}[t]",
        "  \\centering",
        `  \\caption{Concurrent Load Scalability Benchmark (${escapeLatex(report.scalability?.mode ?? "benchmark")})}`,
        "  \\label{tab:nfr_scalability}",
        "  \\begin{tabular}{cccccccc}",
        "    \\toprule",
        "    Workers & Requests & Throughput (RPS) & Mean (ms) & p50 (ms) & p95 (ms) & p99 (ms) & Error Rate \\\\",
        "    \\midrule",
        ...scaleRows,
        "    \\bottomrule",
        "  \\end{tabular}",
        "\\end{table*}",
      ].join("\n"),
    );
  }

  return blocks.join("\n\n");
}

/** Builds clean GitHub-Flavored Markdown tables for the 10-dimension NFR matrix and scalability curve. */
export function buildNfrMarkdownTable(report: NFRReportResponse): string {
  const lines: string[] = [
    `### ISO/IEC 25010 NFR Evaluation Matrix (${report.run_label} — ${report.rubric_version})`,
    "",
    "| Dimension | ISO/IEC 25010 Characteristic | Target | Measured Result | Evidence Source | Verdict |",
    "| :--- | :--- | :--- | :--- | :--- | :--- |",
  ];

  for (const d of report.dimensions) {
    const dimName = escapeMarkdownCell(DIM_SHORT_NAMES[d.id] ?? d.id);
    const verdict = escapeMarkdownCell(VERDICT_DISPLAY[d.status] ?? d.status);
    lines.push(
      `| ${dimName} | ${escapeMarkdownCell(d.iso_characteristic)} | ${escapeMarkdownCell(d.target_expression)} | ${escapeMarkdownCell(d.summary_metric)} | ${escapeMarkdownCell(d.evidence_source)} | ${verdict} |`,
    );
  }

  const levels = report.scalability?.levels ?? [];
  if (levels.length > 0) {
    lines.push(
      "",
      `### Scalability Concurrency Curve (${report.scalability?.mode ?? "benchmark"})`,
      "",
      "| Workers | Requests | Throughput (RPS) | Mean (ms) | p50 (ms) | p95 (ms) | p99 (ms) | Error Rate |",
      "| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |",
    );
    for (const lv of levels) {
      lines.push(
        `| ${lv.concurrency} | ${lv.total_requests} | ${lv.throughput_rps.toFixed(1)} | ${lv.mean_ms.toFixed(2)} | ${lv.p50_ms.toFixed(2)} | ${lv.p95_ms.toFixed(2)} | ${lv.p99_ms.toFixed(2)} | ${lv.error_rate_pct.toFixed(2)}% |`,
      );
    }
  }

  return lines.join("\n");
}

/** Generates an honest, publication-ready academic engineering evaluation paragraph in EN or TH. */
export function buildNfrEvaluationParagraph(report: NFRReportResponse, lang: "en" | "th"): string {
  const s = report.summary;
  const measuredCount = s.passed_count + s.warn_count + s.fail_count;
  const evalDateStr = (report.evaluated_at || "").slice(0, 10) || "N/A";
  const nonLlm = report.runtime.non_llm_latency;
  const availStr = isFiniteStat(report.runtime.availability_pct)
    ? `${report.runtime.availability_pct.toFixed(2)}%`
    : "N/A";
  const p50Str = isFiniteStat(nonLlm.p50) ? `${nonLlm.p50.toFixed(1)} ms` : "N/A";
  const p95Str = isFiniteStat(nonLlm.p95) ? `${nonLlm.p95.toFixed(1)} ms` : "N/A";
  const p99Str = isFiniteStat(nonLlm.p99) ? `${nonLlm.p99.toFixed(1)} ms` : "N/A";

  const rb = report.reliability_breakdown;
  const snapRate =
    rb.total_snapshot_count > 0
      ? `${((rb.valid_snapshot_count / rb.total_snapshot_count) * 100).toFixed(1)}%`
      : "N/A";

  const secDim = report.dimensions.find((d) => d.id === "security");
  const safeDim = report.dimensions.find((d) => d.id === "safety");
  const secVerdict = VERDICT_DISPLAY[secDim?.status ?? "not_measured"] ?? secDim?.status ?? "Not Measured";
  const safeVerdict = VERDICT_DISPLAY[safeDim?.status ?? "not_measured"] ?? safeDim?.status ?? "Not Measured";
  const secSummary = secDim?.summary_metric || "N/A";
  const safeSummary = safeDim?.summary_metric || "N/A";

  const pendingNames = report.dimensions
    .filter((d) => d.status === "not_measured" || d.status === "insufficient_data")
    .map((d) => DIM_SHORT_NAMES[d.id] ?? d.id);

  const levels = report.scalability?.levels ?? [];
  const maxLevel = levels.length > 0 ? levels[levels.length - 1] : null;

  if (lang === "th") {
    const parts: string[] = [
      `การประเมินคุณสมบัติด้านความต้องการที่ไม่ใช่เชิงฟังก์ชัน (NFR) ตามกรอบมาตรฐาน ISO/IEC 25010 [R10] ภายใต้เกณฑ์ ${report.rubric_version} (ชุดข้อมูล "${report.run_label}" ในสภาพแวดล้อม ${report.environment} วันที่ ${evalDateStr})`,
      `พบว่าจากทั้งหมด ${s.total_dimensions} มิติ มีผลการประเมินที่วัดผลได้แล้ว ${measuredCount} มิติ ผ่านเกณฑ์ ${s.passed_count} มิติ เตือน ${s.warn_count} มิติ และไม่ผ่าน ${s.fail_count} มิติ (อัตราผ่านของมิติที่วัดแล้ว ${s.measured_pass_rate.toFixed(1)}%)`,
      `ด้านประสิทธิภาพ API หลักมีค่าความหน่วง p50 = ${p50Str}, p95 = ${p95Str}, p99 = ${p99Str} (จาก ${nonLlm.count} คำขอ, DB ping ${report.runtime.db_latency_ms.toFixed(2)} ms) และความพร้อมใช้งาน Non-5xx อยู่ที่ ${availStr}`,
      `ด้านความน่าเชื่อถือ RQ4 มีอัตราสถานการณ์ล้มเหลวระดับระบบ (system_failure) เท่ากับ ${rb.system_failure_count}/${rb.total_sessions} และส่งมอบ Scenario Snapshot ที่ผ่านการตรวจสอบแล้ว ${snapRate}`,
      `ด้านความปลอดภัย Security อยู่ในสถานะ ${secVerdict} (${secSummary}) และขอบเขตความปลอดภัยกรอบจำลอง Safety อยู่ในสถานะ ${safeVerdict} (${safeSummary})`,
    ];
    if (maxLevel) {
      parts.push(
        `ในการทดสอบโหลดพร้อมกันสูงสุด ${maxLevel.concurrency} เวิร์กเกอร์ ระบบรองรับอัตราการส่งผ่าน ${maxLevel.throughput_rps.toFixed(1)} RPS ที่ p95 = ${maxLevel.p95_ms.toFixed(1)} ms และอัตราความผิดพลาด ${maxLevel.error_rate_pct.toFixed(2)}%`,
      );
    }
    if (pendingNames.length > 0) {
      parts.push(
        `ทั้งนี้ ยังมีอีก ${pendingNames.length} มิติที่อยู่ในสถานะรอข้อมูลหรือยังไม่ได้วัดผลเพิ่มเติม ได้แก่ ${pendingNames.join(", ")} จึงรายงานสถานะรวมเป็น ${s.overall_status} อย่างตรงไปตรงมา`,
      );
    }
    return parts.join(" ");
  }

  const parts: string[] = [
    `System-level non-functional evaluation under the ISO/IEC 25010 quality model [R10] (${report.rubric_version}; "${report.run_label}" in ${report.environment}, evaluated ${evalDateStr})`,
    `completed across ${measuredCount} of ${s.total_dimensions} dimensions, yielding ${s.passed_count} pass, ${s.warn_count} warning, and ${s.fail_count} fail outcomes (${s.measured_pass_rate.toFixed(1)}% measured pass rate).`,
    `Runtime telemetry across ${nonLlm.count} non-LLM API requests recorded p50 = ${p50Str}, p95 = ${p95Str}, and p99 = ${p99Str} (PostgreSQL ping ${report.runtime.db_latency_ms.toFixed(2)} ms) with ${availStr} Non-5xx availability.`,
    `For RQ4 scenario reliability, system_failure sessions remained ${rb.system_failure_count}/${rb.total_sessions} with ${snapRate} valid scenario snapshot delivery (${rb.ai_generated_count} AI-generated, ${rb.static_fallback_count} static fallback).`,
    `Security verification achieved ${secVerdict} (${secSummary}), and Safety boundary audits recorded ${safeVerdict} (${safeSummary}).`,
  ];
  if (maxLevel) {
    parts.push(
      `Under concurrency benchmarking at ${maxLevel.concurrency} workers, the service sustained ${maxLevel.throughput_rps.toFixed(1)} RPS with p95 = ${maxLevel.p95_ms.toFixed(1)} ms and a ${maxLevel.error_rate_pct.toFixed(2)}% error rate.`,
    );
  }
  if (pendingNames.length > 0) {
    parts.push(
      `${pendingNames.length} dimension(s) remain awaiting data/pending (${pendingNames.join(", ")}), so the overall verdict is reported conservatively as ${s.overall_status}.`,
    );
  }
  return parts.join(" ");
}
