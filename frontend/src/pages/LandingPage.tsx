import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { 
  ArrowRight, 
  Terminal, 
  Loader2, 
  CheckCircle, 
  XCircle, 
  RefreshCw,
  ShieldAlert,
  ShieldCheck,
} from "lucide-react";
import { useI18n } from "../i18n/index.tsx";
import { resolveDomainName } from "../i18n/format.ts";
import QuestionTranslationToggle from "../components/QuestionTranslationToggle";
import { getSortedOptionEntries } from "../lib/examReview";
import {
  evaluateLandingMiniCase,
  getLandingMiniCases,
  type MiniCaseVerdict,
} from "../lib/landingMiniCase";
import { resolveLocalizedQuestion, type QuestionTranslationsMap } from "../lib/questionLocale.ts";

interface Question {
  question_id: string;
  domain_id: string;
  question_text: string;
  options: Record<string, string>;
  correct_answer: string;
  explanation: string;
  translations?: QuestionTranslationsMap | null;
  difficulty: number;
}

export default function LandingPage() {
  const navigate = useNavigate();
  const { locale, t } = useI18n();
  const [redirectLoading, setRedirectLoading] = useState(false);
  
  // Interactive Workbench State (default to zero-login 30s Detective Mini-Case)
  const [activeWorkbenchTab, setActiveWorkbenchTab] = useState<"detective" | "quiz">("detective");
  const [miniCaseIndex, setMiniCaseIndex] = useState(0);
  const [miniCasePick, setMiniCasePick] = useState<MiniCaseVerdict | null>(null);
  const [sampleLoading, setSampleLoading] = useState(false);
  const [question, setQuestion] = useState<Question | null>(null);
  const [selectedAnswer, setSelectedAnswer] = useState<string | null>(null);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [showOriginalEn, setShowOriginalEn] = useState(false);

  useEffect(() => {
    fetchSampleQuestion();
  }, []);

  const handleStartAssessment = async () => {
    const token = localStorage.getItem("token");
    if (!token) {
      navigate("/login");
      return;
    }

    try {
      setRedirectLoading(true);

      // 1. Check for active exam session first
      const sessionRes = await axios.get("/api/exams/session", {
        headers: { Authorization: `Bearer ${token}` }
      });
      
      if (sessionRes.data && sessionRes.data.session) {
        const type = sessionRes.data.session.exam_type;
        if (type === "pre") {
          navigate("/pre-test");
        } else {
          navigate("/post-test");
        }
        return;
      }
    } catch (err) {
      // No active exam session, proceed to check simulation status
    }

    try {
      // 2. Check simulation status
      const statusRes = await axios.get("/api/simulation/status", {
        headers: { Authorization: `Bearer ${token}` }
      });

      if (statusRes.data.error || statusRes.data.error?.includes("pre-test")) {
        navigate("/pre-test");
      } else if (statusRes.data.needs_training) {
        navigate("/simulation");
      } else {
        navigate("/post-test");
      }
    } catch (err) {
      // Fallback
      navigate("/pre-test");
    } finally {
      setRedirectLoading(false);
    }
  };

  const fetchSampleQuestion = async () => {
    try {
      setSampleLoading(true);
      setSelectedAnswer(null);
      setIsSubmitted(false);
      setShowOriginalEn(false);
      const res = await axios.get("/api/exams/sample");
      setQuestion(res.data);
    } catch (err) {
      console.error("Failed to load sample question", err);
    } finally {
      setSampleLoading(false);
    }
  };

  const handleFocusDetectiveHook = () => {
    setActiveWorkbenchTab("detective");
    if (window.innerWidth < 1024) {
      setTimeout(() => {
        document.getElementById("analyst-workbench")?.scrollIntoView({ behavior: "smooth" });
      }, 100);
    }
  };

  const miniCases = getLandingMiniCases(locale);
  const currentMiniCase = miniCases[miniCaseIndex % miniCases.length];
  const miniCaseResult =
    currentMiniCase && miniCasePick
      ? evaluateLandingMiniCase(currentMiniCase.id, miniCasePick, locale)
      : null;

  const handleNextMiniCase = () => {
    setMiniCaseIndex((prev) => (prev + 1) % miniCases.length);
    setMiniCasePick(null);
  };

  const handleAnswerSelect = (optionKey: string) => {
    if (isSubmitted) return;
    setSelectedAnswer(optionKey);
  };

  const handleSubmitAnswer = () => {
    if (!selectedAnswer) return;
    setIsSubmitted(true);
  };

  const localizedQuestion = question
    ? resolveLocalizedQuestion(question, locale, showOriginalEn)
    : null;

  return (
    <div className="flex flex-col w-full max-w-6xl mx-auto pb-24 space-y-24 animate-slide-up">
      {/* Hero Section: Left-Heavy Asymmetric Split Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center mt-12 md:mt-20">
        
        {/* Left Column: Heading and CTAs */}
        <div className="lg:col-span-5 space-y-6 text-left">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-graphite-800 text-graphite-300 text-xs font-mono bg-graphite-900/40">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            <span>{t("landing.hero.badge")}</span>
          </div>
          
          <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tighter text-white leading-[1.08] text-wrap">
            {t("landing.hero.titleLine1")} <br />
            {t("landing.hero.titleLine2")}
          </h1>
          
          <p className="text-base sm:text-lg text-graphite-300 font-normal leading-relaxed max-w-[45ch]">
            {t("landing.hero.subtitle")}
          </p>
          
          <div className="pt-4 flex flex-col sm:flex-row gap-4 w-full sm:w-auto">
            <button 
              onClick={handleStartAssessment}
              disabled={redirectLoading}
              className="group flex items-center justify-center gap-2 px-6 py-3 rounded-md bg-white text-black font-semibold text-sm sm:text-base hover:bg-white/90 disabled:bg-graphite-800 disabled:text-graphite-400 transition-all duration-200 cursor-pointer"
            >
              {redirectLoading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  {t("landing.hero.evaluating")}
                </>
              ) : (
                <>
                  {t("landing.hero.startAssessment")}
                  <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-0.5" />
                </>
              )}
            </button>
            
            <button 
              onClick={handleFocusDetectiveHook}
              className="flex items-center justify-center gap-2 px-6 py-3 rounded-md border border-graphite-800 bg-transparent text-graphite-200 font-semibold text-sm sm:text-base hover:bg-graphite-900 hover:text-white transition-all duration-200 cursor-pointer"
            >
              <ShieldAlert className="w-4 h-4 text-emerald-400" />
              {t("landing.hero.tryDetectiveCase")}
            </button>
          </div>
        </div>

        {/* Right Column: Tabbed Analyst Dashboard (SOC Workbench) */}
        <div id="analyst-workbench" className="lg:col-span-7 bg-graphite-900 border border-graphite-800 rounded-lg overflow-hidden shadow-2xl min-h-[440px] flex flex-col justify-between">
          
          {/* Header Tab Bar */}
          <div className="flex bg-graphite-950 border-b border-graphite-800 px-4 h-12 items-center justify-between select-none">
            <div className="flex gap-2">
              {([
                {
                  id: "detective" as const,
                  label: t("landing.workbench.detectiveTab"),
                  icon: <ShieldAlert className="w-4 h-4 text-emerald-400" />,
                },
                {
                  id: "quiz" as const,
                  label: t("landing.workbench.sampleQuizTab"),
                  icon: <Terminal className="w-4 h-4" />,
                },
              ]).map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveWorkbenchTab(tab.id)}
                  className={`flex items-center gap-2 px-3 py-2 text-sm font-mono transition-all border-b-2 h-12 cursor-pointer ${
                    activeWorkbenchTab === tab.id
                      ? "border-white text-white font-medium"
                      : "border-transparent text-graphite-400 hover:text-graphite-200"
                  }`}
                >
                  {tab.icon}
                  {tab.label}
                </button>
              ))}
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-graphite-800 border border-graphite-700" />
              <span className="w-2.5 h-2.5 rounded-full bg-graphite-800 border border-graphite-700" />
              <span className="w-2.5 h-2.5 rounded-full bg-graphite-800 border border-graphite-700" />
            </div>
          </div>

          {/* Workbench Body */}
          <div className="flex-1 p-6 overflow-y-auto">
            {activeWorkbenchTab === "detective" && currentMiniCase ? (
              <div className="space-y-4">
                <div className="flex flex-wrap justify-between items-center gap-2 border-b border-graphite-800 pb-2.5">
                  <span className="text-sm text-emerald-400 font-semibold font-mono">{currentMiniCase.title}</span>
                  <span className="text-xs font-mono text-graphite-400">
                    {t("landing.workbench.caseProgress", {
                      current: (miniCaseIndex % miniCases.length) + 1,
                      total: miniCases.length,
                    })}
                  </span>
                </div>

                <p className="text-base text-graphite-100 leading-relaxed">{currentMiniCase.briefing}</p>

                {/* Authentic English SIEM Telemetry Lines */}
                <div className="rounded-md border border-graphite-800 bg-graphite-950 p-3.5 space-y-2 font-mono text-log">
                  {currentMiniCase.logs.map((log, idx) => (
                    <div key={`${currentMiniCase.id}-${idx}`} className="flex items-start gap-2.5 leading-relaxed">
                      <span className="text-graphite-400 shrink-0">{log.timestamp}</span>
                      <span
                        className={`px-1.5 py-0.5 rounded text-xs font-bold shrink-0 ${
                          log.severity === "ALERT"
                            ? "bg-red-500/15 text-red-400 border border-red-500/30"
                            : log.severity === "WARN"
                              ? "bg-amber-500/15 text-amber-300 border border-amber-500/30"
                              : "bg-graphite-800 text-graphite-300 border border-graphite-700"
                        }`}
                      >
                        {log.severity}
                      </span>
                      <span className="text-graphite-200 break-words min-w-0">
                        <span className="text-emerald-400">[{log.source}]</span> {log.message}
                      </span>
                    </div>
                  ))}
                </div>

                <div className="space-y-2.5 pt-1">
                  <p className="text-sm font-medium text-graphite-300">
                    {t("landing.workbench.pickVerdictPrompt")}
                  </p>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <button
                      type="button"
                      onClick={() => setMiniCasePick("tp")}
                      className={`p-3 rounded-md border text-sm font-semibold flex items-center justify-center gap-2 transition-all cursor-pointer ${
                        miniCasePick === "tp"
                          ? "border-red-400 bg-red-500/15 text-red-300"
                          : "border-graphite-800 bg-graphite-950/50 text-graphite-200 hover:border-graphite-600 hover:text-white"
                      }`}
                    >
                      <ShieldAlert className="w-4 h-4 shrink-0 text-red-400" />
                      <span>{t("landing.workbench.realAttackBtn")}</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => setMiniCasePick("fp")}
                      className={`p-3 rounded-md border text-sm font-semibold flex items-center justify-center gap-2 transition-all cursor-pointer ${
                        miniCasePick === "fp"
                          ? "border-emerald-400 bg-emerald-500/15 text-emerald-300"
                          : "border-graphite-800 bg-graphite-950/50 text-graphite-200 hover:border-graphite-600 hover:text-white"
                      }`}
                    >
                      <ShieldCheck className="w-4 h-4 shrink-0 text-emerald-400" />
                      <span>{t("landing.workbench.safeRoutineBtn")}</span>
                    </button>
                  </div>
                </div>

                {miniCaseResult && (
                  <div className="rounded border p-4 space-y-2 animate-fade-in bg-graphite-950 border-graphite-800 text-base leading-relaxed">
                    {miniCaseResult.isCorrect ? (
                      <div className="flex items-center gap-2 text-emerald-400 font-semibold">
                        <CheckCircle className="w-4 h-4 shrink-0" />
                        <span>{t("landing.workbench.verdictCorrectTitle")}</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-2 text-amber-300 font-semibold">
                        <XCircle className="w-4 h-4 shrink-0" />
                        <span>{t("landing.workbench.verdictIncorrectTitle")}</span>
                      </div>
                    )}
                    <p className="text-graphite-300">
                      <strong className="text-white font-semibold">
                        {t("landing.workbench.verdictExplanationLabel")}
                      </strong>{" "}
                      {miniCaseResult.rationale}
                    </p>
                  </div>
                )}
              </div>
            ) : sampleLoading ? (
                <div className="h-full min-h-[260px] flex flex-col items-center justify-center gap-3">
                  <Loader2 className="w-6 h-6 text-white animate-spin" />
                  <p className="text-sm font-mono text-graphite-400">{t("landing.workbench.fetchingSample")}</p>
                </div>
              ) : question && localizedQuestion ? (
                <div className="space-y-4">
                  <div className="flex flex-wrap justify-between items-center gap-2 text-xs font-mono text-graphite-400 border-b border-graphite-800 pb-2.5">
                    <span>{t("landing.workbench.questionSource")}</span>
                    <div className="flex items-center gap-2.5">
                      <QuestionTranslationToggle
                        visible={locale === "th" && localizedQuestion.hasThaiTranslation}
                        showingOriginalEn={localizedQuestion.showingOriginalEn}
                        onToggle={() => setShowOriginalEn((prev) => !prev)}
                      />
                      <span>{t("landing.workbench.difficulty", { difficulty: question.difficulty })}</span>
                    </div>
                  </div>
                  <div className="text-base text-graphite-100 font-normal leading-relaxed">
                    {localizedQuestion.questionText}
                  </div>
                  <div className="space-y-2.5 pt-2">
                    {getSortedOptionEntries(localizedQuestion.options).map(([key, value]) => {
                      const isSelected = selectedAnswer === key;
                      const isCorrectAnswer = question.correct_answer === key;
                      
                      let optionStyles = "border-graphite-800 bg-graphite-950/40 hover:bg-graphite-950 text-graphite-200";
                      if (isSelected && !isSubmitted) {
                        optionStyles = "border-white bg-graphite-950 text-white";
                      } else if (isSubmitted) {
                        if (isCorrectAnswer) {
                          optionStyles = "border-emerald-500/30 bg-emerald-950/20 text-emerald-400";
                        } else if (isSelected) {
                          optionStyles = "border-red-500/30 bg-red-950/20 text-red-400";
                        } else {
                          optionStyles = "border-graphite-900 opacity-50 text-graphite-400";
                        }
                      }

                      return (
                        <button
                          key={key}
                          onClick={() => handleAnswerSelect(key)}
                          disabled={isSubmitted}
                          className={`w-full text-left p-3.5 rounded border text-base font-normal leading-relaxed transition-all flex items-start gap-3 ${optionStyles}`}
                        >
                          <span className="font-semibold text-white shrink-0 font-mono">{key})</span>
                          <span>{value}</span>
                        </button>
                      );
                    })}
                  </div>
                  
                  {isSubmitted && (
                    <div className="rounded border p-4 space-y-2 animate-fade-in bg-graphite-950 border-graphite-800 text-base font-normal leading-relaxed">
                      {selectedAnswer === question.correct_answer ? (
                        <div className="flex items-center gap-2 text-emerald-400 font-semibold">
                          <CheckCircle className="w-4 h-4 shrink-0" />
                          <span>{t("landing.workbench.correct")}</span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-2 text-red-400 font-semibold">
                          <XCircle className="w-4 h-4 shrink-0" />
                          <span>{t("landing.workbench.incorrect", { answer: question.correct_answer })}</span>
                        </div>
                      )}
                      <p className="text-graphite-300">
                        <strong className="text-white font-semibold">{t("landing.workbench.explanationLabel")}</strong> {localizedQuestion.explanation}
                      </p>
                    </div>
                  )}
                </div>
              ) : (
                <div className="h-full flex items-center justify-center text-graphite-400 font-mono text-sm">
                  {t("landing.workbench.noQuestion")}
                </div>
              )}
          </div>

          {/* Workbench Footer Actions */}
          {activeWorkbenchTab === "detective" ? (
            <div className="bg-graphite-950 border-t border-graphite-800 px-6 py-3.5 flex justify-end items-center select-none">
              <button
                type="button"
                onClick={handleNextMiniCase}
                className="flex items-center gap-2 px-4 py-2 rounded border border-graphite-700 bg-transparent text-graphite-200 text-sm font-medium hover:bg-graphite-900 hover:text-white transition-colors cursor-pointer"
              >
                <RefreshCw className="w-4 h-4" />
                {t("landing.workbench.nextCaseBtn")}
              </button>
            </div>
          ) : activeWorkbenchTab === "quiz" && question ? (
              <div className="bg-graphite-950 border-t border-graphite-800 px-6 py-3.5 flex justify-between items-center select-none">
                <button
                  onClick={fetchSampleQuestion}
                  className="flex items-center gap-2 px-3.5 py-2 rounded border border-graphite-800 bg-transparent text-graphite-300 text-sm font-medium hover:bg-graphite-900 hover:text-white transition-colors cursor-pointer"
                >
                  <RefreshCw className="w-4 h-4" />
                  {t("landing.workbench.tryAnother")}
                </button>

                {!isSubmitted ? (
                  <button
                    onClick={handleSubmitAnswer}
                    disabled={!selectedAnswer}
                    className="px-4 py-2 rounded bg-white text-black text-sm font-semibold hover:bg-white/90 disabled:bg-graphite-800 disabled:text-graphite-400 transition-all duration-200 cursor-pointer"
                  >
                    {t("landing.workbench.submitAnswer")}
                  </button>
                ) : (
                  <span className="text-xs text-graphite-400 font-mono">{t("landing.workbench.submitted")}</span>
                )}
              </div>
          ) : null}
        </div>
      </div>

      {/* High-Visual Curriculum Bento Grid */}
      <div className="space-y-6 pt-12">
        <div className="text-left space-y-1.5">
          <h2 className="text-2xl font-bold tracking-tight text-white">{t("landing.bento.heading")}</h2>
          <p className="text-xs text-graphite-400 font-mono uppercase tracking-wider">{t("landing.bento.subheading")}</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-6 gap-6 w-full">
          
          {/* Cell 1: Adaptive Training Loop (2 columns) */}
          <div className="md:col-span-2 bg-graphite-900 border border-graphite-800/80 rounded-lg p-6 flex flex-col justify-between hover:border-graphite-700 transition-colors duration-300">
            <div>
              <span className="text-xs uppercase tracking-wider font-mono text-graphite-400">{t("landing.bento.cell1.eyebrow")}</span>
              <h3 className="text-lg font-semibold text-white mt-1.5">{t("landing.bento.cell1.title")}</h3>
              <p className="text-graphite-300 text-sm sm:text-base mt-2 font-normal leading-relaxed">
                {t("landing.bento.cell1.description")}
              </p>
            </div>
            
            {/* 4-Stage Mission Roadmap */}
            <div className="mt-6 pt-4 border-t border-graphite-800 flex flex-col gap-2.5 font-mono text-xs text-graphite-300 select-none">
              <div className="flex items-center gap-2.5">
                <span className="w-6 h-6 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium shrink-0">01</span>
                <span>{t("landing.bento.cell1.step1")}</span>
              </div>
              <div className="w-px h-2.5 bg-graphite-800 ml-3" />
              <div className="flex items-center gap-2.5">
                <span className="w-6 h-6 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium shrink-0">02</span>
                <span>{t("landing.bento.cell1.step2")}</span>
              </div>
              <div className="w-px h-2.5 bg-graphite-800 ml-3" />
              <div className="flex items-center gap-2.5">
                <span className="w-6 h-6 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-emerald-400 font-medium shrink-0">03</span>
                <span className="text-white font-medium">{t("landing.bento.cell1.step3")}</span>
              </div>
              <div className="w-px h-2.5 bg-graphite-800 ml-3" />
              <div className="flex items-center gap-2.5">
                <span className="w-6 h-6 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium shrink-0">04</span>
                <span>{t("landing.bento.cell1.step4")}</span>
              </div>
            </div>
          </div>

          {/* Cell 2: Incident Response Playbooks (4 columns) */}
          <div className="md:col-span-4 bg-graphite-900 border border-graphite-800/80 rounded-lg p-6 hover:border-graphite-700 transition-colors duration-300 flex flex-col justify-between">
            <div className="space-y-2">
              <span className="text-xs uppercase tracking-wider font-mono text-graphite-400">{t("landing.bento.cell2.eyebrow")}</span>
              <h3 className="text-lg font-semibold text-white">{t("landing.bento.cell2.title")}</h3>
              <p className="text-graphite-300 text-sm sm:text-base font-normal leading-relaxed">
                {t("landing.bento.cell2.description")}
              </p>
            </div>
            
            {/* Playbook interactive step visual */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mt-8 pt-4 border-t border-graphite-800 font-mono text-xs select-none">
              <div className="border border-graphite-800 rounded bg-graphite-950/60 p-3.5 text-center space-y-1.5">
                <div className="text-white font-bold text-sm">{t("landing.bento.cell2.phase1Title")}</div>
                <div className="text-xs text-graphite-400 leading-snug">{t("landing.bento.cell2.phase1Desc")}</div>
              </div>
              <div className="border border-graphite-800 rounded bg-graphite-950/60 p-3.5 text-center space-y-1.5">
                <div className="text-white font-bold text-sm">{t("landing.bento.cell2.phase2Title")}</div>
                <div className="text-xs text-graphite-400 leading-snug">{t("landing.bento.cell2.phase2Desc")}</div>
              </div>
              <div className="border border-graphite-800 rounded bg-graphite-950/60 p-3.5 text-center space-y-1.5">
                <div className="text-white font-bold text-sm">{t("landing.bento.cell2.phase3Title")}</div>
                <div className="text-xs text-graphite-400 leading-snug">{t("landing.bento.cell2.phase3Desc")}</div>
              </div>
            </div>
          </div>

          {/* Cell 3: Continuous AI Diagnostics (6 columns) */}
          <div className="md:col-span-6 bg-graphite-900 border border-graphite-800/80 rounded-lg p-6 hover:border-graphite-700 transition-colors duration-300">
            <div className="grid md:grid-cols-12 gap-6 items-center">
              <div className="md:col-span-7 space-y-2 text-left">
                <span className="text-xs uppercase tracking-wider font-mono text-graphite-400">{t("landing.bento.cell3.eyebrow")}</span>
                <h3 className="text-lg font-semibold text-white">{t("landing.bento.cell3.title")}</h3>
                <p className="text-graphite-300 text-sm sm:text-base font-normal leading-relaxed">
                  {t("landing.bento.cell3.description")}
                </p>
              </div>
              
              <div className="md:col-span-5 space-y-3 font-mono text-xs text-graphite-300 select-none">
                {[
                  { id: "domain1", pct: 85 },
                  { id: "domain2", pct: 72 },
                  { id: "domain3", pct: 90 },
                ].map((item) => {
                  const label = resolveDomainName(item.id, item.id, locale);
                  return (
                    <div key={item.id}>
                      <div className="flex justify-between items-start gap-2 text-graphite-200 mb-1.5">
                        <span className="leading-snug break-words min-w-0 flex-1" title={label}>{label}</span>
                        <span className="text-white font-bold shrink-0">{item.pct}%</span>
                      </div>
                      <div className="w-full bg-graphite-950 h-1.5 rounded overflow-hidden border border-graphite-800">
                        <div className="bg-white h-full" style={{ width: `${item.pct}%` }} />
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
