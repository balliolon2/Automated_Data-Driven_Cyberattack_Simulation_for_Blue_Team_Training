import { useRef, type ReactNode } from "react";
import { Download } from "lucide-react";
import { useI18n } from "../i18n/index.tsx";
import { exportSvgElement } from "../lib/svgExport";
import { cn } from "../lib/utils";

interface ChartExportCardProps {
  title: string;
  subtitle?: string;
  badge?: ReactNode;
  headerExtra?: ReactNode;
  exportFilename?: string;
  isEmpty?: boolean;
  emptyMessage?: string;
  chartHeightClass?: string;
  minChartWidthPx?: number;
  headingLevel?: "h2" | "h3";
  className?: string;
  children: ReactNode;
}

/** Reusable dark-mode Recharts card with optional toggle header and 1-click standalone .svg download. */
export default function ChartExportCard({
  title,
  subtitle,
  badge,
  headerExtra,
  exportFilename,
  isEmpty = false,
  emptyMessage,
  chartHeightClass = "h-72",
  minChartWidthPx,
  headingLevel = "h2",
  className,
  children,
}: ChartExportCardProps) {
  const { t } = useI18n();
  const containerRef = useRef<HTMLDivElement | null>(null);
  const HeadingTag = headingLevel;
  const headingClass =
    headingLevel === "h2"
      ? "text-xl font-semibold text-white leading-snug"
      : "text-lg font-semibold text-white leading-snug";

  return (
    <div
      className={cn(
        "bg-graphite-900 border border-graphite-800 rounded-lg p-5 sm:p-6 space-y-4",
        className,
      )}
    >
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
        <div className="space-y-1.5 min-w-0">
          <HeadingTag className={headingClass}>{title}</HeadingTag>
          {subtitle && (
            <p className="text-xs font-mono text-graphite-400 leading-relaxed">{subtitle}</p>
          )}
          {badge && <div className="flex flex-wrap items-center gap-2 pt-0.5">{badge}</div>}
        </div>

        <div className="print:hidden flex flex-wrap items-center gap-2 shrink-0">
          {headerExtra}
          {exportFilename && (
            <button
              type="button"
              disabled={isEmpty}
              onClick={() => exportSvgElement(containerRef.current, exportFilename, title)}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-sm font-medium text-graphite-200 hover:text-white hover:border-white disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
            >
              <Download className="w-4 h-4" />
              <span>{t("admin.evaluations.charts.downloadSvg")}</span>
            </button>
          )}
        </div>
      </div>

      {isEmpty ? (
        <div
          className={cn(
            chartHeightClass,
            "w-full rounded-md border border-dashed border-graphite-800 bg-graphite-950/50 flex items-center justify-center p-6 text-center",
          )}
        >
          <p className="text-sm text-graphite-400 max-w-md leading-relaxed">
            {emptyMessage || t("admin.evaluations.charts.emptyChartFallback")}
          </p>
        </div>
      ) : (
        <div
          ref={containerRef}
          className={cn(minChartWidthPx ? "overflow-x-auto pb-1" : "w-full")}
        >
          <div
            className={cn(chartHeightClass, "w-full")}
            style={minChartWidthPx ? { minWidth: `${minChartWidthPx}px` } : undefined}
          >
            {children}
          </div>
        </div>
      )}
    </div>
  );
}
