import { useEffect, useMemo, useState } from "react";
import axios from "axios";
import {
  AlertTriangle,
  BarChart3,
  BookOpen,
  Check,
  ChevronDown,
  ChevronUp,
  Copy,
  FileText,
  Loader2,
  Printer,
  ShieldAlert,
} from "lucide-react";
import {
  CartesianGrid,
  ResponsiveContainer,
  Scatter,
  ScatterChart,
  Tooltip,
  XAxis,
  YAxis,
  ZAxis,
} from "recharts";
import AdminPortalTabs from "../components/AdminPortalTabs";
import ChartExportCard from "../components/ChartExportCard";
import PairedPrePostChartCard from "../components/PairedPrePostChartCard";
import {
  DARK_CHART_TOOLTIP_PROPS,
  buildPrePostPairedData,
  buildSpearmanScatterSeries,
  isCorrelationChartUnlocked,
} from "../lib/evaluationCharts";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import { CSUQ_ITEMS, NASA_TLX_DIMENSIONS } from "../lib/questionnaire";
import {
  METHODOLOGY_PILLARS,
  buildAcademicResultsParagraph,
  buildAllLatexTables,
  buildLatexTable1,
  buildLatexTable1Items,
  buildLatexTable2,
  buildLatexTable3,
  buildMarkdownTable1,
  buildMarkdownTable1Items,
  buildMarkdownTable2,
  buildMarkdownTable3,
  isFiniteStat,
  type AdminPaperStatisticsResponse,
  type CronbachAlphaStat,
  type DescriptiveStat,
} from "../lib/paperStatistics";
import { cn } from "../lib/utils";

function fmtNum(val: number | null | undefined, decimals = 2): string {
  if (!isFiniteStat(val)) return "—";
  return val.toFixed(decimals);
}

function fmtPValue(val: number | null | undefined): string {
  if (!isFiniteStat(val)) return "—";
  if (val < 0.0001) return "< 0.0001";
  return val.toFixed(4);
}

function fmtCI(stat: DescriptiveStat): string {
  if (!isFiniteStat(stat.ci_low) || !isFiniteStat(stat.ci_high)) return "N/A (n ≥ 2)";
  return `[${stat.ci_low.toFixed(2)}, ${stat.ci_high.toFixed(2)}]`;
}

function fmtMedianIQR(stat: DescriptiveStat): string {
  if (!isFiniteStat(stat.median)) return "—";
  if (!isFiniteStat(stat.iqr)) return stat.median.toFixed(2);
  return `${stat.median.toFixed(2)} (${stat.iqr.toFixed(2)})`;
}

function fmtRange(stat: DescriptiveStat): string {
  if (!isFiniteStat(stat.min) || !isFiniteStat(stat.max)) return "—";
  return `${stat.min.toFixed(2)}–${stat.max.toFixed(2)}`;
}

function statusGateLabel(
  status: string,
  minRequired: number,
  t: (k: string, p?: Record<string, string | number>) => string,
): string {
  if (status === "zero_variance") return t("admin.statistics.gates.zeroVariance");
  return t("admin.statistics.gates.requiresN", { min: minRequired });
}

export default function AdminStatisticsPage() {
  const { locale, t, formatDate } = useI18n();
  const [data, setData] = useState<AdminPaperStatisticsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showItemTable, setShowItemTable] = useState(false);
  const [showMethodology, setShowMethodology] = useState(false);
  const [paperLang, setPaperLang] = useState<"en" | "th">("en");
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const token = localStorage.getItem("token");
  const userRole = localStorage.getItem("userRole");

  useEffect(() => {
    let cancelled = false;

    async function loadStats() {
      if (!token || userRole !== "admin") {
        setLoading(false);
        return;
      }
      try {
        setLoading(true);
        const res = await axios.get<AdminPaperStatisticsResponse>("/api/admin/evaluations/statistics", {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!cancelled) {
          setData(res.data);
          setError("");
        }
      } catch (err: unknown) {
        if (!cancelled) {
          setError(resolveApiError(err, t, t("admin.statistics.loadFailed")));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void loadStats();
    return () => {
      cancelled = true;
    };
  }, [token, userRole, t]);

  const pairedFigureSeries = useMemo(
    () =>
      buildPrePostPairedData(data?.participant_points ?? [], {
        preMean: data?.learning_gain.pre.mean,
        postMean: data?.learning_gain.post.mean,
      }),
    [data],
  );

  const scatterSeries = useMemo(
    () => buildSpearmanScatterSeries(data?.participant_points ?? []),
    [data],
  );

  const alphaByKey = useMemo(() => {
    const m = new Map<string, CronbachAlphaStat>();
    for (const r of data?.reliability ?? []) {
      m.set(r.key, r);
    }
    return m;
  }, [data]);

  const narrativeParagraph = useMemo(() => {
    if (!data) return "";
    return buildAcademicResultsParagraph(data, paperLang);
  }, [data, paperLang]);

  const handleCopy = async (id: string, text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedId(id);
      window.setTimeout(() => {
        setCopiedId((prev) => (prev === id ? null : prev));
      }, 2000);
    } catch {
      // Clipboard API fallback ignored if denied
    }
  };

  const csuqLabelForKey = (key: string): string => {
    if (key === "sysuse") return t("admin.statistics.constructs.sysuse");
    if (key === "infoqual") return t("admin.statistics.constructs.infoqual");
    if (key === "interqual") return t("admin.statistics.constructs.interqual");
    if (key === "overall") return t("admin.statistics.constructs.overall");
    const spec = CSUQ_ITEMS.find((i) => i.key === key);
    if (!spec) return key.toUpperCase();
    return `Q${spec.number}: ${locale === "th" ? spec.th : spec.en}`;
  };

  const tlxLabelForKey = (key: string): string => {
    if (key === "performance_workload") return t("admin.statistics.constructs.performanceWorkload");
    if (key === "overall_workload") return t("admin.statistics.constructs.overallWorkload");
    if (key === "performance_rating") return t("admin.statistics.constructs.performanceRating");
    const dim = NASA_TLX_DIMENSIONS.find((d) => d.key === key);
    if (!dim) return key;
    return locale === "th" ? dim.th.label : dim.en.label;
  };

  if (userRole !== "admin") {
    return (
      <div className="min-h-[50vh] flex flex-col items-center justify-center space-y-4">
        <ShieldAlert className="w-12 h-12 text-rose-500" />
        <h2 className="text-xl font-bold text-white">{t("admin.accessDeniedTitle")}</h2>
        <p className="text-base text-graphite-300">{t("admin.accessDeniedDesc")}</p>
      </div>
    );
  }

  const renderCopyActions = (idPrefix: string, latexText: string, mdText: string) => (
    <div className="print:hidden flex flex-wrap items-center gap-2">
      <button
        type="button"
        onClick={() => void handleCopy(`${idPrefix}-latex`, latexText)}
        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white transition-colors cursor-pointer"
      >
        {copiedId === `${idPrefix}-latex` ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
        <span>
          {copiedId === `${idPrefix}-latex`
            ? t("admin.statistics.copyTools.copied")
            : t("admin.statistics.copyTools.copyLatex")}
        </span>
      </button>
      <button
        type="button"
        onClick={() => void handleCopy(`${idPrefix}-md`, mdText)}
        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white transition-colors cursor-pointer"
      >
        {copiedId === `${idPrefix}-md` ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
        <span>
          {copiedId === `${idPrefix}-md`
            ? t("admin.statistics.copyTools.copied")
            : t("admin.statistics.copyTools.copyMarkdown")}
        </span>
      </button>
    </div>
  );

  const renderDescriptiveRow = (
    stat: DescriptiveStat,
    label: string,
    alphaEntry?: CronbachAlphaStat,
    compactLabel = false,
  ) => (
    <tr key={stat.key} className="border-b border-graphite-800/70">
      <td className={cn("py-3 pr-4 text-graphite-100", compactLabel ? "text-sm max-w-md" : "text-sm font-medium")}>
        {label}
      </td>
      <td className="py-3 px-2 font-mono text-sm text-graphite-200">{stat.n}</td>
      <td className="py-3 px-2 font-mono text-sm font-semibold text-white">{fmtNum(stat.mean)}</td>
      <td className="py-3 px-2 font-mono text-sm text-graphite-300">
        {isFiniteStat(stat.sd) ? fmtNum(stat.sd) : statusGateLabel(stat.status, 2, t)}
      </td>
      <td className="py-3 px-2 font-mono text-xs text-graphite-300 whitespace-nowrap">{fmtCI(stat)}</td>
      <td className="py-3 px-2 font-mono text-xs text-graphite-300 whitespace-nowrap">{fmtMedianIQR(stat)}</td>
      <td className="py-3 px-2 font-mono text-xs text-graphite-400 whitespace-nowrap">{fmtRange(stat)}</td>
      {alphaEntry !== undefined && (
        <td className="py-3 px-2 font-mono text-xs text-emerald-300 whitespace-nowrap">
          {isFiniteStat(alphaEntry.alpha)
            ? `${alphaEntry.alpha.toFixed(3)} (n=${alphaEntry.complete_cases_n})`
            : statusGateLabel(alphaEntry.status, 2, t)}
        </td>
      )}
      <td className="py-3 pl-2 text-xs font-mono">
        <span className="px-2 py-0.5 rounded border border-graphite-700 bg-graphite-950 text-graphite-300">
          {t(`admin.statistics.bands.${stat.band}`)}
        </span>
      </td>
    </tr>
  );

  return (
    <div className="max-w-6xl mx-auto space-y-8 animate-fade-in">
      <AdminPortalTabs active="statistics" />

      {/* Header + Global Paper Toolbar */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 border-b border-graphite-800 pb-6">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2.5">
            <BarChart3 className="w-6 h-6 text-emerald-400" />
            <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              {t("admin.statistics.title")}
            </h1>
          </div>
          <p className="text-sm sm:text-base text-graphite-300 leading-relaxed">
            {t("admin.statistics.subtitle")}
          </p>
          {data && (
            <p className="text-xs font-mono text-graphite-400">
              {t("admin.statistics.metadataLine", {
                date: formatDate(data.generated_at),
                protocol: data.protocol_version,
                csuq: data.csuq_schema_version,
                tlx: data.nasa_tlx_schema_version,
              })}
            </p>
          )}
        </div>

        {data && (
          <div className="print:hidden flex flex-wrap items-center gap-2.5 shrink-0">
            <button
              type="button"
              onClick={() => void handleCopy("bundle-latex", buildAllLatexTables(data))}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-md border border-graphite-700 bg-graphite-900 hover:bg-graphite-800 text-sm font-medium text-white transition-colors cursor-pointer"
            >
              {copiedId === "bundle-latex" ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
              <span>
                {copiedId === "bundle-latex"
                  ? t("admin.statistics.copyTools.copied")
                  : t("admin.statistics.copyTools.copyAllLatex")}
              </span>
            </button>

            <button
              type="button"
              onClick={() => window.print()}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-md bg-white text-black text-sm font-semibold hover:bg-white/90 transition-colors cursor-pointer"
            >
              <Printer className="w-4 h-4" />
              <span>{t("admin.statistics.copyTools.printPdf")}</span>
            </button>
          </div>
        )}
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-mono">
          {error}
        </div>
      )}

      {loading ? (
        <div className="py-20 flex flex-col items-center justify-center gap-3">
          <Loader2 className="w-8 h-8 animate-spin text-white" />
          <p className="text-sm font-mono text-graphite-400">{t("admin.statistics.loading")}</p>
        </div>
      ) : data && (
        <>
          {/* Pilot Sample Caveat Banner */}
          {data.is_pilot_sample && (
            <div className="paper-report-section rounded-lg border border-amber-500/40 bg-amber-500/10 p-5 flex items-start gap-3.5">
              <AlertTriangle className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
              <div className="space-y-1">
                <h2 className="text-base font-semibold text-amber-300">
                  {t("admin.statistics.pilotCaveatTitle", { n: data.cohort.questionnaire_respondents })}
                </h2>
                <p className="text-sm text-graphite-200 leading-relaxed">
                  {t("admin.statistics.pilotCaveatDesc")}
                </p>
              </div>
            </div>
          )}

          {/* Section 1: Participant Flow Funnel */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-4">
            <div>
              <h2 className="text-lg font-semibold text-white">{t("admin.statistics.funnel.title")}</h2>
              <p className="text-sm text-graphite-400">{t("admin.statistics.funnel.subtitle")}</p>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
              {[
                { label: t("admin.statistics.funnel.totalLearners"), val: String(data.cohort.total_learners) },
                { label: t("admin.statistics.funnel.preCompleted"), val: String(data.cohort.completed_pre_test_learners) },
                { label: t("admin.statistics.funnel.postCompleted"), val: String(data.cohort.completed_post_test_learners) },
                { label: t("admin.statistics.funnel.respondents"), val: String(data.cohort.questionnaire_respondents) },
                { label: t("admin.statistics.funnel.responseRate"), val: `${data.cohort.response_rate_pct.toFixed(1)}%` },
              ].map((cell) => (
                <div key={cell.label} className="paper-report-card rounded-md border border-graphite-800 bg-graphite-950/60 p-4">
                  <div className="text-xs font-mono text-graphite-400">{cell.label}</div>
                  <div className="text-2xl font-bold font-mono text-white mt-1">{cell.val}</div>
                </div>
              ))}
            </div>
          </section>

          {/* Auto-Generated Academic Results Paragraph */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-graphite-800 pb-4">
              <div className="space-y-1">
                <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                  <FileText className="w-5 h-5 text-emerald-400" />
                  {t("admin.statistics.narrative.title")}
                </h2>
                <p className="text-sm text-graphite-400">{t("admin.statistics.narrative.subtitle")}</p>
              </div>

              <div className="print:hidden flex flex-wrap items-center gap-2.5">
                <div className="inline-flex rounded-md border border-graphite-800 bg-graphite-950 p-1">
                  <button
                    type="button"
                    onClick={() => setPaperLang("en")}
                    className={cn(
                      "px-3 py-1 rounded text-xs font-mono font-semibold transition-colors cursor-pointer",
                      paperLang === "en" ? "bg-white text-black" : "text-graphite-400 hover:text-white",
                    )}
                  >
                    {t("admin.statistics.narrative.langEn")}
                  </button>
                  <button
                    type="button"
                    onClick={() => setPaperLang("th")}
                    className={cn(
                      "px-3 py-1 rounded text-xs font-mono font-semibold transition-colors cursor-pointer",
                      paperLang === "th" ? "bg-white text-black" : "text-graphite-400 hover:text-white",
                    )}
                  >
                    {t("admin.statistics.narrative.langTh")}
                  </button>
                </div>

                <button
                  type="button"
                  onClick={() => void handleCopy("narrative", narrativeParagraph)}
                  className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white transition-colors cursor-pointer"
                >
                  {copiedId === "narrative" ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>
                    {copiedId === "narrative"
                      ? t("admin.statistics.copyTools.copied")
                      : t("admin.statistics.narrative.copyParagraph")}
                  </span>
                </button>
              </div>
            </div>

            <p className="text-base text-graphite-100 leading-relaxed bg-graphite-950/60 paper-report-card rounded-md border border-graphite-800 p-4">
              {narrativeParagraph}
            </p>
          </section>

          {data.cohort.questionnaire_respondents === 0 && (
            <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-8 text-center space-y-2">
              <p className="text-base font-semibold text-white">{t("admin.statistics.emptyTitle")}</p>
              <p className="text-sm text-graphite-400">{t("admin.statistics.emptyDesc")}</p>
            </div>
          )}

          {/* Section 2: Table I - CSUQ Usability & Reliability */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-5">
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3 border-b border-graphite-800 pb-4">
              <div>
                <h2 className="text-lg font-semibold text-white">{t("admin.statistics.table1.title")}</h2>
                <p className="text-sm text-graphite-400">{t("admin.statistics.table1.subtitle")}</p>
              </div>
              <div className="print:hidden flex flex-wrap items-center gap-2">
                {renderCopyActions("table1", buildLatexTable1(data), buildMarkdownTable1(data))}
                <button
                  type="button"
                  onClick={() => setShowItemTable((prev) => !prev)}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white transition-colors cursor-pointer"
                >
                  <span>
                    {showItemTable
                      ? t("admin.statistics.table1.hideItems")
                      : t("admin.statistics.table1.showItems")}
                  </span>
                  {showItemTable ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                </button>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-graphite-800 text-xs font-mono text-graphite-400 uppercase">
                    <th className="py-2.5 pr-4">{t("admin.statistics.cols.construct")}</th>
                    <th className="py-2.5 px-2">n</th>
                    <th className="py-2.5 px-2">M</th>
                    <th className="py-2.5 px-2">SD</th>
                    <th className="py-2.5 px-2">95% CI</th>
                    <th className="py-2.5 px-2">Median (IQR)</th>
                    <th className="py-2.5 px-2">Min–Max</th>
                    <th className="py-2.5 px-2">Cronbach&apos;s α</th>
                    <th className="py-2.5 pl-2">{t("admin.statistics.cols.band")}</th>
                  </tr>
                </thead>
                <tbody>
                  {data.csuq_subscales.map((stat) =>
                    renderDescriptiveRow(
                      stat,
                      csuqLabelForKey(stat.key),
                      alphaByKey.get(`csuq_${stat.key}`),
                    ),
                  )}
                </tbody>
              </table>
            </div>

            {showItemTable && (
              <div className="pt-4 border-t border-graphite-800 space-y-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <h3 className="text-base font-semibold text-emerald-400">
                    {t("admin.statistics.table1.itemsTitle")}
                  </h3>
                  {renderCopyActions("table1-items", buildLatexTable1Items(data), buildMarkdownTable1Items(data))}
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse">
                    <thead>
                      <tr className="border-b border-graphite-800 text-xs font-mono text-graphite-400 uppercase">
                        <th className="py-2 pr-4">{t("admin.statistics.cols.item")}</th>
                        <th className="py-2 px-2">n</th>
                        <th className="py-2 px-2">M</th>
                        <th className="py-2 px-2">SD</th>
                        <th className="py-2 px-2">95% CI</th>
                        <th className="py-2 px-2">Median (IQR)</th>
                        <th className="py-2 px-2">Min–Max</th>
                        <th className="py-2 pl-2">{t("admin.statistics.cols.band")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {data.csuq_items.map((item) =>
                        renderDescriptiveRow(item, csuqLabelForKey(item.key), undefined, true),
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </section>

          {/* Section 3: Table II - Raw NASA-TLX Cognitive Workload */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-4">
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3 border-b border-graphite-800 pb-4">
              <div>
                <h2 className="text-lg font-semibold text-white">{t("admin.statistics.table2.title")}</h2>
                <p className="text-sm text-graphite-400">{t("admin.statistics.table2.subtitle")}</p>
              </div>
              <div className="flex flex-wrap items-center gap-2.5">
                {(() => {
                  const tlxAlpha = alphaByKey.get("nasa_tlx_overall");
                  return (
                    <div className="px-3 py-1.5 rounded border border-graphite-800 bg-graphite-950 text-xs font-mono text-emerald-300">
                      Cronbach&apos;s α (6 dims):{" "}
                      {isFiniteStat(tlxAlpha?.alpha)
                        ? `${tlxAlpha.alpha.toFixed(3)} (${t(`admin.statistics.bands.${tlxAlpha.band}`)})`
                        : statusGateLabel(tlxAlpha?.status ?? "insufficient_n", 2, t)}
                    </div>
                  );
                })()}
                {renderCopyActions("table2", buildLatexTable2(data), buildMarkdownTable2(data))}
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-graphite-800 text-xs font-mono text-graphite-400 uppercase">
                    <th className="py-2.5 pr-4">{t("admin.statistics.cols.dimension")}</th>
                    <th className="py-2.5 px-2">n</th>
                    <th className="py-2.5 px-2">M</th>
                    <th className="py-2.5 px-2">SD</th>
                    <th className="py-2.5 px-2">95% CI</th>
                    <th className="py-2.5 px-2">Median (IQR)</th>
                    <th className="py-2.5 px-2">Min–Max</th>
                    <th className="py-2.5 pl-2">{t("admin.statistics.cols.band")}</th>
                  </tr>
                </thead>
                <tbody>
                  {data.nasa_tlx.map((stat) =>
                    renderDescriptiveRow(stat, tlxLabelForKey(stat.key)),
                  )}
                </tbody>
              </table>
            </div>
          </section>

          {/* Visual Figures Section (Paired Pre->Post + 3 Spearman Correlation Scatter Plots) */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-6">
            <div className="border-b border-graphite-800 pb-4 space-y-1">
              <h2 className="text-xl font-semibold text-white flex items-center gap-2">
                <BarChart3 className="w-5 h-5 text-emerald-400" />
                {t("admin.statistics.figures.sectionTitle")}
              </h2>
              <p className="text-sm text-graphite-400">
                {t("admin.statistics.figures.sectionSubtitle")}
              </p>
            </div>

            <PairedPrePostChartCard
              className="paper-report-card"
              headingLevel="h3"
              title={t("admin.statistics.figures.fig1Title")}
              subtitle={t("admin.statistics.figures.fig1Subtitle")}
              exportFilename="paper_figure1_pre_post_paired.svg"
              emptyMessage={t("admin.statistics.emptyDesc")}
              series={pairedFigureSeries}
            />

            <div className="space-y-3">
              <h3 className="text-lg font-semibold text-white">
                {t("admin.statistics.figures.fig2Title")}
              </h3>
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
                {[
                  {
                    corrKey: "learning_gain_vs_csuq_overall",
                    title: t("admin.statistics.figures.scatterGainCsuqTitle"),
                    xLabel: t("admin.statistics.figures.axisGainPct"),
                    yLabel: t("admin.statistics.figures.axisCsuqOverall"),
                    yDomain: [1, 7] as [number, number],
                    points: scatterSeries.gainVsCsuq,
                    color: "#10b981",
                    file: "paper_figure2a_gain_vs_csuq.svg",
                  },
                  {
                    corrKey: "learning_gain_vs_nasa_tlx_overall",
                    title: t("admin.statistics.figures.scatterGainTlxTitle"),
                    xLabel: t("admin.statistics.figures.axisGainPct"),
                    yLabel: t("admin.statistics.figures.axisTlxWorkload"),
                    yDomain: [0, 100] as [number, number],
                    points: scatterSeries.gainVsTlx,
                    color: "#f59e0b",
                    file: "paper_figure2b_gain_vs_tlx.svg",
                  },
                  {
                    corrKey: "csuq_overall_vs_nasa_tlx_overall",
                    title: t("admin.statistics.figures.scatterCsuqTlxTitle"),
                    xLabel: t("admin.statistics.figures.axisCsuqOverall"),
                    yLabel: t("admin.statistics.figures.axisTlxWorkload"),
                    yDomain: [0, 100] as [number, number],
                    points: scatterSeries.csuqVsTlx,
                    color: "#38bdf8",
                    file: "paper_figure2c_csuq_vs_tlx.svg",
                  },
                ].map((panel) => {
                  const corr = data.correlations.find((c) => c.key === panel.corrKey);
                  const rawP = corr ? fmtPValue(corr.p_value) : "—";
                  const pDisplay = rawP.startsWith("<") ? rawP : `= ${rawP}`;
                  const unlocked = isCorrelationChartUnlocked(panel.points.length);
                  const corrBadgeText =
                    corr && corr.status === "ok" && isFiniteStat(corr.rho)
                      ? t("admin.statistics.figures.rhoBadge", {
                          rho: corr.rho.toFixed(3),
                          p: pDisplay,
                        })
                      : statusGateLabel(corr?.status ?? "insufficient_n", 3, t);

                  return (
                    <ChartExportCard
                      key={panel.corrKey}
                      className="paper-report-card"
                      headingLevel="h3"
                      title={panel.title}
                      exportFilename={panel.file}
                      chartHeightClass="h-64"
                      isEmpty={!unlocked}
                      emptyMessage={t("admin.statistics.figures.smallNScatterGuard", {
                        n: panel.points.length,
                      })}
                      badge={
                        <span
                          className={cn(
                            "inline-block whitespace-nowrap px-2.5 py-0.5 rounded text-xs font-mono font-semibold border",
                            corr?.status === "ok"
                              ? "bg-emerald-500/15 text-emerald-300 border-emerald-500/30"
                              : "bg-amber-500/10 text-amber-300 border-amber-500/30",
                          )}
                        >
                          {corrBadgeText}
                        </span>
                      }
                    >
                      <ResponsiveContainer width="100%" height="100%">
                        <ScatterChart margin={{ top: 12, right: 16, left: -12, bottom: 16 }}>
                          <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                          <XAxis
                            type="number"
                            dataKey="x"
                            name={panel.xLabel}
                            domain={panel.corrKey === "csuq_overall_vs_nasa_tlx_overall" ? [1, 7] : ["auto", "auto"]}
                            tick={{ fill: "#d4d4d8", fontSize: 12 }}
                          />
                          <YAxis
                            type="number"
                            dataKey="y"
                            name={panel.yLabel}
                            domain={panel.yDomain}
                            tick={{ fill: "#a1a1aa", fontSize: 12 }}
                          />
                          <ZAxis
                            dataKey="participantCode"
                            name={t("admin.evaluations.table.colParticipant")}
                            range={[64, 64]}
                          />
                          <Tooltip {...DARK_CHART_TOOLTIP_PROPS} />
                          <Scatter
                            name={t("admin.statistics.figures.participantsLegend")}
                            data={panel.points}
                            fill={panel.color}
                          />
                        </ScatterChart>
                      </ResponsiveContainer>
                    </ChartExportCard>
                  );
                })}
              </div>
            </div>
          </section>

          {/* Section 4: Table III - Learning Gain Inference & Spearman Correlations */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-6">
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3 border-b border-graphite-800 pb-4">
              <div>
                <h2 className="text-lg font-semibold text-white">{t("admin.statistics.table3.title")}</h2>
                <p className="text-sm text-graphite-400">{t("admin.statistics.table3.subtitle")}</p>
              </div>
              {renderCopyActions("table3", buildLatexTable3(data), buildMarkdownTable3(data))}
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-graphite-800 text-xs font-mono text-graphite-400 uppercase">
                    <th className="py-2 pr-4">{t("admin.statistics.cols.metric")}</th>
                    <th className="py-2 px-2">n</th>
                    <th className="py-2 px-2">M (%)</th>
                    <th className="py-2 px-2">SD</th>
                    <th className="py-2 px-2">95% CI</th>
                    <th className="py-2 px-2">Median (IQR)</th>
                    <th className="py-2 px-2">Min–Max</th>
                    <th className="py-2 pl-2">{t("admin.statistics.cols.band")}</th>
                  </tr>
                </thead>
                <tbody>
                  {renderDescriptiveRow(data.learning_gain.pre, t("admin.statistics.constructs.preScore"))}
                  {renderDescriptiveRow(data.learning_gain.post, t("admin.statistics.constructs.postScore"))}
                  {renderDescriptiveRow(data.learning_gain.difference, t("admin.statistics.constructs.gainDelta"))}
                </tbody>
              </table>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="paper-report-card rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-2.5">
                <div className="text-sm font-semibold text-white">
                  {t("admin.statistics.table3.pairedTTitle")}
                </div>
                {data.learning_gain.paired_t_test.status === "ok" ? (
                  <div className="grid grid-cols-2 gap-2 text-sm font-mono">
                    <div className="text-graphite-300">t({data.learning_gain.paired_t_test.df}) = {fmtNum(data.learning_gain.paired_t_test.t_stat, 3)}</div>
                    <div className="text-emerald-400">p = {fmtPValue(data.learning_gain.paired_t_test.p_value)}</div>
                    <div className="text-graphite-300">
                      95% CI Δ: [{fmtNum(data.learning_gain.paired_t_test.ci_low)}, {fmtNum(data.learning_gain.paired_t_test.ci_high)}]
                    </div>
                    <div className="text-white">
                      Cohen&apos;s d_z = {fmtNum(data.learning_gain.paired_t_test.cohens_dz, 3)} ({t(`admin.statistics.bands.${data.learning_gain.paired_t_test.effect_band}`)})
                    </div>
                  </div>
                ) : (
                  <p className="text-sm font-mono text-amber-300">
                    {statusGateLabel(data.learning_gain.paired_t_test.status, 3, t)}
                  </p>
                )}
              </div>

              <div className="paper-report-card rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-2.5">
                <div className="text-sm font-semibold text-white">
                  {t("admin.statistics.table3.wilcoxonTitle")}
                </div>
                {data.learning_gain.wilcoxon.status === "ok" ? (
                  <div className="grid grid-cols-2 gap-2 text-sm font-mono">
                    <div className="text-graphite-300">
                      W = {fmtNum(data.learning_gain.wilcoxon.w_stat, 1)} (n_eff={data.learning_gain.wilcoxon.effective_n})
                    </div>
                    <div className="text-emerald-400">
                      p = {fmtPValue(data.learning_gain.wilcoxon.p_value)} ({data.learning_gain.wilcoxon.method})
                    </div>
                    <div className="text-graphite-300">
                      W+ = {fmtNum(data.learning_gain.wilcoxon.w_plus, 1)} / W- = {fmtNum(data.learning_gain.wilcoxon.w_minus, 1)}
                    </div>
                    <div className="text-white">
                      r_rb = {fmtNum(data.learning_gain.wilcoxon.rank_biserial, 3)}
                    </div>
                  </div>
                ) : (
                  <p className="text-sm font-mono text-amber-300">
                    {statusGateLabel(data.learning_gain.wilcoxon.status, 5, t)}
                  </p>
                )}
              </div>
            </div>

            <div className="space-y-3">
              <h3 className="text-base font-semibold text-white">
                {t("admin.statistics.table3.correlationsTitle")}
              </h3>
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-graphite-800 text-xs font-mono text-graphite-400 uppercase">
                      <th className="py-2 pr-4">{t("admin.statistics.cols.pair")}</th>
                      <th className="py-2 px-2">n</th>
                      <th className="py-2 px-2">Spearman ρ</th>
                      <th className="py-2 px-2">p (two-tailed)</th>
                      <th className="py-2 pl-2">{t("admin.statistics.cols.strength")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.correlations.map((corr) => (
                      <tr key={corr.key} className="border-b border-graphite-800/70 text-sm">
                        <td className="py-3 pr-4 font-medium text-graphite-100">
                          {t(`admin.statistics.correlationPairs.${corr.key}`)}
                        </td>
                        <td className="py-3 px-2 font-mono text-graphite-300">{corr.n}</td>
                        <td className="py-3 px-2 font-mono font-semibold text-white">
                          {corr.status === "ok" ? fmtNum(corr.rho, 3) : statusGateLabel(corr.status, 3, t)}
                        </td>
                        <td className="py-3 px-2 font-mono text-emerald-400">
                          {corr.status === "ok" ? fmtPValue(corr.p_value) : "—"}
                        </td>
                        <td className="py-3 pl-2 text-xs font-mono">
                          <span className="px-2 py-0.5 rounded border border-graphite-700 bg-graphite-950 text-graphite-300">
                            {t(`admin.statistics.bands.${corr.band}`)}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </section>

          {/* Section 5: Collapsible Bilingual Statistical Methodology & Formula Guide */}
          <section className="paper-report-section bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="space-y-1">
                <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                  <BookOpen className="w-5 h-5 text-emerald-400" />
                  {t("admin.statistics.methodology.title")}
                </h2>
                <p className="text-sm text-graphite-400">
                  {t("admin.statistics.methodology.subtitle")}
                </p>
              </div>
              <button
                type="button"
                onClick={() => setShowMethodology((prev) => !prev)}
                className="print:hidden inline-flex items-center gap-1.5 px-3.5 py-2 rounded-md border border-graphite-700 bg-graphite-950 text-xs font-mono font-semibold text-graphite-200 hover:text-white hover:border-white transition-colors cursor-pointer"
              >
                <span>
                  {showMethodology
                    ? t("admin.statistics.methodology.collapseAction")
                    : t("admin.statistics.methodology.expandAction")}
                </span>
                {showMethodology ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
              </button>
            </div>

            {showMethodology && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-3 border-t border-graphite-800">
                {METHODOLOGY_PILLARS.map((pillar) => {
                  const copy = locale === "th" ? pillar.th : pillar.en;
                  return (
                    <div
                      key={pillar.id}
                      className="paper-report-card rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-2"
                    >
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <h3 className="text-base font-semibold text-white">{copy.title}</h3>
                        {pillar.citation && (
                          <span className="px-2 py-0.5 rounded text-xs font-mono bg-emerald-500/10 text-emerald-300 border border-emerald-500/30">
                            {pillar.citation}
                          </span>
                        )}
                      </div>
                      <div className="px-3 py-1.5 rounded bg-graphite-900 border border-graphite-800 font-mono text-xs text-amber-300 overflow-x-auto">
                        {copy.formula}
                      </div>
                      <p className="text-sm text-graphite-300 leading-relaxed">{copy.body}</p>
                    </div>
                  );
                })}
              </div>
            )}
          </section>
        </>
      )}
    </div>
  );
}
