import { Shield, Sparkles, RefreshCw } from "lucide-react";
import { cn } from "../lib/utils";

interface ActionOption {
  id: string;
  label: string;
}

interface PlaybookSteps {
  containment: ActionOption[];
  eradication: ActionOption[];
  recovery: ActionOption[];
}

interface ResponsePanelProps {
  playbookSteps: PlaybookSteps;
  selectedActions: string[];
  onChange: (selected: string[]) => void;
}

export default function ResponsePanel({ playbookSteps, selectedActions, onChange }: ResponsePanelProps) {
  const handleToggleAction = (actionId: string) => {
    const nextSelected = selectedActions.includes(actionId)
      ? selectedActions.filter((id) => id !== actionId)
      : [...selectedActions, actionId];
    onChange(nextSelected);
  };

  const renderSection = (title: string, actions: ActionOption[], icon: React.ReactNode, themeColor: string) => {
    if (!actions || actions.length === 0) return null;

    return (
      <div className="glass-panel p-5 rounded-xl border border-graphite-800 space-y-4">
        <div className="flex items-center gap-2 pb-2 border-b border-graphite-800/80">
          <div className={cn("p-1.5 rounded-lg bg-opacity-10", themeColor)}>
            {icon}
          </div>
          <h4 className="font-semibold text-sm tracking-wider uppercase text-graphite-200">
            {title} Phase
          </h4>
        </div>
        
        <div className="space-y-3">
          {actions.map((act) => {
            const isChecked = selectedActions.includes(act.id);
            return (
              <label 
                key={act.id}
                className={cn(
                  "flex items-start gap-3 p-3.5 rounded-lg border cursor-pointer select-none transition-all duration-200",
                  isChecked 
                    ? "bg-emerald-950/15 border-emerald-500/30 text-emerald-100 shadow-[0_0_8px_rgba(52,211,153,0.05)]" 
                    : "bg-graphite-950/40 border-graphite-800/60 text-graphite-300 hover:bg-graphite-800/20 hover:border-graphite-700/60"
                )}
              >
                <input
                  type="checkbox"
                  checked={isChecked}
                  onChange={() => handleToggleAction(act.id)}
                  className="mt-0.5 w-4 h-4 rounded text-emerald-500 bg-graphite-950 border-graphite-800 focus:ring-emerald-500/40 focus:ring-offset-graphite-900 focus:ring-2 cursor-pointer accent-emerald-500"
                />
                <span className="text-xs leading-relaxed font-medium">
                  {act.label}
                </span>
              </label>
            );
          })}
        </div>
      </div>
    );
  };

  return (
    <div className="space-y-6">
      {renderSection(
        "Containment",
        playbookSteps.containment,
        <Shield className="w-4 h-4 text-amber-400" />,
        "bg-amber-400 text-amber-400"
      )}

      {renderSection(
        "Eradication",
        playbookSteps.eradication,
        <Sparkles className="w-4 h-4 text-emerald-400" />,
        "bg-emerald-400 text-emerald-400"
      )}

      {renderSection(
        "Recovery",
        playbookSteps.recovery,
        <RefreshCw className="w-4 h-4 text-sky-400" />,
        "bg-sky-400 text-sky-400"
      )}
    </div>
  );
}
