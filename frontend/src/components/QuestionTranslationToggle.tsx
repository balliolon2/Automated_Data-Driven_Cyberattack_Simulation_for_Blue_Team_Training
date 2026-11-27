import { Languages } from "lucide-react";
import { useI18n } from "../i18n/index.tsx";
import { cn } from "../lib/utils";

interface QuestionTranslationToggleProps {
  visible: boolean;
  showingOriginalEn: boolean;
  onToggle: () => void;
  className?: string;
}

/** Compact dark-mode button allowing Thai-locale learners to inspect canonical English question text. */
export default function QuestionTranslationToggle({
  visible,
  showingOriginalEn,
  onToggle,
  className,
}: QuestionTranslationToggleProps) {
  const { t } = useI18n();
  if (!visible) return null;

  return (
    <button
      type="button"
      onClick={onToggle}
      aria-pressed={showingOriginalEn}
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded text-xs font-mono border transition-colors cursor-pointer shrink-0",
        showingOriginalEn
          ? "bg-indigo-500/20 border-indigo-400/50 text-indigo-200"
          : "bg-graphite-950 border-graphite-700 text-graphite-200 hover:border-graphite-500 hover:text-white",
        className,
      )}
    >
      <Languages className="w-3.5 h-3.5" />
      {showingOriginalEn ? t("exam.showThaiTranslation") : t("exam.showOriginalEn")}
    </button>
  );
}
