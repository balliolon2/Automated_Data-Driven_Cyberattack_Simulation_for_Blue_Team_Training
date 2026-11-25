import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import axios from "axios";
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  ClipboardCheck,
  Eye,
  Gauge,
  Loader2,
  Lock,
  MessageSquareHeart,
  SlidersHorizontal,
} from "lucide-react";
import QualitativeFeedbackList from "../components/QualitativeFeedbackList";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import {
  CSUQ_ITEMS,
  MAX_QUALITATIVE_CHARS,
  NASA_TLX_DIMENSIONS,
  canAdvanceCSUQStep,
  canAdvanceTLXStep,
  computeCSUQScoresPreview,
  computeNASATLXPreview,
  countCSUQAnswered,
  countQualitativeChars,
  getQuestionnaireDraftStorageKey,
  isQualitativeStepValid,
  parseQuestionnaireDraft,
  serializeQuestionnaireDraft,
  type CSUQAnswerValue,
  type CSUQSubscale,
  type NASATLXKey,
  type QuestionnaireDraft,
  type QuestionnaireStatusResponse,
} from "../lib/questionnaire";
import { cn } from "../lib/utils";

const CSUQ_SCALE_VALUES = [1, 2, 3, 4, 5, 6, 7] as const;

export default function QuestionnairePage() {
  const navigate = useNavigate();
  const { locale, t, formatDate } = useI18n();

  const [status, setStatus] = useState<QuestionnaireStatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [submitError, setSubmitError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const userId = localStorage.getItem("userId") || "";
  const draftKey = useMemo(() => getQuestionnaireDraftStorageKey(userId), [userId]);

  const [draft, setDraft] = useState<QuestionnaireDraft>(() =>
    parseQuestionnaireDraft(localStorage.getItem(getQuestionnaireDraftStorageKey(localStorage.getItem("userId") || ""))),
  );
  const [stepValidationMsg, setStepValidationMsg] = useState("");

  useEffect(() => {
    let cancelled = false;

    async function fetchStatus() {
      const token = localStorage.getItem("token");
      if (!token) {
        navigate("/login");
        return;
      }

      try {
        const res = await axios.get<QuestionnaireStatusResponse>("/api/evaluation/questionnaire/status", {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!cancelled) {
          setStatus(res.data);
        }
      } catch (err: unknown) {
        if (axios.isAxiosError(err) && err.response?.status === 401) {
          navigate("/login");
          return;
        }
        if (!cancelled) {
          setLoadError(resolveApiError(err, t, t("questionnaire.loadError")));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void fetchStatus();
    return () => {
      cancelled = true;
    };
  }, [navigate, t]);

  // Persist in-progress answers cleanly across refreshes and EN/TH toggles.
  useEffect(() => {
    if (status?.submitted) return;
    try {
      localStorage.setItem(draftKey, serializeQuestionnaireDraft(draft));
    } catch {
      // Ignore storage quota errors cleanly
    }
  }, [draft, draftKey, status?.submitted]);

  const { answered: csuqAnsweredCount, numericCount: csuqNumericCount } = useMemo(
    () => countCSUQAnswered(draft.csuq),
    [draft.csuq],
  );
  const csuqReady = useMemo(() => canAdvanceCSUQStep(draft.csuq), [draft.csuq]);
  const tlxReady = useMemo(() => canAdvanceTLXStep(draft.tlxConfirmed), [draft.tlxConfirmed]);
  const tlxConfirmedCount = useMemo(
    () => NASA_TLX_DIMENSIONS.filter((d) => draft.tlxConfirmed[d.key]).length,
    [draft.tlxConfirmed],
  );
  const qualitativeValid = useMemo(() => isQualitativeStepValid(draft.qualitative), [draft.qualitative]);

  const csuqPreview = useMemo(() => computeCSUQScoresPreview(draft.csuq), [draft.csuq]);
  const tlxPreview = useMemo(() => computeNASATLXPreview(draft.tlx), [draft.tlx]);

  const setCSUQAnswer = (key: string, value: CSUQAnswerValue) => {
    setStepValidationMsg("");
    setSubmitError("");
    setDraft((prev) => ({
      ...prev,
      csuq: {
        ...prev.csuq,
        [key]: value,
      },
    }));
  };

  const setTLXValue = (key: NASATLXKey, value: number) => {
    setStepValidationMsg("");
    setSubmitError("");
    setDraft((prev) => ({
      ...prev,
      tlx: {
        ...prev.tlx,
        [key]: value,
      },
      tlxConfirmed: {
        ...prev.tlxConfirmed,
        [key]: true,
      },
    }));
  };

  const toggleTLXConfirmed = (key: NASATLXKey) => {
    setStepValidationMsg("");
    setDraft((prev) => ({
      ...prev,
      tlxConfirmed: {
        ...prev.tlxConfirmed,
        [key]: !prev.tlxConfirmed[key],
      },
    }));
  };

  const handleConfirmAllTLX = () => {
    setStepValidationMsg("");
    setDraft((prev) => ({
      ...prev,
      tlxConfirmed: {
        mental_demand: true,
        physical_demand: true,
        temporal_demand: true,
        performance_rating: true,
        effort: true,
        frustration: true,
      },
    }));
  };

  const handleNextFromStep1 = () => {
    if (csuqAnsweredCount < CSUQ_ITEMS.length) {
      setStepValidationMsg(
        t("questionnaire.validation.csuqIncomplete", {
          remaining: CSUQ_ITEMS.length - csuqAnsweredCount,
        }),
      );
      return;
    }
    if (csuqNumericCount < 1) {
      setStepValidationMsg(t("questionnaire.validation.csuqAllNA"));
      return;
    }
    setStepValidationMsg("");
    setDraft((prev) => ({ ...prev, step: 2 }));
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const handleNextFromStep2 = () => {
    if (!tlxReady) {
      setStepValidationMsg(
        t("questionnaire.validation.tlxIncomplete", {
          remaining: NASA_TLX_DIMENSIONS.length - tlxConfirmedCount,
        }),
      );
      return;
    }
    setStepValidationMsg("");
    setDraft((prev) => ({ ...prev, step: 3 }));
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const handleStepTabClick = (targetStep: 1 | 2 | 3) => {
    if (status?.is_preview_only || targetStep <= draft.step) {
      setStepValidationMsg("");
      setDraft((prev) => ({ ...prev, step: targetStep }));
      return;
    }
    if (targetStep === 2) {
      handleNextFromStep1();
      return;
    }
    if (!csuqReady) {
      handleNextFromStep1();
      return;
    }
    handleNextFromStep2();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!status || status.is_preview_only || submitting) return;
    if (!csuqReady) {
      setDraft((prev) => ({ ...prev, step: 1 }));
      setStepValidationMsg(t("questionnaire.validation.csuqIncomplete", { remaining: CSUQ_ITEMS.length - csuqAnsweredCount }));
      return;
    }
    if (!tlxReady) {
      setDraft((prev) => ({ ...prev, step: 2 }));
      setStepValidationMsg(t("questionnaire.validation.tlxIncomplete", { remaining: NASA_TLX_DIMENSIONS.length - tlxConfirmedCount }));
      return;
    }
    if (!qualitativeValid) {
      setStepValidationMsg(t("questionnaire.validation.feedbackTooLong"));
      return;
    }

    const token = localStorage.getItem("token");
    if (!token) {
      navigate("/login");
      return;
    }

    setSubmitting(true);
    setSubmitError("");
    try {
      const csuqPayload: Record<string, number | null> = {};
      for (const item of CSUQ_ITEMS) {
        const val = draft.csuq[item.key];
        csuqPayload[item.key] = val === undefined ? null : val;
      }

      const res = await axios.post<{ message: string; submission: QuestionnaireStatusResponse["submission"] }>(
        "/api/evaluation/questionnaire",
        {
          locale,
          csuq: csuqPayload,
          nasa_tlx: draft.tlx,
          qualitative_feedback: {
            most_impressive: draft.qualitative.most_impressive.trim(),
            areas_to_improve: draft.qualitative.areas_to_improve.trim(),
            additional_comments: draft.qualitative.additional_comments.trim(),
          },
        },
        { headers: { Authorization: `Bearer ${token}` } },
      );

      try {
        localStorage.removeItem(draftKey);
      } catch {
        // Ignore storage errors
      }

      setStatus((prev) => ({
        role: prev?.role ?? "learner",
        is_preview_only: false,
        has_completed_post_test: true,
        latest_post_session_id: prev?.latest_post_session_id ?? null,
        eligible: false,
        ineligible_reason: "already_submitted",
        submitted: true,
        submission: res.data.submission,
      }));
      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err: unknown) {
      setSubmitError(resolveApiError(err, t, t("questionnaire.submitError")));
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-sm font-mono text-graphite-400">{t("questionnaire.loading")}</p>
      </div>
    );
  }

  if (loadError || !status) {
    return (
      <div className="max-w-lg mx-auto mt-12 bg-graphite-900 border border-graphite-800 rounded-lg p-8 text-center space-y-5 animate-fade-in">
        <Lock className="w-10 h-10 text-red-400 mx-auto" />
        <h2 className="text-xl font-bold text-white">{t("questionnaire.unavailableTitle")}</h2>
        <p className="text-base text-graphite-300 leading-relaxed">{loadError || t("questionnaire.loadError")}</p>
        <Link
          to="/evaluation"
          className="inline-flex items-center justify-center px-5 py-2.5 rounded-md bg-white text-black font-semibold text-sm hover:bg-white/90 transition-colors"
        >
          {t("questionnaire.backToEvaluation")}
        </Link>
      </div>
    );
  }

  // Locked state for learners who have not finished Post-Test
  if (!status.is_preview_only && !status.submitted && !status.eligible) {
    return (
      <div className="max-w-xl mx-auto mt-12 bg-graphite-900 border border-graphite-800 rounded-lg p-8 text-center space-y-6 animate-fade-in">
        <div className="w-12 h-12 rounded-lg bg-amber-500/10 border border-amber-500/30 flex items-center justify-center mx-auto">
          <Lock className="w-6 h-6 text-amber-400" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-white tracking-tight">
            {t("questionnaire.lockedTitle")}
          </h1>
          <p className="text-base text-graphite-300 leading-relaxed">
            {t("questionnaire.lockedDesc")}
          </p>
        </div>
        <div className="flex flex-col sm:flex-row gap-3 justify-center pt-2">
          <Link
            to="/evaluation"
            className="px-5 py-2.5 rounded-md border border-graphite-700 bg-graphite-950 text-sm font-medium text-graphite-200 hover:text-white transition-colors"
          >
            {t("questionnaire.backToEvaluation")}
          </Link>
          <Link
            to="/post-test"
            className="px-5 py-2.5 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 transition-colors inline-flex items-center justify-center gap-2"
          >
            <span>{t("questionnaire.goToPostTest")}</span>
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    );
  }

  // Read-only confirmation view after submission
  if (status.submitted && status.submission) {
    const sub = status.submission;
    const subscales = sub.csuq_scores?.subscales;
    const tlx = sub.nasa_tlx_scores;

    return (
      <div className="max-w-4xl mx-auto space-y-8 animate-fade-in">
        <div className="bg-graphite-900 border border-emerald-500/30 rounded-lg p-6 sm:p-8 space-y-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <span className="inline-flex items-center gap-2 px-3 py-1 rounded text-xs font-mono font-semibold bg-emerald-500/15 text-emerald-300 border border-emerald-500/30">
              <CheckCircle2 className="w-4 h-4" />
              {t("questionnaire.submittedView.badge")}
            </span>
            <span className="text-xs font-mono text-graphite-400">
              {t("questionnaire.submittedView.submittedAt", { date: formatDate(sub.eval_date) })}
            </span>
          </div>

          <div className="space-y-2">
            <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              {t("questionnaire.submittedView.title")}
            </h1>
            <p className="text-base text-graphite-300 leading-relaxed">
              {t("questionnaire.submittedView.subtitle")}
            </p>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2">
            <div className="rounded-md border border-graphite-800 bg-graphite-950/70 p-4">
              <div className="text-xs font-mono text-graphite-400">{t("questionnaire.submittedView.preScore")}</div>
              <div className="text-xl font-bold font-mono text-white mt-1">{sub.pre_score.toFixed(1)}%</div>
            </div>
            <div className="rounded-md border border-graphite-800 bg-graphite-950/70 p-4">
              <div className="text-xs font-mono text-graphite-400">{t("questionnaire.submittedView.postScore")}</div>
              <div className="text-xl font-bold font-mono text-white mt-1">{sub.post_score.toFixed(1)}%</div>
            </div>
            <div className="rounded-md border border-graphite-800 bg-graphite-950/70 p-4">
              <div className="text-xs font-mono text-graphite-400">{t("questionnaire.submittedView.csuqOverall")}</div>
              <div className="text-xl font-bold font-mono text-emerald-400 mt-1">
                {subscales?.overall !== null && subscales?.overall !== undefined ? `${subscales.overall.toFixed(2)} / 7` : "N/A"}
              </div>
            </div>
            <div className="rounded-md border border-graphite-800 bg-graphite-950/70 p-4">
              <div className="text-xs font-mono text-graphite-400">{t("questionnaire.submittedView.tlxOverall")}</div>
              <div className="text-xl font-bold font-mono text-amber-300 mt-1">
                {tlx ? `${tlx.overall_workload.toFixed(1)} / 100` : "N/A"}
              </div>
            </div>
          </div>
        </div>

        {/* Subscale + Workload Summary */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-4">
            <h2 className="text-lg font-semibold text-white border-b border-graphite-800 pb-3">
              {t("questionnaire.submittedView.csuqCardTitle")}
            </h2>
            <div className="space-y-3 text-sm font-mono">
              {[
                { label: t("questionnaire.step1.groups.SYSUSE"), value: subscales?.sysuse },
                { label: t("questionnaire.step1.groups.INFOQUAL"), value: subscales?.infoqual },
                { label: t("questionnaire.step1.groups.INTERQUAL"), value: subscales?.interqual },
                { label: t("questionnaire.step1.groups.OVERALL"), value: subscales?.overall },
              ].map((row) => (
                <div key={row.label} className="flex items-center justify-between py-1.5 border-b border-graphite-800/60">
                  <span className="text-graphite-300 font-sans">{row.label}</span>
                  <span className="text-white font-semibold">
                    {row.value !== null && row.value !== undefined ? `${row.value.toFixed(2)} / 7.00` : "N/A"}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-4">
            <h2 className="text-lg font-semibold text-white border-b border-graphite-800 pb-3">
              {t("questionnaire.submittedView.tlxCardTitle")}
            </h2>
            <div className="space-y-2.5 text-sm">
              {NASA_TLX_DIMENSIONS.map((dim) => {
                const rawVal = tlx?.dimensions?.[dim.key] ?? 0;
                const workloadVal = dim.invertForWorkload
                  ? tlx?.dimensions?.performance_workload ?? 100 - rawVal
                  : rawVal;
                return (
                  <div key={dim.key} className="space-y-1">
                    <div className="flex items-center justify-between">
                      <span className="text-graphite-300">
                        {locale === "th" ? dim.th.label : dim.en.label}
                      </span>
                      <span className="font-mono text-white font-semibold">
                        {dim.invertForWorkload
                          ? t("questionnaire.submittedView.performancePair", {
                              rating: rawVal,
                              workload: workloadVal,
                            })
                          : `${workloadVal} / 100`}
                      </span>
                    </div>
                    <div className="h-1.5 w-full bg-graphite-950 rounded-full overflow-hidden border border-graphite-800">
                      <div
                        className="h-full bg-emerald-500/70 rounded-full"
                        style={{ width: `${Math.min(100, Math.max(0, workloadVal))}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Qualitative Notes */}
        <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-4">
          <h2 className="text-lg font-semibold text-white border-b border-graphite-800 pb-3">
            {t("questionnaire.submittedView.feedbackCardTitle")}
          </h2>
          <QualitativeFeedbackList
            feedback={sub.qualitative_feedback}
            hasNotes={Boolean(sub.notes)}
            labels={{
              mostImpressive: t("questionnaire.step3.mostImpressiveLabel"),
              areasToImprove: t("questionnaire.step3.areasToImproveLabel"),
              additionalComments: t("questionnaire.step3.additionalCommentsLabel"),
              emptyText: t("questionnaire.submittedView.noFeedback"),
            }}
          />
        </div>

        <div className="flex flex-wrap items-center justify-end gap-3">
          <Link
            to="/profile"
            className="px-4 py-2.5 rounded-md border border-graphite-700 bg-graphite-950 text-sm font-medium text-graphite-200 hover:text-white transition-colors"
          >
            {t("questionnaire.backToProfile")}
          </Link>
          <Link
            to="/evaluation"
            className="px-5 py-2.5 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 transition-colors inline-flex items-center gap-2"
          >
            <span>{t("questionnaire.backToEvaluation")}</span>
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    );
  }

  // Group CSUQ items for Step 1
  const csuqGroups: Array<{ key: CSUQSubscale; items: typeof CSUQ_ITEMS[number][] }> = [
    { key: "SYSUSE", items: CSUQ_ITEMS.filter((i) => i.subscale === "SYSUSE") },
    { key: "INFOQUAL", items: CSUQ_ITEMS.filter((i) => i.subscale === "INFOQUAL") },
    {
      key: "INTERQUAL",
      items: CSUQ_ITEMS.filter((i) => i.subscale === "INTERQUAL" || i.subscale === "OVERALL"),
    },
  ];

  return (
    <div className="max-w-4xl mx-auto space-y-6 animate-fade-in">
      {/* Header */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 sm:p-8 space-y-5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Link
            to="/evaluation"
            className="inline-flex items-center gap-1.5 text-sm text-graphite-400 hover:text-white transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            {t("questionnaire.backToEvaluation")}
          </Link>
          <span className="px-2.5 py-1 rounded text-xs font-mono font-semibold border border-emerald-500/30 bg-emerald-500/10 text-emerald-300">
            {t("questionnaire.badge")}
          </span>
        </div>

        <div className="space-y-2">
          <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
            {t("questionnaire.title")}
          </h1>
          <p className="text-base text-graphite-300 leading-relaxed">
            {t("questionnaire.subtitle")}
          </p>
        </div>

        {status.is_preview_only && (
          <div className="rounded-md border border-amber-500/40 bg-amber-500/10 p-4 flex items-start gap-3">
            <Eye className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
            <div className="space-y-1">
              <div className="text-sm font-semibold text-amber-300">
                {t("questionnaire.previewBannerTitle")}
              </div>
              <p className="text-sm text-graphite-200 leading-relaxed">
                {t("questionnaire.previewBannerDesc")}
              </p>
            </div>
          </div>
        )}

        {/* 3-Step Guided Stepper */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2">
          {[
            {
              step: 1 as const,
              icon: ClipboardCheck,
              title: t("questionnaire.steps.step1Title"),
              meta: t("questionnaire.steps.step1Meta", { answered: csuqAnsweredCount, total: CSUQ_ITEMS.length }),
            },
            {
              step: 2 as const,
              icon: SlidersHorizontal,
              title: t("questionnaire.steps.step2Title"),
              meta: t("questionnaire.steps.step2Meta", { confirmed: tlxConfirmedCount, total: NASA_TLX_DIMENSIONS.length }),
            },
            {
              step: 3 as const,
              icon: MessageSquareHeart,
              title: t("questionnaire.steps.step3Title"),
              meta: t("questionnaire.steps.step3Meta"),
            },
          ].map((s) => {
            const Icon = s.icon;
            const active = draft.step === s.step;
            const completed =
              (s.step === 1 && csuqReady) ||
              (s.step === 2 && tlxReady) ||
              (s.step === 3 && status.submitted);
            return (
              <button
                key={s.step}
                type="button"
                onClick={() => handleStepTabClick(s.step)}
                className={cn(
                  "text-left rounded-md border p-3.5 transition-colors cursor-pointer",
                  active
                    ? "border-white bg-graphite-950 text-white"
                    : completed
                      ? "border-emerald-500/35 bg-emerald-500/10 text-emerald-200"
                      : "border-graphite-800 bg-graphite-950/60 text-graphite-300 hover:border-graphite-700",
                )}
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="inline-flex items-center gap-2 text-sm font-semibold">
                    <Icon className="w-4 h-4" />
                    {s.title}
                  </span>
                  {completed && <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />}
                </div>
                <div className="text-xs font-mono text-graphite-400 mt-1">{s.meta}</div>
              </button>
            );
          })}
        </div>
      </div>

      {stepValidationMsg && (
        <div className="rounded-md border border-amber-500/40 bg-amber-500/10 p-4 text-sm text-amber-200">
          {stepValidationMsg}
        </div>
      )}

      {submitError && (
        <div className="rounded-md border border-red-500/40 bg-red-500/10 p-4 text-sm font-mono text-red-300">
          {submitError}
        </div>
      )}

      {/* Step 1: CSUQ v3 (16 items) */}
      {draft.step === 1 && (
        <div className="space-y-6">
          <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-2">
            <h2 className="text-lg font-semibold text-white">{t("questionnaire.step1.heading")}</h2>
            <p className="text-base text-graphite-300 leading-relaxed">{t("questionnaire.step1.instructions")}</p>
            <div className="flex flex-wrap items-center justify-between gap-2 pt-2 text-xs font-mono text-graphite-400 border-t border-graphite-800">
              <span>1 = {t("questionnaire.step1.scaleLow")}</span>
              <span>7 = {t("questionnaire.step1.scaleHigh")}</span>
              <span>{t("questionnaire.step1.naLegend")}</span>
            </div>
          </div>

          {csuqGroups.map((group) => (
            <div key={group.key} className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 sm:p-6 space-y-5">
              <h3 className="text-base font-semibold text-emerald-400 border-b border-graphite-800 pb-3">
                {group.key === "INTERQUAL"
                  ? t("questionnaire.step1.groups.INTERQUAL_COMBINED")
                  : t(`questionnaire.step1.groups.${group.key}`)}
              </h3>

              <div className="space-y-5">
                {group.items.map((item) => {
                  const current = draft.csuq[item.key];
                  const hasValue = current !== undefined;
                  return (
                    <div
                      key={item.key}
                      className={cn(
                        "rounded-md border p-4 space-y-3 transition-colors",
                        hasValue
                          ? "border-graphite-700 bg-graphite-950/60"
                          : "border-graphite-800 bg-graphite-950/30",
                      )}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <p className="text-base text-graphite-100 leading-relaxed">
                          <span className="font-mono text-sm text-graphite-400 mr-2">
                            Q{item.number.toString().padStart(2, "0")}.
                          </span>
                          {locale === "th" ? item.th : item.en}
                        </p>
                      </div>

                      <div
                        role="radiogroup"
                        aria-label={`Q${item.number}`}
                        className="flex flex-wrap items-center gap-2"
                      >
                        {CSUQ_SCALE_VALUES.map((val) => {
                          const selected = current === val;
                          return (
                            <button
                              key={val}
                              type="button"
                              role="radio"
                              aria-checked={selected}
                              onClick={() => setCSUQAnswer(item.key, val)}
                              className={cn(
                                "min-w-[44px] min-h-[40px] px-3 py-1.5 rounded-md border text-sm font-mono font-semibold transition-colors cursor-pointer",
                                selected
                                  ? "bg-white text-black border-white"
                                  : "bg-graphite-900 text-graphite-300 border-graphite-800 hover:border-graphite-600 hover:text-white",
                              )}
                            >
                              {val}
                            </button>
                          );
                        })}

                        <button
                          type="button"
                          role="radio"
                          aria-checked={current === null}
                          onClick={() => setCSUQAnswer(item.key, null)}
                          className={cn(
                            "min-h-[40px] px-3.5 py-1.5 rounded-md border text-xs font-mono font-semibold transition-colors cursor-pointer ml-auto sm:ml-2",
                            current === null
                              ? "bg-amber-500/20 text-amber-200 border-amber-400"
                              : "bg-graphite-900 text-graphite-400 border-graphite-800 hover:border-graphite-600 hover:text-graphite-200",
                          )}
                        >
                          {t("questionnaire.step1.naButton")}
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          ))}

          <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
            <span className="text-xs font-mono text-graphite-400">
              {t("questionnaire.autoDraftNotice")}
            </span>
            <button
              type="button"
              onClick={handleNextFromStep1}
              className="px-5 py-2.5 rounded-md bg-white text-black font-semibold text-sm hover:bg-white/90 transition-colors inline-flex items-center gap-2 cursor-pointer"
            >
              <span>{t("questionnaire.nextToStep2")}</span>
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Step 2: Raw NASA-TLX (6 dimensions) */}
      {draft.step === 2 && (
        <div className="space-y-6">
          <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="space-y-1.5">
              <h2 className="text-lg font-semibold text-white">{t("questionnaire.step2.heading")}</h2>
              <p className="text-base text-graphite-300 leading-relaxed">{t("questionnaire.step2.instructions")}</p>
            </div>
            <button
              type="button"
              onClick={handleConfirmAllTLX}
              className="px-3.5 py-2 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:border-white hover:text-white transition-colors whitespace-nowrap shrink-0 cursor-pointer"
            >
              {t("questionnaire.step2.confirmAllAction")}
            </button>
          </div>

          <div className="space-y-4">
            {NASA_TLX_DIMENSIONS.map((dim) => {
              const copy = locale === "th" ? dim.th : dim.en;
              const val = draft.tlx[dim.key];
              const confirmed = draft.tlxConfirmed[dim.key];

              return (
                <div
                  key={dim.key}
                  className={cn(
                    "bg-graphite-900 border rounded-lg p-5 sm:p-6 space-y-4 transition-colors",
                    confirmed ? "border-emerald-500/35" : "border-graphite-800",
                  )}
                >
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div className="space-y-1">
                      <h3 className="text-base font-semibold text-white">{copy.label}</h3>
                      <p className="text-sm text-graphite-300 leading-relaxed">{copy.prompt}</p>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <span className="px-3 py-1 rounded-md bg-graphite-950 border border-graphite-800 font-mono text-sm font-bold text-white">
                        {val} / 100
                      </span>
                      <button
                        type="button"
                        onClick={() => toggleTLXConfirmed(dim.key)}
                        className={cn(
                          "px-2.5 py-1 rounded text-xs font-mono font-semibold border transition-colors cursor-pointer",
                          confirmed
                            ? "bg-emerald-500/15 text-emerald-300 border-emerald-500/30"
                            : "bg-graphite-950 text-graphite-400 border-graphite-800 hover:text-white",
                        )}
                      >
                        {confirmed ? t("questionnaire.step2.confirmedBadge") : t("questionnaire.step2.confirmBadge")}
                      </button>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <input
                      type="range"
                      min={0}
                      max={100}
                      step={5}
                      value={val}
                      aria-label={copy.label}
                      onChange={(e) => setTLXValue(dim.key, Number(e.target.value))}
                      className="w-full accent-white cursor-pointer"
                    />
                    <div className="flex items-center justify-between text-xs font-mono text-graphite-400">
                      <span>0 · {copy.lowAnchor}</span>
                      <span>50</span>
                      <span>100 · {copy.highAnchor}</span>
                    </div>
                  </div>

                  {dim.invertForWorkload && (
                    <p className="text-xs text-graphite-400 leading-relaxed border-t border-graphite-800 pt-2">
                      {t("questionnaire.step2.performancePolarityHint", {
                        rating: val,
                        workload: 100 - val,
                      })}
                    </p>
                  )}
                </div>
              );
            })}
          </div>

          <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
            <button
              type="button"
              onClick={() => {
                setStepValidationMsg("");
                setDraft((prev) => ({ ...prev, step: 1 }));
              }}
              className="px-4 py-2.5 rounded-md border border-graphite-700 bg-graphite-950 text-sm font-medium text-graphite-200 hover:text-white transition-colors inline-flex items-center gap-2 cursor-pointer"
            >
              <ArrowLeft className="w-4 h-4" />
              <span>{t("questionnaire.backToStep1")}</span>
            </button>
            <button
              type="button"
              onClick={handleNextFromStep2}
              className="px-5 py-2.5 rounded-md bg-white text-black font-semibold text-sm hover:bg-white/90 transition-colors inline-flex items-center gap-2 cursor-pointer"
            >
              <span>{t("questionnaire.nextToStep3")}</span>
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Qualitative Impressions & Final Submit */}
      {draft.step === 3 && (
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 sm:p-6 space-y-5">
            <div className="space-y-1.5 border-b border-graphite-800 pb-3">
              <h2 className="text-lg font-semibold text-white">{t("questionnaire.step3.heading")}</h2>
              <p className="text-base text-graphite-300 leading-relaxed">{t("questionnaire.step3.instructions")}</p>
            </div>

            {[
              {
                key: "most_impressive" as const,
                label: t("questionnaire.step3.mostImpressiveLabel"),
                placeholder: t("questionnaire.step3.mostImpressivePlaceholder"),
              },
              {
                key: "areas_to_improve" as const,
                label: t("questionnaire.step3.areasToImproveLabel"),
                placeholder: t("questionnaire.step3.areasToImprovePlaceholder"),
              },
              {
                key: "additional_comments" as const,
                label: t("questionnaire.step3.additionalCommentsLabel"),
                placeholder: t("questionnaire.step3.additionalCommentsPlaceholder"),
              },
            ].map((field) => {
              const textVal = draft.qualitative[field.key];
              const charCount = countQualitativeChars(textVal);
              const overLimit = charCount > MAX_QUALITATIVE_CHARS;

              return (
                <div key={field.key} className="space-y-1.5">
                  <div className="flex items-center justify-between gap-2">
                    <label className="text-sm font-medium text-graphite-200">{field.label}</label>
                    <span
                      className={cn(
                        "text-xs font-mono",
                        overLimit ? "text-red-400 font-semibold" : "text-graphite-400",
                      )}
                    >
                      {charCount} / {MAX_QUALITATIVE_CHARS}
                    </span>
                  </div>
                  <textarea
                    rows={3}
                    value={textVal}
                    onChange={(e) => {
                      setStepValidationMsg("");
                      setSubmitError("");
                      setDraft((prev) => ({
                        ...prev,
                        qualitative: {
                          ...prev.qualitative,
                          [field.key]: e.target.value,
                        },
                      }));
                    }}
                    placeholder={field.placeholder}
                    className="w-full px-3.5 py-2.5 bg-graphite-950 border border-graphite-800 rounded-md text-base text-white placeholder-graphite-400 leading-relaxed focus:outline-none focus:border-white transition-colors"
                  />
                </div>
              );
            })}
          </div>

          {/* Pre-submit summary preview */}
          <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 sm:p-6 space-y-4">
            <h3 className="text-base font-semibold text-white flex items-center gap-2">
              <Gauge className="w-4 h-4 text-emerald-400" />
              {t("questionnaire.step3.reviewSummaryTitle")}
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-1">
                <div className="text-xs font-mono text-graphite-400">{t("questionnaire.step3.previewCsuqLabel")}</div>
                <div className="text-2xl font-bold font-mono text-emerald-400">
                  {csuqPreview.overall !== null ? `${csuqPreview.overall.toFixed(2)} / 7.00` : "N/A"}
                </div>
                <div className="text-xs font-mono text-graphite-400">
                  {t("questionnaire.step3.previewCsuqItems", { count: csuqPreview.validOverallCount })}
                </div>
              </div>
              <div className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-1">
                <div className="text-xs font-mono text-graphite-400">{t("questionnaire.step3.previewTlxLabel")}</div>
                <div className="text-2xl font-bold font-mono text-amber-300">
                  {tlxPreview.overallWorkload.toFixed(1)} / 100
                </div>
                <div className="text-xs font-mono text-graphite-400">
                  {t("questionnaire.step3.previewTlxNote")}
                </div>
              </div>
            </div>
          </div>

          <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
            <button
              type="button"
              onClick={() => {
                setStepValidationMsg("");
                setDraft((prev) => ({ ...prev, step: 2 }));
              }}
              className="px-4 py-2.5 rounded-md border border-graphite-700 bg-graphite-950 text-sm font-medium text-graphite-200 hover:text-white transition-colors inline-flex items-center gap-2 cursor-pointer"
            >
              <ArrowLeft className="w-4 h-4" />
              <span>{t("questionnaire.backToStep2")}</span>
            </button>

            <button
              type="submit"
              disabled={status.is_preview_only || submitting || !csuqReady || !tlxReady || !qualitativeValid}
              className="px-6 py-2.5 rounded-md bg-white text-black font-semibold text-sm hover:bg-white/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors inline-flex items-center gap-2 cursor-pointer"
            >
              {submitting ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span>{t("questionnaire.submitting")}</span>
                </>
              ) : (
                <>
                  <CheckCircle2 className="w-4 h-4" />
                  <span>
                    {status.is_preview_only
                      ? t("questionnaire.submitPreviewDisabled")
                      : t("questionnaire.submitButton")}
                  </span>
                </>
              )}
            </button>
          </div>
        </form>
      )}
    </div>
  );
}
