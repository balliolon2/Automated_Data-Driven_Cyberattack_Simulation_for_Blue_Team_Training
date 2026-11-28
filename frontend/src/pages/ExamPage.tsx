import { useState, useEffect, useRef, useCallback } from "react";
import { Clock, Flag, ChevronLeft, ChevronRight, CheckCircle2, Sparkles, X } from "lucide-react";
import axios from "axios";
import { useNavigate } from "react-router-dom";
import QuestionTranslationToggle from "../components/QuestionTranslationToggle";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import { EXAM_MAX_DURATION_SECONDS, getSortedOptionEntries } from "../lib/examReview";
import { resolveLocalizedQuestion, type QuestionTranslationsMap } from "../lib/questionLocale.ts";

const EXAM_DURATION_SECONDS = 30 * 60;

interface Question {
  question_id: string;
  domain_id: string;
  type: string;
  question_text: string;
  options: Record<string, string>;
  translations?: QuestionTranslationsMap;
  order_index: number;
  user_answer?: string;
}

interface ExamSession {
  session_id: string;
  user_id: string;
  exam_type: string;
  status: string;
  total_questions: number;
  time_spent_seconds: number | null;
}

export default function ExamPage({ mode }: { mode: "pre" | "post" }) {
  const { locale, t } = useI18n();
  const [currentQuestionIdx, setCurrentQuestionIdx] = useState(0);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [session, setSession] = useState<ExamSession | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [timeLeft, setTimeLeft] = useState(EXAM_DURATION_SECONDS);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [autoSubmitting, setAutoSubmitting] = useState(false);
  const [showOriginalEn, setShowOriginalEn] = useState(false);
  const [warmupDismissed, setWarmupDismissed] = useState(false);
  const navigate = useNavigate();

  const goToQuestion = useCallback((nextIdx: number) => {
    setCurrentQuestionIdx(nextIdx);
    setShowOriginalEn(false);
  }, []);

  const hasTickedRef = useRef(false);
  const autoFinalizeTriggeredRef = useRef(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    let cancelled = false;
    hasTickedRef.current = false;
    autoFinalizeTriggeredRef.current = false;

    const fetchSession = async () => {
      try {
        const token = localStorage.getItem("token");
        const res = await axios.get("/api/exams/session", {
          headers: { Authorization: `Bearer ${token}` }
        });
        if (cancelled) return;

        if (res.data && res.data.session) {
          const activeType = res.data.session.exam_type;
          if (activeType !== mode) {
            if (activeType === "pre") {
              navigate("/pre-test");
            } else {
              navigate("/post-test");
            }
            return;
          }
          setSession(res.data.session);
          setQuestions(res.data.questions);

          const initialAnswers: Record<string, string> = {};
          res.data.questions.forEach((q: Question) => {
            if (q.user_answer) {
              initialAnswers[q.question_id] = q.user_answer;
            }
          });
          setAnswers(initialAnswers);
          setTimeLeft(EXAM_DURATION_SECONDS);
        } else {
          setSession(null);
          setQuestions([]);
          setAnswers({});
        }
      } catch (err) {
        console.error("Failed to load exam session:", err);
        if (!cancelled) {
          setSession(null);
          setQuestions([]);
          setAnswers({});
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    void fetchSession();
    return () => {
      cancelled = true;
    };
  }, [mode, navigate]);

  useEffect(() => {
    if (loading || !session || questions.length === 0) return;
    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 0) return 0;
        hasTickedRef.current = true;
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(timer);
  }, [loading, session, questions.length]);

  const formatTime = (seconds: number) => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  const handleAnswerSelect = (qId: string, answer: string) => {
    setAnswers(prev => ({ ...prev, [qId]: answer }));
  };

  const submitAndFinalize = useCallback(async (skipConfirm: boolean) => {
    if (!session || submittingRef.current) return;

    const unanswered = questions.filter(q => !answers[q.question_id]).length;
    if (!skipConfirm && unanswered > 0) {
      if (!window.confirm(t("exam.confirmUnansweredSubmit", { count: unanswered }))) {
        return;
      }
    }

    submittingRef.current = true;
    setSubmitting(true);
    const token = localStorage.getItem("token");

    const previousTimeSpent = session.time_spent_seconds ?? 0;
    const currentRunElapsed = Math.max(1, EXAM_DURATION_SECONDS - timeLeft);
    const totalTimeSpent = Math.min(EXAM_MAX_DURATION_SECONDS, previousTimeSpent + currentRunElapsed);

    try {
      const unsavedQuestions = questions.filter((q) => {
        const userAns = answers[q.question_id];
        return Boolean(userAns && userAns !== q.user_answer);
      });

      let completedViaAnswer = false;
      for (let i = 0; i < unsavedQuestions.length; i++) {
        const q = unsavedQuestions[i];
        const userAns = answers[q.question_id];
        const isLastCompletePick = unanswered === 0 && i === unsavedQuestions.length - 1;
        const ansRes = await axios.post("/api/exams/submit-answer", {
          session_id: session.session_id,
          question_id: q.question_id,
          user_answer: userAns,
          time_spent_seconds: isLastCompletePick ? currentRunElapsed : 0
        }, {
          headers: { Authorization: `Bearer ${token}` }
        });
        if (ansRes.data?.completed) {
          completedViaAnswer = true;
        }
      }

      if (!completedViaAnswer) {
        await axios.post(
          `/api/exams/sessions/${encodeURIComponent(session.session_id)}/finalize`,
          { time_spent_seconds: totalTimeSpent },
          { headers: { Authorization: `Bearer ${token}` } }
        );
      }

      navigate(`/exams/review/${encodeURIComponent(session.session_id)}`, { replace: true });
    } catch (err: unknown) {
      console.error(err);
      alert(resolveApiError(err, t, t("exam.submitFailed")));
      setAutoSubmitting(false);
    } finally {
      submittingRef.current = false;
      setSubmitting(false);
    }
  }, [answers, navigate, questions, session, t, timeLeft]);

  useEffect(() => {
    if (
      hasTickedRef.current &&
      timeLeft === 0 &&
      session &&
      questions.length > 0 &&
      !loading &&
      !submitting &&
      !autoFinalizeTriggeredRef.current
    ) {
      autoFinalizeTriggeredRef.current = true;
      setAutoSubmitting(true);
      void submitAndFinalize(true);
    }
  }, [loading, questions.length, session, submitAndFinalize, submitting, timeLeft]);

  const handleSubmitExam = () => {
    void submitAndFinalize(false);
  };

  const startExam = async (type: "pre" | "post") => {
    try {
      setLoading(true);
      const token = localStorage.getItem("token");
      const endpoint = type === "pre" ? "/api/exams/pre-test" : "/api/exams/post-test";
      const res = await axios.post(endpoint, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      hasTickedRef.current = false;
      autoFinalizeTriggeredRef.current = false;
      setSession(res.data.session);
      setQuestions(res.data.questions);
      setAnswers({});
      setCurrentQuestionIdx(0);
      setTimeLeft(EXAM_DURATION_SECONDS);
    } catch (err: unknown) {
      console.error("Failed to start exam:", err);
      alert(resolveApiError(err, t, t("exam.startFailed")));
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="p-8 text-white font-mono flex items-center justify-center min-h-[50vh]">{t("exam.loadingSession")}</div>;
  }

  if (!questions.length) {
    return (
      <div className="max-w-xl mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-slide-up">
        {mode === "pre" ? (
          <>
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">{t("exam.preTest.title")}</h2>
            <p className="text-graphite-300 font-normal text-base leading-relaxed">
              {t("exam.preTest.descBeforeCount")}
              <strong className="text-white">{t("exam.preTest.questionCountText")}</strong>
              {t("exam.preTest.descAfterCount")}
            </p>
            <div className="rounded-md border border-emerald-500/30 bg-emerald-500/10 p-4 text-left space-y-1.5">
              <div className="flex items-center gap-2 text-sm font-semibold text-emerald-300">
                <Sparkles className="w-4 h-4 shrink-0" />
                <span>{t("exam.preTest.warmupTitle")}</span>
              </div>
              <p className="text-sm text-graphite-200 leading-relaxed">
                {t("exam.preTest.warmupBody")}
              </p>
            </div>
            <div className="pt-2">
              <button 
                onClick={() => startExam("pre")}
                className="w-full py-3 bg-white hover:bg-white/90 text-black font-semibold text-sm sm:text-base rounded-md transition-all duration-200 cursor-pointer"
              >
                {t("exam.preTest.startButton")}
              </button>
            </div>
          </>
        ) : (
          <>
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">{t("exam.postTest.title")}</h2>
            <p className="text-graphite-300 font-normal text-base leading-relaxed">
              {t("exam.postTest.descBeforeCount")}
              <strong className="text-white">{t("exam.postTest.questionCountText")}</strong>
              {t("exam.postTest.descAfterCount")}
            </p>
            <div className="pt-4">
              <button 
                onClick={() => startExam("post")}
                className="w-full py-3 bg-white hover:bg-white/90 text-black font-semibold text-sm sm:text-base rounded-md transition-all duration-200 cursor-pointer"
              >
                {t("exam.postTest.startButton")}
              </button>
            </div>
          </>
        )}
        <button 
          onClick={() => navigate("/")}
          className="mt-4 text-sm text-graphite-400 hover:text-white transition-colors underline decoration-dotted cursor-pointer"
        >
          {t("exam.backToDashboard")}
        </button>
      </div>
    );
  }

  const currentQ = questions[currentQuestionIdx];
  if (!currentQ) {
    return null;
  }
  const localizedQ = resolveLocalizedQuestion(currentQ, locale, showOriginalEn);
  const totalQuestions = questions.length;
  const answeredCount = questions.filter(q => !!answers[q.question_id]).length;
  const progressPercent = totalQuestions > 0 ? Math.round((answeredCount / totalQuestions) * 100) : 0;
  const isComplete = totalQuestions > 0 && answeredCount === totalQuestions;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 min-h-[75vh]">
      {/* Sidebar: Navigator */}
      <div className="lg:col-span-1 bg-graphite-900 border border-graphite-800 rounded-lg p-4 flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-2">
            <h3 className="font-semibold text-sm text-white">{t("exam.navigatorTitle")}</h3>
            <span
              className={`text-xs font-mono px-2 py-0.5 rounded border transition-colors ${
                isComplete
                  ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400 font-medium'
                  : 'bg-graphite-950 border-graphite-800 text-graphite-300'
              }`}
            >
              {t("exam.answeredCount", { answered: answeredCount, total: totalQuestions })}
            </span>
          </div>

          {/* Micro Progress Bar */}
          <div className="mb-4">
            <div className="w-full bg-graphite-950 rounded-full h-1.5 border border-graphite-800/80 overflow-hidden">
              <div
                className={`h-full transition-all duration-300 rounded-full ${
                  isComplete ? 'bg-emerald-500' : 'bg-white'
                }`}
                style={{ width: `${progressPercent}%` }}
              />
            </div>
          </div>
          <div className="max-h-[350px] overflow-y-auto pr-2 custom-scrollbar">
            <div className="grid grid-cols-5 gap-2">
              {questions.map((q, i) => {
                const isAnswered = !!answers[q.question_id];
                const isCurrent = currentQuestionIdx === i;
                
                let btnStyles = "bg-graphite-950 border-graphite-800 text-graphite-400 hover:border-graphite-700";
                if (isCurrent) {
                  btnStyles = "bg-white/10 border-white text-white";
                } else if (isAnswered) {
                  btnStyles = "bg-graphite-800 border-graphite-600 text-white";
                }

                return (
                  <button
                    key={q.question_id}
                    onClick={() => goToQuestion(i)}
                    className={`w-10 h-10 rounded text-xs font-mono border transition-colors flex items-center justify-center ${btnStyles}`}
                  >
                    {(i + 1).toString().padStart(2, '0')}
                  </button>
                );
              })}
            </div>
          </div>
        </div>
        <div className="mt-6 pt-4 border-t border-graphite-800">
          <button 
            onClick={handleSubmitExam}
            disabled={submitting}
            className="w-full py-2.5 bg-white text-black hover:bg-white/90 disabled:bg-graphite-800 disabled:text-graphite-400 text-sm font-semibold rounded-md transition-all flex items-center justify-center gap-2 cursor-pointer"
          >
            <CheckCircle2 className="w-4 h-4" />
            {submitting ? t("exam.submitting") : t("exam.submitExam")}
          </button>
        </div>
      </div>

      {/* Main Content: Question (Hybrid Bilingual Thai when locale=th with per-question Show EN toggle) */}
      <div className="lg:col-span-3 bg-graphite-900 border border-graphite-800 rounded-lg p-6 sm:p-8 flex flex-col justify-between">
        <div>
          <div className="flex flex-wrap justify-between items-center gap-3 border-b border-graphite-800 pb-4 mb-6">
            <div className="flex flex-wrap items-center gap-2.5">
              <h2 className="text-lg font-semibold text-white">
                {t("exam.questionOf", { current: currentQuestionIdx + 1, total: totalQuestions })}
              </h2>
              <QuestionTranslationToggle
                visible={locale === "th" && localizedQ.hasThaiTranslation}
                showingOriginalEn={localizedQ.showingOriginalEn}
                onToggle={() => setShowOriginalEn((prev) => !prev)}
              />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs font-mono text-indigo-300 bg-indigo-500/10 px-2.5 py-1.5 rounded border border-indigo-500/30 select-none">
                {t("exam.targetPaceBadge")}
              </span>
              <div className="flex items-center gap-2 text-white font-mono text-sm bg-graphite-950 px-3 py-1.5 rounded border border-graphite-800 select-none">
                <Clock className="w-4 h-4 text-graphite-400" />
                {formatTime(timeLeft)}
              </div>
            </div>
          </div>

          {mode === "pre" && !warmupDismissed && (
            <div className="mb-4 px-3.5 py-2.5 rounded-md bg-emerald-500/10 border border-emerald-500/30 text-emerald-200 text-sm flex items-start justify-between gap-3">
              <div className="flex items-start gap-2 leading-relaxed">
                <Sparkles className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                <span>{t("exam.preTest.warmupCompactStrip")}</span>
              </div>
              <button
                type="button"
                onClick={() => setWarmupDismissed(true)}
                aria-label={t("exam.preTest.dismissWarmupAria")}
                className="text-emerald-300 hover:text-white p-0.5 rounded transition-colors cursor-pointer shrink-0"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          )}

          {locale === "th" && (
            <div className="mb-4 px-3.5 py-2.5 rounded bg-graphite-900/70 border border-graphite-800 text-graphite-300 text-xs">
              {localizedQ.hasThaiTranslation ? t("exam.bilingualToggleHint") : t("exam.englishOnlyNotice")}
            </div>
          )}

          {timeLeft === 0 && (
            <div className="mb-6 bg-red-500/10 border border-red-500/20 rounded p-4 text-red-400 text-sm">
              <p className="font-medium flex items-center gap-2">
                <span>⏰</span> {autoSubmitting ? t("exam.autoSubmittingNotice") : t("exam.timesUpNotice")}
              </p>
            </div>
          )}

          <div className="space-y-6">
            <div className="text-base sm:text-lg text-graphite-100 font-normal leading-relaxed">
              {localizedQ.questionText}
            </div>
            <div className="space-y-3">
              {getSortedOptionEntries(localizedQ.options).map(([key, optText]) => {
                const isChecked = answers[currentQ.question_id] === key;
                return (
                  <label 
                    key={key} 
                    className={`flex items-start gap-4 p-4 border rounded-md cursor-pointer transition-colors text-base font-normal leading-relaxed ${
                      isChecked 
                        ? 'border-white bg-graphite-950 text-white' 
                        : 'border-graphite-800 bg-graphite-950/40 hover:bg-graphite-950 text-graphite-200'
                    }`}
                  >
                    <input 
                      type="radio" 
                      name={`q${currentQ.question_id}`} 
                      className={`w-4 h-4 mt-1 accent-white ${timeLeft === 0 ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'}`} 
                      checked={isChecked}
                      onChange={() => {
                        if (timeLeft > 0) {
                          handleAnswerSelect(currentQ.question_id, key);
                        }
                      }}
                      disabled={timeLeft === 0}
                    />
                    <span className="font-semibold font-mono text-white shrink-0">{key})</span>
                    <span>{optText as string}</span>
                  </label>
                );
              })}
            </div>
          </div>
        </div>

        <div className="flex justify-between items-center mt-12 pt-4 border-t border-graphite-800">
          <button className="px-3 py-2 flex items-center gap-2 text-sm text-graphite-400 hover:text-white transition-colors cursor-pointer">
            <Flag className="w-4 h-4" /> {t("exam.flagForReview")}
          </button>
          <div className="flex gap-3">
            <button 
              onClick={() => goToQuestion(Math.max(0, currentQuestionIdx - 1))}
              className="px-4 py-2 rounded-md border border-graphite-800 text-sm text-graphite-200 hover:bg-graphite-900 hover:text-white flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              disabled={currentQuestionIdx === 0}
            >
              <ChevronLeft className="w-4 h-4" /> {t("exam.previous")}
            </button>
            <button 
              onClick={() => goToQuestion(Math.min(totalQuestions - 1, currentQuestionIdx + 1))}
              className="px-4 py-2 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              disabled={currentQuestionIdx === totalQuestions - 1}
            >
              {t("exam.next")} <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
