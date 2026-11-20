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
import { useI18n } from "../i18n";
import { resolveScenarioLocale } from "../i18n/format";
import type { ScenarioTranslationPayload } from "../i18n/types";

interface SubmissionSummary {
  session_id: string;
  learner_nickname: string;
  scenario_id: string;
  scenario_title: string;
  scenario_translations?: {
    th?: ScenarioTranslationPayload;
    [locale: string]: ScenarioTranslationPayload | undefined;
  };
  domain_id: string;
  domain_name: string;
  final_score: number;
  tp_fp_correct: boolean;
  findings_found_count: number;
  total_findings_count: number;
  evidence_selected_count: number;
  noise_selected_count: number;
  evidence_penalty: number;
  actions_count: number;
  completed_at: string;
}

interface EvidenceLog {
  index: number;
  timestamp: string;
  source: string;
  severity: string;
  message: string;
}

interface EvidenceFinding {
  id: string;
  description: string;
  domain_id: string;
  found: boolean;
  points: number;
  earned: number;
  evidence_log_indices: number[];
  matched_log_indices: number[];
  evidence_logs: EvidenceLog[];
  explanation: string;
}

interface TimelineAction {
  action_id: number;
  step_order: number;
  action_type: string;
  payload: unknown;
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
  scenario_translations?: {
    th?: ScenarioTranslationPayload;
    [locale: string]: ScenarioTranslationPayload | undefined;
  };
  domain_id: string;
  is_true_positive: boolean;
  tp_fp_explanation: string;
  learner_choice: boolean | null;
  tp_fp_correct: boolean;
  tp_fp_points: number;
  discovered_findings: string[];
  expected_findings: unknown;
  selected_actions: string[];
  expected_actions: unknown;
  final_score: number;
  selected_log_indices: number[] | null;
  evidence_findings: EvidenceFinding[] | null;
  noise_selections: EvidenceLog[] | null;
  gross_findings_points: number;
  findings_points: number;
  evidence_penalty: number;
  evidence_review_available: boolean;
  timeline: TimelineAction[];
  completed_at: string;
}

interface SelectedLogRow {
  log: EvidenceLog;
  findingIds: string[];
  isNoise: boolean;
}

function buildSelectedLogRows(detail: SubmissionDetail, logUnavailableText: string): SelectedLogRow[] {
  const snippets = new Map<number, EvidenceLog>();
  const matchedFindings = new Map<number, string[]>();

  (detail.evidence_findings || []).forEach((finding) => {
    finding.evidence_logs.forEach((log) => snippets.set(log.index, log));
    finding.matched_log_indices.forEach((index) => {
      const ids = matchedFindings.get(index) || [];
      ids.push(finding.id);
      matchedFindings.set(index, ids);
    });
  });
  (detail.noise_selections || []).forEach((log) => snippets.set(log.index, log));

  return (detail.selected_log_indices || [])
    .map((index) => {
      const findingIds = matchedFindings.get(index) || [];
      const log = snippets.get(index) || {
        index,
        timestamp: "",
        source: "",
        severity: "",
        message: logUnavailableText,
      };
      return { log, findingIds, isNoise: findingIds.length === 0 };
    })
    .sort((a, b) => a.log.index - b.log.index);
}

function describeEvidenceAction(
  payload: unknown,
  t: (key: string, params?: Record<string, string | number>) => string,
): string | null {
  if (!payload || typeof payload !== "object") return null;
  const data = payload as Record<string, unknown>;
  if (data.operation === "clear") return t("specialist.modal.clearEvidenceOp");
  if (data.operation === "toggle") {
    return data.selected
      ? t("specialist.modal.selectedLogOp", { index: String(data.log_index) })
      : t("specialist.modal.deselectedLogOp", { index: String(data.log_index) });
  }
  return null;
}

export default function SpecialistReviewPage() {
  const { locale, t, formatDate } = useI18n();
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
  const [modalCanonicalMode, setModalCanonicalMode] = useState(false);

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
    setModalCanonicalMode(false);
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
    setModalCanonicalMode(false);
  };

  // Filter by search query (nickname or scenario title in either EN or localized locale)
  const filteredSubmissions = submissions.filter((s) => {
    const q = searchQuery.toLowerCase();
    const localizedTitle = resolveScenarioLocale(
      { title: s.scenario_title, translations: s.scenario_translations },
      locale,
    ).title.toLowerCase();
    return (
      s.learner_nickname.toLowerCase().includes(q) ||
      s.scenario_title.toLowerCase().includes(q) ||
      localizedTitle.includes(q) ||
      s.domain_id.toLowerCase().includes(q)
    );
  });

  if (userRole !== "specialist" && userRole !== "admin") {
    return (
      <div className="max-w-xl mx-auto text-center py-16 space-y-4">
        <div className="w-12 h-12 rounded-full bg-amber-500/10 text-amber-400 flex items-center justify-center mx-auto border border-amber-500/20">
          <ShieldAlert className="w-6 h-6" />
        </div>
        <h2 className="text-xl font-bold text-white">{t("specialist.accessTitle")}</h2>
        <p className="text-graphite-400 text-sm leading-relaxed">
          {t("specialist.accessDesc")}
        </p>
        <button
          onClick={() => navigate("/profile")}
          className="mt-4 px-4 py-2 rounded-md bg-graphite-100 hover:bg-white text-graphite-950 font-semibold text-xs transition-colors cursor-pointer"
        >
          {t("specialist.goToProfile")}
        </button>
      </div>
    );
  }

  const effectiveModalLocale = modalCanonicalMode ? "en" : locale;
  const resolvedDetail = detail
    ? resolveScenarioLocale(
        {
          title: detail.scenario_title,
          description: detail.scenario_description,
          tp_fp_explanation: detail.tp_fp_explanation,
          translations: detail.scenario_translations,
        },
        effectiveModalLocale,
      )
    : null;

  const resolveActionLabel = (act: string): string => {
    if (!detail?.expected_actions || typeof detail.expected_actions !== "object" || !resolvedDetail) {
      return act;
    }
    const phases = detail.expected_actions as Record<string, Array<{ id?: string; action_id?: string; label?: string }>>;
    for (const items of Object.values(phases)) {
      if (!Array.isArray(items)) continue;
      for (const item of items) {
        const actionId = item.id || item.action_id || "";
        if (actionId && (actionId === act || item.label === act)) {
          return resolvedDetail.playbookLabel(actionId, item.label || act);
        }
      }
    }
    return act;
  };

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-graphite-800 pb-6">
        <div>
          <div className="flex items-center gap-2.5 mb-1">
            <h1 className="text-2xl font-bold tracking-tight text-white">{t("specialist.title")}</h1>
            <span className="px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              {t("specialist.privacyBadge")}
            </span>
          </div>
          <p className="text-xs text-graphite-400">
            {t("specialist.subtitle")}
          </p>
        </div>

        <button
          onClick={() => navigate("/discussions/new")}
          className="flex items-center gap-2 px-4 py-2 rounded-md bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm transition-all cursor-pointer self-start md:self-auto"
        >
          <FileText className="w-4 h-4" />
          <span>{t("specialist.writePost")}</span>
        </button>
      </div>

      {/* Filters Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 p-4 rounded-lg bg-graphite-900/50 border border-graphite-800 text-xs">
        {/* Search */}
        <div className="relative">
          <Search className="w-3.5 h-3.5 text-graphite-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder={t("specialist.searchPlaceholder")}
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
            <option value="">{t("specialist.allDomains")}</option>
            <option value="domain1">{t("specialist.domainOption1")}</option>
            <option value="domain2">{t("specialist.domainOption2")}</option>
            <option value="domain3">{t("specialist.domainOption3")}</option>
            <option value="domain4">{t("specialist.domainOption4")}</option>
            <option value="domain5">{t("specialist.domainOption5")}</option>
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
            <option value="">{t("specialist.allScores")}</option>
            <option value="50">{t("specialist.scoreLe50")}</option>
            <option value="70">{t("specialist.scoreLe70")}</option>
            <option value="85">{t("specialist.scoreLe85")}</option>
          </select>
        </div>
      </div>

      {/* Submissions Table / Feed */}
      {loading ? (
        <div className="py-20 text-center text-graphite-500 text-xs font-mono">
          {t("specialist.loading")}
        </div>
      ) : filteredSubmissions.length === 0 ? (
        <div className="py-16 text-center border border-dashed border-graphite-800 rounded-lg space-y-2">
          <HelpCircle className="w-8 h-8 text-graphite-600 mx-auto" />
          <p className="text-sm font-semibold text-graphite-300">{t("specialist.emptyTitle")}</p>
          <p className="text-xs text-graphite-500">{t("specialist.emptyDesc")}</p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-graphite-800 bg-graphite-950">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-graphite-900 border-b border-graphite-800 text-graphite-400 font-mono">
                <th className="p-3 font-semibold">{t("specialist.table.learner")}</th>
                <th className="p-3 font-semibold">{t("specialist.table.scenarioDomain")}</th>
                <th className="p-3 font-semibold">{t("specialist.table.score")}</th>
                <th className="p-3 font-semibold">{t("specialist.table.triage")}</th>
                <th className="p-3 font-semibold">{t("specialist.table.findings")}</th>
                <th className="p-3 font-semibold">{t("specialist.table.evidence")}</th>
                <th className="p-3 font-semibold">{t("specialist.table.noise")}</th>
                <th className="p-3 font-semibold">{t("specialist.table.penalty")}</th>
                <th className="p-3 font-semibold">{t("specialist.table.completed")}</th>
                <th className="p-3 text-right font-semibold">{t("specialist.table.action")}</th>
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
                const subLocale = resolveScenarioLocale(
                  { title: sub.scenario_title, translations: sub.scenario_translations },
                  locale,
                );

                return (
                  <tr key={sub.session_id} className="hover:bg-graphite-900/40 transition-colors">
                    <td className="p-3">
                      <span className="font-mono font-medium text-graphite-200 bg-graphite-900 px-2 py-0.5 rounded border border-graphite-800 text-xs">
                        {sub.learner_nickname}
                      </span>
                    </td>
                    <td className="p-3">
                      <div className="font-medium text-white truncate max-w-xs">{subLocale.title}</div>
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
                          <CheckCircle2 className="w-3.5 h-3.5" /> {t("specialist.table.correct")}
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 text-[11px] font-mono text-red-400">
                          <XCircle className="w-3.5 h-3.5" /> {t("specialist.table.misjudged")}
                        </span>
                      )}
                    </td>
                    <td className="p-3 font-mono text-graphite-300">
                      {sub.findings_found_count}/{sub.total_findings_count}
                    </td>
                    <td className="p-3 font-mono text-cyan-300">
                      {sub.evidence_selected_count ?? 0}
                    </td>
                    <td className="p-3 font-mono text-amber-300">
                      {sub.noise_selected_count ?? 0}
                    </td>
                    <td className="p-3 font-mono text-red-300">
                      {sub.evidence_penalty ? `-${sub.evidence_penalty}` : "0"}
                    </td>
                    <td className="p-3 text-graphite-400 font-mono text-[11px]">
                      {sub.completed_at ? formatDate(sub.completed_at) : t("specialist.table.inProgress")}
                    </td>
                    <td className="p-3 text-right">
                      <button
                        onClick={() => openDetailModal(sub.session_id)}
                        className="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-graphite-900 border border-graphite-800 hover:border-graphite-600 text-graphite-300 hover:text-white font-medium text-xs transition-colors cursor-pointer"
                      >
                        <Eye className="w-3.5 h-3.5" />
                        <span>{t("specialist.table.inspectTrace")}</span>
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
            <div className="p-4 sm:p-5 border-b border-graphite-800 flex items-center justify-between gap-3 bg-graphite-900/70">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-xs font-mono px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 font-semibold">
                    {t("specialist.modal.traceBadge")}
                  </span>
                  {detail && (
                    <span className="text-xs font-mono text-graphite-400">
                      {t("specialist.modal.learnerPrefix")} <span className="text-white font-semibold">{detail.learner_nickname}</span>
                    </span>
                  )}
                </div>
                <h3 className="text-base font-bold text-white mt-1 truncate">
                  {resolvedDetail?.title || t("specialist.modal.loadingTitle")}
                </h3>
                {locale === "th" && detail && resolvedDetail && resolvedDetail.title !== detail.scenario_title && (
                  <p className="text-[11px] font-mono text-graphite-400 mt-0.5 truncate">
                    {t("specialist.modal.canonicalScenarioTitle", { title: detail.scenario_title })}
                  </p>
                )}
              </div>

              <div className="flex items-center gap-2 shrink-0">
                {locale === "th" && detail?.scenario_translations?.th && (
                  <div className="inline-flex rounded-md border border-graphite-800 bg-graphite-950 p-0.5 text-[10px] font-mono">
                    <button
                      type="button"
                      onClick={() => setModalCanonicalMode(false)}
                      className={cn(
                        "px-2 py-1 rounded transition-colors cursor-pointer",
                        !modalCanonicalMode ? "bg-indigo-600 text-white" : "text-graphite-400 hover:text-white"
                      )}
                    >
                      {t("specialist.modal.viewLocalized")}
                    </button>
                    <button
                      type="button"
                      onClick={() => setModalCanonicalMode(true)}
                      className={cn(
                        "px-2 py-1 rounded transition-colors cursor-pointer",
                        modalCanonicalMode ? "bg-indigo-600 text-white" : "text-graphite-400 hover:text-white"
                      )}
                    >
                      {t("specialist.modal.viewCanonical")}
                    </button>
                  </div>
                )}

                <button
                  onClick={closeDetailModal}
                  className="p-1.5 rounded-lg text-graphite-400 hover:text-white hover:bg-graphite-800 transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Modal Body */}
            <div className="flex-1 overflow-y-auto p-5 space-y-6 text-xs">
              {detailLoading || !detail || !resolvedDetail ? (
                <div className="py-20 text-center text-graphite-500 font-mono">
                  {t("specialist.modal.loadingBody")}
                </div>
              ) : (
                <>
                  {/* Score & Triage Summary Banner */}
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 p-4 rounded-lg bg-graphite-900/60 border border-graphite-800">
                    <div>
                      <div className="text-graphite-400 text-[10px] font-mono uppercase">{t("specialist.modal.finalScore")}</div>
                      <div className="text-xl font-bold font-mono text-white mt-0.5">
                        {detail.final_score.toFixed(1)}%
                      </div>
                    </div>

                    <div>
                      <div className="text-graphite-400 text-[10px] font-mono uppercase">{t("specialist.modal.triageAccuracy")}</div>
                      <div className="mt-0.5 flex items-center gap-1.5">
                        {detail.tp_fp_correct ? (
                          <span className="inline-flex items-center gap-1 text-emerald-400 font-semibold font-mono">
                            <CheckCircle2 className="w-3.5 h-3.5" />{" "}
                            {t("specialist.modal.correctTpFp", {
                              verdict: detail.learner_choice
                                ? t("specialist.modal.truePositiveFull")
                                : t("specialist.modal.falsePositiveFull"),
                            })}
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-red-400 font-semibold font-mono">
                            <XCircle className="w-3.5 h-3.5" />{" "}
                            {t("specialist.modal.misjudgedTpFp", {
                              verdict: detail.learner_choice
                                ? t("specialist.modal.truePositiveShort")
                                : t("specialist.modal.falsePositiveShort"),
                            })}
                          </span>
                        )}
                      </div>
                    </div>

                    <div>
                      <div className="text-graphite-400 text-[10px] font-mono uppercase">{t("specialist.modal.domain")}</div>
                      <div className="text-xs font-mono text-graphite-200 mt-0.5">
                        {detail.domain_id}
                      </div>
                    </div>
                  </div>

                  {/* Triage Explanation */}
                  <div className="space-y-1.5 p-3.5 rounded-lg bg-graphite-900/40 border border-graphite-800/80">
                    <div className="font-semibold text-graphite-200 flex items-center gap-1.5">
                      <HelpCircle className="w-3.5 h-3.5 text-indigo-400" />
                      <span>{t("specialist.modal.triageTruthTitle")}</span>
                    </div>
                    <p className="text-graphite-300 leading-relaxed text-xs">
                      {resolvedDetail.tp_fp_explanation}
                    </p>
                  </div>

                  {/* Legacy Findings Breakdown (hidden for evidence-review sessions) */}
                  {!detail.evidence_review_available && (
                    <div className="space-y-2">
                      <h4 className="font-semibold text-white flex items-center gap-1.5">
                        <Search className="w-3.5 h-3.5 text-amber-400" />
                        <span>{t("specialist.modal.legacyFindingsTitle")}</span>
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
                        <p className="text-graphite-500 italic">{t("specialist.modal.noFindingsDiscovered")}</p>
                      )}
                    </div>
                  )}

                  {/* Evidence Review (Specialist only) */}
                  {detail.evidence_review_available &&
                    (() => {
                      const selectedLogs = buildSelectedLogRows(detail, t("specialist.modal.logUnavailable"));
                      return (
                        <div className="space-y-3 rounded-lg border border-cyan-500/20 bg-cyan-950/10 p-4">
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <h4 className="font-semibold text-white flex items-center gap-1.5">
                              <Search className="w-3.5 h-3.5 text-cyan-400" />
                              <span>{t("specialist.modal.evidenceReviewTitle")}</span>
                            </h4>
                            <span className="px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
                              {t("specialist.modal.specialistOnlyBadge")}
                            </span>
                          </div>

                          <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 font-mono">
                            <div className="p-2.5 rounded bg-graphite-950/70 border border-graphite-800">
                              <div className="text-[9px] uppercase text-graphite-500">{t("specialist.modal.statSelected")}</div>
                              <div className="text-sm font-bold text-cyan-300">
                                {detail.selected_log_indices?.length ?? 0}
                              </div>
                            </div>
                            <div className="p-2.5 rounded bg-graphite-950/70 border border-graphite-800">
                              <div className="text-[9px] uppercase text-graphite-500">{t("specialist.modal.statFindings")}</div>
                              <div className="text-sm font-bold text-white">
                                {(detail.evidence_findings || []).filter((f) => f.found).length}/
                                {(detail.evidence_findings || []).length}
                              </div>
                            </div>
                            <div className="p-2.5 rounded bg-graphite-950/70 border border-graphite-800">
                              <div className="text-[9px] uppercase text-graphite-500">{t("specialist.modal.statNoise")}</div>
                              <div className="text-sm font-bold text-amber-300">
                                {detail.noise_selections?.length ?? 0}
                              </div>
                            </div>
                            <div className="p-2.5 rounded bg-graphite-950/70 border border-graphite-800">
                              <div className="text-[9px] uppercase text-graphite-500">{t("specialist.modal.statGross")}</div>
                              <div className="text-sm font-bold text-emerald-300">
                                {detail.gross_findings_points}
                              </div>
                            </div>
                            <div className="p-2.5 rounded bg-graphite-950/70 border border-graphite-800">
                              <div className="text-[9px] uppercase text-graphite-500">{t("specialist.modal.statNetPenalty")}</div>
                              <div className="text-sm font-bold text-white">
                                {detail.findings_points}
                                <span className="text-red-300 text-xs"> / -{detail.evidence_penalty}</span>
                              </div>
                            </div>
                          </div>

                          {/* Selected logs */}
                          <div className="space-y-1.5">
                            <h5 className="text-[11px] font-semibold uppercase tracking-wide text-graphite-300">
                              {t("specialist.modal.selectedLogsTitle")}
                            </h5>
                            {selectedLogs.length === 0 ? (
                              <p className="text-graphite-500 italic">
                                {t("specialist.modal.noSelectedLogs")}
                              </p>
                            ) : (
                              <div className="space-y-1.5">
                                {selectedLogs.map((row) => (
                                  <div
                                    key={row.log.index}
                                    className="flex flex-wrap items-start justify-between gap-2 p-2.5 rounded bg-graphite-950/70 border border-graphite-800"
                                  >
                                    <div className="min-w-0 flex-1">
                                      <div className="flex flex-wrap items-center gap-2 font-mono text-[10px] text-graphite-400">
                                        <span className="text-cyan-300 font-semibold">
                                          #{row.log.index}
                                        </span>
                                        {row.log.timestamp && <span>{row.log.timestamp}</span>}
                                        {row.log.source && <span>{row.log.source}</span>}
                                        {row.log.severity && (
                                          <span className="uppercase text-amber-300">
                                            {row.log.severity}
                                          </span>
                                        )}
                                      </div>
                                      <div className="text-graphite-200 text-xs mt-0.5 break-words font-mono">
                                        {row.log.message}
                                      </div>
                                    </div>
                                    {row.isNoise ? (
                                      <span className="px-2 py-0.5 rounded bg-red-950/40 border border-red-500/20 text-red-300 text-[10px] font-mono font-semibold shrink-0">
                                        {t("specialist.modal.notEvidenceBadge")}
                                      </span>
                                    ) : (
                                      <span className="px-2 py-0.5 rounded bg-emerald-950/40 border border-emerald-500/20 text-emerald-300 text-[10px] font-mono font-semibold shrink-0">
                                        {t("specialist.modal.matchedPrefix", { ids: row.findingIds.join(", ") })}
                                      </span>
                                    )}
                                  </div>
                                ))}
                              </div>
                            )}
                          </div>

                          {/* Per-finding matched / missed */}
                          <div className="space-y-1.5">
                            <h5 className="text-[11px] font-semibold uppercase tracking-wide text-graphite-300">
                              {t("specialist.modal.keyFindingsTitle")}
                            </h5>
                            <div className="space-y-2">
                              {(detail.evidence_findings || []).map((finding) => {
                                const localizedDesc = resolvedDetail.findingDescription(finding.id, finding.description);
                                const localizedExpl = resolvedDetail.findingExplanation(finding.id, finding.explanation);
                                return (
                                  <div
                                    key={finding.id}
                                    className="rounded bg-graphite-950/70 border border-graphite-800 p-2.5 space-y-1.5"
                                  >
                                    <div className="flex flex-wrap items-center justify-between gap-2">
                                      <div className="flex items-center gap-1.5">
                                        {finding.found ? (
                                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                                        ) : (
                                          <XCircle className="w-3.5 h-3.5 text-red-400 shrink-0" />
                                        )}
                                        <span className="font-medium text-white text-xs">
                                          {localizedDesc}
                                        </span>
                                        <span className="font-mono text-[10px] text-graphite-500">
                                          {finding.id} · {finding.domain_id}
                                        </span>
                                      </div>
                                      <span
                                        className={cn(
                                          "font-mono text-[10px] font-semibold",
                                          finding.found ? "text-emerald-300" : "text-red-300"
                                        )}
                                      >
                                        {finding.found ? t("specialist.modal.foundBadge") : t("specialist.modal.missedBadge")} · {finding.earned}/
                                        {finding.points} {t("specialist.modal.ptsSuffix")}
                                      </span>
                                    </div>

                                    <div className="space-y-1">
                                      <div className="text-[10px] font-mono uppercase text-graphite-500">
                                        {finding.found
                                          ? t("specialist.modal.matchedEvidenceLabel")
                                          : t("specialist.modal.acceptableEvidenceLabel")}
                                      </div>
                                      {finding.evidence_logs.length === 0 ? (
                                        <p className="text-graphite-500 italic text-[11px]">
                                          {t("specialist.modal.noAuthoredEvidence")}
                                        </p>
                                      ) : (
                                        finding.evidence_logs.map((log) => {
                                          const matched = finding.matched_log_indices.includes(log.index);
                                          return (
                                            <div
                                              key={`${finding.id}-${log.index}`}
                                              className={cn(
                                                "p-2 rounded border font-mono text-[10px]",
                                                matched
                                                  ? "bg-emerald-950/30 border-emerald-500/25 text-emerald-100"
                                                  : "bg-graphite-900/60 border-graphite-800 text-graphite-300"
                                              )}
                                            >
                                              <span className="text-cyan-300">#{log.index}</span>{" "}
                                              <span className="text-graphite-500">
                                                [{log.source} · {log.severity}]
                                              </span>{" "}
                                              {log.message}
                                              {matched && (
                                                <span className="ml-1 text-emerald-400 font-semibold">
                                                  ✓ {t("specialist.modal.selectedTick")}
                                                </span>
                                              )}
                                            </div>
                                          );
                                        })
                                      )}
                                    </div>

                                    {localizedExpl && (
                                      <p className="text-graphite-400 text-[11px] leading-relaxed">
                                        {localizedExpl}
                                      </p>
                                    )}
                                  </div>
                                );
                              })}
                            </div>
                          </div>

                          {/* Noise picks */}
                          {(detail.noise_selections || []).length > 0 && (
                            <div className="space-y-1.5">
                              <h5 className="text-[11px] font-semibold uppercase tracking-wide text-red-300">
                                {t("specialist.modal.noisePicksTitle")}
                              </h5>
                              <div className="space-y-1">
                                {(detail.noise_selections || []).map((log) => (
                                  <div
                                    key={log.index}
                                    className="flex items-start justify-between gap-2 p-2 rounded bg-red-950/20 border border-red-500/20 font-mono text-[10px]"
                                  >
                                    <div className="min-w-0">
                                      <span className="text-cyan-300">#{log.index}</span>{" "}
                                      <span className="text-graphite-400">
                                        [{log.source} · {log.severity}]
                                      </span>{" "}
                                      <span className="text-graphite-200">{log.message}</span>
                                    </div>
                                    <span className="text-red-300 font-semibold shrink-0">-2 {t("specialist.modal.ptsSuffix")}</span>
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}

                          <div className="flex flex-wrap items-center justify-end gap-3 border-t border-graphite-800 pt-2 font-mono text-[11px]">
                            <span className="text-graphite-400">
                              {t("specialist.modal.summaryGross")} <span className="text-emerald-300">{detail.gross_findings_points} {t("specialist.modal.ptsSuffix")}</span>
                            </span>
                            <span className="text-graphite-400">
                              {t("specialist.modal.summaryPenalty")} <span className="text-red-300">-{detail.evidence_penalty} {t("specialist.modal.ptsSuffix")}</span>
                            </span>
                            <span className="text-graphite-400">
                              {t("specialist.modal.summaryNet")} <span className="text-white">{detail.findings_points} {t("specialist.modal.ptsSuffix")}</span>
                            </span>
                          </div>
                        </div>
                      );
                    })()}

                  {/* Actions Taken */}
                  <div className="space-y-2">
                    <h4 className="font-semibold text-white flex items-center gap-1.5">
                      <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                      <span>{t("specialist.modal.actionsTakenTitle")}</span>
                    </h4>
                    {detail.selected_actions && detail.selected_actions.length > 0 ? (
                      <div className="flex flex-wrap gap-1.5">
                        {detail.selected_actions.map((act, idx) => (
                          <span
                            key={idx}
                            className="px-2 py-1 rounded bg-graphite-900 border border-graphite-800 text-graphite-200 font-mono text-[11px]"
                          >
                            • {resolveActionLabel(act)}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <p className="text-graphite-500 italic">{t("specialist.modal.noActionsRecorded")}</p>
                    )}
                  </div>

                  {/* Chronological Timeline */}
                  <div className="space-y-2.5">
                    <h4 className="font-semibold text-white flex items-center gap-1.5">
                      <Activity className="w-3.5 h-3.5 text-indigo-400" />
                      <span>{t("specialist.modal.timelineTitle")}</span>
                    </h4>
                    <div className="border border-graphite-800 rounded-lg divide-y divide-graphite-800/60 overflow-hidden">
                      {detail.timeline && detail.timeline.length > 0 ? (
                        detail.timeline.map((item) => {
                          const evDesc =
                            item.action_type === "select_evidence"
                              ? describeEvidenceAction(item.payload, t)
                              : null;
                          return (
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
                                    {formatDate(item.timestamp, {
                                      year: "numeric",
                                      month: "short",
                                      day: "numeric",
                                      hour: "2-digit",
                                      minute: "2-digit",
                                      second: "2-digit",
                                    })}
                                  </div>
                                  {evDesc && (
                                    <div className="text-[10px] text-cyan-300 font-mono">
                                      {evDesc}
                                    </div>
                                  )}
                                </div>
                              </div>

                              <div className="flex items-center gap-2 font-mono">
                                {item.points > 0 ? (
                                  <span className="text-emerald-400 text-xs">+{item.points} {t("specialist.modal.ptsSuffix")}</span>
                                ) : item.points < 0 ? (
                                  <span className="text-red-400 text-xs">{item.points} {t("specialist.modal.ptsSuffix")}</span>
                                ) : (
                                  <span className="text-graphite-500 text-xs">0 {t("specialist.modal.ptsSuffix")}</span>
                                )}
                              </div>
                            </div>
                          );
                        })
                      ) : (
                        <div className="p-3 text-center text-graphite-500">{t("specialist.modal.noTimelineActions")}</div>
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
                {t("specialist.modal.close")}
              </button>

              {detail && (
                <button
                  onClick={() => {
                    closeDetailModal();
                    navigate(
                      `/discussions/new?scenario_id=${detail.scenario_id}&scenario_title=${encodeURIComponent(
                        detail.scenario_title
                      )}&session_id=${detail.session_id}`
                    );
                  }}
                  className="flex items-center gap-2 px-4 py-1.5 rounded-md bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs transition-all shadow-sm cursor-pointer"
                >
                  <BookOpen className="w-3.5 h-3.5" />
                  <span>{t("specialist.modal.authorAnalysis")}</span>
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
