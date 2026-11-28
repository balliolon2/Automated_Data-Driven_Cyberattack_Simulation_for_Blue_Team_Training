import { useState } from "react";
import { BookOpen, ChevronDown, ChevronUp, Compass, ShieldCheck } from "lucide-react";
import { useI18n } from "../i18n";
import {
  getBeginnerGlossaryTerms,
  getSimulationGuideDismissStorageKey,
  shouldAutoExpandBeginnerGuide,
} from "../lib/simulationGuide";
import { cn } from "../lib/utils";

interface SimulationBeginnerGuideProps {
  completedScenarios: number | null;
  userId?: string | null;
}

function readGuideDismissedFlag(storageKey: string): boolean {
  if (typeof window === "undefined") return false;
  try {
    return window.localStorage.getItem(storageKey) === "true";
  } catch {
    return false;
  }
}

/** Static, zero-spoiler collapsible beginner tutorial & 10-term SOC glossary above Simulation tabs. */
export default function SimulationBeginnerGuide({
  completedScenarios,
  userId,
}: SimulationBeginnerGuideProps) {
  const { locale, t } = useI18n();
  const [activeGuideTab, setActiveGuideTab] = useState<"steps" | "glossary">("steps");
  const [manualOverride, setManualOverride] = useState<{
    key: string;
    open: boolean;
  } | null>(null);

  const storageKey = getSimulationGuideDismissStorageKey(userId);
  const isDismissed = readGuideDismissedFlag(storageKey);
  const isOpen =
    manualOverride?.key === storageKey
      ? manualOverride.open
      : shouldAutoExpandBeginnerGuide(completedScenarios, isDismissed);

  const handleToggle = (nextOpen: boolean) => {
    setManualOverride({ key: storageKey, open: nextOpen });
    try {
      if (!nextOpen && completedScenarios === 0) {
        window.localStorage.setItem(storageKey, "true");
      } else if (nextOpen && completedScenarios === 0) {
        window.localStorage.removeItem(storageKey);
      }
    } catch {
      // Ignore storage errors
    }
  };

  if (!isOpen) {
    return (
      <div className="bg-graphite-900/70 border border-graphite-800 rounded-lg px-4 py-2.5 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5 text-sm text-graphite-300">
          <Compass className="w-4 h-4 text-emerald-400 shrink-0" />
          <span>{t("simulation.beginnerGuide.collapsedBarTitle")}</span>
        </div>
        <button
          type="button"
          onClick={() => handleToggle(true)}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-950 text-sm font-medium text-emerald-300 hover:text-white hover:border-emerald-400 transition-colors cursor-pointer"
        >
          <BookOpen className="w-4 h-4 shrink-0" />
          <span>{t("simulation.beginnerGuide.openAction")}</span>
          <ChevronDown className="w-4 h-4 shrink-0" />
        </button>
      </div>
    );
  }

  const glossaryTerms = getBeginnerGlossaryTerms(locale);

  return (
    <div className="bg-graphite-900 border border-emerald-500/30 rounded-lg overflow-hidden shadow-lg animate-fade-in">
      <div className="bg-graphite-950 border-b border-graphite-800 px-5 py-3.5 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-2.5">
          <Compass className="w-5 h-5 text-emerald-400 shrink-0" />
          <h2 className="text-xl font-semibold text-white leading-snug">
            {t("simulation.beginnerGuide.expandedTitle")}
          </h2>
          <span className="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-emerald-500/15 text-emerald-300 border border-emerald-500/30">
            {t("simulation.beginnerGuide.expandedBadge")}
          </span>
        </div>

        <button
          type="button"
          onClick={() => handleToggle(false)}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-700 bg-graphite-900 text-sm font-medium text-graphite-300 hover:text-white hover:border-white transition-colors cursor-pointer"
        >
          <span>{t("simulation.beginnerGuide.collapseAction")}</span>
          <ChevronUp className="w-4 h-4 shrink-0" />
        </button>
      </div>

      <div className="border-b border-graphite-800 px-5 pt-3 flex flex-wrap gap-2 bg-graphite-900/60">
        {([
          { id: "steps" as const, label: t("simulation.beginnerGuide.tabHowToPlay") },
          { id: "glossary" as const, label: t("simulation.beginnerGuide.tabGlossary") },
        ]).map((tab) => (
          <button
            key={tab.id}
            type="button"
            onClick={() => setActiveGuideTab(tab.id)}
            className={cn(
              "px-3.5 py-2 text-sm font-medium rounded-t-md border-b-2 transition-colors cursor-pointer",
              activeGuideTab === tab.id
                ? "border-emerald-400 text-white bg-graphite-900"
                : "border-transparent text-graphite-400 hover:text-graphite-200",
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="p-5 space-y-4">
        {activeGuideTab === "steps" ? (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {[
              {
                title: t("simulation.beginnerGuide.step1Title"),
                desc: t("simulation.beginnerGuide.step1Desc"),
              },
              {
                title: t("simulation.beginnerGuide.step2Title"),
                desc: t("simulation.beginnerGuide.step2Desc"),
              },
              {
                title: t("simulation.beginnerGuide.step3Title"),
                desc: t("simulation.beginnerGuide.step3Desc"),
              },
            ].map((step) => (
              <div
                key={step.title}
                className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-2"
              >
                <h3 className="text-lg font-semibold text-white leading-snug">{step.title}</h3>
                <p className="text-base text-graphite-300 leading-relaxed">{step.desc}</p>
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-80 overflow-y-auto pr-1">
            {glossaryTerms.map((item) => (
              <div
                key={item.id}
                className="rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-1.5"
              >
                <div className="text-sm font-mono font-semibold text-emerald-300">{item.term}</div>
                <p className="text-base text-graphite-200 leading-relaxed">{item.definition}</p>
                <p className="text-log font-mono text-graphite-400">
                  <span className="text-graphite-300 font-semibold">
                    {t("simulation.beginnerGuide.glossaryExampleLabel")}{" "}
                  </span>
                  {item.example}
                </p>
              </div>
            ))}
          </div>
        )}

        <div className="flex items-center gap-2 text-xs text-graphite-400 pt-1">
          <ShieldCheck className="w-4 h-4 text-emerald-400 shrink-0" />
          <span>{t("simulation.beginnerGuide.zeroSpoilerNote")}</span>
        </div>
      </div>
    </div>
  );
}
