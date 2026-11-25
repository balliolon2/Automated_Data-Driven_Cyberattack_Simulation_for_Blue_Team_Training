import { useEffect, useMemo, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import axios from "axios";
import {
  ArrowLeft,
  ArrowRight,
  Award,
  BookOpen,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Clock,
  Layers,
  ListOrdered,
  Loader2,
  ShieldAlert,
  TrendingDown,
  TrendingUp,
  XCircle,
  AlertCircle,
} from "lucide-react";
import DomainBreakdown from "../components/DomainBreakdown";
import QuestionnairePromptBanner from "../components/QuestionnairePromptBanner";
import type { QuestionnaireStatusResponse } from "../lib/questionnaire";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError, resolveDomainName } from "../i18n/format.ts";
import { cn } from "../lib/utils";
import {
  createInitialExpandedIds,
  filterReviewItems,
  formatDurationSeconds,
  formatSignedDelta,
  getExamTypeBadgeClass,
  getReviewItemOutcome,
  getSortedOptionEntries,
  groupReviewItems,
  orderDistractorsForLearner,
  parseExamExplanation,
  resolveJumpToQuestion,
  type ExamPreComparison,
  type ExamReviewItem,
  type ExamReviewResponse,
  type ReviewGroupMode,
  type ReviewStatusFilter,
} from "../lib/examReview";

type ReviewErrorState =
  | { kind: "not_completed"; examType: string }
  | { kind: "api"; err: unknown }
  | null;

function PrePostGainCard({
  comparison,
  totalQuestions,
}: {
  comparison: ExamPreComparison;
  totalQuestions: number;
}) {
  const { locale, t } = useI18n();
  const isPositive = comparison.score_delta > 0;
  const isNegative = comparison.score_delta < 0;

  return (
    <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-5">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-graphite-800 pb-4">
        <div className="space-y-1">
          <h2 className="text-lg font-semibold text-white flex items-center gap-2">
            {isNegative ? (
              <TrendingDown className="w-5 h-5 text-amber-400" />
            ) : (
              <TrendingUp className="w-5 h-5 text-emerald-400" />
            )}
            {t("examReview.gainCardTitle")}
          </h2>
          <p className="text-sm text-graphite-400">
            {t("examReview.gainCardSubtitle", {
              preCorrect: comparison.pre_correct_count,
              postCorrect: comparison.post_correct_count,
              total: totalQuestions,
            })}
          </p>
        </div>

        <div className="flex items-center gap-3 font-mono shrink-0 bg-graphite-950 border border-graphite-800 rounded-md px-4 py-2.5">
          <div className="text-center">
            <div className="text-xs text-graphite-400">{t("examReview.gainPreLabel")}</div>
            <div className="text-base font-semibold text-graphite-200">
              {comparison.pre_score.toFixed(1)}%
            </div>
          </div>
          <ArrowRight className="w-4 h-4 text-graphite-400" />
          <div className="text-center">
            <div className="text-xs text-graphite-400">{t("examReview.gainPostLabel")}</div>
            <div className="text-base font-bold text-white">
              {comparison.post_score.toFixed(1)}%
            </div>
          </div>
          <span
            className={cn(
              "ml-1 px-2.5 py-1 rounded text-xs font-semibold border",
              isPositive
                ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                : isNegative
                  ? "bg-amber-500/10 text-amber-400 border-amber-500/30"
                  : "bg-graphite-800 text-graphite-300 border-graphite-700",
            )}
          >
            {formatSignedDelta(comparison.score_delta)}
          </span>
        </div>
      </div>

      {comparison.domain_deltas.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
          {comparison.domain_deltas.map((dom) => {
            const domName = resolveDomainName(dom.domain_id, dom.domain_name, locale);
            const domPositive = dom.delta > 0;
            const domNegative = dom.delta < 0;
            return (
              <div
                key={dom.domain_id}
                className="rounded-md border border-graphite-800 bg-graphite-950/60 p-3.5 space-y-2 flex flex-col justify-between"
              >
                <div className="text-sm font-medium text-graphite-200 leading-snug break-words">
                  {domName}
                </div>
                <div className="flex items-center justify-between gap-2 pt-1 border-t border-graphite-800/70 font-mono">
                  <span className="text-xs text-graphite-400">
                    {dom.pre_percentage.toFixed(0)}% → {dom.post_percentage.toFixed(0)}%
                  </span>
                  <span
                    className={cn(
                      "text-xs font-semibold",
                      domPositive
                        ? "text-emerald-400"
                        : domNegative
                          ? "text-amber-400"
                          : "text-graphite-400",
                    )}
                  >
                    {formatSignedDelta(dom.delta)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

function StructuredExplanationView({
  explanation,
  correctAnswer,
  userAnswer,
}: {
  explanation: string;
  correctAnswer: string;
  userAnswer: string;
}) {
  const { t } = useI18n();
  const parsed = useMemo(
    () => parseExamExplanation(explanation, correctAnswer),
    [explanation, correctAnswer],
  );

  if (!parsed.structured) {
    if (!parsed.raw) return null;
    return (
      <div className="rounded-md border border-graphite-800 bg-graphite-950/70 p-4 space-y-2">
        <h4 className="text-sm font-semibold text-white">{t("examReview.explanationTitle")}</h4>
        <p className="text-base text-graphite-200 leading-relaxed whitespace-pre-line">
          {parsed.raw}
        </p>
      </div>
    );
  }

  const orderedDistractors = orderDistractorsForLearner(parsed.distractors, userAnswer);
  const pickedUpper = (userAnswer ?? "").trim().toUpperCase();

  return (
    <div className="rounded-md border border-graphite-800 bg-graphite-950/70 p-4 sm:p-5 space-y-4">
      <div className="space-y-1.5 border-l-2 border-emerald-500 pl-3.5">
        <h4 className="text-sm font-semibold text-emerald-400">
          {t("examReview.whyCorrectTitle", { option: parsed.correctOption })}
        </h4>
        <p className="text-base text-graphite-100 leading-relaxed">{parsed.correctRationale}</p>
      </div>

      {orderedDistractors.length > 0 && (
        <div className="space-y-2.5 pt-3 border-t border-graphite-800/80">
          <h4 className="text-sm font-semibold text-graphite-300">
            {t("examReview.whyIncorrectTitle")}
          </h4>
          <ul className="space-y-2.5">
            {orderedDistractors.map((dist) => {
              const isUserWrongPick = pickedUpper !== "" && pickedUpper === dist.option;
              return (
                <li
                  key={dist.option}
                  className={cn(
                    "rounded-md border px-3.5 py-2.5 space-y-1",
                    isUserWrongPick
                      ? "border-red-500/40 bg-red-950/15 border-l-2 border-l-red-400"
                      : "border-graphite-800/80 bg-graphite-900/50",
                  )}
                >
                  <div className="flex flex-wrap items-center gap-2">
                    <span
                      className={cn(
                        "text-xs font-mono font-semibold",
                        isUserWrongPick ? "text-red-300" : "text-graphite-300",
                      )}
                    >
                      {t("examReview.optionLabel", { option: dist.option })}
                    </span>
                    {isUserWrongPick && (
                      <span className="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-red-500/15 text-red-300 border border-red-500/30">
                        {t("examReview.yourPickBadge")}
                      </span>
                    )}
                  </div>
                  <p className="text-base text-graphite-200 leading-relaxed">{dist.rationale}</p>
                </li>
              );
            })}
          </ul>
        </div>
      )}
    </div>
  );
}

function ReviewQuestionCard({
  item,
  isExpanded,
  isHighlighted,
  showDomainBadge,
  onToggleExpand,
  cardRef,
}: {
  item: ExamReviewItem;
  isExpanded: boolean;
  isHighlighted: boolean;
  showDomainBadge: boolean;
  onToggleExpand: () => void;
  cardRef: (el: HTMLDivElement | null) => void;
}) {
  const { locale, t } = useI18n();
  const outcome = getReviewItemOutcome(item);
  const unanswered = outcome === "unanswered";
  const questionNum = (item.order_index + 1).toString().padStart(2, "0");
  const domainLabel = resolveDomainName(item.domain_id, item.domain_name, locale);
  const subtopicHref = item.sub_topic_id
    ? `/learn?domain=${encodeURIComponent(item.domain_id)}&topic=${encodeURIComponent(item.sub_topic_id)}`
    : `/learn?domain=${encodeURIComponent(item.domain_id)}`;

  const statusBadge = outcome === "correct" ? (
    <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded text-xs font-mono font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 whitespace-nowrap">
      <CheckCircle2 className="w-3.5 h-3.5" />
      {t("examReview.statusCorrect")}
    </span>
  ) : outcome === "unanswered" ? (
    <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded text-xs font-mono font-semibold bg-amber-500/10 text-amber-300 border border-amber-500/30 whitespace-nowrap">
      <AlertCircle className="w-3.5 h-3.5" />
      {t("examReview.statusUnanswered")}
    </span>
  ) : (
    <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded text-xs font-mono font-semibold bg-red-500/10 text-red-400 border border-red-500/30 whitespace-nowrap">
      <XCircle className="w-3.5 h-3.5" />
      {t("examReview.statusIncorrect")}
    </span>
  );

  if (!isExpanded) {
    return (
      <div
        id={`review-q-${item.order_index}`}
        ref={cardRef}
        className={cn(
          "bg-graphite-900 border rounded-lg transition-colors",
          isHighlighted
            ? "border-white ring-2 ring-white/40"
            : "border-graphite-800 hover:border-graphite-700",
        )}
      >
        <button
          type="button"
          onClick={onToggleExpand}
          aria-expanded={false}
          className="w-full p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-left cursor-pointer"
        >
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <span className="px-2 py-0.5 rounded bg-graphite-950 border border-graphite-800 text-xs font-mono font-semibold text-white shrink-0">
              {t("examReview.questionNumber", { number: questionNum })}
            </span>
            {statusBadge}
            {item.sub_topic_id && (
              <span className="px-2 py-0.5 rounded bg-graphite-950 border border-graphite-800 text-xs font-mono text-graphite-300 shrink-0">
                {item.sub_topic_id}
              </span>
            )}
            <span className="text-sm text-graphite-200 truncate">{item.question_text}</span>
          </div>
          <span className="inline-flex items-center gap-1 text-sm font-medium text-graphite-400 hover:text-white shrink-0">
            {t("examReview.showExplanation")}
            <ChevronDown className="w-4 h-4" />
          </span>
        </button>
      </div>
    );
  }

  const sortedOptions = getSortedOptionEntries(item.options);
  const pickedKey = (item.user_answer ?? "").trim().toUpperCase();
  const correctKey = (item.correct_answer ?? "").trim().toUpperCase();

  return (
    <div
      id={`review-q-${item.order_index}`}
      ref={cardRef}
      className={cn(
        "bg-graphite-900 border rounded-lg p-5 sm:p-6 space-y-5 transition-all",
        isHighlighted
          ? "border-white ring-2 ring-white/40"
          : outcome === "correct"
            ? "border-graphite-800"
            : outcome === "unanswered"
              ? "border-amber-500/30"
              : "border-red-500/30",
      )}
    >
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-graphite-800 pb-4">
        <div className="flex flex-wrap items-center gap-2">
          <span className="px-2.5 py-1 rounded bg-graphite-950 border border-graphite-800 text-xs font-mono font-bold text-white">
            {t("examReview.questionNumber", { number: questionNum })}
          </span>
          {statusBadge}
          {showDomainBadge && (
            <span className="px-2.5 py-0.5 rounded bg-graphite-950 border border-graphite-800 text-xs text-graphite-300">
              {domainLabel}
            </span>
          )}
          {item.sub_topic_id && (
            <span className="px-2.5 py-0.5 rounded bg-graphite-950 border border-graphite-800 text-xs font-mono text-graphite-300">
              {item.sub_topic_title
                ? t("examReview.subtopicBadge", {
                    id: item.sub_topic_id,
                    title: item.sub_topic_title,
                  })
                : item.sub_topic_id}
            </span>
          )}
        </div>

        <button
          type="button"
          onClick={onToggleExpand}
          aria-expanded={true}
          className="inline-flex items-center gap-1 text-sm font-medium text-graphite-400 hover:text-white transition-colors cursor-pointer"
        >
          {t("examReview.hideExplanation")}
          <ChevronUp className="w-4 h-4" />
        </button>
      </div>

      <p className="text-base text-graphite-100 leading-relaxed font-normal">
        {item.question_text}
      </p>

      {unanswered && (
        <div className="px-3.5 py-2.5 rounded-md bg-amber-500/10 border border-amber-500/30 text-amber-300 text-sm flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{t("examReview.unansweredCallout")}</span>
        </div>
      )}

      <div className="space-y-2.5">
        {sortedOptions.map(([optKey, optText]) => {
          const normKey = optKey.toUpperCase();
          const isOptionCorrect = normKey === correctKey;
          const isUserChoice = pickedKey !== "" && normKey === pickedKey;

          let rowStyle = "border-graphite-800 bg-graphite-950/40 text-graphite-300";
          if (isOptionCorrect) {
            rowStyle = "border-emerald-500/45 bg-emerald-500/10 text-white";
          } else if (isUserChoice) {
            rowStyle = "border-red-500/45 bg-red-500/10 text-red-100";
          }

          return (
            <div
              key={optKey}
              className={cn(
                "flex flex-col sm:flex-row sm:items-start justify-between gap-2.5 p-3.5 rounded-md border text-base leading-relaxed",
                rowStyle,
              )}
            >
              <div className="flex items-start gap-3 min-w-0">
                <span className="font-mono font-semibold text-white shrink-0">{optKey})</span>
                <span className="break-words">{optText}</span>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                {isUserChoice && (
                  <span
                    className={cn(
                      "inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-mono font-semibold border whitespace-nowrap",
                      isOptionCorrect
                        ? "bg-emerald-500/15 text-emerald-300 border-emerald-500/40"
                        : "bg-red-500/15 text-red-300 border-red-500/40",
                    )}
                  >
                    {!isOptionCorrect && <XCircle className="w-3.5 h-3.5" />}
                    {t("examReview.yourAnswerBadge")}
                  </span>
                )}
                {isOptionCorrect && (
                  <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-mono font-semibold bg-emerald-500/15 text-emerald-300 border border-emerald-500/40 whitespace-nowrap">
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    {t("examReview.correctAnswerBadge")}
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>

      <StructuredExplanationView
        explanation={item.explanation}
        correctAnswer={item.correct_answer}
        userAnswer={item.user_answer}
      />

      <div className="flex flex-wrap items-center justify-end gap-3 pt-2 border-t border-graphite-800/70">
        <Link
          to={subtopicHref}
          className="inline-flex items-center gap-2 px-3.5 py-2 rounded-md border border-graphite-700 bg-graphite-950 hover:bg-graphite-800 text-sm font-medium text-white transition-colors"
        >
          <BookOpen className="w-4 h-4 text-emerald-400" />
          <span>
            {item.sub_topic_id
              ? t("examReview.studySubtopicCta", { id: item.sub_topic_id })
              : t("examReview.studyDomainFallbackCta", { domain: domainLabel })}
          </span>
          <ArrowRight className="w-4 h-4" />
        </Link>
      </div>
    </div>
  );
}

export default function ExamReviewPage() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { locale, t, formatDate } = useI18n();

  const [review, setReview] = useState<ExamReviewResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [errorState, setErrorState] = useState<ReviewErrorState>(null);

  const [statusFilter, setStatusFilter] = useState<ReviewStatusFilter>("all");
  const [domainFilter, setDomainFilter] = useState<string>("all");
  const [groupMode, setGroupMode] = useState<ReviewGroupMode>("domain");
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() => new Set());
  const [highlightedOrderIdx, setHighlightedOrderIdx] = useState<number | null>(null);
  const [questionnaireStatus, setQuestionnaireStatus] = useState<QuestionnaireStatusResponse | null>(null);

  const cardRefs = useRef<Record<number, HTMLDivElement | null>>({});

  useEffect(() => {
    let cancelled = false;

    async function loadReview() {
      const token = localStorage.getItem("token");
      if (!token) {
        navigate("/login");
        return;
      }
      if (!sessionId) {
        if (!cancelled) {
          setErrorState({ kind: "api", err: new Error("Missing session ID") });
          setLoading(false);
        }
        return;
      }

      try {
        const res = await axios.get<ExamReviewResponse>(
          `/api/exams/sessions/${encodeURIComponent(sessionId)}/review`,
          { headers: { Authorization: `Bearer ${token}` } },
        );
        if (cancelled) return;
        setReview(res.data);
        setExpandedIds(createInitialExpandedIds(res.data.results ?? []));
        setStatusFilter("all");

        if (res.data.session?.exam_type === "post") {
          axios
            .get<QuestionnaireStatusResponse>("/api/evaluation/questionnaire/status", {
              headers: { Authorization: `Bearer ${token}` },
            })
            .then((qRes) => {
              if (!cancelled) {
                setQuestionnaireStatus(qRes.data);
              }
            })
            .catch(() => {
              // Fail-open so questionnaire status fetch never blocks Post-Test review
            });
        } else {
          setQuestionnaireStatus(null);
        }
        setDomainFilter("all");
        setGroupMode("domain");
        setErrorState(null);
      } catch (err: unknown) {
        if (cancelled) return;
        const resp = axios.isAxiosError(err) ? err.response : undefined;
        if (resp?.status === 401) {
          navigate("/login");
          return;
        }
        if (resp?.status === 403 && resp?.data?.code === "EXAM_NOT_COMPLETED") {
          const examTypeRaw = resp.data?.params?.exam_type ?? resp.data?.exam_type;
          setErrorState({
            kind: "not_completed",
            examType: typeof examTypeRaw === "string" ? examTypeRaw : "pre",
          });
        } else {
          setErrorState({ kind: "api", err });
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void loadReview();
    return () => {
      cancelled = true;
    };
  }, [sessionId, navigate]);

  useEffect(() => {
    if (highlightedOrderIdx === null) return;
    const timer = window.setTimeout(() => {
      setHighlightedOrderIdx(null);
    }, 1800);
    return () => window.clearTimeout(timer);
  }, [highlightedOrderIdx]);

  const visibleItems = useMemo(
    () => filterReviewItems(review?.results ?? [], statusFilter, domainFilter),
    [review, statusFilter, domainFilter],
  );

  const sections = useMemo(
    () => groupReviewItems(visibleItems, groupMode, review?.domain_scores ?? []),
    [visibleItems, groupMode, review],
  );

  const allVisibleExpanded = useMemo(
    () => visibleItems.length > 0 && visibleItems.every((q) => expandedIds.has(q.question_id)),
    [visibleItems, expandedIds],
  );

  const handleToggleExpandAll = () => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (allVisibleExpanded) {
        for (const q of visibleItems) {
          next.delete(q.question_id);
        }
      } else {
        for (const q of visibleItems) {
          next.add(q.question_id);
        }
      }
      return next;
    });
  };

  const handleToggleCard = (questionId: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(questionId)) {
        next.delete(questionId);
      } else {
        next.add(questionId);
      }
      return next;
    });
  };

  const handleJumpToQuestion = (target: ExamReviewItem) => {
    const resolved = resolveJumpToQuestion(target, statusFilter, domainFilter, expandedIds);
    setStatusFilter(resolved.nextStatusFilter);
    setDomainFilter(resolved.nextDomainFilter);
    setExpandedIds(resolved.nextExpandedIds);
    setHighlightedOrderIdx(target.order_index);

    window.requestAnimationFrame(() => {
      window.requestAnimationFrame(() => {
        cardRefs.current[target.order_index]?.scrollIntoView({
          behavior: "smooth",
          block: "center",
        });
      });
    });
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-sm font-mono text-graphite-400">{t("examReview.loading")}</p>
      </div>
    );
  }

  if (errorState?.kind === "not_completed") {
    const resumePath = errorState.examType === "post" ? "/post-test" : "/pre-test";
    const resumeLabel =
      errorState.examType === "post"
        ? t("examReview.resumePostTest")
        : t("examReview.resumePreTest");

    return (
      <div className="max-w-lg mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
        <ShieldAlert className="w-10 h-10 text-amber-400 mx-auto" />
        <div className="space-y-2">
          <h2 className="text-xl font-bold text-white tracking-tight">
            {t("examReview.notCompletedTitle")}
          </h2>
          <p className="text-base text-graphite-300 leading-relaxed">
            {t("examReview.notCompletedDesc")}
          </p>
        </div>
        <button
          type="button"
          onClick={() => navigate(resumePath)}
          className="w-full py-2.5 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 transition-colors cursor-pointer"
        >
          {resumeLabel}
        </button>
      </div>
    );
  }

  if (errorState?.kind === "api" || !review) {
    const msg = resolveApiError(errorState?.err, t, t("errors.EXAM_SESSION_NOT_FOUND"));
    return (
      <div className="max-w-lg mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
        <ShieldAlert className="w-10 h-10 text-red-400 mx-auto" />
        <div className="space-y-2">
          <h2 className="text-xl font-bold text-white tracking-tight">
            {t("examReview.loadErrorTitle")}
          </h2>
          <p className="text-base text-graphite-300 leading-relaxed">{msg}</p>
        </div>
        <button
          type="button"
          onClick={() => navigate("/evaluation")}
          className="w-full py-2.5 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 transition-colors cursor-pointer"
        >
          {t("examReview.backToEvaluation")}
        </button>
      </div>
    );
  }

  const { session, domain_scores, pre_comparison, results } = review;
  const isPostTest = session.exam_type === "post";

  const breakdownDomains = domain_scores.map((ds) => ({
    domain_id: ds.domain_id,
    domain_name: ds.domain_name,
    proficiency_score: ds.percentage,
    threshold: ds.threshold ?? 70,
    passed: ds.passed,
    score_meta: `${ds.correct}/${ds.total}`,
  }));

  return (
    <div className="max-w-6xl mx-auto space-y-8 animate-fade-in">
      {/* Header & Summary Card */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 sm:p-8 space-y-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Link
            to="/evaluation"
            className="inline-flex items-center gap-1.5 text-sm text-graphite-400 hover:text-white transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            {t("examReview.backToEvaluation")}
          </Link>

          <div className="flex flex-wrap items-center gap-2.5 text-xs font-mono text-graphite-400">
            <span
              className={cn(
                "px-2.5 py-0.5 rounded font-semibold border",
                getExamTypeBadgeClass(session.exam_type),
              )}
            >
              {isPostTest ? t("examReview.postBadge") : t("examReview.preBadge")}
            </span>
            {session.completed_at && (
              <span>
                {t("examReview.completedAtLabel", {
                  date: formatDate(session.completed_at),
                })}
              </span>
            )}
          </div>
        </div>

        <div className="space-y-2">
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white">
            {isPostTest ? t("exam.postTest.title") : t("exam.preTest.title")}
          </h1>
          <p className="text-base text-graphite-300 leading-relaxed max-w-4xl">
            {isPostTest
              ? t("examReview.postFramingSubtitle")
              : t("examReview.preFramingSubtitle")}
          </p>
        </div>

        <div className="grid grid-cols-2 lg:grid-cols-5 gap-3">
          <div className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4">
            <div className="text-xs font-mono text-graphite-400 uppercase">
              {t("examReview.statScore")}
            </div>
            <div className="text-2xl font-bold font-mono text-white mt-1">
              {session.score.toFixed(1)}%
            </div>
          </div>
          <div className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4">
            <div className="text-xs font-mono text-graphite-400 uppercase">
              {t("examReview.statCorrect")}
            </div>
            <div className="text-2xl font-bold font-mono text-emerald-400 mt-1">
              {session.correct_count}/{session.total_questions}
            </div>
          </div>
          <div className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4">
            <div className="text-xs font-mono text-graphite-400 uppercase">
              {t("examReview.statIncorrect")}
            </div>
            <div className="text-2xl font-bold font-mono text-red-400 mt-1">
              {session.incorrect_count}
            </div>
          </div>
          <div className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4">
            <div className="text-xs font-mono text-graphite-400 uppercase">
              {t("examReview.statUnanswered")}
            </div>
            <div className="text-2xl font-bold font-mono text-amber-300 mt-1">
              {session.unanswered_count}
            </div>
          </div>
          <div className="col-span-2 lg:col-span-1 rounded-md border border-graphite-800 bg-graphite-950/60 p-4">
            <div className="text-xs font-mono text-graphite-400 uppercase flex items-center gap-1.5">
              <Clock className="w-3.5 h-3.5" />
              {t("examReview.statTimeSpent")}
            </div>
            <div className="text-2xl font-bold font-mono text-white mt-1">
              {formatDurationSeconds(session.time_spent_seconds)}
            </div>
          </div>
        </div>

        {locale === "th" && (
          <div className="px-3.5 py-2.5 rounded bg-graphite-950/80 border border-graphite-800 text-graphite-300 text-xs">
            {t("exam.englishOnlyNotice")}
          </div>
        )}
      </div>

      {/* Post-Test Pre->Post Learning Gain Card */}
      {isPostTest && pre_comparison && (
        <PrePostGainCard
          comparison={pre_comparison}
          totalQuestions={session.total_questions}
        />
      )}

      {/* Non-blocking Post-Test Evaluation Questionnaire Banner */}
      {isPostTest && questionnaireStatus && (
        <QuestionnairePromptBanner status={questionnaireStatus} variant="prominent" />
      )}

      {/* Domain Breakdown + 01-30 Jump Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <DomainBreakdown
            domains={breakdownDomains}
            title={t("examReview.domainBreakdownTitle")}
            studyDomainHref={(id) => `/learn?domain=${encodeURIComponent(id)}`}
          />
        </div>

        <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 flex flex-col justify-between space-y-4">
          <div className="space-y-2">
            <h3 className="text-lg font-semibold tracking-tight text-white border-b border-graphite-800 pb-3">
              {t("examReview.jumpGridTitle")}
            </h3>
            <p className="text-xs text-graphite-400 leading-relaxed">
              {t("examReview.jumpGridHint")}
            </p>
          </div>

          <div className="grid grid-cols-6 sm:grid-cols-5 gap-2">
            {results.map((q) => {
              const itemOutcome = getReviewItemOutcome(q);
              const colorClass =
                itemOutcome === "correct"
                  ? "bg-emerald-500/10 border-emerald-500/35 text-emerald-300 hover:bg-emerald-500/20"
                  : itemOutcome === "unanswered"
                    ? "bg-amber-500/10 border-amber-500/35 border-dashed text-amber-300 hover:bg-amber-500/20"
                    : "bg-red-500/10 border-red-500/35 text-red-300 hover:bg-red-500/20";

              return (
                <button
                  key={q.question_id}
                  type="button"
                  onClick={() => handleJumpToQuestion(q)}
                  className={cn(
                    "h-10 rounded text-xs font-mono font-semibold border transition-colors flex items-center justify-center cursor-pointer",
                    colorClass,
                  )}
                >
                  {(q.order_index + 1).toString().padStart(2, "0")}
                </button>
              );
            })}
          </div>

          <div className="flex flex-wrap items-center gap-3 pt-2 border-t border-graphite-800 text-xs text-graphite-400">
            <span className="inline-flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-sm bg-emerald-500/60" />
              {t("examReview.legendCorrect")}
            </span>
            <span className="inline-flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-sm bg-red-500/60" />
              {t("examReview.legendIncorrect")}
            </span>
            <span className="inline-flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-sm bg-amber-400/70" />
              {t("examReview.legendUnanswered")}
            </span>
          </div>
        </div>
      </div>

      {/* Filter & Alignment Toolbar */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-4 flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={() => setStatusFilter("all")}
            className={cn(
              "px-3 py-1.5 rounded-md text-sm font-mono font-medium border transition-colors cursor-pointer",
              statusFilter === "all"
                ? "bg-white text-black border-white font-semibold"
                : "bg-graphite-950 text-graphite-300 border-graphite-800 hover:border-graphite-700",
            )}
          >
            {t("examReview.filterAll", { count: session.total_questions })}
          </button>
          <button
            type="button"
            onClick={() => setStatusFilter("incorrect")}
            className={cn(
              "px-3 py-1.5 rounded-md text-sm font-mono font-medium border transition-colors cursor-pointer",
              statusFilter === "incorrect"
                ? "bg-red-500/20 text-red-200 border-red-400 font-semibold"
                : "bg-graphite-950 text-graphite-300 border-graphite-800 hover:border-graphite-700",
            )}
          >
            {t("examReview.filterIncorrect", { count: session.incorrect_count })}
          </button>
          {session.unanswered_count > 0 && (
            <button
              type="button"
              onClick={() => setStatusFilter("unanswered")}
              className={cn(
                "px-3 py-1.5 rounded-md text-sm font-mono font-medium border transition-colors cursor-pointer",
                statusFilter === "unanswered"
                  ? "bg-amber-500/20 text-amber-200 border-amber-400 font-semibold"
                  : "bg-graphite-950 text-graphite-300 border-graphite-800 hover:border-graphite-700",
              )}
            >
              {t("examReview.filterUnanswered", { count: session.unanswered_count })}
            </button>
          )}
          <button
            type="button"
            onClick={() => setStatusFilter("correct")}
            className={cn(
              "px-3 py-1.5 rounded-md text-sm font-mono font-medium border transition-colors cursor-pointer",
              statusFilter === "correct"
                ? "bg-emerald-500/20 text-emerald-200 border-emerald-400 font-semibold"
                : "bg-graphite-950 text-graphite-300 border-graphite-800 hover:border-graphite-700",
            )}
          >
            {t("examReview.filterCorrect", { count: session.correct_count })}
          </button>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <select
            value={domainFilter}
            onChange={(e) => setDomainFilter(e.target.value)}
            aria-label={t("examReview.allDomains")}
            className="px-3 py-1.5 rounded-md border border-graphite-800 bg-graphite-950 text-sm text-graphite-200 focus:outline-none focus:border-white cursor-pointer"
          >
            <option value="all">{t("examReview.allDomains")}</option>
            {domain_scores.map((ds) => (
              <option key={ds.domain_id} value={ds.domain_id}>
                {resolveDomainName(ds.domain_id, ds.domain_name, locale)}
              </option>
            ))}
          </select>

          <div
            role="group"
            className="inline-flex items-center rounded-md border border-graphite-800 bg-graphite-950 p-0.5"
          >
            <button
              type="button"
              onClick={() => setGroupMode("domain")}
              className={cn(
                "inline-flex items-center gap-1.5 px-3 py-1.5 rounded text-sm font-medium transition-colors cursor-pointer",
                groupMode === "domain"
                  ? "bg-white text-black font-semibold"
                  : "text-graphite-400 hover:text-white",
              )}
            >
              <Layers className="w-3.5 h-3.5" />
              {t("examReview.groupByDomain")}
            </button>
            <button
              type="button"
              onClick={() => setGroupMode("order")}
              className={cn(
                "inline-flex items-center gap-1.5 px-3 py-1.5 rounded text-sm font-medium transition-colors cursor-pointer",
                groupMode === "order"
                  ? "bg-white text-black font-semibold"
                  : "text-graphite-400 hover:text-white",
              )}
            >
              <ListOrdered className="w-3.5 h-3.5" />
              {t("examReview.groupByOrder")}
            </button>
          </div>

          <button
            type="button"
            onClick={handleToggleExpandAll}
            disabled={visibleItems.length === 0}
            className="px-3 py-1.5 rounded-md border border-graphite-800 bg-graphite-950 text-sm font-medium text-graphite-300 hover:text-white hover:border-graphite-700 disabled:opacity-50 cursor-pointer"
          >
            {allVisibleExpanded ? t("examReview.collapseAll") : t("examReview.expandAll")}
          </button>
        </div>
      </div>

      {/* Question Cards List */}
      {sections.length === 0 ? (
        <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-8 text-center space-y-4">
          <p className="text-base text-graphite-300">{t("examReview.emptyFiltered")}</p>
          <button
            type="button"
            onClick={() => {
              setStatusFilter("all");
              setDomainFilter("all");
            }}
            className="px-4 py-2 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 cursor-pointer"
          >
            {t("examReview.resetFilters")}
          </button>
        </div>
      ) : (
        <div className="space-y-8">
          {sections.map((section) => (
            <section key={section.key} className="space-y-4">
              {groupMode === "domain" && section.domainId && (
                <div className="flex flex-wrap items-center justify-between gap-3 border-b border-graphite-800 pb-3">
                  <div className="flex flex-wrap items-center gap-2.5">
                    <h2 className="text-lg font-semibold text-white">
                      {resolveDomainName(
                        section.domainId,
                        section.domainScore?.domain_name ?? section.domainId,
                        locale,
                      )}
                    </h2>
                    {section.domainScore && (
                      <>
                        <span className="px-2.5 py-0.5 rounded text-xs font-mono border border-graphite-800 bg-graphite-900 text-graphite-300">
                          {t("examReview.domainGroupScore", {
                            correct: section.domainScore.correct,
                            total: section.domainScore.total,
                            percentage: section.domainScore.percentage.toFixed(0),
                          })}
                        </span>
                        {!section.domainScore.passed && (
                          <span className="px-2.5 py-0.5 rounded text-xs font-mono font-semibold bg-amber-500/10 text-amber-300 border border-amber-500/30">
                            {t("examReview.needsFocusBadge", {
                              threshold: section.domainScore.threshold,
                            })}
                          </span>
                        )}
                      </>
                    )}
                  </div>

                  <Link
                    to={`/learn?domain=${encodeURIComponent(section.domainId)}`}
                    className="inline-flex items-center gap-1.5 text-sm font-medium text-emerald-400 hover:text-emerald-300 transition-colors"
                  >
                    <BookOpen className="w-4 h-4" />
                    {t("examReview.studyDomainLink")}
                    <ArrowRight className="w-4 h-4" />
                  </Link>
                </div>
              )}

              <div className="space-y-4">
                {section.items.map((item) => (
                  <ReviewQuestionCard
                    key={item.question_id}
                    item={item}
                    isExpanded={expandedIds.has(item.question_id)}
                    isHighlighted={highlightedOrderIdx === item.order_index}
                    showDomainBadge={groupMode === "order"}
                    onToggleExpand={() => handleToggleCard(item.question_id)}
                    cardRef={(el) => {
                      cardRefs.current[item.order_index] = el;
                    }}
                  />
                ))}
              </div>
            </section>
          ))}
        </div>
      )}

      {/* Bottom Next-Step CTAs */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-3 text-sm text-graphite-300">
          <Award className="w-5 h-5 text-emerald-400 shrink-0" />
          <span>
            {isPostTest
              ? t("examReview.postFramingSubtitle")
              : t("examReview.preFramingSubtitle")}
          </span>
        </div>

        <div className="flex flex-wrap items-center gap-3 shrink-0">
          {isPostTest && questionnaireStatus?.eligible && !questionnaireStatus?.submitted && (
            <Link
              to="/questionnaire"
              className="px-4 py-2.5 rounded-md border border-emerald-500/40 bg-emerald-500/15 text-sm font-semibold text-emerald-300 hover:bg-emerald-500/25 transition-colors inline-flex items-center gap-2"
            >
              <span>{t("questionnaire.banner.startAction")}</span>
              <ArrowRight className="w-4 h-4" />
            </Link>
          )}
          <button
            type="button"
            onClick={() => navigate(isPostTest ? "/profile" : "/learn")}
            className="px-4 py-2.5 rounded-md border border-graphite-700 bg-graphite-950 text-sm font-medium text-graphite-200 hover:text-white hover:border-graphite-600 transition-colors cursor-pointer"
          >
            {isPostTest ? t("examReview.postCtaSecondary") : t("examReview.preCtaSecondary")}
          </button>
          <button
            type="button"
            onClick={() => navigate("/evaluation")}
            className="px-5 py-2.5 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 transition-colors inline-flex items-center gap-2 cursor-pointer"
          >
            <span>
              {isPostTest ? t("examReview.postCtaPrimary") : t("examReview.preCtaPrimary")}
            </span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
