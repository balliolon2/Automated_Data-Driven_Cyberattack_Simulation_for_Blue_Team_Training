import { BookOpen, ShieldAlert, ShieldCheck } from "lucide-react";
import { Link } from "react-router-dom";
import { cn } from "../lib/utils";
import { useI18n } from "../i18n/index.tsx";
import { resolveDomainName } from "../i18n/format.ts";

interface DomainStatus {
  domain_id: string;
  domain_name: string;
  proficiency_score: number;
  threshold: number;
  passed: boolean;
  score_meta?: string;
}

interface DomainBreakdownProps {
  domains: DomainStatus[];
  title?: string;
  showThresholdLine?: boolean;
  studyDomainHref?: (domainId: string) => string;
}

export default function DomainBreakdown({ domains, title, showThresholdLine = true, studyDomainHref }: DomainBreakdownProps) {
  const { locale, t } = useI18n();
  const resolvedTitle = title ?? t("evaluation.domainBreakdownDefaultTitle");

  return (
    <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-6 space-y-6">
      <h3 className="text-lg font-semibold tracking-tight text-white border-b border-graphite-800 pb-3">
        {resolvedTitle}
      </h3>
      <div className="space-y-4">
        {domains.map((domain) => {
          const percentage = Math.min(100, Math.max(0, domain.proficiency_score));
          const thresholdPercent = domain.threshold;
          const domainName = resolveDomainName(domain.domain_id, domain.domain_name, locale);
          
          return (
            <div key={domain.domain_id} className="space-y-2 group">
              <div className="flex items-start justify-between gap-3 text-sm">
                <div className="flex items-start gap-2 min-w-0 flex-1">
                  {domain.passed ? (
                    <ShieldCheck className="w-4 h-4 text-emerald-500 shrink-0 mt-0.5" />
                  ) : (
                    <ShieldAlert className="w-4 h-4 text-amber-500 shrink-0 mt-0.5" />
                  )}
                  <span
                    className="font-medium text-graphite-200 leading-snug break-words group-hover:text-white transition-colors duration-200"
                    title={domainName}
                  >
                    {domainName}
                  </span>
                </div>
                <div className="flex items-center gap-3 font-mono shrink-0">
                  {studyDomainHref && (
                    <Link
                      to={studyDomainHref(domain.domain_id)}
                      className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded border border-graphite-700 bg-graphite-800/80 hover:bg-graphite-700 text-graphite-200 hover:text-white text-xs font-sans transition-colors whitespace-nowrap"
                    >
                      <BookOpen className="w-3.5 h-3.5" />
                      {t("evaluation.studyTheory")}
                    </Link>
                  )}
                  <div className="flex items-baseline gap-1.5">
                    {domain.score_meta && (
                      <span className="text-xs text-graphite-400 mr-1">
                        ({domain.score_meta})
                      </span>
                    )}
                    <span className={cn(
                      "text-base font-bold",
                      domain.passed ? "text-emerald-400" : "text-amber-400"
                    )}>
                      {domain.proficiency_score.toFixed(1)}%
                    </span>
                    <span className="text-xs text-graphite-400">/ {domain.threshold}%</span>
                  </div>
                </div>
              </div>
              
              <div className="relative h-2 w-full bg-graphite-950 rounded overflow-hidden border border-graphite-800/80">
                {/* Threshold Marker Line */}
                {showThresholdLine && (
                  <div 
                    className="absolute top-0 bottom-0 w-0.5 bg-graphite-800 z-10"
                    style={{ left: `${thresholdPercent}%` }}
                    title={t("evaluation.thresholdTooltip", { threshold: thresholdPercent })}
                  />
                )}
                
                {/* Progress Bar (Flat) */}
                <div 
                  className={cn(
                    "h-full rounded-sm transition-all duration-1000 ease-out",
                    domain.passed ? "bg-emerald-500" : "bg-amber-500"
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
