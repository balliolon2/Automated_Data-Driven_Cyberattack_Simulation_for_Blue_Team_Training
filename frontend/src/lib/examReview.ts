import type { QuestionTranslationsMap } from "./questionLocale";

export type ReviewStatusFilter = "all" | "incorrect" | "unanswered" | "correct";
export type ReviewGroupMode = "domain" | "order";

export interface ExamReviewItem {
  question_id: string;
  domain_id: string;
  domain_name: string;
  sub_topic_id?: string | null;
  sub_topic_title?: string;
  type: string;
  question_text: string;
  options: Record<string, string>;
  correct_answer: string;
  explanation: string;
  translations?: QuestionTranslationsMap | null;
  mitre_ref?: string | null;
  order_index: number;
  user_answer: string;
  is_correct: boolean;
  is_unanswered?: boolean;
}

export interface ExamDomainScore {
  domain_id: string;
  domain_name: string;
  total: number;
  correct: number;
  incorrect: number;
  unanswered: number;
  percentage: number;
  threshold: number;
  passed: boolean;
}

export interface ExamHistoryItem {
  session_id: string;
  exam_type: "pre" | "post" | string;
  status: string;
  score: number;
  correct_count: number;
  incorrect_count: number;
  unanswered_count: number;
  total_questions: number;
  started_at: string;
  completed_at?: string | null;
  time_spent_seconds?: number | null;
}

export interface ExamHistoryResponse {
  latest_pre_session_id: string | null;
  latest_post_session_id?: string | null;
  latest_pre_to_post_delta?: number | null;
  sessions: ExamHistoryItem[];
}

export const EXAM_MAX_DURATION_SECONDS = 7200;

export type ReviewOutcome = "correct" | "unanswered" | "incorrect";

export function getReviewItemOutcome(
  item: Pick<ExamReviewItem, "is_correct" | "is_unanswered" | "user_answer">,
): ReviewOutcome {
  if (item.is_correct) return "correct";
  if (isReviewItemUnanswered(item)) return "unanswered";
  return "incorrect";
}

export function formatSignedDelta(delta: number): string {
  if (delta > 0) return `+${delta.toFixed(1)}%`;
  if (delta < 0) return `${delta.toFixed(1)}%`;
  return "0.0%";
}

export function getExamTypeBadgeClass(examType: string): string {
  return examType === "post"
    ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
    : "bg-sky-500/10 text-sky-300 border-sky-500/30";
}

export interface ExamDomainDelta {
  domain_id: string;
  domain_name: string;
  pre_percentage: number;
  post_percentage: number;
  delta: number;
}

export interface ExamPreComparison {
  pre_session_id: string;
  pre_score: number;
  post_score: number;
  score_delta: number;
  pre_correct_count: number;
  post_correct_count: number;
  domain_deltas: ExamDomainDelta[];
}

export interface ExamReviewResponse {
  session: ExamHistoryItem;
  domain_scores: ExamDomainScore[];
  pre_comparison: ExamPreComparison | null;
  results: ExamReviewItem[];
}

export interface DistractorExplanation {
  option: string;
  rationale: string;
}

export type ParsedExplanation =
  | {
      structured: true;
      correctOption: string;
      correctRationale: string;
      distractors: DistractorExplanation[];
    }
  | {
      structured: false;
      raw: string;
    };

export interface ReviewSection {
  key: string;
  domainId: string | null;
  domainScore: ExamDomainScore | null;
  items: ExamReviewItem[];
}

export const CANONICAL_DOMAIN_ORDER = [
  "domain1",
  "domain2",
  "domain3",
  "domain4",
  "domain5",
] as const;

export function isReviewItemUnanswered(
  item: Pick<ExamReviewItem, "is_unanswered" | "user_answer">,
): boolean {
  return Boolean(item.is_unanswered || !item.user_answer || item.user_answer.trim() === "");
}

export function getSortedOptionEntries(
  options: Record<string, string> | null | undefined,
): Array<[string, string]> {
  if (!options || typeof options !== "object") {
    return [];
  }
  const orderWeight = (key: string) => {
    const idx = ["A", "B", "C", "D"].indexOf(key.toUpperCase());
    return idx >= 0 ? idx : 99;
  };
  return Object.entries(options).sort(([ka], [kb]) => {
    const wa = orderWeight(ka);
    const wb = orderWeight(kb);
    if (wa !== wb) return wa - wb;
    return ka.localeCompare(kb);
  });
}

export function parseExamExplanation(
  explanation: string,
  fallbackCorrectAnswer?: string,
): ParsedExplanation {
  const trimmed = (explanation ?? "").trim();
  if (!trimmed) {
    return { structured: false, raw: "" };
  }

  const markerRegex =
    /(?:\bOption\s+([A-D])\s+is\s+incorrect\.?\s*|ตัวเลือก\s*([A-D])\s*ไม่ถูกต้อง[.:：]?\s*)/gi;
  const matches: Array<{ option: string; start: number; bodyStart: number }> = [];
  for (const m of trimmed.matchAll(markerRegex)) {
    const matchedOpt = (m[1] ?? m[2] ?? "").toUpperCase();
    if (typeof m.index === "number" && matchedOpt) {
      matches.push({
        option: matchedOpt,
        start: m.index,
        bodyStart: m.index + m[0].length,
      });
    }
  }

  if (matches.length === 0) {
    return { structured: false, raw: trimmed };
  }

  const rawPrefix = trimmed.slice(0, matches[0].start).trim();
  const correctHeader = rawPrefix.match(
    /^(?:Correct\s+Answer:\s*([A-D])\.?\s*|คำตอบที่ถูกต้อง\s*[:：]?\s*([A-D])[.:：]?\s*)/i,
  );
  const headerOption = (correctHeader?.[1] ?? correctHeader?.[2] ?? "").toUpperCase();
  const correctOption = headerOption
    ? headerOption
    : (fallbackCorrectAnswer ?? "").trim().toUpperCase();
  const correctRationale = correctHeader
    ? rawPrefix.slice(correctHeader[0].length).trim()
    : rawPrefix;

  if (!correctOption || !correctRationale) {
    return { structured: false, raw: trimmed };
  }

  const distractors: DistractorExplanation[] = [];
  for (let i = 0; i < matches.length; i++) {
    const cur = matches[i];
    const end = i + 1 < matches.length ? matches[i + 1].start : trimmed.length;
    const rationale = trimmed.slice(cur.bodyStart, end).trim();
    if (!rationale) {
      return { structured: false, raw: trimmed };
    }
    distractors.push({ option: cur.option, rationale });
  }

  return {
    structured: true,
    correctOption,
    correctRationale,
    distractors,
  };
}

export function orderDistractorsForLearner(
  distractors: readonly DistractorExplanation[],
  userAnswer: string,
): DistractorExplanation[] {
  const sorted = [...distractors].sort((a, b) => a.option.localeCompare(b.option));
  const pick = (userAnswer ?? "").trim().toUpperCase();
  if (!pick) {
    return sorted;
  }
  const pickIdx = sorted.findIndex((d) => d.option.toUpperCase() === pick);
  if (pickIdx <= 0) {
    return sorted;
  }
  return [sorted[pickIdx], ...sorted.slice(0, pickIdx), ...sorted.slice(pickIdx + 1)];
}

export function createInitialExpandedIds(results: readonly ExamReviewItem[]): Set<string> {
  const expanded = new Set<string>();
  for (const item of results) {
    if (!item.is_correct) {
      expanded.add(item.question_id);
    }
  }
  return expanded;
}

export function countReviewOutcomes(results: readonly ExamReviewItem[]): {
  all: number;
  correct: number;
  incorrect: number;
  unanswered: number;
} {
  let correct = 0;
  let incorrect = 0;
  let unanswered = 0;
  for (const item of results) {
    if (item.is_correct) {
      correct++;
    } else if (isReviewItemUnanswered(item)) {
      unanswered++;
    } else {
      incorrect++;
    }
  }
  return {
    all: results.length,
    correct,
    incorrect,
    unanswered,
  };
}

export function matchesStatusFilter(
  item: ExamReviewItem,
  statusFilter: ReviewStatusFilter,
): boolean {
  if (statusFilter === "all") return true;
  return getReviewItemOutcome(item) === statusFilter;
}

export function filterReviewItems(
  results: readonly ExamReviewItem[],
  statusFilter: ReviewStatusFilter,
  domainFilter: string,
): ExamReviewItem[] {
  return results.filter((item) => {
    if (!matchesStatusFilter(item, statusFilter)) {
      return false;
    }
    if (domainFilter && domainFilter !== "all" && item.domain_id !== domainFilter) {
      return false;
    }
    return true;
  });
}

export function groupReviewItems(
  items: readonly ExamReviewItem[],
  groupMode: ReviewGroupMode,
  domainScores: readonly ExamDomainScore[],
): ReviewSection[] {
  const sorted = [...items].sort((a, b) => a.order_index - b.order_index);
  if (groupMode === "order") {
    if (sorted.length === 0) return [];
    return [
      {
        key: "order",
        domainId: null,
        domainScore: null,
        items: sorted,
      },
    ];
  }

  const byDomain = new Map<string, ExamReviewItem[]>();
  for (const item of sorted) {
    const list = byDomain.get(item.domain_id) ?? [];
    list.push(item);
    byDomain.set(item.domain_id, list);
  }

  const scoreMap = new Map<string, ExamDomainScore>();
  for (const ds of domainScores) {
    scoreMap.set(ds.domain_id, ds);
  }

  const orderedIds: string[] = [...CANONICAL_DOMAIN_ORDER];
  for (const domId of byDomain.keys()) {
    if (!orderedIds.includes(domId)) {
      orderedIds.push(domId);
    }
  }

  const sections: ReviewSection[] = [];
  for (const domId of orderedIds) {
    const domItems = byDomain.get(domId);
    if (!domItems || domItems.length === 0) {
      continue;
    }
    sections.push({
      key: domId,
      domainId: domId,
      domainScore: scoreMap.get(domId) ?? null,
      items: domItems,
    });
  }
  return sections;
}

export function resolveJumpToQuestion(
  target: ExamReviewItem,
  currentStatusFilter: ReviewStatusFilter,
  currentDomainFilter: string,
  currentExpandedIds: ReadonlySet<string>,
): {
  nextStatusFilter: ReviewStatusFilter;
  nextDomainFilter: string;
  nextExpandedIds: Set<string>;
} {
  const nextStatusFilter: ReviewStatusFilter = matchesStatusFilter(target, currentStatusFilter)
    ? currentStatusFilter
    : "all";
  const nextDomainFilter =
    currentDomainFilter !== "all" && currentDomainFilter !== target.domain_id
      ? "all"
      : currentDomainFilter;
  const nextExpandedIds = new Set(currentExpandedIds);
  nextExpandedIds.add(target.question_id);
  return {
    nextStatusFilter,
    nextDomainFilter,
    nextExpandedIds,
  };
}

export function formatDurationSeconds(seconds?: number | null): string {
  if (seconds === undefined || seconds === null || Number.isNaN(seconds)) {
    return "—";
  }
  const safe = Math.max(0, Math.floor(seconds));
  const hours = Math.floor(safe / 3600);
  const mins = Math.floor((safe % 3600) / 60);
  const secs = safe % 60;
  if (hours > 0) {
    return `${hours.toString().padStart(2, "0")}:${mins.toString().padStart(2, "0")}:${secs
      .toString()
      .padStart(2, "0")}`;
  }
  return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
}
