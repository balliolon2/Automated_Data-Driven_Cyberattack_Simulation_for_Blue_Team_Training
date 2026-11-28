import {
  Legend,
  PolarAngleAxis,
  PolarGrid,
  PolarRadiusAxis,
  Radar,
  RadarChart,
  ResponsiveContainer,
  Tooltip,
} from "recharts";
import ChartExportCard from "./ChartExportCard";
import { DARK_CHART_TOOLTIP_PROPS, type TLXRadarDatum } from "../lib/evaluationCharts";

interface TLXWorkloadRadarCardProps {
  title: string;
  subtitle: string;
  exportFilename: string;
  data: TLXRadarDatum[];
  seriesName: string;
  midpointName: string;
  isEmpty?: boolean;
  emptyMessage?: string;
  chartHeightClass?: string;
  headingLevel?: "h2" | "h3";
}

/** Shared 6-axis Raw NASA-TLX Radar chart card spanning 0..100 with midpoint reference overlay. */
export default function TLXWorkloadRadarCard({
  title,
  subtitle,
  exportFilename,
  data,
  seriesName,
  midpointName,
  isEmpty = false,
  emptyMessage,
  chartHeightClass = "h-72",
  headingLevel = "h2",
}: TLXWorkloadRadarCardProps) {
  return (
    <ChartExportCard
      headingLevel={headingLevel}
      title={title}
      subtitle={subtitle}
      exportFilename={exportFilename}
      isEmpty={isEmpty}
      emptyMessage={emptyMessage}
      chartHeightClass={chartHeightClass}
    >
      <ResponsiveContainer width="100%" height="100%">
        <RadarChart data={data} cx="50%" cy="50%" outerRadius="64%">
          <PolarGrid stroke="#27272a" />
          <PolarAngleAxis dataKey="axis" tick={{ fill: "#d4d4d8", fontSize: 12 }} />
          <PolarRadiusAxis angle={30} domain={[0, 100]} tick={{ fill: "#a1a1aa", fontSize: 12 }} />
          <Radar
            name={midpointName}
            dataKey="benchmark"
            stroke="#64748b"
            strokeDasharray="4 4"
            fill="transparent"
          />
          <Radar
            name={seriesName}
            dataKey="workload"
            stroke="#f59e0b"
            fill="#f59e0b"
            fillOpacity={0.28}
          />
          <Legend wrapperStyle={{ fontSize: 12 }} />
          <Tooltip {...DARK_CHART_TOOLTIP_PROPS} />
        </RadarChart>
      </ResponsiveContainer>
    </ChartExportCard>
  );
}
