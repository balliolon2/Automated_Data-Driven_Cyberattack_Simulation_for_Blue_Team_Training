import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { Terminal, Shield, Eye, CheckCircle2, ChevronRight, Loader2, AlertCircle } from "lucide-react";
import KqlSearchBar from "../components/KqlSearchBar";
import LogTable from "../components/LogTable";
import ResponsePanel from "../components/ResponsePanel";
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
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

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
          setErrorMessage(statusRes.data.error);
          setLoading(false);
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
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-4">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
        <p className="text-sm font-mono text-graphite-400">Loading virtual environment...</p>
      </div>
    );
  }

  if (errorMessage) {
    return (
      <div className="max-w-md mx-auto mt-12 glass-panel p-6 rounded-xl border-red-500/20 text-center space-y-4">
        <AlertCircle className="w-12 h-12 text-red-400 mx-auto" />
        <h3 className="text-lg font-bold text-graphite-100">Setup Required</h3>
        <p className="text-sm text-graphite-300">{errorMessage}</p>
        <button
          onClick={() => navigate("/pre-test")}
          className="px-5 py-2.5 bg-emerald-500 text-emerald-950 font-bold rounded-lg hover:bg-emerald-400 transition-all text-sm"
        >
          Go to Pre-Test
        </button>
      </div>
    );
  }

  if (!scenario || !session) return null;

  return (
    <div className="space-y-6">
      {/* Simulation Header */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-graphite-800 pb-5">
        <div>
          <div className="flex items-center gap-2 mb-1.5">
            <span className="px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20 text-[10px] font-mono font-bold text-emerald-400 uppercase tracking-wider">
              Active Incident
            </span>
            <span className="text-xs font-mono text-graphite-500">
              ID: {session.session_id.substring(0, 8)}
            </span>
          </div>
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
              <div className="glass-panel p-6 rounded-xl space-y-4">
                <h3 className="text-lg font-bold text-graphite-100">Incident Narrative</h3>
                <p className="text-sm text-graphite-300 leading-relaxed whitespace-pre-wrap">
                  {scenario.description}
                </p>
              </div>
            </div>

            <div className="space-y-6">
              <div className="glass-panel p-6 rounded-xl space-y-4">
                <h3 className="text-base font-bold text-graphite-100">Triage Classification</h3>
                <p className="text-xs text-graphite-400">
                  Assess the alert. Is this activity malicious (True Positive) or is there a legitimate business explanation (False Positive)?
                </p>
                
                <div className="space-y-3 pt-2">
                  <button
                    onClick={() => setIsTruePositive(true)}
                    className={cn(
                      "w-full text-left p-4 rounded-lg border font-mono text-sm transition-all duration-200 flex items-center justify-between",
                      isTruePositive === true
                        ? "bg-red-950/20 border-red-500/30 text-red-200 shadow-[0_0_8px_rgba(239,68,68,0.05)]"
                        : "bg-graphite-950/40 border-graphite-800/80 text-graphite-300 hover:bg-graphite-800/20"
                    )}
                  >
                    <span>True Positive (Threat)</span>
                    <span className={cn(
                      "w-3 h-3 rounded-full border border-graphite-600",
                      isTruePositive === true && "bg-red-500 border-red-400"
                    )} />
                  </button>

                  <button
                    onClick={() => setIsTruePositive(false)}
                    className={cn(
                      "w-full text-left p-4 rounded-lg border font-mono text-sm transition-all duration-200 flex items-center justify-between",
                      isTruePositive === false
                        ? "bg-emerald-950/20 border-emerald-500/30 text-emerald-200 shadow-[0_0_8px_rgba(52,211,153,0.05)]"
                        : "bg-graphite-950/40 border-graphite-800/80 text-graphite-300 hover:bg-graphite-800/20"
                    )}
                  >
                    <span>False Positive (Benign)</span>
                    <span className={cn(
                      "w-3 h-3 rounded-full border border-graphite-600",
                      isTruePositive === false && "bg-emerald-500 border-emerald-400"
                    )} />
                  </button>
                </div>
              </div>

              <div className="glass-panel p-6 rounded-xl flex items-center justify-between bg-emerald-500/5 border-emerald-500/10">
                <div className="space-y-0.5">
                  <h4 className="text-xs font-semibold text-emerald-400 font-mono">Next step:</h4>
                  <p className="text-xs text-graphite-300">Run queries to audit the log files.</p>
                </div>
                <button
                  onClick={() => setActiveTab("detect")}
                  className="p-2 rounded bg-graphite-800 hover:bg-graphite-700 text-white transition-colors"
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
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse-slow" />
                <h3 className="text-base font-bold text-graphite-100">SIEM Log Terminal</h3>
              </div>
              <div className="text-xs font-mono text-graphite-400 bg-graphite-900 px-3 py-1 rounded border border-graphite-800">
                Evidence Discovered: <span className="text-emerald-400 font-bold">{discoveredFindings.length}</span>
                <span className="text-graphite-600"> / {scenario.sanitized_findings?.length ?? 0}</span>
              </div>
            </div>

            <KqlSearchBar
              sessionId={session.session_id}
              logs={scenario.initial_logs}
              onFilter={setFilteredLogs}
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
              />
            </div>

            <div className="space-y-6">
              <div className="glass-panel p-6 rounded-xl space-y-4">
                <h3 className="text-base font-bold text-graphite-100">Final Submission</h3>
                <p className="text-xs text-graphite-400">
                  Review your analysis. Ensure you have classified the triage correctly and selected response options for each phase before submitting.
                </p>

                <div className="space-y-3 pt-2">
                  <div className="flex justify-between text-xs font-mono border-b border-graphite-800 pb-2">
                    <span className="text-graphite-400">Triage Classification:</span>
                    {isTruePositive === null ? (
                      <span className="text-amber-500 font-bold">MISSING</span>
                    ) : (
                      <span className="text-emerald-400 font-bold">
                        {isTruePositive ? "True Positive" : "False Positive"}
                      </span>
                    )}
                  </div>
                  <div className="flex justify-between text-xs font-mono border-b border-graphite-800 pb-2">
                    <span className="text-graphite-400">Response Plan:</span>
                    <span className="text-graphite-200 font-bold">
                      {selectedActions.length} actions selected
                    </span>
                  </div>
                  <div className="flex justify-between text-xs font-mono border-b border-graphite-800 pb-2">
                    <span className="text-graphite-400">Evidence Found:</span>
                    <span className="text-graphite-200 font-bold">
                      {discoveredFindings.length} indicators
                    </span>
                  </div>
                </div>

                <button
                  onClick={handleSubmit}
                  disabled={submitting}
                  className="w-full mt-4 py-3 bg-emerald-500 hover:bg-emerald-400 text-emerald-950 font-bold rounded-lg shadow-lg hover:shadow-emerald-500/20 hover:-translate-y-0.5 active:translate-y-0 disabled:opacity-50 disabled:pointer-events-none transition-all flex items-center justify-center gap-2"
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
