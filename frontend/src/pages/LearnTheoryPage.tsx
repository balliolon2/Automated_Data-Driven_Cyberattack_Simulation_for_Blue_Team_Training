import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import axios from "axios";
import {
  AlertTriangle,
  BookOpen,
  CheckCircle2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Circle,
  ExternalLink,
  Eye,
  FileText,
  HelpCircle,
  Layers,
  Loader2,
  Lock,
  PlayCircle,
  Sparkles,
  XCircle,
} from "lucide-react";
import MarkdownView from "../components/MarkdownView";
import { cn } from "../lib/utils";

interface TheorySubtopicSummary {
  sub_topic_id: string;
  domain_id: string;
  order_index: number;
  title: string;
  is_completed: boolean;
  completed_at?: string;
  sample_attempted_count: number;
  sample_total_count: number;
}

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
  completed_subtopics_in_domain: number;
  total_subtopics_in_domain: number;
  subtopics: TheorySubtopicSummary[];
}

interface LearnDomainsPayload {
  unlocked: boolean;
  theory_completed: boolean;
  completed_count: number;
  total_domains: number;
  completed_subtopic_count: number;
  total_subtopics: number;
  completed_domains: string[];
  completed_subtopics: string[];
  domains: TheoryDomain[];
}

interface LearnSampleQuestion {
  question_id: string;
  order_index: number;
  question_text: string;
  options: Record<string, string>;
  is_fallback: boolean;
  submitted_answer?: string;
  is_correct?: boolean;
  answered_at?: string;
  correct_answer?: string;
  explanation?: string;
}

interface LearnSubtopicDetail {
  sub_topic_id: string;
  domain_id: string;
  order_index: number;
  title: string;
  content: string;
  source_url: string;
  is_completed: boolean;
  completed_at?: string;
  preview_mode: boolean;
  previous_sub_topic_id?: string;
  next_sub_topic_id?: string;
  attempted_count: number;
  required_count: number;
  ready_to_complete: boolean;
  sample_questions: LearnSampleQuestion[];
}

interface AttemptSampleResponse {
  sub_topic_id: string;
  question_id: string;
  submitted_answer: string;
  is_correct: boolean;
  correct_answer: string;
  explanation: string;
  attempted_count: number;
  required_count: number;
  ready_to_complete: boolean;
  already_submitted: boolean;
  preview_mode: boolean;
}

interface CompleteSubtopicResponse {
  sub_topic_id: string;
  domain_id: string;
  is_completed: boolean;
  domain_auto_completed: boolean;
  completed_subtopics_in_domain: number;
  total_subtopics_in_domain: number;
  completed_subtopic_count: number;
  total_subtopics: number;
  completed_subtopics: string[];
  completed_domains: string[];
  completed_count: number;
  total_domains: number;
  theory_completed: boolean;
  preview_mode: boolean;
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
  const [selectedTopicId, setSelectedTopicId] = useState<string | null>(
    searchParams.get("topic")
  );
  const [expandedDomains, setExpandedDomains] = useState<Record<string, boolean>>({});

  const [subDetail, setSubDetail] = useState<LearnSubtopicDetail | null>(null);
  const [subLoading, setSubLoading] = useState(false);
  const [subError, setSubError] = useState<string | null>(null);

  // Local radio choice per question before clicking Submit
  const [draftAnswers, setDraftAnswers] = useState<Record<string, string>>({});
  const [submittingQuestionId, setSubmittingQuestionId] = useState<string | null>(null);
  const [completingSubtopicId, setCompletingSubtopicId] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [unlockNotice, setUnlockNotice] = useState<{
    title: string;
    detail: string;
    allComplete: boolean;
  } | null>(null);

  const userRole = localStorage.getItem("userRole") || "learner";
  const initialQueryDomainRef = useRef<string | null>(searchParams.get("domain"));
  const initialQueryTopicRef = useRef<string | null>(searchParams.get("topic"));

  const loadSubtopicDetail = useCallback(
    async (subTopicId: string) => {
      const token = localStorage.getItem("token");
      if (!token) {
        navigate("/login");
        return;
      }

      setSubLoading(true);
      setSubError(null);
      setActionError(null);
      try {
        const res = await axios.get<LearnSubtopicDetail>(
          `/api/learn/subtopics/${encodeURIComponent(subTopicId)}`,
          { headers: { Authorization: `Bearer ${token}` } }
        );
        setSubDetail(res.data);

        // Keep sidebar attempted count in sync once assigned
        setData((prev) => {
          if (!prev) return prev;
          return {
            ...prev,
            domains: prev.domains.map((dom) => ({
              ...dom,
              subtopics: dom.subtopics.map((st) =>
                st.sub_topic_id === res.data.sub_topic_id
                  ? {
                      ...st,
                      sample_attempted_count: res.data.attempted_count,
                      sample_total_count: Math.max(1, res.data.required_count || 2),
                    }
                  : st
              ),
            })),
          };
        });
      } catch (err: unknown) {
        if (axios.isAxiosError(err) && err.response?.status === 401) {
          navigate("/login");
          return;
        }
        console.error("Failed to load sub-topic detail:", err);
        setSubError(
          axios.isAxiosError(err)
            ? err.response?.data?.error || "Unable to load this Sub-topic right now."
            : "Unable to load this Sub-topic right now."
        );
      } finally {
        setSubLoading(false);
      }
    },
    [navigate]
  );

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

        const domains = res.data.domains || [];
        const requestedTopic = initialQueryTopicRef.current;
        const requestedDomain = initialQueryDomainRef.current;

        let targetDomainId = domains[0]?.domain_id || "domain1";
        let targetTopicId: string | null = null;

        if (requestedTopic && requestedTopic !== "overview") {
          for (const d of domains) {
            if (d.subtopics?.some((s) => s.sub_topic_id === requestedTopic)) {
              targetDomainId = d.domain_id;
              targetTopicId = requestedTopic;
              break;
            }
          }
        }

        if (!targetTopicId) {
          if (requestedDomain && domains.some((d) => d.domain_id === requestedDomain)) {
            targetDomainId = requestedDomain;
          } else if (domains.length > 0) {
            const firstRecIncomplete = domains.find(
              (d) => d.recommended_focus && !d.is_completed
            );
            const firstIncomplete = domains.find((d) => !d.is_completed);
            targetDomainId = (firstRecIncomplete || firstIncomplete || domains[0]).domain_id;
          }
        }

        setSelectedDomainId(targetDomainId);
        setSelectedTopicId(targetTopicId);
        setExpandedDomains({ [targetDomainId]: true });

        if (targetTopicId) {
          loadSubtopicDetail(targetTopicId);
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
  }, [navigate, loadSubtopicDetail]);

  const handleSelectOverview = (domainId: string) => {
    setSelectedDomainId(domainId);
    setSelectedTopicId(null);
    setSubDetail(null);
    setSubError(null);
    setActionError(null);
    setUnlockNotice(null);
    setExpandedDomains((prev) => ({ ...prev, [domainId]: true }));
    setSearchParams({ domain: domainId, topic: "overview" }, { replace: true });
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const handleSelectSubtopic = (subTopicId: string, parentDomainId?: string) => {
    const resolvedDomain =
      parentDomainId ||
      data?.domains.find((d) => d.subtopics.some((s) => s.sub_topic_id === subTopicId))
        ?.domain_id ||
      selectedDomainId;

    setSelectedDomainId(resolvedDomain);
    setSelectedTopicId(subTopicId);
    setExpandedDomains((prev) => ({ ...prev, [resolvedDomain]: true }));
    setUnlockNotice(null);
    setSearchParams({ domain: resolvedDomain, topic: subTopicId }, { replace: true });
    window.scrollTo({ top: 0, behavior: "smooth" });
    loadSubtopicDetail(subTopicId);
  };

  const toggleDomainExpand = (domainId: string) => {
    setExpandedDomains((prev) => ({ ...prev, [domainId]: !prev[domainId] }));
  };

  const handleSubmitSampleAnswer = async (questionId: string) => {
    if (!subDetail) return;
    const chosen = draftAnswers[questionId];
    if (!chosen) return;
    const token = localStorage.getItem("token");
    if (!token) {
      navigate("/login");
      return;
    }

    setSubmittingQuestionId(questionId);
    setActionError(null);
    try {
      const res = await axios.post<AttemptSampleResponse>(
        `/api/learn/subtopics/${encodeURIComponent(
          subDetail.sub_topic_id
        )}/sample-questions/${encodeURIComponent(questionId)}/attempt`,
        { answer: chosen },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      const result = res.data;

      setSubDetail((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          attempted_count: result.attempted_count,
          required_count: result.required_count,
          ready_to_complete: result.ready_to_complete,
          sample_questions: prev.sample_questions.map((q) =>
            q.question_id === questionId
              ? {
                  ...q,
                  submitted_answer: result.submitted_answer,
                  is_correct: result.is_correct,
                  correct_answer: result.correct_answer,
                  explanation: result.explanation,
                  answered_at: q.answered_at || new Date().toISOString(),
                }
              : q
          ),
        };
      });

      setData((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          domains: prev.domains.map((dom) =>
            dom.domain_id !== subDetail.domain_id
              ? dom
              : {
                  ...dom,
                  subtopics: dom.subtopics.map((st) =>
                    st.sub_topic_id === subDetail.sub_topic_id
                      ? {
                          ...st,
                          sample_attempted_count: result.attempted_count,
                          sample_total_count: Math.max(1, result.required_count),
                        }
                      : st
                  ),
                }
          ),
        };
      });
    } catch (err: unknown) {
      console.error("Failed to submit sample question answer:", err);
      const errMsg = axios.isAxiosError(err) ? err.response?.data?.error : undefined;
      setActionError(errMsg || "Could not evaluate your answer. Please retry.");
    } finally {
      setSubmittingQuestionId(null);
    }
  };

  const handleCompleteSubtopic = async (subTopicId: string) => {
    const token = localStorage.getItem("token");
    if (!token || !data) return;

    setCompletingSubtopicId(subTopicId);
    setActionError(null);
    setUnlockNotice(null);
    try {
      const res = await axios.post<CompleteSubtopicResponse>(
        `/api/learn/subtopics/${encodeURIComponent(subTopicId)}/complete`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
      const payload = res.data;

      if (payload.preview_mode) {
        setSubDetail((prev) => (prev ? { ...prev, is_completed: true } : prev));
        setUnlockNotice({
          title: "Staff Preview Mode",
          detail: "Completion check simulated without writing to learner tracking tables.",
          allComplete: false,
        });
        return;
      }

      const nowIso = new Date().toISOString();
      const doneSubSet = new Set(payload.completed_subtopics || []);
      const doneDomSet = new Set(payload.completed_domains || []);

      setSubDetail((prev) =>
        prev && prev.sub_topic_id === subTopicId
          ? { ...prev, is_completed: true, completed_at: prev.completed_at || nowIso }
          : prev
      );

      setData((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          theory_completed: Boolean(payload.theory_completed),
          completed_count: payload.completed_count ?? doneDomSet.size,
          completed_subtopic_count: payload.completed_subtopic_count ?? doneSubSet.size,
          completed_domains: payload.completed_domains || prev.completed_domains,
          completed_subtopics: payload.completed_subtopics || prev.completed_subtopics,
          domains: prev.domains.map((dom) => {
            const updatedSubs = dom.subtopics.map((st) =>
              doneSubSet.has(st.sub_topic_id)
                ? { ...st, is_completed: true, completed_at: st.completed_at || nowIso }
                : st
            );
            const doneInDom = updatedSubs.filter((s) => s.is_completed).length;
            const domDone = doneDomSet.has(dom.domain_id);
            return {
              ...dom,
              is_completed: domDone,
              completed_at: domDone ? dom.completed_at || nowIso : dom.completed_at,
              completed_subtopics_in_domain: doneInDom,
              subtopics: updatedSubs,
            };
          }),
        };
      });

      if (payload.theory_completed) {
        setUnlockNotice({
          title: "28 / 28 Sub-topics Completed — Simulation Unlocked!",
          detail:
            "All 5 Security+ domains are now complete. You are ready to enter Adaptive SOC Scenarios.",
          allComplete: true,
        });
      } else if (payload.domain_auto_completed) {
        setUnlockNotice({
          title: `Parent Domain Auto-Completed (${payload.domain_id.toUpperCase()})`,
          detail: `Every Sub-topic in ${payload.domain_id} is finished. Continue to the next Sub-topic to unlock Simulations.`,
          allComplete: false,
        });
      }
    } catch (err: unknown) {
      console.error("Failed to complete sub-topic:", err);
      const errMsg = axios.isAxiosError(err) ? err.response?.data?.error : undefined;
      setActionError(errMsg || "Could not record sub-topic completion. Please retry.");
    } finally {
      setCompletingSubtopicId(null);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-xs font-mono text-graphite-500">
          Loading 5-domain / 28-subtopic Security+ curriculum...
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
            skill profile across all 5 CompTIA Security+ domains so you know which areas to focus
            on.
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

  const queryTopic = searchParams.get("topic");
  const queryDomain = searchParams.get("domain");

  let effectiveDomainId =
    queryDomain && data.domains.some((d) => d.domain_id === queryDomain)
      ? queryDomain
      : selectedDomainId;
  const effectiveTopicId =
    queryTopic === "overview" ? null : queryTopic ?? selectedTopicId;

  if (effectiveTopicId) {
    const owner = data.domains.find((d) =>
      d.subtopics.some((st) => st.sub_topic_id === effectiveTopicId)
    );
    if (owner) {
      effectiveDomainId = owner.domain_id;
    }
  }

  const activeIdx = Math.max(
    0,
    data.domains.findIndex((d) => d.domain_id === effectiveDomainId)
  );
  const activeDomain = data.domains[activeIdx];
  const prevDomain = activeIdx > 0 ? data.domains[activeIdx - 1] : null;
  const nextDomain =
    activeIdx < data.domains.length - 1 ? data.domains[activeIdx + 1] : null;

  const totalSubtopics = Math.max(1, data.total_subtopics || 28);
  const completedSubtopicsCount = data.completed_subtopic_count || 0;
  const progressPct = Math.round((completedSubtopicsCount / totalSubtopics) * 100);
  const showProficiencyMetrics =
    userRole === "learner" || data.domains.some((d) => d.proficiency_score > 0);

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header & Progress Gate Banner */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
        <div className="space-y-2 flex-1 min-w-0">
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
            Study all 5 Domain Overviews and 28 web-grounded Sub-topics with formative Security+
            practice questions. Completing both practice questions unlocks each Sub-topic; finish
            all 28 Sub-topics to unlock Adaptive Simulations.
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
              {completedSubtopicsCount} / {totalSubtopics} Sub-topics
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
          <div className="flex items-center justify-between text-[11px] font-mono text-graphite-400">
            <span>
              Domains Auto-Completed: {data.completed_count} / {data.total_domains}
            </span>
            <span>{progressPct}%</span>
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
              Complete all {totalSubtopics} Sub-topics & their sample questions to unlock Adaptive
              Simulations.
            </p>
          )}
        </div>
      </div>

      {/* Main Layout: Expandable Nested Tree + Workspace */}
      <div className="grid lg:grid-cols-12 gap-6 items-start">
        <aside className="lg:col-span-4 space-y-2.5">
          {data.domains.map((domain, idx) => {
            const isDomainSelected = domain.domain_id === activeDomain.domain_id;
            const isOverviewActive = isDomainSelected && !effectiveTopicId;
            const isExpanded = Boolean(expandedDomains[domain.domain_id] ?? isDomainSelected);

            return (
              <div
                key={domain.domain_id}
                className={cn(
                  "rounded-lg border transition-all duration-150 overflow-hidden",
                  isDomainSelected
                    ? "bg-graphite-900 border-white/30 shadow-sm"
                    : "bg-graphite-900/50 border-graphite-800/80"
                )}
              >
                {/* Domain Tree Header */}
                <div className="p-3.5 space-y-2">
                  <div className="flex items-start justify-between gap-2">
                    <button
                      type="button"
                      onClick={() => toggleDomainExpand(domain.domain_id)}
                      className="flex items-start gap-2 text-left flex-1 min-w-0 cursor-pointer group"
                    >
                      <ChevronDown
                        className={cn(
                          "w-4 h-4 text-graphite-400 shrink-0 mt-0.5 transition-transform duration-150",
                          !isExpanded && "-rotate-90"
                        )}
                      />
                      <div className="min-w-0">
                        <span className="text-[10px] font-mono uppercase tracking-wider text-graphite-500 block">
                          Domain {idx + 1}
                        </span>
                        <h3 className="text-sm font-semibold text-white leading-snug group-hover:text-indigo-200 transition-colors">
                          {domain.domain_name}
                        </h3>
                      </div>
                    </button>

                    <div className="flex items-center gap-1.5 shrink-0 mt-0.5">
                      <span
                        className={cn(
                          "px-1.5 py-0.5 rounded text-[10px] font-mono border",
                          domain.is_completed
                            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                            : "bg-graphite-950 text-graphite-400 border-graphite-800"
                        )}
                      >
                        {domain.completed_subtopics_in_domain}/{domain.total_subtopics_in_domain}
                      </span>
                      {domain.is_completed ? (
                        <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                      ) : (
                        <Circle className="w-4 h-4 text-graphite-600" />
                      )}
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center gap-1.5 pl-6">
                    {showProficiencyMetrics && (
                      <span
                        className={cn(
                          "px-2 py-0.5 rounded text-[10px] font-mono font-semibold border",
                          domain.passed_threshold
                            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                            : "bg-amber-500/10 text-amber-400 border-amber-500/20"
                        )}
                      >
                        {domain.proficiency_score.toFixed(1)}%
                      </span>
                    )}

                    {domain.recommended_focus && (
                      <span className="px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-indigo-500/15 text-indigo-300 border border-indigo-500/30 flex items-center gap-1">
                        <Sparkles className="w-2.5 h-2.5" />
                        Recommended
                      </span>
                    )}
                  </div>
                </div>

                {/* Expanded Child Tree: Overview + Sub-topics */}
                {isExpanded && (
                  <div className="border-t border-graphite-800/80 bg-graphite-950/40 px-2.5 py-2 space-y-1">
                    <button
                      type="button"
                      onClick={() => handleSelectOverview(domain.domain_id)}
                      className={cn(
                        "w-full text-left px-2.5 py-1.5 rounded text-xs flex items-center justify-between gap-2 transition-colors cursor-pointer",
                        isOverviewActive
                          ? "bg-white/10 text-white font-semibold"
                          : "text-graphite-300 hover:bg-graphite-800/60 hover:text-white"
                      )}
                    >
                      <span className="flex items-center gap-2 min-w-0">
                        <Layers className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                        <span className="truncate">Domain Overview</span>
                      </span>
                      <span className="text-[10px] font-mono text-graphite-500 shrink-0">
                        Roadmap
                      </span>
                    </button>

                    {(domain.subtopics || []).map((sub) => {
                      const isSubActive = effectiveTopicId === sub.sub_topic_id;
                      return (
                        <button
                          key={sub.sub_topic_id}
                          type="button"
                          onClick={() => handleSelectSubtopic(sub.sub_topic_id, domain.domain_id)}
                          className={cn(
                            "w-full text-left px-2.5 py-1.5 rounded text-xs flex items-center justify-between gap-2 transition-colors cursor-pointer",
                            isSubActive
                              ? "bg-indigo-500/20 border border-indigo-500/40 text-white font-semibold"
                              : "text-graphite-300 hover:bg-graphite-800/60 hover:text-white border border-transparent"
                          )}
                        >
                          <span className="flex items-center gap-2 min-w-0">
                            {sub.is_completed ? (
                              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                            ) : (
                              <FileText className="w-3.5 h-3.5 text-graphite-500 shrink-0" />
                            )}
                            <span className="font-mono text-[11px] text-indigo-300 shrink-0">
                              {sub.sub_topic_id}
                            </span>
                            <span className="truncate text-[11px]">{sub.title}</span>
                          </span>
                          <span
                            className={cn(
                              "text-[10px] font-mono shrink-0 px-1.5 py-0.5 rounded",
                              sub.sample_attempted_count >= sub.sample_total_count
                                ? "text-emerald-400 bg-emerald-500/10"
                                : "text-graphite-500 bg-graphite-900"
                            )}
                          >
                            {sub.sample_attempted_count}/{sub.sample_total_count}
                          </span>
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })}
        </aside>

        {/* Right Reader & Formative Sample Question Workspace */}
        <section className="lg:col-span-8 bg-graphite-900 border border-graphite-800 rounded-lg overflow-hidden">
          {!effectiveTopicId ? (
            /* ---------------- DOMAIN OVERVIEW MODE ---------------- */
            <>
              <div className="px-6 py-4 border-b border-graphite-800 bg-graphite-950/60 flex flex-wrap items-center justify-between gap-3">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="px-2 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-[10px] font-mono text-graphite-300 uppercase">
                    {activeDomain.domain_id} • Overview
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

                <div className="flex flex-wrap items-center gap-3">
                  {/^https?:\/\//i.test(activeDomain.source_url || "") && (
                    <a
                      href={activeDomain.source_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-indigo-500/10 hover:bg-indigo-500/20 border border-indigo-500/30 text-[11px] font-mono text-indigo-300 transition-colors"
                    >
                      <ExternalLink className="w-3 h-3" />
                      Primary Reference
                    </a>
                  )}
                  {activeDomain.is_completed && (
                    <span className="inline-flex items-center gap-1.5 text-xs font-mono font-semibold text-emerald-400">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      Domain Auto-Completed
                    </span>
                  )}
                </div>
              </div>

              <div className="p-6 sm:p-8 space-y-8">
                <MarkdownView content={activeDomain.content} />

                {/* Sub-topic Checklist Roadmap Grid */}
                <div className="pt-6 border-t border-graphite-800 space-y-4">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div>
                      <h3 className="text-sm font-bold text-white tracking-tight">
                        SY0-701 Sub-topics in this Domain (
                        {activeDomain.completed_subtopics_in_domain}/
                        {activeDomain.total_subtopics_in_domain} Completed)
                      </h3>
                      <p className="text-xs text-graphite-400 mt-0.5">
                        Complete every Sub-topic and its 2 practice questions below to
                        automatically complete {activeDomain.domain_name}.
                      </p>
                    </div>
                    {activeDomain.subtopics?.[0] && (
                      <button
                        type="button"
                        onClick={() =>
                          handleSelectSubtopic(
                            activeDomain.subtopics[0].sub_topic_id,
                            activeDomain.domain_id
                          )
                        }
                        className="px-3 py-1.5 rounded-md bg-white text-black font-semibold text-xs hover:bg-white/90 transition-colors flex items-center gap-1.5 cursor-pointer"
                      >
                        Start Sub-topic {activeDomain.subtopics[0].sub_topic_id}
                        <ChevronRight className="w-3.5 h-3.5" />
                      </button>
                    )}
                  </div>

                  <div className="grid sm:grid-cols-2 gap-3">
                    {(activeDomain.subtopics || []).map((sub) => (
                      <button
                        key={sub.sub_topic_id}
                        type="button"
                        onClick={() => handleSelectSubtopic(sub.sub_topic_id, activeDomain.domain_id)}
                        className={cn(
                          "text-left p-4 rounded-lg border transition-all flex flex-col justify-between gap-3 cursor-pointer",
                          sub.is_completed
                            ? "bg-emerald-950/20 border-emerald-500/30 hover:border-emerald-400/50"
                            : "bg-graphite-950/60 border-graphite-800 hover:border-indigo-500/40 hover:bg-graphite-950"
                        )}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <div className="space-y-1">
                            <span className="inline-block px-1.5 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-[10px] font-mono text-indigo-300 font-semibold">
                              Objective {sub.sub_topic_id}
                            </span>
                            <h4 className="text-xs font-semibold text-white leading-snug">
                              {sub.title}
                            </h4>
                          </div>
                          {sub.is_completed ? (
                            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                          ) : (
                            <Circle className="w-4 h-4 text-graphite-600 shrink-0" />
                          )}
                        </div>

                        <div className="flex items-center justify-between text-[11px] font-mono pt-1 border-t border-graphite-800/60">
                          <span
                            className={
                              sub.sample_attempted_count >= sub.sample_total_count
                                ? "text-emerald-400"
                                : "text-graphite-400"
                            }
                          >
                            Questions: {sub.sample_attempted_count}/{sub.sample_total_count}
                          </span>
                          <span className="text-indigo-300 flex items-center gap-1">
                            {sub.is_completed ? "Review" : "Study"}
                            <ChevronRight className="w-3 h-3" />
                          </span>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              <div className="px-6 py-4 border-t border-graphite-800 bg-graphite-950/60 flex items-center justify-between gap-3">
                <button
                  type="button"
                  disabled={!prevDomain}
                  onClick={() => prevDomain && handleSelectOverview(prevDomain.domain_id)}
                  className="px-3 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-xs font-mono text-graphite-300 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                >
                  <ChevronLeft className="w-3.5 h-3.5" />
                  Previous Overview
                </button>
                <button
                  type="button"
                  disabled={!nextDomain}
                  onClick={() => nextDomain && handleSelectOverview(nextDomain.domain_id)}
                  className="px-3 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-xs font-mono text-graphite-300 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                >
                  Next Overview
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </>
          ) : subLoading ? (
            /* ---------------- SUB-TOPIC LOADING ---------------- */
            <div className="min-h-[420px] flex flex-col items-center justify-center gap-3 p-8">
              <Loader2 className="w-6 h-6 text-white animate-spin" />
              <p className="text-xs font-mono text-graphite-400">
                Loading Sub-topic {effectiveTopicId} & formative sample questions...
              </p>
            </div>
          ) : subError || !subDetail ? (
            /* ---------------- SUB-TOPIC ERROR ---------------- */
            <div className="p-8 text-center space-y-4">
              <AlertTriangle className="w-8 h-8 text-amber-400 mx-auto" />
              <p className="text-xs font-mono text-graphite-300">
                {subError || "Sub-topic details unavailable."}
              </p>
              <button
                type="button"
                onClick={() => handleSelectOverview(activeDomain.domain_id)}
                className="px-4 py-2 rounded-md bg-graphite-800 text-xs font-semibold text-white hover:bg-graphite-700 cursor-pointer"
              >
                Back to Domain Overview
              </button>
            </div>
          ) : (
            /* ---------------- SUB-TOPIC WORKSPACE ---------------- */
            <>
              <div className="px-6 py-4 border-b border-graphite-800 bg-graphite-950/60 flex flex-wrap items-center justify-between gap-3">
                <div className="flex flex-wrap items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handleSelectOverview(subDetail.domain_id)}
                    className="px-2 py-0.5 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-[10px] font-mono text-graphite-300 uppercase cursor-pointer"
                  >
                    {subDetail.domain_id} Overview
                  </button>
                  <span className="text-graphite-600 text-xs">/</span>
                  <span className="px-2 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-[10px] font-mono text-indigo-300 font-semibold">
                    Sub-topic {subDetail.sub_topic_id}
                  </span>
                  {subDetail.preview_mode && (
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-amber-500/15 border border-amber-500/30 text-[10px] font-mono text-amber-300 font-semibold">
                      <Eye className="w-3 h-3" />
                      Preview Mode (Staff)
                    </span>
                  )}
                </div>

                <div className="flex flex-wrap items-center gap-3">
                  {/^https?:\/\//i.test(subDetail.source_url || "") && (
                    <a
                      href={subDetail.source_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-indigo-500/10 hover:bg-indigo-500/20 border border-indigo-500/30 text-[11px] font-mono text-indigo-300 transition-colors"
                    >
                      <ExternalLink className="w-3 h-3" />
                      Web Reference
                    </a>
                  )}
                  {subDetail.is_completed && (
                    <span className="inline-flex items-center gap-1.5 text-xs font-mono font-semibold text-emerald-400">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      Sub-topic Completed
                    </span>
                  )}
                </div>
              </div>

              {/* Markdown Lesson Body */}
              <div className="p-6 sm:p-8 space-y-8">
                <MarkdownView content={subDetail.content} />

                {/* Interactive Formative Sample Questions */}
                <div className="pt-6 border-t border-graphite-800 space-y-4">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="space-y-1">
                      <h3 className="text-sm font-bold text-white flex items-center gap-2">
                        <HelpCircle className="w-4 h-4 text-indigo-400" />
                        Formative Sample Practice Questions
                      </h3>
                      <p className="text-xs text-graphite-400">
                        {subDetail.preview_mode
                          ? "Staff Preview Mode: answers are evaluated statelessly without recording learner attempts."
                          : "Your first submitted answer locks cleanly for review and does not modify your Proficiency Score."}
                      </p>
                    </div>
                    <span
                      className={cn(
                        "px-2.5 py-1 rounded text-xs font-mono font-semibold border",
                        subDetail.ready_to_complete
                          ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                          : "bg-graphite-950 text-graphite-300 border-graphite-800"
                      )}
                    >
                      Attempted: {subDetail.attempted_count} / {subDetail.required_count}
                    </span>
                  </div>

                  <div className="space-y-4">
                    {subDetail.sample_questions.map((q, qIdx) => {
                      const isSubmitted = Boolean(q.submitted_answer);
                      const chosenOption = q.submitted_answer || draftAnswers[q.question_id] || "";
                      const optionEntries = Object.entries(q.options || {}).sort(([a], [b]) =>
                        a.localeCompare(b)
                      );

                      return (
                        <div
                          key={q.question_id}
                          className="bg-graphite-950/70 border border-graphite-800 rounded-lg p-5 space-y-4"
                        >
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <span className="text-xs font-mono font-semibold text-graphite-300">
                              Sample Question #{qIdx + 1}
                            </span>
                            <div className="flex items-center gap-2">
                              {q.is_fallback ? (
                                <span className="px-2 py-0.5 rounded bg-amber-500/15 border border-amber-500/30 text-[10px] font-mono text-amber-300">
                                  Domain Related Fallback
                                </span>
                              ) : (
                                <span className="px-2 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-[10px] font-mono text-indigo-300">
                                  Objective {subDetail.sub_topic_id}
                                </span>
                              )}

                              {isSubmitted && (
                                <span
                                  className={cn(
                                    "inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-mono font-semibold border",
                                    q.is_correct
                                      ? "bg-emerald-500/15 text-emerald-300 border-emerald-500/30"
                                      : "bg-red-500/15 text-red-300 border-red-500/30"
                                  )}
                                >
                                  {q.is_correct ? (
                                    <>
                                      <CheckCircle2 className="w-3 h-3" />
                                      Correct ({q.submitted_answer})
                                    </>
                                  ) : (
                                    <>
                                      <XCircle className="w-3 h-3" />
                                      Locked ({q.submitted_answer}) • Correct: {q.correct_answer}
                                    </>
                                  )}
                                </span>
                              )}
                            </div>
                          </div>

                          <p className="text-sm text-white leading-relaxed">{q.question_text}</p>

                          <div className="space-y-2">
                            {optionEntries.map(([optKey, optText]) => {
                              const isChosen = chosenOption === optKey;
                              const isActualCorrect = isSubmitted && q.correct_answer === optKey;
                              const isWrongPick =
                                isSubmitted && isChosen && q.correct_answer !== optKey;

                              return (
                                <button
                                  key={optKey}
                                  type="button"
                                  disabled={isSubmitted || submittingQuestionId === q.question_id}
                                  onClick={() =>
                                    setDraftAnswers((prev) => ({
                                      ...prev,
                                      [q.question_id]: optKey,
                                    }))
                                  }
                                  className={cn(
                                    "w-full text-left px-3.5 py-2.5 rounded-md border text-xs transition-all flex items-start gap-3",
                                    isSubmitted ? "cursor-default" : "cursor-pointer",
                                    isActualCorrect
                                      ? "bg-emerald-500/15 border-emerald-500/50 text-emerald-100"
                                      : isWrongPick
                                      ? "bg-red-500/15 border-red-500/50 text-red-100"
                                      : isChosen
                                      ? "bg-indigo-500/20 border-indigo-400 text-white"
                                      : "bg-graphite-900 border-graphite-800 text-graphite-300 hover:border-graphite-700"
                                  )}
                                >
                                  <span className="font-mono font-bold shrink-0 mt-0.5">
                                    {optKey}.
                                  </span>
                                  <span className="leading-relaxed">{optText}</span>
                                </button>
                              );
                            })}
                          </div>

                          {!isSubmitted ? (
                            <div className="flex items-center justify-between pt-1">
                              <span className="text-[11px] font-mono text-graphite-500">
                                {draftAnswers[q.question_id]
                                  ? `Selected option ${draftAnswers[q.question_id]}`
                                  : "Select A, B, C, or D to check answer"}
                              </span>
                              <button
                                type="button"
                                disabled={
                                  !draftAnswers[q.question_id] ||
                                  submittingQuestionId === q.question_id
                                }
                                onClick={() => handleSubmitSampleAnswer(q.question_id)}
                                className="px-3.5 py-1.5 rounded-md bg-white text-black font-semibold text-xs hover:bg-white/90 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                              >
                                {submittingQuestionId === q.question_id ? (
                                  <>
                                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                    Checking...
                                  </>
                                ) : (
                                  "Check & Lock Answer"
                                )}
                              </button>
                            </div>
                          ) : (
                            q.explanation && (
                              <div className="p-3.5 rounded-md bg-graphite-900 border border-graphite-800 space-y-1">
                                <span className="text-[10px] font-mono uppercase tracking-wider text-indigo-300 font-semibold">
                                  Explanation
                                </span>
                                <p className="text-xs text-graphite-300 leading-relaxed">
                                  {q.explanation}
                                </p>
                              </div>
                            )
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>

              {/* Bottom Checkpoint & Navigation Bar */}
              <div className="px-6 py-5 border-t border-graphite-800 bg-graphite-950/60 space-y-4">
                {actionError && (
                  <p className="text-xs font-mono text-red-400 bg-red-950/30 border border-red-500/30 rounded px-3 py-2">
                    {actionError}
                  </p>
                )}

                {unlockNotice && (
                  <div className="p-4 rounded-lg bg-emerald-950/40 border border-emerald-500/40 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div className="space-y-0.5">
                      <p className="text-xs font-bold text-emerald-300">{unlockNotice.title}</p>
                      <p className="text-xs text-graphite-300">{unlockNotice.detail}</p>
                    </div>
                    {unlockNotice.allComplete && (
                      <button
                        type="button"
                        onClick={() => navigate("/simulation")}
                        className="px-3.5 py-2 rounded-md bg-white text-black font-semibold text-xs hover:bg-white/90 flex items-center gap-1.5 shrink-0 cursor-pointer"
                      >
                        <PlayCircle className="w-4 h-4" />
                        Enter Simulation
                      </button>
                    )}
                  </div>
                )}

                <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      disabled={!subDetail.previous_sub_topic_id}
                      onClick={() =>
                        subDetail.previous_sub_topic_id &&
                        handleSelectSubtopic(subDetail.previous_sub_topic_id)
                      }
                      className="px-3 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-xs font-mono text-graphite-300 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                    >
                      <ChevronLeft className="w-3.5 h-3.5" />
                      Previous Sub-topic
                    </button>
                    <button
                      type="button"
                      disabled={!subDetail.next_sub_topic_id}
                      onClick={() =>
                        subDetail.next_sub_topic_id &&
                        handleSelectSubtopic(subDetail.next_sub_topic_id)
                      }
                      className="px-3 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-xs font-mono text-graphite-300 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                    >
                      Next Sub-topic
                      <ChevronRight className="w-3.5 h-3.5" />
                    </button>
                  </div>

                  <div className="flex flex-wrap items-center gap-3 justify-end">
                    {!subDetail.is_completed ? (
                      <button
                        type="button"
                        disabled={
                          !subDetail.ready_to_complete ||
                          completingSubtopicId === subDetail.sub_topic_id
                        }
                        onClick={() => handleCompleteSubtopic(subDetail.sub_topic_id)}
                        title={
                          subDetail.ready_to_complete
                            ? "Mark this Sub-topic completed"
                            : "Attempt all assigned sample questions first"
                        }
                        className="px-4 py-2 rounded-md bg-emerald-500 hover:bg-emerald-400 text-black font-semibold text-xs transition-all flex items-center justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
                      >
                        {completingSubtopicId === subDetail.sub_topic_id ? (
                          <>
                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                            Saving...
                          </>
                        ) : (
                          <>
                            <CheckCircle2 className="w-4 h-4" />
                            {subDetail.ready_to_complete
                              ? "Mark Sub-topic as Completed"
                              : `Attempt ${Math.max(
                                  0,
                                  subDetail.required_count - subDetail.attempted_count
                                )} more question(s) to unlock`}
                          </>
                        )}
                      </button>
                    ) : subDetail.next_sub_topic_id ? (
                      <button
                        type="button"
                        onClick={() =>
                          subDetail.next_sub_topic_id &&
                          handleSelectSubtopic(subDetail.next_sub_topic_id)
                        }
                        className="px-4 py-2 rounded-md bg-white hover:bg-white/90 text-black font-semibold text-xs transition-all flex items-center justify-center gap-2 cursor-pointer"
                      >
                        Completed ✓ — Next Sub-topic ({subDetail.next_sub_topic_id})
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
                        Sub-topic Completed
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </>
          )}
        </section>
      </div>
    </div>
  );
}
