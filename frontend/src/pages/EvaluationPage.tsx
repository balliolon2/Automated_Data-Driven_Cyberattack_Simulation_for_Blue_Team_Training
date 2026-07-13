import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { AlertTriangle, CheckCircle, ShieldAlert, Award, PlayCircle, Loader2 } from "lucide-react";
import DomainBreakdown from "../components/DomainBreakdown";

interface DomainStatus {
  domain_id: string;
  domain_name: string;
  proficiency_score: number;
  threshold: number;
  passed: boolean;
}

export default function EvaluationPage() {
  const [loading, setLoading] = useState(true);
  const [domains, setDomains] = useState<DomainStatus[]>([]);
  const [completedScenarios, setCompletedScenarios] = useState(0);
  const [needsTraining, setNeedsTraining] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

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
          setErrorMessage(res.data.error);
          setLoading(false);
          return;
        }

        setDomains(res.data.domains || []);
        setCompletedScenarios(res.data.completed_scenarios || 0);
        setNeedsTraining(res.data.needs_training);
      } catch (err) {
        console.error("Failed to load evaluation details:", err);
        setErrorMessage("Please complete the pre-test before viewing evaluation metrics.");
      } finally {
        setLoading(false);
      }
    };

    fetchStatus();
  }, [navigate]);

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-4">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
        <p className="text-sm font-mono text-graphite-400">Loading skill profiles...</p>
      </div>
    );
  }

  if (errorMessage) {
    return (
      <div className="max-w-md mx-auto mt-12 glass-panel p-6 rounded-xl border-red-500/20 text-center space-y-4 animate-fade-in">
        <ShieldAlert className="w-12 h-12 text-red-400 mx-auto" />
        <h3 className="text-lg font-bold text-graphite-100">Assessment Required</h3>
        <p className="text-sm text-graphite-300">{errorMessage}</p>
        <button
          onClick={() => navigate("/pre-test")}
          className="px-5 py-2.5 bg-emerald-500 text-emerald-950 font-bold rounded-lg hover:bg-emerald-400 transition-all text-sm"
        >
          Take Pre-Test
        </button>
      </div>
    );
  }

  const failedDomains = domains.filter((d) => !d.passed);
  const averageScore = domains.reduce((acc, d) => acc + d.proficiency_score, 0) / Math.max(1, domains.length);

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-fade-in">
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">Skill Evaluation</h1>
        <p className="text-sm text-graphite-400">Practical Assessment and Training Plan</p>
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        {/* Left Side: Summary Card */}
        <div className="glass-panel p-6 rounded-xl flex flex-col items-center justify-center text-center space-y-3 relative overflow-hidden">
          <div className="absolute top-0 right-0 w-20 h-20 bg-emerald-500/5 blur-2xl rounded-full" />
          <div className="text-5xl font-black text-emerald-400 font-mono tracking-tight">
            {averageScore.toFixed(0)}%
          </div>
          <div className="text-sm font-semibold text-graphite-200 uppercase tracking-wider">
            Average Proficiency
          </div>
          <p className="text-xs text-graphite-400 font-mono">
            {failedDomains.length > 0
              ? `${failedDomains.length} domains below threshold`
              : "All domains passed"}
          </p>
        </div>

        {/* Right Side: Detailed Breakdown List */}
        <div className="md:col-span-2">
          <DomainBreakdown domains={domains} title="Current Domain Proficiency Levels" />
        </div>
      </div>

      <div className="glass-panel p-6 rounded-xl space-y-4">
        <h3 className="text-lg font-bold flex items-center gap-2 text-graphite-100 border-b border-graphite-800 pb-3">
          <AlertTriangle className="text-amber-500 w-5 h-5" />
          Skill Gap Analysis
        </h3>
        
        <ul className="space-y-4 text-xs text-graphite-300">
          {failedDomains.length > 0 ? (
            <>
              <li className="flex gap-3 items-start">
                <ShieldAlert className="text-amber-500 shrink-0 w-5 h-5" />
                <span>
                  <strong>Identified Weakness:</strong> You have domains performing below the <strong>70%</strong> proficiency threshold. The adaptive simulation system will recommend custom scenarios focusing on your weakest domains first.
                </span>
              </li>
              <li className="flex gap-3 items-start">
                <CheckCircle className="text-emerald-500 shrink-0 w-5 h-5" />
                <span>
                  <strong>Adaptive Roadmap:</strong> Recommended training includes hands-on investigations of alerts, log audits using KQL-like search, and playbook response steps.
                </span>
              </li>
            </>
          ) : (
            <li className="flex gap-3 items-start">
              <CheckCircle className="text-emerald-500 shrink-0 w-5 h-5" />
              <span>
                <strong>Excellent Work!</strong> You have achieved the target proficiency of <strong>70%</strong> across all domains. You are fully prepared to proceed to the final Post-Test.
              </span>
            </li>
          )}
        </ul>

        <div className="pt-4 border-t border-graphite-800 flex justify-end">
          {needsTraining ? (
            <button
              onClick={() => navigate("/simulation")}
              className="px-6 py-2.5 bg-emerald-500 hover:bg-emerald-400 text-emerald-950 font-bold rounded-lg shadow-lg hover:shadow-emerald-500/20 hover:-translate-y-0.5 active:translate-y-0 transition-all flex items-center gap-2 text-sm"
            >
              <PlayCircle className="w-4 h-4" />
              Start Recommended Scenarios ({completedScenarios} completed)
            </button>
          ) : (
            <button
              onClick={() => navigate("/simulation/complete")}
              className="px-6 py-2.5 bg-emerald-500 hover:bg-emerald-400 text-emerald-950 font-bold rounded-lg shadow-lg hover:shadow-emerald-500/20 hover:-translate-y-0.5 active:translate-y-0 transition-all flex items-center gap-2 text-sm"
            >
              <Award className="w-4 h-4" />
              View Completion Certificate
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
