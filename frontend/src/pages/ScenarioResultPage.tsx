import { useState, useEffect } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import axios from "axios";
import { CheckCircle2, XCircle, ChevronRight, Award, ShieldAlert, ShieldCheck, Loader2, MessageSquare, Plus, Search } from "lucide-react";
import DomainBreakdown from "../components/DomainBreakdown";
import { cn } from "../lib/utils";

interface TelemetryLog {
  timestamp?: string;
  source?: string;
  severity?: string;
  message?: string;
  [key: string]: unknown;
}

interface FindingResult {
  id: string;
  description: string;
  domain_id: string;
  found: boolean;
  points: number;
  earned?: number;
  evidence_log_indices?: number[] | null;
  matched_log_indices?: number[] | null;
  explanation: string;
}

interface ResponseActionResult {
  id: string;
  label: string;
  phase: string;
  domain_id: string;
  selected: boolean;
  is_correct: boolean;
  points: number;
  earned: number;
  explanation: string;
}

interface DomainStatus {
  domain_id: string;
  domain_name: string;
  proficiency_score: number;
  threshold: number;
  passed: boolean;
}

interface SimulationSession {
  session_id: string;
  scenario_id: string;
  status: string;
  final_score: number;
  skill_gap: any;
  rubric_version?: string;
}

interface Scenario {
  title: string;
  domain_id: string;
  initial_logs?: TelemetryLog[] | null;
}

interface ResultData {
  total_score: number;
  tp_fp_correct: boolean;
  tp_fp_points: number;
  findings_points: number;
  response_points: number;
  gross_findings_points?: number;
  evidence_penalty?: number;
  noise_log_indices?: number[] | null;
  selected_log_indices?: number[] | null;
  domain_scores: Record<string, any>;
  findings_detail: FindingResult[];
  response_detail: ResponseActionResult[];
  tp_fp_explanation: string;
}

// New sessions serialise evidence fields as arrays (empty when nothing was
// selected). Legacy stored results lack them and arrive as null, so presence
// checks must be structural rather than truthy.
function hasEvidenceReview(
  result: ResultData | null,
  session: SimulationSession | null,
): boolean {
  if (!result) return false;
  if (
    !Array.isArray(result.selected_log_indices) ||
    !Array.isArray(result.noise_log_indices) ||
    typeof result.evidence_penalty !== "number" ||
    typeof result.findings_points !== "number"
  ) {
    return false;
  }
  const findings = result.findings_detail ?? [];
  if (
    !findings.every(
      (finding) =>
        typeof finding.earned === "number" &&
        Array.isArray(finding.evidence_log_indices) &&
        Array.isArray(finding.matched_log_indices),
    )
  ) {
    return false;
  }
  // Only a recognised new-rubric session may show this panel. A missing or
  // unknown rubric_version means a legacy result, even if evidence arrays were
  // reconstructed, so those keep the old findings presentation.
  const rubricVersion = session?.rubric_version;
  if (
    typeof rubricVersion !== "string" ||
    rubricVersion.length === 0 ||
    rubricVersion.startsWith("rubric-v1")
  ) {
    return false;
  }
  return true;
}

function logField(log: TelemetryLog | undefined, ...keys: string[]): string {
  if (!log) return "—";
  for (const key of keys) {
    const value = log[key];
    if (value !== undefined && value !== null && String(value).length > 0) {
      return String(value);
    }
  }
  return "—";
}

function EvidenceLogRow({
  index,
  log,
  tone = "default",
}: {
  index: number;
  log?: TelemetryLog;
  tone?: "default" | "matched" | "missed" | "noise";
}) {
  const accent =
    tone === "matched"
      ? "border-emerald-500/20 bg-emerald-950/5"
      : tone === "missed"
        ? "border-graphite-800/80 bg-graphite-950/40"
        : tone === "noise"
          ? "border-red-500/15 bg-red-950/5"
          : "border-graphite-800/80 bg-graphite-950/40";

  return (
    <div className={cn("rounded border px-3 py-2 font-mono text-[10px] leading-relaxed", accent)}>
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-graphite-400">
        <span className="text-cyan-300/80">#{index}</span>
        <span>{logField(log, "timestamp", "Timestamp")}</span>
        <span className="text-sky-300/80">{logField(log, "source", "Source")}</span>
        <span className="uppercase font-bold text-amber-400/90">
          {logField(log, "severity", "Severity")}
        </span>
      </div>
      <p className="text-graphite-200 break-words">
        {logField(log, "message", "Message")}
      </p>
    </div>
  );
}

export default function ScenarioResultPage() {
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(true);
  const [session, setSession] = useState<SimulationSession | null>(null);
  const [scenario, setScenario] = useState<Scenario | null>(null);
  const [result, setResult] = useState<ResultData | null>(null);
  const [domainProficiencies, setDomainProficiencies] = useState<DomainStatus[]>([]);
  const [needsTraining, setNeedsTraining] = useState<boolean>(true);
  const [linkedThreads, setLinkedThreads] = useState<any[]>([]);
  const [threadsLoading, setThreadsLoading] = useState(false);

  const navigate = useNavigate();

  useEffect(() => {
    const fetchResults = async () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          navigate("/login");
          return;
        }

        // 1. Fetch result payload directly from backend evaluation endpoint
        const res = await axios.get(`/api/simulation/result/${id}`, {
          headers: { Authorization: `Bearer ${token}` }
        });

        setSession(res.data.session);
        setScenario(res.data.scenario);
        setDomainProficiencies(res.data.domain_proficiencies);
        setResult(res.data.result);

        // 2. Fetch the simulation status to see if training is still needed
        const statusRes = await axios.get("/api/simulation/status", {
          headers: { Authorization: `Bearer ${token}` }
        });
        setNeedsTraining(statusRes.data.needs_training);

        // 3. Fetch linked analysis threads if scenario exists
        if (res.data.session?.scenario_id) {
          setThreadsLoading(true);
          try {
            const threadRes = await axios.get(`/api/threads?scenario_id=${res.data.session.scenario_id}`, {
              headers: { Authorization: `Bearer ${token}` },
            });
            setLinkedThreads(threadRes.data.threads || []);
          } catch (tErr) {
            console.error("Failed to load linked threads:", tErr);
          } finally {
            setThreadsLoading(false);
          }
        }

      } catch (err) {
        console.error("Failed to load scenario results:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchResults();
  }, [id, navigate]);

  const handleNext = () => {
    if (id) {
      localStorage.setItem(`ack_session_${id}`, "true");
    }
    if (needsTraining) {
      navigate("/simulation");
    } else {
      navigate("/simulation/complete");
    }
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-xs font-mono text-graphite-500">Analyzing triage report...</p>
      </div>
    );
  }

  if (!session || !scenario || !result) {
    return (
      <div className="text-center py-12 text-graphite-500 font-mono text-xs">
        Failed to load incident results.
      </div>
    );
  }

  const showEvidenceReview = hasEvidenceReview(result, session);
  const initialLogs: TelemetryLog[] = Array.isArray(scenario.initial_logs)
    ? scenario.initial_logs
    : [];
  const logAt = (index: number): TelemetryLog | undefined => initialLogs[index];
  const selectedLogIndices = showEvidenceReview ? result.selected_log_indices ?? [] : [];
  const noiseLogIndices = showEvidenceReview ? result.noise_log_indices ?? [] : [];
  const findingsDetail = result.findings_detail ?? [];
  // Gross credit is server-owned. Older responses without the field fall back
  // to the server's net findings points plus its evidence penalty; the client
  // never re-derives scoring from per-finding earned values.
  const grossFindingCredit =
    typeof result.gross_findings_points === "number"
      ? result.gross_findings_points
      : typeof result.findings_points === "number" &&
          typeof result.evidence_penalty === "number"
        ? result.findings_points + result.evidence_penalty
        : null;

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Page Header */}
      <div className="text-center space-y-2 max-w-2xl mx-auto">
        <div className="inline-flex items-center justify-center p-3 rounded-md bg-graphite-900 border border-graphite-800 text-white mb-2">
          <Award className="w-6 h-6" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-white">Scenario Evaluation</h1>
        <p className="text-xs text-graphite-400 font-mono">
          Incident Audit for: <span className="text-white font-semibold">{scenario.title}</span>
        </p>
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        {/* Left Side: Score & Profiles */}
        <div className="space-y-6">
          {/* Main Score Panel */}
          <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg flex flex-col items-center justify-center text-center space-y-3">
            <div className="text-4xl font-extrabold text-white font-mono tracking-tighter">
              {result.total_score.toFixed(0)}
              <span className="text-xs text-graphite-500 font-medium ml-1">/ 100</span>
            </div>
            <div className="text-[10px] font-semibold text-graphite-300 uppercase tracking-wider font-mono">
              Performance Score
            </div>
            <div className={cn(
              "px-2.5 py-0.5 rounded text-[10px] font-mono font-medium border",
              result.total_score >= 70 
                ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" 
                : "bg-amber-500/10 text-amber-400 border-amber-500/20"
            )}>
              {result.total_score >= 70 ? "Tactical Pass" : "Retraining Advised"}
            </div>
          </div>

          {/* Domain Breakdown */}
          <DomainBreakdown domains={domainProficiencies} title="Updated Domain Proficiency" />

          {/* Action Button */}
          <button
            onClick={handleNext}
            className="w-full py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200 flex items-center justify-center gap-2"
          >
            {needsTraining ? (
              <>
                Next Training Scenario
                <ChevronRight className="w-4 h-4" />
              </>
            ) : (
              <>
                Complete Simulation
                <CheckCircle2 className="w-4 h-4" />
              </>
            )}
          </button>
        </div>

        {/* Right Side: Analysis reviews */}
        <div className="md:col-span-2 space-y-6">
          {/* Triage Decision */}
          <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
            <div className="flex items-center justify-between border-b border-graphite-800 pb-3">
              <h3 className="font-semibold text-sm text-white flex items-center gap-2">
                Triage classification
              </h3>
              {result.tp_fp_correct ? (
                <span className="flex items-center gap-1 text-xs font-semibold text-emerald-400">
                  <ShieldCheck className="w-4 h-4" /> Correct
                </span>
              ) : (
                <span className="flex items-center gap-1 text-xs font-semibold text-red-400">
                  <ShieldAlert className="w-4 h-4" /> Incorrect
                </span>
              )}
            </div>

            <div className={cn(
              "p-4 rounded border",
              result.tp_fp_correct 
                ? "bg-emerald-950/10 border-emerald-500/10 text-emerald-300" 
                : "bg-red-950/10 border-red-500/10 text-red-300"
            )}>
              <div className="font-semibold text-[10px] mb-1.5 uppercase tracking-wider font-mono text-white">
                Explanation:
              </div>
              <p className="text-xs leading-relaxed font-light">
                {result.tp_fp_explanation}
              </p>
            </div>
          </div>

          {/* Key Findings */}
          <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
            <h3 className="font-semibold text-sm text-white border-b border-graphite-800 pb-3">
              Detect & Analysis Findings
            </h3>
            
            <div className="space-y-3">
              {findingsDetail.map((finding) => {
                const matched =
                  Array.isArray(finding.matched_log_indices) ? finding.matched_log_indices : [];
                const acceptable =
                  Array.isArray(finding.evidence_log_indices) ? finding.evidence_log_indices : [];
                const earned = finding.earned ?? (finding.found ? finding.points : 0);

                return (
                  <div
                    key={finding.id}
                    className={cn(
                      "p-4 rounded border flex gap-3 items-start transition-all text-xs font-light",
                      finding.found
                        ? "bg-emerald-950/5 border-emerald-500/10"
                        : "bg-graphite-950/40 border-graphite-800/80"
                    )}
                  >
                    {finding.found ? (
                      <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                    ) : (
                      <XCircle className="w-4 h-4 text-graphite-600 shrink-0 mt-0.5" />
                    )}
                    {showEvidenceReview ? (
                      <div className="min-w-0 flex-1 space-y-2">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-semibold text-white">{finding.description}</span>
                          <span className="px-1.5 py-0.5 rounded bg-graphite-950 border border-graphite-800 text-[8px] font-mono text-graphite-400 uppercase">
                            {finding.domain_id}
                          </span>
                          <span
                            className={cn(
                              "text-[8px] font-mono font-bold uppercase tracking-wider",
                              finding.found ? "text-emerald-400" : "text-graphite-500"
                            )}
                          >
                            {finding.found ? "Found" : "Missed"}
                          </span>
                          <span className="text-[9px] font-mono text-graphite-400">
                            {earned} / {finding.points} pts
                          </span>
                        </div>
                        <p className="text-graphite-400 leading-relaxed">
                          {finding.explanation}
                        </p>

                        {finding.found && matched.length > 0 && (
                          <div className="space-y-1.5 pt-1">
                            <div className="text-[9px] font-mono uppercase tracking-wider text-emerald-400/80">
                              Matched evidence
                            </div>
                            {matched.map((index) => (
                              <EvidenceLogRow
                                key={`matched-${finding.id}-${index}`}
                                index={index}
                                log={logAt(index)}
                                tone="matched"
                              />
                            ))}
                          </div>
                        )}

                        {!finding.found && acceptable.length > 0 && (
                          <div className="space-y-1.5 pt-1">
                            <div className="text-[9px] font-mono uppercase tracking-wider text-graphite-500">
                              Acceptable evidence that would have proven this finding
                            </div>
                            {acceptable.map((index) => (
                              <EvidenceLogRow
                                key={`acceptable-${finding.id}-${index}`}
                                index={index}
                                log={logAt(index)}
                                tone="missed"
                              />
                            ))}
                          </div>
                        )}
                      </div>
                    ) : (
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-white">{finding.description}</span>
                          <span className="px-1.5 py-0.5 rounded bg-graphite-950 border border-graphite-800 text-[8px] font-mono text-graphite-400 uppercase">
                            {finding.domain_id}
                          </span>
                        </div>
                        <p className="text-graphite-400 leading-relaxed">
                          {finding.explanation}
                        </p>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          {/* Evidence Review */}
          {showEvidenceReview && (
            <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
              <div className="flex flex-wrap items-center justify-between gap-2 border-b border-graphite-800 pb-3">
                <h3 className="font-semibold text-sm text-white flex items-center gap-2">
                  <Search className="w-4 h-4 text-cyan-300/80" />
                  Evidence Review
                </h3>
                <span className="text-[10px] font-mono text-graphite-400">
                  Selected evidence: {selectedLogIndices.length} log
                  {selectedLogIndices.length === 1 ? "" : "s"}
                </span>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div className="rounded border border-graphite-800/80 bg-graphite-950/40 p-3">
                  <div className="text-[9px] font-mono uppercase tracking-wider text-graphite-500">
                    Gross finding credit
                  </div>
                  <div className="mt-1 text-lg font-mono font-bold text-white">
                    {grossFindingCredit ?? "—"} pts
                  </div>
                </div>
                <div className="rounded border border-graphite-800/80 bg-graphite-950/40 p-3">
                  <div className="text-[9px] font-mono uppercase tracking-wider text-graphite-500">
                    Evidence penalty
                  </div>
                  <div
                    className={cn(
                      "mt-1 text-lg font-mono font-bold",
                      result.evidence_penalty && result.evidence_penalty > 0
                        ? "text-red-400"
                        : "text-white"
                    )}
                  >
                    {result.evidence_penalty && result.evidence_penalty > 0
                      ? `-${result.evidence_penalty}`
                      : "0"}{" "}
                    pts
                  </div>
                </div>
                <div className="rounded border border-graphite-800/80 bg-graphite-950/40 p-3">
                  <div className="text-[9px] font-mono uppercase tracking-wider text-graphite-500">
                    Net findings points
                  </div>
                  <div className="mt-1 text-lg font-mono font-bold text-emerald-400">
                    {result.findings_points} pts
                  </div>
                </div>
              </div>

              {selectedLogIndices.length === 0 ? (
                <div className="rounded border border-dashed border-graphite-800/80 bg-graphite-950/30 px-4 py-6 text-center space-y-1">
                  <p className="text-xs text-graphite-300">
                    No evidence was selected for this attempt.
                  </p>
                  <p className="text-[10px] font-mono text-graphite-500">
                    Detect &amp; Analysis scored 0 findings points; no false-lead penalty was applied.
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  {selectedLogIndices.map((index) => {
                    const proven = findingsDetail.filter(
                      (finding) =>
                        Array.isArray(finding.matched_log_indices) &&
                        finding.matched_log_indices.includes(index)
                    );
                    const isNoise = noiseLogIndices.includes(index);

                    return (
                      <div key={`selected-${index}`} className="space-y-1.5">
                        <EvidenceLogRow
                          index={index}
                          log={logAt(index)}
                          tone={proven.length > 0 ? "matched" : isNoise ? "noise" : "default"}
                        />
                        {proven.length > 0 ? (
                          <div className="flex flex-wrap items-center gap-1.5 pl-1">
                            <span className="text-[9px] font-mono uppercase tracking-wider text-emerald-400/80">
                              Proved:
                            </span>
                            {proven.map((finding) => (
                              <span
                                key={`proven-${index}-${finding.id}`}
                                className="px-1.5 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20 text-[9px] font-mono text-emerald-300"
                              >
                                {finding.description}
                              </span>
                            ))}
                          </div>
                        ) : (
                          <div className="pl-1 text-[9px] font-mono text-red-400/90">
                            Not evidence for any Key Finding — 2-point penalty
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {/* Response Plan */}
          <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
            <h3 className="font-semibold text-sm text-white border-b border-graphite-800 pb-3">
              Response Plan Execution
            </h3>
            
            <div className="space-y-3">
              {result.response_detail.map((action) => {
                return (
                  <div 
                    key={action.id}
                    className={cn(
                      "p-4 rounded border flex gap-3 items-start transition-all text-xs font-light",
                      action.selected 
                        ? action.is_correct
                          ? "bg-emerald-950/5 border-emerald-500/10"
                          : "bg-red-950/5 border-red-500/10"
                        : "bg-graphite-950/40 border-graphite-800/80"
                    )}
                  >
                    {action.selected ? (
                      action.is_correct ? (
                        <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                      ) : (
                        <XCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
                      )
                    ) : (
                      <div className="w-4 h-4 rounded-full border border-graphite-800 shrink-0 mt-0.5 flex items-center justify-center text-[10px] font-bold text-graphite-500 font-mono">
                        -
                      </div>
                    )}
                    
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-white">{action.label}</span>
                        <span className="px-1.5 py-0.5 rounded bg-graphite-950 border border-graphite-800 text-[8px] font-mono text-graphite-400 uppercase">
                          {action.phase} · {action.domain_id}
                        </span>
                        {action.selected && (
                          <span className={cn(
                            "text-[9px] font-mono font-bold",
                            action.is_correct ? "text-emerald-400" : "text-red-400"
                          )}>
                            ({action.points > 0 ? `+${action.points}` : action.points} pts)
                          </span>
                        )}
                      </div>
                      <p className="text-graphite-400 leading-relaxed">
                        {action.explanation}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Specialist Insights & Community Discussions */}
          <div className="p-5 rounded-lg bg-graphite-900/40 border border-graphite-800 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 border-b border-graphite-800 pb-3">
              <div className="flex items-center gap-2">
                <MessageSquare className="w-4 h-4 text-indigo-400" />
                <h3 className="font-bold text-sm text-white">Specialist Insights & Analysis Debriefs</h3>
              </div>

              {(localStorage.getItem("userRole") === "specialist" || localStorage.getItem("userRole") === "admin") && (
                <button
                  onClick={() =>
                    navigate(
                      `/discussions/new?scenario_id=${session?.scenario_id}&scenario_title=${encodeURIComponent(
                        scenario?.title || ""
                      )}`
                    )
                  }
                  className="flex items-center gap-1.5 px-3 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold transition-colors cursor-pointer self-start sm:self-auto"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Write Analysis on this Scenario</span>
                </button>
              )}
            </div>

            {threadsLoading ? (
              <div className="py-6 text-center text-graphite-500 font-mono text-xs">
                Loading linked analysis threads...
              </div>
            ) : linkedThreads.length === 0 ? (
              <div className="py-8 text-center space-y-2 border border-dashed border-graphite-800/80 rounded-lg">
                <p className="text-xs text-graphite-400">
                  No specialist analysis published yet for this specific scenario.
                </p>
                {localStorage.getItem("userRole") === "specialist" || localStorage.getItem("userRole") === "admin" ? (
                  <button
                    onClick={() =>
                      navigate(
                        `/discussions/new?scenario_id=${session?.scenario_id}&scenario_title=${encodeURIComponent(
                          scenario?.title || ""
                        )}`
                      )
                    }
                    className="text-xs text-indigo-400 hover:text-indigo-300 font-semibold underline underline-offset-2 cursor-pointer"
                  >
                    Be the first specialist to deconstruct this attack attempt!
                  </button>
                ) : (
                  <Link
                    to="/discussions"
                    className="text-xs text-graphite-400 hover:text-white underline underline-offset-2 inline-block"
                  >
                    Browse general discussions & debriefs &rarr;
                  </Link>
                )}
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {linkedThreads.map((t) => (
                  <div
                    key={t.thread_id}
                    onClick={() => navigate(`/discussions/${t.thread_id}`)}
                    className="p-3.5 rounded-lg bg-graphite-950 border border-graphite-800 hover:border-graphite-700 transition-all cursor-pointer space-y-2"
                  >
                    <div className="flex items-center justify-between text-xs">
                      <span className="font-mono text-[11px] text-graphite-300">
                        by {t.author.nickname}
                      </span>
                      <span className="px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                        {t.author.role.toUpperCase()}
                      </span>
                    </div>

                    <h4 className="font-semibold text-white text-xs hover:text-indigo-300 transition-colors line-clamp-1">
                      {t.title}
                    </h4>

                    <p className="text-[11px] text-graphite-400 line-clamp-2 leading-relaxed">
                      {t.content.replace(/[#*`>|]/g, "").slice(0, 100)}...
                    </p>

                    <div className="flex items-center justify-between text-[10px] font-mono text-graphite-500 pt-1.5 border-t border-graphite-800/60">
                      <span>{t.upvote_count} helpful</span>
                      <span className="text-indigo-400 flex items-center gap-0.5">
                        Read full &rarr;
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
