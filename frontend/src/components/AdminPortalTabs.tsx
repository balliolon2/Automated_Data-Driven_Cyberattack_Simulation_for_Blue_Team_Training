import { Link } from "react-router-dom";
import { BarChart3, FileSpreadsheet, ShieldCheck } from "lucide-react";
import { useI18n } from "../i18n/index.tsx";
import { cn } from "../lib/utils";

interface AdminPortalTabsProps {
  active: "specialists" | "evaluations" | "statistics";
}

export default function AdminPortalTabs({ active }: AdminPortalTabsProps) {
  const { t } = useI18n();

  const tabs = [
    {
      key: "specialists" as const,
      to: "/admin/specialists",
      icon: ShieldCheck,
      label: t("admin.portalTabs.specialists"),
    },
    {
      key: "evaluations" as const,
      to: "/admin/evaluations",
      icon: FileSpreadsheet,
      label: t("admin.portalTabs.evaluations"),
    },
    {
      key: "statistics" as const,
      to: "/admin/evaluations/statistics",
      icon: BarChart3,
      label: t("admin.portalTabs.statistics"),
    },
  ];

  return (
    <nav
      aria-label={t("admin.portalTabs.ariaLabel")}
      className="print:hidden flex flex-wrap items-center gap-2 border-b border-graphite-800 pb-4"
    >
      {tabs.map((tab) => {
        const Icon = tab.icon;
        const isActive = active === tab.key;
        return (
          <Link
            key={tab.key}
            to={tab.to}
            className={cn(
              "inline-flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium border transition-colors whitespace-nowrap",
              isActive
                ? "bg-white text-black border-white font-semibold"
                : "bg-graphite-900 text-graphite-300 border-graphite-800 hover:text-white hover:border-graphite-700",
            )}
          >
            <Icon className="w-4 h-4 shrink-0" />
            <span>{tab.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
