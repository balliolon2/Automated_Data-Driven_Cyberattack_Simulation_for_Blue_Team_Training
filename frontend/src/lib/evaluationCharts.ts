import { round2 } from "./questionnaire.ts";

export const NASA_TLX_WORKLOAD_AXIS_KEYS = [
  "mental_demand",
  "physical_demand",
  "temporal_demand",
  "performance_workload",
  "effort",
  "frustration",
] as const;

export const CORRELATION_MIN_PAIRS = 3;

export const DARK_CHART_TOOLTIP_PROPS = {
  cursor: { fill: "rgba(255, 255, 255, 0.06)", stroke: "#52525b" },
  contentStyle: {
    backgroundColor: "#09090b",
    borderColor: "#3f3f46",
    color: "#f4f4f5",
    fontSize: 12,
    borderRadius: "0.375rem",
  },
  labelStyle: { color: "#ffffff", fontWeight: 600 },
  itemStyle: { color: "#e4e4e7" },
} as const;

export type NASATLXAxisKey = (typeof NASA_TLX_WORKLOAD_AXIS_KEYS)[number];

export interface TLXRadarDatum {
  key: NASATLXAxisKey;
  axis: string;
  workload: number;
  benchmark: number;
}

export interface CohortFunnelInput {
  total_learners: number;
  completed_pre_test_learners: number;
  completed_post_test_learners: number;
  questionnaire_respondents: number;
}

export interface CohortFunnelDatum {
  key: string;
  stage: string;
  count: number;
}

export type CSUQDistributionMode = "subscales" | "items";

export interface CSUQRowLike {
  csuq_scores?: {
    items?: Record<string, number | null>;
  };
}

export interface CSUQDistributionDatum {
  key: string;
  label: string;
  n: number;
  naCount: number;
  mean: number | null;
  lowCount: number;
  neutralCount: number;
  positiveCount: number;
  strongCount: number;
  lowPct: number;
  neutralPct: number;
  positivePct: number;
  strongPct: number;
  naPct: number;
}

export interface PersonalCSUQSubscaleDatum {
  key: "sysuse" | "infoqual" | "interqual" | "overall";
  subscale: string;
  score: number | null;
}

export interface PrePostParticipantLike {
  participant_code: string;
  pre_score: number;
  post_score: number;
  improvement_pct?: number;
  eval_date?: string;
}

export interface PrePostPairedDatum {
  label: string;
  pre: number;
  post: number;
}

export interface PrePostPairedSeries {
  rows: PrePostPairedDatum[];
  preMean: number | null;
  postMean: number | null;
}

export interface ParticipantScatterInput {
  participant_code: string;
  pre_score: number;
  post_score: number;
  improvement_pct: number;
  csuq_overall?: number | null;
  nasa_tlx_overall?: number | null;
}

export interface ScatterPlotPoint {
  participantCode: string;
  x: number;
  y: number;
}

export interface SpearmanScatterSeries {
  gainVsCsuq: ScatterPlotPoint[];
  gainVsTlx: ScatterPlotPoint[];
  csuqVsTlx: ScatterPlotPoint[];
}

export interface PersonalProgressDatum {
  key: "pre" | "post";
  stage: string;
  score: number;
}

const CSUQ_SUBSCALE_ITEM_GROUPS: ReadonlyArray<{ key: string; itemKeys: string[] }> = [
  { key: "sysuse", itemKeys: ["q1", "q2", "q3", "q4", "q5", "q6"] },
  { key: "infoqual", itemKeys: ["q7", "q8", "q9", "q10", "q11", "q12"] },
  { key: "interqual", itemKeys: ["q13", "q14", "q15"] },
  {
    key: "overall",
    itemKeys: [
      "q1", "q2", "q3", "q4", "q5", "q6", "q7", "q8",
      "q9", "q10", "q11", "q12", "q13", "q14", "q15", "q16",
    ],
  },
];

const ALL_CSUQ_ITEM_KEYS = [
  "q1", "q2", "q3", "q4", "q5", "q6", "q7", "q8",
  "q9", "q10", "q11", "q12", "q13", "q14", "q15", "q16",
] as const;

function clamp(val: number, min: number, max: number): number {
  if (!Number.isFinite(val)) return min;
  return Math.min(max, Math.max(min, val));
}

/** Shapes Raw NASA-TLX metrics into 6 canonical workload axes (0..100, using inverted performance_workload). */
export function buildTLXRadarData(
  dimensions: Partial<Record<NASATLXAxisKey | "performance_rating", number | null>> | null | undefined,
  labels: Record<NASATLXAxisKey, string>,
  benchmark = 50,
): TLXRadarDatum[] {
  const dims = dimensions ?? {};
  return NASA_TLX_WORKLOAD_AXIS_KEYS.map((key) => {
    let raw = dims[key];
    if (key === "performance_workload" && (raw === undefined || raw === null || !Number.isFinite(raw))) {
      const rating = dims.performance_rating;
      raw = typeof rating === "number" && Number.isFinite(rating) ? 100 - rating : 0;
    }
    return {
      key,
      axis: labels[key] ?? key,
      workload: round2(clamp(typeof raw === "number" ? raw : 0, 0, 100)),
      benchmark: round2(clamp(benchmark, 0, 100)),
    };
  });
}

/** Builds bounded Pre-Test vs. Post-Test percentage bars for individual learner summaries. */
export function buildPersonalProgressData(
  preScore: number,
  postScore: number,
  labels: { pre: string; post: string },
): PersonalProgressDatum[] {
  return [
    { key: "pre", stage: labels.pre, score: round2(clamp(preScore, 0, 100)) },
    { key: "post", stage: labels.post, score: round2(clamp(postScore, 0, 100)) },
  ];
}

/** Builds learner personal CSUQ subscale scores (1-7 scale) for bar chart rendering. */
export function buildPersonalCSUQData(
  subscales: Partial<Record<"sysuse" | "infoqual" | "interqual" | "overall", number | null>> | null | undefined,
  labels: Record<"sysuse" | "infoqual" | "interqual" | "overall", string>,
): PersonalCSUQSubscaleDatum[] {
  const keys = ["sysuse", "infoqual", "interqual", "overall"] as const;
  return keys.map((key) => {
    const val = subscales?.[key];
    const score = typeof val === "number" && Number.isFinite(val) ? round2(clamp(val, 0, 7)) : null;
    return {
      key,
      subscale: labels[key],
      score,
    };
  });
}

/** Shapes 4-stage cohort funnel counts from cohort totals. */
export function buildCohortFunnelData(
  cohort: CohortFunnelInput,
  labels: {
    totalLearners: string;
    preCompleted: string;
    postCompleted: string;
    respondents: string;
  },
): CohortFunnelDatum[] {
  return [
    { key: "total_learners", stage: labels.totalLearners, count: Math.max(0, cohort.total_learners) },
    { key: "completed_pre_test_learners", stage: labels.preCompleted, count: Math.max(0, cohort.completed_pre_test_learners) },
    { key: "completed_post_test_learners", stage: labels.postCompleted, count: Math.max(0, cohort.completed_post_test_learners) },
    { key: "questionnaire_respondents", stage: labels.respondents, count: Math.max(0, cohort.questionnaire_respondents) },
  ];
}

function normalizeLikertBuckets(counts: [number, number, number, number, number]): [number, number, number, number, number] {
  const total = counts[0] + counts[1] + counts[2] + counts[3] + counts[4];
  if (total <= 0) return [0, 0, 0, 0, 0];

  const pcts = counts.map((c) => round2((c / total) * 100)) as [number, number, number, number, number];
  let maxIdx = 0;
  for (let i = 1; i < 5; i++) {
    if (counts[i] > counts[maxIdx]) maxIdx = i;
  }
  const otherSum = pcts.reduce((acc, v, idx) => (idx === maxIdx ? acc : acc + v), 0);
  pcts[maxIdx] = Math.max(0, round2(100 - otherSum));
  return pcts;
}

/** Aggregates CSUQ Likert items (1-7 + N/A) into exact 100%-stacked distribution buckets. */
export function buildCSUQDistributionData(
  rows: CSUQRowLike[],
  mode: CSUQDistributionMode,
  labels: Record<string, string>,
): CSUQDistributionDatum[] {
  if (!rows || rows.length === 0) return [];

  const groups =
    mode === "subscales"
      ? CSUQ_SUBSCALE_ITEM_GROUPS
      : ALL_CSUQ_ITEM_KEYS.map((k) => ({ key: k, itemKeys: [k] }));

  return groups.map((group) => {
    let low = 0;
    let neutral = 0;
    let positive = 0;
    let strong = 0;
    let na = 0;
    let sum = 0;
    let validN = 0;

    for (const row of rows) {
      const itemMap = row.csuq_scores?.items ?? {};
      for (const itemKey of group.itemKeys) {
        const val = itemMap[itemKey];
        if (val === null || val === undefined || !Number.isFinite(val)) {
          na++;
          continue;
        }
        validN++;
        sum += val;
        if (val <= 2) low++;
        else if (val <= 4) neutral++;
        else if (val === 5) positive++;
        else strong++;
      }
    }

    const [lowPct, neutralPct, positivePct, strongPct, naPct] = normalizeLikertBuckets([
      low,
      neutral,
      positive,
      strong,
      na,
    ]);

    return {
      key: group.key,
      label: labels[group.key] ?? group.key.toUpperCase(),
      n: validN,
      naCount: na,
      mean: validN > 0 ? round2(sum / validN) : null,
      lowCount: low,
      neutralCount: neutral,
      positiveCount: positive,
      strongCount: strong,
      lowPct,
      neutralPct,
      positivePct,
      strongPct,
      naPct,
    };
  });
}

/**
 * Shapes paired Pre-Test vs. Post-Test cohort data oldest-to-newest.
 * Preserves input order when eval_date is absent (server-sorted participant_points) and prefers server-certified means when supplied.
 */
export function buildPrePostPairedData(
  items: PrePostParticipantLike[] | null | undefined,
  serverMeans?: { preMean?: number | null; postMean?: number | null },
): PrePostPairedSeries {
  if (!items || items.length === 0) {
    return { rows: [], preMean: null, postMean: null };
  }

  const indexed = items.map((item, idx) => ({
    item,
    idx,
    ts: item.eval_date ? Date.parse(item.eval_date) : Number.NaN,
  }));

  indexed.sort((a, b) => {
    if (Number.isFinite(a.ts) && Number.isFinite(b.ts) && a.ts !== b.ts) {
      return a.ts - b.ts;
    }
    return a.idx - b.idx;
  });

  let preSum = 0;
  let postSum = 0;
  const rows: PrePostPairedDatum[] = indexed.map(({ item }) => {
    const pre = round2(clamp(item.pre_score, 0, 100));
    const post = round2(clamp(item.post_score, 0, 100));
    preSum += pre;
    postSum += post;
    return {
      label: item.participant_code,
      pre,
      post,
    };
  });

  const preMean =
    typeof serverMeans?.preMean === "number" && Number.isFinite(serverMeans.preMean)
      ? round2(serverMeans.preMean)
      : round2(preSum / rows.length);
  const postMean =
    typeof serverMeans?.postMean === "number" && Number.isFinite(serverMeans.postMean)
      ? round2(serverMeans.postMean)
      : round2(postSum / rows.length);

  return { rows, preMean, postMean };
}

/** Builds 3 Spearman scatter pair series while filtering out null or non-finite coordinates per construct pair. */
export function buildSpearmanScatterSeries(
  points: ParticipantScatterInput[] | null | undefined,
): SpearmanScatterSeries {
  const gainVsCsuq: ScatterPlotPoint[] = [];
  const gainVsTlx: ScatterPlotPoint[] = [];
  const csuqVsTlx: ScatterPlotPoint[] = [];

  for (const pt of points ?? []) {
    const gain = typeof pt.improvement_pct === "number" && Number.isFinite(pt.improvement_pct)
      ? round2(pt.improvement_pct)
      : null;
    const csuq = typeof pt.csuq_overall === "number" && Number.isFinite(pt.csuq_overall)
      ? round2(pt.csuq_overall)
      : null;
    const tlx = typeof pt.nasa_tlx_overall === "number" && Number.isFinite(pt.nasa_tlx_overall)
      ? round2(pt.nasa_tlx_overall)
      : null;

    if (gain !== null && csuq !== null) {
      gainVsCsuq.push({ participantCode: pt.participant_code, x: gain, y: csuq });
    }
    if (gain !== null && tlx !== null) {
      gainVsTlx.push({ participantCode: pt.participant_code, x: gain, y: tlx });
    }
    if (csuq !== null && tlx !== null) {
      csuqVsTlx.push({ participantCode: pt.participant_code, x: csuq, y: tlx });
    }
  }

  return { gainVsCsuq, gainVsTlx, csuqVsTlx };
}

/** Gates correlation scatter figures to require at least CORRELATION_MIN_PAIRS complete pairs. */
export function isCorrelationChartUnlocked(pairCount: number): boolean {
  return Number.isFinite(pairCount) && pairCount >= CORRELATION_MIN_PAIRS;
}
