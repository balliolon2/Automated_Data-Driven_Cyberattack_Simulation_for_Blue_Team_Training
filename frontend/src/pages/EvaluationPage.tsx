import { useState, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import axios from "axios";
import { AlertTriangle, BookOpen, CheckCircle, ShieldAlert, Award, PlayCircle, Loader2, ClipboardCheck, ArrowRight } from "lucide-react";
import DomainBreakdown from "../components/DomainBreakdown";
import QuestionnairePromptBanner from "../components/QuestionnairePromptBanner";
import type { QuestionnaireStatusResponse } from "../lib/questionnaire";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import { formatSignedDelta, getExamTypeBadgeClass, type ExamHistoryItem, type ExamHistoryResponse } from "../lib/examReview";

interface DomainStatus {
  domain_id: string;
  domain_name: string;
  proficiency_score: number;
  threshold: number;
  passed: boolean;
}

type EvalErrorState =
  | { kind: "status"; data?: { code?: string; error?: string; params?: Record<string, unknown> } }
  | { kind: "api"; err: unknown }
  | null;

export default function EvaluationPage() {
  const { t, formatDate } = useI18n();
  const [loading, setLoading] = useState(true);
  const [domains, setDomains] = useState<DomainStatus[]>([]);
  const [completedScenarios, setCompletedScenarios] = useState(0);
  const [maxScenarios, setMaxScenarios] = useState(15);
  const [needsTraining, setNeedsTraining] = useState(false);
  const [theoryCompleted, setTheoryCompleted] = useState(false);
  const [completedTheoryDomains, setCompletedTheoryDomains] = useState<string[]>([]);
  const [examHistory, setExamHistory] = useState<ExamHistoryResponse | null>(null);
  const [questionnaireStatus, setQuestionnaireStatus] = useState<QuestionnaireStatusResponse | null>(null);
  const [errorState, setErrorState] = useState<EvalErrorState>(null);

  const navigate = useNavigate();

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          navigate("/login");
          return;
        }

        const [statusRes, historyRes, qStatusRes] = await Promise.allSettled([
          axios.get("/api/simulation/status", {
            headers: { Authorization: `Bearer ${token}` }
          }),
          axios.get<ExamHistoryResponse>("/api/exams/history", {
            headers: { Authorization: `Bearer ${token}` }
          }),
          axios.get<QuestionnaireStatusResponse>("/api/evaluation/questionnaire/status", {
            headers: { Authorization: `Bearer ${token}` }
          })
        ]);

        if (historyRes.status === "fulfilled" && historyRes.value.data) {
          setExamHistory(historyRes.value.data);
        }
        if (qStatusRes.status === "fulfilled" && qStatusRes.value.data) {
          setQuestionnaireStatus(qStatusRes.value.data);
        }

        if (statusRes.status === "rejected") {
          throw statusRes.reason;
        }

        const res = statusRes.value;
        if (res.data.error) {
          setErrorState({
            kind: "status",
            data: {
              code: res.data.code,
              error: res.data.error,
              params: res.data.params,
            },
          });
          setLoading(false);
          return;
        }

        setDomains(res.data.domains || []);
        setCompletedScenarios(res.data.completed_scenarios || 0);
        setMaxScenarios(res.data.max_scenarios || 15);
        setNeedsTraining(res.data.needs_training);
        setTheoryCompleted(Boolean(res.data.theory_completed));
        setCompletedTheoryDomains(Array.isArray(res.data.completed_theory_domains) ? res.data.completed_theory_domains : []);
      } catch (err: unknown) {
        console.error("Failed to load evaluation details:", err);
        setErrorState({ kind: "api", err });
      } finally {
        setLoading(false);
      }
    };

    fetchStatus();
  }, [navigate]);

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-sm font-mono text-graphite-400">{t("evaluation.loading")}</p>
      </div>
    );
  }

  if (errorState) {
    let resolvedErrorMessage = t("evaluation.preTestRequiredError");
    if (errorState.kind === "status") {
      const data = errorState.data;
      if (data?.code) {
        resolvedErrorMessage = resolveApiError({ response: { data } }, t, t("evaluation.preTestRequiredError"));
      } else if (data?.error?.toLowerCase().includes("pre-test")) {
        resolvedErrorMessage = t("evaluation.preTestRequiredError");
      } else if (data?.error) {
        resolvedErrorMessage = data.error;
      }
    } else {
      resolvedErrorMessage = resolveApiError(errorState.err, t, t("evaluation.preTestRequiredError"));
    }

    return (
      <div className="max-w-md mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
        <ShieldAlert className="w-10 h-10 text-red-500 mx-auto" />
        <h3 className="text-lg font-bold text-white tracking-tight">{t("evaluation.assessmentRequiredTitle")}</h3>
        <p className="text-base text-graphite-300 font-normal leading-relaxed">{resolvedErrorMessage}</p>
        <button
          onClick={() => navigate("/pre-test")}
          className="w-full py-2.5 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 transition-all duration-200 cursor-pointer"
        >
          {t("evaluation.takePreTest")}
        </button>
      </div>
    );
  }

  const failedDomains = domains.filter((d) => !d.passed);
  const averageScore = domains.reduce((acc, d) => acc + d.proficiency_score, 0) / Math.max(1, domains.length);

  const latestPreExam: ExamHistoryItem | undefined = examHistory?.latest_pre_session_id
    ? examHistory.sessions.find((s) => s.session_id === examHistory.latest_pre_session_id)
    : undefined;
  const latestPostExam: ExamHistoryItem | undefined = examHistory?.latest_post_session_id
    ? examHistory.sessions.find((s) => s.session_id === examHistory.latest_post_session_id)
    : undefined;
  const preToPostDelta = examHistory?.latest_pre_to_post_delta ?? null;

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-fade-in">
      <div className="text-center space-y-2">
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white">{t("evaluation.title")}</h1>
        <p className="text-sm text-graphite-400 font-mono">{t("evaluation.subtitle")}</p>
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        {/* Left Side: Summary Card */}
        <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg flex flex-col items-center justify-center text-center space-y-3">
          <div className="text-5xl font-extrabold text-white font-mono tracking-tighter">
            {averageScore.toFixed(0)}%
          </div>
          <div className="text-xs font-semibold text-graphite-300 uppercase tracking-wider font-mono">
            {t("evaluation.averageProficiency")}
          </div>
          <p className="text-xs text-graphite-400 font-mono">
            {failedDomains.length > 0
              ? t("evaluation.domainsBelowThreshold", { count: failedDomains.length })
              : t("evaluation.allDomainsPassed")}
          </p>
        </div>

        {/* Right Side: Detailed Breakdown List */}
        <div className="md:col-span-2">
          <DomainBreakdown
            domains={domains}
            title={t("evaluation.currentDomainProficiency")}
            studyDomainHref={(id) => `/learn?domain=${id}`}
          />
        </div>
      </div>

      <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
        <h3 className="text-lg font-semibold flex items-center gap-2 text-white border-b border-graphite-800 pb-3">
          <AlertTriangle className="text-amber-500 w-4 h-4" />
          {t("evaluation.skillGapAnalysis")}
        </h3>
        
        <ul className="space-y-4 text-base text-graphite-300 font-normal leading-relaxed">
          {failedDomains.length > 0 ? (
            <>
              <li className="flex gap-3 items-start">
                <ShieldAlert className="text-amber-500 shrink-0 w-5 h-5 mt-0.5" />
                <span>
                  <strong className="text-white">{t("evaluation.identifiedWeaknessLabel")}</strong>{" "}
                  {t("evaluation.identifiedWeaknessBeforeThreshold")}
                  <strong className="text-white">70%</strong>
                  {t("evaluation.identifiedWeaknessAfterThreshold")}
                </span>
              </li>
              <li className="flex gap-3 items-start">
                <CheckCircle className="text-emerald-500 shrink-0 w-5 h-5 mt-0.5" />
                <span>
                  <strong className="text-white">{t("evaluation.adaptiveRoadmapLabel")}</strong> {t("evaluation.adaptiveRoadmapText")}
                </span>
              </li>
              {!theoryCompleted && (
                <li className="flex gap-3 items-start">
                  <BookOpen className="text-amber-500 shrink-0 w-5 h-5 mt-0.5" />
                  <span>
                    <strong className="text-white">{t("evaluation.requiredPreSimLabel")}</strong>{" "}
                    {t("evaluation.requiredPreSimBeforeLink")}
                    <button
                      onClick={() => navigate("/learn")}
                      className="text-white underline hover:text-emerald-400 transition-colors cursor-pointer"
                    >
                      {t("evaluation.learningTheoryHubLink")}
                    </button>
                    {t("evaluation.requiredPreSimAfterLink", { completed: completedTheoryDomains.length })}
                  </span>
                </li>
              )}
            </>
          ) : (
            <li className="flex gap-3 items-start">
              <CheckCircle className="text-emerald-500 shrink-0 w-5 h-5 mt-0.5" />
              <span>
                <strong className="text-white">{t("evaluation.excellentWorkLabel")}</strong>{" "}
                {t("evaluation.excellentWorkBeforeThreshold")}
                <strong className="text-white">70%</strong>
                {t("evaluation.excellentWorkAfterThreshold")}
              </span>
            </li>
          )}
        </ul>

        <div className="pt-4 border-t border-graphite-800 flex justify-end">
          {needsTraining ? (
            !theoryCompleted ? (
              <button
                onClick={() => navigate("/learn")}
                className="px-5 py-2.5 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 transition-all duration-200 flex items-center gap-2 cursor-pointer"
              >
                <BookOpen className="w-4 h-4" />
                {t("evaluation.studyTheoryFirst", { completed: completedTheoryDomains.length })}
              </button>
            ) : (
              <button
                onClick={() => navigate("/simulation")}
                className="px-5 py-2.5 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 transition-all duration-200 flex items-center gap-2 cursor-pointer"
              >
                <PlayCircle className="w-4 h-4" />
                {t("evaluation.startRecommendedScenarios", {
                  completed: completedScenarios,
                  max: maxScenarios,
                })}
              </button>
            )
          ) : (
            <button
              onClick={() => navigate("/simulation/complete")}
              className="px-5 py-2.5 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 transition-all duration-200 flex items-center gap-2 cursor-pointer"
            >
              <Award className="w-4 h-4" />
              {t("evaluation.viewCompletionCertificate")}
            </button>
          )}
        </div>
      </div>

      {latestPostExam && questionnaireStatus && (
        <QuestionnairePromptBanner status={questionnaireStatus} variant="compact" />
      )}

      {(latestPreExam || latestPostExam) && (
        <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
          <div className="space-y-1 border-b border-graphite-800 pb-3">
            <h3 className="text-lg font-semibold flex items-center gap-2 text-white">
              <ClipboardCheck className="text-emerald-400 w-5 h-5" />
              {t("evaluation.examReviews.title")}
            </h3>
            <p className="text-sm text-graphite-400">
              {t("evaluation.examReviews.subtitle")}
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {latestPreExam && (
              <div className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4 flex flex-col justify-between gap-4">
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-base font-semibold text-white">
                      {t("evaluation.examReviews.preTitle")}
                    </span>
                    <span className={`px-2 py-0.5 rounded text-xs font-mono border ${getExamTypeBadgeClass("pre")}`}>
                      {latestPreExam.score.toFixed(1)}%
                    </span>
                  </div>
                  <p className="text-sm text-graphite-300 font-mono">
                    {t("evaluation.examReviews.scoreMeta", {
                      score: latestPreExam.score.toFixed(1),
                      correct: latestPreExam.correct_count,
                      total: latestPreExam.total_questions,
                    })}
                  </p>
                  {latestPreExam.completed_at && (
                    <p className="text-xs text-graphite-400 font-mono">
                      {formatDate(latestPreExam.completed_at)}
                    </p>
                  )}
                </div>

                <Link
                  to={`/exams/review/${encodeURIComponent(latestPreExam.session_id)}`}
                  className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-md border border-graphite-700 bg-graphite-900 hover:bg-graphite-800 text-sm font-medium text-white transition-colors"
                >
                  <span>{t("evaluation.examReviews.reviewAction")}</span>
                  <ArrowRight className="w-4 h-4" />
                </Link>
              </div>
            )}

            {latestPostExam && (
              <div className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4 flex flex-col justify-between gap-4">
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-base font-semibold text-white">
                      {t("evaluation.examReviews.postTitle")}
                    </span>
                    <span className={`px-2 py-0.5 rounded text-xs font-mono border ${getExamTypeBadgeClass("post")}`}>
                      {latestPostExam.score.toFixed(1)}%
                    </span>
                  </div>
                  <p className="text-sm text-graphite-300 font-mono">
                    {t("evaluation.examReviews.scoreMeta", {
                      score: latestPostExam.score.toFixed(1),
                      correct: latestPostExam.correct_count,
                      total: latestPostExam.total_questions,
                    })}
                  </p>
                  <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
                    {latestPostExam.completed_at && (
                      <span className="text-graphite-400">
                        {formatDate(latestPostExam.completed_at)}
                      </span>
                    )}
                    {preToPostDelta !== null && (
                      <span
                        className={
                          preToPostDelta > 0
                            ? "text-emerald-400"
                            : preToPostDelta < 0
                              ? "text-amber-400"
                              : "text-graphite-300"
                        }
                      >
                        {t("evaluation.examReviews.gainMeta", {
                          delta: formatSignedDelta(preToPostDelta),
                        })}
                      </span>
                    )}
                  </div>
                </div>

                <Link
                  to={`/exams/review/${encodeURIComponent(latestPostExam.session_id)}`}
                  className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-md bg-white hover:bg-white/90 text-sm font-semibold text-black transition-colors"
                >
                  <span>{t("evaluation.examReviews.reviewAction")}</span>
                  <ArrowRight className="w-4 h-4" />
                </Link>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
