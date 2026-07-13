import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import axios from "axios";
import { CheckCircle2, XCircle, ChevronRight, Award, ShieldAlert, ShieldCheck, Loader2 } from "lucide-react";
import DomainBreakdown from "../components/DomainBreakdown";
import { cn } from "../lib/utils";

interface FindingResult {
  id: string;
  description: string;
  domain_id: string;
  found: boolean;
  points: number;
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
}

interface Scenario {
  title: string;
  domain_id: string;
}

interface ResultData {
  total_score: number;
  tp_fp_correct: boolean;
  tp_fp_points: number;
  findings_points: number;
  response_points: number;
  domain_scores: Record<string, any>;
  findings_detail: FindingResult[];
  response_detail: ResponseActionResult[];
  tp_fp_explanation: string;
}

export default function ScenarioResultPage() {
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(true);
  const [session, setSession] = useState<SimulationSession | null>(null);
  const [scenario, setScenario] = useState<Scenario | null>(null);
  const [result, setResult] = useState<ResultData | null>(null);
  const [domainProficiencies, setDomainProficiencies] = useState<DomainStatus[]>([]);
  const [needsTraining, setNeedsTraining] = useState<boolean>(true);

  const navigate = useNavigate();

  useEffect(() => {
    const fetchResults = async () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          navigate("/login");
          return;
        }

        // 1. Fetch result payload from result endpoint
        const res = await axios.get(`/api/simulation/result/${id}`, {
          headers: { Authorization: `Bearer ${token}` }
        });

        setSession(res.data.session);
        setScenario(res.data.scenario);
        setDomainProficiencies(res.data.domain_proficiencies);

        // Standardize the result model
        // If coming directly from submission, the structure might be nested
        if (res.data.actions) {
          // Re-evaluate matching or use the stored session data
          // We can reconstruct or fetch the result mapping
          // Let's check how result is stored in the database
        }
        
        // Wait, the backend endpoint returns:
        // { session: models.SimulationSession, scenario: models.Scenario, actions: [], domain_proficiencies: [] }
        // Let's reconstruct the score results client-side for review
        const reconResult = reconstructResult(res.data.scenario, res.data.actions);
        setResult(reconResult);

        // 2. Fetch the simulation status to see if training is still needed
        const statusRes = await axios.get("/api/simulation/status", {
          headers: { Authorization: `Bearer ${token}` }
        });
        setNeedsTraining(statusRes.data.needs_training);

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
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-4">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
        <p className="text-sm font-mono text-graphite-400">Analyzing triage report...</p>
      </div>
    );
  }

  if (!session || !scenario || !result) {
    return (
      <div className="text-center py-12 text-graphite-400 font-mono">
        Failed to load incident results.
      </div>
    );
  }

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Page Header */}
      <div className="text-center space-y-2 max-w-2xl mx-auto">
        <div className="inline-flex items-center justify-center p-3 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 mb-2">
          <Award className="w-8 h-8" />
        </div>
        <h1 className="text-3xl font-bold tracking-tight">Scenario Evaluation</h1>
        <p className="text-sm text-graphite-400">
          Incident Audit for: <span className="text-graphite-200 font-semibold">{scenario.title}</span>
        </p>
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        {/* Left Side: Score & Profiles */}
        <div className="space-y-6">
          {/* Main Score Panel */}
          <div className="glass-panel p-6 rounded-xl flex flex-col items-center justify-center text-center space-y-3 relative overflow-hidden">
            <div className="absolute top-0 right-0 w-24 h-24 bg-emerald-500/5 blur-2xl rounded-full" />
            <div className="text-5xl font-black text-emerald-400 font-mono tracking-tight">
              {result.total_score.toFixed(0)}
              <span className="text-sm text-graphite-500 font-medium ml-1">/ 100</span>
            </div>
            <div className="text-sm font-semibold text-graphite-200 uppercase tracking-wider">
              Performance Score
            </div>
            <div className={cn(
              "px-3 py-1 rounded-full text-xs font-mono font-medium border",
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
            className="w-full py-3.5 bg-emerald-500 hover:bg-emerald-400 text-emerald-950 font-bold rounded-lg shadow-lg hover:shadow-emerald-500/20 hover:-translate-y-0.5 active:translate-y-0 transition-all flex items-center justify-center gap-2"
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
          <div className="glass-panel p-6 rounded-xl space-y-4">
            <div className="flex items-center justify-between border-b border-graphite-800 pb-3">
              <h3 className="font-bold text-graphite-100 flex items-center gap-2">
                Triage classification
              </h3>
              {result.tp_fp_correct ? (
                <span className="flex items-center gap-1 text-xs font-bold text-emerald-400">
                  <ShieldCheck className="w-4 h-4" /> Correct
                </span>
              ) : (
                <span className="flex items-center gap-1 text-xs font-bold text-red-400">
                  <ShieldAlert className="w-4 h-4" /> Incorrect
                </span>
              )}
            </div>

            <div className={cn(
              "p-4 rounded-lg border",
              result.tp_fp_correct 
                ? "bg-emerald-950/10 border-emerald-500/10 text-emerald-200/90" 
                : "bg-red-950/10 border-red-500/10 text-red-200/90"
            )}>
              <div className="font-semibold text-xs mb-1 uppercase tracking-wider font-mono">
                Explanation:
              </div>
              <p className="text-xs leading-relaxed">
                {result.tp_fp_explanation}
              </p>
            </div>
          </div>

          {/* Key Findings */}
          <div className="glass-panel p-6 rounded-xl space-y-4">
            <h3 className="font-bold text-graphite-100 border-b border-graphite-800 pb-3">
              Detect & Analysis Findings
            </h3>
            
            <div className="space-y-3">
              {result.findings_detail.map((finding) => (
                <div 
                  key={finding.id}
                  className={cn(
                    "p-4 rounded-lg border flex gap-3 items-start transition-all",
                    finding.found 
                      ? "bg-emerald-950/5 border-emerald-500/10" 
                      : "bg-graphite-950/40 border-graphite-800/80"
                  )}
                >
                  {finding.found ? (
                    <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0 mt-0.5" />
                  ) : (
                    <XCircle className="w-5 h-5 text-graphite-600 shrink-0 mt-0.5" />
                  )}
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold text-graphite-200">{finding.description}</span>
                      <span className="px-1.5 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-[9px] font-mono text-graphite-400 uppercase">
                        {finding.domain_id}
                      </span>
                    </div>
                    <p className="text-xs text-graphite-400 leading-relaxed">
                      {finding.explanation}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Response Plan */}
          <div className="glass-panel p-6 rounded-xl space-y-4">
            <h3 className="font-bold text-graphite-100 border-b border-graphite-800 pb-3">
              Response Plan Execution
            </h3>
            
            <div className="space-y-3">
              {result.response_detail.map((action) => {
                return (
                  <div 
                    key={action.id}
                    className={cn(
                      "p-4 rounded-lg border flex gap-3 items-start transition-all",
                      action.selected 
                        ? action.is_correct
                          ? "bg-emerald-950/5 border-emerald-500/10"
                          : "bg-red-950/5 border-red-500/10"
                        : "bg-graphite-950/40 border-graphite-800/80"
                    )}
                  >
                    {action.selected ? (
                      action.is_correct ? (
                        <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0 mt-0.5" />
                      ) : (
                        <XCircle className="w-5 h-5 text-red-400 shrink-0 mt-0.5" />
                      )
                    ) : (
                      <div className="w-5 h-5 rounded-full border border-graphite-800 shrink-0 mt-0.5 flex items-center justify-center text-[10px] font-bold text-graphite-500 font-mono">
                        -
                      </div>
                    )}
                    
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-graphite-200">{action.label}</span>
                        <span className="px-1.5 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-[9px] font-mono text-graphite-400 uppercase">
                          {action.phase} · {action.domain_id}
                        </span>
                        {action.selected && (
                          <span className={cn(
                            "text-[10px] font-mono font-bold",
                            action.is_correct ? "text-emerald-400" : "text-red-400"
                          )}>
                            ({action.points > 0 ? `+${action.points}` : action.points} pts)
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-graphite-400 leading-relaxed">
                        {action.explanation}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// Reconstructs the detailed score result client-side from the scenario data and actions logged
function reconstructResult(scenario: any, actions: any[]): ResultData {
  const result: ResultData = {
    total_score: 0,
    tp_fp_correct: false,
    tp_fp_points: 0,
    findings_points: 0,
    response_points: 0,
    domain_scores: {},
    findings_detail: [],
    response_detail: [],
    tp_fp_explanation: scenario.tp_fp_explanation
  };

  const domainEarned: Record<string, number> = {};
  const domainMaxPoints: Record<string, number> = {};

  // Find user actions in history
  const triageAction = actions.find(a => a.action_type === "triage_alert");
  const respondAction = actions.find(a => a.action_type === "respond");
  const findingsAction = actions.find(a => a.action_type === "submit_decision");

  // 1. Score TP/FP
  const userChoice = triageAction?.payload?.user_choice;
  result.tp_fp_correct = userChoice === scenario.is_true_positive;
  const tpfpMaxPoints = 20.0;
  if (result.tp_fp_correct) {
    result.tp_fp_points = tpfpMaxPoints;
    domainEarned[scenario.domain_id] = (domainEarned[scenario.domain_id] || 0) + tpfpMaxPoints;
  }
  domainMaxPoints[scenario.domain_id] = (domainMaxPoints[scenario.domain_id] || 0) + tpfpMaxPoints;

  // 2. Score Findings
  const discoveredSet = new Set<string>(findingsAction?.payload?.discovered_findings || []);
  const keyFindings = scenario.expected_outcomes?.key_findings || [];
  
  keyFindings.forEach((kf: any) => {
    const found = discoveredSet.has(kf.id);
    const pts = parseFloat(kf.points || "0");
    
    result.findings_detail.push({
      id: kf.id,
      description: kf.description,
      domain_id: kf.domain_id,
      found,
      points: pts,
      explanation: kf.explanation
    });

    if (found) {
      result.findings_points += pts;
      domainEarned[kf.domain_id] = (domainEarned[kf.domain_id] || 0) + pts;
    }
    domainMaxPoints[kf.domain_id] = (domainMaxPoints[kf.domain_id] || 0) + pts;
  });

  // 3. Score Playbooks
  const selectedSet = new Set<string>(respondAction?.payload?.selected_actions || []);
  const phases = ["containment", "eradication", "recovery"];
  
  phases.forEach(phase => {
    const phaseActions = scenario.playbook_steps?.[phase] || [];
    phaseActions.forEach((act: any) => {
      const selected = selectedSet.has(act.id);
      const pts = parseFloat(act.points || "0");
      
      let earned = 0;
      if (selected) {
        earned = pts;
        result.response_points += pts;
        domainEarned[act.domain_id] = (domainEarned[act.domain_id] || 0) + pts;
      }

      if (pts > 0) {
        domainMaxPoints[act.domain_id] = (domainMaxPoints[act.domain_id] || 0) + pts;
      }

      result.response_detail.push({
        id: act.id,
        label: act.label,
        phase,
        domain_id: act.domain_id,
        selected,
        is_correct: act.is_correct,
        points: pts,
        earned,
        explanation: act.explanation
      });
    });
  });

  // Calculate total score
  let totalMax = 0;
  let totalEarned = 0;
  
  Object.keys(domainMaxPoints).forEach(d => {
    totalMax += domainMaxPoints[d];
    totalEarned += Math.max(0, domainEarned[d] || 0);
  });

  if (totalMax > 0) {
    result.total_score = Math.round((totalEarned / totalMax) * 100 * 100) / 100;
  }

  return result;
}
