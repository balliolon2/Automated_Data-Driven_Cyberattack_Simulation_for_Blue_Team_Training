import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { Terminal, Shield, Eye, CheckCircle2, ChevronRight, Loader2, AlertCircle, Sparkles, Search, X, Trash2, BookOpen } from "lucide-react";
import KqlSearchBar from "../components/KqlSearchBar";
import LogTable from "../components/LogTable";
import ResponsePanel from "../components/ResponsePanel";
import ScenarioLoadingState from "../components/ScenarioLoadingState";
import { buildInvestigationShortcuts } from "../lib/siemShortcuts";
import { useI18n } from "../i18n";
import { resolveApiError, resolveScenarioLocale } from "../i18n/format";
import type { LocalizableScenario } from "../i18n/types";
import { cn } from "../lib/utils";

interface ActionOption {
  id: string;
  action_id?: string;
  label: string;
}

interface PlaybookSteps {
  containment: ActionOption[];
  eradication: ActionOption[];
  recovery: ActionOption[];
}

interface Scenario {
  scenario_id: string;
  title: string;
  description: string;
  domain_id: string;
  difficulty: number;
  initial_logs: any[];
  playbook_steps: PlaybookSteps;
  total_findings?: number;
  translations?: LocalizableScenario["translations"];
  tp_fp_explanation?: string;
}

interface SimulationSession {
  session_id: string;
  user_id: string;
  scenario_id: string;
  status: string;
  generation_type?: string;
  fallback_reason?: string;
}

const hydrateLogs = (logs: any[]): any[] =>
  Array.isArray(logs)
    ? logs.map((log, index) => ({ ...log, __log_index: index }))
    : [];

const normalizeIndices = (indices: number[] | undefined | null): number[] =>
  Array.from(new Set(Array.isArray(indices) ? indices : [])).sort((a, b) => a - b);

export default function SimulationPage() {
  const { locale, t } = useI18n();
  const [session, setSession] = useState<SimulationSession | null>(null);
  const [scenario, setScenario] = useState<Scenario | null>(null);
  const [filteredLogs, setFilteredLogs] = useState<any[]>([]);

  // User answers
  const [isTruePositive, setIsTruePositive] = useState<boolean | null>(null);
  const [selectedActions, setSelectedActions] = useState<string[]>([]);
  const [selectedLogIndices, setSelectedLogIndices] = useState<number[]>([]);
  const [selectionError, setSelectionError] = useState<string | null>(null);
  const selectionRef = useRef<number[]>([]);

  // Evidence mutations run through a per-session promise queue so optimistic
  // toggles cannot overwrite one another out of order. The final state is always
  // reconciled from the server once the queue drains.
  const selectionQueueRef = useRef<Promise<void>>(Promise.resolve());
  const pendingSelectionMutationsRef = useRef(0);
  const selectionMutationSeqRef = useRef(0);
  const lastConfirmedSelectionRef = useRef<number[]>([]);
  const selectionFailureRef = useRef(false);
  const activeSessionIdRef = useRef<string | null>(null);
  const submissionInFlightRef = useRef(false);

  const [backfillingTranslation, setBackfillingTranslation] = useState(false);
  const localeRef = useRef(locale);
  useEffect(() => {
    localeRef.current = locale;
  }, [locale]);
  const attemptedThSessionRef = useRef<string | null>(null);
  const initialFetchLocaleRef = useRef(locale);
  const tRef = useRef(t);
  useEffect(() => {
    tRef.current = t;
  }, [t]);

  // Navigation and UI state
  const [activeTab, setActiveTab] = useState<"scenario" | "detect" | "response">("scenario");
  const [siemQuery, setSiemQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [theoryGateInfo, setTheoryGateInfo] = useState<{
    completedCount: number;
    totalDomains: number;
    completedDomains: string[];
    completedSubtopics?: number;
    totalSubtopics?: number;
  } | null>(null);

  const applySelection = (indices: number[] | undefined | null) => {
    const normalized = normalizeIndices(indices);
    selectionRef.current = normalized;
    setSelectedLogIndices(normalized);
  };

  const adoptServerSelection = (indices: number[] | undefined | null) => {
    const normalized = normalizeIndices(indices);
    lastConfirmedSelectionRef.current = normalized;
    applySelection(normalized);
  };

  const fetchServerSelection = async (sessionID: string): Promise<number[]> => {
    const token = localStorage.getItem("token");
    const res = await axios.get("/api/simulation/session", {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (res.data?.type !== "in_progress" || res.data?.session?.session_id !== sessionID) {
      throw new Error("Simulation session changed while reconciling evidence");
    }
    return normalizeIndices(res.data.selected_log_indices);
  };

  const enqueueSelectionMutation = (task: () => Promise<void>) => {
    pendingSelectionMutationsRef.current += 1;
    const run = selectionQueueRef.current.then(task, task);
    selectionQueueRef.current = run.catch(() => undefined);
    return run;
  };

  const finalizeSelection = async (sessionID: string, seq: number) => {
    if (activeSessionIdRef.current !== sessionID) return;
    if (seq !== selectionMutationSeqRef.current) return;
    if (pendingSelectionMutationsRef.current !== 0) return;

    if (selectionFailureRef.current) {
      selectionFailureRef.current = false;
      try {
        const reconciled = await fetchServerSelection(sessionID);
        if (
          activeSessionIdRef.current !== sessionID ||
          seq !== selectionMutationSeqRef.current ||
          pendingSelectionMutationsRef.current !== 0
        ) {
          return;
        }
        adoptServerSelection(reconciled);
      } catch (err) {
        console.error("Failed to reconcile evidence selection:", err);
        applySelection(lastConfirmedSelectionRef.current);
      }
      setSelectionError(tRef.current("simulation.selectionRestoreError"));
      return;
    }

    // Reconcile the optimistic UI with the last server-confirmed selection.
    applySelection(lastConfirmedSelectionRef.current);
  };

  const handleToggleEvidence = (logIndex: number) => {
    if (!session || submissionInFlightRef.current) return;

    const sessionID = session.session_id;
    const wasSelected = selectionRef.current.includes(logIndex);
    applySelection(
      wasSelected
        ? selectionRef.current.filter((index) => index !== logIndex)
        : [...selectionRef.current, logIndex]
    );
    setSelectionError(null);

    const seq = ++selectionMutationSeqRef.current;
    void enqueueSelectionMutation(async () => {
      try {
        const token = localStorage.getItem("token");
        const res = await axios.post("/api/simulation/select-log", {
          session_id: sessionID,
          log_index: logIndex,
          selected: !wasSelected
        }, {
          headers: { Authorization: `Bearer ${token}` }
        });
        lastConfirmedSelectionRef.current = normalizeIndices(res.data.selected_log_indices);
      } catch (err) {
        selectionFailureRef.current = true;
        console.error("Failed to persist evidence selection:", err);
      } finally {
        pendingSelectionMutationsRef.current -= 1;
        await finalizeSelection(sessionID, seq);
      }
    });
  };

  const handleClearEvidence = () => {
    if (!session || submissionInFlightRef.current || selectionRef.current.length === 0) return;

    const sessionID = session.session_id;
    applySelection([]);
    setSelectionError(null);

    const seq = ++selectionMutationSeqRef.current;
    void enqueueSelectionMutation(async () => {
      try {
        const token = localStorage.getItem("token");
        const res = await axios.post("/api/simulation/select-log/clear", {
          session_id: sessionID
        }, {
          headers: { Authorization: `Bearer ${token}` }
        });
        lastConfirmedSelectionRef.current = normalizeIndices(res.data.selected_log_indices);
      } catch (err) {
        selectionFailureRef.current = true;
        console.error("Failed to clear evidence selection:", err);
      } finally {
        pendingSelectionMutationsRef.current -= 1;
        await finalizeSelection(sessionID, seq);
      }
    });
  };

  const handleQuickSiemSearch = (queryStr: string) => {
    setSiemQuery(queryStr);
    setActiveTab("detect");
  };
  const navigate = useNavigate();

  useEffect(() => {
    let cancelled = false;
    const initSimulation = async () => {
      initialFetchLocaleRef.current = localeRef.current;
      attemptedThSessionRef.current = null;
      let preTestVerified = false;
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          navigate("/login");
          return;
        }

        // 1. Check if there's already an active session
        try {
          const activeRes = await axios.get("/api/simulation/session", {
            headers: { Authorization: `Bearer ${token}` }
          });
          if (cancelled) return;

          if (activeRes.data.type === "completed") {
            const isAck = localStorage.getItem(`ack_session_${activeRes.data.session_id}`);
            if (isAck !== "true") {
              // If recently completed but not acknowledged, redirect to results
              navigate(`/simulation/result/${activeRes.data.session_id}`);
              return;
            }
          }

          if (activeRes.data.type === "in_progress") {
            const hydratedLogs = hydrateLogs(activeRes.data.scenario.initial_logs);
            const restoredSelection = normalizeIndices(activeRes.data.selected_log_indices);
            activeSessionIdRef.current = activeRes.data.session.session_id;
            lastConfirmedSelectionRef.current = restoredSelection;
            selectionRef.current = restoredSelection;
            setSession(activeRes.data.session);
            setScenario({ ...activeRes.data.scenario, initial_logs: hydratedLogs });
            setFilteredLogs(hydratedLogs);
            setSelectedLogIndices(restoredSelection);
            setLoading(false);
            return;
          }
        } catch (sessionErr) {
          // No active session, start a new one
          console.log("No active session, starting a new one...");
        }

        // 2. No session, query status to see if training is needed
        const statusRes = await axios.get("/api/simulation/status", {
          headers: { Authorization: `Bearer ${token}` }
        });
        if (cancelled) return;

        if (statusRes.data.error) {
          navigate("/pre-test");
          return;
        }
        preTestVerified = true;

        if (!statusRes.data.needs_training) {
          navigate("/simulation/complete");
          return;
        }

        // 3. Start a new scenario (allow full backend 120s LLM budget + retry headroom)
        const startRes = await axios.post("/api/simulation/start", {}, {
          headers: { Authorization: `Bearer ${token}` },
          timeout: 185000,
        });
        if (cancelled) return;

        if (startRes.data.status === "scenario_unavailable" || startRes.data.error) {
          setErrorMessage(
            startRes.data.error || tRef.current("simulation.tempInitError"),
          );
          setLoading(false);
          return;
        }

        if (startRes.data.completed) {
          navigate("/simulation/complete");
          return;
        }

        const hydratedLogs = hydrateLogs(startRes.data.scenario.initial_logs);
        const restoredSelection = normalizeIndices(startRes.data.selected_log_indices);
        activeSessionIdRef.current = startRes.data.session.session_id;
        lastConfirmedSelectionRef.current = restoredSelection;
        selectionRef.current = restoredSelection;
        setSession(startRes.data.session);
        setScenario({ ...startRes.data.scenario, initial_logs: hydratedLogs });
        setFilteredLogs(hydratedLogs);
        setSelectedLogIndices(restoredSelection);
      } catch (err: any) {
        if (cancelled) return;
        console.error("Failed to initialize simulation:", err);
        if (err.response?.status === 403 && err.response?.data?.reason === "theory_incomplete") {
          setTheoryGateInfo({
            completedCount: err.response.data.completed_count ?? 0,
            totalDomains: err.response.data.total_domains ?? 5,
            completedDomains: Array.isArray(err.response.data.completed_domains)
              ? err.response.data.completed_domains
              : [],
            completedSubtopics: err.response.data.completed_subtopic_count,
            totalSubtopics: err.response.data.total_subtopics,
          });
          setErrorMessage(
            resolveApiError(
              err,
              tRef.current,
              tRef.current("errors.SIM_THEORY_INCOMPLETE", {
                completed_count: err.response.data.completed_count ?? 0,
                total_domains: err.response.data.total_domains ?? 5,
              }),
            ),
          );
          return;
        }
        if (err.response?.status === 504 || err.code === "ECONNABORTED") {
          setErrorMessage(tRef.current("simulation.timeoutError"));
          return;
        }
        if (preTestVerified) {
          setErrorMessage(
            resolveApiError(err, tRef.current, tRef.current("simulation.tempInitError")),
          );
          return;
        }
        setErrorMessage(resolveApiError(err, tRef.current, tRef.current("simulation.genericInitError")));
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    initSimulation();
    return () => {
      cancelled = true;
    };
  }, [navigate, retryCount]);

  useEffect(() => {
    if (locale !== "th" || !session || !scenario || scenario.translations?.th?.description) {
      return;
    }
    const sid = session.session_id;
    if (!sid || attemptedThSessionRef.current === sid) {
      return;
    }
    attemptedThSessionRef.current = sid;

    // Initial load already ran under locale="th", so the backend already attempted any available backfill.
    if (initialFetchLocaleRef.current === "th") {
      return;
    }

    const token = localStorage.getItem("token");
    if (!token) return;

    let cancelled = false;
    Promise.resolve()
      .then(() => {
        if (cancelled) return null;
        setBackfillingTranslation(true);
        return axios.get("/api/simulation/session", {
          headers: { Authorization: `Bearer ${token}` },
        });
      })
      .then((res) => {
        if (
          !cancelled &&
          res?.data?.type === "in_progress" &&
          res?.data?.session?.session_id === sid &&
          res?.data?.scenario
        ) {
          setScenario(res.data.scenario);
        }
      })
      .catch((err) => {
        console.error("Failed to backfill Thai simulation translation:", err);
      })
      .finally(() => {
        if (!cancelled) {
          setBackfillingTranslation(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [locale, session, scenario]);

  const handleSubmit = async () => {
    if (!session || !scenario) return;
    if (isTruePositive === null) {
      alert(t("simulation.selectTriageFirstAlert"));
      setActiveTab("scenario");
      return;
    }

    submissionInFlightRef.current = true;
    setSubmitting(true);
    const token = localStorage.getItem("token");
    const sessionID = session.session_id;

    try {
      // Wait for every queued toggle/clear to commit before submitting.
      await selectionQueueRef.current;
      if (pendingSelectionMutationsRef.current > 0) {
        await selectionQueueRef.current;
      }

      // Reconcile from the server so the submitted evidence set is
      // server-authoritative even if a mutation failed or arrived out of order.
      let finalSelection = selectionRef.current;
      try {
        finalSelection = await fetchServerSelection(sessionID);
        adoptServerSelection(finalSelection);
      } catch (reconcileErr) {
        console.error("Failed to reconcile evidence before submit:", reconcileErr);
        finalSelection = lastConfirmedSelectionRef.current;
      }

      await axios.post("/api/simulation/submit", {
        session_id: sessionID,
        is_true_positive: isTruePositive,
        tp_fp_selected: true,
        selected_actions: selectedActions,
        selected_log_indices: finalSelection
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });

      navigate(`/simulation/result/${sessionID}`);
    } catch (err: any) {
      console.error("Failed to submit scenario:", err);
      alert(resolveApiError(err, t, t("simulation.submitFailedAlert")));
    } finally {
      submissionInFlightRef.current = false;
      setSubmitting(false);
    }
  };

  const selectedLogs = scenario
    ? selectedLogIndices
        .map(index => scenario.initial_logs.find((log: any) => log.__log_index === index))
        .filter(Boolean)
    : [];

  if (loading) {
    return <ScenarioLoadingState />;
  }

  if (errorMessage) {
    if (theoryGateInfo) {
      return (
        <div className="max-w-md mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
          <BookOpen className="w-10 h-10 text-amber-500 mx-auto" />
          <div className="space-y-2">
            <h3 className="text-lg font-bold text-white tracking-tight">{t("simulation.theoryRequiredTitle")}</h3>
            <p className="text-sm text-graphite-300 font-mono leading-relaxed">
              {t("errors.SIM_THEORY_INCOMPLETE", {
                completed_count: theoryGateInfo.completedCount,
                total_domains: theoryGateInfo.totalDomains,
              }) || errorMessage}
            </p>
            <p className="text-sm font-mono text-graphite-200">
              {t("simulation.completedDomainsSummary", {
                completed: theoryGateInfo.completedCount,
                total: theoryGateInfo.totalDomains,
              })}
            </p>
          </div>
          <div className="flex flex-col gap-2">
            <button
              onClick={() => navigate("/learn")}
              className="w-full py-2.5 bg-white hover:bg-white/90 text-black font-semibold text-sm rounded-md transition-all duration-200 flex items-center justify-center gap-2 cursor-pointer"
            >
              <BookOpen className="w-4 h-4" />
              {t("simulation.goToLearnHub")}
            </button>
            <button
              onClick={() => navigate("/evaluation")}
              className="w-full py-2.5 bg-graphite-800 hover:bg-graphite-700 text-graphite-200 font-semibold text-sm rounded-md transition-all duration-200 cursor-pointer"
            >
              {t("simulation.backToEvaluation")}
            </button>
          </div>
        </div>
      );
    }

    return (
      <div className="max-w-md mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
        <AlertCircle className="w-10 h-10 text-amber-500 mx-auto" />
        <h3 className="text-lg font-bold text-white tracking-tight">{t("nav.simulation")}</h3>
        <p className="text-sm text-graphite-300 font-mono leading-relaxed">{errorMessage}</p>
        <div className="flex flex-col gap-2">
          <button
            onClick={() => {
              setErrorMessage(null);
              setTheoryGateInfo(null);
              setLoading(true);
              setRetryCount((prev) => prev + 1);
            }}
            className="w-full py-2.5 bg-emerald-500 hover:bg-emerald-600 text-black font-semibold text-sm rounded-md transition-all duration-200 cursor-pointer"
          >
            {t("simulation.retryLoading")}
          </button>
          <button
            onClick={() => navigate("/pre-test")}
            className="w-full py-2.5 bg-graphite-800 hover:bg-graphite-700 text-graphite-200 font-semibold text-sm rounded-md transition-all duration-200 cursor-pointer"
          >
            {t("simulation.goToPreTestOverview")}
          </button>
        </div>
      </div>
    );
  }

  if (!scenario || !session) return null;

  const localizedScenario = resolveScenarioLocale(scenario, locale);
  const localizePhaseSteps = (steps?: ActionOption[]) =>
    (steps || []).map((act) => {
      const stepId = act.id || act.action_id || "";
      return {
        ...act,
        id: stepId,
        label: localizedScenario.playbookLabel(stepId, act.label),
      };
    });
  const localizedPlaybookSteps: PlaybookSteps = {
    containment: localizePhaseSteps(scenario.playbook_steps?.containment),
    eradication: localizePhaseSteps(scenario.playbook_steps?.eradication),
    recovery: localizePhaseSteps(scenario.playbook_steps?.recovery),
  };

  // Canonical EN title + description MUST be passed so keyword/entity scoring behaves identically across languages.
  const shortcuts = buildInvestigationShortcuts({
    title: scenario.title,
    description: scenario.description,
    initial_logs: scenario.initial_logs,
  });

  return (
    <div className="space-y-6">
      {/* Simulation Header */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-graphite-800 pb-5">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2 mb-2">
            <span className="px-2.5 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20 text-xs font-mono font-bold text-emerald-400 uppercase tracking-wider">
              {t("simulation.activeIncidentBadge")}
            </span>

            {session.generation_type === "ai_generated" ? (
              <span className="px-2.5 py-0.5 rounded bg-cyan-500/10 border border-cyan-500/30 text-xs font-mono font-bold text-cyan-400 uppercase tracking-wider flex items-center gap-1.5 shadow-sm shadow-cyan-500/10">
                <Sparkles className="w-3.5 h-3.5 text-cyan-400 animate-pulse" />
                {t("simulation.aiScenarioBadge")}
              </span>
            ) : (
              <span
                title={
                  session.fallback_reason
                    ? t("simulation.fallbackReasonTooltip", { reason: session.fallback_reason })
                    : t("simulation.staticPoolTooltip")
                }
                className="px-2.5 py-0.5 rounded bg-amber-500/10 border border-amber-500/30 text-xs font-mono font-bold text-amber-400 uppercase tracking-wider flex items-center gap-1.5 cursor-help hover:bg-amber-500/20 transition-all duration-200"
              >
                <AlertCircle className="w-3.5 h-3.5 text-amber-400" />
                {t("simulation.staticScenarioBadge")}
              </span>
            )}

            <span className="text-xs font-mono text-graphite-400">
              {t("simulation.sessionIdPrefix", { id: session.session_id.substring(0, 8) })}
            </span>
          </div>

          {session.generation_type !== "ai_generated" && session.fallback_reason && (
            <p className="text-xs font-mono text-amber-300/90 mb-1.5 flex items-center gap-1.5">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-amber-400"></span>
              <span className="font-semibold">{t("simulation.fallbackReasonLabel")}</span> {session.fallback_reason}
            </p>
          )}

          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-graphite-50 break-words max-w-full" title={localizedScenario.title}>
            {localizedScenario.title}
          </h1>
        </div>

        {/* Tab Navigation */}
        <div className="flex bg-graphite-900/50 p-1 rounded-lg border border-graphite-800/80 font-mono text-sm w-full md:w-auto shrink-0">
          <button
            onClick={() => setActiveTab("scenario")}
            className={cn(
              "flex-1 md:flex-none flex items-center justify-center gap-2 px-4 py-2 rounded-md font-medium transition-all duration-200 cursor-pointer",
              activeTab === "scenario"
                ? "bg-graphite-800 text-white shadow-sm"
                : "text-graphite-400 hover:text-graphite-200"
            )}
          >
            <Eye className="w-4 h-4" />
            {t("simulation.tabScenario")}
          </button>
          <button
            onClick={() => setActiveTab("detect")}
            className={cn(
              "flex-1 md:flex-none flex items-center justify-center gap-2 px-4 py-2 rounded-md font-medium transition-all duration-200 cursor-pointer",
              activeTab === "detect"
                ? "bg-graphite-800 text-white shadow-sm"
                : "text-graphite-400 hover:text-graphite-200"
            )}
          >
            <Terminal className="w-4 h-4" />
            {t("simulation.tabDetect")}
          </button>
          <button
            onClick={() => setActiveTab("response")}
            className={cn(
              "flex-1 md:flex-none flex items-center justify-center gap-2 px-4 py-2 rounded-md font-medium transition-all duration-200 cursor-pointer",
              activeTab === "response"
                ? "bg-graphite-800 text-white shadow-sm"
                : "text-graphite-400 hover:text-graphite-200"
            )}
          >
            <Shield className="w-4 h-4" />
            {t("simulation.tabResponse")}
          </button>
        </div>
      </div>

      {/* Main Panel Content */}
      <div className="min-h-[50vh]">
        {/* Panel 1: Scenario Narrative & TP/FP selection */}
        {activeTab === "scenario" && (
          <div className="grid md:grid-cols-3 gap-6 animate-fade-in">
            <div className="md:col-span-2 space-y-6">
              <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
                <h3 className="text-lg font-semibold text-white">{t("simulation.incidentNarrative")}</h3>
                {locale === "th" && !scenario.translations?.th?.description && !backfillingTranslation && (
                  <p className="text-xs text-amber-300/90 font-mono">
                    {t("simulation.legacyUntranslatedNotice")}
                  </p>
                )}
                <div className="space-y-4 text-base text-graphite-200 leading-relaxed">
                  {localizedScenario.description
                    .split(/\n{2,}/)
                    .map((para) => para.trim())
                    .filter(Boolean)
                    .map((para, idx) => (
                      <p key={idx}>{para}</p>
                    ))}
                </div>
              </div>

              {/* Investigation Clues & SIEM Shortcuts */}
              <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
                <div className="flex flex-wrap items-center justify-between gap-2 border-b border-graphite-800 pb-3">
                  <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                    <Search className="w-4 h-4 text-emerald-400" />
                    {t("simulation.cluesAndShortcutsTitle")}
                  </h3>
                  <span className="text-xs font-mono text-graphite-400">{t("simulation.cluesClickHint")}</span>
                </div>

                <p className="text-sm text-graphite-300 font-normal leading-relaxed">
                  {t("simulation.cluesDesc")}
                </p>

                <div className="flex flex-wrap gap-2.5 pt-1">
                  {shortcuts.map((shortcut) => {
                    const toneClasses =
                      shortcut.tone === "amber"
                        ? "hover:border-amber-500/40 text-amber-400"
                        : shortcut.tone === "sky"
                        ? "hover:border-sky-500/40 text-sky-400"
                        : "hover:border-emerald-500/40 text-emerald-400";
                    return (
                      <button
                        key={`${shortcut.labelPrefix}:${shortcut.query}`}
                        onClick={() => handleQuickSiemSearch(shortcut.query)}
                        className={cn(
                          "px-3.5 py-2 rounded bg-graphite-950 border border-graphite-800 font-mono text-log flex items-center gap-1.5 transition-all cursor-pointer",
                          toneClasses
                        )}
                      >
                        <span className="text-graphite-400 text-xs">{shortcut.labelPrefix}</span>
                        {shortcut.label}
                      </button>
                    );
                  })}
                </div>
              </div>
            </div>

            <div className="space-y-6">
              <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
                <h3 className="text-lg font-semibold text-white">{t("simulation.triageClassification")}</h3>
                <p className="text-sm text-graphite-300 font-normal leading-relaxed">
                  {t("simulation.triagePrompt")}
                </p>

                <div className="space-y-3 pt-2">
                  <button
                    onClick={() => setIsTruePositive(true)}
                    className={cn(
                      "w-full text-left p-4 rounded border font-mono text-sm transition-all duration-200 flex items-center justify-between cursor-pointer",
                      isTruePositive === true
                        ? "bg-red-950/20 border-red-500/30 text-red-200 font-semibold"
                        : "bg-graphite-950/40 border-graphite-800/80 text-graphite-200 hover:bg-graphite-950 hover:text-white"
                    )}
                  >
                    <span>{t("simulation.truePositiveOption")}</span>
                    <span className={cn(
                      "w-3 h-3 rounded-full border border-graphite-600 shrink-0",
                      isTruePositive === true && "bg-red-500 border-red-400"
                    )} />
                  </button>

                  <button
                    onClick={() => setIsTruePositive(false)}
                    className={cn(
                      "w-full text-left p-4 rounded border font-mono text-sm transition-all duration-200 flex items-center justify-between cursor-pointer",
                      isTruePositive === false
                        ? "bg-emerald-950/20 border-emerald-500/30 text-emerald-200 font-semibold"
                        : "bg-graphite-950/40 border-graphite-800/80 text-graphite-200 hover:bg-graphite-950 hover:text-white"
                    )}
                  >
                    <span>{t("simulation.falsePositiveOption")}</span>
                    <span className={cn(
                      "w-3 h-3 rounded-full border border-graphite-600 shrink-0",
                      isTruePositive === false && "bg-emerald-500 border-emerald-400"
                    )} />
                  </button>
                </div>
              </div>

              <div className="bg-graphite-900 border border-graphite-800/80 p-4 rounded-lg flex items-center justify-between gap-3">
                <div className="space-y-0.5">
                  <h4 className="text-xs font-semibold text-white font-mono uppercase tracking-wider">{t("simulation.nextStepLabel")}</h4>
                  <p className="text-xs text-graphite-300 font-normal">{t("simulation.nextStepDesc")}</p>
                </div>
                <button
                  onClick={() => setActiveTab("detect")}
                  className="p-2 rounded border border-graphite-800 bg-transparent text-graphite-200 hover:bg-graphite-950 hover:text-white transition-colors cursor-pointer shrink-0"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Panel 2: Detect & Analyze (KQL + Log Table) */}
        {activeTab === "detect" && (
          <div className="space-y-6 animate-fade-in">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div className="flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-emerald-500" />
                <h3 className="text-lg font-semibold text-white">{t("simulation.siemTerminalTitle")}</h3>
              </div>
              <div className="text-xs font-mono text-graphite-300 bg-graphite-900 px-3 py-1.5 rounded border border-graphite-800/80 select-none">
                {t("simulation.selectedEvidenceCount", { count: selectedLogIndices.length })}
              </div>
            </div>

            <p className="text-sm font-mono text-graphite-300 leading-relaxed">
              {t("simulation.evidenceInstruction")}
            </p>

            {typeof scenario.total_findings === "number" && (
              <p className="text-xs font-mono text-graphite-400">
                {t("simulation.keyFindingsCount")}{" "}
                <span className="text-graphite-200 font-semibold">{scenario.total_findings}</span>
              </p>
            )}

            <KqlSearchBar
              sessionId={session.session_id}
              logs={scenario.initial_logs}
              onFilter={setFilteredLogs}
              searchQuery={siemQuery}
              onSearchQueryChange={setSiemQuery}
            />

            {locale === "th" && (
              <p className="text-xs font-mono text-sky-300 bg-sky-950/20 border border-sky-500/20 rounded-lg px-3.5 py-2">
                {t("simulation.logsEnglishOnlyNotice")}
              </p>
            )}

            <LogTable
              logs={filteredLogs}
              selectedIndices={selectedLogIndices}
              onToggle={handleToggleEvidence}
            />

            {selectionError && (
              <div className="flex items-center gap-2 text-amber-400 text-sm font-mono bg-amber-950/20 px-3.5 py-2 rounded-lg border border-amber-500/20">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{selectionError}</span>
              </div>
            )}

            <div className="bg-graphite-900 border border-cyan-500/25 rounded-xl overflow-hidden">
              <div className="px-5 py-3 bg-cyan-950/20 border-b border-cyan-500/20 flex items-center justify-between gap-3">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-semibold text-cyan-200">{t("simulation.selectedEvidenceTitle")}</span>
                  <span className="text-xs font-mono text-cyan-300 bg-cyan-500/10 border border-cyan-500/20 rounded px-2 py-0.5">
                    {selectedLogIndices.length}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={handleClearEvidence}
                  disabled={selectedLogIndices.length === 0}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded border border-cyan-500/30 text-xs font-mono font-bold uppercase tracking-wider text-cyan-300 hover:bg-cyan-500/10 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  {t("simulation.clearAll")}
                </button>
              </div>

              {selectedLogs.length === 0 ? (
                <p className="px-5 py-4 text-sm font-mono text-graphite-400">
                  {t("simulation.noEvidenceSelected")}
                </p>
              ) : (
                <ul className="divide-y divide-graphite-800/60 max-h-64 overflow-y-auto">
                  {selectedLogs.map((log: any) => (
                    <li key={log.__log_index} className="flex items-start gap-3 px-5 py-3">
                      <div className="min-w-0 flex-1 font-mono space-y-1">
                        <div className="flex flex-wrap items-center gap-2 text-xs text-graphite-400">
                          <span className="text-cyan-300">#{log.__log_index}</span>
                          <span>{log.timestamp}</span>
                          <span className="text-cyan-300/90">{log.source}</span>
                          <span className="uppercase font-bold text-amber-400">{log.severity}</span>
                        </div>
                        <p className="text-log text-graphite-200 truncate" title={String(log.message ?? "")}>
                          {String(log.message ?? "")}
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={() => handleToggleEvidence(log.__log_index)}
                        aria-label={t("simulation.removeEvidenceAria", { index: log.__log_index })}
                        className="shrink-0 p-1.5 rounded border border-graphite-800 text-graphite-400 hover:text-cyan-300 hover:border-cyan-500/40 hover:bg-cyan-500/10 transition-colors cursor-pointer"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        )}

        {/* Panel 3: Response Actions & Submit */}
        {activeTab === "response" && (
          <div className="grid md:grid-cols-3 gap-6 animate-fade-in">
            <div className="md:col-span-2">
              <ResponsePanel
                playbookSteps={localizedPlaybookSteps}
                selectedActions={selectedActions}
                onChange={setSelectedActions}
                isTruePositive={isTruePositive}
                onSelectTriage={setIsTruePositive}
              />
            </div>

            <div className="space-y-6">
              <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
                <h3 className="text-lg font-semibold text-white">{t("simulation.finalSubmissionTitle")}</h3>
                <p className="text-sm text-graphite-300 font-normal leading-relaxed">
                  {t("simulation.finalSubmissionDesc")}
                </p>

                <div className="space-y-3 pt-2">
                  <div className="flex justify-between text-xs font-mono border-b border-graphite-800 pb-2">
                    <span className="text-graphite-400">{t("simulation.summaryTriageLabel")}</span>
                    {isTruePositive === null ? (
                      <span className="text-amber-400 font-bold">{t("simulation.summaryMissing")}</span>
                    ) : (
                      <span className="text-emerald-400 font-bold">
                        {isTruePositive ? t("simulation.summaryTruePositive") : t("simulation.summaryFalsePositive")}
                      </span>
                    )}
                  </div>
                  <div className="flex justify-between text-xs font-mono border-b border-graphite-800 pb-2">
                    <span className="text-graphite-400">{t("simulation.summaryResponsePlanLabel")}</span>
                    <span className="text-white font-bold">
                      {t("simulation.summaryActionsCount", { count: selectedActions.length })}
                    </span>
                  </div>
                  <div className="flex justify-between text-xs font-mono border-b border-graphite-800 pb-2">
                    <span className="text-graphite-400">{t("simulation.summaryEvidenceLabel")}</span>
                    <span className="text-white font-bold">
                      {t("simulation.summaryLogsCount", { count: selectedLogIndices.length })}
                    </span>
                  </div>
                </div>

                {selectedLogIndices.length === 0 && (
                  <p className="text-xs font-mono text-amber-300 bg-amber-950/20 border border-amber-500/20 rounded px-3 py-2 leading-relaxed">
                    {t("simulation.noEvidenceWarning")}
                  </p>
                )}

                <button
                  onClick={handleSubmit}
                  disabled={submitting}
                  className="w-full mt-4 py-2.5 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 disabled:bg-graphite-800 disabled:text-graphite-400 transition-all duration-200 flex items-center justify-center gap-2 cursor-pointer"
                >
                  {submitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      {t("simulation.submittingTriage")}
                    </>
                  ) : (
                    <>
                      <CheckCircle2 className="w-4 h-4" />
                      {t("simulation.submitResponsePlan")}
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
