import { Link } from "react-router-dom";
import { ArrowRight, CheckCircle2, ClipboardList, Sparkles } from "lucide-react";
import { useI18n } from "../i18n/index.tsx";
import type { QuestionnaireStatusResponse } from "../lib/questionnaire";
import { cn } from "../lib/utils";

interface QuestionnairePromptBannerProps {
  status: QuestionnaireStatusResponse | null;
  variant?: "prominent" | "compact";
  className?: string;
}

export default function QuestionnairePromptBanner({
  status,
  variant = "compact",
  className,
}: QuestionnairePromptBannerProps) {
  const { t } = useI18n();

  if (!status || status.is_preview_only) {
    return null;
  }

  if (status.submitted) {
    return (
      <div
        className={cn(
          "rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4",
          className,
        )}
      >
        <div className="flex items-start gap-3">
          <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0 mt-0.5" />
          <div className="space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-base font-semibold text-white">
                {t("questionnaire.banner.submittedTitle")}
              </h3>
              <span className="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                {t("questionnaire.banner.submittedBadge")}
              </span>
            </div>
            <p className="text-sm text-graphite-300 leading-relaxed">
              {t("questionnaire.banner.submittedDesc")}
            </p>
          </div>
        </div>

        <Link
          to="/questionnaire"
          className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-md border border-emerald-500/40 bg-graphite-950 hover:bg-graphite-900 text-sm font-medium text-emerald-300 transition-colors whitespace-nowrap shrink-0"
        >
          <span>{t("questionnaire.banner.viewSummaryAction")}</span>
          <ArrowRight className="w-4 h-4" />
        </Link>
      </div>
    );
  }

  if (!status.eligible) {
    return null;
  }

  const isProminent = variant === "prominent";

  return (
    <div
      className={cn(
        "rounded-lg border p-5 sm:p-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4",
        isProminent
          ? "border-emerald-500/40 bg-gradient-to-r from-emerald-500/15 via-graphite-900 to-graphite-900"
          : "border-graphite-700 bg-graphite-900",
        className,
      )}
    >
      <div className="flex items-start gap-3.5">
        <div className="p-2.5 rounded-md bg-emerald-500/15 border border-emerald-500/30 text-emerald-400 shrink-0">
          {isProminent ? <Sparkles className="w-5 h-5" /> : <ClipboardList className="w-5 h-5" />}
        </div>
        <div className="space-y-1">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="text-base font-semibold text-white">
              {t("questionnaire.banner.pendingTitle")}
            </h3>
            <span className="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-amber-500/15 text-amber-300 border border-amber-500/30">
              {t("questionnaire.banner.durationBadge")}
            </span>
          </div>
          <p className="text-sm sm:text-base text-graphite-300 leading-relaxed">
            {t("questionnaire.banner.pendingDesc")}
          </p>
        </div>
      </div>

      <Link
        to="/questionnaire"
        className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-md bg-white text-black font-semibold text-sm hover:bg-white/90 transition-colors whitespace-nowrap shrink-0"
      >
        <span>{t("questionnaire.banner.startAction")}</span>
        <ArrowRight className="w-4 h-4" />
      </Link>
    </div>
  );
}
