import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { ShieldCheck, Award, ChevronRight, Loader2 } from "lucide-react";
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
      // Start Post Test
      await axios.post("/api/exams/post-test", {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      navigate("/post-test");
    } catch (err) {
      console.error("Failed to start post-test:", err);
      // If already active session exists, navigate to /post-test
      navigate("/post-test");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-4">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
        <p className="text-sm font-mono text-graphite-400">Finalizing training reports...</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto space-y-8 animate-fade-in">
      <div className="text-center space-y-3 max-w-xl mx-auto">
        <div className="inline-flex items-center justify-center p-4 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 mb-2 relative">
          <Award className="w-10 h-10" />
          <ShieldCheck className="w-5 h-5 text-emerald-400 absolute bottom-2 right-2 bg-graphite-950 rounded-full" />
        </div>
        <h1 className="text-3xl font-bold tracking-tight">Training Complete!</h1>
        <p className="text-sm text-graphite-400 leading-relaxed">
          Congratulations! You have completed the practical SOC Incident Simulation training loop. All security domains have reached the required proficiency threshold.
        </p>
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        {/* Stats card */}
        <div className="glass-panel p-6 rounded-xl space-y-5 flex flex-col justify-between">
          <div className="space-y-4">
            <h3 className="text-base font-bold text-graphite-100 border-b border-graphite-800 pb-3">
              Training Metrics
            </h3>
            
            <div className="space-y-3 font-mono text-xs">
              <div className="flex justify-between py-1 border-b border-graphite-900/60">
                <span className="text-graphite-400">Scenarios Completed:</span>
                <span className="text-graphite-100 font-bold">{completedScenarios}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-graphite-900/60">
                <span className="text-graphite-400">Average Proficiency:</span>
                <span className="text-emerald-400 font-bold">
                  {(domainProficiencies.reduce((acc, d) => acc + d.proficiency_score, 0) / Math.max(1, domainProficiencies.length)).toFixed(1)}%
                </span>
              </div>
              <div className="flex justify-between py-1 border-b border-graphite-900/60">
                <span className="text-graphite-400">Status:</span>
                <span className="text-emerald-400 font-bold">READY FOR EXAM</span>
              </div>
            </div>
          </div>

          <button
            onClick={handleStartPostTest}
            className="w-full py-3 bg-emerald-500 hover:bg-emerald-400 text-emerald-950 font-bold rounded-lg shadow-lg hover:shadow-emerald-500/20 hover:-translate-y-0.5 active:translate-y-0 transition-all flex items-center justify-center gap-2 text-sm"
          >
            Start Final Post-Test
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>

        {/* Breakdown display */}
        <DomainBreakdown domains={domainProficiencies} title="Final Domain Status" />
      </div>
    </div>
  );
}
