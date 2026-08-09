import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { Award, ChevronRight, Loader2 } from "lucide-react";
import DomainBreakdown from "../components/DomainBreakdown";

interface DomainStatus {
  domain_id: string;
  domain_name: string;
  proficiency_score: number;
  threshold: number;
  passed: boolean;
}

export default function SimulationCompletePage() {
  const [loading, setLoading] = useState(true);
  const [domainProficiencies, setDomainProficiencies] = useState<DomainStatus[]>([]);
  const [completedScenarios, setCompletedScenarios] = useState(0);
  const [allPassed, setAllPassed] = useState(false);

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

        setDomainProficiencies(res.data.domains || []);
        setCompletedScenarios(res.data.completed_scenarios || 0);
        setAllPassed(res.data.all_domains_passed || false);
      } catch (err) {
        console.error("Failed to load simulation status:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchStatus();
  }, [navigate]);

  const handleStartPostTest = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem("token");
      await axios.post("/api/exams/post-test", {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      navigate("/post-test");
    } catch (err) {
      console.error("Failed to start post-test:", err);
      navigate("/post-test");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-xs font-mono text-graphite-500">Finalizing training reports...</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto space-y-8 animate-fade-in">
      <div className="text-center space-y-3 max-w-xl mx-auto">
        <div className="inline-flex items-center justify-center p-3 rounded-md bg-graphite-900 border border-graphite-800 text-white mb-2 relative">
          <Award className="w-8 h-8 text-emerald-400" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-white">
          {allPassed ? "Training Complete!" : "Simulation Training Overview"}
        </h1>
        <p className="text-xs text-graphite-400 leading-relaxed font-light">
          {allPassed
            ? "Congratulations! You have completed the practical SOC Incident Simulation training loop. All security domains have reached the required 70% proficiency threshold."
            : "You have reviewed your active training metrics. Work through additional targeted scenarios to raise all security domains above the 70% threshold."}
        </p>
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        {/* Stats card */}
        <div className="bg-graphite-900 border border-graphite-800/80 p-6 rounded-lg space-y-6 flex flex-col justify-between">
          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-white border-b border-graphite-800 pb-3">
              Training Metrics
            </h3>
            
            <div className="space-y-3 font-mono text-[10px]">
              <div className="flex justify-between py-1 border-b border-graphite-800/60">
                <span className="text-graphite-400">Scenarios Completed:</span>
                <span className="text-white font-semibold">{completedScenarios}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-graphite-800/60">
                <span className="text-graphite-400">Average Proficiency:</span>
                <span className="text-emerald-400 font-semibold">
                  {(domainProficiencies.reduce((acc, d) => acc + d.proficiency_score, 0) / Math.max(1, domainProficiencies.length)).toFixed(1)}%
                </span>
              </div>
              <div className="flex justify-between py-1 border-b border-graphite-800/60">
                <span className="text-graphite-400">Status:</span>
                <span className={allPassed ? "text-emerald-400 font-semibold" : "text-amber-400 font-semibold"}>
                  {allPassed ? "READY FOR EXAM" : "TRAINING IN PROGRESS"}
                </span>
              </div>
            </div>
          </div>

          <div className="space-y-2">
            {!allPassed && (
              <button
                onClick={() => navigate("/simulation")}
                className="w-full py-2 bg-emerald-500 hover:bg-emerald-600 text-black font-semibold text-xs rounded-md transition-all duration-200"
              >
                Continue Incident Simulation
              </button>
            )}
            <button
              onClick={handleStartPostTest}
              className="w-full py-2 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200 flex items-center justify-center gap-2"
            >
              Start Final Post-Test
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Breakdown display */}
        <DomainBreakdown domains={domainProficiencies} title="Final Domain Status" />
      </div>
    </div>
  );
}
