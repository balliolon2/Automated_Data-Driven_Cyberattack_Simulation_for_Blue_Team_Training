import { useCallback, useEffect, useMemo, useRef, useState, type ChangeEvent } from "react";
import axios from "axios";
import {
  Activity,
  BookOpen,
  Camera,
  Check,
  ChevronDown,
  ChevronUp,
  Copy,
  Download,
  FileJson,
  FileText,
  GitCompare,
  Loader2,
  Play,
  Printer,
  RefreshCw,
  RotateCcw,
  ShieldAlert,
  Trash2,
  Upload,
} from "lucide-react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ComposedChart,
  Legend,
  Line,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import AdminPortalTabs from "../components/AdminPortalTabs";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import { isFiniteStat } from "../lib/paperStatistics";
import {
  NFR_METHODOLOGY_PILLARS,
  buildNfrEvaluationParagraph,
  buildNfrLatexTable,
  buildNfrMarkdownTable,
  computeNfrBaselineDelta,
  type NFRReportResponse,
  type NFRRunDetailResponse,
  type NFRRunListItem,
  type ScalabilityProbeResponse,
} from "../lib/nfrReport";
import { cn } from "../lib/utils";

type DimFilter = "all" | "pass" | "warn" | "fail" | "pending";
type ProbePreset = "light" | "standard" | "burst";

function verdictBadgeClass(status: string): string {
  switch (status) {
    case "all_passed":
    case "pass":
      return "bg-emerald-500/15 text-emerald-300 border-emerald-500/40";
    case "pass_with_warnings":
    case "warn":
      return "bg-amber-500/15 text-amber-300 border-amber-500/40";
    case "fail":
      return "bg-rose-500/15 text-rose-300 border-rose-500/40";
    default:
      return "bg-sky-500/15 text-sky-300 border-sky-500/40";
  }
}

function formatDurationSeconds(sec: number): string {
  const s = Math.max(0, Math.floor(sec));
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  const rem = s % 60;
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${rem}s`;
  return `${rem}s`;
}

function formatSigned(val: number | null, decimals = 2, suffix = "", lowerIsBetter = false): { text: string; tone: string } {
  if (!isFiniteStat(val)) return { text: "N/A", tone: "text-graphite-400" };
  if (val === 0) return { text: `0${suffix}`, tone: "text-graphite-300" };
  const sign = val > 0 ? "+" : "";
  const good = lowerIsBetter ? val < 0 : val > 0;
  return {
    text: `${sign}${val.toFixed(decimals)}${suffix}`,
    tone: good ? "text-emerald-400" : "text-amber-300",
  };
}

function shortEndpointLabel(route: string): string {
  const cleaned = route.replace("/api/", "/");
  return cleaned.length > 28 ? `${cleaned.slice(0, 26)}…` : cleaned;
}

export default function AdminNfrPage() {
  const { locale, t, formatDate } = useI18n();
  const token = localStorage.getItem("token");
  const userRole = localStorage.getItem("userRole");

  const [report, setReport] = useState<NFRReportResponse | null>(null);
  const [liveCache, setLiveCache] = useState<NFRReportResponse | null>(null);
  const [runs, setRuns] = useState<NFRRunListItem[]>([]);
  const [selectedRunId, setSelectedRunId] = useState<string>("live");
  const [baselineRunId, setBaselineRunId] = useState<string>("");
  const [baselineReport, setBaselineReport] = useState<NFRReportResponse | null>(null);

  const [loading, setLoading] = useState(true);
  const [busyAction, setBusyAction] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const [dimFilter, setDimFilter] = useState<DimFilter>("all");
  const [expandedDims, setExpandedDims] = useState<Set<string>>(() => new Set());
  const [showCaptureForm, setShowCaptureForm] = useState(false);
  const [captureLabel, setCaptureLabel] = useState("");
  const [captureEnv, setCaptureEnv] = useState("local");
  const [captureNotes, setCaptureNotes] = useState("");

  const [probePreset, setProbePreset] = useState<ProbePreset>("standard");
  const [probeSaveRun, setProbeSaveRun] = useState(true);

  const [paperLang, setPaperLang] = useState<"en" | "th">("en");
  const [showMethodology, setShowMethodology] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const authHeaders = useMemo(
    () => (token ? { Authorization: `Bearer ${token}` } : {}),
    [token],
  );

  const loadRunsList = useCallback(async () => {
    if (!token || userRole !== "admin") return;
    const res = await axios.get<{ runs: NFRRunListItem[] }>("/api/admin/nfr/runs?limit=50", {
      headers: authHeaders,
    });
    setRuns(res.data.runs ?? []);
  }, [token, userRole, authHeaders]);

  useEffect(() => {
    let cancelled = false;

    async function loadInitial() {
      if (!token || userRole !== "admin") {
        setLoading(false);
        return;
      }
      try {
        setLoading(true);
        const [liveRes, runsRes] = await Promise.all([
          axios.get<NFRReportResponse>("/api/admin/nfr/live", { headers: { Authorization: `Bearer ${token}` } }),
          axios.get<{ runs: NFRRunListItem[] }>("/api/admin/nfr/runs?limit=50", {
            headers: { Authorization: `Bearer ${token}` },
          }),
        ]);
        if (!cancelled) {
          setLiveCache(liveRes.data);
          setRuns(runsRes.data.runs ?? []);
          setReport(liveRes.data);
          setError("");
        }
      } catch (err: unknown) {
        if (!cancelled) {
          setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void loadInitial();
    return () => {
      cancelled = true;
    };
  }, [token, userRole, t]);

  const handleSelectView = async (runId: string) => {
    setSelectedRunId(runId);
    setError("");
    if (runId === "live") {
      if (liveCache) {
        setReport(liveCache);
      }
      setBusyAction("refresh");
      try {
        const res = await axios.get<NFRReportResponse>("/api/admin/nfr/live", { headers: authHeaders });
        setLiveCache(res.data);
        setReport(res.data);
      } catch (err: unknown) {
        setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
      } finally {
        setBusyAction(null);
      }
      return;
    }

    setBusyAction("view");
    try {
      const res = await axios.get<NFRRunDetailResponse>(`/api/admin/nfr/runs/${runId}`, { headers: authHeaders });
      setReport(res.data.report);
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
    } finally {
      setBusyAction(null);
    }
  };

  const handleSelectBaseline = async (runId: string) => {
    setBaselineRunId(runId);
    if (!runId) {
      setBaselineReport(null);
      return;
    }
    try {
      const res = await axios.get<NFRRunDetailResponse>(`/api/admin/nfr/runs/${runId}`, { headers: authHeaders });
      setBaselineReport(res.data.report);
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
    }
  };

  const handleCaptureSnapshot = async () => {
    if (!token || busyAction) return;
    setBusyAction("capture");
    setError("");
    setNotice("");
    try {
      const res = await axios.post<NFRRunDetailResponse>(
        "/api/admin/nfr/runs",
        {
          run_label: captureLabel.trim() || "Live NFR Snapshot",
          environment: captureEnv.trim() || "local",
          source: "live_snapshot",
          notes: captureNotes.trim(),
        },
        { headers: authHeaders },
      );
      await loadRunsList();
      setSelectedRunId(res.data.run.run_id);
      setReport(res.data.report);
      setShowCaptureForm(false);
      setCaptureLabel("");
      setCaptureNotes("");
      setNotice(t("admin.nfr.controls.snapshotSaved"));
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
    } finally {
      setBusyAction(null);
    }
  };

  const handleRunProbe = async (overridePreset?: ProbePreset) => {
    if (!token || busyAction) return;
    const presetToRun = overridePreset ?? probePreset;
    setBusyAction("probe");
    setError("");
    setNotice("");
    try {
      const res = await axios.post<ScalabilityProbeResponse>(
        "/api/admin/nfr/probe/scalability",
        {
          preset: presetToRun,
          save_run: probeSaveRun,
          environment: captureEnv.trim() || "local",
        },
        { headers: authHeaders },
      );
      const lv = res.data.benchmark.levels[0];
      setNotice(
        t("admin.nfr.controls.probeSuccess", {
          rps: lv ? lv.throughput_rps.toFixed(1) : "0.0",
          p95: lv ? lv.p95_ms.toFixed(1) : "0.0",
        }),
      );
      const [liveRes] = await Promise.all([
        axios.get<NFRReportResponse>("/api/admin/nfr/live", { headers: authHeaders }),
        loadRunsList(),
      ]);
      setLiveCache(liveRes.data);
      if (res.data.saved_run) {
        setSelectedRunId(res.data.saved_run.run_id);
        const detail = await axios.get<NFRRunDetailResponse>(`/api/admin/nfr/runs/${res.data.saved_run.run_id}`, {
          headers: authHeaders,
        });
        setReport(detail.data.report);
      } else {
        setSelectedRunId("live");
        setReport(liveRes.data);
      }
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
    } finally {
      setBusyAction(null);
    }
  };

  const handleResetWindow = async () => {
    if (!token || busyAction) return;
    if (!window.confirm(t("admin.nfr.controls.resetConfirm"))) return;
    setBusyAction("reset");
    setError("");
    setNotice("");
    try {
      await axios.post("/api/admin/nfr/telemetry/reset", {}, { headers: authHeaders });
      const liveRes = await axios.get<NFRReportResponse>("/api/admin/nfr/live", { headers: authHeaders });
      setLiveCache(liveRes.data);
      setSelectedRunId("live");
      setReport(liveRes.data);
      setNotice(t("admin.nfr.controls.resetSuccess"));
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
    } finally {
      setBusyAction(null);
    }
  };

  const handleImportFile = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file || !token) return;
    setError("");
    setNotice("");

    if (file.size > 1024 * 1024) {
      setError(t("admin.nfr.controls.importTooLarge"));
      return;
    }

    let parsed: Record<string, unknown>;
    try {
      const txt = await file.text();
      parsed = JSON.parse(txt) as Record<string, unknown>;
    } catch {
      setError(t("admin.nfr.controls.importInvalidJson"));
      return;
    }

    const payload = {
      run_label:
        typeof parsed.run_label === "string" && parsed.run_label.trim()
          ? parsed.run_label.trim()
          : file.name.replace(/\.json$/i, ""),
      environment:
        typeof parsed.environment === "string" && parsed.environment.trim()
          ? parsed.environment.trim()
          : "imported",
      source: "imported_json",
      notes: typeof parsed.notes === "string" ? parsed.notes : "",
      scalability: parsed.scalability ?? undefined,
      maintainability_report: parsed.maintainability_report ?? undefined,
    };

    setBusyAction("import");
    try {
      const res = await axios.post<NFRRunDetailResponse>("/api/admin/nfr/runs", payload, {
        headers: authHeaders,
      });
      await loadRunsList();
      setSelectedRunId(res.data.run.run_id);
      setReport(res.data.report);
      setNotice(t("admin.nfr.controls.importSuccess"));
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
    } finally {
      setBusyAction(null);
    }
  };

  const handleDeleteRun = async (run: NFRRunListItem) => {
    if (!token || busyAction) return;
    if (!window.confirm(t("admin.nfr.history.deleteConfirm", { label: run.run_label }))) return;
    setBusyAction("delete");
    setError("");
    setNotice("");
    try {
      await axios.delete(`/api/admin/nfr/runs/${run.run_id}`, { headers: authHeaders });
      await loadRunsList();
      if (baselineRunId === run.run_id) {
        setBaselineRunId("");
        setBaselineReport(null);
      }
      if (selectedRunId === run.run_id) {
        await handleSelectView("live");
      }
      setNotice(t("admin.nfr.history.deletedSuccess"));
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.nfr.loadFailed")));
    } finally {
      setBusyAction(null);
    }
  };

  const handleExport = async (format: "csv" | "json") => {
    if (!token || busyAction) return;
    setBusyAction(`export-${format}`);
    setError("");
    try {
      const runParam = selectedRunId !== "live" ? `&run_id=${encodeURIComponent(selectedRunId)}` : "";
      const res = await axios.get(`/api/admin/nfr/export?format=${format}${runParam}`, {
        headers: authHeaders,
        responseType: "blob",
      });
      const blob = new Blob([res.data], {
        type: format === "csv" ? "text/csv;charset=utf-8;" : "application/json;charset=utf-8;",
      });
      const disposition = String(res.headers?.["content-disposition"] ?? "");
      const match = disposition.match(/filename\*?=(?:UTF-8'')?["']?([^"';]+)["']?/i);
      const fallbackDate = new Date().toISOString().slice(0, 10).replace(/-/g, "");
      const filename = match?.[1]?.trim() || `nfr_evaluation_${fallbackDate}.${format}`;

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.nfr.exportFailed")));
    } finally {
      setBusyAction(null);
    }
  };

  const handleCopy = async (id: string, text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedId(id);
      window.setTimeout(() => {
        setCopiedId((prev) => (prev === id ? null : prev));
      }, 2000);
    } catch {
      // Clipboard fallback ignored if denied
    }
  };

  const toggleExpandDim = (id: string) => {
    setExpandedDims((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const filteredDimensions = useMemo(() => {
    const dims = report?.dimensions ?? [];
    if (dimFilter === "all") return dims;
    if (dimFilter === "pending") {
      return dims.filter((d) => d.status === "not_measured" || d.status === "insufficient_data");
    }
    return dims.filter((d) => d.status === dimFilter);
  }, [report, dimFilter]);

  const baselineDelta = useMemo(() => {
    if (!report || !baselineReport) return null;
    return computeNfrBaselineDelta(report, baselineReport);
  }, [report, baselineReport]);

  const endpointChartData = useMemo(() => {
    const eps = report?.runtime.endpoints ?? [];
    return eps.slice(0, 8).map((ep) => ({
      name: shortEndpointLabel(ep.route),
      p50: ep.p50_ms ?? 0,
      p95: ep.p95_ms ?? 0,
      p99: ep.p99_ms ?? 0,
    }));
  }, [report]);

  const scalabilityChartData = useMemo(() => {
    const levels = report?.scalability?.levels ?? [];
    return levels.map((lv) => ({
      workers: `${lv.concurrency}w`,
      rps: lv.throughput_rps,
      p95: lv.p95_ms,
      err: lv.error_rate_pct,
    }));
  }, [report]);

  const reliabilityChartData = useMemo(() => {
    const rb = report?.reliability_breakdown;
    if (!rb) return [];
    return [
      { name: t("admin.nfr.charts.barAiGenerated"), count: rb.ai_generated_count, fill: "#10b981" },
      { name: t("admin.nfr.charts.barStaticFallback"), count: rb.static_fallback_count, fill: "#f59e0b" },
      { name: t("admin.nfr.charts.barValidSnapshots"), count: rb.valid_snapshot_count, fill: "#38bdf8" },
      { name: t("admin.nfr.charts.barSystemFailures"), count: rb.system_failure_count, fill: "#f43f5e" },
    ];
  }, [report, t]);

  const paperParagraph = useMemo(() => {
    if (!report) return "";
    return buildNfrEvaluationParagraph(report, paperLang);
  }, [report, paperLang]);

  if (userRole !== "admin") {
    return (
      <div className="min-h-[50vh] flex flex-col items-center justify-center space-y-4">
        <ShieldAlert className="w-12 h-12 text-rose-500" />
        <h2 className="text-xl font-bold text-white">{t("admin.accessDeniedTitle")}</h2>
        <p className="text-base text-graphite-300">{t("admin.accessDeniedDesc")}</p>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto space-y-8 animate-fade-in">
      <AdminPortalTabs active="nfr" />

      {/* Header */}
      <div className="paper-report-section flex flex-col lg:flex-row lg:items-center justify-between gap-4 border-b border-graphite-800 pb-6 print:border-b print:pb-4 print:p-4 print:rounded-md">
        <div className="space-y-1.5">
          <div className="flex flex-wrap items-center gap-2.5">
            <Activity className="w-6 h-6 text-emerald-400 shrink-0" />
            <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              {t("admin.nfr.title")}
            </h1>
            {report && (
              <span className="px-2.5 py-0.5 rounded border border-graphite-700 bg-graphite-900 text-xs font-mono text-graphite-200">
                {report.rubric_version}
              </span>
            )}
          </div>
          <p className="text-base text-graphite-300 leading-relaxed">
            {t("admin.nfr.subtitle")}
          </p>
          {report && (
            <p className="text-xs font-mono text-graphite-400">
              {t("admin.nfr.metadataLine", {
                date: formatDate(report.evaluated_at),
                label: report.run_label,
                env: report.environment,
                source: t(`admin.nfr.sources.${report.source}`),
                ginMode: report.runtime.gin_mode,
              })}
            </p>
          )}
        </div>

        <div className="print:hidden flex flex-wrap items-center gap-2 shrink-0">
          <button
            type="button"
            onClick={() => void handleSelectView("live")}
            disabled={busyAction !== null || loading}
            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-md border border-graphite-700 bg-graphite-900 hover:bg-graphite-800 text-sm font-medium text-white disabled:opacity-50 transition-colors cursor-pointer"
          >
            <RefreshCw className={cn("w-4 h-4", busyAction === "refresh" && "animate-spin")} />
            <span>{t("admin.nfr.controls.refreshLive")}</span>
          </button>

          <button
            type="button"
            onClick={() => setShowCaptureForm((prev) => !prev)}
            disabled={busyAction !== null || loading}
            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-md border border-graphite-700 bg-graphite-900 hover:bg-graphite-800 text-sm font-medium text-white disabled:opacity-50 transition-colors cursor-pointer"
          >
            <Camera className="w-4 h-4 text-emerald-400" />
            <span>{t("admin.nfr.controls.captureSnapshot")}</span>
          </button>

          <button
            type="button"
            onClick={() => window.print()}
            disabled={!report}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 disabled:opacity-50 transition-colors cursor-pointer"
          >
            <Printer className="w-4 h-4" />
            <span>{t("admin.nfr.controls.printPdf")}</span>
          </button>
        </div>
      </div>

      {/* Section 1: Control & Run Bar */}
      <section className="print:hidden bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <label className="space-y-1.5 block">
            <span className="text-xs font-mono text-graphite-400 uppercase">{t("admin.nfr.controls.activeView")}</span>
            <select
              value={selectedRunId}
              onChange={(e) => void handleSelectView(e.target.value)}
              className="w-full px-3 py-2 rounded-md bg-graphite-950 border border-graphite-800 text-sm text-white focus:outline-none focus:border-white"
            >
              <option value="live">{t("admin.nfr.controls.liveNow")}</option>
              {runs.map((r) => (
                <option key={r.run_id} value={r.run_id}>
                  {`${r.run_label} (${r.environment}) · ${formatDate(r.created_at)} · ${r.passed_count}/${r.passed_count + r.warn_count + r.fail_count}`}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-1.5 block">
            <span className="text-xs font-mono text-graphite-400 uppercase flex items-center gap-1.5">
              <GitCompare className="w-3.5 h-3.5" />
              {t("admin.nfr.controls.compareBaseline")}
            </span>
            <select
              value={baselineRunId}
              onChange={(e) => void handleSelectBaseline(e.target.value)}
              className="w-full px-3 py-2 rounded-md bg-graphite-950 border border-graphite-800 text-sm text-white focus:outline-none focus:border-white"
            >
              <option value="">{t("admin.nfr.controls.noBaseline")}</option>
              {runs
                .filter((r) => r.run_id !== selectedRunId)
                .map((r) => (
                  <option key={r.run_id} value={r.run_id}>
                    {`${r.run_label} (${r.environment}) · ${formatDate(r.created_at)}`}
                  </option>
                ))}
            </select>
          </label>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3 pt-3 border-t border-graphite-800">
          <div className="flex flex-wrap items-center gap-2.5">
            <select
              aria-label={t("admin.nfr.controls.runProbe")}
              value={probePreset}
              onChange={(e) => setProbePreset(e.target.value as ProbePreset)}
              className="px-3 py-1.5 rounded-md bg-graphite-950 border border-graphite-800 text-sm text-white"
            >
              <option value="light">{t("admin.nfr.controls.probePresetLight")}</option>
              <option value="standard">{t("admin.nfr.controls.probePresetStandard")}</option>
              <option value="burst">{t("admin.nfr.controls.probePresetBurst")}</option>
            </select>

            <label className="inline-flex items-center gap-2 text-sm text-graphite-300 cursor-pointer">
              <input
                type="checkbox"
                checked={probeSaveRun}
                onChange={(e) => setProbeSaveRun(e.target.checked)}
                className="h-4 w-4 rounded border-graphite-700 bg-graphite-950 text-emerald-500"
              />
              <span>{t("admin.nfr.controls.saveProbeRun")}</span>
            </label>

            <button
              type="button"
              onClick={() => void handleRunProbe()}
              disabled={busyAction !== null || loading}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-md border border-emerald-500/40 bg-emerald-500/15 hover:bg-emerald-500/25 text-sm font-medium text-emerald-200 disabled:opacity-50 transition-colors cursor-pointer"
            >
              {busyAction === "probe" ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
              <span>{busyAction === "probe" ? t("admin.nfr.controls.probeRunning") : t("admin.nfr.controls.runProbe")}</span>
            </button>
          </div>

          <div className="flex flex-wrap items-center gap-2.5">
            <input
              ref={fileInputRef}
              type="file"
              accept=".json,application/json"
              onChange={(e) => void handleImportFile(e)}
              className="hidden"
            />
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={busyAction !== null || loading}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 hover:bg-graphite-800 text-sm font-medium text-graphite-200 disabled:opacity-50 transition-colors cursor-pointer"
            >
              <Upload className="w-4 h-4" />
              <span>{t("admin.nfr.controls.importJson")}</span>
            </button>

            <button
              type="button"
              onClick={() => void handleResetWindow()}
              disabled={busyAction !== null || loading}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-md border border-amber-500/30 bg-amber-500/10 hover:bg-amber-500/20 text-sm font-medium text-amber-200 disabled:opacity-50 transition-colors cursor-pointer"
            >
              <RotateCcw className="w-4 h-4" />
              <span>{t("admin.nfr.controls.resetWindow")}</span>
            </button>
          </div>
        </div>

        {/* Inline Capture Snapshot Form */}
        {showCaptureForm && (
          <div className="pt-4 border-t border-graphite-800 grid grid-cols-1 md:grid-cols-3 gap-3">
            <input
              type="text"
              value={captureLabel}
              onChange={(e) => setCaptureLabel(e.target.value)}
              placeholder={t("admin.nfr.controls.labelPlaceholder")}
              className="px-3 py-2 rounded-md bg-graphite-950 border border-graphite-800 text-sm text-white placeholder-graphite-400"
            />
            <input
              type="text"
              value={captureEnv}
              onChange={(e) => setCaptureEnv(e.target.value)}
              placeholder={t("admin.nfr.controls.envPlaceholder")}
              className="px-3 py-2 rounded-md bg-graphite-950 border border-graphite-800 text-sm text-white placeholder-graphite-400"
            />
            <div className="flex gap-2">
              <input
                type="text"
                value={captureNotes}
                onChange={(e) => setCaptureNotes(e.target.value)}
                placeholder={t("admin.nfr.controls.notesPlaceholder")}
                className="flex-1 px-3 py-2 rounded-md bg-graphite-950 border border-graphite-800 text-sm text-white placeholder-graphite-400"
              />
              <button
                type="button"
                onClick={() => void handleCaptureSnapshot()}
                disabled={busyAction !== null}
                className="px-4 py-2 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 disabled:opacity-50 cursor-pointer whitespace-nowrap"
              >
                {busyAction === "capture" ? t("admin.nfr.controls.savingSnapshot") : t("admin.nfr.controls.saveSnapshotBtn")}
              </button>
            </div>
          </div>
        )}
      </section>

      {error && (
        <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-mono">
          {error}
        </div>
      )}
      {notice && (
        <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-sm font-mono">
          {notice}
        </div>
      )}

      {loading ? (
        <div className="py-20 flex flex-col items-center justify-center gap-3">
          <Loader2 className="w-8 h-8 animate-spin text-white" />
          <p className="text-sm font-mono text-graphite-400">{t("admin.nfr.loading")}</p>
        </div>
      ) : report && (
        <>
          {/* Section 2: Executive KPI Strip */}
          <section className="paper-report-section space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
              <div className="paper-report-card bg-graphite-900 border border-graphite-800 rounded-lg p-4 space-y-1.5">
                <div className="text-xs font-mono text-graphite-400 uppercase">{t("admin.nfr.kpi.overallVerdict")}</div>
                <div>
                  <span className={cn("inline-block px-2.5 py-0.5 rounded border text-xs font-mono font-semibold", verdictBadgeClass(report.summary.overall_status))}>
                    {t(`admin.nfr.verdicts.${report.summary.overall_status}`)}
                  </span>
                </div>
                <div className="text-xs font-mono text-graphite-300">
                  {t("admin.nfr.kpi.measuredPassRate", { rate: report.summary.measured_pass_rate.toFixed(1) })}
                </div>
              </div>

              <div className="paper-report-card bg-graphite-900 border border-graphite-800 rounded-lg p-4 space-y-1.5">
                <div className="text-xs font-mono text-graphite-400 uppercase">{t("admin.nfr.kpi.rollupTitle")}</div>
                <div className="text-2xl font-bold font-mono text-white">
                  {`${report.summary.passed_count}/${report.summary.total_dimensions}`}
                </div>
                <div className="text-xs font-mono text-graphite-400">
                  {t("admin.nfr.kpi.rollupBreakdown", {
                    pass: report.summary.passed_count,
                    warn: report.summary.warn_count,
                    fail: report.summary.fail_count,
                    pending: report.summary.not_measured_count,
                  })}
                </div>
              </div>

              <div className="paper-report-card bg-graphite-900 border border-graphite-800 rounded-lg p-4 space-y-1.5">
                <div className="text-xs font-mono text-graphite-400 uppercase">{t("admin.nfr.kpi.nonLlmP95")}</div>
                <div className="text-2xl font-bold font-mono text-emerald-400">
                  {isFiniteStat(report.runtime.non_llm_latency.p95) ? `${report.runtime.non_llm_latency.p95.toFixed(1)} ms` : "N/A"}
                </div>
                <div className="text-xs font-mono text-graphite-400">
                  {t("admin.nfr.kpi.nonLlmTarget", {
                    target: report.thresholds.non_llm_p95_target_ms.toFixed(0),
                    count: report.runtime.non_llm_latency.count,
                  })}
                </div>
              </div>

              <div className="paper-report-card bg-graphite-900 border border-graphite-800 rounded-lg p-4 space-y-1.5">
                <div className="text-xs font-mono text-graphite-400 uppercase">{t("admin.nfr.kpi.availability")}</div>
                <div className="text-2xl font-bold font-mono text-white">
                  {isFiniteStat(report.runtime.availability_pct) ? `${report.runtime.availability_pct.toFixed(2)}%` : report.runtime.db_status.toUpperCase()}
                </div>
                <div className="text-xs font-mono text-graphite-400">
                  {t("admin.nfr.kpi.dbLine", {
                    status: report.runtime.db_status.toUpperCase(),
                    ms: report.runtime.db_latency_ms.toFixed(2),
                  })}
                </div>
              </div>

              <div className="paper-report-card bg-graphite-900 border border-graphite-800 rounded-lg p-4 space-y-1.5">
                <div className="text-xs font-mono text-graphite-400 uppercase">{t("admin.nfr.kpi.reliabilityRq4")}</div>
                <div className="text-2xl font-bold font-mono text-white">
                  {`${report.reliability_breakdown.valid_snapshot_count}/${report.reliability_breakdown.total_snapshot_count}`}
                </div>
                <div className="text-xs font-mono text-graphite-400">
                  {t("admin.nfr.kpi.reliabilitySub", {
                    ai: report.reliability_breakdown.ai_generated_count,
                    fallback: report.reliability_breakdown.static_fallback_count,
                    sysfail: report.reliability_breakdown.system_failure_count,
                  })}
                </div>
              </div>

              <div className="paper-report-card bg-graphite-900 border border-graphite-800 rounded-lg p-4 space-y-1.5">
                <div className="text-xs font-mono text-graphite-400 uppercase">{t("admin.nfr.kpi.uptimeWindow")}</div>
                <div className="text-2xl font-bold font-mono text-white">
                  {formatDurationSeconds(report.runtime.uptime_seconds)}
                </div>
                <div className="text-xs font-mono text-graphite-400">
                  {t("admin.nfr.kpi.requestsInWindow", { count: report.runtime.total_requests })}
                </div>
              </div>
            </div>

            {baselineDelta && baselineReport && (
              <div className="paper-report-card bg-graphite-900/90 border border-sky-500/30 rounded-lg p-4 space-y-2">
                <div className="text-xs font-mono font-semibold text-sky-300 uppercase">
                  {t("admin.nfr.delta.title", { baseline: baselineReport.run_label })}
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
                  {[
                    { label: t("admin.nfr.delta.nonLlmP95"), fmt: formatSigned(baselineDelta.nonLlmP95DeltaMs, 1, " ms", true) },
                    { label: t("admin.nfr.delta.availability"), fmt: formatSigned(baselineDelta.availabilityDeltaPct, 2, "%", false) },
                    { label: t("admin.nfr.delta.passedDims"), fmt: formatSigned(baselineDelta.passedCountDelta, 0, "", false) },
                    { label: t("admin.nfr.delta.passRate"), fmt: formatSigned(baselineDelta.measuredPassRateDelta, 1, "%", false) },
                    { label: t("admin.nfr.delta.peakRps"), fmt: formatSigned(baselineDelta.peakRpsDelta, 1, " RPS", false) },
                  ].map((item) => (
                    <div key={item.label} className="rounded bg-graphite-950 border border-graphite-800 px-3 py-2">
                      <div className="text-xs font-mono text-graphite-400">{item.label}</div>
                      <div className={cn("text-base font-mono font-bold", item.fmt.tone)}>{item.fmt.text}</div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </section>

          {/* Section 3: 10-Dimension ISO/IEC 25010 Scorecard Grid */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-5">
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 border-b border-graphite-800 pb-4">
              <div>
                <h2 className="text-lg font-semibold text-white">{t("admin.nfr.scorecard.title")}</h2>
                <p className="text-sm text-graphite-400">{t("admin.nfr.scorecard.subtitle")}</p>
              </div>

              <div className="print:hidden flex flex-wrap items-center gap-2">
                {(
                  [
                    { key: "all", label: t("admin.nfr.scorecard.filterAll", { count: report.dimensions.length }) },
                    { key: "pass", label: t("admin.nfr.scorecard.filterPass", { count: report.summary.passed_count }) },
                    { key: "warn", label: t("admin.nfr.scorecard.filterWarn", { count: report.summary.warn_count }) },
                    { key: "fail", label: t("admin.nfr.scorecard.filterFail", { count: report.summary.fail_count }) },
                    { key: "pending", label: t("admin.nfr.scorecard.filterPending", { count: report.summary.not_measured_count }) },
                  ] as const
                ).map((pill) => (
                  <button
                    key={pill.key}
                    type="button"
                    onClick={() => setDimFilter(pill.key)}
                    className={cn(
                      "px-3 py-1 rounded text-xs font-mono font-semibold border transition-colors cursor-pointer",
                      dimFilter === pill.key
                        ? "bg-white text-black border-white"
                        : "bg-graphite-950 text-graphite-300 border-graphite-800 hover:border-graphite-600",
                    )}
                  >
                    {pill.label}
                  </button>
                ))}

                <button
                  type="button"
                  onClick={() => setExpandedDims(new Set(report.dimensions.map((d) => d.id)))}
                  className="px-2.5 py-1 rounded text-xs font-mono text-graphite-300 border border-graphite-800 bg-graphite-950 hover:text-white cursor-pointer"
                >
                  {t("admin.nfr.scorecard.expandAll")}
                </button>
                <button
                  type="button"
                  onClick={() => setExpandedDims(new Set())}
                  className="px-2.5 py-1 rounded text-xs font-mono text-graphite-300 border border-graphite-800 bg-graphite-950 hover:text-white cursor-pointer"
                >
                  {t("admin.nfr.scorecard.collapseAll")}
                </button>
              </div>
            </div>

            {filteredDimensions.length === 0 ? (
              <p className="text-sm text-graphite-400 py-6 text-center">{t("admin.nfr.scorecard.emptyFilter")}</p>
            ) : (
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {filteredDimensions.map((dim) => {
                  const isOpen = expandedDims.has(dim.id);
                  return (
                    <div key={dim.id} className="paper-report-card rounded-lg border border-graphite-800 bg-graphite-950/60 p-4 space-y-3">
                      <div className="flex items-start justify-between gap-3">
                        <div className="space-y-1 min-w-0">
                          <h3 className="text-base font-semibold text-white">
                            {t(`admin.nfr.dimensions.${dim.id}`)}
                          </h3>
                          <p className="text-xs font-mono text-graphite-400">{dim.iso_characteristic}</p>
                        </div>

                        <div className="flex items-center gap-2 shrink-0">
                          <span className={cn("px-2.5 py-0.5 rounded border text-xs font-mono font-semibold", verdictBadgeClass(dim.status))}>
                            {t(`admin.nfr.statuses.${dim.status}`)}
                          </span>
                          <button
                            type="button"
                            onClick={() => toggleExpandDim(dim.id)}
                            aria-label={dim.id}
                            className="print:hidden p-1 rounded border border-graphite-800 text-graphite-300 hover:text-white cursor-pointer"
                          >
                            {isOpen ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                          </button>
                        </div>
                      </div>

                      <div className="text-sm font-mono font-semibold text-graphite-100">{dim.summary_metric}</div>
                      <div className="text-xs text-graphite-400 space-y-0.5">
                        <div>
                          <span className="font-mono text-graphite-300">{t("admin.nfr.scorecard.targetLabel")}</span>{" "}
                          {dim.target_expression}
                        </div>
                        <div>
                          <span className="font-mono text-graphite-300">{t("admin.nfr.scorecard.sourceLabel")}</span>{" "}
                          {dim.evidence_source}
                        </div>
                      </div>

                      {isOpen && dim.checks.length > 0 && (
                        <div className="pt-2 border-t border-graphite-800 overflow-x-auto">
                          <table className="w-full table-fixed text-left border-collapse">
                            <colgroup>
                              <col className="w-[34%]" />
                              <col className="w-[18%]" />
                              <col className="w-[23%]" />
                              <col className="w-[25%]" />
                            </colgroup>
                            <thead>
                              <tr className="text-xs font-mono text-graphite-400 border-b border-graphite-800">
                                <th className="py-1.5 pr-2">{t("admin.nfr.scorecard.colCheck")}</th>
                                <th className="py-1.5 px-2">{t("admin.nfr.scorecard.colMeasured")}</th>
                                <th className="py-1.5 px-2">{t("admin.nfr.scorecard.colTarget")}</th>
                                <th className="py-1.5 pl-1 text-right">{t("admin.nfr.scorecard.colStatus")}</th>
                              </tr>
                            </thead>
                            <tbody>
                              {dim.checks.map((chk) => (
                                <tr key={chk.key} className="border-b border-graphite-800/60 text-log font-mono align-top">
                                  <td className="py-1.5 pr-2 text-graphite-200 break-words">{chk.key}</td>
                                  <td className="py-1.5 px-2 text-white break-words">
                                    {chk.measured} {chk.unit !== "posture" && chk.unit !== "state" ? chk.unit : ""}
                                  </td>
                                  <td className="py-1.5 px-2 text-graphite-400 break-words">{chk.target}</td>
                                  <td className="py-1.5 pl-1 text-right whitespace-nowrap">
                                    <span className={cn("inline-block px-1.5 py-0.5 rounded border text-xs", verdictBadgeClass(chk.status))}>
                                      {t(`admin.nfr.statuses.${chk.status}`)}
                                    </span>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </section>

          {/* Section 4: Three Recharts Telemetry & Reliability Visualizations */}
          <section className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Chart 1: Endpoint API Latency */}
            <div className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-4">
              <div className="space-y-1">
                <h2 className="text-base font-semibold text-white">{t("admin.nfr.charts.latencyTitle")}</h2>
                <p className="text-xs font-mono text-graphite-400">
                  {t("admin.nfr.charts.latencySubtitle", { target: report.thresholds.non_llm_p95_target_ms.toFixed(0) })}
                </p>
              </div>
              {endpointChartData.length === 0 ? (
                <div className="h-64 flex items-center justify-center text-center p-4 rounded border border-graphite-800 bg-graphite-950/60 text-sm text-graphite-400">
                  {t("admin.nfr.charts.latencyEmpty")}
                </div>
              ) : (
                <div className="h-72 w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={endpointChartData} margin={{ top: 4, right: 8, left: -12, bottom: 12 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                      <XAxis dataKey="name" tick={{ fill: "#d4d4d8", fontSize: 12 }} interval={0} angle={-22} textAnchor="end" height={68} />
                      <YAxis tick={{ fill: "#a1a1aa", fontSize: 12 }} />
                      <Tooltip
                        contentStyle={{ backgroundColor: "#09090b", borderColor: "#3f3f46", color: "#f4f4f5", fontSize: 12 }}
                      />
                      <Legend verticalAlign="top" height={28} wrapperStyle={{ fontSize: 12 }} />
                      <ReferenceLine y={report.thresholds.non_llm_p95_target_ms} stroke="#f59e0b" strokeDasharray="4 4" />
                      <Bar dataKey="p50" name={t("admin.nfr.charts.seriesP50")} fill="#38bdf8" radius={[3, 3, 0, 0]} />
                      <Bar dataKey="p95" name={t("admin.nfr.charts.seriesP95")} fill="#10b981" radius={[3, 3, 0, 0]} />
                      <Bar dataKey="p99" name={t("admin.nfr.charts.seriesP99")} fill="#a855f7" radius={[3, 3, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              )}
            </div>

            {/* Chart 2: Scalability Concurrency Curve */}
            <div className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-4">
              <div className="space-y-1">
                <h2 className="text-base font-semibold text-white">{t("admin.nfr.charts.scalabilityTitle")}</h2>
                <p className="text-xs font-mono text-graphite-400">
                  {t("admin.nfr.charts.scalabilitySubtitle", { target: report.thresholds.scalability_p95_target_ms.toFixed(0) })}
                </p>
              </div>
              {scalabilityChartData.length === 0 ? (
                <div className="h-72 flex flex-col items-center justify-center gap-3 text-center p-4 rounded border border-graphite-800 bg-graphite-950/60">
                  <p className="text-sm text-graphite-400">{t("admin.nfr.charts.scalabilityEmpty")}</p>
                  <button
                    type="button"
                    onClick={() => void handleRunProbe("standard")}
                    disabled={busyAction !== null}
                    className="print:hidden px-3.5 py-1.5 rounded-md bg-emerald-500/20 border border-emerald-500/40 text-sm font-medium text-emerald-200 hover:bg-emerald-500/30 cursor-pointer"
                  >
                    {t("admin.nfr.charts.scalabilityRunCta")}
                  </button>
                </div>
              ) : (
                <div className="h-72 w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <ComposedChart data={scalabilityChartData} margin={{ top: 4, right: 8, left: -12, bottom: 8 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                      <XAxis dataKey="workers" tick={{ fill: "#d4d4d8", fontSize: 12 }} />
                      <YAxis yAxisId="left" tick={{ fill: "#10b981", fontSize: 12 }} />
                      <YAxis yAxisId="right" orientation="right" tick={{ fill: "#f59e0b", fontSize: 12 }} />
                      <Tooltip
                        contentStyle={{ backgroundColor: "#09090b", borderColor: "#3f3f46", color: "#f4f4f5", fontSize: 12 }}
                      />
                      <Legend verticalAlign="top" height={28} wrapperStyle={{ fontSize: 12 }} />
                      <ReferenceLine yAxisId="right" y={report.thresholds.scalability_p95_target_ms} stroke="#f59e0b" strokeDasharray="4 4" />
                      <Bar yAxisId="left" dataKey="rps" name={t("admin.nfr.charts.seriesRps")} fill="#10b981" radius={[3, 3, 0, 0]} />
                      <Line yAxisId="right" type="monotone" dataKey="p95" name={t("admin.nfr.charts.seriesP95")} stroke="#f59e0b" strokeWidth={2} />
                    </ComposedChart>
                  </ResponsiveContainer>
                </div>
              )}
            </div>

            {/* Chart 3: Simulation Pipeline Reliability */}
            <div className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-4">
              <div className="space-y-1">
                <h2 className="text-base font-semibold text-white">{t("admin.nfr.charts.reliabilityTitle")}</h2>
                <p className="text-xs font-mono text-graphite-400">{t("admin.nfr.charts.reliabilitySubtitle")}</p>
              </div>
              <div className="h-48 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={reliabilityChartData} margin={{ top: 8, right: 8, left: -12, bottom: 8 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                    <XAxis dataKey="name" tick={{ fill: "#d4d4d8", fontSize: 12 }} interval={0} angle={-12} textAnchor="end" height={42} />
                    <YAxis allowDecimals={false} tick={{ fill: "#a1a1aa", fontSize: 12 }} />
                    <Tooltip
                      contentStyle={{ backgroundColor: "#09090b", borderColor: "#3f3f46", color: "#f4f4f5", fontSize: 12 }}
                    />
                    <Bar dataKey="count" name={t("admin.nfr.charts.seriesCount")} radius={[3, 3, 0, 0]}>
                      {reliabilityChartData.map((entry) => (
                        <Cell key={entry.name} fill={entry.fill} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
              <div className="space-y-1.5 border-t border-graphite-800 pt-3">
                <div className="text-xs font-mono text-graphite-400">{t("admin.nfr.charts.fallbackReasonsTitle")}</div>
                {Object.keys(report.reliability_breakdown.fallback_reasons ?? {}).length === 0 ? (
                  <p className="text-xs text-graphite-400">{t("admin.nfr.charts.noFallbackReasons")}</p>
                ) : (
                  <div className="flex flex-wrap gap-1.5">
                    {Object.entries(report.reliability_breakdown.fallback_reasons).map(([reason, cnt]) => (
                      <span key={reason} className="px-2 py-0.5 rounded bg-graphite-950 border border-graphite-800 text-xs font-mono text-graphite-300">
                        {reason}: {cnt}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </section>

          {/* Section 5: Run History Table */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-4">
            <div>
              <h2 className="text-lg font-semibold text-white">{t("admin.nfr.history.title")}</h2>
              <p className="text-sm text-graphite-400">{t("admin.nfr.history.subtitle")}</p>
            </div>

            {runs.length === 0 ? (
              <p className="text-sm text-graphite-400 py-4">{t("admin.nfr.history.empty")}</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-graphite-800 text-xs font-mono text-graphite-400 uppercase">
                      <th className="py-2 pr-3">{t("admin.nfr.history.colDate")}</th>
                      <th className="py-2 px-2">{t("admin.nfr.history.colLabel")}</th>
                      <th className="py-2 px-2">{t("admin.nfr.history.colSource")}</th>
                      <th className="py-2 px-2">{t("admin.nfr.history.colVerdict")}</th>
                      <th className="py-2 px-2">{t("admin.nfr.history.colCounts")}</th>
                      <th className="py-2 px-2">{t("admin.nfr.history.colP95")}</th>
                      <th className="py-2 px-2">{t("admin.nfr.history.colAvail")}</th>
                      <th className="print:hidden py-2 pl-2 text-right">{t("admin.nfr.history.colActions")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {runs.map((run) => (
                      <tr key={run.run_id} className="border-b border-graphite-800/70 text-sm">
                        <td className="py-2.5 pr-3 font-mono text-xs text-graphite-300 whitespace-nowrap">
                          {formatDate(run.created_at)}
                        </td>
                        <td className="py-2.5 px-2">
                          <div className="font-medium text-white">{run.run_label}</div>
                          <div className="text-xs font-mono text-graphite-400">{run.environment}</div>
                        </td>
                        <td className="py-2.5 px-2 text-xs font-mono text-graphite-300 whitespace-nowrap">
                          {t(`admin.nfr.sources.${run.source}`)}
                        </td>
                        <td className="py-2.5 px-2 whitespace-nowrap">
                          <span className={cn("px-2 py-0.5 rounded border text-xs font-mono font-semibold", verdictBadgeClass(run.overall_status))}>
                            {t(`admin.nfr.verdicts.${run.overall_status}`)}
                          </span>
                        </td>
                        <td className="py-2.5 px-2 font-mono text-xs text-graphite-200 whitespace-nowrap">
                          {`${run.passed_count} / ${run.warn_count} / ${run.fail_count} / ${run.not_measured_count}`}
                        </td>
                        <td className="py-2.5 px-2 font-mono text-xs text-graphite-200 whitespace-nowrap">
                          {isFiniteStat(run.non_llm_p95_ms) ? `${run.non_llm_p95_ms.toFixed(1)} ms` : "N/A"}
                        </td>
                        <td className="py-2.5 px-2 font-mono text-xs text-graphite-200 whitespace-nowrap">
                          {isFiniteStat(run.availability_pct) ? `${run.availability_pct.toFixed(2)}%` : "N/A"}
                        </td>
                        <td className="print:hidden py-2.5 pl-2 text-right whitespace-nowrap space-x-1.5">
                          <button
                            type="button"
                            onClick={() => void handleSelectView(run.run_id)}
                            className="px-2.5 py-1 rounded border border-graphite-700 bg-graphite-950 text-xs font-mono text-graphite-200 hover:text-white cursor-pointer"
                          >
                            {t("admin.nfr.history.viewBtn")}
                          </button>
                          <button
                            type="button"
                            onClick={() => void handleSelectBaseline(baselineRunId === run.run_id ? "" : run.run_id)}
                            className={cn(
                              "px-2.5 py-1 rounded border text-xs font-mono cursor-pointer",
                              baselineRunId === run.run_id
                                ? "border-sky-500 bg-sky-500/20 text-sky-200"
                                : "border-graphite-700 bg-graphite-950 text-graphite-300 hover:text-white",
                            )}
                          >
                            {baselineRunId === run.run_id
                              ? t("admin.nfr.history.activeBaselineBadge")
                              : t("admin.nfr.history.baselineBtn")}
                          </button>
                          <button
                            type="button"
                            onClick={() => void handleDeleteRun(run)}
                            aria-label={`${t("admin.nfr.history.deleteBtn")} ${run.run_label}`}
                            className="inline-flex items-center px-2 py-1 rounded border border-rose-500/30 bg-rose-500/10 text-xs font-mono text-rose-300 hover:bg-rose-500/20 cursor-pointer"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          {/* Section 6: Paper-Ready Engineering Export & [R10] Methodology Guide */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-5">
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 border-b border-graphite-800 pb-4">
              <div>
                <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                  <FileText className="w-5 h-5 text-emerald-400" />
                  {t("admin.nfr.paper.title")}
                </h2>
                <p className="text-sm text-graphite-400">{t("admin.nfr.paper.subtitle")}</p>
              </div>

              <div className="print:hidden flex flex-wrap items-center gap-2">
                <button
                  type="button"
                  onClick={() => void handleCopy("nfr-latex", buildNfrLatexTable(report))}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white cursor-pointer"
                >
                  {copiedId === "nfr-latex" ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copiedId === "nfr-latex" ? t("admin.nfr.paper.copied") : t("admin.nfr.paper.copyLatex")}</span>
                </button>

                <button
                  type="button"
                  onClick={() => void handleCopy("nfr-md", buildNfrMarkdownTable(report))}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white cursor-pointer"
                >
                  {copiedId === "nfr-md" ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copiedId === "nfr-md" ? t("admin.nfr.paper.copied") : t("admin.nfr.paper.copyMarkdown")}</span>
                </button>

                <button
                  type="button"
                  onClick={() => void handleExport("csv")}
                  disabled={busyAction !== null}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white cursor-pointer"
                >
                  <Download className="w-3.5 h-3.5" />
                  <span>{t("admin.nfr.paper.exportCsv")}</span>
                </button>

                <button
                  type="button"
                  onClick={() => void handleExport("json")}
                  disabled={busyAction !== null}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white cursor-pointer"
                >
                  <FileJson className="w-3.5 h-3.5" />
                  <span>{t("admin.nfr.paper.exportJson")}</span>
                </button>
              </div>
            </div>

            {/* Auto-Generated NFR Evaluation Paragraph */}
            <div className="space-y-3">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <h3 className="text-base font-semibold text-white">{t("admin.nfr.paper.paragraphTitle")}</h3>
                <div className="print:hidden flex items-center gap-2">
                  <div className="inline-flex rounded-md border border-graphite-800 bg-graphite-950 p-1">
                    <button
                      type="button"
                      onClick={() => setPaperLang("en")}
                      className={cn(
                        "px-3 py-1 rounded text-xs font-mono font-semibold cursor-pointer",
                        paperLang === "en" ? "bg-white text-black" : "text-graphite-400 hover:text-white",
                      )}
                    >
                      {t("admin.nfr.paper.langEn")}
                    </button>
                    <button
                      type="button"
                      onClick={() => setPaperLang("th")}
                      className={cn(
                        "px-3 py-1 rounded text-xs font-mono font-semibold cursor-pointer",
                        paperLang === "th" ? "bg-white text-black" : "text-graphite-400 hover:text-white",
                      )}
                    >
                      {t("admin.nfr.paper.langTh")}
                    </button>
                  </div>
                  <button
                    type="button"
                    onClick={() => void handleCopy("nfr-para", paperParagraph)}
                    className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white cursor-pointer"
                  >
                    {copiedId === "nfr-para" ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                    <span>{copiedId === "nfr-para" ? t("admin.nfr.paper.copied") : t("admin.nfr.paper.copyParagraph")}</span>
                  </button>
                </div>
              </div>
              <p className="paper-report-card text-base text-graphite-100 leading-relaxed bg-graphite-950/60 rounded-md border border-graphite-800 p-4">
                {paperParagraph}
              </p>
            </div>

            {/* Collapsible ISO/IEC 25010 [R10] Methodology Guide */}
            <div className="border-t border-graphite-800 pt-4 space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <h3 className="text-base font-semibold text-white flex items-center gap-2">
                  <BookOpen className="w-4 h-4 text-emerald-400" />
                  {t("admin.nfr.paper.methodologyTitle")}
                </h3>
                <button
                  type="button"
                  onClick={() => setShowMethodology((prev) => !prev)}
                  className="print:hidden inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white cursor-pointer"
                >
                  <span>{showMethodology ? t("admin.nfr.paper.hideMethodology") : t("admin.nfr.paper.showMethodology")}</span>
                  {showMethodology ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                </button>
              </div>

              {showMethodology && (
                <div className="grid grid-cols-1 gap-4">
                  {NFR_METHODOLOGY_PILLARS.map((pillar) => {
                    const content = locale === "th" ? pillar.th : pillar.en;
                    return (
                      <div key={pillar.id} className="paper-report-card rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-2">
                        <div className="flex flex-wrap items-center justify-between gap-2">
                          <h4 className="text-base font-semibold text-white">{content.title}</h4>
                          {pillar.citation && (
                            <span className="text-xs font-mono text-emerald-400">{pillar.citation}</span>
                          )}
                        </div>
                        <div className="text-log font-mono text-amber-300 bg-graphite-900 rounded px-3 py-1.5 border border-graphite-800">
                          {content.formula}
                        </div>
                        <p className="text-base text-graphite-300 leading-relaxed">{content.body}</p>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </section>
        </>
      )}
    </div>
  );
}
