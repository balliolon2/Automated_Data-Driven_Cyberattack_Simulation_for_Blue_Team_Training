import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import ChartExportCard from "./ChartExportCard";
import { useI18n } from "../i18n/index.tsx";
import { DARK_CHART_TOOLTIP_PROPS, type PrePostPairedSeries } from "../lib/evaluationCharts";

interface PairedPrePostChartCardProps {
  title: string;
  subtitle: string;
  exportFilename: string;
  emptyMessage: string;
  series: PrePostPairedSeries;
  headingLevel?: "h2" | "h3";
  className?: string;
}

/** Shared chronological Pre-Test vs. Post-Test paired bar chart with labeled cohort mean reference lines. */
export default function PairedPrePostChartCard({
  title,
  subtitle,
  exportFilename,
  emptyMessage,
  series,
  headingLevel = "h2",
  className,
}: PairedPrePostChartCardProps) {
  const { t } = useI18n();

  return (
    <ChartExportCard
      className={className}
      headingLevel={headingLevel}
      title={title}
      subtitle={subtitle}
      exportFilename={exportFilename}
      isEmpty={series.rows.length === 0}
      emptyMessage={emptyMessage}
      chartHeightClass="h-72"
      minChartWidthPx={series.rows.length > 16 ? Math.max(580, series.rows.length * 48) : undefined}
    >
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={series.rows} margin={{ top: 18, right: 16, left: -12, bottom: 8 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
          <XAxis dataKey="label" tick={{ fill: "#d4d4d8", fontSize: 12 }} />
          <YAxis domain={[0, 100]} ticks={[0, 25, 50, 75, 100]} tick={{ fill: "#a1a1aa", fontSize: 12 }} />
          <Tooltip {...DARK_CHART_TOOLTIP_PROPS} />
          <Legend wrapperStyle={{ fontSize: 12 }} />
          {series.preMean !== null && (
            <ReferenceLine
              y={series.preMean}
              stroke="#94a3b8"
              strokeDasharray="4 4"
              label={{
                value: t("admin.evaluations.charts.pairedPreMean", { mean: series.preMean.toFixed(1) }),
                position: "insideBottomLeft",
                fill: "#cbd5e1",
                fontSize: 12,
              }}
            />
          )}
          {series.postMean !== null && (
            <ReferenceLine
              y={series.postMean}
              stroke="#10b981"
              strokeDasharray="4 4"
              label={{
                value: t("admin.evaluations.charts.pairedPostMean", { mean: series.postMean.toFixed(1) }),
                position: "insideTopRight",
                fill: "#6ee7b7",
                fontSize: 12,
              }}
            />
          )}
          <Bar
            dataKey="pre"
            name={t("admin.evaluations.charts.pairedPreBar")}
            fill="#64748b"
            radius={[3, 3, 0, 0]}
          />
          <Bar
            dataKey="post"
            name={t("admin.evaluations.charts.pairedPostBar")}
            fill="#10b981"
            radius={[3, 3, 0, 0]}
          />
        </BarChart>
      </ResponsiveContainer>
    </ChartExportCard>
  );
}
