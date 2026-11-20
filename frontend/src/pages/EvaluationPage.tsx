import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { AlertTriangle, BookOpen, CheckCircle, ShieldAlert, Award, PlayCircle, Loader2 } from "lucide-react";
import DomainBreakdown from "../components/DomainBreakdown";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";

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
  const { t } = useI18n();
  const [loading, setLoading] = useState(true);
  const [domains, setDomains] = useState<DomainStatus[]>([]);
  const [completedScenarios, setCompletedScenarios] = useState(0);
  const [needsTraining, setNeedsTraining] = useState(false);
  const [theoryCompleted, setTheoryCompleted] = useState(false);
  const [completedTheoryDomains, setCompletedTheoryDomains] = useState<string[]>([]);
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

        const res = await axios.get("/api/simulation/status", {
          headers: { Authorization: `Bearer ${token}` }
        });

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
        <p className="text-xs font-mono text-graphite-500">{t("evaluation.loading")}</p>
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
        <h3 className="text-base font-bold text-white tracking-tight">{t("evaluation.assessmentRequiredTitle")}</h3>
        <p className="text-xs text-graphite-400 font-light leading-relaxed">{resolvedErrorMessage}</p>
        <button
          onClick={() => navigate("/pre-test")}
          className="w-full py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200"
        >
          {t("evaluation.takePreTest")}
        </button>
      </div>
    );
  }

  const failedDomains = domains.filter((d) => !d.passed);
  const averageScore = domains.reduce((acc, d) => acc + d.proficiency_score, 0) / Math.max(1, domains.length);

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-fade-in">
      <div className="text-center space-y-2">
        <h1 className="text-2xl font-bold tracking-tight text-white">{t("evaluation.title")}</h1>
        <p className="text-xs text-graphite-400 font-mono">{t("evaluation.subtitle")}</p>
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        {/* Left Side: Summary Card */}
        <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg flex flex-col items-center justify-center text-center space-y-3">
          <div className="text-4xl font-extrabold text-white font-mono tracking-tighter">
            {averageScore.toFixed(0)}%
          </div>
          <div className="text-[10px] font-semibold text-graphite-300 uppercase tracking-wider font-mono">
            {t("evaluation.averageProficiency")}
          </div>
          <p className="text-[10px] text-graphite-500 font-mono">
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
        <h3 className="text-sm font-semibold flex items-center gap-2 text-white border-b border-graphite-800 pb-3">
          <AlertTriangle className="text-amber-500 w-4 h-4" />
          {t("evaluation.skillGapAnalysis")}
        </h3>
        
        <ul className="space-y-4 text-xs text-graphite-400 font-light">
          {failedDomains.length > 0 ? (
            <>
              <li className="flex gap-3 items-start">
                <ShieldAlert className="text-amber-500 shrink-0 w-4 h-4 mt-0.5" />
                <span>
                  <strong className="text-white">{t("evaluation.identifiedWeaknessLabel")}</strong>{" "}
                  {t("evaluation.identifiedWeaknessBeforeThreshold")}
                  <strong className="text-white">70%</strong>
                  {t("evaluation.identifiedWeaknessAfterThreshold")}
                </span>
              </li>
              <li className="flex gap-3 items-start">
                <CheckCircle className="text-emerald-500 shrink-0 w-4 h-4 mt-0.5" />
                <span>
                  <strong className="text-white">{t("evaluation.adaptiveRoadmapLabel")}</strong> {t("evaluation.adaptiveRoadmapText")}
                </span>
              </li>
              {!theoryCompleted && (
                <li className="flex gap-3 items-start">
                  <BookOpen className="text-amber-500 shrink-0 w-4 h-4 mt-0.5" />
                  <span>
                    <strong className="text-white">{t("evaluation.requiredPreSimLabel")}</strong>{" "}
                    {t("evaluation.requiredPreSimBeforeLink")}
                    <button
                      onClick={() => navigate("/learn")}
                      className="text-white underline hover:text-emerald-400 transition-colors"
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
              <CheckCircle className="text-emerald-500 shrink-0 w-4 h-4 mt-0.5" />
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
                className="px-4 py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200 flex items-center gap-2"
              >
                <BookOpen className="w-4 h-4" />
                {t("evaluation.studyTheoryFirst", { completed: completedTheoryDomains.length })}
              </button>
            ) : (
              <button
                onClick={() => navigate("/simulation")}
                className="px-4 py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200 flex items-center gap-2"
              >
                <PlayCircle className="w-4 h-4" />
                {t("evaluation.startRecommendedScenarios", { completed: completedScenarios })}
              </button>
            )
          ) : (
            <button
              onClick={() => navigate("/simulation/complete")}
              className="px-4 py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200 flex items-center gap-2"
            >
              <Award className="w-4 h-4" />
              {t("evaluation.viewCompletionCertificate")}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
