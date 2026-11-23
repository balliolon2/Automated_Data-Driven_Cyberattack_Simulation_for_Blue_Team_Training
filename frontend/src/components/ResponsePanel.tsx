import { Shield, Sparkles, RefreshCw, AlertTriangle, CheckCircle, AlertOctagon } from "lucide-react";
import { useI18n } from "../i18n";
import { cn } from "../lib/utils";

interface ActionOption {
  id: string;
  label: string;
}

interface PlaybookSteps {
  containment?: ActionOption[];
  eradication?: ActionOption[];
  recovery?: ActionOption[];
}

interface ResponsePanelProps {
  playbookSteps?: PlaybookSteps;
  selectedActions: string[];
  onChange: (selected: string[]) => void;
  isTruePositive?: boolean | null;
  onSelectTriage?: (val: boolean) => void;
}

export default function ResponsePanel({ 
  playbookSteps, 
  selectedActions, 
  onChange,
  isTruePositive,
  onSelectTriage 
}: ResponsePanelProps) {
  const { t } = useI18n();

  const handleToggleAction = (actionId: string) => {
    const nextSelected = selectedActions.includes(actionId)
      ? selectedActions.filter((id) => id !== actionId)
      : [...selectedActions, actionId];
    onChange(nextSelected);
  };

  const renderSection = (title: string, actions: ActionOption[] | undefined, icon: React.ReactNode, themeColor: string) => {
    if (!actions || actions.length === 0) return null;

    return (
      <div className="glass-panel p-5 rounded-xl border border-graphite-800 space-y-4">
        <div className="flex items-center gap-2 pb-2 border-b border-graphite-800/80">
          <div className={cn("p-1.5 rounded-lg bg-opacity-10", themeColor)}>
            {icon}
          </div>
          <h4 className="font-semibold text-base tracking-wide uppercase text-graphite-100">
            {t("responsePanel.phaseSuffix", { title })}
          </h4>
        </div>
        
        <div className="space-y-3">
          {actions.map((act) => {
            const isChecked = selectedActions.includes(act.id);
            return (
              <label 
                key={act.id}
                className={cn(
                  "flex items-start gap-3 p-4 rounded-lg border cursor-pointer select-none transition-all duration-200",
                  isChecked 
                    ? "bg-emerald-950/15 border-emerald-500/30 text-emerald-100 shadow-[0_0_8px_rgba(52,211,153,0.05)]" 
                    : "bg-graphite-950/40 border-graphite-800/60 text-graphite-200 hover:bg-graphite-800/20 hover:border-graphite-700/60"
                )}
              >
                <input
                  type="checkbox"
                  checked={isChecked}
                  onChange={() => handleToggleAction(act.id)}
                  className="mt-1 w-4 h-4 rounded text-emerald-500 bg-graphite-950 border-graphite-800 focus:ring-emerald-500/40 focus:ring-offset-graphite-900 focus:ring-2 cursor-pointer accent-emerald-500 shrink-0"
                />
                <span className="text-sm leading-relaxed font-medium">
                  {act.label}
                </span>
              </label>
            );
          })}
        </div>
      </div>
    );
  };

  const containmentActions = playbookSteps?.containment || [];
  const eradicationActions = playbookSteps?.eradication || [];
  const recoveryActions = playbookSteps?.recovery || [];

  return (
    <div className="space-y-6">
      {/* Triage Classification Selection */}
      <div className="glass-panel p-5 rounded-xl border border-graphite-800 space-y-4">
        <div className="flex items-center gap-2 pb-2 border-b border-graphite-800/80">
          <div className="p-1.5 rounded-lg bg-opacity-10 bg-amber-400 text-amber-400">
            <AlertOctagon className="w-4 h-4" />
          </div>
          <h4 className="font-semibold text-base tracking-wide uppercase text-graphite-100">
            {t("responsePanel.phase1Title")}
          </h4>
        </div>

        <p className="text-sm text-graphite-300 leading-relaxed font-normal">
          {t("responsePanel.phase1Prompt")}
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-1">
          <button
            type="button"
            onClick={() => onSelectTriage && onSelectTriage(true)}
            className={cn(
              "p-4 rounded-lg border text-left flex items-center justify-between font-mono text-sm transition-all duration-200 cursor-pointer",
              isTruePositive === true
                ? "bg-red-950/30 border-red-500/50 text-red-200 font-semibold shadow-[0_0_12px_rgba(239,68,68,0.15)]"
                : "bg-graphite-950/40 border-graphite-800/60 text-graphite-200 hover:bg-graphite-800/20"
            )}
          >
            <div className="flex items-center gap-2">
              <AlertTriangle className={cn("w-4 h-4 shrink-0", isTruePositive === true ? "text-red-400" : "text-graphite-400")} />
              <span>{t("responsePanel.truePositive")}</span>
            </div>
            <span className={cn("w-3 h-3 rounded-full border border-graphite-600 shrink-0", isTruePositive === true && "bg-red-500 border-red-400")} />
          </button>

          <button
            type="button"
            onClick={() => onSelectTriage && onSelectTriage(false)}
            className={cn(
              "p-4 rounded-lg border text-left flex items-center justify-between font-mono text-sm transition-all duration-200 cursor-pointer",
              isTruePositive === false
                ? "bg-emerald-950/30 border-emerald-500/50 text-emerald-200 font-semibold shadow-[0_0_12px_rgba(16,185,129,0.15)]"
                : "bg-graphite-950/40 border-graphite-800/60 text-graphite-200 hover:bg-graphite-800/20"
            )}
          >
            <div className="flex items-center gap-2">
              <CheckCircle className={cn("w-4 h-4 shrink-0", isTruePositive === false ? "text-emerald-400" : "text-graphite-400")} />
              <span>{t("responsePanel.falsePositive")}</span>
            </div>
            <span className={cn("w-3 h-3 rounded-full border border-graphite-600 shrink-0", isTruePositive === false && "bg-emerald-500 border-emerald-400")} />
          </button>
        </div>
      </div>

      {/* Response Action Steps */}
      {renderSection(
        t("responsePanel.phase2Title"),
        containmentActions,
        <Shield className="w-4 h-4 text-amber-400" />,
        "bg-amber-400 text-amber-400"
      )}

      {renderSection(
        t("responsePanel.phase3Title"),
        eradicationActions,
        <Sparkles className="w-4 h-4 text-emerald-400" />,
        "bg-emerald-400 text-emerald-400"
      )}

      {renderSection(
        t("responsePanel.phase4Title"),
        recoveryActions,
        <RefreshCw className="w-4 h-4 text-sky-400" />,
        "bg-sky-400 text-sky-400"
      )}
    </div>
  );
}
