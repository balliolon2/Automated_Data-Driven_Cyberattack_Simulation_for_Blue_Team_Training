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
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-xs font-mono text-graphite-500">Loading skill profiles...</p>
      </div>
    );
  }

  if (errorMessage) {
    return (
      <div className="max-w-md mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
        <ShieldAlert className="w-10 h-10 text-red-500 mx-auto" />
        <h3 className="text-base font-bold text-white tracking-tight">Assessment Required</h3>
        <p className="text-xs text-graphite-400 font-light leading-relaxed">{errorMessage}</p>
        <button
          onClick={() => navigate("/pre-test")}
          className="w-full py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200"
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
        <h1 className="text-2xl font-bold tracking-tight text-white">Skill Evaluation</h1>
        <p className="text-xs text-graphite-400 font-mono">Practical Assessment and Training Plan</p>
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        {/* Left Side: Summary Card */}
        <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg flex flex-col items-center justify-center text-center space-y-3">
          <div className="text-4xl font-extrabold text-white font-mono tracking-tighter">
            {averageScore.toFixed(0)}%
          </div>
          <div className="text-[10px] font-semibold text-graphite-300 uppercase tracking-wider font-mono">
            Average Proficiency
          </div>
          <p className="text-[10px] text-graphite-500 font-mono">
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

      <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-4">
        <h3 className="text-sm font-semibold flex items-center gap-2 text-white border-b border-graphite-800 pb-3">
          <AlertTriangle className="text-amber-500 w-4 h-4" />
          Skill Gap Analysis
        </h3>
        
        <ul className="space-y-4 text-xs text-graphite-400 font-light">
          {failedDomains.length > 0 ? (
            <>
              <li className="flex gap-3 items-start">
                <ShieldAlert className="text-amber-500 shrink-0 w-4 h-4 mt-0.5" />
                <span>
                  <strong className="text-white">Identified Weakness:</strong> You have domains performing below the <strong className="text-white">70%</strong> proficiency threshold. The adaptive simulation system will recommend custom scenarios focusing on your weakest domains first.
                </span>
              </li>
              <li className="flex gap-3 items-start">
                <CheckCircle className="text-emerald-500 shrink-0 w-4 h-4 mt-0.5" />
                <span>
                  <strong className="text-white">Adaptive Roadmap:</strong> Recommended training includes hands-on investigations of alerts, log audits using KQL-like search, and playbook response steps.
                </span>
              </li>
            </>
          ) : (
            <li className="flex gap-3 items-start">
              <CheckCircle className="text-emerald-500 shrink-0 w-4 h-4 mt-0.5" />
              <span>
                <strong className="text-white">Excellent Work!</strong> You have achieved the target proficiency of <strong className="text-white">70%</strong> across all domains. You are fully prepared to proceed to the final Post-Test.
              </span>
            </li>
          )}
        </ul>

        <div className="pt-4 border-t border-graphite-800 flex justify-end">
          {needsTraining ? (
            <button
              onClick={() => navigate("/simulation")}
              className="px-4 py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200 flex items-center gap-2"
            >
              <PlayCircle className="w-4 h-4" />
              Start Recommended Scenarios ({completedScenarios} completed)
            </button>
          ) : (
            <button
              onClick={() => navigate("/simulation/complete")}
              className="px-4 py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200 flex items-center gap-2"
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
