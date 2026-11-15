import { useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import axios from "axios";
import {
  BookOpen,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Circle,
  Loader2,
  Lock,
  PlayCircle,
  Sparkles,
  AlertTriangle,
} from "lucide-react";
import MarkdownView from "../components/MarkdownView";
import { cn } from "../lib/utils";

interface TheoryDomain {
  domain_id: string;
  domain_name: string;
  title: string;
  content: string;
  source_url: string;
  proficiency_score: number;
  threshold: number;
  passed_threshold: boolean;
  recommended_focus: boolean;
  is_completed: boolean;
  completed_at?: string;
}

interface LearnDomainsPayload {
  unlocked: boolean;
  theory_completed: boolean;
  completed_count: number;
  total_domains: number;
  completed_domains: string[];
  domains: TheoryDomain[];
}

export default function LearnTheoryPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const [loading, setLoading] = useState(true);
  const [lockedMessage, setLockedMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [data, setData] = useState<LearnDomainsPayload | null>(null);
  const [selectedDomainId, setSelectedDomainId] = useState<string>(
    searchParams.get("domain") || "domain1"
  );
  const [completingDomainId, setCompletingDomainId] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const userRole = localStorage.getItem("userRole") || "learner";
  const initialQueryDomainRef = useRef<string | null>(searchParams.get("domain"));

  useEffect(() => {
    const fetchTheory = async () => {
      const token = localStorage.getItem("token");
      if (!token) {
        navigate("/login");
        return;
      }

      try {
        const res = await axios.get<LearnDomainsPayload>("/api/learn/domains", {
          headers: { Authorization: `Bearer ${token}` },
        });
        setData(res.data);

        const requested = initialQueryDomainRef.current;
        const domains = res.data.domains || [];
        if (requested && domains.some((d) => d.domain_id === requested)) {
          setSelectedDomainId(requested);
        } else if (domains.length > 0) {
          const firstRecommendedIncomplete = domains.find(
            (d) => d.recommended_focus && !d.is_completed
          );
          const firstIncomplete = domains.find((d) => !d.is_completed);
          setSelectedDomainId(
            (firstRecommendedIncomplete || firstIncomplete || domains[0]).domain_id
          );
        }
      } catch (err: unknown) {
        const errPayload = axios.isAxiosError(err) ? err.response : undefined;
        if (errPayload?.status === 401) {
          navigate("/login");
          return;
        }
        if (
          errPayload?.status === 403 &&
          errPayload?.data?.reason === "pre_test_required"
        ) {
          setLockedMessage(
            errPayload.data.error ||
              "Please complete the Diagnostic Pre-Test first to unlock Learning Theory."
          );
        } else {
          console.error("Failed to load learning theory:", err);
          setErrorMessage(
            errPayload?.data?.error ||
              "Unable to load Learning Theory modules right now. Please try again."
          );
        }
      } finally {
        setLoading(false);
      }
    };

    fetchTheory();
  }, [navigate]);

  const handleSelectDomain = (domainId: string) => {
    setSelectedDomainId(domainId);
    setSearchParams({ domain: domainId }, { replace: true });
    setActionError(null);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const handleMarkCompleted = async (domainId: string) => {
    const token = localStorage.getItem("token");
    if (!token || !data) return;

    setCompletingDomainId(domainId);
    setActionError(null);
    try {
      const res = await axios.post(
        `/api/learn/domains/${domainId}/complete`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
      const completedDomains: string[] = res.data.completed_domains || [];
      setData({
        ...data,
        theory_completed: Boolean(res.data.theory_completed),
        completed_count: res.data.completed_count ?? completedDomains.length,
        completed_domains: completedDomains,
        domains: data.domains.map((d) =>
          d.domain_id === domainId
            ? {
                ...d,
                is_completed: true,
                completed_at: d.completed_at || new Date().toISOString(),
              }
            : d
        ),
      });
    } catch (err: unknown) {
      console.error("Failed to mark domain completed:", err);
      const errMsg = axios.isAxiosError(err) ? err.response?.data?.error : undefined;
      setActionError(errMsg || "Could not save domain completion. Please retry.");
    } finally {
      setCompletingDomainId(null);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-xs font-mono text-graphite-500">
          Loading 5-domain security curriculum...
        </p>
      </div>
    );
  }

  if (lockedMessage) {
    return (
      <div className="max-w-lg mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
        <div className="w-12 h-12 rounded-full bg-amber-500/10 border border-amber-500/30 flex items-center justify-center mx-auto">
          <Lock className="w-5 h-5 text-amber-400" />
        </div>
        <div className="space-y-2">
          <h2 className="text-lg font-bold tracking-tight text-white">
            Pre-Test Required to Unlock Theory
          </h2>
          <p className="text-xs text-graphite-400 leading-relaxed">
            {lockedMessage} Completing the 30-question Diagnostic Pre-Test builds your baseline
            skill profile across all 5 CompTIA Security+ domains so you know which areas to focus on.
          </p>
        </div>
        <button
          onClick={() => navigate("/pre-test")}
          className="w-full py-2.5 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all duration-200 cursor-pointer"
        >
          Take Diagnostic Pre-Test
        </button>
      </div>
    );
  }

  if (errorMessage || !data || data.domains.length === 0) {
    return (
      <div className="max-w-lg mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-4">
        <AlertTriangle className="w-8 h-8 text-amber-400 mx-auto" />
        <h2 className="text-base font-bold text-white">Learning Theory Unavailable</h2>
        <p className="text-xs font-mono text-graphite-400">
          {errorMessage || "No curriculum modules were found."}
        </p>
      </div>
    );
  }

  const queryDomain = searchParams.get("domain");
  const effectiveDomainId =
    queryDomain && data.domains.some((d) => d.domain_id === queryDomain)
      ? queryDomain
      : selectedDomainId;

  const activeIdx = Math.max(
    0,
    data.domains.findIndex((d) => d.domain_id === effectiveDomainId)
  );
  const activeDomain = data.domains[activeIdx];
  const prevDomain = activeIdx > 0 ? data.domains[activeIdx - 1] : null;
  const nextDomain =
    activeIdx < data.domains.length - 1 ? data.domains[activeIdx + 1] : null;

  const progressPct = Math.round(
    (data.completed_count / Math.max(1, data.total_domains)) * 100
  );
  const showProficiencyMetrics = userRole === "learner" || data.domains.some((d) => d.proficiency_score > 0);

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header & Progress Gate Banner */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-xs font-mono text-emerald-400">
            <BookOpen className="w-4 h-4" />
            <span className="uppercase tracking-wider font-semibold">
              CompTIA Security+ (SY0-701) Curriculum
            </span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-white">
            Learning Theory Hub
          </h1>
          <p className="text-xs text-graphite-400 max-w-2xl leading-relaxed">
            Study all 5 domains — core theory, SIEM/KQL telemetry indicators, TP/FP triage, and
            3-phase Incident Response playbooks. Complete all 5 domains to unlock Adaptive Scenarios.
          </p>
        </div>

        <div className="w-full lg:w-80 space-y-3 bg-graphite-950/70 border border-graphite-800 rounded-lg p-4 shrink-0">
          <div className="flex items-center justify-between text-xs font-mono">
            <span className="text-graphite-400">Study Progress</span>
            <span
              className={cn(
                "font-bold",
                data.theory_completed ? "text-emerald-400" : "text-white"
              )}
            >
              {data.completed_count} / {data.total_domains} Domains Completed
            </span>
          </div>
          <div className="h-2 w-full bg-graphite-900 rounded overflow-hidden border border-graphite-800">
            <div
              className={cn(
                "h-full transition-all duration-500",
                data.theory_completed ? "bg-emerald-500" : "bg-indigo-500"
              )}
              style={{ width: `${progressPct}%` }}
            />
          </div>

          {data.theory_completed ? (
            <button
              onClick={() => navigate("/simulation")}
              className="w-full py-2 px-3 bg-white text-black font-semibold text-xs rounded-md hover:bg-white/90 transition-all flex items-center justify-center gap-2 cursor-pointer"
            >
              <PlayCircle className="w-4 h-4" />
              Start Adaptive Simulation
            </button>
          ) : (
            <p className="text-[11px] font-mono text-graphite-400">
              Mark all {data.total_domains} domains as completed below to unlock Simulation Scenarios.
            </p>
          )}
        </div>
      </div>

      {/* Main Layout: Domain Navigation Sidebar + Reader Panel */}
      <div className="grid lg:grid-cols-12 gap-6 items-start">
        <aside className="lg:col-span-4 space-y-2.5">
          {data.domains.map((domain, idx) => {
            const isActive = domain.domain_id === activeDomain.domain_id;
            return (
              <button
                key={domain.domain_id}
                type="button"
                onClick={() => handleSelectDomain(domain.domain_id)}
                className={cn(
                  "w-full text-left p-4 rounded-lg border transition-all duration-150 space-y-2 cursor-pointer",
                  isActive
                    ? "bg-graphite-900 border-white/40 shadow-sm"
                    : "bg-graphite-900/50 border-graphite-800/80 hover:border-graphite-700 hover:bg-graphite-900"
                )}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="space-y-0.5">
                    <span className="text-[10px] font-mono uppercase tracking-wider text-graphite-500">
                      Domain {idx + 1}
                    </span>
                    <h3 className="text-sm font-semibold text-white leading-snug">
                      {domain.domain_name}
                    </h3>
                  </div>
                  {domain.is_completed ? (
                    <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                  ) : (
                    <Circle className="w-4 h-4 text-graphite-600 shrink-0 mt-0.5" />
                  )}
                </div>

                <div className="flex flex-wrap items-center gap-2 pt-1">
                  {showProficiencyMetrics && (
                    <span
                      className={cn(
                        "px-2 py-0.5 rounded text-[10px] font-mono font-semibold border",
                        domain.passed_threshold
                          ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                          : "bg-amber-500/10 text-amber-400 border-amber-500/20"
                      )}
                    >
                      Proficiency: {domain.proficiency_score.toFixed(1)}%
                    </span>
                  )}

                  {domain.recommended_focus && (
                    <span className="px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-indigo-500/15 text-indigo-300 border border-indigo-500/30 flex items-center gap-1">
                      <Sparkles className="w-2.5 h-2.5" />
                      Recommended Focus
                    </span>
                  )}

                  <span
                    className={cn(
                      "text-[10px] font-mono ml-auto",
                      domain.is_completed ? "text-emerald-400" : "text-graphite-500"
                    )}
                  >
                    {domain.is_completed ? "Completed" : "Unread"}
                  </span>
                </div>
              </button>
            );
          })}
        </aside>

        <section className="lg:col-span-8 bg-graphite-900 border border-graphite-800 rounded-lg overflow-hidden">
          {/* Domain Reader Top Bar */}
          <div className="px-6 py-4 border-b border-graphite-800 bg-graphite-950/60 flex flex-wrap items-center justify-between gap-3">
            <div className="flex flex-wrap items-center gap-2">
              <span className="px-2 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-[10px] font-mono text-graphite-300 uppercase">
                {activeDomain.domain_id}
              </span>
              {activeDomain.recommended_focus && (
                <span className="px-2 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-[10px] font-mono text-indigo-300 font-semibold">
                  Recommended Focus Domain
                </span>
              )}
              {showProficiencyMetrics && (
                <span className="text-[11px] font-mono text-graphite-400">
                  Current Proficiency:{" "}
                  <strong
                    className={
                      activeDomain.passed_threshold ? "text-emerald-400" : "text-amber-400"
                    }
                  >
                    {activeDomain.proficiency_score.toFixed(1)}%
                  </strong>{" "}
                  / {activeDomain.threshold}%
                </span>
              )}
            </div>

            {activeDomain.is_completed && (
              <span className="inline-flex items-center gap-1.5 text-xs font-mono font-semibold text-emerald-400">
                <CheckCircle2 className="w-3.5 h-3.5" />
                Domain Completed
              </span>
            )}
          </div>

          {/* Markdown Curriculum Body */}
          <div className="p-6 sm:p-8">
            <MarkdownView content={activeDomain.content} />
          </div>

          {/* Bottom Checkpoint & Navigation Bar */}
          <div className="px-6 py-5 border-t border-graphite-800 bg-graphite-950/60 space-y-4">
            {actionError && (
              <p className="text-xs font-mono text-red-400 bg-red-950/30 border border-red-500/30 rounded px-3 py-2">
                {actionError}
              </p>
            )}

            <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  disabled={!prevDomain}
                  onClick={() => prevDomain && handleSelectDomain(prevDomain.domain_id)}
                  className="px-3 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-xs font-mono text-graphite-300 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                >
                  <ChevronLeft className="w-3.5 h-3.5" />
                  Previous
                </button>
                <button
                  type="button"
                  disabled={!nextDomain}
                  onClick={() => nextDomain && handleSelectDomain(nextDomain.domain_id)}
                  className="px-3 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-xs font-mono text-graphite-300 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                >
                  Next
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>

              <div className="flex flex-wrap items-center gap-3 justify-end">
                {!activeDomain.is_completed ? (
                  <button
                    type="button"
                    disabled={completingDomainId === activeDomain.domain_id}
                    onClick={() => handleMarkCompleted(activeDomain.domain_id)}
                    className="px-4 py-2 rounded-md bg-emerald-500 hover:bg-emerald-400 text-black font-semibold text-xs transition-all flex items-center justify-center gap-2 disabled:opacity-50 cursor-pointer"
                  >
                    {completingDomainId === activeDomain.domain_id ? (
                      <>
                        <Loader2 className="w-3.5 h-3.5 animate-spin" />
                        Saving Progress...
                      </>
                    ) : (
                      <>
                        <CheckCircle2 className="w-4 h-4" />
                        Mark Domain as Completed
                      </>
                    )}
                  </button>
                ) : nextDomain && !nextDomain.is_completed ? (
                  <button
                    type="button"
                    onClick={() => handleSelectDomain(nextDomain.domain_id)}
                    className="px-4 py-2 rounded-md bg-white hover:bg-white/90 text-black font-semibold text-xs transition-all flex items-center justify-center gap-2 cursor-pointer"
                  >
                    Completed ✓ — Next Domain
                    <ChevronRight className="w-4 h-4" />
                  </button>
                ) : data.theory_completed ? (
                  <button
                    type="button"
                    onClick={() => navigate("/simulation")}
                    className="px-4 py-2 rounded-md bg-white hover:bg-white/90 text-black font-semibold text-xs transition-all flex items-center justify-center gap-2 cursor-pointer"
                  >
                    <PlayCircle className="w-4 h-4" />
                    Start Adaptive Simulation
                  </button>
                ) : (
                  <span className="inline-flex items-center gap-1.5 px-3 py-2 rounded-md bg-emerald-500/10 border border-emerald-500/30 text-xs font-mono font-semibold text-emerald-400">
                    <CheckCircle2 className="w-4 h-4" />
                    Completed ({data.completed_count}/{data.total_domains})
                  </span>
                )}
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
