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
  Clock,
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
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError, resolveDomainName } from "../i18n/format.ts";

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

interface CompleteDomainResponse {
  domain_id: string;
  is_completed: boolean;
  completed_domains: string[];
  completed_count: number;
  total_domains: number;
  theory_completed: boolean;
  preview_mode?: boolean;
}

type UnlockNoticeState =
  | { kind: "staff_preview" }
  | { kind: "all_complete" }
  | { kind: "domain_complete"; domainId: string }
  | { kind: "domain_auto"; domainId: string }
  | null;

export default function LearnTheoryPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { locale, t } = useI18n();

  const [loading, setLoading] = useState(true);
  const [lockedErr, setLockedErr] = useState<unknown>(null);
  const [loadErr, setLoadErr] = useState<unknown>(null);
  const [data, setData] = useState<LearnDomainsPayload | null>(null);

  const [selectedDomainId, setSelectedDomainId] = useState<string>(
    searchParams.get("domain") || "domain1"
  );
  const [selectedTopicId, setSelectedTopicId] = useState<string | null>(
    searchParams.get("topic") === "overview" ? null : searchParams.get("topic")
  );
  const [expandedDomains, setExpandedDomains] = useState<Record<string, boolean>>({});
  const [expandedDeepDiveDomains, setExpandedDeepDiveDomains] = useState<Record<string, boolean>>({});
  const [overviewDeepDiveOpen, setOverviewDeepDiveOpen] = useState<boolean>(false);

  const [subDetail, setSubDetail] = useState<LearnSubtopicDetail | null>(null);
  const [subLoading, setSubLoading] = useState(false);
  const [subErrorObj, setSubErrorObj] = useState<unknown>(null);

  // Local radio choice per question before clicking Submit
  const [draftAnswers, setDraftAnswers] = useState<Record<string, string>>({});
  const [submittingQuestionId, setSubmittingQuestionId] = useState<string | null>(null);
  const [completingDomainId, setCompletingDomainId] = useState<string | null>(null);
  const [completingSubtopicId, setCompletingSubtopicId] = useState<string | null>(null);
  const [actionErrorObj, setActionErrorObj] = useState<unknown>(null);
  const [unlockNotice, setUnlockNotice] = useState<UnlockNoticeState>(null);

  const userRole = localStorage.getItem("userRole") || "learner";
  const initialQueryDomainRef = useRef<string | null>(searchParams.get("domain"));
  const initialQueryTopicRef = useRef<string | null>(searchParams.get("topic"));
  const hasInitializedSelectionRef = useRef(false);
  const activeTopicRef = useRef<string | null>(selectedTopicId);
  useEffect(() => {
    activeTopicRef.current = selectedTopicId;
  }, [selectedTopicId]);

  const loadSubtopicDetail = useCallback(
    async (subTopicId: string, langOverride?: string) => {
      const token = localStorage.getItem("token");
      if (!token) {
        navigate("/login");
        return;
      }

      const activeLang = langOverride || locale;
      setSubLoading(true);
      setSubErrorObj(null);
      setActionErrorObj(null);
      try {
        const res = await axios.get<LearnSubtopicDetail>(
          `/api/learn/subtopics/${encodeURIComponent(subTopicId)}`,
          {
            params: { lang: activeLang },
            headers: { Authorization: `Bearer ${token}` },
          }
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
        setSubErrorObj(err);
      } finally {
        setSubLoading(false);
      }
    },
    [navigate, locale]
  );

  useEffect(() => {
    let cancelled = false;
    const fetchTheory = async () => {
      const token = localStorage.getItem("token");
      if (!token) {
        navigate("/login");
        return;
      }

      try {
        const res = await axios.get<LearnDomainsPayload>("/api/learn/domains", {
          params: { lang: locale },
          headers: { Authorization: `Bearer ${token}` },
        });
        if (cancelled) return;
        setLockedErr(null);
        setLoadErr(null);
        setData(res.data);

        if (!hasInitializedSelectionRef.current) {
          hasInitializedSelectionRef.current = true;
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
            setExpandedDeepDiveDomains({ [targetDomainId]: true });
            loadSubtopicDetail(targetTopicId, locale);
          }
        } else if (activeTopicRef.current) {
          loadSubtopicDetail(activeTopicRef.current, locale);
        }
      } catch (err: unknown) {
        if (cancelled) return;
        const errPayload = axios.isAxiosError(err) ? err.response : undefined;
        if (errPayload?.status === 401) {
          navigate("/login");
          return;
        }
        if (
          errPayload?.status === 403 &&
          errPayload?.data?.reason === "pre_test_required"
        ) {
          setLockedErr(err);
        } else {
          console.error("Failed to load learning theory:", err);
          setLoadErr(err);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    fetchTheory();
    return () => {
      cancelled = true;
    };
  }, [navigate, locale, loadSubtopicDetail]);

  const handleSelectOverview = (domainId: string) => {
    setSelectedDomainId(domainId);
    setSelectedTopicId(null);
    setSubDetail(null);
    setSubErrorObj(null);
    setActionErrorObj(null);
    setUnlockNotice(null);
    setOverviewDeepDiveOpen(false);
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
    setExpandedDeepDiveDomains((prev) => ({ ...prev, [resolvedDomain]: true }));
    setUnlockNotice(null);
    setSearchParams({ domain: resolvedDomain, topic: subTopicId }, { replace: true });
    window.scrollTo({ top: 0, behavior: "smooth" });
    loadSubtopicDetail(subTopicId);
  };

  const toggleDomainExpand = (domainId: string) => {
    setExpandedDomains((prev) => ({ ...prev, [domainId]: !prev[domainId] }));
  };

  const toggleDeepDiveExpand = (domainId: string) => {
    setExpandedDeepDiveDomains((prev) => ({ ...prev, [domainId]: !prev[domainId] }));
  };

  const handleCompleteDomain = async (domainId: string) => {
    const token = localStorage.getItem("token");
    if (!token || !data) return;

    setCompletingDomainId(domainId);
    setActionErrorObj(null);
    setUnlockNotice(null);
    try {
      const res = await axios.post<CompleteDomainResponse>(
        `/api/learn/domains/${encodeURIComponent(domainId)}/complete`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
      const payload = res.data;

      if (payload.preview_mode) {
        setUnlockNotice({ kind: "staff_preview" });
        return;
      }

      const nowIso = new Date().toISOString();
      const doneDomSet = new Set(payload.completed_domains || []);

      setData((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          theory_completed: Boolean(payload.theory_completed),
          completed_count: payload.completed_count ?? doneDomSet.size,
          completed_domains: payload.completed_domains || prev.completed_domains,
          domains: prev.domains.map((dom) => {
            const domDone = doneDomSet.has(dom.domain_id);
            return {
              ...dom,
              is_completed: domDone,
              completed_at: domDone ? dom.completed_at || nowIso : dom.completed_at,
            };
          }),
        };
      });

      if (payload.theory_completed) {
        setUnlockNotice({ kind: "all_complete" });
      } else {
        setUnlockNotice({ kind: "domain_complete", domainId: payload.domain_id });
      }
    } catch (err: unknown) {
      console.error("Failed to complete domain essential:", err);
      setActionErrorObj(err);
    } finally {
      setCompletingDomainId(null);
    }
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
    setActionErrorObj(null);
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
      setActionErrorObj(err);
    } finally {
      setSubmittingQuestionId(null);
    }
  };

  const handleCompleteSubtopic = async (subTopicId: string) => {
    const token = localStorage.getItem("token");
    if (!token || !data) return;

    setCompletingSubtopicId(subTopicId);
    setActionErrorObj(null);
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
        setUnlockNotice({ kind: "staff_preview" });
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
        setUnlockNotice({ kind: "all_complete" });
      } else if (payload.domain_auto_completed) {
        setUnlockNotice({ kind: "domain_auto", domainId: payload.domain_id });
      }
    } catch (err: unknown) {
      console.error("Failed to complete sub-topic:", err);
      setActionErrorObj(err);
    } finally {
      setCompletingSubtopicId(null);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-6 h-6 text-white animate-spin" />
        <p className="text-sm font-mono text-graphite-400">
          {t("learn.loadingCurriculum")}
        </p>
      </div>
    );
  }

  if (lockedErr) {
    const lockedMessage = resolveApiError(lockedErr, t, t("errors.LEARN_PRETEST_REQUIRED"));
    return (
      <div className="max-w-lg mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-fade-in">
        <div className="w-12 h-12 rounded-full bg-amber-500/10 border border-amber-500/30 flex items-center justify-center mx-auto">
          <Lock className="w-5 h-5 text-amber-400" />
        </div>
        <div className="space-y-2">
          <h2 className="text-xl font-bold tracking-tight text-white">
            {t("learn.preTestRequiredTitle")}
          </h2>
          <p className="text-base text-graphite-300 leading-relaxed">
            {lockedMessage} {t("learn.preTestRequiredBody")}
          </p>
        </div>
        <button
          onClick={() => navigate("/pre-test")}
          className="w-full py-2.5 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 transition-all duration-200 cursor-pointer"
        >
          {t("learn.takePreTest")}
        </button>
      </div>
    );
  }

  if (loadErr || !data || data.domains.length === 0) {
    const errorMessage = loadErr ? resolveApiError(loadErr, t, t("learn.loadError")) : t("learn.unavailableEmpty");
    return (
      <div className="max-w-lg mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-4">
        <AlertTriangle className="w-8 h-8 text-amber-400 mx-auto" />
        <h2 className="text-lg font-bold text-white">{t("learn.unavailableTitle")}</h2>
        <p className="text-sm font-mono text-graphite-300">
          {errorMessage}
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
  const nextActionDomain =
    nextDomain ||
    data.domains.find((d) => !d.is_completed && d.domain_id !== activeDomain.domain_id) ||
    null;

  const totalDomains = Math.max(1, data.total_domains || 5);
  const completedDomainsCount = data.completed_count || 0;
  const progressPct = Math.round((completedDomainsCount / totalDomains) * 100);
  const totalSubtopics = Math.max(1, data.total_subtopics || 28);
  const completedSubtopicsCount = data.completed_subtopic_count || 0;
  const showProficiencyMetrics =
    userRole === "learner" || data.domains.some((d) => d.proficiency_score > 0);

  const subError = subErrorObj ? resolveApiError(subErrorObj, t, t("learn.subtopicLoadError")) : null;
  const actionError = actionErrorObj ? resolveApiError(actionErrorObj, t, t("learn.completeError")) : null;
  const activeDomainDisplayName = resolveDomainName(activeDomain.domain_id, activeDomain.domain_name, locale);

  const actionErrorBanner = actionError ? (
    <p className="text-sm font-mono text-red-400 bg-red-950/30 border border-red-500/30 rounded px-3.5 py-2">
      {actionError}
    </p>
  ) : null;

  const unlockNoticeBanner = unlockNotice ? (
    <div className="p-4 rounded-lg bg-emerald-950/40 border border-emerald-500/40 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
      <div className="space-y-1">
        <p className="text-sm font-bold text-emerald-300">
          {unlockNotice.kind === "staff_preview"
            ? t("learn.staffPreviewNoticeTitle")
            : unlockNotice.kind === "all_complete"
            ? t("learn.allCompleteNoticeTitle")
            : unlockNotice.kind === "domain_complete"
            ? t("learn.domainEssentialCompleteNoticeTitle", {
                domainUpper: unlockNotice.domainId.toUpperCase(),
              })
            : t("learn.domainAutoCompleteNoticeTitle", {
                domainUpper: unlockNotice.domainId.toUpperCase(),
              })}
        </p>
        <p className="text-sm text-graphite-200">
          {unlockNotice.kind === "staff_preview"
            ? t("learn.staffPreviewNoticeDetail")
            : unlockNotice.kind === "all_complete"
            ? t("learn.allCompleteNoticeDetail")
            : unlockNotice.kind === "domain_complete"
            ? t("learn.domainEssentialCompleteNoticeDetail", {
                domainName: resolveDomainName(unlockNotice.domainId, unlockNotice.domainId, locale),
              })
            : t("learn.domainAutoCompleteNoticeDetail", {
                domainName: resolveDomainName(unlockNotice.domainId, unlockNotice.domainId, locale),
              })}
        </p>
      </div>
      {unlockNotice.kind === "all_complete" && (
        <button
          type="button"
          onClick={() => navigate("/simulation")}
          className="px-4 py-2 rounded-md bg-white text-black font-semibold text-sm hover:bg-white/90 flex items-center gap-1.5 shrink-0 cursor-pointer"
        >
          <PlayCircle className="w-4 h-4" />
          {t("learn.enterSimulation")}
        </button>
      )}
    </div>
  ) : null;

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header & Progress Gate Banner */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
        <div className="space-y-2 flex-1 min-w-0">
          <div className="flex items-center gap-2 text-xs font-mono text-emerald-400">
            <BookOpen className="w-4 h-4" />
            <span className="uppercase tracking-wider font-semibold">
              {t("learn.curriculumBadge")}
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white">
            {t("learn.hubTitle")}
          </h1>
          <p className="text-sm sm:text-base text-graphite-300 max-w-2xl leading-relaxed">
            {t("learn.hubSubtitle")}
          </p>
        </div>

        <div className="w-full lg:w-96 space-y-3 bg-graphite-950/70 border border-graphite-800 rounded-lg p-4 shrink-0">
          <div className="flex items-center justify-between gap-3 text-xs sm:text-sm font-mono">
            <span className="text-graphite-300">{t("learn.studyProgress")}</span>
            <span
              className={cn(
                "font-bold shrink-0",
                data.theory_completed ? "text-emerald-400" : "text-white"
              )}
            >
              {t("learn.domainsRatio", { completed: completedDomainsCount, total: totalDomains })}
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
          <div className="flex items-center justify-between gap-3 text-xs font-mono text-graphite-400">
            <span>
              {t("learn.deepDiveProgressSummary", { completed: completedSubtopicsCount, total: totalSubtopics })}
            </span>
            <span className="shrink-0">{progressPct}%</span>
          </div>

          {data.theory_completed ? (
            <button
              onClick={() => navigate("/simulation")}
              className="w-full py-2.5 px-3 bg-white text-black font-semibold text-sm rounded-md hover:bg-white/90 transition-all flex items-center justify-center gap-2 cursor-pointer"
            >
              <PlayCircle className="w-4 h-4" />
              {t("learn.startAdaptiveSimulation")}
            </button>
          ) : (
            <p className="text-xs font-mono text-graphite-400 leading-relaxed">
              {t("learn.unlockRequirementHint", { total: totalDomains })}
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
            const domainDisplayName = resolveDomainName(domain.domain_id, domain.domain_name, locale);

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
                        <span className="text-xs font-mono uppercase tracking-wider text-graphite-400 block">
                          {t("learn.domainNumberLabel", { index: idx + 1 })}
                        </span>
                        <h3 className="text-sm font-semibold text-white leading-snug group-hover:text-indigo-200 transition-colors">
                          {domainDisplayName}
                        </h3>
                      </div>
                    </button>

                    <div className="flex items-center gap-1.5 shrink-0 mt-0.5">
                      <span
                        className={cn(
                          "px-2 py-0.5 rounded text-xs font-mono border",
                          domain.is_completed
                            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                            : "bg-graphite-950 text-graphite-300 border-graphite-800"
                        )}
                      >
                        {domain.is_completed ? t("learn.doneBadge") : t("learn.roadmapBadge")}
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
                          "px-2 py-0.5 rounded text-xs font-mono font-semibold border",
                          domain.passed_threshold
                            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                            : "bg-amber-500/10 text-amber-400 border-amber-500/20"
                        )}
                      >
                        {domain.proficiency_score.toFixed(1)}%
                      </span>
                    )}

                    {domain.recommended_focus && (
                      <span className="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-indigo-500/15 text-indigo-300 border border-indigo-500/30 flex items-center gap-1">
                        <Sparkles className="w-3 h-3" />
                        {t("learn.recommendedBadge")}
                      </span>
                    )}
                  </div>
                </div>

                {/* Expanded Child Tree: Domain Essential + Collapsible Optional Deep Dive */}
                {isExpanded && (
                  <div className="border-t border-graphite-800/80 bg-graphite-950/40 px-2.5 py-2 space-y-1">
                    <button
                      type="button"
                      onClick={() => handleSelectOverview(domain.domain_id)}
                      className={cn(
                        "w-full text-left px-2.5 py-2 rounded text-sm flex items-center justify-between gap-2 transition-colors cursor-pointer",
                        isOverviewActive
                          ? "bg-white/10 text-white font-semibold"
                          : "text-graphite-200 hover:bg-graphite-800/60 hover:text-white"
                      )}
                    >
                      <span className="flex items-center gap-2 min-w-0">
                        <Layers className="w-4 h-4 text-indigo-400 shrink-0" />
                        <span className="truncate">{t("learn.domainOverviewLabel")}</span>
                      </span>
                      <span className="text-xs font-mono text-graphite-400 shrink-0">
                        {t("learn.roadmapBadge")}
                      </span>
                    </button>

                    {(domain.subtopics || []).length > 0 && (
                      <>
                        <button
                          type="button"
                          onClick={() => toggleDeepDiveExpand(domain.domain_id)}
                          aria-expanded={Boolean(expandedDeepDiveDomains[domain.domain_id])}
                          className="w-full text-left px-2.5 py-1.5 rounded text-xs font-mono text-graphite-400 hover:text-graphite-200 hover:bg-graphite-800/40 flex items-center justify-between gap-2 transition-colors cursor-pointer"
                        >
                          <span className="flex items-center gap-1.5 min-w-0">
                            <ChevronDown
                              className={cn(
                                "w-3.5 h-3.5 shrink-0 transition-transform duration-150",
                                !expandedDeepDiveDomains[domain.domain_id] && "-rotate-90"
                              )}
                            />
                            <span className="truncate">
                              {t("learn.optionalDeepDiveLabel", { count: domain.subtopics.length })}
                            </span>
                          </span>
                          <span className="shrink-0">
                            {domain.completed_subtopics_in_domain}/{domain.total_subtopics_in_domain}
                          </span>
                        </button>

                        {expandedDeepDiveDomains[domain.domain_id] && (
                          <div className="pl-2 space-y-1">
                            {(domain.subtopics || []).map((sub) => {
                              const isSubActive = effectiveTopicId === sub.sub_topic_id;
                              return (
                                <button
                                  key={sub.sub_topic_id}
                                  type="button"
                                  onClick={() => handleSelectSubtopic(sub.sub_topic_id, domain.domain_id)}
                                  className={cn(
                                    "w-full text-left px-2.5 py-2 rounded text-sm flex items-center justify-between gap-2 transition-colors cursor-pointer",
                                    isSubActive
                                      ? "bg-indigo-500/20 border border-indigo-500/40 text-white font-semibold"
                                      : "text-graphite-200 hover:bg-graphite-800/60 hover:text-white border border-transparent"
                                  )}
                                >
                                  <span className="flex items-center gap-2 min-w-0">
                                    {sub.is_completed ? (
                                      <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                                    ) : (
                                      <FileText className="w-4 h-4 text-graphite-400 shrink-0" />
                                    )}
                                    <span className="font-mono text-xs text-indigo-300 shrink-0">
                                      {sub.sub_topic_id}
                                    </span>
                                    <span className="truncate text-sm">{sub.title}</span>
                                  </span>
                                  <span
                                    className={cn(
                                      "text-xs font-mono shrink-0 px-1.5 py-0.5 rounded",
                                      sub.sample_attempted_count >= sub.sample_total_count
                                        ? "text-emerald-400 bg-emerald-500/10"
                                        : "text-graphite-400 bg-graphite-900"
                                    )}
                                  >
                                    {sub.sample_attempted_count}/{sub.sample_total_count}
                                  </span>
                                </button>
                              );
                            })}
                          </div>
                        )}
                      </>
                    )}
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
                  <span className="px-2.5 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-xs font-mono text-graphite-200 uppercase">
                    {t("learn.overviewPill", { domainId: activeDomain.domain_id })}
                  </span>
                  <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-xs font-mono text-graphite-300">
                    <Clock className="w-3.5 h-3.5 text-indigo-400" />
                    {t("learn.estimatedReadTimeBadge")}
                  </span>
                  {activeDomain.recommended_focus && (
                    <span className="px-2.5 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-xs font-mono text-indigo-300 font-semibold">
                      {t("learn.recommendedFocusBadge")}
                    </span>
                  )}
                  {showProficiencyMetrics && (
                    <span className="text-xs font-mono text-graphite-300">
                      {t("learn.currentProficiency")}{" "}
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
                      className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-indigo-500/10 hover:bg-indigo-500/20 border border-indigo-500/30 text-xs font-mono text-indigo-300 transition-colors"
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                      {t("learn.primaryReference")}
                    </a>
                  )}
                  {activeDomain.is_completed && (
                    <span className="inline-flex items-center gap-1.5 text-xs font-mono font-semibold text-emerald-400">
                      <CheckCircle2 className="w-4 h-4" />
                      {t("learn.domainAutoCompletedBadge")}
                    </span>
                  )}
                </div>
              </div>

              <div className="p-6 sm:p-8 space-y-6">
                <MarkdownView content={activeDomain.content} />

                {actionErrorBanner}

                {unlockNoticeBanner}

                {/* Direct Domain Essential Completion Checkpoint */}
                <div className="p-4 rounded-lg bg-graphite-950/70 border border-graphite-800 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div className="space-y-1">
                    <p className="text-sm font-semibold text-white">
                      {activeDomain.is_completed
                        ? t("learn.domainEssentialDoneTitle")
                        : activeDomainDisplayName}
                    </p>
                    <p className="text-sm text-graphite-300">
                      {t("learn.domainEssentialCompletedHint")}
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2.5 shrink-0">
                    {!activeDomain.is_completed ? (
                      <button
                        type="button"
                        disabled={completingDomainId === activeDomain.domain_id}
                        onClick={() => handleCompleteDomain(activeDomain.domain_id)}
                        className="px-4 py-2 rounded-md bg-emerald-500 hover:bg-emerald-400 text-black font-semibold text-sm transition-all flex items-center justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
                      >
                        {completingDomainId === activeDomain.domain_id ? (
                          <>
                            <Loader2 className="w-4 h-4 animate-spin" />
                            {t("learn.saving")}
                          </>
                        ) : (
                          <>
                            <CheckCircle2 className="w-4 h-4" />
                            {t("learn.markDomainCompleted")}
                          </>
                        )}
                      </button>
                    ) : data.theory_completed ? (
                      <button
                        type="button"
                        onClick={() => navigate("/simulation")}
                        className="px-4 py-2 rounded-md bg-white hover:bg-white/90 text-black font-semibold text-sm transition-all flex items-center justify-center gap-2 cursor-pointer"
                      >
                        <PlayCircle className="w-4 h-4" />
                        {t("learn.enterSimulation")}
                      </button>
                    ) : nextActionDomain ? (
                      <button
                        type="button"
                        onClick={() => handleSelectOverview(nextActionDomain.domain_id)}
                        className="px-4 py-2 rounded-md bg-white hover:bg-white/90 text-black font-semibold text-sm transition-all flex items-center justify-center gap-2 cursor-pointer"
                      >
                        {t("learn.nextOverview")}
                        <ChevronRight className="w-4 h-4" />
                      </button>
                    ) : null}
                  </div>
                </div>

                {/* Collapsible Optional Deep Dive (Sub-topic Checklist Roadmap Grid) */}
                <div className="pt-4 border-t border-graphite-800 space-y-4">
                  <button
                    type="button"
                    onClick={() => setOverviewDeepDiveOpen((prev) => !prev)}
                    aria-expanded={overviewDeepDiveOpen}
                    className="w-full p-4 rounded-lg bg-graphite-950/60 border border-graphite-800 hover:border-graphite-700 text-left flex items-center justify-between gap-3 transition-colors cursor-pointer"
                  >
                    <div>
                      <h3 className="text-base font-bold text-white tracking-tight">
                        {t("learn.optionalDeepDiveTitle", {
                          completed: activeDomain.completed_subtopics_in_domain,
                          total: activeDomain.total_subtopics_in_domain,
                        })}
                      </h3>
                      <p className="text-sm text-graphite-300 mt-0.5">
                        {t("learn.subtopicsSectionDesc", { domainName: activeDomainDisplayName })}
                      </p>
                    </div>
                    <ChevronDown
                      className={cn(
                        "w-5 h-5 text-graphite-400 shrink-0 transition-transform duration-150",
                        !overviewDeepDiveOpen && "-rotate-90"
                      )}
                    />
                  </button>

                  {overviewDeepDiveOpen && (
                    <div className="space-y-4">
                      {activeDomain.subtopics?.[0] && (
                        <div className="flex justify-end">
                          <button
                            type="button"
                            onClick={() =>
                              handleSelectSubtopic(
                                activeDomain.subtopics[0].sub_topic_id,
                                activeDomain.domain_id
                              )
                            }
                            className="px-3.5 py-2 rounded-md bg-white text-black font-semibold text-sm hover:bg-white/90 transition-colors flex items-center gap-1.5 cursor-pointer"
                          >
                            {t("learn.startSubtopicButton", { subTopicId: activeDomain.subtopics[0].sub_topic_id })}
                            <ChevronRight className="w-4 h-4" />
                          </button>
                        </div>
                      )}

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
                              <div className="space-y-1.5">
                                <span className="inline-block px-2 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-xs font-mono text-indigo-300 font-semibold">
                                  {t("learn.objectiveLabel", { subTopicId: sub.sub_topic_id })}
                                </span>
                                <h4 className="text-sm font-semibold text-white leading-snug">
                                  {sub.title}
                                </h4>
                              </div>
                              {sub.is_completed ? (
                                <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                              ) : (
                                <Circle className="w-4 h-4 text-graphite-600 shrink-0" />
                              )}
                            </div>

                            <div className="flex items-center justify-between text-xs font-mono pt-1.5 border-t border-graphite-800/60">
                              <span
                                className={
                                  sub.sample_attempted_count >= sub.sample_total_count
                                    ? "text-emerald-400"
                                    : "text-graphite-400"
                                }
                              >
                                {t("learn.questionsRatio", {
                                  attempted: sub.sample_attempted_count,
                                  total: sub.sample_total_count,
                                })}
                              </span>
                              <span className="text-indigo-300 flex items-center gap-1">
                                {sub.is_completed ? t("learn.reviewAction") : t("learn.studyAction")}
                                <ChevronRight className="w-3.5 h-3.5" />
                              </span>
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>

              <div className="px-6 py-4 border-t border-graphite-800 bg-graphite-950/60 flex items-center justify-between gap-3">
                <button
                  type="button"
                  disabled={!prevDomain}
                  onClick={() => prevDomain && handleSelectOverview(prevDomain.domain_id)}
                  className="px-3.5 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-sm font-mono text-graphite-200 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                >
                  <ChevronLeft className="w-4 h-4" />
                  {t("learn.previousOverview")}
                </button>
                <button
                  type="button"
                  disabled={!nextDomain}
                  onClick={() => nextDomain && handleSelectOverview(nextDomain.domain_id)}
                  className="px-3.5 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-sm font-mono text-graphite-200 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                >
                  {t("learn.nextOverview")}
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </>
          ) : subLoading ? (
            /* ---------------- SUB-TOPIC LOADING ---------------- */
            <div className="min-h-[420px] flex flex-col items-center justify-center gap-3 p-8">
              <Loader2 className="w-6 h-6 text-white animate-spin" />
              <p className="text-sm font-mono text-graphite-400">
                {t("learn.loadingSubtopic", { subTopicId: effectiveTopicId })}
              </p>
            </div>
          ) : subError || !subDetail ? (
            /* ---------------- SUB-TOPIC ERROR ---------------- */
            <div className="p-8 text-center space-y-4">
              <AlertTriangle className="w-8 h-8 text-amber-400 mx-auto" />
              <p className="text-sm font-mono text-graphite-300">
                {subError || t("learn.subtopicDetailsUnavailable")}
              </p>
              <button
                type="button"
                onClick={() => handleSelectOverview(activeDomain.domain_id)}
                className="px-4 py-2 rounded-md bg-graphite-800 text-sm font-semibold text-white hover:bg-graphite-700 cursor-pointer"
              >
                {t("learn.backToOverview")}
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
                    className="px-2.5 py-0.5 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-xs font-mono text-graphite-200 uppercase cursor-pointer"
                  >
                    {t("learn.overviewBreadcrumb", { domainId: subDetail.domain_id })}
                  </button>
                  <span className="text-graphite-500 text-xs">/</span>
                  <span className="px-2.5 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-xs font-mono text-indigo-300 font-semibold">
                    {t("learn.subtopicBadge", { subTopicId: subDetail.sub_topic_id })}
                  </span>
                  {subDetail.preview_mode && (
                    <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded bg-amber-500/15 border border-amber-500/30 text-xs font-mono text-amber-300 font-semibold">
                      <Eye className="w-3.5 h-3.5" />
                      {t("learn.previewModeBadge")}
                    </span>
                  )}
                </div>

                <div className="flex flex-wrap items-center gap-3">
                  {/^https?:\/\//i.test(subDetail.source_url || "") && (
                    <a
                      href={subDetail.source_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-indigo-500/10 hover:bg-indigo-500/20 border border-indigo-500/30 text-xs font-mono text-indigo-300 transition-colors"
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                      {t("learn.webReference")}
                    </a>
                  )}
                  {subDetail.is_completed && (
                    <span className="inline-flex items-center gap-1.5 text-xs font-mono font-semibold text-emerald-400">
                      <CheckCircle2 className="w-4 h-4" />
                      {t("learn.subtopicCompletedBadge")}
                    </span>
                  )}
                </div>
              </div>

              {/* Markdown Lesson Body */}
              <div className="p-6 sm:p-8 space-y-8">
                <MarkdownView content={subDetail.content} />

                {/* Interactive Formative Sample Questions (Strictly English-only content) */}
                <div className="pt-6 border-t border-graphite-800 space-y-4">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="space-y-1">
                      <h3 className="text-lg font-bold text-white flex items-center gap-2">
                        <HelpCircle className="w-5 h-5 text-indigo-400" />
                        {t("learn.sampleSectionTitle")}
                      </h3>
                      <p className="text-sm text-graphite-300">
                        {subDetail.preview_mode
                          ? t("learn.samplePreviewDesc")
                          : t("learn.sampleLearnerDesc")}
                      </p>
                    </div>
                    <span
                      className={cn(
                        "px-3 py-1 rounded text-xs font-mono font-semibold border",
                        subDetail.ready_to_complete
                          ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                          : "bg-graphite-950 text-graphite-200 border-graphite-800"
                      )}
                    >
                      {t("learn.attemptedRatio", {
                        attempted: subDetail.attempted_count,
                        required: subDetail.required_count,
                      })}
                    </span>
                  </div>

                  {locale === "th" && (
                    <div className="px-3.5 py-2.5 rounded bg-graphite-900/70 border border-graphite-800 text-graphite-300 text-xs">
                      {t("learn.sampleEnglishOnlyNotice")}
                    </div>
                  )}

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
                              {t("learn.sampleQuestionNumber", { index: qIdx + 1 })}
                            </span>
                            <div className="flex items-center gap-2">
                              {q.is_fallback ? (
                                <span className="px-2 py-0.5 rounded bg-amber-500/15 border border-amber-500/30 text-xs font-mono text-amber-300">
                                  {t("learn.domainFallbackBadge")}
                                </span>
                              ) : (
                                <span className="px-2 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-xs font-mono text-indigo-300">
                                  {t("learn.objectiveLabel", { subTopicId: subDetail.sub_topic_id })}
                                </span>
                              )}

                              {isSubmitted && (
                                <span
                                  className={cn(
                                    "inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-mono font-semibold border",
                                    q.is_correct
                                      ? "bg-emerald-500/15 text-emerald-300 border-emerald-500/30"
                                      : "bg-red-500/15 text-red-300 border-red-500/30"
                                  )}
                                >
                                  {q.is_correct ? (
                                    <>
                                      <CheckCircle2 className="w-3.5 h-3.5" />
                                      {t("learn.correctLockedBadge", { answer: q.submitted_answer ?? "" })}
                                    </>
                                  ) : (
                                    <>
                                      <XCircle className="w-3.5 h-3.5" />
                                      {t("learn.wrongLockedBadge", {
                                        submitted: q.submitted_answer ?? "",
                                        correct: q.correct_answer ?? "",
                                      })}
                                    </>
                                  )}
                                </span>
                              )}
                            </div>
                          </div>

                          <p className="text-base text-white leading-relaxed">{q.question_text}</p>

                          <div className="space-y-2.5">
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
                                    "w-full text-left px-4 py-3 rounded-md border text-base transition-all flex items-start gap-3",
                                    isSubmitted ? "cursor-default" : "cursor-pointer",
                                    isActualCorrect
                                      ? "bg-emerald-500/15 border-emerald-500/50 text-emerald-100"
                                      : isWrongPick
                                      ? "bg-red-500/15 border-red-500/50 text-red-100"
                                      : isChosen
                                      ? "bg-indigo-500/20 border-indigo-400 text-white"
                                      : "bg-graphite-900 border-graphite-800 text-graphite-200 hover:border-graphite-700"
                                  )}
                                >
                                  <span className="font-mono font-bold shrink-0">
                                    {optKey}.
                                  </span>
                                  <span className="leading-relaxed">{optText}</span>
                                </button>
                              );
                            })}
                          </div>

                          {!isSubmitted ? (
                            <div className="flex items-center justify-between pt-1">
                              <span className="text-xs font-mono text-graphite-400">
                                {draftAnswers[q.question_id]
                                  ? t("learn.selectedOptionStatus", { answer: draftAnswers[q.question_id] })
                                  : t("learn.selectedOptionHint")}
                              </span>
                              <button
                                type="button"
                                disabled={
                                  !draftAnswers[q.question_id] ||
                                  submittingQuestionId === q.question_id
                                }
                                onClick={() => handleSubmitSampleAnswer(q.question_id)}
                                className="px-4 py-2 rounded-md bg-white text-black font-semibold text-sm hover:bg-white/90 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                              >
                                {submittingQuestionId === q.question_id ? (
                                  <>
                                    <Loader2 className="w-4 h-4 animate-spin" />
                                    {t("learn.checkingAnswer")}
                                  </>
                                ) : (
                                  t("learn.checkLockAnswer")
                                )}
                              </button>
                            </div>
                          ) : (
                            q.explanation && (
                              <div className="p-4 rounded-md bg-graphite-900 border border-graphite-800 space-y-1.5">
                                <span className="text-xs font-mono uppercase tracking-wider text-indigo-300 font-semibold">
                                  {t("learn.explanationLabel")}
                                </span>
                                <p className="text-base text-graphite-200 leading-relaxed">
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
                {actionErrorBanner}

                {unlockNoticeBanner}

                <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      disabled={!subDetail.previous_sub_topic_id}
                      onClick={() =>
                        subDetail.previous_sub_topic_id &&
                        handleSelectSubtopic(subDetail.previous_sub_topic_id)
                      }
                      className="px-3.5 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-sm font-mono text-graphite-200 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                    >
                      <ChevronLeft className="w-4 h-4" />
                      {t("learn.previousSubtopic")}
                    </button>
                    <button
                      type="button"
                      disabled={!subDetail.next_sub_topic_id}
                      onClick={() =>
                        subDetail.next_sub_topic_id &&
                        handleSelectSubtopic(subDetail.next_sub_topic_id)
                      }
                      className="px-3.5 py-2 rounded-md border border-graphite-800 bg-graphite-900 text-sm font-mono text-graphite-200 hover:text-white hover:border-graphite-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 cursor-pointer"
                    >
                      {t("learn.nextSubtopic")}
                      <ChevronRight className="w-4 h-4" />
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
                            ? t("learn.markCompletedTitle")
                            : t("learn.attemptFirstTitle")
                        }
                        className="px-4 py-2 rounded-md bg-emerald-500 hover:bg-emerald-400 text-black font-semibold text-sm transition-all flex items-center justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
                      >
                        {completingSubtopicId === subDetail.sub_topic_id ? (
                          <>
                            <Loader2 className="w-4 h-4 animate-spin" />
                            {t("learn.saving")}
                          </>
                        ) : (
                          <>
                            <CheckCircle2 className="w-4 h-4" />
                            {subDetail.ready_to_complete
                              ? t("learn.markSubtopicCompleted")
                              : t("learn.attemptMoreToUnlock", {
                                  remaining: Math.max(
                                    0,
                                    subDetail.required_count - subDetail.attempted_count
                                  ),
                                })}
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
                        className="px-4 py-2 rounded-md bg-white hover:bg-white/90 text-black font-semibold text-sm transition-all flex items-center justify-center gap-2 cursor-pointer"
                      >
                        {t("learn.completedNextSubtopic", { subTopicId: subDetail.next_sub_topic_id })}
                        <ChevronRight className="w-4 h-4" />
                      </button>
                    ) : data.theory_completed ? (
                      <button
                        type="button"
                        onClick={() => navigate("/simulation")}
                        className="px-4 py-2 rounded-md bg-white hover:bg-white/90 text-black font-semibold text-sm transition-all flex items-center justify-center gap-2 cursor-pointer"
                      >
                        <PlayCircle className="w-4 h-4" />
                        {t("learn.startAdaptiveSimulation")}
                      </button>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-md bg-emerald-500/10 border border-emerald-500/30 text-sm font-mono font-semibold text-emerald-400">
                        <CheckCircle2 className="w-4 h-4" />
                        {t("learn.subtopicCompletedBadge")}
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
