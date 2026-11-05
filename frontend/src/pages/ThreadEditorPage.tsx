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

interface ScenarioOption {
  scenario_id: string;
  title: string;
  domain_id: string;
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
          if (item.scenario_id && !map.has(item.scenario_id)) {
            map.set(item.scenario_id, {
              scenario_id: item.scenario_id,
              title: item.scenario_title,
              domain_id: item.domain_id,
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
      const t = res.data;
      setTitle(t.title);
      setContent(t.content);
      if (t.scenario?.scenario_id) {
        setScenarioId(t.scenario.scenario_id);
      }
      if (Array.isArray(t.tags)) {
        setTags(t.tags);
      }
    } catch (err) {
      console.error("Failed to load thread:", err);
      setErrorMsg("Failed to load thread details.");
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
    setTags(tags.filter((t) => t !== tagToRemove));
  };

  const insertSnippet = (snippet: string) => {
    setContent((prev) => (prev ? `${prev}\n\n${snippet}` : snippet));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg("");

    if (title.trim().length < 5) {
      setErrorMsg("Title must be at least 5 characters long.");
      return;
    }
    if (content.trim().length < 20) {
      setErrorMsg("Content must be at least 20 characters long.");
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
        const payload: any = {
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
    } catch (err: any) {
      console.error("Submit thread error:", err);
      setErrorMsg(err.response?.data?.error || "Failed to publish analysis post.");
    } finally {
      setLoading(false);
    }
  };

  if (initialLoading) {
    return (
      <div className="py-24 text-center text-graphite-500 font-mono text-xs">
        Loading post editor...
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6 animate-fade-in pb-12">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-graphite-800 pb-4">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate(-1)}
            className="p-1.5 rounded-lg text-graphite-400 hover:text-white hover:bg-graphite-900 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
          </button>
          <div>
            <h1 className="text-xl font-bold tracking-tight text-white">
              {isEditing ? "Edit Analysis Post" : "Author New Specialist Analysis"}
            </h1>
            <p className="text-xs text-graphite-400">
              Share your tactical breakdown, triage methodology, and playbook guidance.
            </p>
          </div>
        </div>

        {/* View Switcher on larger screens */}
        <div className="flex items-center gap-1 p-1 rounded-lg bg-graphite-900 border border-graphite-800 text-xs">
          <button
            type="button"
            onClick={() => setActiveTab("write")}
            className={`px-3 py-1 rounded font-medium transition-colors ${
              activeTab === "write" ? "bg-graphite-800 text-white" : "text-graphite-400 hover:text-white"
            }`}
          >
            <Edit3 className="w-3.5 h-3.5 inline mr-1" />
            Write
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("preview")}
            className={`px-3 py-1 rounded font-medium transition-colors ${
              activeTab === "preview" ? "bg-graphite-800 text-white" : "text-graphite-400 hover:text-white"
            }`}
          >
            <Eye className="w-3.5 h-3.5 inline mr-1" />
            Preview
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("split")}
            className={`hidden md:block px-3 py-1 rounded font-medium transition-colors ${
              activeTab === "split" ? "bg-graphite-800 text-white" : "text-graphite-400 hover:text-white"
            }`}
          >
            Split
          </button>
        </div>
      </div>

      {errorMsg && (
        <div className="p-3.5 rounded-lg bg-red-950/30 border border-red-500/20 text-red-400 text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Scenario Link */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-graphite-300 flex items-center gap-1.5">
            <Crosshair className="w-3.5 h-3.5 text-indigo-400" />
            <span>Link to Scenario (Optional)</span>
          </label>
          <select
            value={scenarioId}
            onChange={(e) => setScenarioId(e.target.value)}
            disabled={isEditing}
            className="w-full px-3 py-2 rounded-lg bg-graphite-950 border border-graphite-800 text-graphite-200 text-xs focus:outline-none focus:border-graphite-600 disabled:opacity-60"
          >
            <option value="">-- General Analysis / Theoretical Concept --</option>
            {scenarios.map((sc) => (
              <option key={sc.scenario_id} value={sc.scenario_id}>
                {sc.title} ({sc.domain_id})
              </option>
            ))}
          </select>
          <p className="text-[11px] text-graphite-500">
            Linking a scenario embeds this post directly beneath the scenario's simulation evaluation page.
          </p>
        </div>

        {/* Title */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-graphite-300">
            Post Title <span className="text-red-400">*</span>
          </label>
          <input
            type="text"
            required
            placeholder="e.g., Tactical Deconstruction: Identifying False-Positive SIEM Alerts in Domain 4"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full px-3.5 py-2.5 rounded-lg bg-graphite-950 border border-graphite-800 text-white placeholder-graphite-600 text-sm focus:outline-none focus:border-graphite-600"
          />
        </div>

        {/* Tags */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-graphite-300">Tags</label>
          <div className="flex flex-wrap items-center gap-1.5 p-2 rounded-lg bg-graphite-950 border border-graphite-800">
            {tags.map((tag) => (
              <span
                key={tag}
                className="px-2 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-graphite-200 font-mono text-[11px] flex items-center gap-1"
              >
                #{tag}
                <button
                  type="button"
                  onClick={() => handleRemoveTag(tag)}
                  className="text-graphite-400 hover:text-white"
                >
                  <X className="w-3 h-3" />
                </button>
              </span>
            ))}
            <input
              type="text"
              placeholder="Type tag and press Enter..."
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleAddTag();
                }
              }}
              className="flex-1 min-w-[140px] bg-transparent text-xs text-white placeholder-graphite-600 focus:outline-none px-1"
            />
          </div>
        </div>

        {/* Quick-Insert Toolbar */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-graphite-400">Quick-Insert Blue Team Templates</span>
          </div>

          <div className="flex flex-wrap gap-2 text-xs">
            <button
              type="button"
              onClick={() =>
                insertSnippet(
                  `### Investigation Timeline\n| Step | Action Taken | Discovered Finding |\n|---|---|---|\n| 1 | Triage SIEM Alert | Suspicious PowerShell invocation |\n| 2 | Endpoint Inspection | Outbound C2 connection detected |\n| 3 | Host Isolation | VLAN quarantine applied |`
                )
              }
              className="px-2.5 py-1.5 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-300 hover:text-white flex items-center gap-1.5 text-xs transition-colors cursor-pointer"
            >
              <Clock className="w-3.5 h-3.5 text-indigo-400" />
              <span>Timeline Table</span>
            </button>

            <button
              type="button"
              onClick={() =>
                insertSnippet(
                  `### Indicators of Compromise (IOCs)\n| Type | Value | Context |\n|---|---|---|\n| IPv4 | 198.51.100.45 | C2 Command Beacon |\n| Hash (SHA256) | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 | Malicious Dropper |\n| Domain | update.security-patch.net | Phishing Delivery |`
                )
              }
              className="px-2.5 py-1.5 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-300 hover:text-white flex items-center gap-1.5 text-xs transition-colors cursor-pointer"
            >
              <Search className="w-3.5 h-3.5 text-amber-400" />
              <span>IOC Table</span>
            </button>

            <button
              type="button"
              onClick={() =>
                insertSnippet(
                  "### Recommended Detection Rule (Sigma)\n```yaml\ntitle: Suspicious Base64 Encoded PowerShell\nstatus: stable\nlogsource:\n  product: windows\n  service: security\ndetection:\n  selection:\n    EventID: 4688\n    CommandLine|contains:\n      - '-enc'\n      - '-encodedcommand'\n  condition: selection\nlevel: high\n```"
                )
              }
              className="px-2.5 py-1.5 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-300 hover:text-white flex items-center gap-1.5 text-xs transition-colors cursor-pointer"
            >
              <Shield className="w-3.5 h-3.5 text-emerald-400" />
              <span>Sigma Rule</span>
            </button>

            <button
              type="button"
              onClick={() =>
                insertSnippet(
                  "> 💡 **Key Takeaways & Lessons Learned**:\n> Always trace parent process execution lineage before validating alert fidelity to prevent false negative dismissals."
                )
              }
              className="px-2.5 py-1.5 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-300 hover:text-white flex items-center gap-1.5 text-xs transition-colors cursor-pointer"
            >
              <Lightbulb className="w-3.5 h-3.5 text-amber-400" />
              <span>Key Takeaways</span>
            </button>

            {evidenceReview?.evidence_review_available && (
              <button
                type="button"
                onClick={() => insertSnippet(buildEvidenceReviewMarkdown(evidenceReview))}
                className="px-2.5 py-1.5 rounded bg-cyan-950/40 hover:bg-cyan-900/50 border border-cyan-500/30 text-cyan-200 hover:text-white flex items-center gap-1.5 text-xs transition-colors cursor-pointer"
              >
                <Search className="w-3.5 h-3.5 text-cyan-400" />
                <span>Evidence Review</span>
              </button>
            )}
          </div>
        </div>

        {/* Content Area */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-graphite-300">
            Post Body (Markdown Supported) <span className="text-red-400">*</span>
          </label>

          {activeTab === "write" && (
            <textarea
              required
              rows={16}
              placeholder="Write comprehensive analysis, steps to triage, playbook recommendations, and evidence notes..."
              value={content}
              onChange={(e) => setContent(e.target.value)}
              className="w-full p-4 rounded-lg bg-graphite-950 border border-graphite-800 text-white placeholder-graphite-600 text-xs font-mono leading-relaxed focus:outline-none focus:border-graphite-600 resize-y"
            />
          )}

          {activeTab === "preview" && (
            <div className="p-5 rounded-lg bg-graphite-950 border border-graphite-800 min-h-[350px]">
              {content.trim() ? (
                <MarkdownView content={content} />
              ) : (
                <p className="text-xs text-graphite-500 italic">No content to preview.</p>
              )}
            </div>
          )}

          {activeTab === "split" && (
            <div className="grid grid-cols-2 gap-4">
              <textarea
                required
                rows={16}
                placeholder="Write markdown here..."
                value={content}
                onChange={(e) => setContent(e.target.value)}
                className="w-full p-4 rounded-lg bg-graphite-950 border border-graphite-800 text-white placeholder-graphite-600 text-xs font-mono leading-relaxed focus:outline-none focus:border-graphite-600 resize-y"
              />
              <div className="p-4 rounded-lg bg-graphite-950 border border-graphite-800 overflow-y-auto max-h-[420px]">
                {content.trim() ? (
                  <MarkdownView content={content} />
                ) : (
                  <p className="text-xs text-graphite-500 italic">Preview will appear here.</p>
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
            className="px-4 py-2 rounded-md border border-graphite-800 text-graphite-300 hover:text-white hover:bg-graphite-900 text-xs font-medium transition-colors"
          >
            Cancel
          </button>

          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-5 py-2 rounded-md bg-indigo-600 hover:bg-indigo-500 disabled:opacity-60 text-white text-xs font-semibold shadow-sm transition-all cursor-pointer"
          >
            {loading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>Saving...</span>
              </>
            ) : (
              <>
                <Save className="w-4 h-4" />
                <span>{isEditing ? "Update Post" : "Publish Post"}</span>
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
