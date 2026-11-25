export interface DescriptiveStat {
  key: string;
  group?: string;
  n: number;
  mean?: number | null;
  sd?: number | null;
  se?: number | null;
  ci_low?: number | null;
  ci_high?: number | null;
  median?: number | null;
  q1?: number | null;
  q3?: number | null;
  iqr?: number | null;
  min?: number | null;
  max?: number | null;
  band: string;
  status: string;
}

export interface CronbachAlphaStat {
  key: string;
  item_count: number;
  complete_cases_n: number;
  excluded_na_cases: number;
  alpha?: number | null;
  band: string;
  status: string;
}

export interface PairedTTestStat {
  n: number;
  df: number;
  mean_diff?: number | null;
  sd_diff?: number | null;
  ci_low?: number | null;
  ci_high?: number | null;
  t_stat?: number | null;
  p_value?: number | null;
  cohens_dz?: number | null;
  effect_band: string;
  status: string;
}

export interface WilcoxonTestStat {
  n: number;
  effective_n: number;
  zero_diff_count: number;
  method: string;
  w_stat?: number | null;
  w_plus?: number | null;
  w_minus?: number | null;
  p_value?: number | null;
  rank_biserial?: number | null;
  status: string;
}

export interface SpearmanCorrelationStat {
  key: string;
  n: number;
  rho?: number | null;
  p_value?: number | null;
  band: string;
  status: string;
}

export interface AdminPaperStatisticsResponse {
  generated_at: string;
  protocol_version: string;
  csuq_schema_version: string;
  nasa_tlx_schema_version: string;
  is_pilot_sample: boolean;
  cohort: {
    total_learners: number;
    completed_pre_test_learners: number;
    completed_post_test_learners: number;
    questionnaire_respondents: number;
    response_rate_pct: number;
  };
  csuq_subscales: DescriptiveStat[];
  csuq_items: DescriptiveStat[];
  nasa_tlx: DescriptiveStat[];
  reliability: CronbachAlphaStat[];
  learning_gain: {
    pre: DescriptiveStat;
    post: DescriptiveStat;
    difference: DescriptiveStat;
    paired_t_test: PairedTTestStat;
    wilcoxon: WilcoxonTestStat;
  };
  correlations: SpearmanCorrelationStat[];
}

export interface MethodologyPillar {
  id: string;
  citation?: string;
  en: {
    title: string;
    formula: string;
    body: string;
  };
  th: {
    title: string;
    formula: string;
    body: string;
  };
}

export const METHODOLOGY_PILLARS: readonly MethodologyPillar[] = [
  {
    id: "csuq_v3",
    citation: "[R8] Lewis (1995, 2018)",
    en: {
      title: "1. CSUQ v3 Usability Scoring & N/A Omission",
      formula: "Subscale Mean = (1 / m) × Σ x_i  (over non-N/A items)",
      body: "System usability is measured via the 16-item Computer System Usability Questionnaire Version 3 on an ascending 1–7 Likert scale (higher = better) across SYSUSE (Q1–Q6), INFOQUAL (Q7–Q12), INTERQUAL (Q13–Q15), and OVERALL (Q1–Q16). Items marked N/A are excluded from respondent-level denominators per Lewis (1995, 2018), while all-16 N/A submissions are disallowed.",
    },
    th: {
      title: "1. การให้คะแนน CSUQ v3 และการตัดข้อ N/A",
      formula: "ค่าเฉลี่ยด้าน = (1 / m) × Σ x_i  (คำนวณเฉพาะข้อที่ให้คะแนน 1–7)",
      body: "วัดความพึงพอใจและความง่ายในการใช้งานด้วยแบบสอบถามมาตรฐาน CSUQ Version 3 จำนวน 16 ข้อ สเกลขาขึ้น 1–7 (คะแนนสูง = ดี) ครอบคลุม SYSUSE (Q1–Q6), INFOQUAL (Q7–Q12), INTERQUAL (Q13–Q15) และ OVERALL (Q1–Q16) โดยตัดข้อที่ตอบ N/A ออกจากตัวหารตามแนวทาง Lewis (1995, 2018) และห้ามตอบ N/A ทั้ง 16 ข้อ",
    },
  },
  {
    id: "raw_nasa_tlx",
    citation: "[R9] Hart & Staveland (1988); Hart (2006)",
    en: {
      title: "2. Raw NASA-TLX & Performance Polarity Normalization",
      formula: "Performance_workload = 100 − Performance_rating;  RTLX = (1 / 6) × Σ d_k",
      body: "Cognitive and operational task load is assessed via unweighted Raw NASA-TLX (RTLX) on a 0–100 scale (step 5) across Mental, Physical, Temporal, Performance, Effort, and Frustration. To prevent respondent polarity errors, the UI anchors Performance from 0 (Failure) to 100 (Perfect), while the backend stores both raw self-rating and inverted workload (100 − rating) before averaging all 6 workload dimensions.",
    },
    th: {
      title: "2. Raw NASA-TLX และการกลับขั้วคะแนน Performance",
      formula: "ภาระงาน Performance = 100 − คะแนนประเมินตนเอง;  RTLX = (1 / 6) × Σ d_k",
      body: "ประเมินภาระงานทางความคิดด้วย Raw NASA-TLX แบบไม่ถ่วงน้ำหนัก (สเกล 0–100 เพิ่มทีละ 5) เพื่อไม่ให้ผู้เรียนสับสนทิศทาง หน้าเว็บจึงวางแกน Performance จาก 0 (ไม่สำเร็จ) ไป 100 (ดีเยี่ยม) แล้วฝั่ง Backend จะคำนวณค่าภาระงานกลับด้าน (100 − rating) เก็บควบคู่กับค่าเดิมก่อนหาค่าเฉลี่ย 6 มิติ",
    },
  },
  {
    id: "descriptives_ci",
    en: {
      title: "3. Descriptive Statistics & Student-t 95% Confidence Intervals",
      formula: "s = √[ Σ(x_i − x̄)² / (n − 1) ];  CI_95% = x̄ ± t_{0.975, n−1} · (s / √n)",
      body: "Every scale reports sample size (n), arithmetic mean, Bessel-corrected sample standard deviation (n − 1), exact two-tailed Student-t 95% CI (evaluated via regularized incomplete beta function I_x(a,b)), median, R-7 linear-interpolated IQR (Q3 − Q1), and min–max range.",
    },
    th: {
      title: "3. สถิติเชิงพรรณนาและช่วงความเชื่อมั่น 95% (Student-t CI)",
      formula: "s = √[ Σ(x_i − x̄)² / (n − 1) ];  CI_95% = x̄ ± t_{0.975, n−1} · (s / √n)",
      body: "รายงานครบทั้ง n, ค่าเฉลี่ย, ส่วนเบี่ยงเบนมาตรฐานตัวอย่าง (ตัวหาร n − 1), ช่วงความเชื่อมั่น 95% แบบ Student-t (คำนวณตรงจากฟังก์ชัน Regularized Incomplete Beta), ค่ามัธยฐาน, พิสัยควอไทล์ R-7 (IQR = Q3 − Q1) และค่าต่ำสุด–สูงสุด",
    },
  },
  {
    id: "cronbach_alpha",
    en: {
      title: "4. Internal Consistency Reliability (Cronbach's α)",
      formula: "α = [k / (k − 1)] × (1 − Σ s_j² / s_total²)",
      body: "Internal consistency is computed separately for CSUQ Overall (k = 16), its subscales, and NASA-TLX (k = 6) over listwise complete cases, transparently disclosing excluded N/A cases and interpreting α against George & Mallery (2003) bands (≥0.90 Excellent, ≥0.80 Good, ≥0.70 Acceptable).",
    },
    th: {
      title: "4. ความเที่ยงตรงภายในด้วยสัมประสิทธิ์ Cronbach's α",
      formula: "α = [k / (k − 1)] × (1 − Σ s_j² / s_total²)",
      body: "คำนวณความสอดคล้องภายในของชุดข้อคำถาม CSUQ รวม (k = 16), รายด้านย่อย และ NASA-TLX (k = 6) โดยใช้เฉพาะชุดข้อมูลที่ตอบตัวเลขครบทุกข้อในสเกลนั้น (Listwise complete cases) พร้อมรายงานจำนวนเคสที่มี N/A ที่ถูกแยกออกอย่างโปร่งใส",
    },
  },
  {
    id: "paired_inference",
    en: {
      title: "5. Pre→Post Hypothesis Testing & Effect Sizes",
      formula: "t = d̄ / (s_d / √n),  d_z = d̄ / s_d;   r_rb = (W⁺ − W⁻) / (W⁺ + W⁻)",
      body: "Paired Pre-to-Post learning gains (d_i = Post_i − Pre_i) are evaluated with both the parametric Paired t-test (with Cohen's d_z and 95% CI of Δ) and the non-parametric Wilcoxon signed-rank test (exact dynamic-programming distribution for tie-free n_eff ≤ 20, or continuity- and tie-corrected normal approximation otherwise, paired with rank-biserial correlation r_rb).",
    },
    th: {
      title: "5. การทดสอบสมมติฐาน Pre→Post และขนาดอิทธิพล (Effect Size)",
      formula: "t = d̄ / (s_d / √n),  d_z = d̄ / s_d;   r_rb = (W⁺ − W⁻) / (W⁺ + W⁻)",
      body: "วิเคราะห์ผลต่างคะแนนรายบุคคล (d_i = Post_i − Pre_i) ด้วยทั้งวิธีพาราเมตริก (Paired t-test พร้อม Cohen's d_z และ 95% CI) และนอนพาราเมตริก (Wilcoxon Signed-Rank แบบ Exact DP เมื่อ n_eff ≤ 20 และไม่มีค่าซ้ำ หรือ Normal Approximation แบบแก้ค่าซ้ำและต่อเนื่อง พร้อมค่า r_rb)",
    },
  },
  {
    id: "spearman_rho",
    en: {
      title: "6. Spearman's Rank Correlation (ρ)",
      formula: "ρ = Pearson_r(Rank(X), Rank(Y));   t = ρ √[(n − 2) / (1 − ρ²)]",
      body: "Associations between Learning Gain (Δ), CSUQ Overall, and Raw NASA-TLX Workload are quantified using tie-safe average-rank Spearman's ρ with two-tailed Student-t p-values (df = n − 2), avoiding linearity and normality assumptions on bounded ordinal ratings.",
    },
    th: {
      title: "6. สหสัมพันธ์แบบอันดับ Spearman (ρ)",
      formula: "ρ = Pearson_r(Rank(X), Rank(Y));   t = ρ √[(n − 2) / (1 − ρ²)]",
      body: "วัดความสัมพันธ์ระหว่างพัฒนาการคะแนน (Δ), คะแนนความพึงพอใจ CSUQ และภาระงาน NASA-TLX ด้วยสหสัมพันธ์เชิงอันดับ Spearman (จัดการอันดับเฉลี่ยเมื่อมีค่าซ้ำ) เพื่อไม่ให้ยึดติดข้อสมมติฐานการแจกแจงปกติบนสเกลแบบลำดับ",
    },
  },
  {
    id: "small_n_guardrails",
    en: {
      title: "7. Small-Sample Guardrails & Pilot Caveats",
      formula: "SD/CI/α: n ≥ 2  |  t-test/d_z/ρ: n ≥ 3  |  Wilcoxon: n_eff ≥ 5  |  Pilot flag: n < 30",
      body: "Statistical outputs are gated strictly by minimum valid sample thresholds and non-zero variance checks. Whenever respondent count n < 30, a Pilot Sample Caveat flags results as exploratory and reminds authors to report participant attrition alongside outcomes.",
    },
    th: {
      title: "7. เกณฑ์ป้องกันกลุ่มตัวอย่างน้อยและข้อจำกัด Pilot Study",
      formula: "SD/CI/α: n ≥ 2  |  t-test/d_z/ρ: n ≥ 3  |  Wilcoxon: n_eff ≥ 5  |  Pilot: n < 30",
      body: "ระบบบังคับเงื่อนไขจำนวนตัวอย่างขั้นต่ำและความแปรปรวนไม่เป็นศูนย์ก่อนแสดงผลทดสอบสถิติทุกตัว และหากจำนวนผู้ตอบ n < 30 จะติดป้าย Pilot Sample Caveat เพื่อกำชับให้รายงานผลเป็นข้อมูลนำร่องและแนบอัตราการถอนตัว (Attrition)",
    },
  },
] as const;

const CSUQ_LABELS: Record<string, string> = {
  sysuse: "System Usefulness (SYSUSE, Q1–Q6)",
  infoqual: "Information Quality (INFOQUAL, Q7–Q12)",
  interqual: "Interface Quality (INTERQUAL, Q13–Q15)",
  overall: "Overall Usability (OVERALL, Q1–Q16)",
};

const TLX_LABELS: Record<string, string> = {
  mental_demand: "Mental Demand",
  physical_demand: "Physical Demand",
  temporal_demand: "Temporal Demand",
  performance_rating: "Performance (Self-Rating)",
  performance_workload: "Performance (Workload = 100 - Rating)",
  effort: "Effort",
  frustration: "Frustration",
  overall_workload: "Overall Raw NASA-TLX",
};

const CORR_LABELS: Record<string, string> = {
  learning_gain_vs_csuq_overall: "Learning Gain (Delta) vs. CSUQ Overall",
  learning_gain_vs_nasa_tlx_overall: "Learning Gain (Delta) vs. NASA-TLX Workload",
  csuq_overall_vs_nasa_tlx_overall: "CSUQ Overall vs. NASA-TLX Workload",
};

/** Type guard confirming a nullable/optional numeric stat field is safe to format. */
export function isFiniteStat(val: number | null | undefined): val is number {
  return typeof val === "number" && Number.isFinite(val);
}

export function escapeLatex(text: string): string {
  return text
    .replace(/\\/g, "\\textbackslash{}")
    .replace(/([%&#_{}$])/g, "\\$1");
}

function fmtOpt(val: number | null | undefined, decimals: number, naToken: string): string {
  return isFiniteStat(val) ? val.toFixed(decimals) : naToken;
}

function fmtCIWithFallback(stat: DescriptiveStat, naToken: string): string {
  if (!isFiniteStat(stat.ci_low) || !isFiniteStat(stat.ci_high)) return naToken;
  return `[${stat.ci_low.toFixed(2)}, ${stat.ci_high.toFixed(2)}]`;
}

function fmtMedIQRWithFallback(stat: DescriptiveStat, naToken: string): string {
  if (!isFiniteStat(stat.median)) return naToken;
  if (!isFiniteStat(stat.iqr)) return stat.median.toFixed(2);
  return `${stat.median.toFixed(2)} (${stat.iqr.toFixed(2)})`;
}

function fmtPWithFallback(val: number | null | undefined, naToken: string, smallToken: string): string {
  if (!isFiniteStat(val)) return naToken;
  if (val < 0.0001) return smallToken;
  return val.toFixed(4);
}

const BAND_DISPLAY: Record<string, string> = {
  excellent: "Excellent",
  good: "Good",
  very_high: "Very High",
  high: "High",
  above_average: "Above Average",
  moderate: "Moderate",
  neutral: "Neutral",
  acceptable: "Acceptable",
  questionable: "Questionable",
  low: "Low",
  poor: "Poor",
  unacceptable: "Unacceptable",
  small: "Small",
  medium: "Medium",
  large: "Large",
  negligible: "Negligible",
  weak: "Weak",
  strong: "Strong",
  zero_variance: "Zero variance",
  insufficient_n: "N/A",
};

function formatBandDisplay(band: string): string {
  return BAND_DISPLAY[band] ?? band;
}

function formatGateText(status: string, minRequired: number): string {
  if (status === "zero_variance") return "Zero variance";
  return `N/A (requires n >= ${minRequired})`;
}

function fmtRangeWithFallback(stat: DescriptiveStat, naToken: string): string {
  if (!isFiniteStat(stat.min) || !isFiniteStat(stat.max)) return naToken;
  return `${stat.min.toFixed(2)}--${stat.max.toFixed(2)}`;
}

const fmtLatexOpt = (val: number | null | undefined, decimals = 2) => fmtOpt(val, decimals, "\\text{N/A}");
const fmtLatexCI = (stat: DescriptiveStat) => fmtCIWithFallback(stat, "\\text{N/A ($n \\ge 2$)}");
const fmtLatexMedIQR = (stat: DescriptiveStat) => fmtMedIQRWithFallback(stat, "\\text{N/A}");
const fmtLatexRange = (stat: DescriptiveStat) => fmtRangeWithFallback(stat, "\\text{N/A}");
const fmtLatexP = (val: number | null | undefined) => fmtPWithFallback(val, "\\text{N/A}", "$< 0.0001$");
const fmtLatexGate = (status: string, minRequired: number) =>
  status === "zero_variance" ? "\\text{Zero variance}" : `\\text{N/A (requires } n \\ge ${minRequired})`;

const fmtMdOpt = (val: number | null | undefined, decimals = 2) => fmtOpt(val, decimals, "N/A");
const fmtMdCI = (stat: DescriptiveStat) => fmtCIWithFallback(stat, "N/A (requires n >= 2)");
const fmtMdMedIQR = (stat: DescriptiveStat) => fmtMedIQRWithFallback(stat, "N/A");
const fmtMdRange = (stat: DescriptiveStat) => {
  if (!isFiniteStat(stat.min) || !isFiniteStat(stat.max)) return "N/A";
  return `${stat.min.toFixed(2)}–${stat.max.toFixed(2)}`;
};
const fmtMdP = (val: number | null | undefined) => fmtPWithFallback(val, "N/A", "< 0.0001");

export function buildLatexTable1(stats: AdminPaperStatisticsResponse): string {
  const alphaMap = new Map(stats.reliability.map((r) => [r.key, r]));
  const rows = stats.csuq_subscales.map((s) => {
    const label = escapeLatex(CSUQ_LABELS[s.key] ?? s.key);
    const alphaEntry = alphaMap.get(`csuq_${s.key}`);
    const sdStr = isFiniteStat(s.sd) ? fmtLatexOpt(s.sd) : fmtLatexGate(s.status, 2);
    const alphaStr = isFiniteStat(alphaEntry?.alpha)
      ? fmtLatexOpt(alphaEntry.alpha, 3)
      : fmtLatexGate(alphaEntry?.status ?? "insufficient_n", 2);
    const bandStr = escapeLatex(formatBandDisplay(s.band));
    return `    ${label} & ${s.n} & ${fmtLatexOpt(s.mean)} & ${sdStr} & ${fmtLatexCI(s)} & ${fmtLatexMedIQR(s)} & ${fmtLatexRange(s)} & ${alphaStr} & ${bandStr} \\\\`;
  });

  return [
    "\\begin{table*}[t]",
    "  \\centering",
    `  \\caption{CSUQ Version 3 System Usability and Internal Consistency ($N = ${stats.cohort.questionnaire_respondents}$)}`,
    "  \\label{tab:csuq_usability}",
    "  \\begin{tabular}{lcccccccc}",
    "    \\toprule",
    "    Construct & $n$ & $M$ & $SD$ & 95\\% CI & Median (IQR) & Min--Max & $\\alpha$ & Band \\\\",
    "    \\midrule",
    ...rows,
    "    \\bottomrule",
    "  \\end{tabular}",
    "\\end{table*}",
  ].join("\n");
}

export function buildLatexTable1Items(stats: AdminPaperStatisticsResponse): string {
  const rows = stats.csuq_items.map((item) => {
    const label = escapeLatex(`${item.key.toUpperCase()} (${item.group ?? "CSUQ"})`);
    const sdStr = isFiniteStat(item.sd) ? fmtLatexOpt(item.sd) : fmtLatexGate(item.status, 2);
    const bandStr = escapeLatex(formatBandDisplay(item.band));
    return `    ${label} & ${item.n} & ${fmtLatexOpt(item.mean)} & ${sdStr} & ${fmtLatexCI(item)} & ${fmtLatexMedIQR(item)} & ${fmtLatexRange(item)} & ${bandStr} \\\\`;
  });

  return [
    "\\begin{table*}[t]",
    "  \\centering",
    `  \\caption{Item-Level CSUQ Version 3 Descriptive Statistics ($N = ${stats.cohort.questionnaire_respondents}$)}`,
    "  \\label{tab:csuq_items}",
    "  \\begin{tabular}{lccccccc}",
    "    \\toprule",
    "    Item & $n$ & $M$ & $SD$ & 95\\% CI & Median (IQR) & Min--Max & Band \\\\",
    "    \\midrule",
    ...rows,
    "    \\bottomrule",
    "  \\end{tabular}",
    "\\end{table*}",
  ].join("\n");
}

export function buildLatexTable2(stats: AdminPaperStatisticsResponse): string {
  const tlxAlpha = stats.reliability.find((r) => r.key === "nasa_tlx_overall");
  const rows = stats.nasa_tlx.map((s) => {
    const label = escapeLatex(TLX_LABELS[s.key] ?? s.key);
    const sdStr = isFiniteStat(s.sd) ? fmtLatexOpt(s.sd) : fmtLatexGate(s.status, 2);
    const bandStr = escapeLatex(formatBandDisplay(s.band));
    return `    ${label} & ${s.n} & ${fmtLatexOpt(s.mean)} & ${sdStr} & ${fmtLatexCI(s)} & ${fmtLatexMedIQR(s)} & ${fmtLatexRange(s)} & ${bandStr} \\\\`;
  });

  const tlxAlphaText = isFiniteStat(tlxAlpha?.alpha)
    ? fmtLatexOpt(tlxAlpha.alpha, 3)
    : fmtLatexGate(tlxAlpha?.status ?? "insufficient_n", 2);

  return [
    "\\begin{table*}[t]",
    "  \\centering",
    `  \\caption{Raw NASA-TLX Cognitive Workload Profile ($N = ${stats.cohort.questionnaire_respondents}$, Cronbach's $\\alpha = ${tlxAlphaText}$)}`,
    "  \\label{tab:nasa_tlx}",
    "  \\begin{tabular}{lccccccc}",
    "    \\toprule",
    "    Dimension & $n$ & $M$ & $SD$ & 95\\% CI & Median (IQR) & Min--Max & Band \\\\",
    "    \\midrule",
    ...rows,
    "    \\bottomrule",
    "  \\end{tabular}",
    "\\end{table*}",
  ].join("\n");
}

export function buildLatexTable3(stats: AdminPaperStatisticsResponse): string {
  const lg = stats.learning_gain;
  const t = lg.paired_t_test;
  const w = lg.wilcoxon;
  const examRows = [
    ["Pre-Test (\\%)", lg.pre],
    ["Post-Test (\\%)", lg.post],
    ["Learning Gain ($\\Delta$)", lg.difference],
  ] as const;

  const partA = examRows.map(([lbl, stat]) => {
    const sdStr = isFiniteStat(stat.sd) ? fmtLatexOpt(stat.sd) : fmtLatexGate(stat.status, 2);
    return `    ${lbl} & ${stat.n} & ${fmtLatexOpt(stat.mean)} & ${sdStr} & ${fmtLatexCI(stat)} & ${fmtLatexMedIQR(stat)} & ${fmtLatexRange(stat)} \\\\`;
  });

  const corrRows = stats.correlations.map((c) => {
    const label = escapeLatex(CORR_LABELS[c.key] ?? c.key);
    const rhoStr = c.status === "ok" ? fmtLatexOpt(c.rho, 3) : fmtLatexGate(c.status, 3);
    const pStr = c.status === "ok" ? fmtLatexP(c.p_value) : "\\text{---}";
    const strengthStr = c.status === "ok" ? escapeLatex(formatBandDisplay(c.band)) : fmtLatexGate(c.status, 3);
    return `    ${label} & ${c.n} & ${rhoStr} & ${pStr} & ${strengthStr} & & \\\\`;
  });

  const pairedSummary =
    t.status === "ok"
      ? `Paired $t$-test: $t(${t.df}) = ${fmtLatexOpt(t.t_stat, 3)}$, $p = ${fmtLatexP(t.p_value)}$, Cohen's $d_z = ${fmtLatexOpt(t.cohens_dz, 3)}$`
      : `Paired $t$-test: $${fmtLatexGate(t.status, 3)}$`;

  const wilcoxonSummary =
    w.status === "ok"
      ? `Wilcoxon signed-rank: $W = ${fmtLatexOpt(w.w_stat, 1)}$ ($n_{\\text{eff}} = ${w.effective_n}$), $p = ${fmtLatexP(w.p_value)}$, $r_{rb} = ${fmtLatexOpt(w.rank_biserial, 3)}$`
      : `Wilcoxon signed-rank: $${fmtLatexGate(w.status, 5)}$`;

  return [
    "\\begin{table*}[t]",
    "  \\centering",
    `  \\caption{Pre-to-Post Learning Gain Inference and Spearman Correlations ($N = ${stats.cohort.questionnaire_respondents}$)}`,
    "  \\label{tab:learning_gain_inference}",
    "  \\begin{tabular}{lcccccc}",
    "    \\toprule",
    "    Assessment Metric & $n$ & $M$ & $SD$ & 95\\% CI & Median (IQR) & Min--Max \\\\",
    "    \\midrule",
    ...partA,
    "    \\midrule",
    `    \\multicolumn{7}{l}{${pairedSummary}} \\\\`,
    `    \\multicolumn{7}{l}{${wilcoxonSummary}} \\\\`,
    "    \\midrule",
    "    Spearman Pair & $n$ & $\\rho$ & $p$ (2-tailed) & Strength & & \\\\",
    "    \\midrule",
    ...corrRows,
    "    \\bottomrule",
    "  \\end{tabular}",
    "\\end{table*}",
  ].join("\n");
}

export function buildAllLatexTables(stats: AdminPaperStatisticsResponse): string {
  return [
    "% Auto-generated by CyberSim/SOC Trainer Paper Statistics (IEEEtran + booktabs)",
    buildLatexTable1(stats),
    "",
    buildLatexTable1Items(stats),
    "",
    buildLatexTable2(stats),
    "",
    buildLatexTable3(stats),
  ].join("\n");
}

export function buildMarkdownTable1(stats: AdminPaperStatisticsResponse): string {
  const alphaMap = new Map(stats.reliability.map((r) => [r.key, r]));
  const lines = [
    "| Construct | n | M | SD | 95% CI | Median (IQR) | Min–Max | Cronbach's α | Band |",
    "| --- | --- | --- | --- | --- | --- | --- | --- | --- |",
  ];
  for (const s of stats.csuq_subscales) {
    const a = alphaMap.get(`csuq_${s.key}`);
    const sdStr = isFiniteStat(s.sd) ? fmtMdOpt(s.sd) : formatGateText(s.status, 2);
    const alphaStr = isFiniteStat(a?.alpha) ? fmtMdOpt(a.alpha, 3) : formatGateText(a?.status ?? "insufficient_n", 2);
    lines.push(
      `| ${CSUQ_LABELS[s.key] ?? s.key} | ${s.n} | ${fmtMdOpt(s.mean)} | ${sdStr} | ${fmtMdCI(s)} | ${fmtMdMedIQR(s)} | ${fmtMdRange(s)} | ${alphaStr} | ${formatBandDisplay(s.band)} |`,
    );
  }
  return lines.join("\n");
}

export function buildMarkdownTable1Items(stats: AdminPaperStatisticsResponse): string {
  const lines = [
    "| Item | Subscale | n | M | SD | 95% CI | Median (IQR) | Min–Max | Band |",
    "| --- | --- | --- | --- | --- | --- | --- | --- | --- |",
  ];
  for (const item of stats.csuq_items) {
    const sdStr = isFiniteStat(item.sd) ? fmtMdOpt(item.sd) : formatGateText(item.status, 2);
    lines.push(
      `| ${item.key.toUpperCase()} | ${item.group ?? "CSUQ"} | ${item.n} | ${fmtMdOpt(item.mean)} | ${sdStr} | ${fmtMdCI(item)} | ${fmtMdMedIQR(item)} | ${fmtMdRange(item)} | ${formatBandDisplay(item.band)} |`,
    );
  }
  return lines.join("\n");
}

export function buildMarkdownTable2(stats: AdminPaperStatisticsResponse): string {
  const lines = [
    "| Dimension | n | M | SD | 95% CI | Median (IQR) | Min–Max | Band |",
    "| --- | --- | --- | --- | --- | --- | --- | --- |",
  ];
  for (const s of stats.nasa_tlx) {
    const sdStr = isFiniteStat(s.sd) ? fmtMdOpt(s.sd) : formatGateText(s.status, 2);
    lines.push(
      `| ${TLX_LABELS[s.key] ?? s.key} | ${s.n} | ${fmtMdOpt(s.mean)} | ${sdStr} | ${fmtMdCI(s)} | ${fmtMdMedIQR(s)} | ${fmtMdRange(s)} | ${formatBandDisplay(s.band)} |`,
    );
  }
  return lines.join("\n");
}

export function buildMarkdownTable3(stats: AdminPaperStatisticsResponse): string {
  const lg = stats.learning_gain;
  const t = lg.paired_t_test;
  const w = lg.wilcoxon;
  const descLine = (label: string, s: DescriptiveStat) => {
    const sdStr = isFiniteStat(s.sd) ? fmtMdOpt(s.sd) : formatGateText(s.status, 2);
    return `| ${label} | ${s.n} | ${fmtMdOpt(s.mean)} | ${sdStr} | ${fmtMdCI(s)} | ${fmtMdMedIQR(s)} | ${fmtMdRange(s)} |`;
  };
  const lines = [
    "| Metric | n | M | SD | 95% CI | Median (IQR) | Min–Max |",
    "| --- | --- | --- | --- | --- | --- | --- |",
    descLine("Pre-Test (%)", lg.pre),
    descLine("Post-Test (%)", lg.post),
    descLine("Learning Gain (Δ)", lg.difference),
    "",
    t.status === "ok"
      ? `- **Paired t-test:** t(${t.df}) = ${fmtMdOpt(t.t_stat, 3)}, p = ${fmtMdP(t.p_value)}, Cohen's d_z = ${fmtMdOpt(t.cohens_dz, 3)}`
      : `- **Paired t-test:** ${formatGateText(t.status, 3)}`,
    w.status === "ok"
      ? `- **Wilcoxon signed-rank:** W = ${fmtMdOpt(w.w_stat, 1)} (n_eff = ${w.effective_n}), p = ${fmtMdP(w.p_value)}, r_rb = ${fmtMdOpt(w.rank_biserial, 3)}`
      : `- **Wilcoxon signed-rank:** ${formatGateText(w.status, 5)}`,
    "",
    "| Construct Pair | n | Spearman ρ | p (2-tailed) | Strength |",
    "| --- | --- | --- | --- | --- |",
  ];
  for (const c of stats.correlations) {
    const rhoStr = c.status === "ok" ? fmtMdOpt(c.rho, 3) : formatGateText(c.status, 3);
    const pStr = c.status === "ok" ? fmtMdP(c.p_value) : "—";
    const bandStr = c.status === "ok" ? formatBandDisplay(c.band) : formatGateText(c.status, 3);
    lines.push(
      `| ${CORR_LABELS[c.key] ?? c.key} | ${c.n} | ${rhoStr} | ${pStr} | ${bandStr} |`,
    );
  }
  return lines.join("\n");
}

export function buildAcademicResultsParagraph(
  stats: AdminPaperStatisticsResponse,
  lang: "en" | "th",
): string {
  const n = stats.cohort.questionnaire_respondents;
  if (n === 0) {
    return lang === "th"
      ? "ยังไม่มีข้อมูลผู้เรียนที่ตอบแบบสอบถามหลังจบ Post-Test จึงยังไม่สามารถสร้างสรุปผลเชิงสถิติได้"
      : "No learner post-test evaluation responses have been recorded yet; statistical results will appear once participants complete the questionnaire.";
  }

  const csuqOverall = stats.csuq_subscales.find((s) => s.key === "overall");
  const csuqAlpha = stats.reliability.find((r) => r.key === "csuq_overall");
  const tlxOverall = stats.nasa_tlx.find((s) => s.key === "overall_workload");
  const tlxAlpha = stats.reliability.find((r) => r.key === "nasa_tlx_overall");
  const lg = stats.learning_gain;
  const t = lg.paired_t_test;
  const w = lg.wilcoxon;
  const corrGainCSUQ = stats.correlations.find((c) => c.key === "learning_gain_vs_csuq_overall");
  const corrGainTLX = stats.correlations.find((c) => c.key === "learning_gain_vs_nasa_tlx_overall");
  const corrCSUQTLX = stats.correlations.find((c) => c.key === "csuq_overall_vs_nasa_tlx_overall");

  if (lang === "th") {
    const parts: string[] = [];
    const pilotNote = stats.is_pilot_sample
      ? `ในการศึกษาแบบกลุ่มตัวอย่างนำร่อง (Pilot Study, N = ${n} < 30)`
      : `จากกลุ่มตัวอย่างผู้เรียนทั้งหมด N = ${n} คน`;
    parts.push(
      `${pilotNote} (คิดเป็นอัตราการตอบกลับ ${stats.cohort.response_rate_pct.toFixed(1)}% จากผู้สอบ Post-Test ${stats.cohort.completed_post_test_learners} คน)`,
    );

    if (isFiniteStat(csuqOverall?.mean)) {
      const sdCi =
        isFiniteStat(csuqOverall.sd) && isFiniteStat(csuqOverall.ci_low) && isFiniteStat(csuqOverall.ci_high)
          ? `, SD = ${csuqOverall.sd.toFixed(2)}, 95% CI [${csuqOverall.ci_low.toFixed(2)}, ${csuqOverall.ci_high.toFixed(2)}]`
          : "";
      const alphaStr = isFiniteStat(csuqAlpha?.alpha)
        ? `, Cronbach's α = ${csuqAlpha.alpha.toFixed(3)}`
        : "";
      parts.push(
        `ผู้เรียนประเมินความพึงพอใจและการใช้งานระบบ (CSUQ v3) โดยรวมอยู่ที่ M = ${csuqOverall.mean.toFixed(2)} / 7.00${sdCi}${alphaStr}`,
      );
    }

    if (isFiniteStat(tlxOverall?.mean)) {
      const sdCi =
        isFiniteStat(tlxOverall.sd) && isFiniteStat(tlxOverall.ci_low) && isFiniteStat(tlxOverall.ci_high)
          ? `, SD = ${tlxOverall.sd.toFixed(2)}, 95% CI [${tlxOverall.ci_low.toFixed(2)}, ${tlxOverall.ci_high.toFixed(2)}]`
          : "";
      const alphaStr = isFiniteStat(tlxAlpha?.alpha)
        ? `, α = ${tlxAlpha.alpha.toFixed(3)}`
        : "";
      parts.push(
        `และมีภาระงานทางความคิดรวมตามเกณฑ์ Raw NASA-TLX อยู่ที่ M = ${tlxOverall.mean.toFixed(2)} / 100${sdCi}${alphaStr}.`,
      );
    }

    if (isFiniteStat(lg.pre.mean) && isFiniteStat(lg.post.mean) && isFiniteStat(lg.difference.mean)) {
      const deltaVerb =
        lg.difference.mean > 0 ? "เพิ่มขึ้นจาก" : lg.difference.mean < 0 ? "ลดลงจาก" : "คงที่จาก";
      let gainSentence = `ด้านผลสัมฤทธิ์ทางการเรียน คะแนนเฉลี่ย${deltaVerb} ${lg.pre.mean.toFixed(2)}% (Pre-Test) เป็น ${lg.post.mean.toFixed(2)}% (Post-Test) คิดเป็นผลต่างเฉลี่ย ${lg.difference.mean > 0 ? "+" : ""}${lg.difference.mean.toFixed(2)}%`;
      if (t.status === "ok" && isFiniteStat(t.t_stat) && isFiniteStat(t.p_value) && isFiniteStat(t.cohens_dz)) {
        gainSentence += ` ซึ่งทดสอบด้วย Paired t-test พบ t(${t.df}) = ${t.t_stat.toFixed(3)}, p = ${fmtMdP(t.p_value)}, Cohen's d_z = ${t.cohens_dz.toFixed(3)}`;
      }
      if (w.status === "ok" && isFiniteStat(w.w_stat) && isFiniteStat(w.p_value) && isFiniteStat(w.rank_biserial)) {
        gainSentence += ` และสอดคล้องกับผล Wilcoxon signed-rank (W = ${w.w_stat.toFixed(1)}, p = ${fmtMdP(w.p_value)}, r_rb = ${w.rank_biserial.toFixed(3)})`;
      }
      if (t.status !== "ok" || w.status !== "ok") {
        const omitted: string[] = [];
        if (t.status !== "ok") omitted.push(`Paired t-test (${formatGateText(t.status, 3)})`);
        if (w.status !== "ok") omitted.push(`Wilcoxon (${formatGateText(w.status, 5)})`);
        gainSentence += ` (งดแสดงผล ${omitted.join(", ")} เนื่องจากข้อจำกัดจำนวนตัวอย่าง)`;
      }
      parts.push(`${gainSentence}.`);
    }

    const thCorrParts: string[] = [];
    if (corrGainCSUQ?.status === "ok" && isFiniteStat(corrGainCSUQ.rho) && isFiniteStat(corrGainCSUQ.p_value)) {
      thCorrParts.push(`พัฒนาการคะแนน↔CSUQ (ρ = ${corrGainCSUQ.rho.toFixed(3)}, p = ${fmtMdP(corrGainCSUQ.p_value)})`);
    }
    if (corrGainTLX?.status === "ok" && isFiniteStat(corrGainTLX.rho) && isFiniteStat(corrGainTLX.p_value)) {
      thCorrParts.push(`พัฒนาการคะแนน↔NASA-TLX (ρ = ${corrGainTLX.rho.toFixed(3)}, p = ${fmtMdP(corrGainTLX.p_value)})`);
    }
    if (corrCSUQTLX?.status === "ok" && isFiniteStat(corrCSUQTLX.rho) && isFiniteStat(corrCSUQTLX.p_value)) {
      thCorrParts.push(`CSUQ↔NASA-TLX (ρ = ${corrCSUQTLX.rho.toFixed(3)}, p = ${fmtMdP(corrCSUQTLX.p_value)})`);
    }
    if (thCorrParts.length > 0) {
      parts.push(`ผลสหสัมพันธ์แบบอันดับ Spearman พบว่า ${thCorrParts.join("; ")}.`);
    }

    return parts.join(" ");
  }

  // Academic English (IEEE/APA style)
  const sentences: string[] = [];
  const sampleLead = stats.is_pilot_sample
    ? `In this pilot evaluation (N = ${n}, ${stats.cohort.response_rate_pct.toFixed(1)}% response rate among ${stats.cohort.completed_post_test_learners} Post-Test completers)`
    : `Across the learner cohort (N = ${n}, ${stats.cohort.response_rate_pct.toFixed(1)}% response rate among ${stats.cohort.completed_post_test_learners} Post-Test completers)`;

  if (isFiniteStat(csuqOverall?.mean)) {
    const details: string[] = [`M = ${csuqOverall.mean.toFixed(2)} / 7.00`];
    if (isFiniteStat(csuqOverall.sd)) details.push(`SD = ${csuqOverall.sd.toFixed(2)}`);
    if (isFiniteStat(csuqOverall.ci_low) && isFiniteStat(csuqOverall.ci_high)) {
      details.push(`95% CI [${csuqOverall.ci_low.toFixed(2)}, ${csuqOverall.ci_high.toFixed(2)}]`);
    }
    if (isFiniteStat(csuqAlpha?.alpha)) {
      details.push(`Cronbach's α = ${csuqAlpha.alpha.toFixed(3)}`);
    }
    sentences.push(`${sampleLead}, participants reported perceived system usability on the CSUQ v3 of ${details.join(", ")}.`);
  } else {
    sentences.push(`${sampleLead}, evaluation metrics were collected.`);
  }

  if (isFiniteStat(tlxOverall?.mean)) {
    const details: string[] = [`M = ${tlxOverall.mean.toFixed(2)} / 100`];
    if (isFiniteStat(tlxOverall.sd)) details.push(`SD = ${tlxOverall.sd.toFixed(2)}`);
    if (isFiniteStat(tlxOverall.ci_low) && isFiniteStat(tlxOverall.ci_high)) {
      details.push(`95% CI [${tlxOverall.ci_low.toFixed(2)}, ${tlxOverall.ci_high.toFixed(2)}]`);
    }
    if (isFiniteStat(tlxAlpha?.alpha)) {
      details.push(`α = ${tlxAlpha.alpha.toFixed(3)}`);
    }
    sentences.push(`Overall cognitive task load measured by the Raw NASA-TLX was ${details.join(", ")}.`);
  }

  if (isFiniteStat(lg.pre.mean) && isFiniteStat(lg.post.mean) && isFiniteStat(lg.difference.mean)) {
    let gainText = `Assessment performance changed from M = ${lg.pre.mean.toFixed(2)}% on the Pre-Test to M = ${lg.post.mean.toFixed(2)}% on the Post-Test (mean Δ = ${lg.difference.mean > 0 ? "+" : ""}${lg.difference.mean.toFixed(2)}%)`;
    if (t.status === "ok" && isFiniteStat(t.t_stat) && isFiniteStat(t.p_value) && isFiniteStat(t.cohens_dz)) {
      gainText += `, paired t(${t.df}) = ${t.t_stat.toFixed(3)}, p = ${fmtMdP(t.p_value)}, Cohen's d_z = ${t.cohens_dz.toFixed(3)}`;
    }
    if (w.status === "ok" && isFiniteStat(w.w_stat) && isFiniteStat(w.p_value) && isFiniteStat(w.rank_biserial)) {
      gainText += `; non-parametric Wilcoxon W = ${w.w_stat.toFixed(1)} (n_eff = ${w.effective_n}), p = ${fmtMdP(w.p_value)}, r_rb = ${w.rank_biserial.toFixed(3)}`;
    }
    if (t.status !== "ok" || w.status !== "ok") {
      const deferred: string[] = [];
      if (t.status !== "ok") deferred.push(`paired t-test: ${formatGateText(t.status, 3)}`);
      if (w.status !== "ok") deferred.push(`Wilcoxon: ${formatGateText(w.status, 5)}`);
      gainText += ` (${deferred.join("; ")})`;
    }
    sentences.push(gainText + ".");
  }

  const corrClauses: string[] = [];
  if (corrGainCSUQ?.status === "ok" && isFiniteStat(corrGainCSUQ.rho) && isFiniteStat(corrGainCSUQ.p_value)) {
    corrClauses.push(`learning gain vs. CSUQ overall ρ = ${corrGainCSUQ.rho.toFixed(3)} (p = ${fmtMdP(corrGainCSUQ.p_value)}, n = ${corrGainCSUQ.n})`);
  }
  if (corrGainTLX?.status === "ok" && isFiniteStat(corrGainTLX.rho) && isFiniteStat(corrGainTLX.p_value)) {
    corrClauses.push(`learning gain vs. NASA-TLX workload ρ = ${corrGainTLX.rho.toFixed(3)} (p = ${fmtMdP(corrGainTLX.p_value)}, n = ${corrGainTLX.n})`);
  }
  if (corrCSUQTLX?.status === "ok" && isFiniteStat(corrCSUQTLX.rho) && isFiniteStat(corrCSUQTLX.p_value)) {
    corrClauses.push(`CSUQ overall vs. NASA-TLX workload ρ = ${corrCSUQTLX.rho.toFixed(3)} (p = ${fmtMdP(corrCSUQTLX.p_value)}, n = ${corrCSUQTLX.n})`);
  }
  if (corrClauses.length > 0) {
    sentences.push(`Spearman's rank correlations showed ${corrClauses.join("; ")}.`);
  } else if (stats.correlations.length > 0) {
    sentences.push(`Spearman's rank correlations were deferred (${formatGateText(stats.correlations[0].status, 3)}).`);
  }

  if (stats.is_pilot_sample) {
    sentences.push(
      "Because n < 30, these findings should be interpreted as pilot evidence subject to larger-sample replication.",
    );
  }

  return sentences.join(" ");
}
