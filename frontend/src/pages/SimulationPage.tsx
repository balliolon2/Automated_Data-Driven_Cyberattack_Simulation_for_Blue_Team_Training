import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { Terminal, Shield, Eye, CheckCircle2, ChevronRight, Loader2, AlertCircle, Sparkles, Search } from "lucide-react";
import KqlSearchBar from "../components/KqlSearchBar";
import LogTable from "../components/LogTable";
import ResponsePanel from "../components/ResponsePanel";
import ScenarioLoadingState from "../components/ScenarioLoadingState";
import { cn } from "../lib/utils";

interface ActionOption {
  id: string;
  label: string;
}

interface PlaybookSteps {
  containment: ActionOption[];
  eradication: ActionOption[];
  recovery: ActionOption[];
}

interface SanitizedFinding {
  id: string;
  evidence_log_indices: number[];
}

interface Scenario {
  scenario_id: string;
  title: string;
  description: string;
  domain_id: string;
  difficulty: number;
  initial_logs: any[];
  playbook_steps: PlaybookSteps;
  sanitized_findings: SanitizedFinding[];
}

interface SimulationSession {
  session_id: string;
  user_id: string;
  scenario_id: string;
  status: string;
  generation_type?: string;
  fallback_reason?: string;
}

export default function SimulationPage() {
  const [session, setSession] = useState<SimulationSession | null>(null);
  const [scenario, setScenario] = useState<Scenario | null>(null);
  const [filteredLogs, setFilteredLogs] = useState<any[]>([]);
  
  // User answers
  const [isTruePositive, setIsTruePositive] = useState<boolean | null>(null);
  const [selectedActions, setSelectedActions] = useState<string[]>([]);
  const [discoveredFindings, setDiscoveredFindings] = useState<string[]>([]);

  // Navigation and UI state
  const [activeTab, setActiveTab] = useState<"scenario" | "detect" | "response">("scenario");
  const [siemQuery, setSiemQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleQuickSiemSearch = (queryStr: string) => {
    setSiemQuery(queryStr);
    setActiveTab("detect");
  };
  const navigate = useNavigate();

  useEffect(() => {
    const initSimulation = async () => {
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
          
          if (activeRes.data.type === "completed") {
            const isAck = localStorage.getItem(`ack_session_${activeRes.data.session_id}`);
            if (isAck !== "true") {
              // If recently completed but not acknowledged, redirect to results
              navigate(`/simulation/result/${activeRes.data.session_id}`);
              return;
            }
          }
          
          if (activeRes.data.type === "in_progress") {
            setSession(activeRes.data.session);
            setScenario(activeRes.data.scenario);
            setFilteredLogs(activeRes.data.scenario.initial_logs);
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

        if (statusRes.data.error) {
          navigate("/pre-test");
          return;
        }

        if (!statusRes.data.needs_training) {
          navigate("/simulation/complete");
          return;
        }

        // 3. Start a new scenario
        const startRes = await axios.post("/api/simulation/start", {}, {
          headers: { Authorization: `Bearer ${token}` }
        });

        if (startRes.data.error) {
          setErrorMessage(startRes.data.error);
          setLoading(false);
          return;
        }

        if (startRes.data.completed) {
          navigate("/simulation/complete");
          return;
        }

        setSession(startRes.data.session);
        setScenario(startRes.data.scenario);
        setFilteredLogs(startRes.data.scenario.initial_logs);
      } catch (err: any) {
        console.error("Failed to initialize simulation:", err);
        setErrorMessage("An error occurred while loading the simulation. Please ensure your pre-test is completed.");
      } finally {
        setLoading(false);
      }
    };

    initSimulation();
  }, [navigate]);

  // Evaluate matching key findings whenever filtered logs change
  useEffect(() => {
    if (!scenario || !scenario.sanitized_findings || scenario.sanitized_findings.length === 0) return;

    // A finding is discovered if the filtered logs contain ALL of its evidence log indices.
    // To do this, we need to map the current filtered logs back to their original indexes in scenario.initial_logs.
    const filteredOriginalIndices = new Set<number>();
    
    filteredLogs.forEach(log => {
      // Find the index of this log entry in the original logs array
      const origIndex = scenario.initial_logs.findIndex(origLog => 
        JSON.stringify(origLog) === JSON.stringify(log)
      );
      if (origIndex !== -1) {
        filteredOriginalIndices.add(origIndex);
      }
    });

    const newlyDiscovered: string[] = [];

    scenario.sanitized_findings.forEach(finding => {
      // Guard: skip findings with no evidence indices defined
      if (!finding.evidence_log_indices || finding.evidence_log_indices.length === 0) return;
      
      const allEvidencePresent = finding.evidence_log_indices.every(idx => 
        filteredOriginalIndices.has(idx)
      );
      if (allEvidencePresent) {
        newlyDiscovered.push(finding.id);
      }
    });

    // Merge with previously discovered findings so we don't lose progress if user runs a different query
    setDiscoveredFindings(prev => {
      const merged = new Set([...prev, ...newlyDiscovered]);
      return Array.from(merged);
    });

  }, [filteredLogs, scenario]);

  const handleSubmit = async () => {
    if (!session || !scenario) return;
    if (isTruePositive === null) {
      alert("Please select whether this incident is a True Positive or False Positive in the Scenario panel first.");
      setActiveTab("scenario");
      return;
    }

    setSubmitting(true);
    const token = localStorage.getItem("token");

    try {
      await axios.post("/api/simulation/submit", {
        session_id: session.session_id,
        is_true_positive: isTruePositive,
        tp_fp_selected: true,
        selected_actions: selectedActions,
        discovered_findings: discoveredFindings
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });

      navigate(`/simulation/result/${session.session_id}`);
    } catch (err) {
      console.error("Failed to submit scenario:", err);
      alert("Failed to submit scenario. Please try again.");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <ScenarioLoadingState />;
  }

  if (errorMessage) {
    return (
      <div className="max-w-md mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
        <AlertCircle className="w-10 h-10 text-amber-500 mx-auto" />
        <h3 className="text-base font-bold text-white tracking-tight">Simulation Initialization</h3>
        <p className="text-xs text-graphite-400 font-mono leading-relaxed">{errorMessage}</p>
        <div className="flex flex-col gap-2">
          <button
            onClick={() => window.location.reload()}
            className="w-full py-2.5 bg-emerald-500 hover:bg-emerald-600 text-black font-semibold text-xs rounded-md transition-all duration-200"
          >
            Retry Loading Simulation
          </button>
          <button
            onClick={() => navigate("/pre-test")}
            className="w-full py-2 bg-graphite-800 hover:bg-graphite-700 text-graphite-300 font-semibold text-xs rounded-md transition-all duration-200"
          >
            Go to Pre-Test Overview
          </button>
        </div>
      </div>
    );
  }

  if (!scenario || !session) return null;

  return (
    <div className="space-y-6">
      {/* Simulation Header */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-graphite-800 pb-5">
        <div>
          <div className="flex flex-wrap items-center gap-2 mb-1.5">
            <span className="px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20 text-[10px] font-mono font-bold text-emerald-400 uppercase tracking-wider">
              Active Incident
            </span>

            {session.generation_type === "ai_generated" ? (
              <span className="px-2 py-0.5 rounded bg-cyan-500/10 border border-cyan-500/30 text-[10px] font-mono font-bold text-cyan-400 uppercase tracking-wider flex items-center gap-1.5 shadow-sm shadow-cyan-500/10">
                <Sparkles className="w-3 h-3 text-cyan-400 animate-pulse" />
                AI Dynamic Scenario
              </span>
            ) : (
              <span
                title={session.fallback_reason ? `Fallback Reason: ${session.fallback_reason}` : "Static seed scenario pool loaded"}
                className="px-2 py-0.5 rounded bg-amber-500/10 border border-amber-500/30 text-[10px] font-mono font-bold text-amber-400 uppercase tracking-wider flex items-center gap-1.5 cursor-help hover:bg-amber-500/20 transition-all duration-200"
              >
                <AlertCircle className="w-3 h-3 text-amber-400" />
                Static Scenario (Fallback)
              </span>
            )}

            <span className="text-xs font-mono text-graphite-500">
              ID: {session.session_id.substring(0, 8)}
            </span>
          </div>

          {session.generation_type !== "ai_generated" && session.fallback_reason && (
            <p className="text-[11px] font-mono text-amber-400/80 mb-1 flex items-center gap-1.5">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-amber-400"></span>
              <span className="font-semibold">LLM Fallback Reason:</span> {session.fallback_reason}
            </p>
          )}

          <h1 className="text-2xl font-bold tracking-tight text-graphite-50">
            {scenario.title}
          </h1>
        </div>

        {/* Tab Navigation */}
        <div className="flex bg-graphite-900/50 p-1 rounded-lg border border-graphite-800/80 font-mono text-xs w-full md:w-auto">
          <button
            onClick={() => setActiveTab("scenario")}
            className={cn(
              "flex-1 md:flex-none flex items-center justify-center gap-2 px-4 py-2 rounded-md font-medium transition-all duration-200",
              activeTab === "scenario"
                ? "bg-graphite-800 text-white shadow-sm"
                : "text-graphite-400 hover:text-graphite-200"
            )}
          >
            <Eye className="w-3.5 h-3.5" />
            1. Scenario
          </button>
          <button
            onClick={() => setActiveTab("detect")}
            className={cn(
              "flex-1 md:flex-none flex items-center justify-center gap-2 px-4 py-2 rounded-md font-medium transition-all duration-200",
              activeTab === "detect"
                ? "bg-graphite-800 text-white shadow-sm"
                : "text-graphite-400 hover:text-graphite-200"
            )}
          >
            <Terminal className="w-3.5 h-3.5" />
            2. Detect & Analyze
          </button>
          <button
            onClick={() => setActiveTab("response")}
            className={cn(
              "flex-1 md:flex-none flex items-center justify-center gap-2 px-4 py-2 rounded-md font-medium transition-all duration-200",
              activeTab === "response"
                ? "bg-graphite-800 text-white shadow-sm"
                : "text-graphite-400 hover:text-graphite-200"
            )}
          >
            <Shield className="w-3.5 h-3.5" />
            3. Response
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
                <h3 className="text-sm font-semibold text-white">Incident Narrative</h3>
                <p className="text-xs text-graphite-300 leading-relaxed font-light whitespace-pre-wrap">
                  {scenario.description}
                </p>
              </div>

              {/* Investigation Clues & SIEM Shortcuts */}
              <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
                <div className="flex items-center justify-between border-b border-graphite-800 pb-3">
                  <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                    <Search className="w-4 h-4 text-emerald-400" />
                    Investigation Clues & SIEM Shortcuts
                  </h3>
                  <span className="text-[10px] font-mono text-graphite-400">Click to query SIEM Terminal</span>
                </div>
                
                <p className="text-xs text-graphite-400 font-light">
                  Use these identified indicators to filter and audit the telemetry logs in the SIEM terminal:
                </p>

                <div className="flex flex-wrap gap-2 pt-1">
                  {/* Extract clues dynamically from initial_logs if present */}
                  {Array.isArray(scenario.initial_logs) && scenario.initial_logs.find((l: any) => l.host) && (
                    <button
                      onClick={() => handleQuickSiemSearch(scenario.initial_logs.find((l: any) => l.host)?.host)}
                      className="px-3 py-1.5 rounded bg-graphite-950 border border-graphite-800 hover:border-emerald-500/40 font-mono text-xs text-emerald-400 flex items-center gap-1.5 transition-all"
                    >
                      <span className="text-graphite-500 text-[10px]">Host:</span>
                      {scenario.initial_logs.find((l: any) => l.host)?.host}
                    </button>
                  )}

                  {Array.isArray(scenario.initial_logs) && scenario.initial_logs.find((l: any) => l.dst_ip || l.src_ip) && (
                    <button
                      onClick={() => handleQuickSiemSearch(scenario.initial_logs.find((l: any) => l.dst_ip)?.dst_ip || scenario.initial_logs.find((l: any) => l.src_ip)?.src_ip)}
                      className="px-3 py-1.5 rounded bg-graphite-950 border border-graphite-800 hover:border-emerald-500/40 font-mono text-xs text-emerald-400 flex items-center gap-1.5 transition-all"
                    >
                      <span className="text-graphite-500 text-[10px]">IP:</span>
                      {scenario.initial_logs.find((l: any) => l.dst_ip)?.dst_ip || scenario.initial_logs.find((l: any) => l.src_ip)?.src_ip}
                    </button>
                  )}

                  <button
                    onClick={() => handleQuickSiemSearch('severity == "high"')}
                    className="px-3 py-1.5 rounded bg-graphite-950 border border-graphite-800 hover:border-amber-500/40 font-mono text-xs text-amber-400 flex items-center gap-1.5 transition-all"
                  >
                    <span className="text-graphite-500 text-[10px]">Filter:</span>
                    High Severity Alerts
                  </button>

                  <button
                    onClick={() => handleQuickSiemSearch('source == "firewall"')}
                    className="px-3 py-1.5 rounded bg-graphite-950 border border-graphite-800 hover:border-sky-500/40 font-mono text-xs text-sky-400 flex items-center gap-1.5 transition-all"
                  >
                    <span className="text-graphite-500 text-[10px]">Filter:</span>
                    Firewall Logs
                  </button>
                </div>
              </div>
            </div>

            <div className="space-y-6">
              <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
                <h3 className="text-sm font-semibold text-white">Triage Classification</h3>
                <p className="text-[11px] text-graphite-400 font-light">
                  Assess the alert. Is this activity malicious (True Positive) or is there a legitimate business explanation (False Positive)?
                </p>
                
                <div className="space-y-3 pt-2">
                  <button
                    onClick={() => setIsTruePositive(true)}
                    className={cn(
                      "w-full text-left p-4 rounded border font-mono text-xs transition-all duration-200 flex items-center justify-between",
                      isTruePositive === true
                        ? "bg-red-950/20 border-red-500/30 text-red-200 font-medium"
                        : "bg-graphite-950/40 border-graphite-800/80 text-graphite-300 hover:bg-graphite-950 hover:text-white"
                    )}
                  >
                    <span>True Positive (Threat)</span>
                    <span className={cn(
                      "w-2.5 h-2.5 rounded-full border border-graphite-600",
                      isTruePositive === true && "bg-red-500 border-red-400"
                    )} />
                  </button>

                  <button
                    onClick={() => setIsTruePositive(false)}
                    className={cn(
                      "w-full text-left p-4 rounded border font-mono text-xs transition-all duration-200 flex items-center justify-between",
                      isTruePositive === false
                        ? "bg-emerald-950/20 border-emerald-500/30 text-emerald-200 font-medium"
                        : "bg-graphite-950/40 border-graphite-800/80 text-graphite-300 hover:bg-graphite-950 hover:text-white"
                    )}
                  >
                    <span>False Positive (Benign)</span>
                    <span className={cn(
                      "w-2.5 h-2.5 rounded-full border border-graphite-600",
                      isTruePositive === false && "bg-emerald-500 border-emerald-400"
                    )} />
                  </button>
                </div>
              </div>

              <div className="bg-graphite-900 border border-graphite-800/80 p-4 rounded-lg flex items-center justify-between">
                <div className="space-y-0.5">
                  <h4 className="text-[10px] font-semibold text-white font-mono uppercase tracking-wider">Next step:</h4>
                  <p className="text-[10px] text-graphite-400 font-light">Run queries to audit the log files.</p>
                </div>
                <button
                  onClick={() => setActiveTab("detect")}
                  className="p-1.5 rounded border border-graphite-800 bg-transparent text-graphite-300 hover:bg-graphite-950 hover:text-white transition-colors"
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
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                <h3 className="text-sm font-semibold text-white">SIEM Log Terminal</h3>
              </div>
              <div className="text-[10px] font-mono text-graphite-400 bg-graphite-900 px-3 py-1 rounded border border-graphite-800/80 select-none">
                Evidence Discovered: <span className="text-emerald-400 font-bold">{discoveredFindings.length}</span>
                <span className="text-graphite-600"> / {scenario.sanitized_findings?.length ?? 0}</span>
              </div>
            </div>

            <KqlSearchBar
              sessionId={session.session_id}
              logs={scenario.initial_logs}
              onFilter={setFilteredLogs}
              searchQuery={siemQuery}
              onSearchQueryChange={setSiemQuery}
            />

            <LogTable logs={filteredLogs} />
          </div>
        )}

        {/* Panel 3: Response Actions & Submit */}
        {activeTab === "response" && (
          <div className="grid md:grid-cols-3 gap-6 animate-fade-in">
            <div className="md:col-span-2">
              <ResponsePanel
                playbookSteps={scenario.playbook_steps}
                selectedActions={selectedActions}
                onChange={setSelectedActions}
                isTruePositive={isTruePositive}
                onSelectTriage={setIsTruePositive}
              />
            </div>

            <div className="space-y-6">
              <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
                <h3 className="text-sm font-semibold text-white">Final Submission</h3>
                <p className="text-[11px] text-graphite-400 font-light leading-relaxed">
                  Review your analysis. Ensure you have classified the triage correctly and selected response options for each phase before submitting.
                </p>

                <div className="space-y-3 pt-2">
                  <div className="flex justify-between text-[10px] font-mono border-b border-graphite-850 pb-2">
                    <span className="text-graphite-400">Triage Classification:</span>
                    {isTruePositive === null ? (
                      <span className="text-amber-500 font-bold">MISSING</span>
                    ) : (
                      <span className="text-emerald-400 font-bold">
                        {isTruePositive ? "True Positive" : "False Positive"}
                      </span>
                    )}
                  </div>
                  <div className="flex justify-between text-[10px] font-mono border-b border-graphite-850 pb-2">
                    <span className="text-graphite-400">Response Plan:</span>
                    <span className="text-white font-bold">
                      {selectedActions.length} actions
                    </span>
                  </div>
                  <div className="flex justify-between text-[10px] font-mono border-b border-graphite-850 pb-2">
                    <span className="text-graphite-400">Evidence Found:</span>
                    <span className="text-white font-bold">
                      {discoveredFindings.length} indicators
                    </span>
                  </div>
                </div>

                <button
                  onClick={handleSubmit}
                  disabled={submitting}
                  className="w-full mt-4 py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 disabled:bg-graphite-800 disabled:text-graphite-500 transition-all duration-200 flex items-center justify-center gap-2"
                >
                  {submitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Submitting Triage...
                    </>
                  ) : (
                    <>
                      <CheckCircle2 className="w-4 h-4" />
                      Submit Response Plan
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
