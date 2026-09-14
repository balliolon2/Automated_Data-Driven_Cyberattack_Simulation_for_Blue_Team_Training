import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import {
  ShieldAlert,
  Search,
  Filter,
  Eye,
  CheckCircle2,
  XCircle,
  Award,
  BookOpen,
  ArrowRight,
  X,
  FileText,
  HelpCircle,
  Activity,
} from "lucide-react";
import { cn } from "../lib/utils";

interface SubmissionSummary {
  session_id: string;
  learner_nickname: string;
  scenario_id: string;
  scenario_title: string;
  domain_id: string;
  domain_name: string;
  final_score: number;
  tp_fp_correct: boolean;
  findings_found_count: number;
  total_findings_count: number;
  actions_count: number;
  completed_at: string;
}

interface TimelineAction {
  action_id: number;
  step_order: number;
  action_type: string;
  payload: any;
  is_correct: boolean | null;
  points: number;
  timestamp: string;
}

interface SubmissionDetail {
  session_id: string;
  learner_nickname: string;
  scenario_id: string;
  scenario_title: string;
  scenario_description: string;
  domain_id: string;
  is_true_positive: boolean;
  tp_fp_explanation: string;
  learner_choice: boolean | null;
  tp_fp_correct: boolean;
  tp_fp_points: number;
  discovered_findings: string[];
  expected_findings: any;
  selected_actions: string[];
  expected_actions: any;
  final_score: number;
  timeline: TimelineAction[];
  completed_at: string;
}

export default function SpecialistReviewPage() {
  const navigate = useNavigate();
  const token = localStorage.getItem("token");
  const userRole = localStorage.getItem("userRole") || "learner";

  const [submissions, setSubmissions] = useState<SubmissionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDomain, setSelectedDomain] = useState<string>("");
  const [maxScoreFilter, setMaxScoreFilter] = useState<string>("");
  const [searchQuery, setSearchQuery] = useState<string>("");

  // Modal detail state
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detail, setDetail] = useState<SubmissionDetail | null>(null);

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    if (userRole !== "specialist" && userRole !== "admin") {
      setLoading(false);
      return;
    }
    fetchSubmissions();
  }, [selectedDomain, maxScoreFilter]);

  const fetchSubmissions = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (selectedDomain) params.append("domain_id", selectedDomain);
      if (maxScoreFilter) params.append("max_score", maxScoreFilter);

      const res = await axios.get(`/api/specialist/submissions?${params.toString()}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setSubmissions(res.data);
    } catch (err) {
      console.error("Failed to fetch submissions:", err);
    } finally {
      setLoading(false);
    }
  };

  const openDetailModal = async (sessionId: string) => {
    setActiveSessionId(sessionId);
    setDetailLoading(true);
    try {
      const res = await axios.get(`/api/specialist/submissions/${sessionId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setDetail(res.data);
    } catch (err) {
      console.error("Failed to load submission detail:", err);
    } finally {
      setDetailLoading(false);
    }
  };

  const closeDetailModal = () => {
    setActiveSessionId(null);
    setDetail(null);
  };

  // Filter by search query (nickname or scenario title)
  const filteredSubmissions = submissions.filter((s) => {
    const q = searchQuery.toLowerCase();
    return (
      s.learner_nickname.toLowerCase().includes(q) ||
      s.scenario_title.toLowerCase().includes(q) ||
      s.domain_id.toLowerCase().includes(q)
    );
  });

  if (userRole !== "specialist" && userRole !== "admin") {
    return (
      <div className="max-w-xl mx-auto text-center py-16 space-y-4">
        <div className="w-12 h-12 rounded-full bg-amber-500/10 text-amber-400 flex items-center justify-center mx-auto border border-amber-500/20">
          <ShieldAlert className="w-6 h-6" />
        </div>
        <h2 className="text-xl font-bold text-white">Specialist Portal Access Required</h2>
        <p className="text-graphite-400 text-sm leading-relaxed">
          The Submission Review Console is restricted to approved Cyber Specialists and Administrators.
          If you are an industry practitioner or security expert, please apply through your Profile.
        </p>
        <button
          onClick={() => navigate("/profile")}
          className="mt-4 px-4 py-2 rounded-md bg-white text-black font-semibold text-xs hover:bg-white/90 transition-all"
        >
          Go to Profile & Apply
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-graphite-800 pb-6">
        <div>
          <div className="flex items-center gap-2.5 mb-1">
            <h1 className="text-2xl font-bold tracking-tight text-white">Specialist Review Console</h1>
            <span className="px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              PRIVACY ANONYMIZED
            </span>
          </div>
          <p className="text-xs text-graphite-400">
            Inspect learner simulation attempts anonymously by nickname. Identify misconceptions and author educational analysis threads.
          </p>
        </div>

        <button
          onClick={() => navigate("/discussions/new")}
          className="flex items-center gap-2 px-4 py-2 rounded-md bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm transition-all cursor-pointer self-start md:self-auto"
        >
          <FileText className="w-4 h-4" />
          <span>Write Analysis Post</span>
        </button>
      </div>

      {/* Filters Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 p-4 rounded-lg bg-graphite-900/50 border border-graphite-800 text-xs">
        {/* Search */}
        <div className="relative">
          <Search className="w-3.5 h-3.5 text-graphite-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search nickname, scenario..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 rounded bg-graphite-950 border border-graphite-800 text-graphite-200 placeholder-graphite-500 focus:outline-none focus:border-graphite-600 font-mono text-xs"
          />
        </div>

        {/* Domain Filter */}
        <div className="flex items-center gap-2">
          <Filter className="w-3.5 h-3.5 text-graphite-500 shrink-0" />
          <select
            value={selectedDomain}
            onChange={(e) => setSelectedDomain(e.target.value)}
            className="w-full px-2.5 py-1.5 rounded bg-graphite-950 border border-graphite-800 text-graphite-200 focus:outline-none focus:border-graphite-600 text-xs"
          >
            <option value="">All Security Domains</option>
            <option value="domain1">Domain 1: Security Concepts</option>
            <option value="domain2">Domain 2: Threats & Mitigations</option>
            <option value="domain3">Domain 3: Security Architecture</option>
            <option value="domain4">Domain 4: Security Operations</option>
            <option value="domain5">Domain 5: Security Management</option>
          </select>
        </div>

        {/* Score Threshold Filter */}
        <div className="flex items-center gap-2">
          <Award className="w-3.5 h-3.5 text-graphite-500 shrink-0" />
          <select
            value={maxScoreFilter}
            onChange={(e) => setMaxScoreFilter(e.target.value)}
            className="w-full px-2.5 py-1.5 rounded bg-graphite-950 border border-graphite-800 text-graphite-200 focus:outline-none focus:border-graphite-600 text-xs"
          >
            <option value="">All Scores</option>
            <option value="50">Show Struggling (Score &le; 50%)</option>
            <option value="70">Show Moderate (Score &le; 70%)</option>
            <option value="85">Show Below Average (Score &le; 85%)</option>
          </select>
        </div>
      </div>

      {/* Submissions Table / Feed */}
      {loading ? (
        <div className="py-20 text-center text-graphite-500 text-xs font-mono">
          Loading learner simulation submissions...
        </div>
      ) : filteredSubmissions.length === 0 ? (
        <div className="py-16 text-center border border-dashed border-graphite-800 rounded-lg space-y-2">
          <HelpCircle className="w-8 h-8 text-graphite-600 mx-auto" />
          <p className="text-sm font-semibold text-graphite-300">No submissions found</p>
          <p className="text-xs text-graphite-500">Try adjusting your domain or score filter criteria.</p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-graphite-800 bg-graphite-950">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-graphite-900 border-b border-graphite-800 text-graphite-400 font-mono">
                <th className="p-3 font-semibold">LEARNER NICKNAME</th>
                <th className="p-3 font-semibold">SCENARIO & DOMAIN</th>
                <th className="p-3 font-semibold">SCORE</th>
                <th className="p-3 font-semibold">TRIAGE (TP/FP)</th>
                <th className="p-3 font-semibold">FINDINGS</th>
                <th className="p-3 font-semibold">COMPLETED</th>
                <th className="p-3 text-right font-semibold">ACTION</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-graphite-800/60">
              {filteredSubmissions.map((sub) => {
                const scoreColor =
                  sub.final_score >= 80
                    ? "text-emerald-400 bg-emerald-950/30 border-emerald-500/20"
                    : sub.final_score >= 60
                    ? "text-amber-400 bg-amber-950/30 border-amber-500/20"
                    : "text-red-400 bg-red-950/30 border-red-500/20";

                return (
                  <tr key={sub.session_id} className="hover:bg-graphite-900/40 transition-colors">
                    <td className="p-3">
                      <span className="font-mono font-medium text-graphite-200 bg-graphite-900 px-2 py-0.5 rounded border border-graphite-800 text-xs">
                        {sub.learner_nickname}
                      </span>
                    </td>
                    <td className="p-3">
                      <div className="font-medium text-white truncate max-w-xs">{sub.scenario_title}</div>
                      <span className="text-[10px] font-mono text-graphite-400">{sub.domain_id}</span>
                    </td>
                    <td className="p-3">
                      <span className={cn("px-2 py-0.5 rounded text-xs font-mono font-bold border", scoreColor)}>
                        {sub.final_score.toFixed(1)}%
                      </span>
                    </td>
                    <td className="p-3">
                      {sub.tp_fp_correct ? (
                        <span className="inline-flex items-center gap-1 text-[11px] font-mono text-emerald-400">
                          <CheckCircle2 className="w-3.5 h-3.5" /> Correct
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 text-[11px] font-mono text-red-400">
                          <XCircle className="w-3.5 h-3.5" /> Misjudged
                        </span>
                      )}
                    </td>
                    <td className="p-3 font-mono text-graphite-300">
                      {sub.findings_found_count} discovered
                    </td>
                    <td className="p-3 text-graphite-400 font-mono text-[11px]">
                      {sub.completed_at ? new Date(sub.completed_at).toLocaleDateString() : "In progress"}
                    </td>
                    <td className="p-3 text-right">
                      <button
                        onClick={() => openDetailModal(sub.session_id)}
                        className="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-graphite-900 border border-graphite-800 hover:border-graphite-600 text-graphite-300 hover:text-white font-medium text-xs transition-colors cursor-pointer"
                      >
                        <Eye className="w-3.5 h-3.5" />
                        <span>Inspect Trace</span>
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Deep-Dive Investigation Modal */}
      {activeSessionId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in">
          <div className="w-full max-w-3xl max-h-[90vh] flex flex-col bg-graphite-950 border border-graphite-800 rounded-xl shadow-2xl overflow-hidden">
            {/* Modal Header */}
            <div className="p-4 sm:p-5 border-b border-graphite-800 flex items-center justify-between bg-graphite-900/70">
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-mono px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 font-semibold">
                    SUBMISSION TRACE
                  </span>
                  {detail && (
                    <span className="text-xs font-mono text-graphite-400">
                      Learner: <span className="text-white font-semibold">{detail.learner_nickname}</span>
                    </span>
                  )}
                </div>
                <h3 className="text-base font-bold text-white mt-1">
                  {detail?.scenario_title || "Loading scenario attempt..."}
                </h3>
              </div>

              <button
                onClick={closeDetailModal}
                className="p-1.5 rounded-lg text-graphite-400 hover:text-white hover:bg-graphite-800 transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Modal Body */}
            <div className="flex-1 overflow-y-auto p-5 space-y-6 text-xs">
              {detailLoading || !detail ? (
                <div className="py-20 text-center text-graphite-500 font-mono">
                  Loading investigation trace details...
                </div>
              ) : (
                <>
                  {/* Score & Triage Summary Banner */}
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 p-4 rounded-lg bg-graphite-900/60 border border-graphite-800">
                    <div>
                      <div className="text-graphite-400 text-[10px] font-mono uppercase">FINAL SCORE</div>
                      <div className="text-xl font-bold font-mono text-white mt-0.5">
                        {detail.final_score.toFixed(1)}%
                      </div>
                    </div>

                    <div>
                      <div className="text-graphite-400 text-[10px] font-mono uppercase">TRIAGE ACCURACY</div>
                      <div className="mt-0.5 flex items-center gap-1.5">
                        {detail.tp_fp_correct ? (
                          <span className="inline-flex items-center gap-1 text-emerald-400 font-semibold font-mono">
                            <CheckCircle2 className="w-3.5 h-3.5" /> Correct (
                            {detail.learner_choice ? "True Positive" : "False Positive"})
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-red-400 font-semibold font-mono">
                            <XCircle className="w-3.5 h-3.5" /> Misjudged (Learner:{" "}
                            {detail.learner_choice ? "TP" : "FP"})
                          </span>
                        )}
                      </div>
                    </div>

                    <div>
                      <div className="text-graphite-400 text-[10px] font-mono uppercase">DOMAIN</div>
                      <div className="text-xs font-mono text-graphite-200 mt-0.5">
                        {detail.domain_id}
                      </div>
                    </div>
                  </div>

                  {/* Triage Explanation */}
                  <div className="space-y-1.5 p-3.5 rounded-lg bg-graphite-900/40 border border-graphite-800/80">
                    <div className="font-semibold text-graphite-200 flex items-center gap-1.5">
                      <HelpCircle className="w-3.5 h-3.5 text-indigo-400" />
                      <span>Triage Ground Truth & Misconception Notes</span>
                    </div>
                    <p className="text-graphite-300 leading-relaxed text-xs">
                      {detail.tp_fp_explanation}
                    </p>
                  </div>

                  {/* Findings Breakdown */}
                  <div className="space-y-2">
                    <h4 className="font-semibold text-white flex items-center gap-1.5">
                      <Search className="w-3.5 h-3.5 text-amber-400" />
                      <span>Artifact Findings Discovered</span>
                    </h4>
                    {detail.discovered_findings && detail.discovered_findings.length > 0 ? (
                      <div className="flex flex-wrap gap-1.5">
                        {detail.discovered_findings.map((finding, idx) => (
                          <span
                            key={idx}
                            className="px-2 py-1 rounded bg-graphite-900 border border-graphite-800 text-graphite-200 font-mono text-[11px]"
                          >
                            ✓ {finding}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <p className="text-graphite-500 italic">No artifact findings discovered during simulation.</p>
                    )}
                  </div>

                  {/* Actions Taken */}
                  <div className="space-y-2">
                    <h4 className="font-semibold text-white flex items-center gap-1.5">
                      <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                      <span>Containment & Mitigation Actions Selected</span>
                    </h4>
                    {detail.selected_actions && detail.selected_actions.length > 0 ? (
                      <div className="flex flex-wrap gap-1.5">
                        {detail.selected_actions.map((act, idx) => (
                          <span
                            key={idx}
                            className="px-2 py-1 rounded bg-graphite-900 border border-graphite-800 text-graphite-200 font-mono text-[11px]"
                          >
                            • {act}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <p className="text-graphite-500 italic">No remediation actions were executed.</p>
                    )}
                  </div>

                  {/* Chronological Timeline */}
                  <div className="space-y-2.5">
                    <h4 className="font-semibold text-white flex items-center gap-1.5">
                      <Activity className="w-3.5 h-3.5 text-indigo-400" />
                      <span>Investigation Action Log</span>
                    </h4>
                    <div className="border border-graphite-800 rounded-lg divide-y divide-graphite-800/60 overflow-hidden">
                      {detail.timeline && detail.timeline.length > 0 ? (
                        detail.timeline.map((item) => (
                          <div
                            key={item.action_id}
                            className="p-3 flex items-center justify-between bg-graphite-950/40 hover:bg-graphite-900/30 transition-colors"
                          >
                            <div className="flex items-center gap-3">
                              <span className="w-5 h-5 rounded-full bg-graphite-900 border border-graphite-800 text-[10px] font-mono font-bold text-graphite-400 flex items-center justify-center shrink-0">
                                {item.step_order}
                              </span>
                              <div>
                                <div className="font-medium text-white font-mono text-xs">
                                  {item.action_type}
                                </div>
                                <div className="text-[10px] text-graphite-500 font-mono">
                                  {new Date(item.timestamp).toLocaleTimeString()}
                                </div>
                              </div>
                            </div>

                            <div className="flex items-center gap-2 font-mono">
                              {item.points > 0 ? (
                                <span className="text-emerald-400 text-xs">+{item.points} pts</span>
                              ) : item.points < 0 ? (
                                <span className="text-red-400 text-xs">{item.points} pts</span>
                              ) : (
                                <span className="text-graphite-500 text-xs">0 pts</span>
                              )}
                            </div>
                          </div>
                        ))
                      ) : (
                        <div className="p-3 text-center text-graphite-500">No actions recorded.</div>
                      )}
                    </div>
                  </div>
                </>
              )}
            </div>

            {/* Modal Footer */}
            <div className="p-4 border-t border-graphite-800 bg-graphite-900/70 flex items-center justify-between">
              <button
                onClick={closeDetailModal}
                className="px-3.5 py-1.5 rounded-md border border-graphite-800 text-graphite-300 hover:text-white hover:bg-graphite-800 text-xs transition-colors"
              >
                Close
              </button>

              {detail && (
                <button
                  onClick={() => {
                    closeDetailModal();
                    navigate(
                      `/discussions/new?scenario_id=${detail.scenario_id}&scenario_title=${encodeURIComponent(
                        detail.scenario_title
                      )}`
                    );
                  }}
                  className="flex items-center gap-2 px-4 py-1.5 rounded-md bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs transition-all shadow-sm cursor-pointer"
                >
                  <BookOpen className="w-3.5 h-3.5" />
                  <span>Author Analysis on this Scenario</span>
                  <ArrowRight className="w-3.5 h-3.5" />
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
