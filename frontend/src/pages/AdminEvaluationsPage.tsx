import { Fragment, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import axios from "axios";
import {
  ArrowRight,
  BarChart3,
  ChevronDown,
  ChevronUp,
  Download,
  FileJson,
  FileSpreadsheet,
  Loader2,
  Search,
  ShieldAlert,
} from "lucide-react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import AdminPortalTabs from "../components/AdminPortalTabs";
import ChartExportCard from "../components/ChartExportCard";
import PairedPrePostChartCard from "../components/PairedPrePostChartCard";
import QualitativeFeedbackList from "../components/QualitativeFeedbackList";
import TLXWorkloadRadarCard from "../components/TLXWorkloadRadarCard";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import {
  DARK_CHART_TOOLTIP_PROPS,
  buildCSUQDistributionData,
  buildCohortFunnelData,
  buildPrePostPairedData,
  buildTLXRadarData,
  type CSUQDistributionMode,
} from "../lib/evaluationCharts";
import type {
  CSUQScoresPayload,
  NASATLXScoresPayload,
  QualitativeDraft,
} from "../lib/questionnaire";
import { cn } from "../lib/utils";

interface CohortFunnel {
  total_learners: number;
  completed_pre_test_learners: number;
  completed_post_test_learners: number;
  questionnaire_respondents: number;
  response_rate_pct: number;
}

interface AdminMeanSD {
  key: string;
  n: number;
  mean: number | null;
  sd: number | null;
}

interface AdminEvaluationRow {
  eval_id: string;
  participant_code: string;
  nickname: string;
  submitted_locale: string;
  eval_date: string;
  pre_score: number;
  post_score: number;
  improvement_pct: number;
  completed_scenarios: number;
  average_proficiency: number;
  csuq_scores: CSUQScoresPayload;
  nasa_tlx_scores: NASATLXScoresPayload;
  qualitative_feedback: QualitativeDraft;
  notes: string;
}

interface AdminEvaluationsResponse {
  cohort: CohortFunnel;
  summary: {
    csuq_subscales: AdminMeanSD[];
    nasa_tlx_dimensions: AdminMeanSD[];
    learning_gain: {
      n: number;
      pre_mean: number | null;
      post_mean: number | null;
      improvement_mean: number | null;
      improvement_sd: number | null;
      qualitative_count: number;
    };
  };
  responses: AdminEvaluationRow[];
}

function formatMeanSD(mean: number | null, sd: number | null, suffix = ""): string {
  if (mean === null) return "N/A";
  if (sd === null) return `${mean.toFixed(2)}${suffix}`;
  return `${mean.toFixed(2)} ± ${sd.toFixed(2)}${suffix}`;
}

export default function AdminEvaluationsPage() {
  const { t, formatDate } = useI18n();
  const [data, setData] = useState<AdminEvaluationsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [exporting, setExporting] = useState<"csv" | "json" | null>(null);

  const [searchQuery, setSearchQuery] = useState("");
  const [localeFilter, setLocaleFilter] = useState<"all" | "en" | "th">("all");
  const [csuqDistMode, setCsuqDistMode] = useState<CSUQDistributionMode>("subscales");
  const [expandedEvalIds, setExpandedEvalIds] = useState<Set<string>>(() => new Set());

  const token = localStorage.getItem("token");
  const userRole = localStorage.getItem("userRole");

  useEffect(() => {
    let cancelled = false;

    async function loadEvaluations() {
      if (!token || userRole !== "admin") {
        setLoading(false);
        return;
      }
      try {
        setLoading(true);
        const res = await axios.get<AdminEvaluationsResponse>("/api/admin/evaluations", {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!cancelled) {
          setData(res.data);
          setError("");
        }
      } catch (err: unknown) {
        if (!cancelled) {
          setError(resolveApiError(err, t, t("admin.evaluations.loadFailed")));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void loadEvaluations();
    return () => {
      cancelled = true;
    };
  }, [token, userRole, t]);

  const handleExport = async (format: "csv" | "json") => {
    if (!token || exporting) return;
    setExporting(format);
    setError("");
    try {
      const res = await axios.get(`/api/admin/evaluations/export?format=${format}`, {
        headers: { Authorization: `Bearer ${token}` },
        responseType: "blob",
      });
      const blob = new Blob([res.data], {
        type: format === "csv" ? "text/csv;charset=utf-8;" : "application/json;charset=utf-8;",
      });
      const disposition = String(res.headers?.["content-disposition"] ?? "");
      const filenameMatch = disposition.match(/filename\*?=(?:UTF-8'')?["']?([^"';]+)["']?/i);
      const fallbackDate = new Date().toISOString().slice(0, 10).replace(/-/g, "");
      const downloadName = filenameMatch?.[1]?.trim() || `soc_trainer_evaluations_${fallbackDate}.${format}`;

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = downloadName;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (err: unknown) {
      setError(resolveApiError(err, t, t("admin.evaluations.exportFailed")));
    } finally {
      setExporting(null);
    }
  };

  const toggleExpandRow = (evalId: string) => {
    setExpandedEvalIds((prev) => {
      const next = new Set(prev);
      if (next.has(evalId)) {
        next.delete(evalId);
      } else {
        next.add(evalId);
      }
      return next;
    });
  };

  const filteredResponses = useMemo(() => {
    const rows = data?.responses ?? [];
    const q = searchQuery.trim().toLowerCase();
    return rows.filter((row) => {
      if (localeFilter !== "all" && row.submitted_locale !== localeFilter) {
        return false;
      }
      if (!q) return true;
      return (
        row.participant_code.toLowerCase().includes(q) ||
        row.nickname.toLowerCase().includes(q)
      );
    });
  }, [data, searchQuery, localeFilter]);

  const funnelChartData = useMemo(() => {
    if (!data) return [];
    return buildCohortFunnelData(data.cohort, {
      totalLearners: t("admin.evaluations.charts.funnelStageTotal"),
      preCompleted: t("admin.evaluations.charts.funnelStagePre"),
      postCompleted: t("admin.evaluations.charts.funnelStagePost"),
      respondents: t("admin.evaluations.charts.funnelStageSurvey"),
    });
  }, [data, t]);

  const csuqDistData = useMemo(() => {
    const labels: Record<string, string> = {
      sysuse: t("admin.evaluations.chartLabels.csuq.sysuse"),
      infoqual: t("admin.evaluations.chartLabels.csuq.infoqual"),
      interqual: t("admin.evaluations.chartLabels.csuq.interqual"),
      overall: t("admin.evaluations.chartLabels.csuq.overall"),
    };
    for (let i = 1; i <= 16; i++) {
      labels[`q${i}`] = `Q${i}`;
    }
    return buildCSUQDistributionData(data?.responses ?? [], csuqDistMode, labels);
  }, [data, csuqDistMode, t]);

  const tlxRadarData = useMemo(() => {
    const dims: Record<string, number | null> = {};
    for (const d of data?.summary.nasa_tlx_dimensions ?? []) {
      dims[d.key] = d.mean;
    }
    return buildTLXRadarData(dims, {
      mental_demand: t("admin.evaluations.chartLabels.tlx.mental_demand"),
      physical_demand: t("admin.evaluations.chartLabels.tlx.physical_demand"),
      temporal_demand: t("admin.evaluations.chartLabels.tlx.temporal_demand"),
      performance_workload: t("admin.evaluations.chartLabels.tlx.performance_workload"),
      effort: t("admin.evaluations.chartLabels.tlx.effort"),
      frustration: t("admin.evaluations.chartLabels.tlx.frustration"),
    });
  }, [data, t]);

  const pairedGainSeries = useMemo(
    () =>
      buildPrePostPairedData(data?.responses ?? [], {
        preMean: data?.summary.learning_gain.pre_mean,
        postMean: data?.summary.learning_gain.post_mean,
      }),
    [data],
  );

  if (userRole !== "admin") {
    return (
      <div className="min-h-[50vh] flex flex-col items-center justify-center space-y-4">
        <ShieldAlert className="w-12 h-12 text-rose-500" />
        <h2 className="text-xl font-bold text-white">{t("admin.accessDeniedTitle")}</h2>
        <p className="text-base text-graphite-300">{t("admin.accessDeniedDesc")}</p>
      </div>
    );
  }

  const csuqOverallStat = data?.summary.csuq_subscales.find((s) => s.key === "overall");
  const tlxOverallStat = data?.summary.nasa_tlx_dimensions.find((s) => s.key === "overall_workload");
  const gainSummary = data?.summary.learning_gain;

  return (
    <div className="max-w-6xl mx-auto space-y-8 animate-fade-in">
      <AdminPortalTabs active="evaluations" />

      {/* Header + Action Buttons */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 border-b border-graphite-800 pb-6">
        <div>
          <div className="flex items-center gap-2.5">
            <FileSpreadsheet className="w-6 h-6 text-emerald-400" />
            <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              {t("admin.evaluations.title")}
            </h1>
          </div>
          <p className="text-sm sm:text-base text-graphite-300 mt-1.5 leading-relaxed">
            {t("admin.evaluations.subtitle")}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2.5 shrink-0">
          <button
            type="button"
            onClick={() => void handleExport("csv")}
            disabled={exporting !== null || loading}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-md border border-graphite-700 bg-graphite-900 hover:bg-graphite-800 text-sm font-medium text-white disabled:opacity-50 transition-colors cursor-pointer"
          >
            {exporting === "csv" ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
            <span>{t("admin.evaluations.exportCsv")}</span>
          </button>

          <button
            type="button"
            onClick={() => void handleExport("json")}
            disabled={exporting !== null || loading}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-md border border-graphite-700 bg-graphite-900 hover:bg-graphite-800 text-sm font-medium text-white disabled:opacity-50 transition-colors cursor-pointer"
          >
            {exporting === "json" ? <Loader2 className="w-4 h-4 animate-spin" /> : <FileJson className="w-4 h-4" />}
            <span>{t("admin.evaluations.exportJson")}</span>
          </button>

          <Link
            to="/admin/evaluations/statistics"
            className="inline-flex items-center gap-2 px-4 py-2 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 transition-colors"
          >
            <BarChart3 className="w-4 h-4" />
            <span>{t("admin.evaluations.openStatsCta")}</span>
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-mono">
          {error}
        </div>
      )}

      {loading ? (
        <div className="py-20 flex flex-col items-center justify-center gap-3">
          <Loader2 className="w-8 h-8 animate-spin text-white" />
          <p className="text-sm font-mono text-graphite-400">{t("admin.evaluations.loading")}</p>
        </div>
      ) : data && (
        <>
          {/* KPI Summary Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
            <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-1.5">
              <div className="text-xs font-mono text-graphite-400 uppercase">
                {t("admin.evaluations.kpi.respondents")}
              </div>
              <div className="text-2xl font-bold font-mono text-white">
                {data.cohort.questionnaire_respondents}
              </div>
              <div className="text-xs font-mono text-emerald-400">
                {t("admin.evaluations.kpi.responseRate", {
                  rate: data.cohort.response_rate_pct.toFixed(1),
                  postCompleted: data.cohort.completed_post_test_learners,
                })}
              </div>
            </div>

            <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-1.5">
              <div className="text-xs font-mono text-graphite-400 uppercase">
                {t("admin.evaluations.kpi.csuqOverall")}
              </div>
              <div className="text-2xl font-bold font-mono text-emerald-400">
                {formatMeanSD(csuqOverallStat?.mean ?? null, csuqOverallStat?.sd ?? null)}
              </div>
              <div className="text-xs font-mono text-graphite-400">
                {t("admin.evaluations.kpi.csuqScale")}
              </div>
            </div>

            <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-1.5">
              <div className="text-xs font-mono text-graphite-400 uppercase">
                {t("admin.evaluations.kpi.tlxWorkload")}
              </div>
              <div className="text-2xl font-bold font-mono text-amber-300">
                {formatMeanSD(tlxOverallStat?.mean ?? null, tlxOverallStat?.sd ?? null)}
              </div>
              <div className="text-xs font-mono text-graphite-400">
                {t("admin.evaluations.kpi.tlxScale")}
              </div>
            </div>

            <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-1.5">
              <div className="text-xs font-mono text-graphite-400 uppercase">
                {t("admin.evaluations.kpi.learningGain")}
              </div>
              <div className="text-2xl font-bold font-mono text-white">
                {gainSummary?.improvement_mean !== null && gainSummary?.improvement_mean !== undefined
                  ? `${gainSummary.improvement_mean > 0 ? "+" : ""}${gainSummary.improvement_mean.toFixed(1)}%`
                  : "N/A"}
              </div>
              <div className="text-xs font-mono text-graphite-400">
                {gainSummary?.pre_mean !== null && gainSummary?.post_mean !== null && gainSummary
                  ? `${gainSummary.pre_mean?.toFixed(1)}% → ${gainSummary.post_mean?.toFixed(1)}%`
                  : t("admin.evaluations.kpi.noPairedScores")}
              </div>
            </div>

            <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-5 space-y-1.5">
              <div className="text-xs font-mono text-graphite-400 uppercase">
                {t("admin.evaluations.kpi.qualitativeCount")}
              </div>
              <div className="text-2xl font-bold font-mono text-white">
                {gainSummary?.qualitative_count ?? 0}
              </div>
              <div className="text-xs font-mono text-graphite-400">
                {t("admin.evaluations.kpi.withWrittenNotes")}
              </div>
            </div>
          </div>

          {/* 2x2 Recharts Visualization Suite */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* 1. Cohort Participation Funnel */}
            <ChartExportCard
              title={t("admin.evaluations.charts.funnelTitle")}
              subtitle={t("admin.evaluations.charts.funnelSubtitle")}
              exportFilename="cohort_participation_funnel.svg"
              isEmpty={data.cohort.total_learners === 0}
              emptyMessage={t("admin.evaluations.charts.funnelEmpty")}
              chartHeightClass="h-72"
            >
              <ResponsiveContainer width="100%" height="100%">
                <BarChart
                  data={funnelChartData}
                  layout="vertical"
                  margin={{ top: 8, right: 20, left: 8, bottom: 8 }}
                >
                  <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                  <XAxis type="number" allowDecimals={false} tick={{ fill: "#a1a1aa", fontSize: 12 }} />
                  <YAxis
                    type="category"
                    dataKey="stage"
                    width={126}
                    tick={{ fill: "#d4d4d8", fontSize: 12 }}
                  />
                  <Tooltip {...DARK_CHART_TOOLTIP_PROPS} />
                  <Bar
                    dataKey="count"
                    name={t("admin.evaluations.charts.funnelSeriesCount")}
                    fill="#38bdf8"
                    radius={[0, 4, 4, 0]}
                    barSize={26}
                  />
                </BarChart>
              </ResponsiveContainer>
            </ChartExportCard>

            {/* 2. CSUQ Stacked Likert Distribution */}
            <ChartExportCard
              title={t("admin.evaluations.charts.csuqStackedTitle")}
              subtitle={t("admin.evaluations.charts.csuqStackedSubtitle")}
              exportFilename={`csuq_likert_${csuqDistMode}.svg`}
              isEmpty={data.responses.length === 0}
              emptyMessage={t("admin.evaluations.table.emptyDesc")}
              chartHeightClass={csuqDistMode === "items" ? "h-[30rem]" : "h-72"}
              headerExtra={
                <div className="inline-flex rounded-md border border-graphite-800 bg-graphite-950 p-0.5">
                  {(["subscales", "items"] as const).map((m) => (
                    <button
                      key={m}
                      type="button"
                      onClick={() => setCsuqDistMode(m)}
                      className={cn(
                        "px-2.5 py-1 rounded text-xs font-mono font-semibold transition-colors cursor-pointer",
                        csuqDistMode === m
                          ? "bg-white text-black"
                          : "text-graphite-400 hover:text-white",
                      )}
                    >
                      {m === "subscales"
                        ? t("admin.evaluations.charts.csuqModeSubscales")
                        : t("admin.evaluations.charts.csuqModeItems")}
                    </button>
                  ))}
                </div>
              }
            >
              <ResponsiveContainer width="100%" height="100%">
                <BarChart
                  data={csuqDistData}
                  layout="vertical"
                  margin={{ top: 8, right: 16, left: 4, bottom: 8 }}
                >
                  <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                  <XAxis type="number" domain={[0, 100]} unit="%" tick={{ fill: "#a1a1aa", fontSize: 12 }} />
                  <YAxis
                    type="category"
                    dataKey="label"
                    width={96}
                    tick={{ fill: "#d4d4d8", fontSize: 12 }}
                  />
                  <Tooltip {...DARK_CHART_TOOLTIP_PROPS} />
                  <Legend wrapperStyle={{ fontSize: 12 }} />
                  <Bar dataKey="lowPct" stackId="likert" name={t("admin.evaluations.charts.csuqLow")} fill="#ef4444" />
                  <Bar dataKey="neutralPct" stackId="likert" name={t("admin.evaluations.charts.csuqNeutral")} fill="#f59e0b" />
                  <Bar dataKey="positivePct" stackId="likert" name={t("admin.evaluations.charts.csuqPositive")} fill="#38bdf8" />
                  <Bar dataKey="strongPct" stackId="likert" name={t("admin.evaluations.charts.csuqStrong")} fill="#10b981" />
                  <Bar dataKey="naPct" stackId="likert" name={t("admin.evaluations.charts.csuqNA")} fill="#52525b" />
                </BarChart>
              </ResponsiveContainer>
            </ChartExportCard>

            {/* 3. Raw NASA-TLX 6-Axis Radar */}
            <TLXWorkloadRadarCard
              title={t("admin.evaluations.charts.tlxRadarTitle")}
              subtitle={t("admin.evaluations.charts.tlxRadarSubtitle")}
              exportFilename="nasa_tlx_cohort_radar.svg"
              data={tlxRadarData}
              seriesName={t("admin.evaluations.charts.tlxCohortMean")}
              midpointName={t("admin.evaluations.charts.tlxMidpoint")}
              isEmpty={data.responses.length === 0}
              emptyMessage={t("admin.evaluations.table.emptyDesc")}
              chartHeightClass="h-72"
            />

            {/* 4. Pre-Test vs. Post-Test Chronological Paired Comparison */}
            <PairedPrePostChartCard
              title={t("admin.evaluations.charts.pairedGainTitle")}
              subtitle={t("admin.evaluations.charts.pairedGainSubtitle")}
              exportFilename="cohort_pre_post_paired.svg"
              emptyMessage={t("admin.evaluations.table.emptyDesc")}
              series={pairedGainSeries}
            />
          </div>

          {/* Pseudonymized Response Table */}
          <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-5">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-graphite-800 pb-4">
              <div>
                <h2 className="text-lg font-semibold text-white">
                  {t("admin.evaluations.table.title")}
                </h2>
                <p className="text-sm text-graphite-400">
                  {t("admin.evaluations.table.subtitle")}
                </p>
              </div>

              <div className="flex flex-wrap items-center gap-3">
                <div className="relative">
                  <Search className="w-4 h-4 text-graphite-400 absolute left-3 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder={t("admin.evaluations.table.searchPlaceholder")}
                    className="pl-9 pr-3 py-2 rounded-md bg-graphite-950 border border-graphite-800 text-sm text-white placeholder-graphite-400 focus:outline-none focus:border-white"
                  />
                </div>

                <div className="inline-flex rounded-md border border-graphite-800 bg-graphite-950 p-1">
                  {(["all", "en", "th"] as const).map((loc) => (
                    <button
                      key={loc}
                      type="button"
                      onClick={() => setLocaleFilter(loc)}
                      className={cn(
                        "px-3 py-1 rounded text-xs font-mono font-semibold uppercase transition-colors cursor-pointer",
                        localeFilter === loc
                          ? "bg-white text-black"
                          : "text-graphite-400 hover:text-white",
                      )}
                    >
                      {loc === "all" ? t("admin.evaluations.table.filterAll") : loc.toUpperCase()}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {filteredResponses.length === 0 ? (
              <div className="py-12 text-center space-y-2">
                <p className="text-base text-graphite-300 font-medium">
                  {t("admin.evaluations.table.emptyTitle")}
                </p>
                <p className="text-sm text-graphite-400">
                  {t("admin.evaluations.table.emptyDesc")}
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-graphite-800 text-xs font-mono text-graphite-400 uppercase">
                      <th className="py-3 pr-4">{t("admin.evaluations.table.colParticipant")}</th>
                      <th className="py-3 pr-4">{t("admin.evaluations.table.colPrePost")}</th>
                      <th className="py-3 pr-4">{t("admin.evaluations.table.colCsuq")}</th>
                      <th className="py-3 pr-4">{t("admin.evaluations.table.colTlx")}</th>
                      <th className="py-3 pr-4">{t("admin.evaluations.table.colLocale")}</th>
                      <th className="py-3 pr-4">{t("admin.evaluations.table.colDate")}</th>
                      <th className="py-3 text-right">{t("admin.evaluations.table.colDetails")}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-graphite-800/70 text-sm">
                    {filteredResponses.map((row) => {
                      const isExpanded = expandedEvalIds.has(row.eval_id);
                      const csuqVal = row.csuq_scores?.subscales?.overall;
                      const tlxVal = row.nasa_tlx_scores?.overall_workload;
                      const deltaPositive = row.improvement_pct > 0;
                      const deltaNegative = row.improvement_pct < 0;

                      return (
                        <Fragment key={row.eval_id}>
                          <tr className="hover:bg-graphite-950/50 transition-colors">
                            <td className="py-3.5 pr-4">
                              <div className="font-mono font-semibold text-white">{row.participant_code}</div>
                              <div className="text-xs text-graphite-400">{row.nickname}</div>
                            </td>
                            <td className="py-3.5 pr-4 font-mono">
                              <span className="text-graphite-300">
                                {row.pre_score.toFixed(0)}% → {row.post_score.toFixed(0)}%
                              </span>
                              <span
                                className={cn(
                                  "ml-2 px-2 py-0.5 rounded text-xs font-semibold border",
                                  deltaPositive
                                    ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                                    : deltaNegative
                                      ? "bg-amber-500/10 text-amber-300 border-amber-500/30"
                                      : "bg-graphite-800 text-graphite-300 border-graphite-700",
                                )}
                              >
                                {row.improvement_pct > 0 ? `+${row.improvement_pct.toFixed(1)}%` : `${row.improvement_pct.toFixed(1)}%`}
                              </span>
                            </td>
                            <td className="py-3.5 pr-4 font-mono font-semibold text-emerald-400">
                              {csuqVal !== null && csuqVal !== undefined ? `${csuqVal.toFixed(2)} / 7` : "N/A"}
                            </td>
                            <td className="py-3.5 pr-4 font-mono font-semibold text-amber-300">
                              {tlxVal !== undefined ? `${tlxVal.toFixed(1)} / 100` : "N/A"}
                            </td>
                            <td className="py-3.5 pr-4">
                              <span className="px-2 py-0.5 rounded text-xs font-mono uppercase bg-graphite-950 border border-graphite-800 text-graphite-300">
                                {row.submitted_locale}
                              </span>
                            </td>
                            <td className="py-3.5 pr-4 text-xs font-mono text-graphite-400">
                              {formatDate(row.eval_date)}
                            </td>
                            <td className="py-3.5 text-right">
                              <button
                                type="button"
                                onClick={() => toggleExpandRow(row.eval_id)}
                                className="inline-flex items-center gap-1 px-3 py-1.5 rounded border border-graphite-800 bg-graphite-950 hover:bg-graphite-800 text-xs font-medium text-graphite-200 cursor-pointer"
                              >
                                <span>{isExpanded ? t("admin.evaluations.table.hide") : t("admin.evaluations.table.inspect")}</span>
                                {isExpanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                              </button>
                            </td>
                          </tr>

                          {isExpanded && (
                            <tr className="bg-graphite-950/70">
                              <td colSpan={7} className="p-5 space-y-4">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                  <div className="rounded-md border border-graphite-800 bg-graphite-900 p-4 space-y-2">
                                    <div className="text-xs font-mono text-emerald-400 uppercase font-semibold">
                                      {t("admin.evaluations.table.csuqBreakdown")}
                                    </div>
                                    <div className="grid grid-cols-2 gap-2 text-sm font-mono">
                                      <div>SYSUSE: {row.csuq_scores?.subscales?.sysuse?.toFixed(2) ?? "N/A"}</div>
                                      <div>INFOQUAL: {row.csuq_scores?.subscales?.infoqual?.toFixed(2) ?? "N/A"}</div>
                                      <div>INTERQUAL: {row.csuq_scores?.subscales?.interqual?.toFixed(2) ?? "N/A"}</div>
                                      <div>OVERALL: {row.csuq_scores?.subscales?.overall?.toFixed(2) ?? "N/A"}</div>
                                    </div>
                                  </div>

                                  <div className="rounded-md border border-graphite-800 bg-graphite-900 p-4 space-y-2">
                                    <div className="text-xs font-mono text-amber-300 uppercase font-semibold">
                                      {t("admin.evaluations.table.tlxBreakdown")}
                                    </div>
                                    <div className="grid grid-cols-2 gap-2 text-sm font-mono">
                                      <div>Mental: {row.nasa_tlx_scores?.dimensions?.mental_demand ?? 0}</div>
                                      <div>Physical: {row.nasa_tlx_scores?.dimensions?.physical_demand ?? 0}</div>
                                      <div>Temporal: {row.nasa_tlx_scores?.dimensions?.temporal_demand ?? 0}</div>
                                      <div>
                                        Perf: {row.nasa_tlx_scores?.dimensions?.performance_rating ?? 0} (W:{row.nasa_tlx_scores?.dimensions?.performance_workload ?? 0})
                                      </div>
                                      <div>Effort: {row.nasa_tlx_scores?.dimensions?.effort ?? 0}</div>
                                      <div>Frustration: {row.nasa_tlx_scores?.dimensions?.frustration ?? 0}</div>
                                    </div>
                                  </div>
                                </div>

                                <div className="rounded-md border border-graphite-800 bg-graphite-900 p-4 space-y-3">
                                  <div className="text-xs font-mono text-graphite-300 uppercase font-semibold">
                                    {t("admin.evaluations.table.qualitativeTitle")}
                                  </div>
                                  <QualitativeFeedbackList
                                    feedback={row.qualitative_feedback}
                                    hasNotes={Boolean(row.notes)}
                                    cardClassName="space-y-0.5"
                                    labels={{
                                      mostImpressive: t("questionnaire.step3.mostImpressiveLabel"),
                                      areasToImprove: t("questionnaire.step3.areasToImproveLabel"),
                                      additionalComments: t("questionnaire.step3.additionalCommentsLabel"),
                                      emptyText: t("questionnaire.submittedView.noFeedback"),
                                    }}
                                  />
                                </div>
                              </td>
                            </tr>
                          )}
                        </Fragment>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
