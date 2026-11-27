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
  Activity
} from "lucide-react";
import { useI18n } from "../i18n/index.tsx";
import { resolveDomainName } from "../i18n/format.ts";
import QuestionTranslationToggle from "../components/QuestionTranslationToggle";
import { getSortedOptionEntries } from "../lib/examReview";
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
  
  // Sample Question State
  const [activeWorkbenchTab, setActiveWorkbenchTab] = useState<"quiz" | "logs">("quiz");
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

  const handleViewSampleExam = () => {
    setActiveWorkbenchTab("quiz");
    if (!question) {
      fetchSampleQuestion();
    }
    // Scroll smoothly to section on mobile
    if (window.innerWidth < 1024) {
      setTimeout(() => {
        document.getElementById("analyst-workbench")?.scrollIntoView({ behavior: "smooth" });
      }, 100);
    }
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
              onClick={handleViewSampleExam}
              className="flex items-center justify-center gap-2 px-6 py-3 rounded-md border border-graphite-800 bg-transparent text-graphite-200 font-semibold text-sm sm:text-base hover:bg-graphite-900 hover:text-white transition-all duration-200 cursor-pointer"
            >
              <Terminal className="w-4 h-4" />
              {t("landing.hero.viewSampleExam")}
            </button>
          </div>
        </div>

        {/* Right Column: Tabbed Analyst Dashboard (SOC Workbench) */}
        <div id="analyst-workbench" className="lg:col-span-7 bg-graphite-900 border border-graphite-800 rounded-lg overflow-hidden shadow-2xl min-h-[440px] flex flex-col justify-between">
          
          {/* Header Tab Bar */}
          <div className="flex bg-graphite-950 border-b border-graphite-800 px-4 h-12 items-center justify-between select-none">
            <div className="flex gap-2">
              <button
                onClick={() => setActiveWorkbenchTab("quiz")}
                className={`flex items-center gap-2 px-3 py-2 text-sm font-mono transition-all border-b-2 h-12 cursor-pointer ${
                  activeWorkbenchTab === "quiz"
                    ? "border-white text-white font-medium"
                    : "border-transparent text-graphite-400 hover:text-graphite-200"
                }`}
              >
                <Terminal className="w-4 h-4" />
                triage-quiz.md
              </button>
              <button
                onClick={() => setActiveWorkbenchTab("logs")}
                className={`flex items-center gap-2 px-3 py-2 text-sm font-mono transition-all border-b-2 h-12 cursor-pointer ${
                  activeWorkbenchTab === "logs"
                    ? "border-white text-white font-medium"
                    : "border-transparent text-graphite-400 hover:text-graphite-200"
                }`}
              >
                <Activity className="w-4 h-4" />
                live-siem-logs.sh
              </button>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-graphite-800 border border-graphite-700" />
              <span className="w-2.5 h-2.5 rounded-full bg-graphite-800 border border-graphite-700" />
              <span className="w-2.5 h-2.5 rounded-full bg-graphite-800 border border-graphite-700" />
            </div>
          </div>

          {/* Workbench Body */}
          <div className="flex-1 p-6 overflow-y-auto">
            {activeWorkbenchTab === "quiz" ? (
              // Quiz content
              sampleLoading ? (
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
              )
            ) : (
              // Live logs content (strictly English SIEM telemetry)
              <div className="h-full min-h-[260px] flex flex-col justify-between font-mono text-log text-graphite-200">
                <div className="space-y-2.5 select-none">
                  <div className="text-graphite-400 border-b border-graphite-800 pb-2 mb-2 flex justify-between text-xs">
                    <span>active-sim-triage.log</span>
                    <span className="text-emerald-400 flex items-center gap-1.5">
                      <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" /> {t("landing.workbench.liveStream")}
                    </span>
                  </div>
                  <div className="text-red-400 font-semibold">[ALERT] Port Scan T1046 detected from 192.0.2.14</div>
                  <div className="text-graphite-400">[INFO] Executing KQL filter: SecurityAlert | where ProviderName == "MDATP"</div>
                  <div className="text-graphite-400">SrcIP: 198.51.100.45 -{"->"} DstIP: 10.0.2.15</div>
                  <div className="text-graphite-300">&gt; T1110.001 Brute Force target identified</div>
                  <div className="text-emerald-400">&gt; Playbook Containment: Isolation rule dispatched to firewall.</div>
                  <div className="text-graphite-400">[STATUS] Host isolation status: COMPLETED</div>
                </div>
                <div className="border-t border-graphite-800 pt-4 flex justify-between items-center">
                  <span className="text-xs text-graphite-400">{t("landing.workbench.threatMitigationActive")}</span>
                  <span className="text-white bg-graphite-800 px-2.5 py-0.5 rounded border border-graphite-700 text-xs">MITRE T1046</span>
                </div>
              </div>
            )}
          </div>

          {/* Workbench Footer Actions */}
          {activeWorkbenchTab === "quiz" && question && (
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
          )}
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
            
            {/* Visual Loop diagram */}
            <div className="mt-8 pt-4 border-t border-graphite-800 flex flex-col gap-3 font-mono text-xs text-graphite-300 select-none">
              <div className="flex items-center gap-2.5">
                <span className="w-6 h-6 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium shrink-0">01</span>
                <span>{t("landing.bento.cell1.step1")}</span>
              </div>
              <div className="w-px h-3 bg-graphite-800 ml-3" />
              <div className="flex items-center gap-2.5">
                <span className="w-6 h-6 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium shrink-0">02</span>
                <span className="text-white font-medium">{t("landing.bento.cell1.step2")}</span>
              </div>
              <div className="w-px h-3 bg-graphite-800 ml-3" />
              <div className="flex items-center gap-2.5">
                <span className="w-6 h-6 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium shrink-0">03</span>
                <span>{t("landing.bento.cell1.step3")}</span>
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
