import { ShieldAlert, ShieldCheck } from "lucide-react";
import { cn } from "../lib/utils";

interface DomainStatus {
  domain_id: string;
  domain_name: string;
  proficiency_score: number;
  threshold: number;
  passed: boolean;
}

interface DomainBreakdownProps {
  domains: DomainStatus[];
  title?: string;
  showThresholdLine?: boolean;
}

export default function DomainBreakdown({ domains, title = "Domain Proficiency Breakdown", showThresholdLine = true }: DomainBreakdownProps) {
  return (
    <div className="glass-panel p-6 rounded-xl space-y-6">
      <h3 className="text-lg font-semibold tracking-wide text-graphite-100 border-b border-graphite-800 pb-3">
        {title}
      </h3>
      <div className="space-y-4">
        {domains.map((domain) => {
          const percentage = Math.min(100, Math.max(0, domain.proficiency_score));
          const thresholdPercent = domain.threshold;
          
          return (
            <div key={domain.domain_id} className="space-y-2 group">
              <div className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2 max-w-[75%]">
                  {domain.passed ? (
                    <ShieldCheck className="w-4 h-4 text-emerald-400 shrink-0 filter drop-shadow-[0_0_4px_rgba(52,211,153,0.4)]" />
                  ) : (
                    <ShieldAlert className="w-4 h-4 text-amber-500 shrink-0 filter drop-shadow-[0_0_4px_rgba(245,158,11,0.4)]" />
                  )}
                  <span className="font-medium text-graphite-200 truncate group-hover:text-white transition-colors duration-200">
                    {domain.domain_name}
                  </span>
                </div>
                <div className="flex items-baseline gap-1.5 font-mono">
                  <span className={cn(
                    "text-base font-bold",
                    domain.passed ? "text-emerald-400" : "text-amber-400"
                  )}>
                    {domain.proficiency_score.toFixed(1)}%
                  </span>
                  <span className="text-xs text-graphite-500">/ {domain.threshold}%</span>
                </div>
              </div>
              
              <div className="relative h-3 w-full bg-graphite-950 rounded-full overflow-hidden border border-graphite-800/80">
                {/* Threshold Marker Line */}
                {showThresholdLine && (
                  <div 
                    className="absolute top-0 bottom-0 w-0.5 bg-graphite-600/60 z-10"
                    style={{ left: `${thresholdPercent}%` }}
                    title={`Threshold: ${thresholdPercent}%`}
                  />
                )}
                
                {/* Progress Bar with Glow */}
                <div 
                  className={cn(
                    "h-full rounded-full transition-all duration-1000 ease-out relative",
                    domain.passed 
                      ? "bg-gradient-to-r from-emerald-600 to-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.3)]" 
                      : "bg-gradient-to-r from-amber-600 to-amber-400 shadow-[0_0_8px_rgba(245,158,11,0.3)]"
                  )}
                  style={{ width: `${percentage}%` }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
