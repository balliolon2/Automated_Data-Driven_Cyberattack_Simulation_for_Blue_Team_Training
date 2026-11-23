import { useState, useEffect } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import axios from "axios";
import {
  Save,
  Eye,
  Edit3,
  X,
  Crosshair,
  Clock,
  Search,
  Shield,
  Lightbulb,
  AlertCircle,
  ArrowLeft,
  Loader2,
} from "lucide-react";
import MarkdownView from "../components/MarkdownView";
import { useI18n } from "../i18n";
import { resolveApiError, resolveScenarioLocale } from "../i18n/format";
import type { ScenarioTranslationPayload } from "../i18n/types";

interface ScenarioOption {
  scenario_id: string;
  title: string;
  domain_id: string;
  translations?: {
    th?: ScenarioTranslationPayload;
    [locale: string]: ScenarioTranslationPayload | undefined;
  };
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

interface EvidenceReviewSubmission {
  session_id: string;
  learner_nickname: string;
  evidence_review_available: boolean;
  selected_log_indices: number[] | null;
  evidence_findings: EvidenceFinding[] | null;
  noise_selections: EvidenceLog[] | null;
  gross_findings_points: number;
  findings_points: number;
  evidence_penalty: number;
}

function markdownCell(value: string): string {
  return value.replace(/\|/g, "\\|").replace(/\r?\n/g, " ").trim();
}

function buildEvidenceReviewMarkdown(detail: EvidenceReviewSubmission): string {
  const findings = detail.evidence_findings || [];
  const noise = detail.noise_selections || [];
  const selectedCount = detail.selected_log_indices?.length ?? 0;
  const lines: string[] = [
    "### Evidence Review",
    "",
    `> Learner: **${markdownCell(detail.learner_nickname)}** · Selected logs: ${selectedCount} · Noise: ${noise.length} · Net findings: ${detail.findings_points} pts`,
    "",
    "| Finding | Result | Points | Matched evidence |",
    "|---|---|---|---|",
  ];

  findings.forEach((finding) => {
    const matched = finding.matched_log_indices.length
      ? finding.matched_log_indices.map((index) => `#${index}`).join(", ")
      : "—";
    lines.push(
      `| ${markdownCell(finding.description)} (${markdownCell(finding.id)}) | ${
        finding.found ? "Found" : "Missed"
      } | ${finding.earned}/${finding.points} | ${matched} |`
    );
  });

  const missed = findings.filter((finding) => !finding.found);
  if (missed.length > 0) {
    lines.push("", "**Acceptable evidence for missed findings**", "");
    missed.forEach((finding) => {
      finding.evidence_logs.forEach((log) => {
        lines.push(
          `- \`${finding.id}\` #${log.index} [${markdownCell(log.source)} · ${markdownCell(
            log.severity
          )}] ${markdownCell(log.message)}`
        );
      });
    });
  }

  if (noise.length > 0) {
    lines.push("", "**Noise picks (-2 each)**", "");
    noise.forEach((log) => {
      lines.push(
        `- #${log.index} [${markdownCell(log.source)} · ${markdownCell(log.severity)}] ${markdownCell(
          log.message
        )}`
      );
    });
  }

  lines.push(
    "",
    `**Gross credit:** ${detail.gross_findings_points} pts · **Evidence penalty:** -${detail.evidence_penalty} pts · **Net findings:** ${detail.findings_points} pts`
  );

  return lines.join("\n");
}

export default function ThreadEditorPage() {
  const { locale, t } = useI18n();
  const { id } = useParams<{ id: string }>();
  const isEditing = Boolean(id);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const token = localStorage.getItem("token");
  const userRole = localStorage.getItem("userRole") || "learner";
  const isSpecialistOrAdmin = userRole === "specialist" || userRole === "admin";

  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [scenarioId, setScenarioId] = useState<string>(searchParams.get("scenario_id") || "");
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState("");
  const [scenarios, setScenarios] = useState<ScenarioOption[]>([]);
  const [activeTab, setActiveTab] = useState<"write" | "preview" | "split">("write");

  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(isEditing);
  const [errorMsg, setErrorMsg] = useState("");
  const [evidenceReview, setEvidenceReview] = useState<EvidenceReviewSubmission | null>(null);
  const reviewSessionId = searchParams.get("session_id");

  const loadEvidenceReview = async (sessionId: string) => {
    if (!isSpecialistOrAdmin) return;
    try {
      const res = await axios.get(`/api/specialist/submissions/${sessionId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setEvidenceReview(res.data);
    } catch (err) {
      console.error("Failed to load evidence review:", err);
    }
  };

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    if (!isSpecialistOrAdmin) {
      navigate("/discussions");
      return;
    }

    loadScenarios();

    if (reviewSessionId) {
      loadEvidenceReview(reviewSessionId);
    }

    if (isEditing && id) {
      loadExistingThread(id);
    }
  }, [id]);

  const loadScenarios = async () => {
    try {
      // Load recent submissions to collect unique scenarios
      const res = await axios.get("/api/specialist/submissions", {
        headers: { Authorization: `Bearer ${token}` },
      });
      const map = new Map<string, ScenarioOption>();

      // Also ensure param scenario is preserved
      const paramScenarioId = searchParams.get("scenario_id");
      const paramScenarioTitle = searchParams.get("scenario_title");
      if (paramScenarioId && paramScenarioTitle) {
        map.set(paramScenarioId, {
          scenario_id: paramScenarioId,
          title: decodeURIComponent(paramScenarioTitle),
          domain_id: "Simulation",
        });
      }

      if (Array.isArray(res.data)) {
        for (const item of res.data) {
          if (item.scenario_id) {
            const existing = map.get(item.scenario_id);
            map.set(item.scenario_id, {
              scenario_id: item.scenario_id,
              title: item.scenario_title || existing?.title || "",
              domain_id: item.domain_id || existing?.domain_id || "Simulation",
              translations: item.scenario_translations || existing?.translations,
            });
          }
        }
      }
      setScenarios(Array.from(map.values()));
    } catch (err) {
      console.error("Failed to load scenario options:", err);
    }
  };

  const loadExistingThread = async (threadId: string) => {
    setInitialLoading(true);
    try {
      const res = await axios.get(`/api/threads/${threadId}`);
      const threadData = res.data;
      setTitle(threadData.title);
      setContent(threadData.content);
      if (threadData.scenario?.scenario_id) {
        setScenarioId(threadData.scenario.scenario_id);
        setScenarios((prev) => {
          const exists = prev.some((s) => s.scenario_id === threadData.scenario.scenario_id);
          if (exists) return prev;
          return [
            {
              scenario_id: threadData.scenario.scenario_id,
              title: threadData.scenario.title,
              domain_id: threadData.scenario.domain_id,
              translations: threadData.scenario.translations,
            },
            ...prev,
          ];
        });
      }
      if (Array.isArray(threadData.tags)) {
        setTags(threadData.tags);
      }
    } catch (err: unknown) {
      console.error("Failed to load thread:", err);
      setErrorMsg(resolveApiError(err, t, t("discussions.editor.loadFailed")));
    } finally {
      setInitialLoading(false);
    }
  };

  const handleAddTag = () => {
    const clean = tagInput.trim().toLowerCase().replace(/[^a-z0-9-_]/g, "");
    if (clean && !tags.includes(clean)) {
      setTags([...tags, clean]);
      setTagInput("");
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setTags(tags.filter((tag) => tag !== tagToRemove));
  };

  const insertSnippet = (snippet: string) => {
    setContent((prev) => (prev ? `${prev}\n\n${snippet}` : snippet));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg("");

    if (title.trim().length < 5) {
      setErrorMsg(t("discussions.editor.titleTooShort"));
      return;
    }
    if (content.trim().length < 20) {
      setErrorMsg(t("discussions.editor.contentTooShort"));
      return;
    }

    setLoading(true);
    try {
      if (isEditing) {
        await axios.put(
          `/api/threads/${id}`,
          {
            title: title.trim(),
            content: content.trim(),
            tags,
          },
          { headers: { Authorization: `Bearer ${token}` } }
        );
        navigate(`/discussions/${id}`);
      } else {
        const payload: { title: string; content: string; tags: string[]; scenario_id?: string } = {
          title: title.trim(),
          content: content.trim(),
          tags,
        };
        if (scenarioId) {
          payload.scenario_id = scenarioId;
        }
        const res = await axios.post("/api/threads", payload, {
          headers: { Authorization: `Bearer ${token}` },
        });
        navigate(`/discussions/${res.data.thread_id}`);
      }
    } catch (err: unknown) {
      console.error("Submit thread error:", err);
      setErrorMsg(resolveApiError(err, t, t("discussions.editor.submitFailed")));
    } finally {
      setLoading(false);
    }
  };

  if (initialLoading) {
    return (
      <div className="py-24 text-center text-graphite-400 font-mono text-sm">
        {t("discussions.editor.loading")}
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6 animate-fade-in pb-12">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-graphite-800 pb-4">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate(-1)}
            className="p-2 rounded-lg text-graphite-400 hover:text-white hover:bg-graphite-900 transition-colors cursor-pointer"
          >
            <ArrowLeft className="w-4 h-4" />
          </button>
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white">
              {isEditing ? t("discussions.editor.editTitle") : t("discussions.editor.newTitle")}
            </h1>
            <p className="text-sm text-graphite-300">
              {t("discussions.editor.subtitle")}
            </p>
          </div>
        </div>

        {/* View Switcher on larger screens */}
        <div className="flex items-center gap-1 p-1 rounded-lg bg-graphite-900 border border-graphite-800 text-sm">
          <button
            type="button"
            onClick={() => setActiveTab("write")}
            className={`px-3 py-1.5 rounded font-medium transition-colors cursor-pointer ${
              activeTab === "write" ? "bg-graphite-800 text-white" : "text-graphite-400 hover:text-white"
            }`}
          >
            <Edit3 className="w-4 h-4 inline mr-1" />
            {t("discussions.editor.writeTab")}
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("preview")}
            className={`px-3 py-1.5 rounded font-medium transition-colors cursor-pointer ${
              activeTab === "preview" ? "bg-graphite-800 text-white" : "text-graphite-400 hover:text-white"
            }`}
          >
            <Eye className="w-4 h-4 inline mr-1" />
            {t("discussions.editor.previewTab")}
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("split")}
            className={`hidden md:block px-3 py-1.5 rounded font-medium transition-colors cursor-pointer ${
              activeTab === "split" ? "bg-graphite-800 text-white" : "text-graphite-400 hover:text-white"
            }`}
          >
            {t("discussions.editor.splitTab")}
          </button>
        </div>
      </div>

      {errorMsg && (
        <div className="p-3.5 rounded-lg bg-red-950/30 border border-red-500/20 text-red-400 text-sm flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Scenario Link */}
        <div className="space-y-1.5">
          <label className="text-sm font-semibold text-graphite-200 flex items-center gap-1.5">
            <Crosshair className="w-4 h-4 text-indigo-400" />
            <span>{t("discussions.editor.linkScenarioLabel")}</span>
          </label>
          <select
            value={scenarioId}
            onChange={(e) => setScenarioId(e.target.value)}
            disabled={isEditing}
            className="w-full px-3.5 py-2.5 rounded-lg bg-graphite-950 border border-graphite-800 text-graphite-100 text-sm focus:outline-none focus:border-graphite-600 disabled:opacity-60"
          >
            <option value="">{t("discussions.editor.generalAnalysisOption")}</option>
            {scenarios.map((sc) => {
              const loc = resolveScenarioLocale(sc, locale);
              return (
                <option key={sc.scenario_id} value={sc.scenario_id}>
                  {loc.title} ({sc.domain_id})
                </option>
              );
            })}
          </select>
          <p className="text-xs text-graphite-400">
            {t("discussions.editor.linkScenarioHint")}
          </p>
        </div>

        {/* Title */}
        <div className="space-y-1.5">
          <label className="text-sm font-semibold text-graphite-200">
            {t("discussions.editor.postTitleLabel")} <span className="text-red-400">*</span>
          </label>
          <input
            type="text"
            required
            placeholder={t("discussions.editor.postTitlePlaceholder")}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full px-3.5 py-2.5 rounded-lg bg-graphite-950 border border-graphite-800 text-white placeholder-graphite-500 text-base focus:outline-none focus:border-graphite-600"
          />
        </div>

        {/* Tags */}
        <div className="space-y-1.5">
          <label className="text-sm font-semibold text-graphite-200">{t("discussions.editor.tagsLabel")}</label>
          <div className="flex flex-wrap items-center gap-1.5 p-2.5 rounded-lg bg-graphite-950 border border-graphite-800">
            {tags.map((tag) => (
              <span
                key={tag}
                className="px-2.5 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-graphite-200 font-mono text-xs flex items-center gap-1"
              >
                #{tag}
                <button
                  type="button"
                  onClick={() => handleRemoveTag(tag)}
                  className="text-graphite-400 hover:text-white cursor-pointer"
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              </span>
            ))}
            <input
              type="text"
              placeholder={t("discussions.editor.tagsPlaceholder")}
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleAddTag();
                }
              }}
              className="flex-1 min-w-[140px] bg-transparent text-sm text-white placeholder-graphite-500 focus:outline-none px-1"
            />
          </div>
        </div>

        {/* Quick-Insert Toolbar */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-semibold text-graphite-300">{t("discussions.editor.templatesLabel")}</span>
          </div>

          <div className="flex flex-wrap gap-2 text-xs">
            <button
              type="button"
              onClick={() =>
                insertSnippet(
                  `### Investigation Timeline\n| Step | Action Taken | Discovered Finding |\n|---|---|---|\n| 1 | Triage SIEM Alert | Suspicious PowerShell invocation |\n| 2 | Endpoint Inspection | Outbound C2 connection detected |\n| 3 | Host Isolation | VLAN quarantine applied |`
                )
              }
              className="px-3 py-2 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-200 hover:text-white flex items-center gap-1.5 text-xs sm:text-sm transition-colors cursor-pointer"
            >
              <Clock className="w-4 h-4 text-indigo-400" />
              <span>{t("discussions.editor.timelineBtn")}</span>
            </button>

            <button
              type="button"
              onClick={() =>
                insertSnippet(
                  `### Indicators of Compromise (IOCs)\n| Type | Value | Context |\n|---|---|---|\n| IPv4 | 198.51.100.45 | C2 Command Beacon |\n| Hash (SHA256) | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 | Malicious Dropper |\n| Domain | update.security-patch.net | Phishing Delivery |`
                )
              }
              className="px-3 py-2 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-200 hover:text-white flex items-center gap-1.5 text-xs sm:text-sm transition-colors cursor-pointer"
            >
              <Search className="w-4 h-4 text-amber-400" />
              <span>{t("discussions.editor.iocBtn")}</span>
            </button>

            <button
              type="button"
              onClick={() =>
                insertSnippet(
                  "### Recommended Detection Rule (Sigma)\n```yaml\ntitle: Suspicious Base64 Encoded PowerShell\nstatus: stable\nlogsource:\n  product: windows\n  service: security\ndetection:\n  selection:\n    EventID: 4688\n    CommandLine|contains:\n      - '-enc'\n      - '-encodedcommand'\n  condition: selection\nlevel: high\n```"
                )
              }
              className="px-3 py-2 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-200 hover:text-white flex items-center gap-1.5 text-xs sm:text-sm transition-colors cursor-pointer"
            >
              <Shield className="w-4 h-4 text-emerald-400" />
              <span>{t("discussions.editor.sigmaBtn")}</span>
            </button>

            <button
              type="button"
              onClick={() =>
                insertSnippet(
                  "> 💡 **Key Takeaways & Lessons Learned**:\n> Always trace parent process execution lineage before validating alert fidelity to prevent false negative dismissals."
                )
              }
              className="px-3 py-2 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-200 hover:text-white flex items-center gap-1.5 text-xs sm:text-sm transition-colors cursor-pointer"
            >
              <Lightbulb className="w-4 h-4 text-amber-400" />
              <span>{t("discussions.editor.takeawaysBtn")}</span>
            </button>

            {evidenceReview?.evidence_review_available && (
              <button
                type="button"
                onClick={() => insertSnippet(buildEvidenceReviewMarkdown(evidenceReview))}
                className="px-3 py-2 rounded bg-cyan-950/40 hover:bg-cyan-900/50 border border-cyan-500/30 text-cyan-200 hover:text-white flex items-center gap-1.5 text-xs sm:text-sm transition-colors cursor-pointer"
              >
                <Search className="w-4 h-4 text-cyan-400" />
                <span>{t("discussions.editor.evidenceReviewBtn")}</span>
              </button>
            )}
          </div>
        </div>

        {/* Content Area */}
        <div className="space-y-1.5">
          <label className="text-sm font-semibold text-graphite-200">
            {t("discussions.editor.postBodyLabel")} <span className="text-red-400">*</span>
          </label>

          {activeTab === "write" && (
            <textarea
              required
              rows={16}
              placeholder={t("discussions.editor.writePlaceholder")}
              value={content}
              onChange={(e) => setContent(e.target.value)}
              className="w-full p-4 rounded-lg bg-graphite-950 border border-graphite-800 text-white placeholder-graphite-500 text-sm font-mono leading-relaxed focus:outline-none focus:border-graphite-600 resize-y"
            />
          )}

          {activeTab === "preview" && (
            <div className="p-5 rounded-lg bg-graphite-950 border border-graphite-800 min-h-[350px]">
              {content.trim() ? (
                <MarkdownView content={content} />
              ) : (
                <p className="text-sm text-graphite-400 italic">{t("discussions.editor.previewEmpty")}</p>
              )}
            </div>
          )}

          {activeTab === "split" && (
            <div className="grid grid-cols-2 gap-4">
              <textarea
                required
                rows={16}
                placeholder={t("discussions.editor.splitPlaceholder")}
                value={content}
                onChange={(e) => setContent(e.target.value)}
                className="w-full p-4 rounded-lg bg-graphite-950 border border-graphite-800 text-white placeholder-graphite-500 text-sm font-mono leading-relaxed focus:outline-none focus:border-graphite-600 resize-y"
              />
              <div className="p-4 rounded-lg bg-graphite-950 border border-graphite-800 overflow-y-auto max-h-[420px]">
                {content.trim() ? (
                  <MarkdownView content={content} />
                ) : (
                  <p className="text-sm text-graphite-400 italic">{t("discussions.editor.splitPreviewEmpty")}</p>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="flex items-center justify-end gap-3 pt-4 border-t border-graphite-800">
          <button
            type="button"
            onClick={() => navigate(-1)}
            className="px-4 py-2 rounded-md border border-graphite-800 text-graphite-200 hover:text-white hover:bg-graphite-900 text-sm font-medium transition-colors cursor-pointer"
          >
            {t("discussions.editor.cancel")}
          </button>

          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-5 py-2 rounded-md bg-indigo-600 hover:bg-indigo-500 disabled:opacity-60 text-white text-sm font-semibold shadow-sm transition-all cursor-pointer"
          >
            {loading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>{t("discussions.editor.saving")}</span>
              </>
            ) : (
              <>
                <Save className="w-4 h-4" />
                <span>{isEditing ? t("discussions.editor.updatePost") : t("discussions.editor.publishPost")}</span>
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
